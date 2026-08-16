# Lolicount 技术细节

本文档记录 Lolicount 的架构设计、模块职责、接口契约与实现要点。配合 [README.md](./README.md) 阅读。

## 目录

- [整体架构](#整体架构)
- [技术选型](#技术选型)
- [项目结构](#项目结构)
- [核心模块](#核心模块)
- [数据模型](#数据模型)
- [接口契约](#接口契约)
- [路由与参数](#路由与参数)
- [主题系统](#主题系统)
- [底图叠加(方案 C)](#底图叠加方案-c)
- [存储层](#存储层)
- [限流策略](#限流策略)
- [缓存策略](#缓存策略)
- [前端](#前端)
- [CI/CD](#cicd)
- [部署](#部署)
- [配置项](#配置项)

---

## 整体架构

Lolicount 沿用 Moe-Counter 的三层模型,并扩展底图叠加与社区贡献能力:

```
请求 GET /@:name?theme=&bg=&x=&y=
        │
        ▼
┌───────────────────────────────────────────┐
│  Fiber v3 HTTP 层                          │
│  ┌─────────┐ ┌──────────┐ ┌────────────┐  │
│  │ 限流    │→│ 参数校验 │→│ handler    │  │
│  │ IP+name │ │ validator│ │ counter.go │  │
│  └─────────┘ └──────────┘ └─────┬──────┘  │
└─────────────────────────────────┼─────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        ▼                         ▼                         ▼
┌────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│ counter.Buffer │    │ theme.Registry   │    │ bg.Registry      │
│ 内存自增+缓冲  │    │ 数字字形图       │    │ 底图元数据       │
│ ─────────────  │    │ ──────────────   │    │ ──────────────   │
│ store.Repo     │    │ builtin+user     │    │ builtin+user     │
│ memory/redis/  │    │ embed.FS + Redis │    │ JSON + R2/S3     │
│ sqlite         │    └────────┬─────────┘    └────────┬─────────┘
└───────┬────────┘             │                       │
        │                      ▼                       ▼
        │           ┌────────────────────────────────────┐
        │           │ theme.Render / RenderWithBg        │
        │           │ 拼接 SVG:底图<image>+数字<image> │
        │           └────────────────┬───────────────────┘
        │                            │
        └──────────────┬─────────────┘
                       ▼
                  返回 image/svg+xml
```

**核心流程(带底图模式)**:
1. 限流中间件(IP 级 → name 级)
2. 参数校验
3. `counter.Buffer.Incr(name)` 自增计数(超限降级只读)
4. `bg.Registry.Get(bgName)` 取底图元数据
5. `theme.Registry.Get(themeName)` 取数字字形
6. `theme.RenderWithBg(...)` 拼接 SVG
7. 返回 `image/svg+xml`,`Cache-Control: no-store`

---

## 技术选型

### 后端(Go)

| 模块 | 选型 | 理由 |
|---|---|---|
| Web 框架 | Fiber v3 | fasthttp v2,低开销,Express 风易迁移 |
| 路由校验 | go-playground/validator | 结构体标签校验,替代 zod |
| 配置 | envconfig | 读 `.env`,支持默认值 |
| 日志 | zerolog | 结构化、零分配 |
| 计数存储(默认) | memory(map + RWMutex) | 零依赖开箱即用 |
| 计数存储(推荐) | Redis(`INCR`) | 原子自增,多实例共享 |
| 计数存储(可选) | modernc.org/sqlite | 纯 Go 免 CGO,持久化 |
| 限流 | Fiber limiter 中间件 | 支持 Redis 后端 |
| 图片解码 | image/png, image/gif, x/image/webp | 标准库 + 扩展 |
| 对象存储 | aws-sdk-go-v2 | R2/S3 兼容 |
| 资源嵌入 | embed.FS | 主题图 + 前端 dist 打包 |

### 前端

| 模块 | 选型 | 理由 |
|---|---|---|
| 框架 | Nuxt 3 (SSG) | SEO 好,首屏快,静态部署 |
| 样式 | UnoCSS | 原子化,轻量 |
| 动画 | GSAP + CSS + Vue Transition | 数字滚动/过渡/撒花 |
| 部署 | CDN 或 embed 进 Go | 解耦,互不拖累 |

### 主题与底图

| 维度 | 主题 | 底图 |
|---|---|---|
| 内容 | `0~9` 字形图 + 可选 `_start/_end` | 一张完整图片 |
| 存储 | embed.FS(builtin)/ Redis(user) | CDN URL(builtin)/ R2(user) |
| 渲染 | data URI 内嵌 | 外部 URL 引用(方案 C) |
| 贡献 | PR + Web 上传 | PR(JSON)+ Web 上传(图) |

---

## 项目结构

```
loli-counter/
├── cmd/
│   ├── server/              # 主服务入口
│   │   └── main.go
│   └── check-theme/         # CI 主题/底图校验工具
│       └── main.go
├── internal/
│   ├── config/              # 环境变量配置
│   ├── logger/              # zerolog 封装
│   ├── server/              # Fiber 路由与 handler
│   │   ├── server.go
│   │   ├── counter.go       # GET /@:name
│   │   ├── record.go        # GET /record/@:name
│   │   ├── theme_api.go     # 主题 CRUD
│   │   └── bg_api.go        # 底图 CRUD
│   ├── store/               # 计数存储
│   │   ├── repo.go          # Repository 接口
│   │   ├── memory.go
│   │   ├── redis.go
│   │   └── sqlite.go
│   ├── counter/             # 缓冲计数器
│   │   └── counter.go
│   ├── theme/               # 主题系统
│   │   ├── registry.go      # ThemeRegistry
│   │   ├── render.go        # Render 数字 SVG
│   │   └── render_bg.go     # RenderWithBg 底图叠加
│   ├── bg/                  # 底图系统
│   │   ├── registry.go      # BackgroundRegistry
│   │   └── storage.go       # R2/S3 上传
│   ├── ratelimit/           # 限流
│   │   ├── ip.go
│   │   └── name.go
│   ├── assets/
│   │   └── embed.go         # embed.FS 挂载
│   └── logger/
├── assets/
│   ├── theme/               # 内置主题
│   │   ├── loli/
│   │   │   ├── 0.gif ... 9.gif
│   │   │   ├── _start.gif _end.gif
│   │   │   └── meta.json
│   │   └── ...
│   ├── bg/                  # 内置底图元数据
│   │   └── loli-stand.json
│   ├── img/
│   └── themes.json          # CI 自动生成
├── web/                     # Nuxt 3 前端
│   ├── nuxt.config.ts
│   ├── pages/
│   ├── components/
│   └── composables/
├── scripts/
│   ├── validate-theme-meta.js
│   └── gen-themes-json.js
├── .github/workflows/
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
└── README.md
```

---

## 核心模块

### `internal/server`
Fiber v3 应用入口,注册路由、挂载中间件(限流、CORS、校验、日志)、托管静态资源。

### `internal/store`
计数存储抽象,三实现可切换:
- `memoryRepo`:map + `sync.RWMutex`,单实例零依赖
- `redisRepo`:`INCR counter:<name>` 原子自增,多实例
- `sqliteRepo`:`tb_count` 表,批量 upsert

### `internal/counter`
缓冲计数器,衔接 store 与 handler:
- 内存 map 自增,降低 DB 压力
- `time.Ticker` 定时批量落库(sqlite 模式)
- name 级限流降级:超限返回当前值但不 +1

### `internal/theme`
主题加载与渲染:
- `Registry`:合并 builtin(embed.FS)+ user(Redis)
- `Render`:纯数字 SVG(对应 Moe-Counter 原版)
- `RenderWithBg`:底图 + 数字叠加 SVG

### `internal/bg`
底图管理:
- `Registry`:合并 builtin(JSON)+ user(Redis)
- `Storage`:R2/S3 上传客户端
- 只存元数据(URL + 宽高),图在 CDN

### `internal/ratelimit`
双层限流:
- `ip.go`:单 IP 令牌桶(10/s, 300/min)
- `name.go`:单 name 自增限流(5/s),超限降级只读

---

## 数据模型

### 计数

```go
type Counter struct {
    Name string `json:"name"`
    Num  int64  `json:"num"`
}
```

### 主题

```go
type Theme struct {
    Name       string              `json:"name"`
    Source     string              `json:"source"`     // "builtin" | "user"
    Author     string              `json:"author"`
    Description string             `json:"description"`
    Tags       []string            `json:"tags"`
    Characters map[string]ThemeChar `json:"characters"` // "0"~"9", "_start", "_end"
}

type ThemeChar struct {
    Width   int    `json:"width"`
    Height  int    `json:"height"`
    DataURI string `json:"dataURI"`   // base64 内嵌
    Format  string `json:"format"`    // image/gif|png|webp
}
```

### 底图

```go
type Background struct {
    Name        string `json:"name"`
    URL         string `json:"url"`          // CDN 地址,不内嵌
    Width       int    `json:"width"`
    Height      int    `json:"height"`
    Author      string `json:"author"`
    Description string `json:"description"`
}
```

---

## 接口契约

### `store.Repository`

```go
type Repository interface {
    Get(ctx context.Context, name string) (int64, error)
    Incr(ctx context.Context, name string) (int64, error)   // 原子自增,返回新值
    SetMulti(ctx context.Context, counters []Counter) error  // 批量,sqlite 用
}
```

### `theme.Registry`

```go
type Registry interface {
    List() []Theme
    Get(name string) (Theme, bool)
    Add(t Theme) error
    Remove(name string) error
}
```

### `bg.Registry`

```go
type Registry interface {
    List() []Background
    Get(name string) (Background, bool)
    Add(b Background) error
    Remove(name string) error
}
```

---

## 路由与参数

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/@:name` | 计数 +1,返回 SVG |
| GET | `/get/@:name` | 同上(兼容 Moe-Counter) |
| GET | `/record/@:name` | 返回 JSON 计数 |
| GET | `/heart-beat` | 健康检查 |
| GET | `/api/themes` | 主题列表 |
| GET | `/api/themes/:name` | 主题详情 |
| POST | `/api/themes` | 上传主题(multipart) |
| DELETE | `/api/themes/:name` | 删除用户主题 |
| GET | `/api/backgrounds` | 底图列表 |
| POST | `/api/backgrounds` | 上传底图(multipart) |
| DELETE | `/api/backgrounds/:name` | 删除底图 |

### `GET /@:name` 查询参数

| 参数 | 类型 | 范围 | 默认 | 说明 |
|---|---|---|---|---|
| `theme` | string | 白名单 / `random` | `loli` | 主题名 |
| `bg` | string | 白名单 | 无 | 底图名,不传走纯数字模式 |
| `x` | float | 0~底图宽 | `0` | 数字起始 x |
| `y` | float | 0~底图高 | `0` | 数字起始 y |
| `align` | enum | top/center/bottom | `top` | 数字垂直对齐 |
| `padding` | int | 0~16 | `7` | 位数补零 |
| `offset` | float | -500~500 | `0` | 字间距 |
| `scale` | float | 0.1~2 | `1` | 缩放 |
| `pixelated` | enum | 0/1 | `1` | 像素化渲染 |
| `darkmode` | enum | 0/1/auto | `auto` | 暗色模式 |
| `num` | int | 0~1e15 | `0` | 指定数字(不落库) |
| `prefix` | int | -1~999999 | `-1` | 前缀数字 |

---

## 主题系统

### 主题目录约定

```
assets/theme/<theme-name>/
  0.gif 1.gif ... 9.gif        # 必须,10 张
  _start.gif  _end.gif         # 可选,装饰
  meta.json                    # 可选,元数据
```

### `meta.json` schema

```json
{
  "name": "loli-pink",
  "author": "someone",
  "description": "粉色萝节数字",
  "tags": ["loli", "pink"]
}
```

### 加载流程

1. 启动时 `builtinRegistry` 扫描 `embed.FS` 的 `assets/theme/*`
2. 每个主题:遍历图片,`image.DecodeConfig` 读宽高,`base64` 转 data URI
3. 缓存到内存 map
4. `userRegistry` 从 Redis 加载用户主题(JSON,含 data URI)
5. `composedRegistry.Get` 先查 user 再查 builtin

### 渲染流程(`theme.Render`)

移植 Moe-Counter 的 `themify.js:getCountImage`:

1. `count.toString().padStart(padding, '0')` 补零
2. 可选加 `prefix` 前缀
3. 可选加 `_start` / `_end` 装饰
4. 逐位查主题图,`<image id>` 定义 + `<use x>` 引用
5. 计算 `viewBox` = 总宽 × 最大高
6. 拼接暗色模式 `<style>`
7. 输出完整 SVG

### 贡献通道

**PR 通道**:
- fork → `assets/theme/<name>/` 放图 + meta.json
- CI 跑 `cmd/check-theme`:目录名合规、`0~9` 齐全、格式/尺寸/体积合格
- 合并后进下一版 `embed.FS`

**Web 上传**:
- `POST /api/themes` multipart 上传 10 张图
- 服务端校验 + 重编码(防图片马)+ 转 data URI
- 写 `userRegistry`(Redis),立即生效

---

## 底图叠加(方案 C)

### 核心思路

底图用**外部 URL 引用**(CDN/R2),不内嵌 base64;数字图仍用主题系统 data URI。SVG 里底图 `<image href="url">` + 数字 `<image>` 叠加。

**优点**:
- SVG 体积小(底图不内嵌)
- 底图独立管理,走 CDN 缓存
- 数字图层复用现有渲染逻辑

### 底图元数据

`assets/bg/loli-stand.json`:
```json
{
  "name": "loli-stand",
  "url": "https://cdn.lolicount.app/bg/loli-stand.png",
  "width": 400,
  "height": 300,
  "author": "someone",
  "description": "萝莉立绘底图"
}
```

### `theme.RenderWithBg`

```go
func (r *Registry) RenderWithBg(p RenderBgParams) string {
    bg := p.Bg
    digitLayers := r.renderDigits(p.Count, p.Theme, p.X, p.Y, p.Align, p.Scale, p.Padding, p.Offset)
    style := buildStyle(p.Darkmode, p.Pixelated)

    return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg viewBox="0 0 %d %d" width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <title>Lolicount</title>
  <style>%s</style>
  <image href="%s" width="%d" height="%d" preserveAspectRatio="none"/>
  %s
</svg>`, bg.Width, bg.Height, bg.Width, bg.Height, style,
        bg.URL, bg.Width, bg.Height, digitLayers)
}
```

`renderDigits` 与 `Render` 的区别:有起始偏移 `(x, y)`,贴到底图指定位置。

### 输出示例

```xml
<svg viewBox="0 0 400 300" ...>
  <image href="https://cdn.lolicount.app/bg/loli-stand.png" width="400" height="300"/>
  <image href="data:image/gif;base64,..." x="20" y="180" .../>
  <image href="data:image/gif;base64,..." x="40" y="180" .../>
</svg>
```

### 底图上传

- `POST /api/backgrounds`:上传图 → R2 → 拿 URL → 存元数据
- 校验:格式白名单(png/webp)、尺寸上限、重编码
- URL 写入 `bg.userRegistry`(Redis)

---

## 存储层

### 三种实现

| 实现 | 适用 | 自增方式 |
|---|---|---|
| `memoryRepo` | 单实例,演示 | map + RWMutex |
| `redisRepo` | 多实例,生产 | `INCR`(原子) |
| `sqliteRepo` | 持久化,单实例 | 批量缓冲 upsert |

### `sqliteRepo` 表结构

```sql
CREATE TABLE tb_count (
    id    INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
    name  VARCHAR(32) NOT NULL UNIQUE,
    num   BIGINT      NOT NULL DEFAULT 0
);
```

### 缓冲计数器(`counter.Buffer`)

- 内存 map 维护当前计数,避免每次请求读 DB
- `time.Ticker` 按 `DB_INTERVAL` 批量 upsert(仅 sqlite)
- Redis 模式直接 `INCR`,无需缓冲
- 并发安全:`sync.RWMutex` 保护 map

---

## 限流策略

| 维度 | 限制 | 超出行为 |
|---|---|---|
| 单 IP | 10 req/s, 300/min | 429 |
| 单 name 自增 | 5/s | 返回当前值但不 +1(降级只读) |
| 上传接口 | 5 次/小时/IP | 429 |
| 写库缓冲 | 上限 1 万条 | 丢弃 + 日志告警 |

**实现**:
- `ratelimit.ip`:Fiber limiter 中间件,令牌桶,Redis 后端
- `ratelimit.name`:handler 内自定义,按 name 维度计数

---

## 缓存策略

| 对象 | 策略 | 理由 |
|---|---|---|
| 计数器 SVG(非 demo) | `Cache-Control: no-store` | 计数实时,GitHub 代理场景必需 |
| `demo` 主题 | `max-age=31536000` | 固定值,长缓存 |
| 底图(CDN) | `max-age=31536000` | 不变,浏览器/代理缓存 |
| SVG 结果(可选) | Redis LRU,key 含所有渲染参数 | 数字不变命中,降低 CPU |
| 主题字形 | 启动加载进内存 | 避免每次解码 |

---

## 前端

### Nuxt 3 SSG

- 首页内容(主题列表)构建时从 API 拉取,预渲染成静态 HTML
- 主题/底图变更时触发 `rebuild-frontend` workflow 重新 `nuxi generate`
- 部署:CDN(Cloudflare Pages / Vercel)或 embed 进 Go 二进制

### 页面

| 页面 | 功能 |
|---|---|
| `pages/index.vue` | 主题市场网格 + 参数表单 + 实时预览 + 复制 Markdown |
| `pages/themes.vue` | 主题画廊,标记 official/user |
| `pages/upload.vue` | 主题/底图上传表单 |

### 动画

- 数字滚动:GSAP `to` + `TextPlugin` 或自实现翻牌
- 主题切换:`<TransitionGroup>` 渐变
- 撒花:`party.js` 或自实现粒子

---

## CI/CD

四条流水线(详见 [.github/workflows/](./.github/workflows/)):

| 流水线 | 触发 | 产出 |
|---|---|---|
| `ci.yml` | PR / push | go vet + test -race + Nuxt build |
| `theme-check.yml` | PR 改 `assets/theme|bg/**` | 主题完整性校验 |
| `release.yml` | tag `v*` | Docker 镜像 + GitHub Release |
| `rebuild-frontend.yml` | 主题/首页变更 | Nuxt SSG 重建 + CDN 部署 |

### 主题校验规则(`cmd/check-theme`)

- 目录名:`^[a-z0-9_-]{1,32}$`,非保留字
- 必须存在 `0~9` 全部 10 张图
- `_start/_end` 可选
- 每张图:格式 ∈ {gif,png,webp},体积 ≤ 64KB,宽高 ≤ 200px
- `meta.json` 符合 schema

---

## 部署

### 方案 A:单 Docker(简单)

```bash
docker run -d -p 3000:3000 \
  -e STORAGE_TYPE=memory \
  ghcr.io/yourname/lolicount:latest
```

### 方案 B:前后端分离(生产)

- Go 后端:Docker → VPS / k8s
- Nuxt 静态站:CDN
- Redis:托管服务
- 底图:Cloudflare R2

### 方案 C:全 k8s

- Go Deployment + HPA(QPS 扩容)
- Redis StatefulSet
- 静态站走 Ingress 或 CDN

### Dockerfile(多阶段)

```dockerfile
# Stage 1: 前端
FROM node:20-alpine AS web
WORKDIR /app/web
COPY web/ .
RUN corepack enable && pnpm install --frozen-lockfile && pnpm build

# Stage 2: 后端
FROM golang:1.23-alpine AS go
WORKDIR /app
COPY . .
COPY --from=web /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

# Stage 3: 运行
FROM alpine:latest
COPY --from=go /app/server /server
EXPOSE 3000
ENTRYPOINT ["/server"]
```

---

## 配置项

`.env.example`:

```env
# 应用
APP_PORT=3000
SITE_URL=https://lolicount.app

# 存储:memory | redis | sqlite
STORAGE_TYPE=memory
REDIS_URL=redis://127.0.0.1:6379
DB_INTERVAL=60

# 限流
RATE_LIMIT_IP_RPS=10
RATE_LIMIT_IP_RPM=300
RATE_LIMIT_NAME_RPS=5

# 对象存储(R2/S3)
S3_ENDPOINT=
S3_BUCKET=
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_PUBLIC_URL=https://cdn.lolicount.app

# 日志:debug | info | warn | error
LOG_LEVEL=info
```

---

## 实现顺序

完整 TodoList 见 [README.md](./README.md) 的实现计划,共 60 项,按 M1→M8 顺序推进:

1. **M1** 项目骨架(配置/日志/Fiber/embed)
2. **M2** 主题系统(模型/加载/渲染/萝莉主题)
3. **M3** 存储与计数(三实现 + 缓冲 + 路由)
4. **M4** 限流与安全(IP/name/校验/CORS)
5. **M5** 底图叠加(BackgroundRegistry + RenderWithBg)
6. **M6** Web 上传(主题/底图 CRUD + 对象存储)
7. **M7** 前端(Nuxt SSG + 动画)
8. **M8** CI/CD 与部署(校验/流水线/Docker)
