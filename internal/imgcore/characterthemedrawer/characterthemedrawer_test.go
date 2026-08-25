package characterthemedrawer

import (
	"math/rand"
	"strings"
	"testing"
)

// fakeCharacter builds a Character with the same index-range shape as the
// real 莲 manifest so assembly logic can be tested without the PSD assets.
func fakeCharacter() *Character {
	const total = 80
	layers := make([]CharacterLayer, total)
	parts := make(map[int]CharacterPart, 70)
	for i := 0; i < total; i++ {
		layers[i] = CharacterLayer{
			Name:    "L",
			Left:    100 + i,
			Top:     200 + i,
			Width:   50,
			Height:  60,
			LayerID: 1000 + i,
		}
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

func TestAssemblePicksFiveParts(t *testing.T) {
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

func TestAssembleBBoxSane(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(7))
	p, err := c.Assemble(r)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if p.BBox.Width <= 0 || p.BBox.Height <= 0 {
		t.Errorf("bbox must be positive: %+v", p.BBox)
	}
	for _, q := range p.Parts {
		if q.Left < p.BBox.Left || q.Top < p.BBox.Top {
			t.Errorf("part outside bbox left/top: %+v vs %+v", q, p.BBox)
		}
	}
}

func TestAssembleRandomVaries(t *testing.T) {
	c := fakeCharacter()
	p1, _ := c.Assemble(rand.New(rand.NewSource(1)))
	p2, _ := c.Assemble(rand.New(rand.NewSource(2)))
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

func TestAssembleDeterministicWithSeed(t *testing.T) {
	c := fakeCharacter()
	p1, _ := c.Assemble(rand.New(rand.NewSource(42)))
	p2, _ := c.Assemble(rand.New(rand.NewSource(42)))
	for i := range p1.Parts {
		if p1.Parts[i] != p2.Parts[i] {
			t.Errorf("same seed should produce identical assembly at part %d", i)
		}
	}
}

// Draw renders five layered <image> elements in a nested <svg> with the
// original PSD canvas viewBox.
func TestDrawProducesFiveImages(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	p, _ := c.Assemble(r)
	layer := Draw(p, 0)
	if got := strings.Count(layer.Fragment, "<image"); got != 5 {
		t.Errorf("expected 5 <image> layers, got %d", got)
	}
	if !strings.Contains(layer.Fragment, `viewBox="0 0 504 925"`) {
		t.Errorf("expected nested svg with canvas viewBox 0 0 504 925: %s", layer.Fragment)
	}
}

// Draw must scale the whole canvas at once (nested SVG with original PSD
// viewBox), not per-layer. Per-layer truncation shifts parts.
func TestDrawScalesWholeCanvasNotPerLayer(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	p, _ := c.Assemble(r)
	layer := Draw(p, 0)
	for _, line := range strings.Split(layer.Fragment, "\n") {
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


// TestHinataBodyFaceAlignment verifies that for the hinata theme the
// assembled body (lass) layer is large enough to contain the face layer
// on the PSD canvas. This guards the fix that stopped downscaling layer
// PNGs: a downscaled body would no longer span the face coordinates.
func TestHinataBodyFaceAlignment(t *testing.T) {
	reg, _ := NewBuiltinRegistry()
	ch, ok := reg.Get("hinata")
	if !ok {
		t.Skip("hinata theme not loaded; assets unavailable")
	}
	cfg := ch.config()
	lassRange, ok := cfg.Ranges["lass"]
	if !ok {
		t.Fatal("hinata config missing lass range")
	}
	faceRange, ok := cfg.Ranges["face"]
	if !ok {
		t.Fatal("hinata config missing face range")
	}
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 8; i++ {
		_, err := ch.Assemble(r)
		if err != nil {
			t.Fatalf("assemble %d: %v", i, err)
		}
		body := ch.Parts[ch.Layers[lassRange.First+r.Intn(lassRange.Last-lassRange.First+1)].LayerID]
		face := ch.Parts[ch.Layers[faceRange.First+r.Intn(faceRange.Last-faceRange.First+1)].LayerID]
		if body.Width < 500 {
			t.Errorf("trial %d: body width %d too small (downscaled?)", i, body.Width)
		}
		bodyRight := body.Left + body.Width
		bodyBottom := body.Top + body.Height
		if face.Left < body.Left || face.Left > bodyRight ||
			face.Top < body.Top || face.Top > bodyBottom {
			t.Errorf("trial %d: face (%d,%d) outside body rect (%d,%d,%d,%d)",
				i, face.Left, face.Top, body.Left, body.Top, bodyRight, bodyBottom)
		}
	}
}
