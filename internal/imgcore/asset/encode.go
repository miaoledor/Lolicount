package asset

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/png"

	_ "image/gif"
	_ "image/png"
)

// MaxUploadBytes is the per-file size limit for uploaded images.
const MaxUploadBytes = 4 * 1024 * 1024 // 4 MiB

// MaxUploadSide is the maximum width or height in pixels for uploaded
// images.
const MaxUploadSide = 2048

// EncodeFormat is the output format for re-encoded upload images.
type EncodeFormat string

const (
	// EncodeWebP encodes as WebP (preferred for layer images). Currently
	// falls back to PNG because golang.org/x/image/webp is decode-only;
	// a WebP encoder library will be added when the upload channel (M6)
	// is implemented.
	EncodeWebP EncodeFormat = "webp"
	// EncodePNG encodes as PNG (for images requiring lossless alpha).
	EncodePNG EncodeFormat = "png"
	// EncodeGIF encodes as GIF (for animated frame themes).
	EncodeGIF EncodeFormat = "gif"
)

// ReEncodeImage decodes uploaded image bytes (trusting no client-provided
// Content-Type or extension) and re-encodes them in the specified format.
// This enforces AGENTS.md Iron Rule 4: server-side re-encoding to prevent
// image polyglot attacks. Returns the re-encoded bytes and the MIME type.
func ReEncodeImage(raw []byte, format EncodeFormat) ([]byte, string, error) {
	if len(raw) > MaxUploadBytes {
		return nil, "", fmt.Errorf("image exceeds %d bytes limit", MaxUploadBytes)
	}

	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", fmt.Errorf("decode uploaded image: %w", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() > MaxUploadSide || bounds.Dy() > MaxUploadSide {
		return nil, "", fmt.Errorf("image dimensions %dx%d exceed %dpx side limit",
			bounds.Dx(), bounds.Dy(), MaxUploadSide)
	}

	var buf bytes.Buffer
	var mime string

	switch format {
	case EncodeWebP:
		// TODO(M6): use a WebP encoder when the upload channel is
		// implemented. For now, fall back to PNG (lossless, supports
		// alpha) so the function is usable without a new dependency.
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", fmt.Errorf("png encode (webp fallback): %w", err)
		}
		mime = "image/png"
	case EncodePNG:
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", fmt.Errorf("png encode: %w", err)
		}
		mime = "image/png"
	case EncodeGIF:
		if err := gif.Encode(&buf, img, nil); err != nil {
			return nil, "", fmt.Errorf("gif encode: %w", err)
		}
		mime = "image/gif"
	default:
		return nil, "", fmt.Errorf("unsupported encode format: %s", format)
	}

	return buf.Bytes(), mime, nil
}
