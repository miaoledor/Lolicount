// Package assets embeds the theme/static-image trees into the binary via
// go:embed. The declaration lives in this package directory because
// go:embed patterns are relative to the declaring .go file and cannot
// traverse parent directories.
package assets

import "embed"

// FS exposes the embedded asset trees. "all:" is used so files with a
// leading underscore (e.g. _start.gif, _end.gif) are included — without
// it go:embed skips files whose names begin with _ or .
//
//go:embed all:theme all:img README.md
var FS embed.FS
