## 简介
本项目用一张图片作为底图,计数文字叠加在图片上展示。

图片主题:比如 `lian`[0.webp 1.webp 2.webp ... (n-1).webp],
则每次请求从帧集合中随机抽取一张展示,count++。
多图层主题(如 `lian-ren`):由多个透明分层随机组合(类似 galgame 立绘),每次请求重新随机。
默认状态下 count 作为文字在图片正下方正中央展示。
可接受参数 `number` 选择数字进行展示(预览,不落库)。

前端采用 Nuxt 4 SSG,后端采用 Go Fiber。

> 本文档的接口契约与项目结构需与代码保持同步。如发现不符请修正。

## 数据存储

请求->内存->(定时批量写，解决sqlite单写者问题)->sqlite

### 表结构

用户计数表

```sql
-- tb_count: 计数器主表
-- 存储每个 name 的当前计数值,单实例 + 定时批量 upsert 写入
CREATE TABLE IF NOT EXISTS tb_count (
    id    INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    name  VARCHAR(32) NOT NULL UNIQUE,
    num   BIGINT      NOT NULL DEFAULT 0
);

unique自带索引

-- name 上的 UNIQUE 约束是 upsert (ON CONFLICT) 的触发条件,
-- 同时保证并发 upsert 同一 name 不会产生重复行。
-- id 是纯主键,业务不用,只是惯例。
```

## 图片叠加计数方案/使用方式

下面给出完整接口文档。参数契约以 `internal/server/params.go` 为准。

## 接口调用文档

### `GET /@:name` — 计数器图片接口

对 `name` 计数 +1 并返回 SVG 图片。支持底图叠加、数字位置/大小/主题动态调整。

**基础信息**

| 项 | 值 |
|---|---|
| 方法 | `GET` |
| 路径 | `/@:name`(`/get/@:name` 为兼容别名) |
| 域名 | `https://lolicount.top` |
| 返回类型 | `image/svg+xml` |
| 缓存 | 非 demo:`Cache-Control: no-store`;demo:`max-age=31536000` |

**路径参数**

| 参数 | 类型 | 说明 |
|---|---|---|
| `name` | string | 计数器名,最长 32 字符。`demo` 为保留值(固定 `0123456789`,不落库,单帧主题长缓存/多帧主题 no-store) |

**查询参数**

| 参数 | 类型 | 范围 | 默认 | 说明 |
|---|---|---|---|---|
| `theme` | string | 白名单 / `random` | `lian` | 图片主题名,必须在 Registry 中存在,或为保留值 `random` |
| ~~`mode`~~ | ~~enum~~ | ~~`seq`/`random`~~ | — | 已移除。所有主题统一使用随机帧选择 |
| `ftheme` | string | 白名单 / `random` | 无 | 文字风格主题名,或 `random`。不传用默认字体 |
| `number` | int64 | 0 ~ 999999 | `0` | 指定数字直接展示(不落库,不 +1)。`>0` 生效,用于预览 |
| `fsize` | int | 0 ~ 500 | `0` | 计数文字字号(像素)。`0` 表示用默认字号 |
| `scale` | float | 0.1 ~ 4 | `1` | 底图缩放倍数(基于统一最长边) |
| `unshowf` | bool | — | `false` | 隐藏计数文字 |
| `x` | int | -500 ~ 2000 | 无 | 文字像素横坐标(相对图片左上角) |
| `y` | int | -500 ~ 2000 | 无 | 文字像素纵坐标(相对图片左上角) |
| `rx` | float | 0 ~ 1 | 无 | 文字比例横坐标(相对图片宽) |
| `ry` | float | 0 ~ 1 | 无 | 文字比例纵坐标(相对图片高) |

**文字定位**:`x`/`y`(像素)与 `rx`/`ry`(比例)二选一;都不传时文字默认在图片正下方居中。

**`fsize` 与 `scale`**:两者独立。`scale` 控制底图大小,`fsize` 控制计数文字字号。

**三种嵌入方式**(同一个 URL)

```
1. SVG address
   https://lolicount.top/@miaoledor?theme=lian&fsize=16&scale=1

2. Img tag
   <img src="https://lolicount.top/@miaoledor?theme=lian&fsize=16&scale=1" alt="miaoledor" />

3. Markdown
   ![miaoledor](https://lolicount.top/@miaoledor?theme=lian&fsize=16&scale=1)
```

**示例**

```
# 基础计数(默认主题)
https://lolicount.top/@mycounter

# 单图层主题(每次请求随机抽帧)
https://lolicount.top/@mycounter?theme=lian

# 多图层主题(每次请求随机组合分层)
https://lolicount.top/@mycounter?theme=lian-ren

# 调字号 + 缩放
https://lolicount.top/@mycounter?theme=lian&fsize=32&scale=1.2

# 预览(不落库)
https://lolicount.top/@demo?theme=lian&fsize=16

# 隐藏文字只看图
https://lolicount.top/@mycounter?theme=lian&unshowf=true
```

