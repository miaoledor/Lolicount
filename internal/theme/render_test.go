package theme

import (
	"strings"
	"testing"
)

// fakeTheme builds a theme with uniform 10x20 digits for deterministic
// geometry assertions.
func fakeTheme(name string) *Theme {
	th := &Theme{Name: name, Chars: make(map[CharName]ThemeChar)}
	for _, d := range digits {
		th.Chars[d] = ThemeChar{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	}
	return th
}

func TestRenderBasic(t *testing.T) {
	th := fakeTheme("fake")
	svg, err := Render(th, RenderParams{Count: 42, Padding: 0, Prefix: -1, Scale: 1, Align: "top", Pixelated: "1", DarkMode: "auto"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.HasPrefix(svg, "<?xml") {
		t.Errorf("not an xml doc: %q", svg[:20])
	}
	// 2 digits * 10 wide = 20, height 20.
	if !strings.Contains(svg, `viewBox="0 0 20 20"`) {
		t.Errorf("viewBox wrong: %s", substring(svg, "viewBox"))
	}
	if !strings.Contains(svg, `image-rendering: pixelated`) {
		t.Errorf("pixelated style missing")
	}
}

func TestRenderPadding(t *testing.T) {
	th := fakeTheme("fake")
	svg, err := Render(th, RenderParams{Count: 7, Padding: 4, Prefix: -1, Scale: 1, Align: "top", Pixelated: "0", DarkMode: "0"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// 4 digits padded (0007), 10 wide each = 40.
	if !strings.Contains(svg, `viewBox="0 0 40 20"`) {
		t.Errorf("padded viewBox wrong: %s", substring(svg, "viewBox"))
	}
	// pixelated=0 should not emit the directive.
	if strings.Contains(svg, "image-rendering: pixelated") {
		t.Errorf("pixelated should be off")
	}
}

func TestRenderPrefix(t *testing.T) {
	th := fakeTheme("fake")
	svg, err := Render(th, RenderParams{Count: 1, Padding: 0, Prefix: 99, Scale: 1, Align: "top", Pixelated: "1", DarkMode: "0"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// "99" + "1" = 3 digits = 30 wide.
	if !strings.Contains(svg, `viewBox="0 0 30 20"`) {
		t.Errorf("prefix viewBox wrong: %s", substring(svg, "viewBox"))
	}
}

func TestRenderFsizeScalesProportionally(t *testing.T) {
	th := fakeTheme("fake")
	// native 10x20; fsize=40 -> height 40, width 10*(40/20)=20.
	svg, err := Render(th, RenderParams{Count: 1, Padding: 0, Prefix: -1, FontSize: 40, Scale: 1, Align: "top", Pixelated: "1", DarkMode: "0"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `width="20" height="40"`) {
		t.Errorf("fsize scaling wrong: %s", substring(svg, "image id=\"g1\""))
	}
}

func TestRenderScaleMultiplies(t *testing.T) {
	th := fakeTheme("fake")
	// native 10x20; scale=0.5 -> 5x10.
	svg, err := Render(th, RenderParams{Count: 1, Padding: 0, Prefix: -1, Scale: 0.5, Align: "top", Pixelated: "1", DarkMode: "0"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `width="5" height="10"`) {
		t.Errorf("scale sizing wrong: %s", substring(svg, "image id=\"g1\""))
	}
}

func TestRenderOffset(t *testing.T) {
	th := fakeTheme("fake")
	// 2 digits, width 10 each, offset 5 -> 10+5+10 = 25.
	svg, err := Render(th, RenderParams{Count: 12, Padding: 0, Prefix: -1, Offset: 5, Scale: 1, Align: "top", Pixelated: "1", DarkMode: "0"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `viewBox="0 0 25 20"`) {
		t.Errorf("offset viewBox wrong: %s", substring(svg, "viewBox"))
	}
}

func TestRenderAlignCenter(t *testing.T) {
	// Mixed heights: "1" short, others tall. Center align gives a y offset
	// on the short glyph.
	th := &Theme{Name: "mix", Chars: map[CharName]ThemeChar{
		"1": {Width: 5, Height: 10, Data: "data:image/gif;base64,QQ"},
		"2": {Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"},
	}}
	svg, err := Render(th, RenderParams{Count: 12, Padding: 0, Prefix: -1, Scale: 1, Align: "center", Pixelated: "1", DarkMode: "0"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// maxH=20, "1" height 10 -> yOffset (20-10)/2 = 5.
	if !strings.Contains(svg, `y="5"`) {
		t.Errorf("center align y offset missing:\n%s", svg)
	}
}

func TestRenderInvalidAlign(t *testing.T) {
	th := fakeTheme("fake")
	if _, err := Render(th, RenderParams{Align: "diagonal"}); err == nil {
		t.Fatal("expected error for invalid align")
	}
}

func TestRenderMissingDigit(t *testing.T) {
	th := &Theme{Name: "broken", Chars: map[CharName]ThemeChar{}}
	if _, err := Render(th, RenderParams{Count: 1, Padding: 0, Prefix: -1}); err == nil {
		t.Fatal("expected error for missing digit")
	}
}

func TestRenderDecorations(t *testing.T) {
	th := fakeTheme("deco")
	th.Chars["_start"] = ThemeChar{Width: 8, Height: 20, Data: "data:image/gif;base64,SS"}
	th.Chars["_end"] = ThemeChar{Width: 8, Height: 20, Data: "data:image/gif;base64,EE"}
	// _start(8) + digit(10) + _end(8) = 26.
	svg, err := Render(th, RenderParams{Count: 5, Padding: 0, Prefix: -1, Scale: 1, Align: "top", Pixelated: "1", DarkMode: "0"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `viewBox="0 0 26 20"`) {
		t.Errorf("decoration viewBox wrong: %s", substring(svg, "viewBox"))
	}
	if !strings.Contains(svg, `id="g_start"`) || !strings.Contains(svg, `id="g_end"`) {
		t.Errorf("decoration defs missing")
	}
}

func TestFnum(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1.5, "1.5"},
		{1.50000, "1.5"},
		{2.123456, "2.12346"},
		{-0, "0"},
	}
	for _, c := range cases {
		if got := fnum(c.in); got != c.want {
			t.Errorf("fnum(%v) = %q want %q", c.in, got, c.want)
		}
	}
}

func substring(s, marker string) string {
	i := strings.Index(s, marker)
	if i < 0 {
		return "(not found)"
	}
	end := i + 60
	if end > len(s) {
		end = len(s)
	}
	return s[i:end]
}
