# assets

Embedded static resources, packed into the Go binary via `go:embed` (see
`assets/embed.go`).

- `theme/<name>/` — built-in themes (both types live in this unified tree):
  - Single-layer (frame) themes: frame images `0..n-1` (gif/png/webp),
    optional `meta.json`
  - Multi-layer (character) themes: `ren.json` + `config.json` +
    `display.json` + layered webp images (`ren/*.webp`), assembled
    randomly per request
- `f-theme/<name>.json` — built-in font-style themes (counter text
  family/color/weight)
- `img/` — static images for README/frontend (logo, screenshots)
- `themes.json` — CI-generated theme manifest (consumed by the frontend)
- `dist/` — pre-built Nuxt SSG frontend (copied in at build time; only
  `.gitkeep` is tracked)
