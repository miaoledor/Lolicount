// Package assets embeds the theme/static-image trees into the binary via
// go:embed. The declaration lives in this package directory because
// go:embed patterns are relative to the declaring .go file and cannot
// traverse parent directories.
package assets

import "embed"

// FS exposes the embedded asset trees. All themes — both single-layer
// (frame) and multi-layer (character) — live under the unified
// assets/theme/ tree and are loaded into the ThemeRegistry at startup.
// "all:" is used so files with a leading underscore (e.g. _start.gif,
// _end.gif) are included — without it go:embed skips files whose names
// begin with _ or .
//
//go:embed all:theme all:f-theme all:img README.md
var FS embed.FS

// DistFS holds the pre-built Nuxt SSG frontend. At build time the
// release/Docker pipeline copies web/dist into assets/dist so the static
// site is served from the single binary. The placeholder .gitkeep keeps
// the directory non-empty so this embed always compiles even before a
// build has populated it; in that case the server logs a warning and
// skips serving the frontend (local dev uses the Nuxt dev server instead).
//
//go:embed all:dist
var DistFS embed.FS
