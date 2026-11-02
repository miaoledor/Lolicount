package asset

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// CharacterConfig describes a character theme's PSD layout: canvas
// dimensions and closed manifest-index ranges for each part category.
type CharacterConfig struct {
	CanvasW int                  `json:"canvasW"`
	CanvasH int                  `json:"canvasH"`
	Ranges  map[string]PartRange `json:"ranges"`
}

// PartRange is a closed interval [First, Last] of zero-based manifest
// array indexes belonging to one part category.
type PartRange struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

// CharacterManifest is one entry in ren.json: the absolute placement of
// a layer in the original PSD canvas.
type CharacterManifest struct {
	Name         string `json:"name"`
	Left         int    `json:"left"`
	Top          int    `json:"top"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Visible      int    `json:"visible"`
	LayerID      int    `json:"layer_id"`
	GroupLayerID int    `json:"group_layer_id"`
}

// readCharacterManifest parses ren.json plus config.json (when present
// and valid). config.json is required: without its ranges there is no
// slot structure to build.
func readCharacterManifest(fsys fs.FS, dir string) ([]CharacterManifest, *CharacterConfig, error) {
	manifestPath := path.Join(dir, ManifestName)
	raw, err := fs.ReadFile(fsys, manifestPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", manifestPath, err)
	}
	var manifest []CharacterManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", manifestPath, err)
	}
	if len(manifest) == 0 {
		return nil, nil, fmt.Errorf("%s: empty manifest", manifestPath)
	}

	configPath := path.Join(dir, "config.json")
	craw, err := fs.ReadFile(fsys, configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", configPath, err)
	}
	var cfg CharacterConfig
	if err := json.Unmarshal(craw, &cfg); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", configPath, err)
	}
	if cfg.CanvasW <= 0 || cfg.CanvasH <= 0 || len(cfg.Ranges) == 0 {
		return nil, nil, fmt.Errorf("%s: invalid config (canvas/ranges)", configPath)
	}
	return manifest, &cfg, nil
}

// readCharacterDisplay parses display.json when present and valid.
func readCharacterDisplay(fsys fs.FS, dir string) *theme.DisplayConfig {
	raw, err := fs.ReadFile(fsys, path.Join(dir, "display.json"))
	if err != nil {
		return nil
	}
	var dp theme.DisplayConfig
	if err := json.Unmarshal(raw, &dp); err != nil {
		return nil
	}
	if dp.Size > 0 {
		return &dp
	}
	return nil
}

// characterLayerDimensions preserves the manifest's intended display
// size while keeping older manifests without dimensions loadable.
func characterLayerDimensions(layer CharacterManifest, fallbackWidth, fallbackHeight int) (int, int) {
	width := layer.Width
	if width <= 0 {
		width = fallbackWidth
	}
	height := layer.Height
	if height <= 0 {
		height = fallbackHeight
	}
	return width, height
}
