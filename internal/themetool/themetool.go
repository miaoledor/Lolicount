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
	Idx int    // numeric index parsed from the filename stem
	Ext string // original extension (with leading dot)
	Dir string // absolute directory of the theme
}

// Rename is one file rename operation in a plan.
type Rename struct {
	From string // old basename (e.g. "3.png")
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
func CollectFrames(themeDir string) ([]Frame, error) {
	files, err := os.ReadDir(themeDir)
	if err != nil {
		return nil, err
	}
	var frames []Frame
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
		if err != nil || n < 0 {
			continue // non-integer stem, not a frame
		}
		frames = append(frames, Frame{Idx: n, Ext: ext, Dir: themeDir})
	}
	sort.Slice(frames, func(i, j int) bool { return frames[i].Idx < frames[j].Idx })
	return frames, nil
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
	// Check if already contiguous 0..n-1.
	alreadyOK := true
	for i, fr := range frames {
		if fr.Idx != i {
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
		old := fmt.Sprintf("%d%s", fr.Idx, fr.Ext)
		if target == old {
			used[target] = true
			continue
		}
		if used[target] {
			// Collision: assign a unique temp target; ApplyRenames will
			// do a two-pass move to finalize it.
			target = fmt.Sprintf("__tmp_%d%s", i, fr.Ext)
		}
		used[target] = true
		plan = append(plan, Rename{From: old, To: target})
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
