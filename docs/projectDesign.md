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

主题使用量表由同一初始化语句自动创建,用于热门主题统计:

```sql
CREATE TABLE IF NOT EXISTS tb_theme_usage (
    theme VARCHAR(128) NOT NULL PRIMARY KEY,
    num   BIGINT       NOT NULL DEFAULT 0
);
```

第二个 `counter.Buffer` 通过 `store.ThemeUsageSource` 写入该表;它与
`tb_count` 共享数据库文件与批量写语义,但命名空间完全隔离。

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
| 缓存 | 真实计数:`no-store, no-cache, max-age=0, must-revalidate`;确定性 demo:`public, max-age=31536000` |

**路径参数**

| 参数 | 类型 | 说明 |
|---|---|---|
| `name` | string | 非空计数器名。`demo` 为保留值(固定渲染 `0123456789`,不落库;单帧主题长缓存,多帧主题即使传 `number` 也使用 no-cache 组合头) |

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
| `text` | string | 最长 64 字符 | 无 | 自定义文字模板;`{n}` 替换为计数值。传入后强制显示文字层 |

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
`Cache-Control: no-store`,确保调用方始终拿到最新值。该只读 JSON 路径不经过 GitHub 图片代理。

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

### `GET /api/themes` / `GET /api/themes/hot` / `GET /api/fthemes` / `GET /api/config`

前端只读数据接口,`/api/*` 统一挂 CORS(允许嵌入场景跨域)。常规列表使用
`Cache-Control: public, max-age=60`;热门列表单独使用 300 秒缓存。

- `GET /api/themes` → `{ "themes": [{ "name": "...", "variants": 12 }, ...] }`。
  静态主题带 `variants`;PSB/Spine/Live2D 模型追加在末尾,带
  `"animated": true` 与 `"kind": "psb|spine|live2d"`,前端据此切换对应挂件流程。
- `GET /api/themes/hot` → `{ "hot": ["theme-a", ...] }`。固定最多 10 个,
  启动时与每小时按 `tb_theme_usage` 重算,响应缓存 `public, max-age=300`。
- `GET /api/fthemes` → `{ "fthemes": ["default", "neon", ...] }`(文字风格清单)
- `GET /api/config` → `{ "baseUrl": "<BASE_URL or empty>" }`(前端构建嵌入链接用)

### `GET /api/count/@:name` — JSON 计数(自增,Emote 挂件用)

与 `/@:name` 语义完全一致的 JSON 版本,供 Emote 挂件(`web/public/widget/widget.js`)
跨域 fetch:真实 name 自增(name 级限流超限降级只读,铁律 3);`demo` / `number>0`
特例不自增(demo 无 number 时返回 JSON 数字 `123456789`,前导零无法保留)。挂在
`/api` 下自动继承 CORS。包括 demo/number 在内均使用完整 no-cache 组合头。

| 项 | 值 |
|---|---|
| 方法 | `GET` |
| 路径 | `/api/count/@:name` |
| 查询参数 | `number`(0~999999,可选,固定展示值) |
| 返回类型 | `application/json` |
| 缓存 | `no-store, no-cache, max-age=0, must-revalidate` |
| CORS | 继承 `/api` 的 reflect-origin |

响应体:`{ "name": "<name>", "num": <int64> }`

### `GET /api/psb/models` / `GET /api/psb/:model/download` / `GET /psb/:model` — Emote 模型接口

Emote 挂件的模型资产通道,设计详见 `docs/emote-widget.md`。模型存放在
`PSB_DIR/<model>/model.psb.gz`(默认 `assets/psb`,也支持未压缩 `model.psb`),
不经 `embed.FS` 打包。服务端启动时只挂载目录,不预读模型字节;请求时打开磁盘文件并
流式读取,响应结束后关闭文件流,请求之间不保留 PSB 字节或文件句柄。

- `GET /api/psb/models` → `{ "models": ["azuki", ...] }`,短缓存
  `public, max-age=60`,目录为空时返回空列表;
- `GET /api/psb/:model/download` → 模型文件**下载**(附件形式,落盘为真实的
  `<model>-model.psb.gz`,无 Content-Encoding,`application/gzip`),供用户在
  FreeMoteViewer / Emote_Widget 等工具中使用;
- `GET /psb/:model` → 模型字节(`model.psb.gz` 优先,以 `Content-Encoding: gzip`
  下发压缩字节,浏览器透明解压;未压缩 `model.psb` 同样支持),
  `application/octet-stream`,
  `private, max-age=31536000, immutable`(只允许浏览器本地缓存;按目录名约定内容
  不可变;更换模型内容必须更换目录名),`Access-Control-Allow-Origin: *`
  (开发期 Nuxt 跨域加载);前端加载省略 Cookie 并优先复用本地 HTTP 缓存;
  模型名白名单 `^[a-zA-Z0-9-]+$`,不匹配 400,不存在 404。

