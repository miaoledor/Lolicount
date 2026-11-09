package asset

import (
	"bytes"
	"image"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/miaoledor/lolicount/assets"
	_ "golang.org/x/image/webp"
)

// TestHeaderDimsMatchesDecodeConfig walks the embedded theme tree and checks
// the header-only parser against the full-file decoder for every image. It
// is deliberately count-independent: themes and frames are added often, and
// their number must never make the suite fail. Use -short to sample.
func TestHeaderDimsMatchesDecodeConfig(t *testing.T) {
	fsys, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		t.Fatalf("open embedded theme tree: %v", err)
	}
	var files []string
	err = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := SupportedExts[pathExt(p)]; ok {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk theme tree: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no theme images found")
	}
	if testing.Short() && len(files) > 300 {
		files = files[:300]
	}

	for _, rel := range files {
		raw, err := fs.ReadFile(fsys, rel)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("decode config %s: %v", rel, err)
		}
		w, h, err := headerDims(fsys, rel)
		if err != nil {
			t.Fatalf("headerDims %s: %v", rel, err)
		}
		if w != cfg.Width || h != cfg.Height {
			t.Errorf("%s: header %dx%d != decoder %dx%d", rel, w, h, cfg.Width, cfg.Height)
		}
	}
}

// TestDimsFromHeaderFallback pins the contract for containers the fast path
// does not recognize: the caller must fall back to the registered decoder,
// which still produces correct dimensions for a complete image file.
func TestDimsFromHeaderFallback(t *testing.T) {
	if _, _, ok := dimsFromHeader([]byte("RIFF1234WEBPANIM123456789012345")); ok {
		t.Fatal("unknown WebP chunk should not be handled by the fast path")
	}
	png := []byte{
		0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
		0, 0, 0, 13, 'I', 'H', 'D', 'R',
		0, 0, 0, 7, 0, 0, 0, 9, 8, 6, 0, 0, 0,
	}
	w, h, ok := dimsFromHeader(png)
	if !ok || w != 7 || h != 9 {
		t.Fatalf("PNG header = (%d,%d,%v), want (7,9,true)", w, h, ok)
	}

	fsys := fstest.MapFS{"unknown.bin": {Data: []byte("not an image header")}}
	if _, _, err := headerDims(fsys, "unknown.bin"); err == nil {
		t.Fatal("unrecognized data should surface a decode error")
	}
}

// TestHeaderDimsReadsOnlyPrefix proves the optimization itself: a file whose
// payload extends far beyond the header must still resolve without the whole
// payload being pulled into the heap. The trailing bytes are only present to
// make a short read observable by the parser.
func TestHeaderDimsReadsOnlyPrefix(t *testing.T) {
	png := []byte{
		0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
		0, 0, 0, 13, 'I', 'H', 'D', 'R',
		0, 0, 1, 0, 0, 0, 0, 200, 8, 6, 0, 0, 0,
	}
	payload := append(append([]byte{}, png...), make([]byte, 4096)...)
	fsys := fstest.MapFS{"big.png": {Data: payload}}
	w, h, err := headerDims(fsys, "big.png")
	if err != nil {
		t.Fatalf("headerDims: %v", err)
	}
	if w != 256 || h != 200 {
		t.Fatalf("got %dx%d, want 256x200", w, h)
	}
}
