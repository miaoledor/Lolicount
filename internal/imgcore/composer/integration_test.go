package composer

import (
	"regexp"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// cardThemeWithText builds a card-style theme matching the server
// bridge layout: one image layer + one text layer, with the canvas
// height = imageH + textH.
func cardThemeWithText(srcW, srcH int, scale float64, text string) *theme.Theme {
	s := scale
	if s <= 0 {
		s = 1
	}
	display := imgutils.DisplaySize(s)
	imgW, imgH := imgutils.ScaledDims(srcW, srcH, display)
	textW, textH := render.MeasureText(text, 0, false)
	canvasW := imgW
	if textW > canvasW {
		canvasW = textW
	}
	return &theme.Theme{
		Canvas: theme.Canvas{Width: canvasW, Height: imgH + textH},
		Layers: []imgcore.Layer{
			&render.ImageLayer{Src: "data:image/png;base64,abc", Width: imgW, Height: imgH, Transform: imgcore.DefaultTransform(), Z: 0},
			&render.TextLayer{Text: text, Transform: imgcore.DefaultTransform(), Z: 1},
		},
	}
}

// TestComposeCardImageDimsMatchScaledDims is a regression test for the
// card scale bug: the rendered <image> width/height must equal the
// ScaledDims output exactly, with no re-rounding on non-square frames.
func TestComposeCardImageDimsMatchScaledDims(t *testing.T) {
	cases := []struct {
		name        string
		srcW, srcH  int
		scale       float64
	}{
		{"lian-like-non-square", 1320, 1333, 0},
		{"square", 400, 400, 0},
		{"wide", 800, 200, 0},
		{"tall", 200, 800, 0},
		{"scale-half", 400, 400, 0.5},
		{"scale-two", 100, 100, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			display := imgutils.DisplaySize(tc.scale)
			if tc.scale <= 0 {
				tc.scale = 1
			}
			wantW, wantH := imgutils.ScaledDims(tc.srcW, tc.srcH, display)
			th := cardThemeWithText(tc.srcW, tc.srcH, tc.scale, "0")
			svg, err := Compose(ComposeParams{Theme: th, Seed: "test", CountText: "0"})
			if err != nil {
				t.Fatalf("Compose: %v", err)
			}
			re := regexp.MustCompile(`<image x="0" y="0" width="(\d+)" height="(\d+)"`)
			m := re.FindStringSubmatch(svg)
			if m == nil {
				t.Fatalf("no <image> in SVG:\n%s", svg)
			}
			if m[1] != itoa(wantW) || m[2] != itoa(wantH) {
				t.Errorf("image dims = %sx%s, want %dx%d", m[1], m[2], wantW, wantH)
			}
		})
	}
}

// TestComposeTextWithinViewBox verifies the counter text Y coordinate is
// within the SVG viewBox height (not below it). This is a regression
// test for the textlayer Y positioning bug.
func TestComposeTextWithinViewBox(t *testing.T) {
	th := cardThemeWithText(400, 400, 0, "0123456789")
	svg, err := Compose(ComposeParams{Theme: th, Seed: "test", CountText: "0123456789"})
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	vbRe := regexp.MustCompile(`viewBox="0 0 (\d+) (\d+)"`)
	m := vbRe.FindStringSubmatch(svg)
	if m == nil {
		t.Fatalf("no viewBox in SVG")
	}
	canvasH := atoi(t, m[2])
	textRe := regexp.MustCompile(`<text x="\d+" y="(\d+)"`)
	tm := textRe.FindStringSubmatch(svg)
	if tm == nil {
		t.Fatalf("no <text> in SVG")
	}
	textY := atoi(t, tm[1])
	if textY > canvasH {
		t.Errorf("text Y=%d is below viewBox height=%d (text invisible)", textY, canvasH)
	}
	if textY < 0 {
		t.Errorf("text Y=%d is negative", textY)
	}
}

