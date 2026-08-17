package theme

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/png"  // register PNG decoder
	"io/fs"
	"path"
	"strings"
	"strconv"

	_ "golang.org/x/image/webp" // register WebP decoder
)

// supportedExts maps a lower-cased file extension to its MIME type.
// Only gif/png/webp frames are accepted.
var supportedExts = map[string]string{
	".gif":  "image/gif",
	".png":  "image/png",
	".webp": "image/webp",
}

// frameIndexFromName parses a frame's base filename (without extension)
// as an integer index. Non-integer names (e.g. "cover") return -1 so the
// caller can decide whether to include them.
func frameIndexFromName(base string) int {
	ext := strings.ToLower(path.Ext(base))
	stem := strings.TrimSuffix(base, ext)
	if _, ok := supportedExts[ext]; !ok {
		return -1
	}
	n, err := strconv.Atoi(stem)
	if err != nil || n < 0 {
		return -1
	}
	return n
}

// decodeFrame reads an image file from fsys and returns a Frame with
// pixel dimensions and a base64 data URI of the original bytes.
// Width/height come from image.DecodeConfig (header read). The data URI
// embeds the ORIGINAL bytes: built-in themes are trusted in-repo assets,
// so re-encoding would only lose quality. Upload re-encoding for
// untrusted user themes happens in M6.
func decodeFrame(fsys fs.FS, relPath, mime string) (Frame, error) {
	raw, err := fs.ReadFile(fsys, relPath)
	if err != nil {
		return Frame{}, fmt.Errorf("read %s: %w", relPath, err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return Frame{}, fmt.Errorf("decode config %s: %w", relPath, err)
	}
	return Frame{
		Width:  cfg.Width,
		Height: cfg.Height,
		Data:   "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw),
	}, nil
}
