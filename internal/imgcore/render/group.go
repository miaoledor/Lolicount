package render

import (
	"fmt"
	"strings"

	"github.com/miaoledor/lolicount/internal/imgcore"
)

// GroupLayer renders multiple images as a nested <svg> with a viewBox
// that maps a source coordinate space (e.g. PSD canvas) onto the output
// viewport. This preserves sub-pixel alignment between layers and
// supports crop regions.
//
// Each GroupPart can carry multiple Candidates — at render time the PRNG
// picks one candidate per part (e.g. one random expression from the
// "brow" category). When Candidates is empty, Src/X/Y/Width/Height are
// used directly (fixed part).
type GroupLayer struct {
	Parts []GroupPart
	OutW  int
	OutH  int
	VbX   int
	VbY   int
	VbW   int
	VbH   int
	Z     int
}

// GroupPart is one layer slot in a GroupLayer. When Candidates is
// non-empty, the PRNG picks one at render time (random selection per
// request). When empty, the fixed Src/X/Y/Width/Height fields are used.
type GroupPart struct {
	Src    string
	X      int
	Y      int
	Width  int
	Height int
	// Candidates are alternative images for this part slot. Each
	// candidate carries its own Src and placement. At render time one
	// candidate is randomly selected. When empty, the fixed fields above
	// are used.
	Candidates []GroupCandidate
}

// GroupCandidate is one alternative image for a GroupPart slot.
type GroupCandidate struct {
	Src    string
	X      int
	Y      int
	Width  int
	Height int
}

// Kind returns LayerImage since a group renders as image content.
func (l *GroupLayer) Kind() imgcore.LayerKind { return imgcore.LayerImage }

// ZIndex returns the stack order.
func (l *GroupLayer) ZIndex() int { return l.Z }

// Fixed reports whether the layer cannot be deleted.
func (l *GroupLayer) Fixed() bool { return false }

// Render produces a nested <svg> fragment. For each part with
// Candidates, the PRNG picks one candidate; otherwise the fixed part is
// rendered directly.
func (l *GroupLayer) Render(ctx imgcore.RenderCtx) imgcore.LayerOutput {
	var b strings.Builder
	fmt.Fprintf(&b, `  <svg x="0" y="0" width="%d" height="%d" viewBox="%d %d %d %d">`+"\n",
		l.OutW, l.OutH, l.VbX, l.VbY, l.VbW, l.VbH)
	for _, part := range l.Parts {
		src := part.Src
		x := part.X
		y := part.Y
		w := part.Width
		h := part.Height
		if len(part.Candidates) > 0 {
			weights := make([]float64, len(part.Candidates))
			for j := range weights {
				weights[j] = 1
			}
			picked := part.Candidates[ctx.PRNG.WeightedPick(weights)]
			src = picked.Src
			x = picked.X
			y = picked.Y
			w = picked.Width
			h = picked.Height
		}
		fmt.Fprintf(&b, `    <image x="%d" y="%d" width="%d" height="%d" xlink:href="%s" />`+"\n",
			x, y, w, h, src)
	}
	b.WriteString("  </svg>\n")
	return imgcore.LayerOutput{Fragment: b.String(), Width: l.OutW, Height: l.OutH}
}
