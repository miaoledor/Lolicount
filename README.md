# Lolicount

> A cute, themeable SVG visitor counter — pick a built-in theme or upload your own, then paste a link and watch it count.

**中文** · [English](./README.en.md) · [日本語](./README.ja.md)



萌系可换肤访问计数器,以 SVG 图片形式输出。内置多套主题,也可上传自己的数字图或底图打造专属风格。往 README 或主页贴一行链接,每次访问数字自然 +1。

本项目只使用一张图片进行展示
比如default-theme[0.png 1.png 2.png 3.png ... size-1.png]
则每次显示(count+1)%size,count++

## 特性

- 🎀 **萌系主题** — 内置萝莉风格数字图,支持 gif/png/webp
- 🎨 **可换肤** — 多套内置主题,或上传自己的帧图
- 🖼️ **底图叠加** — 把计数器叠加到任意底图上(立绘、徽章、海报)
- 📊 **SVG 输出** — 矢量清晰,嵌入 README 即可,无需 JS
- ⚡ **高性能** — Go + Fiber v3,内存缓冲 + 定时批量落库 SQLite
- 🛡️ **限流防护** — IP 级 + name 级双层限流,防刷防注水
- 🚀 **单二进制** — 前端 + 主题 embed 进 Go,一次构建到处部署
- 🤝 **社区贡献** — PR 通道(CI 自动校验)+ Web 上传通道

## 快速开始

### Docker

```bash
docker run -d -p 9721:9721 \
  -v lolicount-data:/app/data \
  ghcr.io/miaoledor/lolicount:latest
```

访问 `http://localhost:9721/@my-counter` 即可。计数数据持久化到 `lolicount-data` 卷的 SQLite 文件。

### 从源码

```bash
git clone https://github.com/miaoledor/Lolicount.git
cd lolicount
cp .env.example .env
go run ./cmd/server
```

### 同时运行前后端(开发模式)

根目录 `package.json` 用 `concurrently` 同时启动后端(Go :9721)和前端(Nuxt :3721),跨平台兼容 macOS / Windows / Linux:

```bash
pnpm install        # 安装 concurrently(根目录)与前端依赖
pnpm dev            # 同时启动前后端
```

- 后端:http://127.0.0.1:9721
- 前端:http://localhost:3721

也可单独运行:`pnpm dev:server`(仅后端)或 `pnpm dev:web`(仅前端)。

### 使用

在 README 或网页里嵌入:

```markdown
![visitor](https://umi7.top/@my-counter?theme=lian)
```

带底图叠加:

```markdown
![visitor](https://umi7.top/@my-counter?theme=lian&scale=2)
```

三种嵌入方式(同一个 URL):

```
1. SVG address
   https://umi7.top/@my-counter?theme=lian

2. Img tag
   <img src="https://umi7.top/@my-counter?theme=lian" alt="my-counter" />

3. Markdown
   ![my-counter](https://umi7.top/@my-counter?theme=lian)
```

## 参数

| 参数 | 说明 | 默认值 |
|---|---|---|
| `theme` | 主题名,或 `random` | `lian` |
| `fsize` | 计数文字字号(像素) | `16` |
| `scale` | 图片展示尺寸倍数(基于统一最长边 400px) | `1` |
| `number` | 指定数字预览(不落库、不 +1) | 无 |
| `unshowf` | 隐藏计数文字(`true`/`false`) | `false` |

> `scale` 控制图片大小,`fsize` 控制文字大小,二者独立。不传 `scale` 时所有主题图片等比缩放到最长边 400px,保持宽高比不拉伸。

## 默认配置

所有渲染默认值集中在一个文件:`internal/theme/defaults.go`。修改默认行为只需编辑该文件,无需改动渲染逻辑。

| 常量 | 说明 | 默认值 |
|---|---|---|
| `DefaultTheme` | 未传 `?theme=` 时使用的主题 | `lian` |
| `DefaultDisplaySize` | 图片统一最长边目标(像素),不传 `scale` 时生效 | `400` |
| `DefaultFontSize` | 未传 `fsize` 时的计数文字字号 | `16` |
| `MonoCharWidthFactor` | 等宽字体字宽估算系数(相对字号) | `0.6` |
| `DefaultFontFamily` | 计数文字 CSS `font-family` | `monospace` |
| `DefaultFontColor` | 计数文字颜色 | `#333` |
| `TextGapBelowImage` | 图片底部与文字基线的额外间距(像素) | `4` |

