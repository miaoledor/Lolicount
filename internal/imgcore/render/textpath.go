package render

import (
	"fmt"
	"math"
	"strings"
)

// ArcPath generates an SVG <path> defining a circular arc centered at
// (cx, cy) with the given radius, spanning from startAngle to
// startAngle+sweep degrees. The path is used as a <textPath> reference
// so text follows the arc curve. Angles are in degrees, 0 = rightward,
// increasing clockwise (SVG convention).
//
// This supports the edit-design spec's "文字弧度" (text arc) feature:
// the text layer can specify an arc radius and sweep, and the text is
// rendered along the curved path instead of a straight line.
func ArcPath(cx, cy, radius, startAngle, sweep float64) string {
	if radius <= 0 {
		return ""
	}
	startRad := degToRad(startAngle)
	endRad := degToRad(startAngle + sweep)
	x1 := cx + radius*math.Cos(startRad)
	y1 := cy + radius*math.Sin(startRad)
	x2 := cx + radius*math.Cos(endRad)
	y2 := cy + radius*math.Sin(endRad)
	largeArc := 0
	if sweep > 180 {
		largeArc = 1
	}
	return fmt.Sprintf("M %s %s A %s %s 0 %d 1 %s %s",
		formatFloat(x1), formatFloat(y1),
		formatFloat(radius), formatFloat(radius),
		largeArc,
		formatFloat(x2), formatFloat(y2))
}

// RenderArcText produces an SVG fragment with text following a circular
// arc path. The path is defined inline in <defs> with a unique ID so
// the <textPath> can reference it. When arcRadius is 0 or negative,
// returns empty string (caller should use straight text instead).
func RenderArcText(text, fontFamily, color, fontWeight string, fontSize int,
	cx, cy, arcRadius, startAngle, sweep float64) string {
	if arcRadius <= 0 {
		return ""
	}
	pathID := "arc-text"
	pathD := ArcPath(cx, cy, arcRadius, startAngle, sweep)
	if pathD == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`  <defs><path id="%s" d="%s" /></defs>`+"\n", pathID, pathD))
	weightAttr := ""
	if fontWeight != "" {
		weightAttr = ` font-weight="` + fontWeight + `"`
	}
	b.WriteString(fmt.Sprintf(`  <text font-family="%s" font-size="%d" fill="%s"%s>`+"\n",
		fontFamily, fontSize, color, weightAttr))
	b.WriteString(fmt.Sprintf(`    <textPath href="#%s">%s</textPath>`+"\n", pathID, text))
	b.WriteString("  </text>\n")
	return b.String()
}

// degToRad converts degrees to radians.
func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}
