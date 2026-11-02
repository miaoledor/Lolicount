# assets

Static resources for Lolicount. Core resources are packed into the Go binary via
`go:embed` (see `assets/embed.go`); `psb/` is the documented runtime exception.

- `theme/<name>/` — built-in themes (both types live in this unified tree):
  - Single-layer (frame) themes: frame images `0..n-1` (gif/png/webp),
    optional `meta.json`
  - Multi-layer (character) themes: `ren.json` + `config.json` + optional
    `display.json` + layered webp images (`ren/*.webp`), assembled randomly
    per request
- `f-theme/<name>.json` — built-in font-style themes (counter text
  family/color/weight)
- `img/` — static images for README/frontend (logo, screenshots)
- `psb/` — optional E-mote models loaded from disk through `PSB_DIR`; not embedded
- `themes.json` — CI-generated manifest for all built-in themes; the
  runtime API lists themes from ThemeRegistry
- `dist/` — pre-built Nuxt SSG frontend (copied in at build time; only
  `.gitkeep` is tracked)
