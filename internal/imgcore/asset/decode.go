// Package asset provides the shared asset-loading layer for imgcore:
// image decoding (file bytes to data URI + pixel dimensions) and theme
// loading from the embedded assets.FS. Merged and generalized from
// cardthemedrawer/decode.go and characterthemedrawer's readLayerDataURI.
package asset

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io/fs"
	"path"
	"strconv"
	"strings"

	_ "image/gif"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// SupportedExts maps a lower-cased file extension to its MIME type.
// Only gif/png/webp images are accepted.
var SupportedExts = map[string]string{
	".gif":  "image/gif",
	".png":  "image/png",
	".webp": "image/webp",
}

// DecodedImage is an image file decoded into a data URI plus its pixel
// dimensions. The data URI is embedded directly into SVG so the counter
// renders offline (AGENTS.md Iron Rule 2).
type DecodedImage struct {
	Width  int
	Height int
	Data   string // data:<mime>;base64,...
}

// DecodeImage reads an image file from fsys and returns a DecodedImage
// with pixel dimensions and a base64 data URI of the original bytes.
// Merged from cardthemedrawer.decodeFrame and
// characterthemedrawer.readLayerDataURI.
func DecodeImage(fsys fs.FS, relPath, mime string) (DecodedImage, error) {
	raw, err := fs.ReadFile(fsys, relPath)
	if err != nil {
		return DecodedImage{}, fmt.Errorf("read %s: %w", relPath, err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return DecodedImage{}, fmt.Errorf("decode config %s: %w", relPath, err)
	}
	return DecodedImage{
		Width:  cfg.Width,
		Height: cfg.Height,
		Data:   "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw),
	}, nil
}

// FindImageFile searches for an image file at relPath with any supported
// extension (.webp, .png, .gif). Returns the full path and MIME type, or
// an error if no image is found. Used by character theme loading where
// the layer ID is known but the extension is not.
func FindImageFile(fsys fs.FS, relPath string) (string, string, error) {
	for _, ext := range []string{".webp", ".png", ".gif"} {
		p := relPath + ext
		if _, err := fs.Stat(fsys, p); err == nil {
			mime, ok := SupportedExts[strings.ToLower(ext)]
			if !ok {
				continue
			}
			return p, mime, nil
		}
	}
	return "", "", fmt.Errorf("no image found at %s", relPath)
}

// FrameIndexFromName parses a frame's base filename (without extension)
// as an integer index. Non-integer names return -1. Migrated from
// cardthemedrawer.frameIndexFromName.
func FrameIndexFromName(base string) int {
	ext := strings.ToLower(path.Ext(base))
	stem := strings.TrimSuffix(base, ext)
	if _, ok := SupportedExts[ext]; !ok {
		return -1
	}
	n, err := strconv.Atoi(stem)
	if err != nil || n < 0 {
		return -1
	}
	return n
}
