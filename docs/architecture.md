# Lolicount 架构文档

> 本文档描述项目结构、技术选型与核心架构决策。
> 贡献规范见 [功能贡献指南](./contributing-code.md),接口契约见
> [项目设计](./projectDesign.md),主题模型见 [主题文档](./themes.md)。

## 总体架构

Lolicount 是一个核心单二进制萌系访问计数器。后端用 Go + Fiber v3,前端用
Nuxt 4 SSG;主题图、Spine/Live2D 资产与前端 dist 通过 `embed.FS` 打包进
Go 二进制。体积较大的可选 E-mote PSB 模型是例外,通过 `PSB_DIR` 从磁盘读取。

```
                ┌───────────────────────────────────────┐
   HTTP 请求 ──▶│  Fiber v3 (internal/server)           │
                │  ├─ /@:name      计数 +1 → SVG        │
                │  ├─ /get/@:name  兼容别名              │
                │  ├─ /record/@:name  JSON 计数          │
                │  ├─ /heart-beat  健康检查              │
                │  ├─ /api/*       主题/字体/配置(CORS) │
                │  └─ /            前端 SSG dist(embed)  │
                ├───────────────────────────────────────┤
                │  ratelimit  IP 级(429)/ name 级(降级)│
                ├───────────────────────────────────────┤
                │  counter.Buffer(内存 map 自增)       │
                │      │ time.Ticker(DB_INTERVAL)        │
                │      ▼ 批量 upsert                      │
                │  store.sqliteRepo → SQLite(data/count.db)│
                └───────────────────────────────────────┘
```

### 请求生命周期(计数 +1)

1. 请求到达 `GET /@:name`,`sanitizeBackslashEscape` 中间件先把 query 里的
   `\&` 还原成 `&`(兼容 milkdown/remark 等编辑器对 `&` 的反斜杠转义),
   无反斜杠时零开销放行。
2. `ipRateLimit` 做 IP 级限流(超限 429)。
3. `counterHandler` 克隆 `name`(避免 fasthttp 缓冲区复用),`parseParams`
   绑定并校验查询参数。
4. `nameLimiter` 先做 name 级限流;额度内由 `counter.Buffer.Incr` 在内存
   map 自增,超限则调用 `Get` **降级只读**(返回当前值但不 +1,铁律 3)。
5. `composer.Compose` 合成所有图层(底图 + 计数文字)
   生成 SVG,设 `Cache-Control: no-store, no-cache, max-age=0, must-revalidate`
   (铁律 1),返回 `image/svg+xml`。
6. `counter.Buffer` 内的 `time.Ticker` 按 `DB_INTERVAL` 秒触发 `flush()`,
   批量 upsert 到 SQLite。

## 数据存储(铁律 5)

**唯一存储路径**:请求 → `counter.Buffer`(内存 map 自增)→ `time.Ticker`
按 `DB_INTERVAL` 批量 upsert → SQLite(`data/count.db`)。

- `counter.Buffer` 在内存维护当前计数,避免每次请求读/写 DB
- `flush()` 快照 `cache`(绝对值)调 `SetMulti`(事务内批量
  `INSERT ... ON CONFLICT(name) DO UPDATE`),**不换 map、不清空**——
  cache 存绝对计数,SetMulti 是绝对值覆盖,下次 flush 重推增长值即可
- flush 在飞期间的 Incr 直接写当前 map,纳入下次 flush,不丢失
- `len(cache) > 10000` 时降级只读 + 日志告警,防极端流量撑爆内存
- 数据丢失窗口:`DB_INTERVAL` 秒内进程崩溃,内存 cache 全丢(固有代价)
- **严格单实例**:多实例会互相吞计数,当前不支持水平扩展

### `tb_count` 表

```sql
CREATE TABLE IF NOT EXISTS tb_count (
    id    INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    name  VARCHAR(32) NOT NULL UNIQUE,
    num   BIGINT      NOT NULL DEFAULT 0
);
```

