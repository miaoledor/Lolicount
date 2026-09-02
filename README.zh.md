![miaoledor](docs/png/githubSocialPreview.png)
[English](./README.md) · **中文** · [日本語](./README.ja.md)


### 在你的主页或者支持外部图片源的位置展示你喜欢的角色！

萌系可换肤访问计数器,以 SVG 图片形式输出。内置多套主题,也可上传自己的数字图或底图打造专属风格。只需往 README 或主页贴一行链接！ 

展示的角色支持随机抽帧和随机分层，支持类似gal中角色立绘的`动态拼接`

## 少量素材，万种变化——以极低的存储成本带来最丰富变化的主题

多图层主题将角色拆分为表情、眼睛、嘴巴、脸部等独立图层，每次请求随机组合。以 `lian-ren` 主题为例，仅用 **71 张图**（lass ×8 + brow ×18 + eye ×18 + mouth ×20 + face ×6）就能拼出 **311,040 种**不同的立绘组合——每次刷新都是一个全新的姿态。

## 快速加载

Fiber v3 基于 [Fasthttp](https://gofiber.io/) 构建，Fiber 官方称其为 Go 最快的 HTTP 引擎。在 [Fiber 公布的 TechEmpower 基准测试](https://docs.gofiber.io/extra/benchmarks/)中，Fiber v3 的 Plaintext 吞吐约为 **1198.8 万响应/秒**，Express 约为 **120.5 万**；JSON 序列化约为 **236.3 万响应/秒**，Express 约为 **94.97 万**。

作为更广的参照，[TechEmpower Round 23](https://www.techempower.com/benchmarks/#section=data-r23&test=plaintext&hw=ph) 在同一物理硬件上，以最佳 15 秒 Plaintext 运行对比了多款热门框架：

| 框架 | Plaintext RPS |
|---|---:|
| ASP.NET Core | 2753 万 |
| Fiber v2.52.5 prefork | 1351 万 |
| Gin | 164 万 |
| Spring | 83 万 |
| Express | 28 万 |
| Laravel | 2.7 万 |
| Next.js | 0.2 万 |

Round 23 仍使用 Fiber v2.52.5，因此这一行更适合作为生态参照，而不是 Fiber v3 的直接测量。

在 Lolicount 中，这条轻量请求路径还会配合嵌入的主题素材和 Nuxt SSG 前端。计数 SVG 在进程内合成，无需额外请求主题素材，因此能够快速响应。

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

喜欢该项目，如果 Lolicount 对你有帮助,欢迎[请作者喝一杯奶茶](https://github.com/sponsors/miaoledor) 🧋

## 技术栈

**后端**:Go 1.25+ / Fiber v3 / SQLite
**性能**:Fiber v3 提供快速请求处理，让计数 SVG 在嵌入流量下也能快速响应。
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

![lolicount](https://lolicount.top/@lolicount?theme=lian-ren&fsize=16&scale=1&unshowf=true)
