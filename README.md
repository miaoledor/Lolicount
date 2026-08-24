<p align="center"><img src="docs/png/lolicount-icon.png" width="120" alt="Lolicount"></p>

<h1 align="center">Lolicount !</h1>

![miaoledor](docs/png/nbg2.png)
**中文** · [English](./README.en.md) · [日本語](./README.ja.md)


### 在你的主页或者支持外部图片源的位置展示你喜欢的角色！

萌系可换肤访问计数器,以 SVG 图片形式输出。内置多套主题,也可上传自己的数字图或底图打造专属风格。只需往 README 或主页贴一行链接！ 

展示的角色支持顺序模式和随机模式，支持类似gal中角色立绘的`动态拼接`


## 快速开始

### 直接使用
查看 https://lolicount.top

### 测试开发运行

根目录 `package.json` 用 `concurrently` 同时启动后端(Go :9721)和前端(Nuxt :3721),跨平台兼容 macOS / Windows / Linux:

```bash
pnpm install        # 安装 concurrently(根目录)与前端依赖
pnpm dev            # 同时启动前后端
```
也可单独运行:`pnpm dev:server`(仅后端)或 `pnpm dev:web`(仅前端)。

### 服务器部署

```bash
docker run -d -p 9721:9721 \
  -v lolicount-data:/app/data \
  ghcr.io/miaoledor/lolicount:latest
```

访问 `http://localhost:9721/@my-counter` 即可。计数数据持久化到 `lolicount-data` 卷的 SQLite 文件。

项目通过 GitHub Actions 实现 CI/CD:推送代码自动运行 `go vet` + 测试,打 `v*` 标签自动构建前端、编译静态二进制并推送 Docker 镜像到 ghcr.io。

## 贡献

我们真的非常需要你的帮助！

无论是功能的丰富 或者是主题的添加 都需要你的参与
更多贡献的`细节`可以查看：
| 文档 | 内容 |
|---|---|
| [CONTRIBUTING.md](./CONTRIBUTING.md) | 贡献总览 |
| [docs/contributing-themes.md](./docs/contributing-themes.md) | 主题贡献指南 |
| [docs/contributing-code.md](./docs/contributing-code.md) | 功能贡献指南 |

## 致谢

- [kun-galgame-forum](https://github.com/KunMoe/kun-galgame-forum)
- [Moe-Counter](https://github.com/journey-ad/Moe-Counter)

## 赞助

喜欢该项目，如果 Lolicount 对你有帮助,欢迎[请作者喝一杯奶茶] 🧋

## 技术栈

**后端**:Go 1.25+ / Fiber v3 / SQLite
**前端**:Vue(Nuxt 4 SSG)/ UnoCSS / GSAP
**存储**:请求 → 内存 Buffer → 定时批量写 → SQLite
**部署**:单二进制(embed.FS 打包主题 + 前端 dist)
更多的技术细节可以在以下文档中查看：
| 文档 | 内容 |
|---|---|
| [docs/architecture.md](./docs/architecture.md) | 架构、项目结构、技术选型 |
| [docs/deployment.md](./docs/deployment.md) | 使用与部署(Win/Mac/Linux) |
| [docs/projectDesign.md](./docs/projectDesign.md) | 项目设计与接口契约 |
| [docs/TODOlist.md](./docs/TODOlist.md) | 里程碑与任务状态 |

## 开源协议

本项目基于 [AGPL-3.0](./LICENSE) 协议开源。

This project is licensed under the AGPL-3.0 license.

![miaoledor](https://lolicount.top/@miaoledor?theme=lian&fsize=16&scale=1&unshowf=true&mode=seq)
