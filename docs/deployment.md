# 使用与部署

> 本文档覆盖日常使用、开发模式与生产部署,适用于 Windows / macOS / Linux。
> 架构与技术选型见 [架构文档](./architecture.md)。

## 前置要求

| 工具 | 最低版本 | 用途 |
|---|---|---|
| Go | 1.25+ | 后端编译(以 `go.mod` 为准) |
| Node.js | 22+ | 前端构建(与 Docker/CI 一致) |
| pnpm | 11+ | 前端依赖管理(与 Docker/CI 一致) |
| Git | 任意 | 克隆仓库 |
| Docker(可选) | 任意 | 容器化部署 |

三端安装示例:

**macOS(Homebrew)**
```bash
brew install go node pnpm git
```

**Windows(Scoop)**
```powershell
scoop install go nodejs pnpm git
```

**Linux(apt,以 Debian/Ubuntu 为例)**
```bash
# Go 从官方获取最新版
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
# Node + pnpm
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs git
npm install -g pnpm
```

## 快速使用

拿到一个运行中的 Lolicount 实例后,在 README / 网页里嵌入即可:

```markdown
![visitor](https://lolicount.top/@my-counter?theme=lian)
```

三种嵌入方式见 [README](../README.md)。

## 开发模式(三端通用)

根目录 `package.json` 用 `concurrently` 同时启动后端(Go :9721)和前端
(Nuxt :3721),脚本本身跨平台,无需为不同系统改命令:

根目录的 pnpm workspace 包含 `web`,执行一次 `pnpm install` 即可安装开发工具和前端依赖。

```bash
git clone https://github.com/miaoledor/Lolicount.git
cd Lolicount
pnpm install        # 安装 concurrently(根目录)+ 前端依赖
cp .env.example .env
pnpm dev            # 同时启动前后端
```

- 后端: http://127.0.0.1:9721
- 前端: http://localhost:3721

单独运行:

```bash
pnpm dev:server     # 仅后端
pnpm dev:web        # 仅前端
```

> 跨平台说明:`dev:web` 脚本通过 `pnpm --dir web dev` 启动 Nuxt,
> `dev:server` 先跑 `node scripts/preflight-port.js` 清理占用端口再
> `go run`。两者在 Windows / macOS / Linux 上行为一致。

### 端口模型:开发 vs 生产

本地开发与服务器部署的端口拓扑不同,这是有意设计:

| | 前端(Nuxt) | 后端(Go) | 说明 |
|---|---|---|---|
| **本地开发** | `3721` | `9721` | 前后端分离运行,Nuxt dev server 独立监听 3721,通过 `NUXT_PUBLIC_API_BASE=http://127.0.0.1:9721` 跨端口请求后端 API |
| **服务器部署** | — | `9721` | 前端经 `nuxt generate` 产出静态 SSG dist,由 Go 二进制 `embed.FS` 打包并直接 serve,前后端统一在 9721,无独立前端进程 |

**为什么开发时分端口**:Nuxt 的 dev server(HMR / SSR 调试)需要一个独立
进程,不能被 embed 进 Go 二进制;若让 Nuxt 与 Go 共用 9721 会端口冲突(Nuxt
CLI 会读 `PORT` 环境变量作为自己的监听端口)。因此 `scripts/dev-web.js`
显式设置 `NUXT_PORT=3721` 把 Nuxt 钉在 3721,后端独占 9721。

**为什么生产时统一端口**:生产用 SSG(`nuxt generate`),前端只是静态文件
(HTML/JS/CSS),没有独立进程,直接被 Go 二进制 serve。用户访问
`http://your-host:9721/` 拿到前端页面,访问 `/api/*`、`/@:name` 拿到 API
与计数 SVG,全部同源,无需跨端口。


> 生产对外暴露时,设 `HOST=0.0.0.0` 让 Go 监听所有网卡(见环境变量表),
> 再用 Nginx/Caddy 反代 80/443 到 9721,或直接暴露 9721。

