# Lolicount

> A cute, themeable SVG visitor counter — pick a built-in theme or upload your own, then paste a link and watch it count.



萌系可换肤访问计数器,以 SVG 图片形式输出。内置多套主题,也可上传自己的数字图或底图打造专属风格。往 README 或主页贴一行链接,每次访问数字自然 +1。

本项目只使用一张图片进行展示
比如default-theme[0.png 1.png 2.png 3.png ... size-1.png]
则每次显示(count+1)%size,count++

## 特性

- 🎀 **萌系主题** — 内置萝莉风格数字图,支持 gif/png/webp
- 🎨 **可换肤** — 50+ 内置主题,或上传自己的 `0~9` 字形图
- 🖼️ **底图叠加** — 把计数器叠加到任意底图上(立绘、徽章、海报)
- 📊 **SVG 输出** — 矢量清晰,嵌入 README 即可,无需 JS
- ⚡ **高性能** — Go + Fiber v3,内存缓冲 + 定时批量落库 SQLite
- 🛡️ **限流防护** — IP 级 + name 级双层限流,防刷防注水
- 🚀 **单二进制** — 前端 + 主题 embed 进 Go,一次构建到处部署
- 🤝 **社区贡献** — PR 通道(CI 自动校验)+ Web 上传通道

## 快速开始

### Docker

```bash
docker run -d -p 3000:3000 \
  -v lolicount-data:/app/data \
  ghcr.io/yourname/lolicount:latest
```

访问 `http://localhost:3000/@my-counter` 即可。计数数据持久化到 `lolicount-data` 卷的 SQLite 文件。

### 从源码

```bash
git clone https://github.com/yourname/lolicount.git
cd lolicount
cp .env.example .env
go run ./cmd/server
```

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

详见 [docs/detail.md](./docs/detail.md)。

## 贡献主题

两种方式:

**PR 通道** — fork 仓库,在 `assets/theme/<your-theme>/` 放入 `0~9` 图片 + `meta.json`,提 PR。CI 自动校验。

**Web 上传** — 访问 `/upload` 页面,上传 10 张图,立即可用。

详见 [贡献指南](./CONTRIBUTING.md)。

## 技术栈

- **后端**:Go 1.23+ / Fiber v3 / SQLite(`modernc.org/sqlite`,纯 Go 免 CGO)
- **前端**:Vue(计划升级 Nuxt 3 SSG)/ UnoCSS / GSAP
- **存储**:请求 → 内存 Buffer → 定时批量写 → SQLite
- **部署**:单二进制(embed.FS 打包主题 + 前端 dist)