### `GET /record/@:name` — JSON 计数(只读)

返回 `name` 的当前计数值,不 +1。值优先取内存 Buffer,未命中则读 SQLite。
`Cache-Control: no-store`,确保调用方始终拿到最新值(铁律 1)。

| 项 | 值 |
|---|---|
| 方法 | `GET` |
| 路径 | `/record/@:name` |
| 返回类型 | `application/json` |
| 缓存 | `no-store` |

响应体:`{ "name": "<name>", "num": <int64> }`

### `GET /heart-beat` — 健康检查

返回存活信号与 UTC 时间戳,`Cache-Control: no-store`(健康检查不应被代理缓存)。
响应体:`{ "status": "alive", "timestamp": "<RFC3339>" }`

### `GET /api/themes` / `GET /api/fthemes` / `GET /api/config`

前端只读数据接口,`/api/*` 统一挂 CORS(允许嵌入场景跨域)。短缓存
`Cache-Control: public, max-age=60`。

- `GET /api/themes` → `{ "themes": [{ "name": "..." }, ...] }`(卡片主题清单)
- `GET /api/fthemes` → `{ "fthemes": ["default", "neon", ...] }`(文字风格清单)
- `GET /api/config` → `{ "baseUrl": "<BASE_URL or empty>" }`(前端构建嵌入链接用)

### `GET /api/count/@:name` — JSON 计数(自增,Emote 挂件用)

与 `/@:name` 语义完全一致的 JSON 版本,供 Emote 挂件(`web/public/widget/widget.js`)
跨域 fetch:真实 name 自增(name 级限流超限降级只读,铁律 3);`demo` / `number>0`
特例不自增(demo 无 number 时返回 `123456789`)。挂在 `/api` 下自动继承 CORS。
`Cache-Control: no-store`(铁律 1)。

| 项 | 值 |
|---|---|
| 方法 | `GET` |
| 路径 | `/api/count/@:name` |
| 查询参数 | `number`(0~999999,可选,固定展示值) |
| 返回类型 | `application/json` |
| 缓存 | `no-store` |
| CORS | 继承 `/api` 的 reflect-origin |

响应体:`{ "name": "<name>", "num": <int64> }`

### `GET /api/psb/models` / `GET /psb/:model` — Emote 模型接口

Emote 挂件的模型资产通道,设计详见 `docs/emote-widget.md`。模型存放在
`assets/psb/<model>/model.psb`(pure ems 规格的 PSB,不入 git),经 `embed.FS`
打包,启动时载入 `Server.psbFS`。

- `GET /api/psb/models` → `{ "models": ["azuki", ...] }`,短缓存
  `public, max-age=60`,目录为空时返回空列表;
- `GET /psb/:model` → 模型原始字节,`application/octet-stream`,
  `public, max-age=31536000, immutable`(构建期嵌入,字节不可变;更换模型内容
  必须更换目录名),`Access-Control-Allow-Origin: *`(开发期 Nuxt 跨域加载);
  模型名白名单 `^[a-zA-Z0-9-]+$`,不匹配 400,不存在 404。

### `POST /api/editor/preview` / `POST /api/editor/export`

编辑器接口,接收完整图层栈 JSON(`EditorRequest`:canvas、layers、text 等),
preview 渲染 SVG 预览(不落库),export 打包主题 ZIP。均 `no-store`。
请求体契约见 `internal/server/editor_types.go`。

### `GET /api/admin/*` — 管理接口

需 `X-Admin-Key` 请求头,与 `ADMIN_KEY` 环境变量常量时间比对。`ADMIN_KEY`
为空时所有 admin 端点返回 404(不可见,而非仅禁止),避免泄露管理功能存在性。
非空时 key 不匹配返回 403。当前仅 `GET /api/admin/ping`(`→ ok`)。

## 项目结构

