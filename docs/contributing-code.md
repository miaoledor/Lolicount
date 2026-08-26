# 功能贡献指南

> [主题贡献指南](./contributing-themes.md) · [返回贡献总览](../CONTRIBUTING.md)

本指南面向修改 Go / 前端 / CI / 文档的功能贡献。开始前请先通读
`AGENTS.md`,其中「铁律」与「Core Engineering Principles」是硬约束。

## 项目结构

按职责切包(domain-oriented),依赖方向单向:
`internal/server`(HTTP/编排)→ `counter` / `imgcore`(渲染)→ `store`。
`imgcore` 内三个 drawer 互不 import,仅由 `renderer` 合成。出现循环依赖说明分层错了,先修依赖方向再加功能。

```
cmd/server/         入口
cmd/check-theme/    主题校验工具
cmd/fix-theme/      卡片主题帧序号修复工具
internal/config/    环境变量加载
internal/logger/    zerolog 封装
internal/server/    Fiber 路由、handler、中间件
internal/counter/   内存缓冲 + 定时批量落库
internal/store/     SQLite repository
internal/imgcore/   渲染核心(card/character/fdrawer + renderer)
internal/ratelimit/ IP / name 限流
internal/themetool/ 主题元数据工具
web/                Nuxt 4 前端(SSG)
assets/             主题、立绘、字体素材、前端 dist
scripts/            主题校验 / 生成 / 构建 / 图片优化脚本
.github/workflows/  CI/CD
```

## 技术选型(已定,勿擅自换库)

- 后端:Go 1.25+、Fiber v3、go-playground/validator、envconfig + godotenv、zerolog、
  `modernc.org/sqlite`(纯 Go,免 CGO)、golang.org/x/image/webp、embed.FS
- 前端:Nuxt 4(SSG)、Vue 3、UnoCSS、GSAP、clsx、pnpm
- 存储:请求 → `counter.Buffer`(内存)→ `time.Ticker` 批量 → SQLite

换同类库(如 Fiber→Gin、zerolog→slog、modernc.org/sqlite→mattn/go-sqlite3)
需先和 maintainer 确认。

## 开发流程

```bash
pnpm install        # 安装 concurrently + 前端依赖
pnpm dev            # 同时启动前后端(跨平台 macOS/win/linux)
```

- 后端: http://127.0.0.1:9721
- 前端: http://localhost:3721

单独运行:`pnpm dev:server` 或 `pnpm dev:web`。

## 测试

- Go 测试用独立测试 DSN,不要从 `.env` 取、不要打印 DSN
- 共享库集成测试:`go test -race -count=1 -p 1 ./...`
- 主题相关测试不应依赖具体主题内容或数量(用 fixture / mock)
- 新增功能后判断是否需要单元测试;测试与实现可分两个 commit

## Git 提交规范

为保证代码历史可读、可理解,贡献代码时请遵循以下 Git 规范。

### 清晰的 commit message

内部采用类似 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
的格式:第一部分标明改动范围(scope),第二部分说明改动内容。

```
<type>(<scope>): <简短说明>
```

- **type**:`feat` / `fix` / `docs` / `refactor` / `test` / `chore` 等
- **scope**:标明改动的部分,不严格限定,写最贴切的即可。若改动既涉及大模块
  又涉及其中小部分,可同时写大小 scope 以便理解,例如改 internal streams
  相关逻辑可写 `api/stream: fix object not being handled properly`
- commit 标题用英文;如需进一步说明,鼓励在正文写简短解释(说明「为什么」改),
  这能省去来回询问的时间。正文可附中文说明

示例:

```
feat(counter): add buffer overflow guard

当 len(cache) > 10000 时降级只读并告警,防极端流量撑爆内存。
Ref: AGENTS.md Iron Rule 5.
```

> 如果 commit 标题信息不足,可能会被要求 interactively rebase 并 amend 每个
> commit 以补充有意义的标题。

### 清晰的 commit 历史

- **一个 commit 只做一件事**,保持功能单一,便于 review 与回滚
- **分支过期或有冲突时用 rebase,不要 merge**:避免无意义的 merge commit
  进入历史
- **发现已提交的代码有错,不要新开一个 commit 修**,而是修正引入错误的那个
  commit:
  - 若是该分支最新 commit:`git add` 暂存后 `git commit --amend`
  - 若在更深处:`git commit --fixup=HASH`(HASH 为出错的 commit),再
    `git rebase -i current --autosquash`(打开编辑器后直接保存退出即可)
  - 之后需 `git push --force-with-lease` 强推到自己的分支
- **只在本地 commit**:不要 push 到主仓库(由 maintainer 统一推送);fork
  PR 场景下推送到自己的 fork 分支
- **不要提交 `.env`**:它含密钥(R2/S3),只提交 `.env.example`

## 改动检查清单

每次改动后检查是否有意外副作用,尤其三处:

- [ ] **限流**:IP 级(429)与 name 级(降级只读)是否仍各自独立
- [ ] **缓存**:计数 SVG 仍 `no-store`,demo 仍长缓存(铁律 1)
- [ ] **存储**:仍是「请求→内存 Buffer→定时批量→SQLite」单一路径(铁律 5)

涉及 `tb_count` schema 变更时,任务结尾必须告诉用户是否需要迁移、跑哪个命令、
对哪个库。

---

下一步:→ [主题贡献指南](./contributing-themes.md)