`name` 的 `UNIQUE` 约束自带唯一索引,是 upsert 触发条件,并保证并发
upsert 同一 name 不会产生重复行。业务从不按 `num` 查询,无需额外索引。

### `tb_theme_usage` 表(主题热度统计)

```sql
CREATE TABLE IF NOT EXISTS tb_theme_usage (
    theme VARCHAR(128) NOT NULL PRIMARY KEY,
    num   BIGINT       NOT NULL DEFAULT 0
);
```

主题使用量的持久化,数据源是真实计数渲染(每次 `/@:name?theme=X` 渲染
+1;demo / number 渲染不计)。实现上完全复用 `counter.Buffer` 的批处理
语义:第二个 Buffer 实例通过 `store.ThemeUsageSource` 的 `ThemeUsage()` 视图
把读写隔离到本表,绝不混入 `tb_count` 的用户计数命名空间。热门主题列表
(`GET /api/themes/hot`,固定 10 个)按规格在**启动时与每小时**从使用量
缓存重算;无使用量时回退为注册表前 10 个主题。前端 Playground 以可折叠
分组展示:热门组置顶,其余按来源游戏分组、默认折叠。

## 渲染模型

渲染核心在 `internal/imgcore`,server 只调 `composer.Compose`。所有主题
统一为有序图层栈,计数文字作为其中一个图层,不再区分卡片主题与立绘主题。

### 统一主题模型

- **单图层主题**(原卡片):`assets/theme/<name>/` 下帧图 `0..n-1`,显示帧
  每次请求随机抽帧(从帧集合中随机选择一张展示)
- **多图层主题**(原立绘):`assets/theme/<name>/` 下 `ren.json` + 分层图
  (`ren/*.webp`),每次请求随机组合分层。所有主题统一随机抽帧(已移除 `mode` 参数)。
- 图层类型:`ImageLayer`、`RandomPickLayer`、`GroupLayer`、`TextLayer`,
  均实现 `imgcore.Layer` 接口。
- **画布稳定**:`RandomPickLayer` 携带全部候选帧的尺寸元数据(`FrameDims`),
  画布取所有帧按显示缩放后的最大宽高,随机换帧不跳变。
- 帧图 base64 内嵌成 data URI `<image>`(离线可用)。

### 文字风格主题(f-theme)

独立的 JSON 配置(`assets/f-theme/*.json`),定义计数文字的 `family`、
`color`、`weight`。通过 `ftheme` 参数切换,与图片主题解耦。

### 合成

`composer.Compose` 合并所有图层:viewBox = `max(bg宽, 文字宽) × (bg高 + 文字高)`,
底图水平居中,文字默认在图片正下方居中。

## 图片内存策略与占用

**策略:每槽位常驻一张(One Image Per Slot)+ 热门主题有限常驻。** 编译期
`go:embed` 全量打包不变,运行时既不把全部候选图解码为 base64 常驻堆,也
不给全部主题各留一份装配:

- **元数据**:`asset.LoadThemeCatalogs` 为每个主题构建 `ThemeCatalog`
  (路径、摆放坐标、尺寸)。尺寸由 `asset/header.go` 只读文件前 32 字节
  解析,不再 `fs.ReadFile` 全文件,启动因此不会把 1.03 GiB 嵌入 rodata
  全片触页;未识别容器回退 `image.DecodeConfig`。
- **常驻集**:`NewThemeRegistryWithResident(limit)` 只装配前 `limit` 个,
  `Server.refreshHotThemes()` 在启动与每小时用 `tb_theme_usage` 的 Top-N
  调用 `Warm()` 替换常驻集;`limit` 默认 10,由启动参数
  `-theme-resident-limit` 或环境变量 `THEME_RESIDENT_LIMIT` 配置
  (启动参数优先,0 = 全量常驻)。离开常驻集的主题立即释放 base64 装配。
