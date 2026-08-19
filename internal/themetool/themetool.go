// Package themetool contains shared logic for frame-theme file operations
// used by cmd/fix-theme (and potentially other tooling). It is split out
// of package main so the rename planning can be unit-tested.
package themetool

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// SupportedExts mirrors cmd/check-theme: the accepted frame extensions.
var SupportedExts = map[string]bool{
	".gif":  true,
	".png":  true,
	".webp": true,
}

// Frame represents one discovered frame file on disk.
type Frame struct {
	Idx       int    // numeric index parsed from the filename stem (-1 when the stem is non-numeric and the frame is collected by sort order)
	Ext       string // original extension (with leading dot)
	Dir       string // absolute directory of the theme
	OrigName  string // original basename including extension (used as the rename source)
}

// Rename is one file rename operation in a plan.
type Rename struct {
	From string // old basename (e.g. "3.png" or "bs1_um010101_1-1.png")
	To   string // new basename (e.g. "1.png")
}

// IsCharacterTheme reports whether themeDir contains a ren.json manifest,
// marking it as a layered portrait theme whose layer ids must not be
// reindexed.
func IsCharacterTheme(themeDir string) bool {
	_, err := os.Stat(filepath.Join(themeDir, "ren.json"))
	return err == nil
}

// CollectFrames reads themeDir and returns frames sorted by numeric
// index. Non-frame files (meta.json, .DS_Store, dotfiles, non-image
// extensions) are ignored.
//
// Two collection modes:
//   - Numeric mode (default): only files whose stem is a non-negative
//     integer are frames, sorted by that integer. This is the classic
//     "already numeric but possibly non-contiguous" case.
//   - Sort mode: when no file has a numeric stem but the directory
//     contains supported image files, every image is treated as a frame
//     and ordered by filename. This lets fix-theme reindex themes whose
//     frames have arbitrary names (e.g. bs1_um010101_1-1.png) into a
//     contiguous 0..n-1 sequence. Idx is set to the sort position so
//     BuildRenamePlan can detect "already contiguous".
func CollectFrames(themeDir string) ([]Frame, error) {
	files, err := os.ReadDir(themeDir)
	if err != nil {
		return nil, err
	}

	var numeric []Frame
	var images []Frame
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		base := f.Name()
		if strings.HasPrefix(base, ".") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(base))
		if !SupportedExts[ext] {
			continue
		}
		stem := strings.TrimSuffix(base, ext)
		n, err := strconv.Atoi(stem)
		fr := Frame{Ext: ext, Dir: themeDir, OrigName: base}
		if err == nil && n >= 0 {
			fr.Idx = n
			numeric = append(numeric, fr)
		} else {
			images = append(images, fr)
		}
	}

	// Numeric mode: when at least one frame has a numeric stem, use only
	// the numeric frames (classic behavior). Non-numeric image files in
	// the same directory are ignored to avoid mixing conventions.
	if len(numeric) > 0 {
		sort.Slice(numeric, func(i, j int) bool { return numeric[i].Idx < numeric[j].Idx })
		return numeric, nil
	}

	// Sort mode: no numeric stems — treat every image as a frame ordered
	// by filename, assigning synthetic indices 0..n-1.
	sort.Slice(images, func(i, j int) bool { return images[i].OrigName < images[j].OrigName })
	for i := range images {
		images[i].Idx = i
	}
	return images, nil
}

// BuildRenamePlan computes the renames needed so the frame files become
// 0.<ext>, 1.<ext>, ..., n-1.<ext> in numeric order. It returns an empty
// slice when the files are already contiguous from 0.
//
// If two source frames would land on the same target name (a collision),
// the second one is routed through a temp name (__tmp_<i>.<ext>);
// ApplyRenames does a second pass to finalize temp names once the
// conflicting source files have moved away.
func BuildRenamePlan(frames []Frame) []Rename {
	// Check if already contiguous 0..n-1 with matching OrigName.
	alreadyOK := true
	for i, fr := range frames {
		target := fmt.Sprintf("%d%s", i, fr.Ext)
		if fr.Idx != i || fr.OrigName != target {
			alreadyOK = false
			break
		}
	}
	if alreadyOK {
		return nil
	}

	plan := make([]Rename, 0, len(frames))
	used := make(map[string]bool) // target basenames already assigned

	for i, fr := range frames {
		target := fmt.Sprintf("%d%s", i, fr.Ext)
		if target == fr.OrigName {
			used[target] = true
			continue
		}
		if used[target] {
			// Collision: assign a unique temp target; ApplyRenames will
			// do a two-pass move to finalize it.
			target = fmt.Sprintf("__tmp_%d%s", i, fr.Ext)
		}
		used[target] = true
		plan = append(plan, Rename{From: fr.OrigName, To: target})
	}
	return plan
}

// ApplyRenames performs the renames, handling temp-name collisions with
// a second pass. Temp names (__tmp_*) are renamed to their final numeric
// names after the first pass clears the conflicting source files.
func ApplyRenames(themeDir string, plan []Rename) error {
	// Pass 1: rename each source to its (possibly temp) target.
	for _, r := range plan {
		from := filepath.Join(themeDir, r.From)
		to := filepath.Join(themeDir, r.To)
		if err := os.Rename(from, to); err != nil {
			return fmt.Errorf("rename %s -> %s: %w", r.From, r.To, err)
		}
	}
	// Pass 2: collapse any __tmp_* names to their final numeric names.
	for _, r := range plan {
		if !strings.HasPrefix(r.To, "__tmp_") {
			continue
		}
		final := strings.TrimPrefix(r.To, "__tmp_")
		from := filepath.Join(themeDir, r.To)
		to := filepath.Join(themeDir, final)
		if err := os.Rename(from, to); err != nil {
			return fmt.Errorf("finalize %s -> %s: %w", r.To, final, err)
		}
	}
	return nil
}
