package render

import "github.com/miaoledor/lolicount/internal/imgcore/theme"

// MeasureText returns the width and height the text layer would occupy,
// without generating the SVG fragment. The composer calls this before
// rendering so it can compute the final canvas width. Migrated from
func MeasureText(text string, fontSize int, unshowFont bool) (width, height int) {
	if unshowFont {
		return 0, 0
	}
	fs := effectiveFontSize(fontSize)
	charW := int(float64(fs) * theme.MonoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}
	return textWidth(text, fs) + charW, fs + theme.TextGapBelowImage
}

// effectiveFontSize resolves the font size from fsize, falling back to
// the default. Font sizing is independent of image Scale.
func effectiveFontSize(fsize int) int {
	fs := fsize
	if fs <= 0 {
		fs = theme.DefaultFontSize
	}
	if fs < 1 {
		fs = 1
	}
	return fs
}

// textWidth estimates the pixel width of the counter text at a given
// font size so the viewBox can grow to fit it.
func textWidth(text string, fontSize int) int {
	if len(text) == 0 {
		return 0
	}
	return int(float64(len(text)*fontSize) * theme.MonoCharWidthFactor)
}
