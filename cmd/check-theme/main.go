// Package main implements check-theme: a CLI that validates the integrity
// of built-in themes under assets/theme. It is run locally and in CI.
//
// Validation rules:
//
// Frame themes (see AGENTS.md):
//   - directory name: ASCII letters (any case), digits, hyphens; not a reserved word
//   - at least one frame file named <int>.<ext>
//   - frame indices are contiguous starting at 0 (0..n-1)
//   - accepted extensions: .gif .png .webp
//   - each frame decodes to a real image (header read via image.DecodeConfig)
//   - per-file size limit and per-theme max dimensions enforced
//   - optional meta.json, when present, must be valid JSON
//
// Character themes (M9): a directory containing ren.json is a layered
// portrait theme. Validated as:
//   - ren.json is valid JSON and non-empty
//   - ren/ subdir exists with at least one <layer_id>.<ext> image
//   - each image decodes and is within size/dimension limits
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/png"
	"io/fs"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/miaoledor/lolicount/assets"

	_ "golang.org/x/image/webp"
)

const (
	maxFileBytes  = 4 * 1024 * 1024 // 4 MiB per frame
	maxFrameSide  = 2048            // max width or height in pixels
	reservedNames = "demo random"
)

var supportedExts = map[string]bool{
	".gif":  true,
	".png":  true,
	".webp": true,
}

type themeReport struct {
	root   string
	name   string
	frames int
	errors []string
}

