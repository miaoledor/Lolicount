# assets

Embedded static resources, packed into the Go binary via `go:embed` (see
`embed.go` at the repo root).

- `theme/<name>/` — built-in digit-glyph themes (`0`..`9`, optional
  `_start`/`_end`, optional `meta.json`)
- `bg/` — built-in background metadata JSON (image bytes live on a CDN)
- `img/` — static images for README/frontend (logo, screenshots)
- `themes.json` — CI-generated theme manifest (consumed by the frontend)

This directory is intentionally non-empty so `//go:embed all:assets` has
content to mount during M1.
