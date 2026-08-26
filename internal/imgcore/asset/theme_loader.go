package asset

import (
	"fmt"
	"io/fs"
	"sort"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// ManifestName is the marker file that distinguishes a multi-layer
// (character) theme from a single-layer (frame) theme inside the unified
// assets/theme/ tree. Its presence triggers the layered loader; its
// absence triggers the frame loader.
const ManifestName = "ren.json"

// LoadThemes scans the embedded assets/theme/ tree and returns every
// valid theme as a ready-to-render *theme.Theme. Each subdirectory is
// dispatched by structure: a ren.json manifest marks a multi-layer
// (character) theme, otherwise the directory is loaded as a single-layer
// frame theme. This is the sole entry point for builtin theme loading —
// there is no longer a separate card/character registry split.
func LoadThemes() (map[string]*theme.Theme, []error) {
	sub, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		return nil, []error{fmt.Errorf("theme loader: open embedded theme: %w", err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return nil, []error{fmt.Errorf("theme loader: read theme: %w", err)}
	}

	themes := make(map[string]*theme.Theme)
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		t, err := loadTheme(sub, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		themes[name] = t
	}
	return themes, errs
}

// loadTheme dispatches one theme directory to the correct loader based
// on the presence of ren.json. A multi-layer theme without a valid
// config returns an error; a frame theme with no frames returns an
// error. Both paths produce a *theme.Theme consumed by the same compose
// pipeline.
func loadTheme(fsys fs.FS, name string) (*theme.Theme, error) {
	if isManifestTheme(fsys, name) {
		ct, err := LoadCharacterTheme(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		t, err := CharacterThemeToTheme(ct)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		return t, nil
	}

	ct, err := LoadFrameTheme(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("theme %s: %w", name, err)
	}
	return FrameThemeToTheme(ct), nil
}

// isManifestTheme reports whether the theme directory contains a ren.json
// manifest, marking it as a multi-layer (character) theme.
func isManifestTheme(fsys fs.FS, name string) bool {
	_, err := fs.Stat(fsys, name+"/"+ManifestName)
	return err == nil
}

// FrameTheme is a named ordered-frame theme loaded from assets/theme.
// Frames is the ordered set (0..size-1); the caller picks a frame index.
type FrameTheme struct {
	Name   string
	Frames []render.ImageLayer
}

// LoadFrameTheme decodes one frame-theme directory into an ordered
// FrameTheme. Files are named <index>.<ext>; indices need not be
// contiguous but must be non-negative. Frames are stored sorted by index.
func LoadFrameTheme(fsys fs.FS, name string) (*FrameTheme, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
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
		ext := pathExt(base)
		mime, ok := SupportedExts[ext]
		if !ok {
			continue
		}
		decoded, err := DecodeImage(fsys, name+"/"+base, mime)
		if err != nil {
			return nil, err
		}
		frames = append(frames, indexed{idx: idx, img: render.ImageLayer{
			Src:       decoded.Data,
			Width:     decoded.Width,
			Height:    decoded.Height,
			Transform: imgcore.DefaultTransform(),
		}})
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("no frame images found")
	}

	sort.Slice(frames, func(i, j int) bool { return frames[i].idx < frames[j].idx })
	th := &FrameTheme{Name: name, Frames: make([]render.ImageLayer, len(frames))}
	for i, fr := range frames {
		fr.img.Z = i
		th.Frames[i] = fr.img
	}
	return th, nil
}

// FrameThemeToTheme converts a FrameTheme into a *theme.Theme. When the
// theme has multiple frames, they are stored as a RandomPickLayer so each
// request randomly picks one frame. A single-frame theme uses a plain
// ImageLayer. The canvas is the first frame's pixel dimensions; the
// caller (server compose path) applies scale/text at render time.
func FrameThemeToTheme(ft *FrameTheme) *theme.Theme {
	if ft == nil || len(ft.Frames) == 0 {
		return &theme.Theme{Name: "", Canvas: theme.Canvas{}, Layers: nil}
	}
	frame := ft.Frames[0]
	canvasW := frame.Width
	canvasH := frame.Height

	var layer imgcore.Layer
	if len(ft.Frames) == 1 {
		layer = &frame
	} else {
		opts := make([]render.ImageOption, len(ft.Frames))
		for i, fr := range ft.Frames {
			fr.Z = 0
			opts[i] = render.ImageOption{ImageLayer: fr, Weight: 1}
		}
		layer = &render.RandomPickLayer{
			Category:  ft.Name,
			Options:   opts,
			Transform: imgcore.DefaultTransform(),
			Z:         0,
		}
	}

	return &theme.Theme{
		Name:   ft.Name,
		Canvas: theme.Canvas{Width: canvasW, Height: canvasH},
		BgW:    canvasW,
		BgH:    canvasH,
		Layers: []imgcore.Layer{layer},
	}
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
