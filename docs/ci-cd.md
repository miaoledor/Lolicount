# CI/CD

## 当前 CI 失败原因

2026-11-08 的 [CI #157](https://github.com/miaoledor/Lolicount/actions/runs/42460595499)
被 GitHub 以「账号因账单问题锁定」拒绝启动。三个任务的步骤列表均为空，因此没有测试或构建日志。
这与代码编译错误不同，需要账号所有者先在 GitHub 侧解除锁定；修改 YAML 或重跑旧任务无法解除锁定。

同次运行还提示 `ubuntu-latest` 正在迁移至 Ubuntu 26。仓库工作流现固定为 `ubuntu-24.04`，
避免运行器环境随标签迁移发生变化；该提示本身不是本次失败原因。

## 构建与验证流程

| 工作流 | 触发 | 内容 |
|---|---|---|
| `ci.yml` | main/master push、PR、手动运行 | 前端测试与 SSG；复用产物执行 Go vet 和完整 race 测试；独立主题完整性检查 |
| `theme-check.yml` | 主题资源 PR | Go 主题校验、meta.json 校验和 themes.json 同步检查 |
| `rebuild-frontend.yml` | 主分支主题资源变更 | 重建清单、测试并生成前端，提交更新后的 themes.json；不提交静态构建目录 |
| `release.yml` | `v*` tag push | 使用 Dockerfile 构建并推送 GHCR 镜像，不自动发布 GitHub Release |
| `release-assets.yml` | 手动发布 Release 或手动指定 tag | 使用相同 Dockerfile 构建镜像，打包 tar.gz 并上传到对应 Release |

CI 中的 `Nuxt build` 任务负责一次性完成：

1. 根据根 `package.json` 的 `packageManager` 安装 pnpm，执行 frozen-lockfile 安装。
2. 运行 `pnpm test:frontend`，覆盖 About 统计与 Support 物理 / 多语言测试。
3. 生成 `web/.output/public/`，确认首页、`/about`、`/support` 均有预渲染文件。
4. 上传 `frontend-dist`，保留一天，支持同一次运行重跑时替换同名产物。

`Go vet + test` 依赖该任务，将同一次运行的产物下载到 `assets/dist/` 后再编译和测试，
避免重复安装 Node 依赖、重复构建前端，也保证页面回归实际使用嵌入的构建产物。
下载与上传均使用 [GitHub 官方 Artifact Actions](https://github.com/actions/upload-artifact)。

Node 保持 22，Go 根据 `go.mod` 安装；pnpm 的唯一版本声明位于根 `package.json`，
GitHub Actions 与 Docker Corepack 均读取它。相关 checkout、Node、Go、pnpm 和 artifact Actions
已更新为验证过输入定义的版本。

## 发布镜像

`Dockerfile` 的前端阶段安装锁定依赖、运行同一组前端测试并生成 SSG。
后端阶段直接将 `/app/web/.output/public` 复制至 `assets/dist/` 后构建 Go 二进制。
本地 `.output`、`dist`、Node 依赖与已有 `assets/dist` 均排除在 Docker 构建上下文之外。

`release.yml` 不再提前构建一个未被上传的宿主机二进制，前端与 Go 编译统一由 Dockerfile 完成。
`BASE_URL` 继续通过 build arg 传入，GHCR 标签和手动发布 Release 的行为保留。

## 本地验证

```sh
pnpm install --frozen-lockfile
pnpm test:frontend
pnpm --dir web generate
node -e "require('node:fs').cpSync('web/.output/public', 'assets/dist', { recursive: true })"
go vet ./...
go test -race -count=1 -p 1 ./...
go run ./cmd/check-theme
node scripts/validate-theme-meta.js
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
```

单独验证 Docker 前端阶段：

```sh
docker build --target frontend -t lolicount-frontend-check .
```

完整镜像验证仍可按 [部署文档](./deployment.md) 执行 `docker build`。
此次验证已通过工作流 lint、21 项前端测试、SSG 三个入口检查、完整 Go vet / race、主题与元数据校验，
以及干净 Docker 前端阶段构建。没有发布镜像，也没有执行远端工作流；完整 Docker 运行镜像未在此次重建。

若本机 Docker 继承了指向 `127.0.0.1` 的代理，容器中的该地址表示容器自己，可能导致
`ECONNREFUSED`。可仅在本地验证命令中使用容器能够访问的宿主机代理地址，不把个人代理写入项目配置。

## 恢复远端验证

1. 账号所有者解除 GitHub 账单锁定。
2. 修复提交进入远端主分支后，在 Actions 运行新的 CI；也可通过新增的 `workflow_dispatch` 手动触发。
3. 确认前端产物上传、Go 下载及完整测试三个阶段都有实际执行步骤，三个 CI 任务均成功。
4. 若超过一天才重跑 Go 任务，旧前端产物可能已经过期，应重跑完整工作流。
5. 正常发布时再验证 tag → GHCR，以及手动 Release → 镜像压缩包两条发布路径。

本次不修改数据库 schema、缓存 / 限流 / 存储逻辑，**无需数据库迁移**。
