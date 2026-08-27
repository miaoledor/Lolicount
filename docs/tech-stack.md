# 技术选型

> 各层技术选型与理由。选型已定,不擅自引入同类替代库;换库需先与维护者确认。
> 架构总览见 [架构文档](./architecture.md)。

## 后端

| 依赖 | 版本 | 用途 | 选型理由 |
|---|---|---|---|
| Go | 1.25+(`go.mod`) | 编译/运行 | 静态编译、并发原语、交叉编译友好;单二进制部署的基础 |
| Fiber v3 | `v3.5.0` | HTTP 框架 | 高性能(基于 fasthttp),API 风格贴近 Express,上手快 |
| `modernc.org/sqlite` | `v1.56.0` | SQLite 驱动 | 纯 Go 实现,**免 CGO**,交叉编译无需 C 工具链 |
| zerolog | `v1.35.1` | 结构化日志 | 零分配,JSON 输出,性能高 |
| go-playground/validator | `v10.30.3` | 参数校验 | struct tag 声明式校验,与 Fiber Bind 无缝集成 |
| envconfig + godotenv | `v1.4.0` / `v1.5.1` | 配置 | 环境变量驱动;`.env` 仅 dev 便利,生产用真实环境变量 |
| `golang.org/x/image/webp` | `v0.45.0` | WebP 解码 | 标准库不含 webp 解码,补充主题图格式支持 |
| embed.FS | 标准库 | 资源打包 | 主题图 + 前端 dist 编译进同一个二进制,零外部文件 |

### 为什么不用 mattn/go-sqlite3?

`mattn/go-sqlite3` 依赖 CGO,交叉编译需目标平台的 C 工具链,显著增加构建
复杂度。`modernc.org/sqlite` 是纯 Go 实现,功能足够,免 CGO 后交叉编译
只需 `GOOS`/`GOARCH` 即可,适合单二进制多平台发布。

### 为什么不用 Gin?

Fiber v3 基于 fasthttp,性能更高且 API 更贴近 Express 风格。选型已定,
不因个人偏好切换。

## 前端

| 依赖 | 版本 | 用途 | 选型理由 |
|---|---|---|---|
| Nuxt | `^4.5.2` | 前端框架 | SSG 静态生成,产物可 embed 进 Go 二进制;文件路由简洁 |
| Vue | `^3.5.41` | UI 框架 | Nuxt 底层,组合式 API |
| vue-router | `^5.2.0` | 路由 | Nuxt 内置 |
| UnoCSS | `^66.7.5` | 原子化 CSS | 按需生成,体积小,与 Nuxt 集成好 |
| GSAP | `^3.15.0` | 动画 | 数字滚动、主题切换过渡、撒花效果 |
| clsx | `^2.1.1` | 类名合并 | 条件类名拼接(AGENTS.md 要求用 `cn` 合并) |
| pnpm | 9+ | 包管理 | 快、磁盘高效,monorepo 友好 |

### 为什么 SSG 而不是 SSR?

SSG 在构建期生成纯静态文件,运行时不需要 Node 进程,可直接 embed 进 Go
二进制实现真正的单二进制部署。SSR 需常驻 Node 服务,违背单二进制目标。

## 存储

| 组件 | 用途 | 选型理由 |
|---|---|---|
| SQLite(`data/count.db`) | 唯一持久化 | 单文件、零运维、嵌入式;访问计数场景读多写少且单实例足够 |
| `counter.Buffer`(内存) | 内存自增缓冲 | 避免每次请求读写 DB,`time.Ticker` 按 `DB_INTERVAL` 批量 upsert |

### 为什么不用 Redis?

当前是单实例,内存 Buffer 已解决 SQLite 单写者问题,引入 Redis 会增加运维
复杂度且无收益(铁律 5)。多实例水平扩展是未来需求,届时再评估。

### 数据丢失窗口

`DB_INTERVAL` 秒内进程崩溃,内存 `cache` 全丢。这是缓冲方案的固有代价,
生产建议 `DB_INTERVAL=5~10`。

## 工具链

| 工具 | 用途 |
|---|---|
| concurrently | dev 模式同时启动前后端 |
| Docker | 容器化部署(多阶段构建:builder 编 Nuxt+Go → alpine 运行) |
| GitHub Actions | CI/CD(vet + test + check-theme + build + release) |

## 版本备注

- `go.mod` 声明 `go 1.25.0`,AGENTS.md 与文档表述「Go 1.23+」为最低要求
  兼容性提示;实际开发与 CI 以 `go.mod` 的 `go` 指令为准。
- 前端为 Nuxt 4(`web/package.json`),所有文档已统一为 Nuxt 4。
