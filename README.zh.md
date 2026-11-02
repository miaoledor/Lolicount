![Lolicount](docs/png/lolicount-s.png)

[English](./README.md) · **中文** · [日本語](./README.ja.md)

为 GitHub README、博客或个人主页添加一个萌系访问计数器。选好角色，复制嵌入代码即可使用。静态主题支持每次请求随机换图或组合角色图层。

## 开始使用

1. 打开 [lolicount.top](https://lolicount.top)。
2. 填写计数器名称，选择主题，点击生成。
3. 复制 Markdown 或 HTML 代码，粘贴到你的页面。

将下面的 `my-counter` 换成自己的名称。同名计数器共享计数，在多个页面使用同一个名称会合并计数。

**GitHub README / Markdown**

```markdown
![访问量](https://lolicount.top/@my-counter?theme=lian-ren&fsize=16&scale=0.6)
```

**网页 / HTML**

```html
<img src="https://lolicount.top/@my-counter?theme=lian-ren&amp;fsize=16&amp;scale=0.6" alt="访问量" />
```

**图片链接**：[打开示例](https://lolicount.top/@my-counter?theme=lian-ren&fsize=16&scale=0.6)，也可复制其地址，填入支持外部图片链接的位置。

GitHub README 请选择 SVG 主题。E-mote、Spine、Live2D 动画主题使用网站生成的网页挂件代码，需要页面支持 JavaScript。

## 调整样式

直接修改链接参数，或在网站上调整并预览：

| 参数 | 用途 | 示例 |
|---|---|---|
| `theme` | 角色主题；`random` 随机选择 SVG 主题 | `theme=lian-ren` |
| `scale` | 图片缩放，范围 `0.1`～`4`，默认 `1` | `scale=0.6` |
| `fsize` | 计数字号，单位为像素，范围 `0`～`500`，`0` 使用默认字号 | `fsize=16` |
| `number` | 固定展示 `1`～`999999` 的数字，不计数；正常计数时不传 | `number=12345` |

只想预览时，将名称换成 [`demo`](https://lolicount.top/@demo?theme=lian-ren&fsize=16&scale=0.6)，默认展示 `0123456789`，不会写入计数；设置 `number` 可预览指定数字。

字体风格、文字位置等用法见[完整参数说明](./docs/projectDesign.md#接口调用文档)，遇到问题可查看[常见问题](./docs/faq.md)。

## 自行部署

使用 Docker 启动：

```sh
docker run -d --name lolicount --restart unless-stopped -p 9721:9721 -v lolicount-data:/app/data ghcr.io/miaoledor/lolicount:latest
```

打开 [localhost:9721](http://localhost:9721) 即可生成计数器。数据保存在 `lolicount-data` 卷中。请保持单实例运行；使用公网域名时，将 `BASE_URL` 设为对应地址，例如 `https://counter.example.com`。

Docker Compose、环境配置和二进制构建见[部署文档](./docs/deployment.md)。

## 本地开发

安装 Go 1.25+、Node.js 22+ 和 pnpm 11 后，在仓库根目录运行：

```sh
pnpm install
pnpm dev
```

前端访问 [localhost:3721](http://localhost:3721)，后端运行于 [127.0.0.1:9721](http://127.0.0.1:9721)。

## 参与与支持

- [贡献指南](./CONTRIBUTING.md) · [添加主题](./docs/contributing-themes.md) · [贡献代码](./docs/contributing-code.md)
- [文档导航](./docs/index.md) · [项目架构](./docs/architecture.md) · [任务计划](./docs/TODOlist.md)
- 主题仅通过仓库 Pull Request 贡献，流程见[添加主题](./docs/contributing-themes.md)。
- [请作者喝杯奶茶](https://github.com/sponsors/miaoledor) 🧋

感谢 [kun-galgame-forum](https://github.com/KunMoe/kun-galgame-forum) 和 [Moe-Counter](https://github.com/journey-ad/Moe-Counter)。

本项目基于 [AGPL-3.0](./LICENSE) 协议开源。

![Lolicount](https://lolicount.top/@lolicount?theme=lian-ren&fsize=16&scale=1&unshowf=true)