### Spine / Live2D 模型接口

Spine 与 Live2D 模型随构建嵌入二进制,由浏览器端 WebGL 运行时渲染。模型列表
使用 `public, max-age=60`;构建期固定的资产文件使用
`public, max-age=31536000, immutable` 并允许跨域读取。

- `GET /api/spine/models` → `{ "models": ["...", ...] }`;
- `GET /spine/models/:name/:file` → Spine 骨骼、atlas 或纹理文件;
- `GET /spine/anim/:name/:file` → 构建期预渲染的 WebP/PNG/GIF 动画;
- `GET /api/live2d/models` → `{ "models": ["...", ...] }`;
- `GET /live2d/models/:name/:file` → Live2D manifest、moc、纹理、动作或物理文件。

模型名和文件名均走白名单校验:非法输入返回 400,不存在的资产返回 404。详细
格式与嵌入方式见 `docs/spine-widget.md` 与 `docs/live2d-widget.md`。

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
│   │   ├── counter.go                # GET /@:name / /get/@:name:限流→Incr→渲染→缓存策略
│   │   ├── count_api.go              # GET /api/count/@:name:JSON 自增/预览计数
│   │   ├── record.go                 # GET /record/@:name:JSON 计数(只读)
│   │   ├── api.go                    # themes/hot/fthemes/config 列表接口
│   │   ├── heartbeat.go              # GET /heart-beat 健康检查
│   │   ├── params.go                 # queryParams + validator + applyDefaults
│   │   ├── middleware.go             # cors / ipRateLimit / sanitizeBackslashEscape / adminAuth
│   │   ├── compose.go                # 统一图层栈合成(buildThemeLayers/compose/themeIsMultiFrame)
│   │   ├── psb.go                    # PSB_DIR 磁盘模型列表/流式下发
│   │   ├── spine.go / live2d.go      # 嵌入式动态模型资产接口
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
│   └── themetool/                    # 主题元数据工具
├── assets/                           # 编译期静态资源 + 外置 PSB 默认目录
│   ├── embed.go                      # theme/f-theme/img/spine/live2d 与 dist 的 embed 声明
│   ├── theme/                        # 所有主题(单图层+多图层统一):lian kuon hinata lian-ren ...
│   ├── f-theme/                      # 文字风格:default neon pink serif
│   ├── img/                          # 杂项图片(logo、示例截图)
│   ├── spine/ / live2d/              # 编译期嵌入的动态模型资产
│   ├── psb/                          # PSB_DIR 默认值;不嵌入,运行时从磁盘读取
│   ├── themes.json                   # 脚本生成的主题资产清单(构建/工具使用)
│   └── dist/                         # 前端 SSG 构建产物(embed,.gitkeep 占位)
├── web/                              # Nuxt 4 SSG 前端
│   ├── nuxt.config.ts                # SSG 配置、UnoCSS、GSAP
│   ├── app/
│   │   ├── app.vue                   # 根组件(单一真实根元素)
│   │   ├── pages/index.vue           # 首页 + Playground(/themes 为 alias)
│   │   ├── pages/about.vue           # 关于页
│   │   ├── pages/playground/glyph.vue# 文字效果实验页
│   │   ├── components/               # BgPreview/ParamPanel/LinkOutput/NavBar ...
│   │   ├── composables/              # useApi/useI18n/useTheme/useGitHub
│   │   ├── i18n/                     # 中英日 locale
│   │   └── utils/                    # cn(类名合并)
│   └── dist -> .output/public        # SSG 产物软链
├── scripts/                          # CI 辅助脚本(Node.js)
│   ├── validate-theme-meta.js        # 校验 meta.json schema
│   ├── gen-themes-json.js            # 生成 assets/themes.json
│   ├── gen-theme-thumbs-webp.mjs      # 生成画廊 WebP 缩略图
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

主仓只保留计数器运行时、主题资产与贡献校验工具。主题不提供运行时上传或在线编辑,
只能修改仓库中的 `assets/theme/**` 并通过 Pull Request 贡献。


## 部署配置

生产计数保存在 `DB_PATH` 指向的 SQLite 文件中,升级代码不会自动覆盖该文件。
测试使用独立 DSN,不得复用生产库。升级前备份 `data/count.db`;若未来发生 schema
变更,按对应迁移说明执行。
