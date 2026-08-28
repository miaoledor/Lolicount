// Package main implements gen-character: a one-shot generator that
// converts a kirikiri stand-layer export (fgimage/<char>/<char>Ａ_0.txt +
// <char>Ａ_0_<layer_id>.png) into a Lolicount multi-layer character theme
// (assets/theme/<theme>/{ren.json, config.json, display.json, ren/*.webp}).
//
// The export is a layered PSD split: type-0 rows carry absolute
// left/top/width/height + layer_id for each transparent PNG layer.
// Layers are categorized into four part buckets:
//
//   - lass  : body/dress layers (names: 斜め, 正面, 斜め腕上, 斜め腕下)
//   - eye   : eye+eyebrow layers (names: 0/1/2, 目, 目／...)
//   - mouth : mouth layers (name prefix 口／)
//   - face  : cheek/blush layers (name prefix 頬／)
//
// The canvas dimensions come from the _0.txt header row. The display
// crop is auto-computed as the bounding box of all visible layers,
// trimming blank canvas margins so only the portrait area is shown.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// character maps a fgimage source directory to its output theme name.
type character struct {
	dir   string // fgimage subdirectory (Japanese name)
	theme string // output theme name under assets/theme/
}

// built-in character roster. ひなた is included so it regenerates via
// the same unified pipeline as the rest.
var roster = []character{
	{"ひなた", "hinata"},
	{"七海", "nanami"},
	{"柚子", "yuzu"},
	{"湊", "minato"},
	{"美結", "miyu"},
	{"風莉", "furi"},
}

// layerRow is one parsed row of the TJS _0.txt layer-info file.
type layerRow struct {
	layerType string
	name      string
	left      int
	top       int
	width     int
	height    int
	visible   int
	layerID   int
	groupID   int
}

