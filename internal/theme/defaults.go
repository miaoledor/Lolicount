package theme

// defaults.go is the single source of truth for render-time defaults.
// Centralizing them here keeps the rendering contract in one file so
// tuning display size, font, colors or the default theme does not require
// hunting through render.go / params.go.

// DefaultTheme is the theme used when the request omits ?theme=.
const DefaultTheme = "lian"

// DefaultDisplaySize is the longest-edge target (in px) every frame is
// scaled to when no explicit Scale is given (M5.6: all images show at a
// consistent size). Frames smaller than this are scaled up to it as well
// so output is uniform across themes.
const DefaultDisplaySize = 400

// DefaultFontSize is the counter text size in pixels when fsize is not
// set. Font sizing is independent of image Scale (M5.6).
const DefaultFontSize = 16

// MonoCharWidthFactor approximates the advance width of one monospace
// digit relative to font-size: monospace digits are ~0.6em wide. Used to
// estimate text width so the viewBox can grow to fit long counts.
const MonoCharWidthFactor = 0.6

// DefaultFontFamily is the CSS font-family used for the counter text.
const DefaultFontFamily = "monospace"

// DefaultFontColor is the fill color of the counter text.
const DefaultFontColor = "#333"

// TextGapBelowImage is the extra pixels between the image bottom and the
// counter text baseline, on top of the font size.
const TextGapBelowImage = 4
