// Package renderer is the composition layer: it merges the independent
// drawer layers (card/character background + font text) into the final
// SVG document. It is the sole rendering entry point the server calls.
//
// The three drawers (cardthemedrawer, characterthemedrawer, fdrawer) are
// mutually independent — none imports another. The renderer is the only
// place that knows about all three and combines their Layer outputs.
package renderer

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/miaoledor/lolicount/internal/drawer"
	"github.com/miaoledor/lolicount/internal/drawer/cardthemedrawer"
	"github.com/miaoledor/lolicount/internal/drawer/characterthemedrawer"
	"github.com/miaoledor/lolicount/internal/drawer/fdrawer"
)

// RenderParams controls how the background + counter text are composed.
// The server resolves the final text string (from count/number/demo)
// before calling Render — the renderer only draws, it does not decide
// what text to show.
type RenderParams struct {
	// ThemeKind selects which background drawer to use.
	ThemeKind drawer.Kind
	// Frame is the card-theme frame to draw (KindFrame only).
	Frame cardthemedrawer.Frame
	// Portrait is the assembled character portrait (KindCharacter only).
	Portrait *characterthemedrawer.ComposedPortrait
	// Scale controls the image display size (0 = uniform base).
	Scale float64

	// Text is the final counter text string to render.
	Text string
	// FontSize controls the counter text size (0 = default 16).
	FontSize int
	// UnshowFont omits the counter text entirely.
	UnshowFont bool
	// FontStyle applies a font-style theme to the text.
	FontStyle fdrawer.FontStyle
	// Position places the counter text.
	Position fdrawer.TextPos
}

// Render composes the final SVG: the background (layer 0) produced by
// cardthemedrawer or characterthemedrawer, with the counter text
// (layer 1) produced by fdrawer overlaid. The canvas grows to fit
// whichever layer is wider, and the text sits below the image by default.
func Render(p RenderParams) (string, error) {
	// Layer 0: background.
	var bg drawer.Layer
	switch p.ThemeKind {
	case drawer.KindCharacter:
		if p.Portrait == nil {
			return "", fmt.Errorf("renderer: character portrait is nil")
		}
		bg = characterthemedrawer.Draw(p.Portrait, p.Scale)
	default:
		bg = cardthemedrawer.Draw(p.Frame, p.Scale)
	}

	// Layer 1: text. Passes bg dimensions for ratio positioning.
	textLayer := fdrawer.Draw(fdrawer.Params{
		Text:       p.Text,
		FontSize:   p.FontSize,
		UnshowFont: p.UnshowFont,
		FontStyle:  p.FontStyle,
		Position:   p.Position,
	}, bg.Width, bg.Height)

	// Merge canvas: max(bg width, text width) x (bg height + text height).
	canvasWidth := bg.Width
	if textLayer.Width > canvasWidth {
		canvasWidth = textLayer.Width
	}
	canvasHeight := bg.Height + textLayer.Height

	// Center the background horizontally if text widened the canvas.
	imgX := (canvasWidth - bg.Width) / 2
	if imgX < 0 {
		imgX = 0
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		canvasWidth, canvasHeight, canvasWidth, canvasHeight)
	b.WriteString("  <title>Lolicount</title>\n")

	// Layer 0 fragment: if the background needs horizontal centering,
	// wrap it in a group with an x offset. The card drawer emits an
	// <image> at x=0; the character drawer emits a nested <svg> at x=0.
	// Wrapping in <g transform="translate(x,0)"> shifts both uniformly.
	if imgX > 0 {
		fmt.Fprintf(&b, `  <g transform="translate(%d,0)">`+"\n", imgX)
		b.WriteString(bg.Fragment)
		b.WriteString("  </g>\n")
	} else {
		b.WriteString(bg.Fragment)
	}

	// Layer 1 fragment (may be empty when UnshowFont).
	b.WriteString(textLayer.Fragment)

	b.WriteString("</svg>\n")
	return b.String(), nil
}

// FrameIndexForCount picks the background frame for a given count.
// Per M2.5: display frame[(count+1) % size]. size<=1 guards against a
// single-frame theme and division by zero.
func FrameIndexForCount(count int64, size int) int {
	if size <= 1 {
		return 0
	}
	return int((count + 1) % int64(size))
}

// ModeForTheme resolves the effective render Mode for a theme given the
// ?mode= query param. Character themes only support random mode; a
// ?mode=seq on a character theme is coerced to random. Frame themes
// default to sequential unless ?mode=random is requested.
func ModeForTheme(kind drawer.Kind, modeParam string) drawer.Mode {
	if kind == drawer.KindCharacter {
		return drawer.ModeRandom
	}
	if modeParam == "random" {
		return drawer.ModeRandom
	}
	return drawer.ModeSeq
}

// PickFrame selects a frame from a card theme according to the mode.
// ModeSeq uses frameIndex; ModeRandom picks a random frame. Returns
// false if the index is out of range.
func PickFrame(th *cardthemedrawer.Theme, mode drawer.Mode, frameIndex int, r *rand.Rand) (cardthemedrawer.Frame, bool) {
	if mode == drawer.ModeRandom {
		return th.FrameAt(drawerRandomInt(r, th.Size()))
	}
	return th.FrameAt(frameIndex)
}

// drawerRandomInt returns a uniform random int in [0, n). When r is nil
// the package-global source is used.
func drawerRandomInt(r *rand.Rand, n int) int {
	if n <= 0 {
		return 0
	}
	if r != nil {
		return r.Intn(n)
	}
	return rand.Intn(n)
}
