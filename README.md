<div align="center">
<img src="docs/png/lolicount-s.png" alt="Lolicount" width="480" />
<p><b>English</b> · <a href="./README.zh.md">中文</a> · <a href="./README.ja.md">日本語</a></p>
<p>A themeable visitor counter for your GitHub README, blog, or website. Choose a character, copy the embed code, and you're ready to go. Static themes can pick a random frame or combine character layers on each request.</p>
</div>

<div align="center">
<table>
  <tr>
    <td align="center" width="33%"><h2>1,574</h2><b>Built-in themes</b><br /><sub>Static themes + animated models</sub></td>
    <td align="center" width="33%"><h2>311,040</h2><b>Max variations in one theme</b><br /><sub>Sprite combinations of lian-ren</sub></td>
    <td align="center" width="33%"><h2>14,000+</h2><b>Fastest request speed</b><br /><sub>Renders per second, single-core benchmark</sub></td>
  </tr>
  <tr>
    <td colspan="3" align="center"><br /><img src="https://img.shields.io/badge/QQ-1121240595-12B7F5?style=for-the-badge&amp;logo=qq&amp;logoColor=white" alt="QQ group 1121240595" /></td>
  </tr>
</table>
</div>

## Get started

1. Open [lolicount.top](https://lolicount.top).
2. Enter a counter name, choose a theme, and generate your counter.
3. Copy the Markdown or HTML code into your page.

Use your own name in place of `my-counter` below. The same name shares one count across all embeds.

**GitHub README / Markdown**

```markdown
![Visits](https://lolicount.top/@my-counter?theme=lian-ren&fsize=16&scale=0.6)
```

**Website / HTML**

```html
<img src="https://lolicount.top/@my-counter?theme=lian-ren&amp;fsize=16&amp;scale=0.6" alt="Visits" />
```

**Image URL**: [open the example](https://lolicount.top/@my-counter?theme=lian-ren&fsize=16&scale=0.6), or copy its address into any field that accepts an image URL.

For GitHub READMEs, choose an SVG theme. E-mote, Spine, and Live2D themes use the website widget code generated on the site and require JavaScript.

## Customize

Adjust the URL parameters, or use the online generator to preview changes:

| Parameter | What it does | Example |
|---|---|---|
| `theme` | Character theme; `random` picks a random SVG theme | `theme=lian-ren` |
| `scale` | Image scale, from `0.1` to `4`; default `1` | `scale=0.6` |
| `fsize` | Counter font size in pixels, from `0` to `500`; `0` uses the default | `fsize=16` |
| `number` | Display a fixed value without counting, from `1` to `999999`; omit for real counts | `number=12345` |

For a preview without counting, replace the name with [`demo`](https://lolicount.top/@demo?theme=lian-ren&fsize=16&scale=0.6). It displays `0123456789` unless `number` is set.

See the [full parameter reference](./docs/projectDesign.md#接口调用文档) for font styles and text positioning, or the [FAQ](./docs/faq.md) for troubleshooting.

## Self-host

Start with Docker:

```sh
docker run -d --name lolicount --restart unless-stopped -p 9721:9721 -v lolicount-data:/app/data ghcr.io/miaoledor/lolicount:latest
```

Open [localhost:9721](http://localhost:9721) to generate counters. Data is saved in the `lolicount-data` volume. Run a single instance; when using a public domain, set `BASE_URL` to that origin, such as `https://counter.example.com`.

See the [deployment guide](./docs/deployment.md) for Docker Compose, configuration, and binary builds.

## Local development

With Go 1.25+, Node.js 22+, and pnpm 11 installed, run these commands from the repository root:

```sh
pnpm install
pnpm dev
```

Open [localhost:3721](http://localhost:3721) for the frontend. The backend runs at [127.0.0.1:9721](http://127.0.0.1:9721).

## Contribute & support

- [Contributing](./CONTRIBUTING.md) · [Add a theme](./docs/contributing-themes.md) · [Contribute code](./docs/contributing-code.md)
- [Documentation](./docs/index.md) · [Architecture](./docs/architecture.md) · [Roadmap](./docs/TODOlist.md)
- Theme contributions are accepted only through repository pull requests; see [Add a theme](./docs/contributing-themes.md).
- [Buy the author a milk tea](https://github.com/sponsors/miaoledor) 🧋

<div align="center">
<sub>Thanks to <a href="https://github.com/KunMoe/kun-galgame-forum">kun-galgame-forum</a> and <a href="https://github.com/journey-ad/Moe-Counter">Moe-Counter</a>.</sub><br />
<sub>Licensed under <a href="./LICENSE">AGPL-3.0</a>.</sub><br />
<img src="https://lolicount.top/@lolicount?theme=lian-ren&amp;fsize=16&amp;scale=1&amp;unshowf=true" alt="Lolicount" />
</div>
