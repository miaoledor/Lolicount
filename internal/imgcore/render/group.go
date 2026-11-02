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

// GroupPart is one resident image slot in a GroupLayer. Candidate
// selection happens in the lazy registry before rendering.
type GroupPart struct {
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

// Render produces a nested <svg> fragment from the registry's resident
// parts.
func (l *GroupLayer) Render(_ imgcore.RenderCtx) imgcore.LayerOutput {
	var b strings.Builder
	fmt.Fprintf(&b, `  <svg x="0" y="0" width="%d" height="%d" viewBox="%d %d %d %d">`+"\n",
		l.OutW, l.OutH, l.VbX, l.VbY, l.VbW, l.VbH)
	for _, part := range l.Parts {
		fmt.Fprintf(&b, `    <image x="%d" y="%d" width="%d" height="%d" xlink:href="%s" />`+"\n",
			part.X, part.Y, part.Width, part.Height, part.Src)
	}
	b.WriteString("  </svg>\n")
	return imgcore.LayerOutput{Fragment: b.String(), Width: l.OutW, Height: l.OutH}
}
