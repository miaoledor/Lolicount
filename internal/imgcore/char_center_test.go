package imgcore

import (
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
)

// Character theme with wide text should also center on full canvas.
func TestCharacterTextCenteredOnFullCanvas(t *testing.T) {
	ch := fakeCharacter()
	portrait, _ := ch.Assemble(nil)
	svg, err := Render(RenderParams{
		ThemeKind: KindCharacter,
		Portrait:  portrait,
		Text:      "12345678901234567890",
		FontSize:  100,
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `text-anchor="middle"`) {
		t.Errorf("expected middle anchor")
	}
	// Character canvas 504x925 -> display 400 longest edge ->
	// imgW = 400*504/925 = 217, imgH = 400.
	// textW = 20*100*0.6 + 60 = 1260. canvasWidth = 1260.
	// text x = 1260/2 = 630.
	if !strings.Contains(svg, `x="630"`) {
		t.Errorf("character text should center on full canvas (x=630): %s", sub(svg, "<text"))
	}
}

// Character theme: the <g transform> wrapper must correctly shift the
// nested <svg> when text widens the canvas.
func TestCharacterGroupTransformPresent(t *testing.T) {
	ch := fakeCharacter()
	portrait, _ := ch.Assemble(nil)
	svg, _ := Render(RenderParams{
		ThemeKind: KindCharacter,
		Portrait:  portrait,
		Text:      "12345678901234567890",
		FontSize:  100,
	})
	// imgX = (1260 - 217) / 2 = 521. A <g transform="translate(521,0)">
	// should wrap the nested <svg>.
	if !strings.Contains(svg, `transform="translate(`) {
		t.Errorf("expected <g transform> wrapper for character: %s", sub(svg, "<g"))
	}
	// 5 portrait images still present.
	if got := strings.Count(svg, "<image"); got != 5 {
		t.Errorf("expected 5 images, got %d", got)
	}
}
