# Lolicount

> A cute, themeable SVG visitor counter — pick a built-in theme or upload your own, then paste a link and watch it count.

[中文](./README.md) · [日本語](./README.ja.md) · **English**

Lolicount is a cute, themeable visitor counter that renders as an SVG image.
It ships with several built-in themes, and you can also upload your own
frame images to create a custom style. Paste one link in your README or
homepage and the number goes up by one on every visit.

This project renders a single image at a time. For example, a theme with
frames `0.png 1.png 2.png ... (n-1).png` shows frame `(count+1) % n` on
each visit, then increments `count`.

## Features

- 🎀 **Cute themes** — built-in loli-style frame images, supports gif/png/webp
- 🎨 **Themeable** — multiple built-in themes, or upload your own frames
- 🖼️ **Background overlay** — overlay the counter onto any background image
- 📊 **SVG output** — crisp vectors, embed in a README, no JS required
- ⚡ **High performance** — Go + Fiber v3, in-memory buffer + batched SQLite writes
- 🛡️ **Rate limiting** — IP-level + name-level dual rate limiting, anti-spam
- 🚀 **Single binary** — frontend + themes embedded into the Go binary
- 🤝 **Community** — PR channel (CI-validated) + Web upload channel

## Quick Start

### Docker

```bash
docker run -d -p 9721:9721 \
  -v lolicount-data:/app/data \
  ghcr.io/miaoledor/lolicount:latest
```

Open `http://localhost:9721/@my-counter`. Counter data is persisted to the
SQLite file inside the `lolicount-data` volume.

### From Source

```bash
git clone https://github.com/miaoledor/Lolicount.git
cd Lolicount
cp .env.example .env
go run ./cmd/server
```

### Run Frontend + Backend Together (Dev Mode)

The root `package.json` uses `concurrently` to start both the backend
(Go :9721) and the frontend (Nuxt :3721). It is cross-platform compatible
with macOS / Windows / Linux:

```bash
pnpm install        # installs concurrently (root) and frontend deps
pnpm dev            # starts both frontend and backend
```

- Backend: http://127.0.0.1:9721
- Frontend: http://localhost:3721

You can also run them separately: `pnpm dev:server` (backend only) or
`pnpm dev:web` (frontend only).

### Usage

Embed in a README or webpage:

```markdown
![visitor](https://umi7.top/@my-counter?theme=lian)
```

With a background overlay:

```markdown
![visitor](https://umi7.top/@my-counter?theme=lian&scale=2)
```

Three embed formats (same URL):

```
1. SVG address
   https://umi7.top/@my-counter?theme=lian

2. Img tag
   <img src="https://umi7.top/@my-counter?theme=lian" alt="my-counter" />

3. Markdown
   ![my-counter](https://umi7.top/@my-counter?theme=lian)
```

## Parameters

| Parameter | Description | Default |
|---|---|---|
| `theme` | Theme name, or `random` | `lian` |
| `fsize` | Counter font size (px) | `16` |
| `scale` | Image display scale (based on a unified 400px longest edge) | `1` |
| `number` | Preview a specific number (not persisted, no +1) | none |
| `unshowf` | Hide the counter text (`true`/`false`) | `false` |

> `scale` controls the image size, `fsize` controls the text size; the two
> are independent. When `scale` is omitted, all theme images are scaled to
> a 400px longest edge while preserving the aspect ratio.

## Default Configuration

All rendering defaults live in one file: `internal/theme/defaults.go`. To
change the default behavior, just edit that file — no need to touch the
rendering logic.

| Constant | Description | Default |
|---|---|---|
| `DefaultTheme` | Theme used when `?theme=` is omitted | `lian` |
| `DefaultDisplaySize` | Unified longest-edge target (px), used when `scale` is omitted | `400` |
| `DefaultFontSize` | Counter font size when `fsize` is omitted | `16` |
| `MonoCharWidthFactor` | Monospace char-width estimate (relative to font size) | `0.6` |
| `DefaultFontFamily` | Counter text CSS `font-family` | `monospace` |
| `DefaultFontColor` | Counter text color | `#333` |
| `TextGapBelowImage` | Extra gap between image bottom and text baseline (px) | `4` |

Example — set the default image size to 600px and font size to 20:

```go
// internal/theme/defaults.go
const DefaultDisplaySize = 600
const DefaultFontSize   = 20
```

Rebuild to apply: `go build -o lolicount ./cmd/server && ./lolicount`

## API

| Method | Path | Description |
|---|---|---|
| GET | `/@:name` | Count +1, returns SVG |
| GET | `/get/@:name` | Same (compat alias) |
| GET | `/record/@:name` | Returns JSON count |
| GET | `/heart-beat` | Health check |
| GET | `/api/themes` | Theme list |
| POST | `/api/themes` | Upload a theme |
| GET | `/api/backgrounds` | Background list |
| POST | `/api/backgrounds` | Upload a background |

See [docs/detail.md](./docs/detail.md) for details.

## Contributing Themes

Lolicount uses **frame-based themes**: each theme is a directory containing
several frame images `0.<ext> 1.<ext> ... n-1.<ext>`. The visit counter
cycles through them as `(count+1) % n`. Extensions `gif` / `png` / `webp`
are supported, and frame indices must be contiguous starting from 0.

Two ways to contribute:

**PR channel** — fork the repo, put your frames in
`assets/theme/<your-theme>/` (at least 1 frame, indices starting at 0) with
an optional `meta.json`, and open a PR. CI runs automatically:

- `cmd/check-theme` validates the directory name, frame completeness, format and size
- `scripts/validate-theme-meta.js` validates the `meta.json` schema
- `scripts/gen-themes-json.js` validates that `assets/themes.json` is in sync

**Web upload** — visit the `/upload` page and upload frames; they are
available immediately (server-side re-encoded).

`meta.json` example:

```json
{
  "name": "lian",
  "author": "yourname",
  "description": "Loli-style digit frames",
  "tags": ["cute", "anime"],
  "version": "1.0.0"
}
```

Local pre-validation:

```bash
go run ./cmd/check-theme
node scripts/validate-theme-meta.js
node scripts/gen-themes-json.js
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the full contribution guide.

## CI/CD & Deployment

| Workflow | Trigger | Purpose |
|---|---|---|
| `ci.yml` | push / PR | go vet + `go test -race` + check-theme + Nuxt build |
| `theme-check.yml` | PR changes `assets/theme`, `assets/bg` | Theme integrity + meta.json + themes.json sync |
| `release.yml` | tag `v*` | Build Docker image + Release binaries |
| `rebuild-frontend.yml` | theme change on default branch | Rebuild SSG dist and commit |

**Docker**: `docker compose up -d`, open `http://localhost:9721/@my-counter`.
**Release**: tag it `git tag v0.1.0 && git push --tags`, CI builds the image and publishes a Release.

## Tech Stack

- **Backend**: Go 1.23+ / Fiber v3 / SQLite (`modernc.org/sqlite`, pure Go, CGO-free)
- **Frontend**: Nuxt 3 SSG / UnoCSS / GSAP
- **Storage**: request → in-memory buffer → batched writes → SQLite
- **Deployment**: single binary (embed.FS bundles themes + frontend dist)

## Acknowledgements

- [kun-galgame-forum](https://github.com/KunMoe/kun-galgame-forum) — the loli character layering approach
- [Moe-Counter](https://github.com/journey-ad/Moe-Counter) — the original Moe-Counter this project is inspired by

## Sponsor

If Lolicount helps you, consider [sponsoring the author](https://github.com/sponsors/miaoledor) 🧋
