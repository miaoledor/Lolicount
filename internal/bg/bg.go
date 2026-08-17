// Package bg holds the background-image registry for the overlay render
// mode (AGENTS.md "方案 C"). A Background is an external CDN image plus
// its pixel dimensions; the counter digits are composited on top of it
// by theme.RenderWithBg.
//
// Iron Rule 2: background images reference an external URL
// (<image href="https://cdn...">); they are NEVER base64-embedded. Only
// the digit/counter images use data URIs. Mixing the two would either
// bloat the SVG (base64 background) or break digits offline.
package bg

// Background is one composable background image. The URL is referenced
// directly in the SVG <image> tag (not embedded), so it must be a
// publicly fetchable, long-cacheable resource.
type Background struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Registry resolves a background name to its metadata. The reserved
// value "random" is handled by the caller (handler), not the registry.
type Registry interface {
	// Get returns the background for name, or false if not registered.
	Get(name string) (Background, bool)
	// List returns the names of all registered backgrounds.
	List() []string
}
