
前端示例页面 vue 可能会升级到nuxt
后端采用 go fiber

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

已经确认 detail.md 的参数契约。下面给出完整接口文档和简略实现。

## 接口调用文档

### `GET /@:name` — 计数器图片接口

对 `name` 计数 +1 并返回 SVG 图片。支持底图叠加、数字位置/大小/主题动态调整。

**基础信息**

| 项 | 值 |
|---|---|
| 方法 | `GET` |
| 路径 | `/@:name`(`/get/@:name` 为兼容别名) |
| 域名 | `https://umi7.top` |
| 返回类型 | `image/svg+xml` |
| 缓存 | 非 demo:`Cache-Control: no-store`;demo:`max-age=31536000` |

**路径参数**

| 参数 | 类型 | 说明 |
|---|---|---|
| `name` | string | 计数器名,最长 32 字符。`demo` 为保留值(固定 `0123456789`,不落库,长缓存) |

**查询参数**

| 参数 | 类型 | 范围 | 默认 | 说明 |
|---|---|---|---|---|
| `theme` | string | 白名单 / `random` | `loli` | 数字字形主题名,必须在 Registry 中存在,或为保留值 `random` |
| `bg` | string | 白名单 | 无 | 底图名,不传走纯数字模式;传则叠加底图 |
| `x` | float | 0 ~ 底图宽 | `0` | 数字块起始横坐标(相对底图左上角,像素) |
| `y` | float | 0 ~ 底图高 | `0` | 数字块起始纵坐标(相对底图左上角,像素) |
| `fsize` | int | 8 ~ 200 | `0` | 字体目标高度(像素)。`0` 表示不启用,用主题字形原始高度 |
| `scale` | float | 0.1 ~ 2 | `1` | 数字缩放倍数。与 `fsize` 同时传时,先按 `fsize` 归一化再乘 `scale` |
| `align` | enum | `top`/`center`/`bottom` | `top` | 数字块内部各字形的垂直对齐(字形高度不一时生效) |
| `padding` | int | 0 ~ 16 | `7` | 位数补零,如 `padding=7` 则 `123` → `0000123` |
| `offset` | float | -500 ~ 500 | `0` | 字间距(像素,负数重叠,正数拉开) |
| `pixelated` | enum | `0`/`1` | `1` | `1` 像素化渲染(`image-rendering: pixelated`),`0` 平滑 |
| `darkmode` | enum | `0`/`1`/`auto` | `auto` | 暗色模式亮度调节,`auto` 跟随 `prefers-color-scheme` |
| `num` | int | 0 ~ 1e15 | `0` | 指定数字直接展示(不落库,不 +1)。`>0` 生效,用于预览 |
| `prefix` | int | -1 ~ 999999 | `-1` | 前缀数字,`>=0` 时在计数前拼上该数字 |

**`fsize` 与 `scale` 的关系**(关键设计)

两者都控制数字大小,但语义不同:
- `fsize`:**绝对尺寸**,把数字高度归一化到指定像素。适合"我要数字高 40px"这种精确需求
- `scale`:**相对缩放**,基于主题字形原始高度乘倍数。适合"放大 1.5 倍"这种相对需求

计算顺序:`最终高度 = (fsize>0 ? fsize : 字形原始高度) × scale`

这样用户可以单独用 `fsize`(精确)或单独用 `scale`(相对),或组合用(`fsize=40&scale=1.2` = 48px)。两个参数都为默认值时用字形原始高度。

纯数字模式(无 `bg`)时,viewBox 按数字总宽×最高高度计算,无底图 `<image>` 层。

**三种嵌入方式**(同一个 URL)

```
1. SVG address
   https://umi7.top/@miaoledor?theme=loli&bg=loli-stand&x=20&y=180&fsize=40&scale=1

2. Img tag
   <img src="https://umi7.top/@miaoledor?theme=loli&bg=loli-stand&x=20&y=180&fsize=40&scale=1" alt="miaoledor" />

3. Markdown
   ![miaoledor](https://umi7.top/@miaoledor?theme=loli&bg=loli-stand&x=20&y=180&fsize=40&scale=1)
```

