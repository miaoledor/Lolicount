package assets

import (
	"io/fs"
	"strings"
	"testing"
)

// The embedded FS must actually contain the assets tree; a misconfigured
// go:embed pattern would compile but yield an empty FS and silently break
// every module that reads themes from it.
func TestFSContainsReadme(t *testing.T) {
	b, err := fs.ReadFile(FS, "README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if !strings.Contains(string(b), "assets") {
		t.Errorf("README content unexpected: %q", b)
	}
}

func TestFSSubtrees(t *testing.T) {
	for _, sub := range []string{"theme", "character", "f-theme", "img"} {
		entries, err := fs.ReadDir(FS, sub)
		if err != nil {
			t.Fatalf("ReadDir %s: %v", sub, err)
		}
		if len(entries) == 0 {
			t.Errorf("%s/ is empty in embed.FS", sub)
		}
	}
}
