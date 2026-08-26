package render

import (
	"fmt"
	"strings"

	"github.com/miaoledor/lolicount/internal/imgcore"
)

// GroupLayer renders multiple images as a nested <svg> with a viewBox
// that maps a source coordinate space (e.g. PSD canvas) onto the output
// viewport. This preserves sub-pixel alignment between layers and
// supports crop regions — the same approach used by the old
// characterthemedrawer.drawLayeredSVG.
//
// The nested <svg viewBox="vbX vbY vbW vbH" width="outW" height="outH">
// tells the browser to map the source rectangle (vbX,vbY,vbW,vbH) onto
// the output viewport (outW,outH), so all child <image> elements with
// coordinates in the source space are automatically scaled and clipped.
type GroupLayer struct {
	// Parts are the images to render, positioned in the source
	// coordinate space.
	Parts []GroupPart
	// OutW/OutH are the output viewport dimensions in pixels.
	OutW int
	OutH int
	// VbX/VbY/VbW/VbH define the source rectangle mapped to the output
	// viewport. Typically (0,0,canvasW,canvasH) or a crop region.
	VbX int
	VbY int
	VbW int
	VbH int
	Z   int
}

// GroupPart is one image in a GroupLayer, positioned in the source
// coordinate space.
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

// Render produces a nested <svg> fragment containing all parts.
func (l *GroupLayer) Render(ctx imgcore.RenderCtx) imgcore.LayerOutput {
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
