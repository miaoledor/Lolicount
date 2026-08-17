// Package assets embeds the repository-rooted assets/ directory into the
// binary via go:embed. The declaration lives at the repo root (not under
// internal/assets) because go:embed cannot traverse parent directories
// (".."); the embed pattern must be relative to the declaring .go file.
package assets

import "embed"

// FS exposes the embedded assets tree (themes, backgrounds metadata,
// static images and the generated themes.json). Submodules read from it
// at startup instead of touching the filesystem.
//
//go:embed all:assets
var FS embed.FS