**前端接口调用 vs 生成链接**:web 前端调用后端接口(`/api/themes`、
`/api/fthemes`、`/api/config`、计数 SVG 预览)**始终走本地**——dev 模式下
是 `http://127.0.0.1:9721`(`NUXT_PUBLIC_API_BASE`),SSG 生产模式下是同源
相对路径(空 `apiBase`),绝不会走外部域名。只有给用户复制粘贴的**嵌入链接**
用 `BASE_URL`(`https://lolicount.top`,不带端口)。这样接口请求不依赖外部域名
可达性,而嵌入链接指向公网域名。



### 环境变量

复制 `.env.example` 为 `.env` 并按需填写:

| 变量 | 说明 | 默认 |
|---|---|---|
| `HOST` | 监听地址,默认仅本机;部署对外需设 `0.0.0.0` | `127.0.0.1` |
| `PORT` | 后端端口 | `9721` |
| `LOG_LEVEL` | 日志级别(trace/debug/info/warn/error) | `info` |
| `DB_PATH` | SQLite 数据库文件路径 | `data/count.db` |
| `DB_INTERVAL` | 缓冲批量落库间隔(秒),生产建议 5~10 | `10` |
| `TRUST_PROXY` | 信任代理头(X-Forwarded-For),同机反代默认信任 | `true` |
| `TRUST_PROXY_PRIVATE` | 额外信任私网段(跨机反代时开启) | `false` |
| `BASE_URL` | 公开域名(用于 web 嵌入链接),不带端口,如 `https://lolicount.top`;空=用请求自身 origin | 空 |
| `RATE_LIMIT_IP_PER_SEC` / `RATE_LIMIT_IP_PER_MIN` | IP 限流,超限 429 | `60` / `3000` |
| `RATE_LIMIT_NAME_PER_SEC` | name 限流,超限降级只读 | `20` |
| `PSB_DIR` | E-mote PSB 模型磁盘目录 | `assets/psb` |

> 仓库中的 `docker-compose.yml` 示例显式使用更严格的 `10/s`、`300/min`、
> `5/s/name`;删除这些覆盖项后才会回到上表的程序默认值。

**永远不要提交 `.env`**。只提交 `.env.example`。

## 构建生产二进制

生产模式下前端无独立进程:先 `nuxt generate` 产出静态 SSG dist,再编译
Go 二进制把 dist + 主题图 `embed.FS` 打包进去,由 Go 单进程在 9721 同源
serve 前端页面与后端 API(见上文「端口模型」)。

前端先构建 SSG,再编译 Go 二进制(embed 打包前端 dist + 主题):

```bash
# 1. 构建前端静态产物 → web/.output/public
pnpm build:web

# 2. 将前端产物放入 Go embed 目录(Win/Mac/Linux 通用)
node -e "require('node:fs').cpSync('web/.output/public', 'assets/dist', { recursive: true })"

# 3. 编译后端(三端各自的目标)
# macOS / Linux
go build -o lolicount ./cmd/server
# Windows (PowerShell / cmd)
go build -o lolicount.exe ./cmd/server
```

运行:

```bash
# macOS / Linux
./lolicount
# Windows
.\lolicount.exe
```

访问 `http://localhost:9721/@my-counter`。