```
lolicount/
├── cmd/                              # 命令入口,每个子目录编译成一个二进制
│   ├── server/                       # 主服务(计数 API + 主题渲染 + 前端托管)
│   │   └── main.go                   # 装配:config → logger → store → counter.Buffer → renderer → Fiber
│   ├── check-theme/                  # CI 主题校验工具
│   ├── fix-theme/                    # 卡片主题帧序号修复工具(0..n-1 连续重命名)
│   └── gen-character/                # 立绘素材生成工具(多角色:hinata/nanami/yuzu/minato/miyu/furi)
├── internal/
│   ├── config/                       # envconfig + godotenv 读环境变量,带默认值与校验
│   ├── logger/                       # zerolog 封装
│   ├── server/                       # Fiber v3 路由 + handler + 中间件
│   │   ├── server.go                 # Server 构造 + registerRoutes + 优雅关闭
│   │   ├── counter.go                # GET /@:name / /get/@:name:限流→Incr→渲染→no-store
│   │   ├── record.go                 # GET /record/@:name:JSON 计数(只读)
│   │   ├── api.go                    # GET /api/themes /api/fthemes /api/config
│   │   ├── heartbeat.go              # GET /heart-beat 健康检查
│   │   ├── params.go                 # queryParams + validator + applyDefaults
│   │   ├── middleware.go             # cors / ipRateLimit / sanitizeBackslashEscape / adminAuth
│   │   ├── compose.go                # 统一图层栈合成(buildThemeLayers/compose/themeIsMultiFrame)
│   │   ├── editor_preview.go         # POST /api/editor/preview:图层栈 → SVG 预览
│   │   ├── editor_export.go          # POST /api/editor/export:图层栈 → 主题 ZIP
│   │   ├── editor_types.go           # EditorRequest/EditorLayer/EditorImage 契约
│   │   └── frontend.go               # embed 前端 dist + SPA fallback + BASE_URL 注入
│   ├── counter/                      # 内存 Buffer + 定时批量落库(铁律 5)
│   │   └── buffer.go                 # Incr/Get/flush,绝对值 cache 不换 map
│   ├── store/                        # SQLite repository(Repository 接口 + sqliteRepo 唯一实现)
│   │   ├── repository.go             # Repository interface: Get/GetAll/Set/SetMulti
│   │   └── sqlite.go                 # tb_count 表 + SetMulti 事务批量 upsert
│   ├── imgcore/                      # 渲染核心(统一图层栈模型)
│   │   ├── layer.go                  # Layer 契约 + LayerKind 枚举
│   │   ├── asset/                    # 主题加载(统一目录,按 ren.json 分派 → *theme.Theme)
│   │   ├── composer/                 # 图层栈合成 Compose + ThemeRegistry
│   │   ├── render/                   # Layer 实现(ImageLayer/GroupLayer/TextLayer/RandomPickLayer)
│   │   ├── theme/                    # Theme/Canvas/TextStyle 数据模型
│   │   └── imgutils/                 # SVG/geometry 工具
│   ├── ratelimit/                    # IP / name 限流(token bucket)
│   ├── themetool/                    # 主题元数据工具
│   └── assets/                       # (embed.FS 挂载点,见 assets/embed.go)
├── assets/                           # 静态资源,被 assets/embed.go 嵌入二进制
│   ├── embed.go                      # //go:embed all:theme all:f-theme all:img all:dist
│   ├── theme/                        # 所有主题(单图层+多图层统一):lian kuon hinata lian-ren ...
│   ├── f-theme/                      # 文字风格:default neon pink serif
│   ├── img/                          # 杂项图片(logo、示例截图)
│   ├── themes.json                   # CI 生成的卡片主题清单(前端 /api/themes 消费)
│   └── dist/                         # 前端 SSG 构建产物(embed,.gitkeep 占位)
├── web/                              # Nuxt 4 SSG 前端
│   ├── nuxt.config.ts                # SSG 配置、UnoCSS、GSAP
│   ├── app/
│   │   ├── app.vue                   # 根组件(单一真实根元素)
│   │   ├── pages/index.vue           # 首页:介绍 + playground + 嵌入格式
│   │   ├── pages/about.vue           # 关于页
│   │   ├── pages/editor.vue          # 主题编辑器(快速/高级模式)
│   │   ├── components/               # BgPreview/ParamPanel/LinkOutput/NavBar ...
│   │   ├── components/editor/        # 编辑器组件(EditorCanvas/LayerPanel/LayerItem ...)
│   │   ├── composables/              # useApi/useI18n/useTheme/useGitHub/useEditorApi/useEditorStorage
│   │   ├── i18n/                     # 中英双文 locale
│   │   └── utils/                    # cn(类名合并)
│   └── dist -> .output/public        # SSG 产物软链
├── scripts/                          # CI 辅助脚本(Node.js)
│   ├── validate-theme-meta.js        # 校验 meta.json schema
│   ├── gen-themes-json.js            # 生成 assets/themes.json
│   ├── build-web.js / dev-web.js     # 前端构建/开发脚本
│   ├── optimize-images.mjs           # 图片优化脚本
│   └── preflight-port.js             # 开发前端口占用检查
├── .github/workflows/                # CI/CD
│   ├── ci.yml                        # Go test/vet + 前端 build
│   ├── theme-check.yml               # PR 改 assets/theme/** 时跑校验
│   ├── rebuild-frontend.yml          # 主题变更触发 SSG 重建
│   └── release.yml                   # tag v* 构建 Docker + 多平台二进制
├── docs/                             # 文档
├── Dockerfile                        # 多阶段构建:Node 编 Nuxt → Go 编二进制 → alpine 运行
├── docker-compose.yml                # 本地起服(单实例 SQLite)
├── .env.example                      # 配置模板(提交),.env 含密钥不提交
├── go.mod                            # Go 1.25+,依赖 Fiber v3/zerolog/modernc.org/sqlite 等
├── AGENTS.md                         # AI agent 指南(铁律 + 工程原则)
└── README.md                         # 项目介绍 + 嵌入格式 + 部署
```


## 部署配置
部署拉取不会覆盖之前的sqlite 隔离生产和测试的sql
