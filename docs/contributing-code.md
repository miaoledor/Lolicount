# 功能贡献指南

> [主题贡献指南](./contributing-themes.md) · [返回贡献总览](../CONTRIBUTING.md)

本指南面向修改 Go / 前端 / CI / 文档的功能贡献。开始前请先通读
`AGENTS.md`,其中「铁律」与「Core Engineering Principles」是硬约束。

## 项目结构

按职责切包(domain-oriented),依赖方向单向:
`internal/server`(HTTP/编排)→ `counter` / `theme` / `ftheme` → `store`。
出现循环依赖说明分层错了,先修依赖方向再加功能。

```
cmd/server/         入口
cmd/check-theme/    主题校验工具
internal/config/    环境变量加载
internal/logger/    zerolog 封装
internal/server/    Fiber 路由、handler、中间件
internal/counter/   内存缓冲 + 定时批量落库
internal/store/     SQLite repository
internal/theme/     主题注册与渲染(帧图模型)
internal/ftheme/    字体样式主题
internal/ratelimit/ IP / name 限流
internal/assets/    embed.FS 挂载
web/                Nuxt 3 前端(SSG)
assets/             主题、立绘、字体素材、前端 dist
scripts/            主题校验 / 生成脚本
.github/workflows/  CI/CD
```

## 技术选型(已定,勿擅自换库)

- 后端:Go 1.23+、Fiber v3、go-playground/validator、envconfig、zerolog、
  `modernc.org/sqlite`(纯 Go,免 CGO)、golang.org/x/image/webp、embed.FS
- 前端:Nuxt 3、UnoCSS、GSAP、pnpm
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

## 提交规范

- Commit message 英文,Conventional Commits
- 一个 commit 一个功能,保持单一
- 只在本地 commit,不要 push(由 maintainer 统一推送)
- 不要提交 `.env`

## 改动检查清单

每次改动后检查是否有意外副作用,尤其三处:

- [ ] **限流**:IP 级(429)与 name 级(降级只读)是否仍各自独立
- [ ] **缓存**:计数 SVG 仍 `no-store`,demo 仍长缓存(铁律 1)
- [ ] **存储**:仍是「请求→内存 Buffer→定时批量→SQLite」单一路径(铁律 5)

涉及 `tb_count` schema 变更时,任务结尾必须告诉用户是否需要迁移、跑哪个命令、
对哪个库。

---

下一步:→ [主题贡献指南](./contributing-themes.md)