type manifestLayer struct {
	Name         string `json:"name"`
	Left         int    `json:"left"`
	Top          int    `json:"top"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Visible      int    `json:"visible"`
	LayerID      int    `json:"layer_id"`
	GroupLayerID int    `json:"group_layer_id"`
}

// partRange is the 1-based closed index range for a part category.
type partRange struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

type configJSON struct {
	CanvasW int                  `json:"canvasW"`
	CanvasH int                  `json:"canvasH"`
	Ranges  map[string]partRange `json:"ranges"`
}

type cropRect struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type displayJSON struct {
	Size int      `json:"size"`
	Crop *cropRect `json:"crop"`
}

func main() {
	fgRoot := "/Users/miaoledor/devCode/Lolicount/fgimage"
	outRoot := "/Users/miaoledor/devCode/Lolicount/assets/theme"

	// Optional CLI arg: generate only one character by theme name.
	filter := ""
	if len(os.Args) > 1 {
		filter = os.Args[1]
	}

	for _, c := range roster {
		if filter != "" && c.theme != filter {
			continue
		}
		if err := generate(fgRoot, outRoot, c); err != nil {
			die("%s: %v", c.theme, err)
		}
	}
}

// generate converts one character export into a theme directory.
func generate(fgRoot, outRoot string, c character) error {
	fgDir := filepath.Join(fgRoot, c.dir)
	prefix := c.dir + "Ａ_0"
	rows, canvasW, canvasH, err := parseLayerInfo(filepath.Join(fgDir, prefix+".txt"))
	if err != nil {
		return fmt.Errorf("parse layer info: %w", err)
	}

	// Categorize type-0 visible layers into the four part buckets.
	cats := map[string][]layerRow{
		"lass":  {},
		"eye":   {},
		"mouth": {},
		"face":  {},
	}
	var visLayers []layerRow
	for _, r := range rows {
		if r.layerType != "0" || r.layerID == 0 {
			continue
		}
		if r.visible == 1 {
			visLayers = append(visLayers, r)
		}
		switch {
		case isBodyLayer(r.name):
			cats["lass"] = append(cats["lass"], r)
		case r.name == "0" || r.name == "1" || r.name == "2" ||
			r.name == "目" || strings.HasPrefix(r.name, "目／"):
			cats["eye"] = append(cats["eye"], r)
		case strings.HasPrefix(r.name, "口／"):
			cats["mouth"] = append(cats["mouth"], r)
		case strings.HasPrefix(r.name, "頬／"):
			cats["face"] = append(cats["face"], r)
		}
	}

	// Build the manifest. Index 0 is a placeholder (LayerID 0, skipped
	// by assembly). Categories follow in z-stack order: lass, eye,
	// mouth, face.
	stack := []string{"lass", "eye", "mouth", "face"}
	manifest := []manifestLayer{{Name: "placeholder", LayerID: 0}}
	ranges := make(map[string]partRange)
	newID := 1
	for _, cat := range stack {
		layers := cats[cat]
		sort.Slice(layers, func(i, j int) bool {
			return layers[i].layerID < layers[j].layerID
		})
		if len(layers) == 0 {
			return fmt.Errorf("category %s has no layers", cat)
		}
		first := len(manifest)
		for _, r := range layers {
			src := filepath.Join(fgDir, fmt.Sprintf("%s_%d.png", prefix, r.layerID))
			dst := filepath.Join(outRoot, c.theme, "ren", fmt.Sprintf("%d", newID))
			if err := convertLayer(src, dst); err != nil {
				return fmt.Errorf("convert %s -> %s: %w", src, dst, err)
			}
			manifest = append(manifest, manifestLayer{
				Name:         fmt.Sprintf("%s_%d", cat, newID),
				Left:         r.left,
				Top:          r.top,
				Width:        r.width,
				Height:       r.height,
				Visible:      1,
				LayerID:      newID,
				GroupLayerID: newID,
			})
			newID++
		}
		ranges[cat] = partRange{First: first, Last: len(manifest) - 1}
	}

	outDir := filepath.Join(outRoot, c.theme)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Write ren.json.
	renJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ren.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "ren.json"), append(renJSON, '\n'), 0644); err != nil {
		return err
	}

	// Write config.json.
	cfg := configJSON{CanvasW: canvasW, CanvasH: canvasH, Ranges: ranges}
	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "config.json"), append(cfgJSON, '\n'), 0644); err != nil {
		return err
	}

	// Write display.json: crop = bounding box of all visible layers.
	crop := boundingBox(visLayers)
	disp := displayJSON{Size: 400, Crop: &crop}
	dispJSON, err := json.MarshalIndent(disp, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "display.json"), append(dispJSON, '\n'), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated %s: %d layers, canvas %dx%d, crop %dx%d\n",
		c.theme, newID-1, canvasW, canvasH, crop.Width, crop.Height)
	for _, cat := range stack {
		r := ranges[cat]
		fmt.Printf("  %s: indices %d-%d (%d layers)\n", cat, r.First, r.Last, r.Last-r.First+1)
	}
	return nil
}

// isBodyLayer reports whether a layer name is a body/dress layer.
func isBodyLayer(name string) bool {
	switch name {
	case "斜め", "正面", "斜め腕上", "斜め腕下":
		return true
	}
	return false
}

// boundingBox computes the union bounding box of a set of layers.
func boundingBox(layers []layerRow) cropRect {
	if len(layers) == 0 {
		return cropRect{}
	}
	minL, minT := layers[0].left, layers[0].top
	maxR := layers[0].left + layers[0].width
	maxB := layers[0].top + layers[0].height
	for _, r := range layers[1:] {
		if r.left < minL {
			minL = r.left
		}
		if r.top < minT {
			minT = r.top
		}
		if right := r.left + r.width; right > maxR {
			maxR = right
		}
		if bottom := r.top + r.height; bottom > maxB {
			maxB = bottom
		}
	}
	return cropRect{Left: minL, Top: minT, Width: maxR - minL, Height: maxB - minT}
}

// parseLayerInfo reads a UTF-16LE TJS layer-info .txt file and returns
// the rows plus the canvas dimensions from the header.
func parseLayerInfo(path string) (rows []layerRow, canvasW, canvasH int, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, 0, err
	}
	text := decodeUTF16LE(raw)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 11 || strings.HasPrefix(strings.TrimSpace(fields[0]), "#") {
			continue
		}
		// The row right after the header carries canvas dims:
		// "				<width>	<height>	...".
		if i == 1 && len(fields) >= 6 {
			canvasW, _ = strconv.Atoi(strings.TrimSpace(fields[4]))
			canvasH, _ = strconv.Atoi(strings.TrimSpace(fields[5]))
		}
		lt := strings.TrimSpace(fields[0])
		if lt != "0" && lt != "2" {
			continue
		}
		left, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
		top, _ := strconv.Atoi(strings.TrimSpace(fields[3]))
		w, _ := strconv.Atoi(strings.TrimSpace(fields[4]))
		h, _ := strconv.Atoi(strings.TrimSpace(fields[5]))
		vis, _ := strconv.Atoi(strings.TrimSpace(fields[8]))
		lid, _ := strconv.Atoi(strings.TrimSpace(fields[9]))
		gid, _ := strconv.Atoi(strings.TrimSpace(fields[10]))
		rows = append(rows, layerRow{
			layerType: lt, name: strings.TrimSpace(fields[1]),
			left: left, top: top, width: w, height: h,
			visible: vis, layerID: lid, groupID: gid,
		})
	}
	if canvasW == 0 || canvasH == 0 {
		return nil, 0, 0, fmt.Errorf("canvas dimensions not found in %s", path)
	}
	return rows, canvasW, canvasH, nil
}

// decodeUTF16LE decodes a UTF-16LE byte slice to a UTF-8 string,
// stripping the BOM and any embedded null bytes.
func decodeUTF16LE(b []byte) string {
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE {
		b = b[2:]
	}
	var sb strings.Builder
	for i := 0; i+1 < len(b); i += 2 {
		r := rune(b[i]) | rune(b[i+1])<<8
		if r == 0 {
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// convertLayer converts a source PNG to webp at dst (without extension).
// Falls back to copying the verbatim PNG if cwebp is unavailable.
func convertLayer(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	webpDst := dst + ".webp"
	if out, err := exec.Command("cwebp", "-q", "80", src, "-o", webpDst).CombinedOutput(); err == nil {
		return nil
	} else {
		_ = out
		pngDst := dst + ".png"
		in, err := os.Open(src)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(pngDst)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	}
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-character: "+format+"\n", args...)
	os.Exit(1)
}
