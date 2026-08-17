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

	_ "golang.org/x/image/webp" // register WebP decoder
)

// supportedExts maps a lower-cased file extension to its MIME type.
// Only gif/png/webp glyphs are accepted (AGENTS.md theme convention).
var supportedExts = map[string]string{
	".gif":  "image/gif",
	".png":  "image/png",
	".webp": "image/webp",
}

// isSupportedGlyph reports whether name has a supported image extension.
func isSupportedGlyph(name string) (ext, mime string, ok bool) {
	ext = strings.ToLower(path.Ext(name))
	mime, ok = supportedExts[ext]
	return ext, mime, ok
}

// decodeGlyph reads an image file from fsys and returns a ThemeChar with
// the pixel dimensions and a base64 data URI of the original bytes.
//
// Width/height come from image.DecodeConfig (cheap header read). The data
// URI embeds the ORIGINAL bytes, not a re-encode: built-in themes are
// trusted assets curated in-repo, so re-encoding would only lose quality.
// (Upload re-encoding for untrusted user themes happens in M6.)
func decodeGlyph(fsys fs.FS, relPath, mime string) (ThemeChar, error) {
	raw, err := fs.ReadFile(fsys, relPath)
	if err != nil {
		return ThemeChar{}, fmt.Errorf("read %s: %w", relPath, err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return ThemeChar{}, fmt.Errorf("decode config %s: %w", relPath, err)
	}

	return ThemeChar{
		Width:  cfg.Width,
		Height: cfg.Height,
		Data:   "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw),
	}, nil
}
