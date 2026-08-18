// Package main implements check-theme: a CLI that validates the integrity
// of built-in themes under assets/theme. It is run locally and in CI.
//
// Validation rules (frame-based theme model, see AGENTS.md):
//   - directory name: lowercase letters, digits, hyphens; not a reserved word
//   - at least one frame file named <int>.<ext>
//   - frame indices are contiguous starting at 0 (0..n-1)
//   - accepted extensions: .gif .png .webp
//   - each frame decodes to a real image (header read via image.DecodeConfig)
//   - per-file size limit and per-theme max dimensions enforced
//   - optional meta.json, when present, must be valid JSON
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
	maxFileBytes  = 2 * 1024 * 1024 // 2 MiB per frame
	maxFrameSide  = 1024            // max width or height in pixels
	reservedNames = "demo random"
)

var supportedExts = map[string]bool{
	".gif":  true,
	".png":  true,
	".webp": true,
}

type themeReport struct {
	name   string
	frames int
	errors []string
}

func (r *themeReport) fail(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func main() {
	root := flag.String("root", "theme", "embedded subdirectory to scan (relative to assets.FS)")
	flag.Parse()

	sub, err := fs.Sub(assets.FS, *root)
	if err != nil {
		die("open embedded %s: %v", *root, err)
	}

	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		die("read %s: %v", *root, err)
	}

	var reports []themeReport
	hasError := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rep := validateTheme(sub, e.Name())
		reports = append(reports, rep)
		if len(rep.errors) > 0 {
			hasError = true
		}
	}

	for _, rep := range reports {
		if len(rep.errors) == 0 {
			fmt.Printf("OK   %s (%d frames)\n", rep.name, rep.frames)
			continue
		}
		fmt.Printf("FAIL %s (%d frames)\n", rep.name, rep.frames)
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
		rep.fail("invalid directory name %q: use lowercase letters, digits, hyphens; not a reserved word", name)
	}

	themeRoot := name
	files, err := fs.ReadDir(sub, themeRoot)
	if err != nil {
		rep.fail("read theme dir: %v", err)
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

func validThemeName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
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
