package asset

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
)

// CardTheme is a named ordered-frame theme loaded from assets/theme.
// Frames is the ordered set (0..size-1); the caller picks a frame index.
// Migrated from cardthemedrawer.Theme.
type CardTheme struct {
	Name   string
	Frames []render.ImageLayer
}

// Size returns the number of frames in the theme.
func (t *CardTheme) Size() int { return len(t.Frames) }

// FrameAt returns the frame at index, or false if out of range.
func (t *CardTheme) FrameAt(index int) (render.ImageLayer, bool) {
	if index < 0 || index >= len(t.Frames) {
		return render.ImageLayer{}, false
	}
	return t.Frames[index], true
}

// CardRegistry resolves a card theme name to its CardTheme.
type CardRegistry interface {
	Get(name string) (*CardTheme, bool)
	List() []string
}

// builtinCardRegistry loads card themes from the embedded assets/theme
// tree at construction time.
type builtinCardRegistry struct {
	themes map[string]*CardTheme
}

// NewBuiltinCardRegistry scans the embedded assets/theme directory and
// loads every valid frame theme into memory. A theme that fails to load
// is skipped with an error; the registry still returns successfully
// with the themes that did load.
func NewBuiltinCardRegistry() (CardRegistry, []error) {
	reg := &builtinCardRegistry{themes: make(map[string]*CardTheme)}
	var errs []error

	sub, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		return reg, []error{fmt.Errorf("card loader: open embedded theme: %w", err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("card loader: read theme: %w", err)}
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		th, err := LoadCardTheme(sub, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.themes[name] = th
	}
	return reg, errs
}

// LoadCardTheme decodes one frame-theme directory into an ordered
// CardTheme. Files are named <index>.<ext>; indices need not be
// contiguous but must be non-negative. Frames are stored sorted by index.
func LoadCardTheme(fsys fs.FS, name string) (*CardTheme, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("theme %s: read dir: %w", name, err)
	}

	type indexed struct {
		idx int
		img render.ImageLayer
	}
	var frames []indexed
	for _, e := range entries {
		base := e.Name()
		idx := FrameIndexFromName(base)
		if idx < 0 {
			continue
		}
		ext := strings.ToLower(pathExt(base))
		mime, ok := SupportedExts[ext]
		if !ok {
			continue
		}
		decoded, err := DecodeImage(fsys, name+"/"+base, mime)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		frames = append(frames, indexed{idx: idx, img: render.ImageLayer{
			Src:       decoded.Data,
			Width:     decoded.Width,
			Height:    decoded.Height,
			Transform: imgcore.DefaultTransform(),
		}})
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("theme %s: no frame images found", name)
	}

	sort.Slice(frames, func(i, j int) bool { return frames[i].idx < frames[j].idx })
	th := &CardTheme{Name: name, Frames: make([]render.ImageLayer, len(frames))}
	for i, fr := range frames {
		fr.img.Z = i
		th.Frames[i] = fr.img
	}
	return th, nil
}

// Get returns the theme for name, or false if not registered.
func (r *builtinCardRegistry) Get(name string) (*CardTheme, bool) {
	t, ok := r.themes[name]
	return t, ok
}

// List returns registered theme names sorted for stable output.
func (r *builtinCardRegistry) List() []string {
	out := make([]string, 0, len(r.themes))
	for k := range r.themes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// pathExt returns the file extension including the leading dot.
func pathExt(name string) string {
	for i := len(name) - 1; i >= 0 && name[i] != '/'; i-- {
		if name[i] == '.' {
			return name[i:]
		}
	}
	return ""
}

