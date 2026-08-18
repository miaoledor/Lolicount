package theme

import (
	"math/rand"
	"strings"
	"testing"
)

// fakeCharacter builds a Character with the same index-range shape as the
// real 莲 manifest so assembly logic can be tested without the PSD assets.
// Layers 1-70 are real entries; 71-79 are group labels (no parts).
func fakeCharacter() *Character {
	const total = 80
	layers := make([]CharacterLayer, total)
	parts := make(map[int]CharacterPart, 70)
	for i := 0; i < total; i++ {
		layers[i] = CharacterLayer{
			Name:     "L",
			Left:     100 + i,
			Top:      200 + i,
			Width:    50,
			Height:   60,
			LayerID:  1000 + i,
		}
		// Only layers 1-70 have shipped images (group labels 71-79 do not).
		if i >= 1 && i <= 70 {
			parts[1000+i] = CharacterPart{
				Left:   100 + i,
				Top:    200 + i,
				Width:  50,
				Height: 60,
				Data:   "data:image/png;base64,QQ",
			}
		}
	}
	return &Character{Layers: layers, Parts: parts}
}

func TestCharacterAssemblePicksFiveParts(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	p, err := c.Assemble(r)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if len(p.Parts) != 5 {
		t.Fatalf("got %d parts, want 5", len(p.Parts))
	}
}

func TestCharacterAssembleBBoxSane(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(7))
	p, err := c.Assemble(r)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if p.BBox.Width <= 0 || p.BBox.Height <= 0 {
		t.Errorf("bbox must be positive: %+v", p.BBox)
	}
	// bbox must contain every selected part.
	for _, q := range p.Parts {
		if q.Left < p.BBox.Left || q.Top < p.BBox.Top {
			t.Errorf("part outside bbox left/top: %+v vs %+v", q, p.BBox)
		}
		if q.Left+q.Width > p.BBox.Left+p.BBox.Width {
			t.Errorf("part outside bbox right: %+v vs %+v", q, p.BBox)
		}
		if q.Top+q.Height > p.BBox.Top+p.BBox.Height {
			t.Errorf("part outside bbox bottom: %+v vs %+v", q, p.BBox)
		}
	}
}

func TestCharacterAssembleRandomVaries(t *testing.T) {
	c := fakeCharacter()
	r1 := rand.New(rand.NewSource(1))
	r2 := rand.New(rand.NewSource(2))
	p1, _ := c.Assemble(r1)
	p2, _ := c.Assemble(r2)
	// Two different seeds should (almost certainly) differ in at least one
	// chosen layer_id.
	same := true
	for i := range p1.Parts {
		if p1.Parts[i].Data != p2.Parts[i].Data || p1.Parts[i].Left != p2.Parts[i].Left {
			same = false
			break
		}
	}
	if same {
		t.Errorf("two seeds produced identical assembly; expected variation")
	}
}

func TestCharacterAssembleDeterministicWithSeed(t *testing.T) {
	c := fakeCharacter()
	r1 := rand.New(rand.NewSource(42))
	r2 := rand.New(rand.NewSource(42))
	p1, _ := c.Assemble(r1)
	p2, _ := c.Assemble(r2)
	for i := range p1.Parts {
		if p1.Parts[i] != p2.Parts[i] {
			t.Errorf("same seed should produce identical assembly at part %d", i)
		}
	}
}

func TestRenderCharacterProducesSVG(t *testing.T) {
	c := fakeCharacter()
	th := &Theme{Name: "lian-ren", Kind: KindCharacter, Character: c}
	r := rand.New(rand.NewSource(1))
	svg, err := Render(th, RenderParams{Count: 3, Rand: r})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.HasPrefix(svg, "<?xml") {
		t.Errorf("not svg: %q", svg[:16])
	}
	// Five layered <image> elements (the portrait parts).
	if got := strings.Count(svg, "<image"); got != 5 {
		t.Errorf("expected 5 <image> layers, got %d", got)
	}
	// Counter text present (layer 1).
	if !strings.Contains(svg, ">3<") {
		t.Errorf("count text 3 missing")
	}
}

func TestRenderCharacterUnshowFont(t *testing.T) {
	c := fakeCharacter()
	th := &Theme{Name: "lian-ren", Kind: KindCharacter, Character: c}
	r := rand.New(rand.NewSource(1))
	svg, _ := Render(th, RenderParams{Count: 3, Rand: r, UnshowFont: true})
	if strings.Contains(svg, "<text") {
		t.Errorf("unshowf should omit <text>")
	}
	// Portrait still rendered (5 images).
	if got := strings.Count(svg, "<image"); got != 5 {
		t.Errorf("expected 5 <image> layers with unshowf, got %d", got)
	}
}

func TestRenderCharacterNilCharacter(t *testing.T) {
	th := &Theme{Name: "broken", Kind: KindCharacter}
	if _, err := Render(th, RenderParams{}); err == nil {
		t.Error("expected error for character theme with nil Character")
	}
}

// The character render must scale the WHOLE canvas at once (nested SVG
// with the original PSD viewBox), not per-layer. Per-layer int(left*scale)
// truncation shifts parts (e.g. the mouth) relative to each other.
func TestRenderCharacterScalesWholeCanvasNotPerLayer(t *testing.T) {
	c := fakeCharacter()
	th := &Theme{Name: "lian-ren", Kind: KindCharacter, Character: c}
	r := rand.New(rand.NewSource(1))
	svg, _ := Render(th, RenderParams{Count: 1, Rand: r})
	// A nested <svg> with the original PSD canvas viewBox must be present.
	if !strings.Contains(svg, `viewBox="0 0 504 925"`) {
		t.Errorf("expected nested svg with original canvas viewBox 0 0 504 925: %s", sub(svg, "viewBox"))
	}
	// Every portrait part must be placed at its ORIGINAL coordinate inside
	// the nested canvas. fakeCharacter parts have Left >= 101; at display
	// size 400 a per-layer int(left*scale) render would shrink these below
	// 100, so assert each image x stays in the original range.
	for _, line := range strings.Split(svg, "\n") {
		if !strings.Contains(line, "<image") {
			continue
		}
		i := strings.Index(line, "x=\"")
		if i < 0 {
			continue
		}
		rest := line[i+3:]
		j := strings.Index(rest, "\"")
		if j < 0 {
			continue
		}
		x := 0
		for _, ch := range rest[:j] {
			if ch < '0' || ch > '9' {
				break
			}
			x = x*10 + int(ch-'0')
		}
		if x < 100 {
			t.Errorf("image x=%d is per-layer scaled, expected original coord >=100: %s", x, line)
		}
	}
}
