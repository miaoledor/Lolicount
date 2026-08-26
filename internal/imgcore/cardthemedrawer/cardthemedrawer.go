// Package cardthemedrawer owns the card (frame) theme assets and the
// layer-0 background drawer for card themes. It loads frame themes from
// the embedded assets/theme tree, provides a Registry for name lookup,
// and draws one frame as a data-URI <image> Layer.
//
// This package merged the former internal/theme package's frame-theme
// data types (Theme, Frame, Registry) and loading logic, plus the
// background-drawing half of composeSVG.
package cardthemedrawer

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
)

// DefaultTheme is the theme used when the request omits ?theme=.
const DefaultTheme = "lian"

// Frame is one decoded image of a theme: pixel dimensions and a base64
// data URI of the original bytes. The data URI is embedded directly into
// the SVG so the counter renders offline (AGENTS.md Iron Rule 2).
type Frame struct {
	Width  int
	Height int
	Data   string // data:<mime>;base64,...
}

// Theme is a named ordered-frame theme. Frames is the ordered set
// (0..size-1); the caller picks a frame index.
type Theme struct {
	Name   string
	Frames []Frame
}

// Size returns the number of frames in the theme.
func (t *Theme) Size() int { return len(t.Frames) }

// FrameAt returns the frame at index, or false if out of range.
func (t *Theme) FrameAt(index int) (Frame, bool) {
	if index < 0 || index >= len(t.Frames) {
		return Frame{}, false
	}
	return t.Frames[index], true
}

// Registry resolves a card theme name to a Theme. Reserved names "demo"
// and "random" are handled by the caller (renderer), not by the registry.
type Registry interface {
	Get(name string) (*Theme, bool)
	List() []string
}

// Draw renders one frame as the layer-0 background: a data-URI <image>
// scaled to a uniform display size with aspect ratio preserved (M5.6).
// It knows nothing about the text layer — it only returns the image
// fragment and the canvas it occupies.
func Draw(frame Frame, scale float64) imgcore.LegacyLayer {
	display := imgutils.DisplaySize(scale)
	imgW, imgH := imgutils.ScaledDims(frame.Width, frame.Height, display)

	var b strings.Builder
	fmt.Fprintf(&b, `  <image x="0" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		imgW, imgH, frame.Data)
	return imgcore.LegacyLayer{Fragment: b.String(), Width: imgW, Height: imgH}
}

// builtinRegistry loads card themes from the embedded assets/theme tree
// at construction time.
type builtinRegistry struct {
	themes map[string]*Theme
}

// NewBuiltinRegistry scans the embedded assets/theme directory and loads
// every valid frame theme into memory. A theme that fails to load is
// skipped with an error; the registry still returns successfully with
// the themes that did load.
func NewBuiltinRegistry() (Registry, []error) {
	reg := &builtinRegistry{themes: make(map[string]*Theme)}
	var errs []error

	sub, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		return reg, []error{fmt.Errorf("cardthemedrawer: open embedded theme: %w", err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("cardthemedrawer: read theme: %w", err)}
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		th, err := loadFrameTheme(sub, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.themes[name] = th
	}
	return reg, errs
}

// loadFrameTheme decodes one frame-theme directory into an ordered Frame
// slice. Files are named <index>.<ext>; indices need not be contiguous
// but must be non-negative. Frames are stored sorted by index.
func loadFrameTheme(fsys fs.FS, name string) (*Theme, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("theme %s: read dir: %w", name, err)
	}

	type indexed struct {
		idx int
		f   Frame
	}
	var frames []indexed
	for _, e := range entries {
		base := e.Name()
		idx := frameIndexFromName(base)
		if idx < 0 {
			continue
		}
		ext := strings.ToLower(pathExt(base))
		mime, ok := supportedExts[ext]
		if !ok {
			continue
		}
		f, err := decodeFrame(fsys, name+"/"+base, mime)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		frames = append(frames, indexed{idx: idx, f: f})
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("theme %s: no frame images found", name)
	}

	sort.Slice(frames, func(i, j int) bool { return frames[i].idx < frames[j].idx })
	th := &Theme{Name: name, Frames: make([]Frame, len(frames))}
	for i, fr := range frames {
		th.Frames[i] = fr.f
	}
	return th, nil
}

// pathExt is path.Ext without importing path here.
func pathExt(name string) string {
	for i := len(name) - 1; i >= 0 && name[i] != '/'; i-- {
		if name[i] == '.' {
			return name[i:]
		}
	}
	return ""
}

// Get returns the theme for name, or false if not registered.
func (r *builtinRegistry) Get(name string) (*Theme, bool) {
	t, ok := r.themes[name]
	return t, ok
}

// List returns registered theme names sorted for stable output.
func (r *builtinRegistry) List() []string {
	out := make([]string, 0, len(r.themes))
	for k := range r.themes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

