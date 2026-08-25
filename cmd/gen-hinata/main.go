// Package main implements gen-hinata: a one-shot generator that converts
// the ひなた TJS2 stand-layer export (fgimage/ひなた/ひなたＡ_0.txt +
// ひなたＡ_0_<layer_id>.png) into a Lolicount character theme
// (assets/character/hinata/{ren.json, config.json, ren/*.png}).
//
// The ひなた export is a layered PSD split: type-0 rows carry absolute
// left/top/width/height + layer_id for each transparent PNG layer.
// Layers are categorized into five part buckets to match the existing
// character-theme assembly model:
//
//   - lass  : dress/body layers (name "斜め"), 8 variants
//   - brow  : eye+eyebrow layers named "目" (special expressions), 9
//   - eye   : eye+eyebrow layers named "0"/"1"/"2" (gaze directions), 45
//   - mouth : mouth layers (name "口／…"), 13
//   - face  : cheek/blush layers (name "頬／…"), 26
//
// The canvas is 2362 x 4134 (from the _0.txt header row).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// layerRow is one parsed row of the TJS _0.txt layer info file.
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

// manifestLayer is the ren.json entry format expected by characterthemedrawer.
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

// configJSON is the per-theme layout config consumed by characterthemedrawer.
type configJSON struct {
	CanvasW int                  `json:"canvasW"`
	CanvasH int                  `json:"canvasH"`
	Ranges  map[string]partRange `json:"ranges"`
}

func main() {
	fgDir := "/Users/miaoledor/devCode/Lolicount/fgimage/ひなた"
	outDir := "/Users/miaoledor/devCode/Lolicount/assets/character/hinata"
	prefix := "ひなたＡ_0"

	rows, canvasW, canvasH, err := parseLayerInfo(filepath.Join(fgDir, prefix+".txt"))
	if err != nil {
		die("parse layer info: %v", err)
	}

	// Categorize type-0 layers into the five part buckets.
	cats := map[string][]layerRow{
		"lass":  {},
		"eye":   {},
		"mouth": {},
		"face":  {},
	}
	for _, r := range rows {
		if r.layerType != "0" || r.layerID == 0 {
			continue
		}
		switch {
		case r.name == "斜め":
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

	// Build the manifest. Index 0 is a placeholder (LayerID 0, skipped by
	// assembly, matching the 莲 convention). Categories follow in z-stack
	// order: lass, brow, eye, mouth, face.
	stack := []string{"lass", "eye", "mouth", "face"}
	manifest := []manifestLayer{{
		Name: "placeholder", LayerID: 0,
	}}
	ranges := make(map[string]partRange)
	newID := 1 // sequential layer_id used in ren/ filenames
	for _, cat := range stack {
		layers := cats[cat]
		sort.Slice(layers, func(i, j int) bool {
			return layers[i].layerID < layers[j].layerID
		})
		if len(layers) == 0 {
			die("category %s has no layers", cat)
		}
		first := len(manifest) // 0-based array index of first entry
		for _, r := range layers {
			src := filepath.Join(fgDir, fmt.Sprintf("%s_%d.png", prefix, r.layerID))
			dst := filepath.Join(outDir, "ren", fmt.Sprintf("%d", newID))
			if err := convertLayer(src, dst, 0); err != nil {
				die("convert %s -> %s: %v", src, dst, err)
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

	// Write ren.json.
	renJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		die("marshal ren.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "ren.json"), append(renJSON, '\n'), 0644); err != nil {
		die("write ren.json: %v", err)
	}

	// Write config.json.
	cfg := configJSON{
		CanvasW: canvasW,
		CanvasH: canvasH,
		Ranges:  ranges,
	}
	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		die("marshal config.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "config.json"), append(cfgJSON, '\n'), 0644); err != nil {
		die("write config.json: %v", err)
	}

	fmt.Printf("Generated hinata theme: %d layers, canvas %dx%d\n", newID-1, canvasW, canvasH)
	for _, cat := range stack {
		r := ranges[cat]
		fmt.Printf("  %s: indices %d-%d (%d layers)\n", cat, r.First, r.Last, r.Last-r.First+1)
	}
}

// parseLayerInfo reads a UTF-16LE TJS layer-info .txt file and returns
// the type-0/type-2 rows plus the canvas dimensions from the header.
func parseLayerInfo(path string) (rows []layerRow, canvasW, canvasH int, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, 0, err
	}
	text := decodeUTF16LE(raw)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 11 {
			continue
		}
		if strings.HasPrefix(fields[0], "#") {
			continue
		}
		// The row right after the header carries canvas dims. Its format
		// is "				<width>	<height>	..." so width=fields[4], height=fields[5].
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
			layerType: lt,
			name:      strings.TrimSpace(fields[1]),
			left:      left,
			top:       top,
			width:     w,
			height:    h,
			visible:   vis,
			layerID:   lid,
			groupID:   gid,
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

// convertLayer converts a source PNG to webp at dst (without the .webp
// extension, which is added here). Body layers (maxBodyHeight) are
// downscaled to keep file sizes small; the viewBox mapping in Draw
// stretches them back to canvas coordinates so placement stays correct.
// If cwebp is unavailable it falls back to copying the verbatim PNG so
// generation never blocks on a missing tool.
func convertLayer(src, dst string, maxHeight int) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	webpDst := dst + ".webp"
	var args []string
	args = append(args, "-q", "80")
	if maxHeight > 0 {
		args = append(args, "-resize", "0", strconv.Itoa(maxHeight))
	}
	args = append(args, src, "-o", webpDst)
	if out, err := exec.Command("cwebp", args...).CombinedOutput(); err == nil {
		return nil
	} else {
		_ = out
		// Fallback: copy the verbatim PNG.
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
	fmt.Fprintf(os.Stderr, "gen-hinata: "+format+"\n", args...)
	os.Exit(1)
}