**示例**

```
# 基础计数(默认主题,纯数字)
https://umi7.top/@mycounter

# 带底图 + 调位置 + 调字号
https://umi7.top/@mycounter?bg=loli-stand&x=20&y=180&fsize=40

# 纯数字模式调字号 + 缩放 + 补零
https://umi7.top/@mycounter?theme=loli&fsize=32&scale=1.2&padding=8

# 预览(不落库)
https://umi7.top/@demo?bg=loli-stand&x=20&y=180&fsize=40

# 暗色模式 + 像素化关闭
https://umi7.top/@mycounter?darkmode=1&pixelated=0
```

## 项目结构：

```
lolicount/
│
├── cmd/                              # 【命令入口】编译产物的主程序入口,每个子目录编译成一个二进制
│   ├── server/                       # 主服务二进制(计数 API + 主题/底图管理)
│   │   └── main.go                   # 装配依赖(读 config → 建 logger → 按 STORAGE_TYPE 装一个 store → 启动 Fiber),唯一的 main
│   └── check-theme/                  # CI 校验工具二进制(校验主题目录完整性)
│       └── main.go                   # 扫 assets/theme/**,检查目录名合规、0~9 齐全、格式/尺寸/体积合格,供 CI 调用
│
├── internal/                         # 【内部包】业务代码,不对外暴露(go.mod 里不作为公开包)
│   │
│   ├── config/                       # 【配置】读 .env,结构体化所有配置项
│   │   └── config.go                 # envconfig 读 STORAGE_TYPE/DB_INTERVAL/REDIS_URL/R2_*/RATE_LIMIT_* 等,带默认值
│   │
│   ├── logger/                       # 【日志】zerolog 封装,全项目共用
│   │   └── logger.go                 # 初始化 zerolog,按 LOG_LEVEL 设级别,提供全局 logger 实例
│   │
│   ├── assets/                       # 【嵌入资源】把 assets/ 目录经 embed.FS 打包进二进制
│   │   └── embed.go                  # //go:embed assets/* 挂载,builtin 主题图 + 前端 dist 都从这里取
│   │
│   ├── server/                       # 【HTTP 层】Fiber v3 路由 + handler + 参数校验,编排所有模块
│   │   ├── server.go                 # Fiber app 装配:注册路由、挂中间件(限流/CORS/校验/日志)、托管静态资源、优雅关闭
│   │   ├── counter.go                # GET /@:name / /get/@:name handler:限流→Incr→取主题/底图→渲染→设响应头(铁律1)→返 SVG
│   │   ├── record.go                 # GET /record/@:name handler:返 JSON 计数(只读,不 +1)
│   │   ├── heartbeat.go              # GET /heart-beat:健康检查,no-store
│   │   ├── theme_api.go              # /api/themes CRUD handler:列表/详情/上传(重编码防图片马)/删除
│   │   ├── bg_api.go                 # /api/backgrounds CRUD handler:列表/上传(R2)/删除
│   │   ├── params.go                 # QueryParams 结构体 + validator 标签 + applyDefaults(),集中管理 x/y/fsize/scale/theme/bg 等参数
│   │   └── middleware.go             # CORS(仅 /api/*)、IP 级限流(Fiber limiter)、请求日志、recover
│   │
│   ├── store/                        # 【计数存储层】Repository 接口 + 三实现,业务只依赖接口(铁律5)
│   │   ├── repo.go                   # Repository interface: Get/GetAll/Set/SetMulti + Counter 结构体,三实现必须满足
│   │   ├── memory.go                 # memoryRepo:map + sync.RWMutex,单实例零依赖,演示用
│   │   ├── redis.go                  # redisRepo:INCR counter:<name> 原子自增,多实例生产
│   │   └── sqlite.go                 # sqliteRepo:tb_count 表,建表 SQL + SetMulti 事务批量 upsert,单实例持久化
│   │
│   ├── counter/                      # 【缓冲计数器】衔接 handler 与 store,sqlite 模式做内存缓冲 + 定时落库
│   │   ├── counter.go                # Buffer 结构:Incr/Get 直通(memory/redis)或缓冲(sqlite);启动/停止 ticker
│   │   ├── flush.go                  # sqlite 模式的 flush 逻辑:快照 cache→SetMulti→换新 map(修 Moe-Counter flush 期间丢增量)
│   │   └── buffer_test.go            # 并发自增、flush 期间增量不丢、上限守护的测试
│   │
│   ├── theme/                        # 【主题系统】数字字形图管理 + SVG 渲染
│   │   ├── registry.go               # Registry interface + composedRegistry:先查 userRegistry(Redis)再查 builtinRegistry(embed.FS)
│   │   ├── builtin.go                # builtinRegistry:启动扫描 embed.FS 的 assets/theme/*,DecodeConfig 读宽高,base64 转 data URI,缓存内存 map
│   │   ├── user.go                   # userRegistry:从 Redis 加载用户主题(JSON,含 data URI),Add/Remove 立即生效
│   │   ├── render.go                 # Render(纯数字模式):buildDigitChars + renderDigits + 算 viewBox + 拼 SVG
│   │   ├── render_bg.go              # RenderWithBg(底图叠加,方案C):底图 <image href=url> + 数字 <image> 叠加,用 x/y 定位
│   │   ├── digits.go                 # renderDigits 核心:数字转字符串→补零→逐位查字形→按 fsize/scale 算尺寸→按 align 算 y→输出 <image>
│   │   ├── style.go                  # buildStyle:暗色模式 <style> + 像素化 image-rendering,被 render/render_bg 共用
│   │   └── types.go                  # Theme/Glyph/RenderParams/RenderBgParams 结构体定义
│   │
│   ├── bg/                           # 【底图系统】底图元数据管理 + 对象存储
│   │   ├── registry.go               # Registry interface + composedRegistry:builtin(JSON 文件)+ user(Redis),Get/List/Add/Remove
│   │   ├── builtin.go                # builtinRegistry:加载 assets/bg/*.json(含 url/width/height)
│   │   ├── user.go                   # userRegistry:从 Redis 加载用户底图元数据,上传后立即生效
│   │   ├── storage.go                # R2/S3 上传:aws-sdk-go-v2,PutObject 拿 URL,删图 DeleteObject
│   │   └── types.go                  # Background 结构体:Name/URL/Width/Height/Author/Description
│   │
│   ├── ratelimit/                    # 【限流】双层限流,IP 级与 name 级职责分离(铁律3)
│   │   ├── ip.go                     # IP 级:10/s、300/min,超限返 429(Fiber limiter 中间件,可挂 Redis 后端)
│   │   └── name.go                   # name 级:5/s,超限返 true(降级只读,返当前值不 +1),handler 内调用
│   │
│   └── svg/                          # 【SVG 公共工具】跨模块的 SVG 拼装原语,theme 包内部用
│       └── svg.go                    # viewBox 计算、<image> 元素生成、data URI 编码等纯函数
│
├── assets/                           # 【静态资源】被 internal/assets embed.go 嵌入二进制
│   ├── theme/                        # 内置主题字形图,每子目录一个主题
│   │   ├── loli/                     # 默认主题
│   │   │   ├── 0.gif ... 9.gif       # 必须 10 张数字字形,格式 gif/png/webp
│   │   │   ├── _start.gif _end.gif   # 可选装饰图(前缀/后缀)
│   │   │   └── meta.json             # 可选元数据(name/author/description/tags)
│   │   └── ...                       # 其他内置主题(如 moebooru/asoul 等)
│   ├── bg/                           # 内置底图元数据(JSON,图本身在 CDN)
│   │   └── loli-stand.json           # {name,url,width,height,author,description}
│   ├── img/                          # 前端/README 用的静态图(logo、示例截图)
│   └── themes.json                   # CI 自动生成的主题清单(供前端 /api/themes 用)
│
├── web/                              # 【前端】Nuxt 3 SSG,playground 示例页 + 主题/底图浏览
│   ├── nuxt.config.ts                # SSG 配置、UnoCSS、GSAP、build 输出到 dist 供后端 embed
│   ├── app.vue                       # 根组件,单一真实根元素(AGENTS.md 前端原则5)
│   ├── pages/
│   │   ├── index.vue                 # 首页:介绍 + 主题预览 + 三种嵌入格式说明
│   │   └── playground.vue            # 示例调整页:选主题/底图→拖拽调 x/y→调 fsize/scale→实时预览→生成链接
│   ├── components/
│   │   ├── BgPreview.vue             # 预览区:底图背景 + 可拖拽数字块浮层,实时算 x/y
│   │   ├── ParamPanel.vue            # 参数控件:theme/bg/fsize/scale/align/padding/offset/darkmode
│   │   ├── LinkOutput.vue            # 输出区:同一 URL 派生 SVG address / Img tag / Markdown 三种格式 + 复制
│   │   ├── ThemeGallery.vue          # 主题画廊:展示所有主题字形预览
│   │   └── BgGallery.vue             # 底图画廊:展示所有底图缩略图
│   ├── composables/
│   │   ├── useDragPosition.ts        # 拖拽逻辑:pointer 事件→坐标换算(页面坐标→底图真实坐标)→防抖回调
│   │   ├── useCounterUrl.ts          # URL 构造:参数→只拼非默认值→生成 https://umi7.top/@name?...
│   │   └── useClipboard.ts           # 复制到剪贴板封装
│   ├── utils/
│   │   └── defaults.ts               # 参数默认值表(与后端 params.go 保持一致,前端拼 URL 时跳过默认值)
│   └── dist/                         # SSG 构建产物,被 internal/assets embed.go 打包进二进制
│
├── scripts/                          # 【CI 辅助脚本】Node.js 脚本,主题贡献流程用
│   ├── validate-theme-meta.js        # 校验 meta.json schema(name/author/tags 字段类型)
│   └── gen-themes-json.js            # 扫 assets/theme/* 生成 assets/themes.json(前端主题列表数据源)
│
├── .github/workflows/                # 【CI/CD】GitHub Actions
│   ├── theme-check.yml               # PR 改 assets/theme/** 或 assets/bg/** 时跑 check-theme + validate-theme-meta
│   ├── rebuild-frontend.yml          # 主题变更触发重建 SSG(web/dist → 提交或触发后端重建)
│   └── ci.yml                        # Go test/vet + 前端 build + Docker build
│
├── data/                             # 【运行时数据】SQLite 库文件落这里(gitignore)
│   └── .gitkeep                      # 占位,实际 count.db 运行时生成
│
├── docs/                             # 【文档】设计与契约
│   ├── detail.md                     # 技术细节(架构/接口/数据模型,改代码必查)
│   └── TODOlist.md                   # 任务清单
│
├── Dockerfile                        # 多阶段构建:Go 编译 + 前端 build → 单运行镜像
├── docker-compose.yml                # 本地起服:app + 可选 redis
├── .env.example                      # 配置模板(提交),.env 含密钥不提交(AGENTS.md session 规范)
├── .gitignore                        # 忽略 data/*.db、web/dist、.env、node_modules
├── go.mod                            # Go 1.23+,依赖 Fiber v3/zerolog/modernc.org/sqlite/aws-sdk-go-v2 等
├── AGENTS.md                         # AI agent 指南(铁律 + 工程原则,本项目最高约束)
└── README.md                         # 项目介绍 + 三种嵌入格式示例 + 部署说明
```
