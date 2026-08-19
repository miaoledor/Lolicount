// Command fix-theme scans built-in frame themes under assets/theme on
// disk and auto-corrects frame file indices so they are contiguous from
// 0 (0, 1, 2, ..., size-1). This is the M11 task: contributors sometimes
// drop in frames with gaps (e.g. 1,3,5.png) or out-of-order names; this
// tool renames them in place so check-theme and the runtime registry see
// a dense, contiguous sequence.
//
// It operates on the real filesystem (not embed.FS, which is read-only).
// Run it from the repo root before committing new themes. Use --dry-run
// to preview what would change without touching files.
//
// Character themes (assets/character, ren.json + ren/) are skipped —
// their layer ids are not a frame sequence and must not be reindexed.
//
// Usage:
//
//	go run ./cmd/fix-theme              # fix in place
//	go run ./cmd/fix-theme --dry-run    # preview only
//	go run ./cmd/fix-theme --root assets/theme
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/miaoledor/lolicount/internal/themetool"
)

func main() {
	root := flag.String("root", "assets/theme", "directory containing frame themes on disk")
	dryRun := flag.Bool("dry-run", false, "preview renames without writing")
	flag.Parse()

	abs, err := filepath.Abs(*root)
	if err != nil {
		die("resolve root %q: %v", *root, err)
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		die("read %s: %v", abs, err)
	}

	changed := 0
	skipped := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		themeDir := filepath.Join(abs, name)

		// Skip character themes (ren.json marker) — layer ids are not
		// a frame sequence and must not be reindexed.
		if themetool.IsCharacterTheme(themeDir) {
			skipped++
			fmt.Printf("SKIP  %s (character theme)\n", name)
			continue
		}

		frames, err := themetool.CollectFrames(themeDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL  %s: %v\n", name, err)
			continue
		}
		if len(frames) == 0 {
			fmt.Printf("SKIP  %s (no frames)\n", name)
			skipped++
			continue
		}

		plan := themetool.BuildRenamePlan(frames)
		if len(plan) == 0 {
			fmt.Printf("OK    %s (%d frames, already contiguous)\n", name, len(frames))
			continue
		}

		if *dryRun {
			fmt.Printf("DRY   %s (%d frames, %d renames):\n", name, len(frames), len(plan))
		} else {
			fmt.Printf("FIX   %s (%d frames, %d renames):\n", name, len(frames), len(plan))
		}
		for _, r := range plan {
			fmt.Printf("        %s -> %s\n", r.From, r.To)
		}
		if !*dryRun {
			if err := themetool.ApplyRenames(themeDir, plan); err != nil {
				fmt.Fprintf(os.Stderr, "      apply failed: %v\n", err)
				continue
			}
		}
		changed++
	}

	if *dryRun {
		fmt.Printf("\n(dry-run) %d theme(s) would change, %d skipped\n", changed, skipped)
	} else {
		fmt.Printf("\n%d theme(s) fixed, %d skipped\n", changed, skipped)
	}
	if changed > 0 && *dryRun {
		os.Exit(1) // signal "would change" so CI can gate on it
	}
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fix-theme: "+format+"\n", args...)
	os.Exit(2)
}
