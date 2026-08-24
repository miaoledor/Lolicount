package fdrawer

// defaults.go is the single source of truth for text-layer render
// defaults (font size, family, color, spacing).

// DefaultFontSize is the counter text size in pixels when fsize is not
// set. Font sizing is independent of image Scale (M5.6).
const DefaultFontSize = 16

// MonoCharWidthFactor approximates the advance width of one monospace
// digit relative to font-size: monospace digits are ~0.6em wide.
const MonoCharWidthFactor = 0.6

// DefaultFontFamily is the CSS font-family used for the counter text.
const DefaultFontFamily = "monospace"

// DefaultFontColor is the fill color of the counter text.
const DefaultFontColor = "#333"

// TextGapBelowImage is the extra pixels between the image bottom and the
// counter text baseline, on top of the font size.
const TextGapBelowImage = 4
