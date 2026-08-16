# Lolicount

> A cute, themeable SVG visitor counter — pick a built-in theme or upload your own, then paste a link and watch it count.

萌系可换肤访问计数器,以 SVG 图片形式输出。内置多套主题,也可上传自己的数字图或底图打造专属风格。往 README 或主页贴一行链接,每次访问数字自然 +1。

## 特性

- 🎀 **萌系主题** — 内置萝莉风格数字图,支持 gif/png/webp
- 🎨 **可换肤** — 50+ 内置主题,或上传自己的 `0~9` 字形图
- 🖼️ **底图叠加** — 把计数器叠加到任意底图上(立绘、徽章、海报)
- 📊 **SVG 输出** — 矢量清晰,嵌入 README 即可,无需 JS
- ⚡ **高性能** — Go + Fiber v3,内存/Redis 存储,批量落库
- 🛡️ **限流防护** — IP 级 + name 级双层限流,防刷防注水
- 🚀 **单二进制** — 前端 + 主题 embed 进 Go,一次构建到处部署
- 🤝 **社区贡献** — PR 通道(CI 自动校验)+ Web 上传通道

## 快速开始

### Docker

```bash
docker run -d -p 3000:3000 \
  -e STORAGE_TYPE=memory \
  ghcr.io/yourname/lolicount:latest
```

访问 `http://localhost:3000/@my-counter` 即可。

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
![visitor](https://lolicount.app/@my-counter?theme=loli&padding=7)
```

带底图叠加:

```markdown
![visitor](https://lolicount.app/@my-counter?bg=loli-stand&theme=loli&x=20&y=180)
```

## 参数

| 参数 | 说明 | 默认值 |
|---|---|---|
| `theme` | 主题名,或 `random` | `loli` |
| `bg` | 底图名(可选) | 无 |
| `x` `y` | 数字在底图上的坐标 | `0` `0` |
| `align` | 数字对齐:top/center/bottom | `top` |
| `padding` | 位数补零 | `7` |
| `offset` | 字间距 | `0` |
| `scale` | 缩放 | `1` |
| `darkmode` | 0/1/auto | `auto` |
| `pixelated` | 像素化渲染 | `1` |
| `num` | 指定数字(不落库) | `0` |
| `prefix` | 前缀数字 | `-1` |

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

详见 [Detail.md](./Detail.md)。

## 贡献主题

两种方式:

**PR 通道** — fork 仓库,在 `assets/theme/<your-theme>/` 放入 `0~9` 图片 + `meta.json`,提 PR。CI 自动校验。

**Web 上传** — 访问 `/upload` 页面,上传 10 张图,立即可用。

详见 [贡献指南](./CONTRIBUTING.md)。

## 技术栈

- **后端**:Go 1.23+ / Fiber v3 / Redis / SQLite