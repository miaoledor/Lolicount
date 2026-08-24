package characterthemedrawer

import (
	"math/rand"
	"strings"
	"testing"
)

// Assemble on a nil Character returns an error.
func TestAssembleNilCharacter(t *testing.T) {
	var c *Character
	_, err := c.Assemble(rand.New(rand.NewSource(1)))
	if err == nil {
		t.Error("nil character should error")
	}
}

// Assemble on an empty Character returns an error.
func TestAssembleEmptyCharacter(t *testing.T) {
	c := &Character{}
	_, err := c.Assemble(rand.New(rand.NewSource(1)))
	if err == nil {
		t.Error("empty character should error")
	}
}

// Assemble with a missing part (layer referenced by manifest but not
// decoded) returns an error.
func TestAssembleMissingPart(t *testing.T) {
	c := fakeCharacter()
	// Delete a part that will be picked (layer_id in the brow range
	// 1001..1018). Remove one so assembly may hit it.
	delete(c.Parts, 1001)
	// Try multiple seeds to ensure we hit the missing part eventually.
	missingHit := false
	for seed := int64(0); seed < 200; seed++ {
		_, err := c.Assemble(rand.New(rand.NewSource(seed)))
		if err != nil {
			missingHit = true
			break
		}
	}
	if !missingHit {
		t.Error("expected an error when a selected part is missing (tried 200 seeds)")
	}
}

// Draw produces a nested <svg> with the character canvas viewBox.
func TestDrawProducesNestedSVG(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	p, _ := c.Assemble(r)
	layer := Draw(p, 0)
	if !strings.Contains(layer.Fragment, "<svg") {
		t.Error("Draw should produce a nested <svg>")
	}
	if !strings.Contains(layer.Fragment, `viewBox="0 0 504 925"`) {
		t.Errorf("nested svg should have character canvas viewBox: %s", layer.Fragment)
	}
}

// Draw with scale produces the right output dimensions.
func TestDrawScaledDims(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	p, _ := c.Assemble(r)
	layer := Draw(p, 2) // 400*2=800 longest edge; 504x925 → ratio 800/925
	// longest edge = 925, display=800, ratio=800/925
	// w = 504*800/925 = 436, h = 925*800/925 = 800
	if layer.Width <= 0 || layer.Height <= 0 {
		t.Errorf("scaled dims should be positive: %dx%d", layer.Width, layer.Height)
	}
	if layer.Height != 800 {
		t.Errorf("scale=2 height: got %d, want 800", layer.Height)
	}
}

// Draw with negative scale uses default.
func TestDrawNegativeScale(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	p, _ := c.Assemble(r)
	layer := Draw(p, -1)
	// default display 400, longest=925, h=400, w=504*400/925=217
	if layer.Height != 400 {
		t.Errorf("negative scale should use default height 400: got %d", layer.Height)
	}
}

// BBox is the union of all part bounding boxes.
func TestAssembleBBoxUnion(t *testing.T) {
	c := fakeCharacter()
	r := rand.New(rand.NewSource(42))
	p, err := c.Assemble(r)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	minLeft := p.Parts[0].Left
	minTop := p.Parts[0].Top
	maxRight := p.Parts[0].Left + p.Parts[0].Width
	maxBottom := p.Parts[0].Top + p.Parts[0].Height
	for _, q := range p.Parts[1:] {
		if q.Left < minLeft {
			minLeft = q.Left
		}
		if q.Top < minTop {
			minTop = q.Top
		}
		if q.Left+q.Width > maxRight {
			maxRight = q.Left + q.Width
		}
		if q.Top+q.Height > maxBottom {
			maxBottom = q.Top + q.Height
		}
	}
	if p.BBox.Left != minLeft {
		t.Errorf("bbox left %d != min %d", p.BBox.Left, minLeft)
	}
	if p.BBox.Top != minTop {
		t.Errorf("bbox top %d != min %d", p.BBox.Top, minTop)
	}
	if p.BBox.Width != maxRight-minLeft {
		t.Errorf("bbox width %d != %d", p.BBox.Width, maxRight-minLeft)
	}
	if p.BBox.Height != maxBottom-minTop {
		t.Errorf("bbox height %d != %d", p.BBox.Height, maxBottom-minTop)
	}
}

// Assemble is deterministic for the same seed.
func TestAssembleDeterministic(t *testing.T) {
	c := fakeCharacter()
	r1 := rand.New(rand.NewSource(99))
	r2 := rand.New(rand.NewSource(99))
	p1, _ := c.Assemble(r1)
	p2, _ := c.Assemble(r2)
	for i := range p1.Parts {
		if p1.Parts[i] != p2.Parts[i] {
			t.Errorf("part %d differs: %+v vs %+v", i, p1.Parts[i], p2.Parts[i])
		}
	}
}

// Registry Get on unknown name returns false.
func TestRegistryGetUnknown(t *testing.T) {
	reg := &builtinCharRegistry{themes: map[string]*Character{"a": {}}}
	if _, ok := reg.Get("b"); ok {
		t.Error("Get(unknown) should return false")
	}
}

// Registry List returns sorted names.
func TestRegistryListSorted(t *testing.T) {
	reg := &builtinCharRegistry{themes: map[string]*Character{
		"zeta":  {},
		"alpha": {},
		"mid":   {},
	}}
	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("got %d, want 3", len(list))
	}
	if list[0] != "alpha" || list[1] != "mid" || list[2] != "zeta" {
		t.Errorf("List not sorted: %v", list)
	}
}
