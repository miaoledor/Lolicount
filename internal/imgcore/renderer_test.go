package imgcore

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/cardthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/characterthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/fdrawer"
)

// fakeFrame builds a cardthemedrawer.Frame with given dims.
func fakeFrame(w, h int) cardthemedrawer.Frame {
	return cardthemedrawer.Frame{Width: w, Height: h, Data: "data:image/gif;base64,QQ"}
}

// fakeCardRender builds RenderParams for a 10x20 card frame with the
// given count text.
func fakeCardRender(text string, opts ...func(*RenderParams)) (string, error) {
	p := RenderParams{
		ThemeKind: KindFrame,
		Frame:     fakeFrame(10, 20),
		Text:      text,
	}
	for _, opt := range opts {
		opt(&p)
	}
	return Render(p)
}

// sub returns a short snippet around the first occurrence of marker.
func sub(s, marker string) string {
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

// M5.6: without Scale the frame is scaled to the uniform base size
// (longest edge = 400). A 10x20 frame -> 200x400.
func TestRenderUniformDisplaySize(t *testing.T) {
	svg, err := fakeCardRender("5")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.HasPrefix(svg, "<?xml") {
		t.Errorf("not xml: %q", svg[:16])
	}
	if !strings.Contains(svg, `width="200" height="400"`) {
		t.Errorf("image not scaled to uniform size: %s", sub(svg, "image"))
	}
}

// M5.6: the count text sits BELOW the image, centered.
func TestRenderTextBelowImage(t *testing.T) {
	svg, _ := fakeCardRender("5", func(p *RenderParams) { p.FontSize = 16 })
	// canvas height = imgH(400) + fontSize(16) + gap(4) = 420.
	if !strings.Contains(svg, `viewBox="0 0 200 420"`) {
		t.Errorf("viewBox should be image+text height: %s", sub(svg, "viewBox"))
	}
	if !strings.Contains(svg, `y="416"`) {
		t.Errorf("text should be below image (y=416): %s", sub(svg, "y="))
	}
	if !strings.Contains(svg, `text-anchor="middle"`) {
		t.Errorf("text should be centered")
	}
	if !strings.Contains(svg, `>5<`) {
		t.Errorf("count text missing:\n%s", svg)
	}
	// Layer order: image before text.
	if strings.Index(svg, "<image") > strings.Index(svg, "<text") {
		t.Errorf("image must precede text")
	}
}

func TestRenderScaleMultipliesImageSize(t *testing.T) {
	svg, _ := fakeCardRender("1", func(p *RenderParams) { p.Scale = 2 })
	if !strings.Contains(svg, `width="400" height="800"`) {
		t.Errorf("scale=2 should double image display size: %s", sub(svg, "image"))
	}
}

func TestRenderAspectRatioPreserved(t *testing.T) {
	svg, _ := Render(RenderParams{
		ThemeKind: KindFrame,
		Frame:     cardthemedrawer.Frame{Width: 2000, Height: 1000, Data: "data:image/gif;base64,QQ"},
		Text:      "1",
	})
	if !strings.Contains(svg, `width="400" height="200"`) {
		t.Errorf("wide frame aspect not preserved: %s", sub(svg, "image"))
	}
}

func TestRenderUnshowFont(t *testing.T) {
	svg, _ := fakeCardRender("42", func(p *RenderParams) { p.UnshowFont = true })
	if strings.Contains(svg, "<text") {
		t.Errorf("unshowf=true should omit <text>: %s", svg)
	}
	if !strings.Contains(svg, `viewBox="0 0 200 400"`) {
		t.Errorf("unshowf canvas should be image-only: %s", sub(svg, "viewBox"))
	}
}

func TestRenderFontSizeFromParam(t *testing.T) {
	svg, _ := fakeCardRender("1", func(p *RenderParams) { p.FontSize = 40 })
	if !strings.Contains(svg, `font-size="40"`) {
		t.Errorf("fsize=40 should set font-size=40: %s", sub(svg, "font-size"))
	}
}

func TestRenderDefaultFontSize(t *testing.T) {
	svg, _ := fakeCardRender("1")
	if !strings.Contains(svg, `font-size="16"`) {
		t.Errorf("default font-size should be 16: %s", sub(svg, "font-size"))
	}
}

func TestRenderWideTextWidensViewBox(t *testing.T) {
	svg, _ := fakeCardRender("123456")
	if !strings.Contains(svg, `viewBox="0 0 200 420"`) {
		t.Errorf("wide text viewBox wrong: %s", sub(svg, "viewBox"))
	}
}

func TestRenderNilCharacterPortrait(t *testing.T) {
	_, err := Render(RenderParams{ThemeKind: KindCharacter})
	if err == nil {
		t.Error("expected error for nil character portrait")
	}
}

func TestRenderFontStyleApplied(t *testing.T) {
	svg, _ := fakeCardRender("5", func(p *RenderParams) {
		p.FontSize = 16
		p.FontStyle = fdrawer.FontStyle{Family: "serif", Color: "#e91e63", Weight: "bold"}
	})
	if !strings.Contains(svg, `font-family="serif"`) {
		t.Errorf("FontStyle.Family not applied: %s", sub(svg, "font-family"))
	}
	if !strings.Contains(svg, `fill="#e91e63"`) {
		t.Errorf("FontStyle.Color not applied: %s", sub(svg, "fill"))
	}
}

func TestRenderFontStyleDefaults(t *testing.T) {
	svg, _ := fakeCardRender("5")
	if !strings.Contains(svg, `font-family="monospace"`) {
		t.Errorf("default family not applied: %s", sub(svg, "font-family"))
	}
	if !strings.Contains(svg, `fill="#333"`) {
		t.Errorf("default color not applied: %s", sub(svg, "fill"))
	}
}

func TestRenderPositionDefault(t *testing.T) {
	svg, _ := fakeCardRender("5", func(p *RenderParams) { p.FontSize = 16 })
	if !strings.Contains(svg, `text-anchor="middle"`) {
		t.Errorf("default should be centered: %s", sub(svg, "text-anchor"))
	}
	if !strings.Contains(svg, `y="416"`) {
		t.Errorf("default text y should be below image: %s", sub(svg, "y="))
	}
}

func TestRenderPositionPixel(t *testing.T) {
	svg, _ := fakeCardRender("5", func(p *RenderParams) {
		p.FontSize = 16
		p.Position = fdrawer.TextPos{X: 50, Y: 100}
	})
	if !strings.Contains(svg, `text-anchor="start"`) {
		t.Errorf("pixel mode should use start anchor: %s", sub(svg, "text-anchor"))
	}
	if !strings.Contains(svg, `x="50"`) {
		t.Errorf("pixel x should be 50: %s", sub(svg, "x="))
	}
	if !strings.Contains(svg, `y="116"`) {
		t.Errorf("pixel y should be Y+fontSize=116: %s", sub(svg, "y="))
	}
}

func TestRenderPositionRatio(t *testing.T) {
	svg, _ := fakeCardRender("5", func(p *RenderParams) {
		p.FontSize = 16
		p.Position = fdrawer.TextPos{RX: 0.5, RY: 0.25}
	})
	if !strings.Contains(svg, `x="100"`) {
		t.Errorf("ratio x should be imgW*rx=100: %s", sub(svg, "x="))
	}
	if !strings.Contains(svg, `y="116"`) {
		t.Errorf("ratio y should be imgH*ry+fs=116: %s", sub(svg, "y="))
	}
}

func TestRenderPositionPixelOverRatio(t *testing.T) {
	svg, _ := fakeCardRender("5", func(p *RenderParams) {
		p.FontSize = 16
		p.Position = fdrawer.TextPos{X: 10, Y: 20, RX: 0.9, RY: 0.9}
	})
	if !strings.Contains(svg, `x="10"`) {
		t.Errorf("pixel should override ratio: %s", sub(svg, "x="))
	}
}

// distinctFrames builds 4 frames with distinct data URIs (F0..F3).
func distinctFrames() []cardthemedrawer.Frame {
	frames := make([]cardthemedrawer.Frame, 4)
	for i := 0; i < 4; i++ {
		frames[i] = cardthemedrawer.Frame{Width: 10, Height: 20, Data: fmt.Sprintf("data:image/gif;base64,F%d", i)}
	}
	return frames
}

func TestPickFrameModeRandom(t *testing.T) {
	th := &cardthemedrawer.Theme{Name: "fake", Frames: distinctFrames()}
	r := rand.New(rand.NewSource(1))
	frame, ok := PickFrame(th, ModeRandom, 0, r)
	if !ok {
		t.Fatal("PickFrame returned false")
	}
	data := frame.Data
	if !strings.Contains(data, "F0") && !strings.Contains(data, "F1") &&
		!strings.Contains(data, "F2") && !strings.Contains(data, "F3") {
		t.Errorf("random mode did not pick a valid frame: %s", data)
	}
}

func TestPickFrameModeSeq(t *testing.T) {
	th := &cardthemedrawer.Theme{Name: "fake", Frames: distinctFrames()}
	frame, ok := PickFrame(th, ModeSeq, 2, nil)
	if !ok {
		t.Fatal("PickFrame returned false")
	}
	if !strings.Contains(frame.Data, "F2") {
		t.Errorf("seq mode should render FrameIndex 2: %s", frame.Data)
	}
}

// Character theme render produces 5 layered images + text.
func TestRenderCharacterProducesSVG(t *testing.T) {
	ch := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	portrait, err := ch.Assemble(r)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	svg, err := Render(RenderParams{
		ThemeKind: KindCharacter,
		Portrait:  portrait,
		Text:      "3",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.HasPrefix(svg, "<?xml") {
		t.Errorf("not svg: %q", svg[:16])
	}
	if got := strings.Count(svg, "<image"); got != 5 {
		t.Errorf("expected 5 <image> layers, got %d", got)
	}
	if !strings.Contains(svg, ">3<") {
		t.Errorf("count text 3 missing")
	}
}

func TestRenderCharacterUnshowFont(t *testing.T) {
	ch := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	portrait, _ := ch.Assemble(r)
	svg, _ := Render(RenderParams{
		ThemeKind: KindCharacter,
		Portrait:  portrait,
		Text:      "3",
		UnshowFont: true,
	})
	if strings.Contains(svg, "<text") {
		t.Errorf("unshowf should omit <text>")
	}
	if got := strings.Count(svg, "<image"); got != 5 {
		t.Errorf("expected 5 <image> layers with unshowf, got %d", got)
	}
}

func TestRenderCharacterScalesWholeCanvas(t *testing.T) {
	ch := fakeCharacter()
	r := rand.New(rand.NewSource(1))
	portrait, _ := ch.Assemble(r)
	svg, _ := Render(RenderParams{
		ThemeKind: KindCharacter,
		Portrait:  portrait,
		Text:      "1",
	})
	if !strings.Contains(svg, `viewBox="0 0 504 925"`) {
		t.Errorf("expected nested svg with canvas viewBox 0 0 504 925: %s", sub(svg, "viewBox"))
	}
}

// fakeCharacter mirrors characterthemedrawer's test helper.
func fakeCharacter() *characterthemedrawer.Character {
	const total = 80
	layers := make([]characterthemedrawer.CharacterLayer, total)
	parts := make(map[int]characterthemedrawer.CharacterPart, 70)
	for i := 0; i < total; i++ {
		layers[i] = characterthemedrawer.CharacterLayer{
			Name:    "L",
			Left:    100 + i,
			Top:     200 + i,
			Width:   50,
			Height:  60,
			LayerID: 1000 + i,
		}
		if i >= 1 && i <= 70 {
			parts[1000+i] = characterthemedrawer.CharacterPart{
				Left:   100 + i,
				Top:    200 + i,
				Width:  50,
				Height: 60,
				Data:   "data:image/png;base64,QQ",
			}
		}
	}
	return &characterthemedrawer.Character{Layers: layers, Parts: parts}
}

func TestFrameIndexForCount(t *testing.T) {
	if FrameIndexForCount(0, 1) != 0 {
		t.Error("single frame should always be 0")
	}
	if FrameIndexForCount(4, 3) != 2 {
		t.Errorf("(4+1)%%3 = %d, want 2", FrameIndexForCount(4, 3))
	}
}

func TestModeForTheme(t *testing.T) {
	if ModeForTheme(KindCharacter, "seq") != ModeRandom {
		t.Error("character themes should always be random")
	}
	if ModeForTheme(KindFrame, "random") != ModeRandom {
		t.Error("frame with mode=random should be random")
	}
	if ModeForTheme(KindFrame, "") != ModeSeq {
		t.Error("frame default should be seq")
	}
}

// Regression: when text is wider than the image, the text must be
// centered on the FULL canvas width (canvasWidth/2), not just the image
// width (bgW/2). The old composeSVG used textX = canvasWidth / 2.
func TestTextCenteredOnFullCanvasWhenWiderThanImage(t *testing.T) {
	// 10x20 frame -> displayed 200x400. Font 100, 20 chars ->
	// textW = 20*100*0.6 + 60 = 1260. canvasWidth = max(200, 1260) = 1260.
	frame := cardthemedrawer.Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	svg, err := Render(RenderParams{
		ThemeKind: KindFrame,
		Frame:     frame,
		Text:      "12345678901234567890",
		FontSize:  100,
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `text-anchor="middle"`) {
		t.Errorf("expected middle anchor")
	}
	// Text x should be canvasWidth/2 = 630, NOT bgW/2 = 100.
	if !strings.Contains(svg, `x="630"`) {
		t.Errorf("text should be centered on full canvas (x=630), got: %s", sub(svg, "<text"))
	}
}