> **内存要求**:当前采用“每槽位常驻一张”的懒加载,启动只读取候选元数据并
> 为每个槽位解码一张图片,不会把全部主题图预加载为 base64。2026-09-30 的
> 284 主题快照中稳定 RSS 约 0.85 GiB、匿名堆约 0.13 GiB;主题持续增长后会
> 相应增加。生产环境仍建议预留 **≥ 2 GiB 内存**;详细构成与估算方法见
> [架构文档「图片内存策略与占用」](./architecture.md#图片内存策略与占用)。
>
> **E-mote(PSB)模型目录**:模型文件不打包进二进制(444 MB,避免常驻内存),
> 由 `PSB_DIR`(默认 `assets/psb`,相对工作目录)从磁盘按需流式读取。启动不预读
> 模型字节;请求结束即关闭文件流,服务端不保留模型缓存。浏览器以不携带 Cookie 的
> 私有 immutable 请求在本地缓存模型,共享代理不保存。源码
> 部署保持在仓库目录下运行即可;Docker 镜像已在 `/app/assets/psb` 自带
> (可用 `PSB_DIR` 指向挂载卷替换)。目录缺失时挂件端点优雅降级为空列表,
> 不影响启动。

### 交叉编译

`modernc.org/sqlite` 是纯 Go,免 CGO,因此可交叉编译:

```bash
# 从 macOS 编译 Linux amd64
GOOS=linux GOARCH=amd64 go build -o lolicount-linux-amd64 ./cmd/server
# 编译 Windows
GOOS=windows GOARCH=amd64 go build -o lolicount-windows-amd64.exe ./cmd/server
# 编译 macOS arm64
GOOS=darwin GOARCH=arm64 go build -o lolicount-darwin-arm64 ./cmd/server
```

## Docker 部署

> **关于 `BASE_URL` 与 Docker**:后端在 serve `index.html` 时会**运行时**
> 把 `BASE_URL` 环境变量注入到 SSG 的 `__NUXT__` payload,所以**无需重新构建
> 镜像**,改域名只需改环境变量再重启容器即可(build once, configure per env,
> 与 kun-galgame-forum 的 SSR runtimeConfig 同理念)。构建期的 build arg 只
> 决定镜像里的默认值;运行时环境变量优先覆盖。

**预构建镜像 + 运行时配置域名(推荐)**:

```bash
docker run -d -p 9721:9721 \
  -v lolicount-data:/app/data \
  -e BASE_URL=https://lolicount.top \
  ghcr.io/miaoledor/lolicount:latest
```

镜像有两个自动来源(`.github/workflows/`):

- **ghcr.io**:推送 `v*` tag 时 `release.yml` 构建并推送
  `ghcr.io/miaoledor/lolicount:<version>`(另带 `latest` 与主/次版本 tag);
- **Release Assets**:发布 Release 时 `release-assets.yml` 构建镜像并打包为
  `lolicount-<tag>-docker-image.tar.gz` 挂到该 Release 的 Assets,`docker load`
  即可离线导入;发布后才想补挂,在 Actions 手动触发该 workflow 并填 tag。

两条发布路径都使用同一 Dockerfile，构建时执行前端测试并嵌入 `.output/public`。
CI 中前端只构建一次，再将静态产物交给 Go 检查。工具版本、任务依赖和失败排查见
[CI/CD 文档](./ci-cd.md)。

**docker compose**:在 `docker-compose.yml` 的 `environment` 设 `BASE_URL`,
或导出环境变量:

```bash
export BASE_URL=https://lolicount.top
docker compose up -d
```

换域名只需改 `BASE_URL` 后 `docker compose up -d` 重启,无需重新 build。

**带端口访问(绕过 80/443 审核)**:`BASE_URL` 可包含端口,例如直接暴露 9721
或用反向代理监听 9721 做 TLS 终结:

```bash
export BASE_URL=https://lolicount.top:9721
docker compose up -d
```

此时嵌入链接为 `https://lolicount.top:9721/@name?...`。注意:Go 二进制本身
只监听 HTTP,若需 HTTPS 需在前面加反向代理(Nginx/Caddy)监听 9721 并终结 TLS,
转发到容器的 9721。

**本地构建镜像(可选,固化默认域名)**:

```bash
docker build --build-arg BASE_URL=https://lolicount.top -t lolicount .
docker run -d -p 9721:9721 -v lolicount-data:/app/data lolicount
```

CI(`release.yml`)构建时也读 GitHub Variables 的 `BASE_URL` 作为默认值;
运行时的 `BASE_URL` 环境变量优先级更高。

计数数据持久化到 `lolicount-data` 卷的 SQLite 文件。若需在容器外维护 PSB
模型,另挂载目录并设置 `PSB_DIR`,例如
`-v /srv/lolicount/psb:/models:ro -e PSB_DIR=/models`。

## Release 流程

打 tag 触发 `release.yml`,CI 自动构建 Docker 镜像与多平台二进制:

```bash
git tag v0.1.0
git push --tags
```

## 升级与迁移

- SQLite schema 变更需要写迁移脚本,任务结尾必须告知用户是否需要执行
- 升级二进制前备份 `data/count.db`
- 前端 SSG dist 变更由 `rebuild-frontend.yml` 在主题变更时自动重建
