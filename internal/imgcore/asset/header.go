// Package asset header.go recovers pixel dimensions from the leading bytes
// of an embedded image. The catalog build walks tens of thousands of files;
// reading each one in full pulls the whole embedded rodata into the heap and
// the process RSS even though only a width and a height are needed.
package asset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"io/fs"
)

// headerBytes covers the largest header this package parses: a WebP VP8 or
// VP8X chunk needs 30 bytes, PNG IHDR needs 24, GIF needs 10.
const headerBytes = 32

// headerDims opens an image, reads only its leading bytes and returns the
// pixel dimensions. Recognized containers are parsed directly; anything else
// falls back to the registered image decoders, with the consumed prefix
// replayed so the decoder still sees the file from byte zero. DecodeConfig
// stops at the header by contract, so neither path retains image payloads.
func headerDims(fsys fs.FS, relPath string) (int, int, error) {
	f, err := fsys.Open(relPath)
	if err != nil {
		return 0, 0, fmt.Errorf("read %s: %w", relPath, err)
	}
	defer f.Close()

	var head [headerBytes]byte
	n, err := io.ReadFull(f, head[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return 0, 0, fmt.Errorf("read %s: %w", relPath, err)
	}
	if w, h, ok := dimsFromHeader(head[:n]); ok {
		return w, h, nil
	}

	cfg, _, err := image.DecodeConfig(io.MultiReader(bytes.NewReader(head[:n]), f))
	if err != nil {
		return 0, 0, fmt.Errorf("decode config %s: %w", relPath, err)
	}
	return cfg.Width, cfg.Height, nil
}

// dimsFromHeader parses PNG IHDR, the GIF logical screen descriptor, and the
// WebP VP8/VP8L/VP8X payload prefixes. A false result is not an error: it
// means the file needs the registered decoder instead. Dimensions are read
// from the container exactly as image.DecodeConfig reports them, so catalog
// canvas sizes stay byte-for-byte compatible with the previous full-file path.
func dimsFromHeader(head []byte) (int, int, bool) {
	if len(head) >= 24 && string(head[:8]) == "\x89PNG\r\n\x1a\n" && string(head[12:16]) == "IHDR" {
		return int(binary.BigEndian.Uint32(head[16:20])),
			int(binary.BigEndian.Uint32(head[20:24])), true
	}
	if len(head) >= 10 {
		magic := string(head[:6])
		if magic == "GIF87a" || magic == "GIF89a" {
			return int(binary.LittleEndian.Uint16(head[6:8])),
				int(binary.LittleEndian.Uint16(head[8:10])), true
		}
	}
	if len(head) < 30 || string(head[:4]) != "RIFF" || string(head[8:12]) != "WEBP" {
		return 0, 0, false
	}
	switch string(head[12:16]) {
	case "VP8X":
		w := uint32(head[24]) | uint32(head[25])<<8 | uint32(head[26])<<16
		h := uint32(head[27]) | uint32(head[28])<<8 | uint32(head[29])<<16
		return int(w) + 1, int(h) + 1, true
	case "VP8L":
		if head[20] != 0x2f {
			return 0, 0, false
		}
		bits := binary.LittleEndian.Uint32(head[21:25])
		return int(bits&0x3fff) + 1, int((bits>>14)&0x3fff) + 1, true
	case "VP8 ":
		if head[23] != 0x9d || head[24] != 0x01 || head[25] != 0x2a {
			return 0, 0, false
		}
		return int(binary.LittleEndian.Uint16(head[26:28]) & 0x3fff),
			int(binary.LittleEndian.Uint16(head[28:30]) & 0x3fff), true
	}
	return 0, 0, false
}