- **冷主题**:`Get()` / `PrepareRender()` 用名字播种的 PRNG 现装一份私有
  主题,用完即弃,不写回注册表;装配结果与旧版启动装配一致,画廊缩略图
  不变。`List()` 仍是全量主题。
- **常驻主题请求**(`PrepareRender`):换入一个槽位——随机选多候选槽、
  `DecodeImage` 按需加载、替换部件、返回深拷贝快照;并发互不干扰。
- **缓存头判定**:`ThemeInfo.IsMultiFrame` 读元数据即可,不触发冷装配。

### 实测占用(2026-11-09,1203 主题 / 43469 张图 / 1.03 GiB 图片)

| 项目 | 本次优化后(常驻 10) | 优化前(全量常驻) |
|---|---|---|
| 常驻堆 HeapAlloc(GC 后) | **7.10 MB** | 308.68 MB |
| HeapSys / Sys(隔离进程) | 19.5 MB / 26.0 MB | 479.4 MB / 488.9 MB |
| 启动累计分配 | 39.4 MB/op | 2.73 GB/op |
| catalog 扫描分配 | 33.2 MB/op | 1.52 GB/op |
| 可服务主题数 | 1203 | 1203 |

测量方法、逐项贡献分解、代价与未落地候选见
[运行内存优化记录](./memory-optimization.md)。RSS 中高于堆的部分是二进制
映射的 file-backed 页(内核可回收);本次优化把启动触页从整片图片降到每
文件约一个 4 KiB 页,但没有改变 `go:embed` 全量打包的单二进制形态。

## 限流(铁律 3)

两套阈值、两种响应,不可统一:

- **IP 级**:`RATE_LIMIT_IP_PER_SEC`(默认 `60/s`)、`RATE_LIMIT_IP_PER_MIN`(默认 `3000/min`),超限返 `429`
- **name 级**:`RATE_LIMIT_NAME_PER_SEC`(默认 `20/s`),超限**降级只读**(返回当前值但不 +1),
  让正常嵌入不被一次性刷量打挂

## 缓存(铁律 1)

| 资源 | Cache-Control | 理由 |
|---|---|---|
| 计数器 SVG(非 demo) | `no-store, no-cache, max-age=0, must-revalidate` | 计数实时;GitHub 的 camo/Fastly 边缘无视裸 `no-store`,必须靠 `no-cache`/`max-age=0` 强制回源 |
| `demo` 主题(单帧,含 `number`) | `public, max-age=31536000` | 文字和背景都固定,长缓存 |
| `demo` 主题(多帧) | `no-store, no-cache, max-age=0, must-revalidate` | 每次随机选帧,不可长缓存 |
| `/api/count/@:name` | `no-store, no-cache, max-age=0, must-revalidate` | JSON 计数始终实时,包括 demo/number |
| 常规 `/api/*` 列表 | `public, max-age=60` | 主题/字体/配置/模型列表短缓存 |
| `/api/themes/hot` | `public, max-age=300` | 热门主题允许 5 分钟客户端缓存 |
| `/psb/:model` | `private, max-age=31536000, immutable` | 磁盘按需流式读取,只允许浏览器本地保存 |

GitHub 图片代理会缓存,任何给真实计数 SVG 加 `max-age` 的"优化"都会
让计数永久卡死;裸 `no-store` 也挡不住其边缘缓存(实测 MISS→HIT),
真实计数必须带完整的 no-cache 组合头。

动态预览页面会带浏览器累计 Cookie 与 Client Hints。Fiber 请求读取缓冲固定为
32 KiB(同时也是请求头上限),避免默认 4 KiB 在路由处理前误返 431,同时保持每连接
内存有界。公开 PSB 模型请求额外使用 `credentials: "omit"`,不会携带无关 Cookie。

## 主题贡献边界

