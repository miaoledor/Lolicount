package assets

import (
	"io/fs"
	"strings"
	"testing"
)

// The embedded FS must actually contain the assets tree; a misconfigured
// go:embed pattern would compile but yield an empty FS and silently break
// every module that reads themes/backgrounds from it.
func TestFSContainsAssets(t *testing.T) {
	b, err := fs.ReadFile(FS, "assets/README.md")
	if err != nil {
		t.Fatalf("read assets/README.md: %v", err)
	}
	if !strings.Contains(string(b), "assets") {
		t.Errorf("README content unexpected: %q", b)
	}
}

func TestFSSubtree(t *testing.T) {
	sub, err := fs.Sub(FS, "assets")
	if err != nil {
		t.Fatalf("Sub: %v", err)
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("assets subtree is empty")
	}
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}
	for _, want := range []string{"theme", "bg", "img", "README.md"} {
		if !names[want] {
			t.Errorf("missing %q in assets/", want)
		}
	}
}
