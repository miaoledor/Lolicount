package themetool

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// touchFrames creates a temp theme dir with the given frame filenames
// (relative to the dir) and returns its path.
func touchFrames(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		// Create an empty file; content does not matter for rename logic.
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatalf("create %s: %v", n, err)
		}
	}
	return dir
}

// listNames returns the sorted filenames in dir (dotfiles filtered).
func listNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out
}

func TestCollectFrames_SortsByIndex(t *testing.T) {
	// Files given in a non-numeric sort order; must collect sorted.
	dir := touchFrames(t, "2.png", "10.png", "1.png", "0.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	want := []int{0, 1, 2, 10}
	for i, fr := range frames {
		if fr.Idx != want[i] {
			t.Errorf("frame[%d].Idx = %d, want %d", i, fr.Idx, want[i])
		}
	}
}

func TestCollectFrames_IgnoresNonFrames(t *testing.T) {
	dir := touchFrames(t, "0.png", "1.png", "meta.json", ".DS_Store", "readme.txt", "_start.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	if len(frames) != 2 {
		t.Fatalf("got %d frames, want 2 (only 0.png and 1.png)", len(frames))
	}
}

func TestBuildRenamePlan_AlreadyContiguous(t *testing.T) {
	frames := []Frame{{Idx: 0, Ext: ".png", OrigName: "0.png"}, {Idx: 1, Ext: ".png", OrigName: "1.png"}, {Idx: 2, Ext: ".png", OrigName: "2.png"}}
	plan := BuildRenamePlan(frames)
	if plan != nil {
		t.Errorf("expected nil plan for contiguous frames, got %v", plan)
	}
}

func TestBuildRenamePlan_Gaps(t *testing.T) {
	// Indices 1,3,5 with gaps -> should map to 0,1,2.
	frames := []Frame{{Idx: 1, Ext: ".png", OrigName: "1.png"}, {Idx: 3, Ext: ".png", OrigName: "3.png"}, {Idx: 5, Ext: ".png", OrigName: "5.png"}}
	plan := BuildRenamePlan(frames)
	if len(plan) != 3 {
		t.Fatalf("got %d renames, want 3", len(plan))
	}
	wantFrom := []string{"1.png", "3.png", "5.png"}
	wantTo := []string{"0.png", "1.png", "2.png"}
	for i, r := range plan {
		if r.From != wantFrom[i] || r.To != wantTo[i] {
			t.Errorf("rename[%d] = %s->%s, want %s->%s", i, r.From, r.To, wantFrom[i], wantTo[i])
		}
	}
}

func TestApplyRenames_Gaps(t *testing.T) {
	dir := touchFrames(t, "1.png", "3.png", "5.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	plan := BuildRenamePlan(frames)
	if err := ApplyRenames(dir, plan); err != nil {
		t.Fatalf("ApplyRenames: %v", err)
	}
	got := listNames(t, dir)
	want := []string{"0.png", "1.png", "2.png"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestApplyRenames_Swap(t *testing.T) {
	// 0 and 1 swapped: idx 1 should become 0, idx 0 should become 1.
	// This is a collision case (target 0.png is taken by source 0.png
	// which needs to move to 1.png). Temp names resolve it.
	dir := touchFrames(t, "1.png", "0.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	plan := BuildRenamePlan(frames)
	if err := ApplyRenames(dir, plan); err != nil {
		t.Fatalf("ApplyRenames: %v", err)
	}
	got := listNames(t, dir)
	want := []string{"0.png", "1.png"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestApplyRenames_NoChangeWhenContiguous(t *testing.T) {
	dir := touchFrames(t, "0.png", "1.png", "2.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	plan := BuildRenamePlan(frames)
	if plan != nil {
		t.Fatalf("expected no plan, got %v", plan)
	}
	// Files must be untouched.
	got := listNames(t, dir)
	want := []string{"0.png", "1.png", "2.png"}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestApplyRenames_MixedExtensions(t *testing.T) {
	// Gaps with mixed extensions: each frame keeps its own ext.
	dir := touchFrames(t, "0.gif", "5.png", "10.webp")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	plan := BuildRenamePlan(frames)
	if err := ApplyRenames(dir, plan); err != nil {
		t.Fatalf("ApplyRenames: %v", err)
	}
	got := listNames(t, dir)
	want := []string{"0.gif", "1.png", "2.webp"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestIsCharacterTheme(t *testing.T) {
	dir := touchFrames(t, "ren.json", "0.png")
	if !IsCharacterTheme(dir) {
		t.Error("expected character theme (has ren.json)")
	}
	dir2 := touchFrames(t, "0.png", "1.png")
	if IsCharacterTheme(dir2) {
		t.Error("expected non-character theme (no ren.json)")
	}
}

func TestApplyRenames_LargeGap(t *testing.T) {
	// Extreme gap: 0 and 100 -> should become 0 and 1.
	dir := touchFrames(t, "0.png", "100.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	plan := BuildRenamePlan(frames)
	if err := ApplyRenames(dir, plan); err != nil {
		t.Fatalf("ApplyRenames: %v", err)
	}
	got := listNames(t, dir)
	want := []string{"0.png", "1.png"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectFrames_NonNumericSortMode(t *testing.T) {
	// No numeric stems: every image becomes a frame ordered by name,
	// with synthetic indices 0..n-1.
	dir := touchFrames(t, "bs1_um030301_1-1.png", "bs1_um010101_1-1.png", "bs1_um020101_1-1.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	if len(frames) != 3 {
		t.Fatalf("got %d frames, want 3", len(frames))
	}
	want := []string{"bs1_um010101_1-1.png", "bs1_um020101_1-1.png", "bs1_um030301_1-1.png"}
	for i, fr := range frames {
		if fr.Idx != i {
			t.Errorf("frame[%d].Idx = %d, want %d", i, fr.Idx, i)
		}
		if fr.OrigName != want[i] {
			t.Errorf("frame[%d].OrigName = %s, want %s", i, fr.OrigName, want[i])
		}
	}
}

func TestCollectFrames_NumericModeIgnoresNonNumeric(t *testing.T) {
	// When numeric stems exist, non-numeric image files are ignored
	// (do not mix conventions).
	dir := touchFrames(t, "0.png", "2.png", "bs1_um010101_1-1.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	if len(frames) != 2 {
		t.Fatalf("got %d frames, want 2 (numeric only)", len(frames))
	}
}

func TestApplyRenames_NonNumericNames(t *testing.T) {
	// Arbitrary filenames -> renamed to 0..n-1 by sort order.
	dir := touchFrames(t, "bs1_um030301_1-1.png", "bs1_um010101_1-1.png", "bs1_um020101_1-1.png")
	frames, err := CollectFrames(dir)
	if err != nil {
		t.Fatalf("CollectFrames: %v", err)
	}
	plan := BuildRenamePlan(frames)
	if err := ApplyRenames(dir, plan); err != nil {
		t.Fatalf("ApplyRenames: %v", err)
	}
	got := listNames(t, dir)
	want := []string{"0.png", "1.png", "2.png"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}