运行中的 Lolicount 只读取随仓库构建的主题资产,不提供主题上传、在线编辑、审核或
投票接口。主题必须修改 `assets/theme/**` 后通过 Pull Request 提交,由仓库校验工具和
CI 检查后随新版本发布。完整流程见[主题贡献指南](./contributing-themes.md)。

## 技术选型

| 层 | 选型 | 理由 |
|---|---|---|
| HTTP 框架 | Fiber v3 | 高性能,API 贴近 Express |
| SQLite 驱动 | `modernc.org/sqlite` | 纯 Go,免 CGO,交叉编译友好 |
| 日志 | zerolog | 结构化、零分配 |
| 参数校验 | go-playground/validator | struct tag 校验 |
| 配置 | envconfig + godotenv | 环境变量驱动,.env 仅 dev 便利 |
| 前端 | Nuxt 4 SSG | 静态生成,embed 进二进制 |
| CSS | UnoCSS | 原子化,体积小 |
| 动画 | GSAP | 数字滚动、过渡 |
| 包管理 | pnpm | 前端依赖 |
| 并发启动 | concurrently | dev 模式同时启前后端 |

> 选型已定,不擅自引入同类替代库(如 Gin/slog/mattn/go-sqlite3)。
> 换库需先与维护者确认。

## 目录结构

```
cmd/
  server/            入口 main.go
  check-theme/       主题校验工具
internal/
  config/            环境变量加载与默认值
  logger/            zerolog 封装
  server/            Fiber 路由、handler、中间件
    server.go        Server 构造 + registerRoutes
    counter.go       /@:name 计数 + SVG
    count_api.go     /api/count/@:name JSON 计数
    record.go        /record/@:name JSON
    api.go           themes/hot/fthemes/config 列表接口
    psb.go           PSB_DIR 磁盘流式模型接口
    spine.go         Spine 嵌入资产接口
    live2d.go        Live2D 嵌入资产接口
    params.go        queryParams + validator
    middleware.go    cors / ipRateLimit / sanitizeBackslashEscape
    frontend.go      embed 前端 dist
  counter/           内存 Buffer + 定时批量落库
  store/             SQLite repository(Repository 接口)
  imgcore/           渲染核心(统一图层栈模型)
    asset/             主题加载(统一目录,按 ren.json 分派 → *theme.Theme)
    composer/          图层栈合成入口 + ThemeRegistry
    render/            Layer 实现(ImageLayer/GroupLayer/TextLayer/RandomPickLayer)
    theme/             Theme/Canvas/TextStyle 数据模型
    imgutils/          SVG/geometry 工具
  ratelimit/         IP / name 限流(token bucket)
web/                  Nuxt 4 前端(SSG)
  app/
    pages/           页面(index、about、playground/glyph)
    components/      组件
    composables/     useApi / useI18n / useTheme / useGitHub
    i18n/            locale 字典
assets/
  theme/             所有主题(单图层+多图层统一):lian kuon hinata lian-ren ...
  f-theme/           字体样式:default neon pink serif
  spine/ live2d/     编译期嵌入的动态模型
  psb/               PSB_DIR 默认目录(磁盘流式,不嵌入)
  dist/              前端 SSG 构建产物(embed)
  img/               杂项图片
  themes.json        主题清单(脚本生成)
scripts/             主题校验 / 生成 / dev 脚本
.github/workflows/   CI/CD
docs/                文档
```

## 依赖方向

按职责切包(domain-oriented),依赖单向:

```
internal/server (HTTP/编排) → counter / imgcore(renderer) → store
```

`imgcore/render` 内各 Layer 实现(ImageLayer/GroupLayer/TextLayer/RandomPickLayer)互不 import,仅由 `composer` 合成。`store.Repository`
是接口,`sqliteRepo` 是唯一实现,业务代码只依赖接口。
一旦出现循环依赖,说明分层错了,先修依赖方向再加功能。
