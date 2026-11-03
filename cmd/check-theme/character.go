package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	themeasset "github.com/miaoledor/lolicount/internal/imgcore/asset"
)

// validateCharacterTheme checks that a layered theme can be assembled by
// the runtime loader, in addition to validating its image files.
func validateCharacterTheme(sub fs.FS, name string, rep *themeReport) {
	layers, manifestOK := validateCharacterManifest(sub, name, rep)
	config, configOK := validateCharacterConfig(sub, name, rep)
	imageFiles := validateCharacterImages(sub, name, rep)
	if manifestOK && configOK {
		validateCharacterRanges(layers, config, imageFiles, rep)
	}
}

func validateCharacterManifest(sub fs.FS, name string, rep *themeReport) ([]themeasset.CharacterManifest, bool) {
	raw, err := fs.ReadFile(sub, path.Join(name, "ren.json"))
	if err != nil {
		rep.fail("read ren.json: %v", err)
		return nil, false
	}
	if len(raw) > maxFileBytes {
		rep.fail("ren.json: %d bytes exceeds %d limit", len(raw), maxFileBytes)
	}
	var layers []themeasset.CharacterManifest
	if err := json.Unmarshal(raw, &layers); err != nil {
		rep.fail("ren.json: invalid JSON: %v", err)
		return nil, false
	}
	if len(layers) == 0 {
		rep.fail("ren.json: empty manifest")
		return nil, false
	}
	return layers, true
}

func validateCharacterConfig(sub fs.FS, name string, rep *themeReport) (*themeasset.CharacterConfig, bool) {
	raw, err := fs.ReadFile(sub, path.Join(name, "config.json"))
	if err != nil {
		rep.fail("read config.json: %v", err)
		return nil, false
	}
	if len(raw) > maxFileBytes {
		rep.fail("config.json: %d bytes exceeds %d limit", len(raw), maxFileBytes)
	}
	var config themeasset.CharacterConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		rep.fail("config.json: invalid JSON: %v", err)
		return nil, false
	}
	if len(config.Sets) > 0 {
		valid := true
		for index, set := range config.Sets {
			label := characterSetLabel(index, set.Name)
			if set.CanvasW <= 0 || set.CanvasH <= 0 {
				rep.fail("config.json: %s canvasW and canvasH must be positive", label)
				valid = false
			}
			if len(set.Ranges) == 0 {
				rep.fail("config.json: %s ranges must not be empty", label)
				valid = false
			}
		}
		return &config, valid
	}
	if config.CanvasW <= 0 || config.CanvasH <= 0 {
		rep.fail("config.json: canvasW and canvasH must be positive")
	}
	if len(config.Ranges) == 0 {
		rep.fail("config.json: ranges must not be empty")
	}
	return &config, config.CanvasW > 0 && config.CanvasH > 0 && len(config.Ranges) > 0
}

func validateCharacterImages(sub fs.FS, name string, rep *themeReport) map[int]string {
	renDir := path.Join(name, "ren")
	renFiles, err := fs.ReadDir(sub, renDir)
	if err != nil {
		rep.fail("read ren/ dir: %v", err)
		return nil
	}

	imageFiles := make(map[int]string)
	decodedCount := 0
	for _, file := range renFiles {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		base := file.Name()
		ext := strings.ToLower(path.Ext(base))
		if !supportedExts[ext] {
			rep.fail("ren/%s: unsupported extension", base)
			continue
		}
		layerID, err := strconv.Atoi(strings.TrimSuffix(base, ext))
		if err != nil || layerID <= 0 {
			rep.fail("ren/%s: filename must be a positive layer_id integer", base)
			continue
		}
		if previous, exists := imageFiles[layerID]; exists {
			rep.fail("ren/%s: duplicate layer_id %d also used by %s", base, layerID, previous)
			continue
		}
		imageFiles[layerID] = base

		raw, err := fs.ReadFile(sub, path.Join(renDir, base))
		if err != nil {
			rep.fail("ren/%s: %v", base, err)
			continue
		}
		if len(raw) > maxFileBytes {
			rep.fail("ren/%s: %d bytes exceeds %d limit", base, len(raw), maxFileBytes)
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
		if err != nil {
			rep.fail("ren/%s: not a decodable image: %v", base, err)
			continue
		}
		if cfg.Width > maxCharLayerSide || cfg.Height > maxCharLayerSide {
			rep.fail("ren/%s: %dx%d exceeds %dpx side limit", base, cfg.Width, cfg.Height, maxCharLayerSide)
		}
		decodedCount++
	}
	if decodedCount == 0 {
		rep.fail("ren/: no layer images found")
	}
	rep.frames = decodedCount
	return imageFiles
}

func validateCharacterRanges(layers []themeasset.CharacterManifest, config *themeasset.CharacterConfig, imageFiles map[int]string, rep *themeReport) {
	if len(config.Sets) > 0 {
		for index, set := range config.Sets {
			validateCharacterRangeSet(layers, set.Ranges, imageFiles, characterSetLabel(index, set.Name), rep)
		}
		return
	}
	validateCharacterRangeSet(layers, config.Ranges, imageFiles, "", rep)
}

func validateCharacterRangeSet(layers []themeasset.CharacterManifest, ranges map[string]themeasset.PartRange, imageFiles map[int]string, setLabel string, rep *themeReport) {
	categories := make([]string, 0, len(ranges))
	for category := range ranges {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	prefix := "config.json:"
	if setLabel != "" {
		prefix += " " + setLabel
	}

	indexOwners := make(map[int]string)
	overlaps := make(map[string]bool)
	activeLayerIndexes := make(map[int]int)
	for _, category := range categories {
		partRange := ranges[category]
		if partRange.First < 0 || partRange.Last < partRange.First || partRange.Last >= len(layers) {
			rep.fail("%s range %q [%d,%d] is outside ren.json indexes [0,%d]", prefix, category, partRange.First, partRange.Last, len(layers)-1)
			continue
		}
		for index := partRange.First; index <= partRange.Last; index++ {
			if owner, exists := indexOwners[index]; exists && owner != category {
				key := owner + "\x00" + category
				if !overlaps[key] {
					rep.fail("%s ranges %q and %q overlap at manifest index %d", prefix, owner, category, index)
					overlaps[key] = true
				}
			} else {
				indexOwners[index] = category
			}

			layer := layers[index]
			if layer.LayerID <= 0 {
				rep.fail("ren.json: active manifest index %d in range %q must have a positive layer_id", index, category)
				continue
			}
			if layer.Width <= 0 || layer.Height <= 0 {
				rep.fail("ren.json: active layer_id %d must have positive width and height", layer.LayerID)
			}
			if previousIndex, exists := activeLayerIndexes[layer.LayerID]; exists && previousIndex != index {
				rep.fail("ren.json: active layer_id %d is duplicated at manifest indexes %d and %d", layer.LayerID, previousIndex, index)
			} else {
				activeLayerIndexes[layer.LayerID] = index
			}
			if _, exists := imageFiles[layer.LayerID]; !exists {
				rep.fail("ren.json: active layer_id %d has no matching ren/<layer_id> image", layer.LayerID)
			}
		}
	}
}

func characterSetLabel(index int, name string) string {
	if name != "" {
		return fmt.Sprintf("set %q", name)
	}
	return fmt.Sprintf("sets[%d]", index)
}