func (r *themeReport) fail(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func main() {
	roots := flag.String("roots", "theme,character", "comma-separated embedded subdirectories to scan (relative to assets.FS)")
	flag.Parse()

	var reports []themeReport
	hasError := false
	for _, root := range strings.Split(*roots, ",") {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		sub, err := fs.Sub(assets.FS, root)
		if err != nil {
			die("open embedded %s: %v", root, err)
		}
		entries, err := fs.ReadDir(sub, ".")
		if err != nil {
			die("read %s: %v", root, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			rep := validateTheme(sub, e.Name())
			rep.root = root
			reports = append(reports, rep)
			if len(rep.errors) > 0 {
				hasError = true
			}
		}
	}

	for _, rep := range reports {
		if len(rep.errors) == 0 {
			fmt.Printf("OK   %s/%s (%d frames)\n", rep.root, rep.name, rep.frames)
			continue
		}
		fmt.Printf("FAIL %s/%s (%d frames)\n", rep.root, rep.name, rep.frames)
		for _, msg := range rep.errors {
			fmt.Printf("      - %s\n", msg)
		}
	}

	if hasError {
		os.Exit(1)
	}
}

func validateTheme(sub fs.FS, name string) themeReport {
	rep := themeReport{name: name}

	if !validThemeName(name) {
		rep.fail("invalid directory name %q: use ASCII letters, digits, hyphens; not a reserved word", name)
	}

	themeRoot := name
	files, err := fs.ReadDir(sub, themeRoot)
	if err != nil {
		rep.fail("read theme dir: %v", err)
		return rep
	}

	// M9: a ren.json manifest marks this as a character (layered
	// portrait) theme validated by a separate ruleset.
	if fileExists(sub, path.Join(themeRoot, "ren.json")) {
		validateCharacterTheme(sub, themeRoot, &rep)
		return rep
	}

	// Collect frame indices and detect problems.
	indices := map[int]string{}
	var idxList []int
	seenExt := make(map[string]bool)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		base := f.Name()
		// Skip dotfiles such as .DS_Store: they are OS artifacts, not theme assets.
		if strings.HasPrefix(base, ".") {
			continue
		}
		if base == "meta.json" {
			continue
		}
		ext := strings.ToLower(path.Ext(base))
		if !supportedExts[ext] {
			rep.fail("unsupported file %q: only gif/png/webp allowed", base)
			continue
		}
		stem := strings.TrimSuffix(base, ext)
		n, err := strconv.Atoi(stem)
		if err != nil || n < 0 {
			rep.fail("non-integer frame name %q: expected <int>.<ext>", base)
			continue
		}
		if _, dup := indices[n]; dup {
			rep.fail("duplicate frame index %d", n)
			continue
		}
		indices[n] = base
		idxList = append(idxList, n)
		seenExt[ext] = true

		// Per-file checks against the embedded bytes.
		full := path.Join(themeRoot, base)
		raw, err := fs.ReadFile(sub, full)
		if err != nil {
			rep.fail("read %s: %v", base, err)
			continue
		}
		if len(raw) > maxFileBytes {
			rep.fail("%s: %d bytes exceeds %d limit", base, len(raw), maxFileBytes)
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
		if err != nil {
			rep.fail("%s: not a decodable image: %v", base, err)
			continue
		}
		if cfg.Width > maxFrameSide || cfg.Height > maxFrameSide {
			rep.fail("%s: %dx%d exceeds %dpx side limit", base, cfg.Width, cfg.Height, maxFrameSide)
		}
	}

	// Frames must be contiguous from 0.
	sort.Ints(idxList)
	if len(idxList) == 0 {
		rep.fail("no frame files found")
	} else {
		for i, want := 0, 0; i < len(idxList); i, want = i+1, want+1 {
			if idxList[i] != want {
				rep.fail("frame indices not contiguous from 0: missing %d", want)
				break
			}
		}
		// Mixed extensions within one theme are allowed but warned.
		if len(seenExt) > 1 {
			rep.fail("mixed frame extensions in one theme (%v)", strings.Join(keys(seenExt), ", "))
		}
	}

	// Optional meta.json.
	if metaPath := path.Join(themeRoot, "meta.json"); fileExists(sub, metaPath) {
		raw, err := fs.ReadFile(sub, metaPath)
		if err != nil {
			rep.fail("read meta.json: %v", err)
		} else {
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				rep.fail("meta.json: invalid JSON: %v", err)
			}
		}
	}

	rep.frames = len(idxList)
	return rep
}

// validateCharacterTheme checks a ren.json-based layered portrait theme
// (M9). The manifest must be valid JSON and non-empty; the ren/ subdir
// must contain at least one decodable <layer_id>.<ext> image within the
// size/dimension limits.
func validateCharacterTheme(sub fs.FS, name string, rep *themeReport) {
	manifestPath := path.Join(name, "ren.json")
	raw, err := fs.ReadFile(sub, manifestPath)
	if err != nil {
		rep.fail("read ren.json: %v", err)
		return
	}
	if len(raw) > maxFileBytes {
		rep.fail("ren.json: %d bytes exceeds %d limit", len(raw), maxFileBytes)
	}
	var layers []map[string]any
	if err := json.Unmarshal(raw, &layers); err != nil {
		rep.fail("ren.json: invalid JSON: %v", err)
		return
	}
	if len(layers) == 0 {
		rep.fail("ren.json: empty manifest")
		return
	}

	renDir := path.Join(name, "ren")
	renFiles, err := fs.ReadDir(sub, renDir)
	if err != nil {
		rep.fail("read ren/ dir: %v", err)
		return
	}
	imgCount := 0
	for _, f := range renFiles {
		if f.IsDir() {
			continue
		}
		base := f.Name()
		if strings.HasPrefix(base, ".") {
			continue
		}
		ext := strings.ToLower(path.Ext(base))
		if !supportedExts[ext] {
			rep.fail("ren/%s: unsupported extension", base)
			continue
		}
		stem := strings.TrimSuffix(base, ext)
		if _, err := strconv.Atoi(stem); err != nil {
			rep.fail("ren/%s: filename must be a layer_id integer", base)
			continue
		}
		full := path.Join(renDir, base)
		raw, err := fs.ReadFile(sub, full)
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
		if cfg.Width > maxFrameSide || cfg.Height > maxFrameSide {
			rep.fail("ren/%s: %dx%d exceeds %dpx side limit", base, cfg.Width, cfg.Height, maxFrameSide)
		}
		imgCount++
	}
	if imgCount == 0 {
		rep.fail("ren/: no layer images found")
	}
	rep.frames = imgCount
}

func validThemeName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	for _, res := range strings.Fields(reservedNames) {
		if name == res {
			return false
		}
	}
	return true
}

func fileExists(sub fs.FS, p string) bool {
	_, err := fs.Stat(sub, p)
	return err == nil
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "check-theme: "+format+"\n", args...)
	os.Exit(2)
}
