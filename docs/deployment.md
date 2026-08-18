# 使用与部署

> 本文档覆盖日常使用、开发模式与生产部署,适用于 Windows / macOS / Linux。
> 架构与技术选型见 [架构文档](./architecture.md)。

## 前置要求

| 工具 | 最低版本 | 用途 |
|---|---|---|
| Go | 1.23+ | 后端编译 |
| Node.js | 20+ | 前端构建 |
| pnpm | 9+ | 前端依赖管理 |
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
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
# Node + pnpm
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs git
npm install -g pnpm
```

## 快速使用

拿到一个运行中的 Lolicount 实例后,在 README / 网页里嵌入即可:

```markdown
![visitor](https://umi7.top/@my-counter?theme=lian)
```

三种嵌入方式见 [README](../README.md#使用)。

## 开发模式(三端通用)

根目录 `package.json` 用 `concurrently` 同时启动后端(Go :9721)和前端
(Nuxt :3721),脚本本身跨平台,无需为不同系统改命令:

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

### 环境变量

复制 `.env.example` 为 `.env` 并按需填写:

| 变量 | 说明 | 默认 |
|---|---|---|
| `PORT` | 后端端口 | `9721` |
| `DB_INTERVAL` | 缓冲批量落库间隔(秒) | `5` |
| `TRUST_PROXY` | 信任代理(X-Forwarded-For) | `true` |
| `R2_*` / `S3_*` | 对象存储(底图 CDN) | 可选 |

**永远不要提交 `.env`**。只提交 `.env.example`。

## 构建生产二进制

前端先构建 SSG,再编译 Go 二进制(embed 打包前端 dist + 主题):

```bash
# 1. 构建前端静态产物 → web/.output/public (软链到 web/dist)
pnpm build:web

# 2. 编译后端(三端各自的目标)
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

```bash
docker run -d -p 9721:9721 \
  -v lolicount-data:/app/data \
  ghcr.io/miaoledor/lolicount:latest
```

或用 docker compose:

```bash
docker compose up -d
```

计数数据持久化到 `lolicount-data` 卷的 SQLite 文件。

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