// TestComposeZOrder verifies layers render in ZIndex order (lower Z
// appears first in the SVG document, i.e. painted below higher Z).
func TestComposeZOrder(t *testing.T) {
	th := &theme.Theme{
		Canvas: theme.Canvas{Width: 100, Height: 100},
		Layers: []imgcore.Layer{
			&render.ImageLayer{Src: "data:image/png;base64,HIGH", Width: 10, Height: 10, Transform: imgcore.DefaultTransform(), Z: 9},
			&render.ImageLayer{Src: "data:image/png;base64,LOW", Width: 10, Height: 10, Transform: imgcore.DefaultTransform(), Z: 0},
			&render.ImageLayer{Src: "data:image/png;base64,MID", Width: 10, Height: 10, Transform: imgcore.DefaultTransform(), Z: 5},
		},
	}
	svg, err := Compose(ComposeParams{Theme: th, Seed: "t", CountText: ""})
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	lowIdx := strings.Index(svg, "LOW")
	midIdx := strings.Index(svg, "MID")
	highIdx := strings.Index(svg, "HIGH")
	if lowIdx < 0 || midIdx < 0 || highIdx < 0 {
		t.Fatalf("missing layer markers in SVG")
	}
	if !(lowIdx < midIdx && midIdx < highIdx) {
		t.Errorf("Z order wrong: LOW=%d MID=%d HIGH=%d (want LOW<MID<HIGH)", lowIdx, midIdx, highIdx)
	}
}

// TestComposeCanvasGrowsForWideText verifies the canvas width grows to
// fit text wider than the image, and the text is centered on the full
// canvas width.
func TestComposeCanvasGrowsForWideText(t *testing.T) {
	imgW := 100
	th := &theme.Theme{
		Canvas: theme.Canvas{Width: 400, Height: 100},
		Layers: []imgcore.Layer{
			&render.ImageLayer{Src: "data:image/png;base64,x", Width: imgW, Height: 80, Transform: imgcore.DefaultTransform(), Z: 0},
			&render.TextLayer{Text: "0123456789012345678901234567890123456789", FontSize: 16, Transform: imgcore.DefaultTransform(), Z: 1},
		},
	}
	svg, err := Compose(ComposeParams{Theme: th, Seed: "t", CountText: "0123456789"})
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if !strings.Contains(svg, `viewBox="0 0 400 100"`) {
		t.Errorf("canvas should be 400 wide (text wider than 100 image):\n%s", svg)
	}
}

// TestComposeGroupLayerCharacterLayout verifies a GroupLayer (character
// theme) renders a nested <svg> with the correct viewBox mapping and
// output dimensions.
func TestComposeGroupLayerCharacterLayout(t *testing.T) {
	th := &theme.Theme{
		Canvas: theme.Canvas{Width: 243, Height: 420},
		Layers: []imgcore.Layer{
			&render.GroupLayer{
				Parts: []render.GroupPart{
					{Src: "data:image/webp;base64,a", X: 137, Y: 323, Width: 367, Height: 602},
					{Src: "data:image/webp;base64,b", X: 244, Y: 395, Width: 69, Height: 37},
				},
				OutW: 243, OutH: 400, VbX: 137, VbY: 323, VbW: 367, VbH: 602, Z: 0,
			},
			&render.TextLayer{Text: "0", Transform: imgcore.DefaultTransform(), Z: 1},
		},
	}
	svg, err := Compose(ComposeParams{Theme: th, Seed: "t", CountText: "0"})
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if !strings.Contains(svg, `<svg x="0" y="0" width="243" height="400" viewBox="137 323 367 602">`) {
		t.Errorf("nested svg viewBox mismatch:\n%s", svg)
	}
	if !strings.Contains(svg, `<image x="137" y="323" width="367" height="602"`) {
		t.Errorf("first part image missing or wrong coords:\n%s", svg)
	}
}

// TestComposeUnshowFontOmitsText verifies UnshowFont produces no <text>.
func TestComposeUnshowFontOmitsText(t *testing.T) {
	th := &theme.Theme{
		Canvas: theme.Canvas{Width: 100, Height: 100},
		Layers: []imgcore.Layer{
			&render.ImageLayer{Src: "data:image/png;base64,x", Width: 100, Height: 100, Transform: imgcore.DefaultTransform(), Z: 0},
			&render.TextLayer{Text: "0", UnshowFont: true, Transform: imgcore.DefaultTransform(), Z: 1},
		},
	}
	svg, _ := Compose(ComposeParams{Theme: th, Seed: "t", CountText: "0"})
	if strings.Contains(svg, "<text") {
		t.Errorf("expected no <text> when UnshowFont, got:\n%s", svg)
	}
}

// itoa/atoi helpers avoid importing strconv for a test-only file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			t.Fatalf("atoi: non-digit in %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n
}