示例:把默认图片尺寸调到 600px、字号调到 20:

```go
// internal/theme/defaults.go
const DefaultDisplaySize = 600
const DefaultFontSize   = 20
```

修改后重新构建即可生效:`go build -o lolicount ./cmd/server && ./lolicount`

## API

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/@:name` | 计数 +1,返回 SVG |
| GET | `/get/@:name` | 同上(兼容) |
| GET | `/record/@:name` | 返回 JSON 计数 |
| GET | `/heart-beat` | 健康检查 |
| GET | `/api/themes` | 主题列表 |
| POST | `/api/themes` | 上传主题 |
| GET | `/api/backgrounds` | 底图列表 |
| POST | `/api/backgrounds` | 上传底图 |

详见 [docs/projectDesign.md](./docs/projectDesign.md)。

## 文档

| 文档 | 内容 |
|---|---|
| [docs/architecture.md](./docs/architecture.md) | 架构、项目结构、技术选型 |
| [docs/deployment.md](./docs/deployment.md) | 使用与部署(Win/Mac/Linux) |
| [docs/projectDesign.md](./docs/projectDesign.md) | 项目设计与接口契约 |
| [docs/TODOlist.md](./docs/TODOlist.md) | 里程碑与任务状态 |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | 贡献总览 |
| [docs/contributing-themes.md](./docs/contributing-themes.md) | 主题贡献指南 |
| [docs/contributing-code.md](./docs/contributing-code.md) | 功能贡献指南 |

## 贡献主题

Lolicount 采用**帧式主题**:每个主题是一个目录,内含若干帧图片
`0.<ext> 1.<ext> ... n-1.<ext>`,访问计数按 `(count+1) % n` 轮播展示。
扩展名支持 `gif` / `png` / `webp`,帧索引必须从 0 连续递增。

两种贡献方式:

**PR 通道** — fork 仓库,在 `assets/theme/<your-theme>/` 放入帧图片
(至少 1 帧,索引从 0 连续)与可选 `meta.json`,提 PR。CI 自动运行:

- `cmd/check-theme` 校验目录名、帧完整性、格式与尺寸
- `scripts/validate-theme-meta.js` 校验 `meta.json` schema
- `scripts/gen-themes-json.js` 校验 `assets/themes.json` 已同步

**Web 上传** — 访问 `/upload` 页面,上传帧图片,立即可用(服务端重编码)。

`meta.json` 示例:

```json
{
  "name": "lian",
  "author": "yourname",
  "description": "Loli-style digit frames",
  "tags": ["cute", "anime"],
  "version": "1.0.0"
}
```

本地预校验:

```bash
go run ./cmd/check-theme
node scripts/validate-theme-meta.js
node scripts/gen-themes-json.js
```

## CI/CD 与部署

| 工作流 | 触发 | 作用 |
|---|---|---|
| `ci.yml` | push / PR | go vet + `go test -race` + check-theme + Nuxt build |
| `theme-check.yml` | PR 改动 `assets/theme`、`assets/bg` | 主题完整性 + meta.json + themes.json 同步 |
| `release.yml` | tag `v*` | 构建 Docker 镜像 + Release 二进制 |
| `rebuild-frontend.yml` | 默认分支主题变更 | 重建 SSG dist 并提交 |

**Docker**:`docker compose up -d`,访问 `http://localhost:9721/@my-counter`。
**Release**:打 tag `git tag v0.1.0 && git push --tags`,CI 自动构建镜像并发布 Release。

## 致谢

- [kun-galgame-forum](https://github.com/KunMoe/kun-galgame-forum) — 立绘角色分层叠加的灵感来源
- [Moe-Counter](https://github.com/journey-ad/Moe-Counter) — 本项目参考的 Moe-Counter 原作

## 赞助

如果 Lolicount 对你有帮助,欢迎[赞助作者](https://github.com/sponsors/miaoledor) 🧋

## 技术栈

- **后端**:Go 1.23+ / Fiber v3 / SQLite(`modernc.org/sqlite`,纯 Go 免 CGO)
- **前端**:Vue(计划升级 Nuxt 3 SSG)/ UnoCSS / GSAP
- **存储**:请求 → 内存 Buffer → 定时批量写 → SQLite
- **部署**:单二进制(embed.FS 打包主题 + 前端 dist)

## 开源协议

本项目基于 [AGPL-3.0](./LICENSE) 协议开源。

This project is licensed under the AGPL-3.0 license.
