# Lolicount 架构文档

> 本文档描述项目结构、技术选型与核心架构决策。
> 贡献规范见 [功能贡献指南](./contributing-code.md),接口契约见
> [项目设计](./projectDesign.md)。

## 总体架构

Lolicount 是一个单二进制萌系访问计数器。后端用 Go + Fiber v3,前端用
Nuxt 3 SSG,主题图与前端 dist 通过 `embed.FS` 打包进同一个 Go 二进制。

```
                ┌───────────────────────────────────────┐
   HTTP 请求 ──▶│  Fiber v3 (internal/server)           │
                │  ├─ /@:name   计数 +1 → SVG           │
                │  ├─ /record   JSON 计数               │
                │  ├─ /api/*    主题/底图列表(CORS)    │
                │  └─ /         前端 SSG dist(embed)    │
                ├───────────────────────────────────────┤
                │  ratelimit  IP 级(429)/ name 级(降级)│
                ├───────────────────────────────────────┤
                │  counter.Buffer(内存 map 自增)       │
                │      │ time.Ticker(DB_INTERVAL)        │
                │      ▼ 批量 upsert                      │
                │  store.sqliteRepo → SQLite(data/count.db)│
                └───────────────────────────────────────┘
```

## 数据存储(铁律 5)

**唯一存储路径**:请求 → `counter.Buffer`(内存 map 自增)→ `time.Ticker`
按 `DB_INTERVAL` 批量 upsert → SQLite(`data/count.db`)。

- `counter.Buffer` 在内存维护当前计数,避免每次请求读/写 DB
- `flush()` 先快照 `cache` 再 `SetMulti`(事务内批量
  `INSERT ... ON CONFLICT(name) DO UPDATE`),成功后保留基线不换 map
- flush 期间的新增 Incr 写入当前 map,下次 flush 一并落库,不丢失
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

## 渲染模型(帧图模型)

本项目只使用一张图片展示。每个主题是一个**帧集合**:

```
theme[0.png 1.png 2.png ... (n-1).png]
显示帧 = (count+1) % n,count++
```

- **卡片主题(frame)**:支持顺序模式(`mode=seq`,随计数循环帧)与
  随机模式(`mode=random`,每次请求随机抽帧)
- **立绘主题(character)**:由多个透明分层组成,固定随机模式,
  每次请求重新组合服装/表情等
- 两种主题都作为底图,展示层级在 0;计数文字层级在 1
- `theme` 控制底图风格,`fsize`/`scale` 控制文字与图片大小

## 限流(铁律 3)

两套阈值、两种响应,不可统一:

- **IP 级**:`10/s, 300/min`,超限返 `429`
- **name 级**:`5/s`,超限**降级只读**(返回当前值但不 +1),
  让正常嵌入不被一次性刷量打挂

## 缓存(铁律 1)

| 资源 | Cache-Control | 理由 |
|---|---|---|
| 计数器 SVG(非 demo) | `no-store` | 计数实时,GitHub 代理场景必需 |
| `demo` 主题 | `max-age=31536000` | 固定值,长缓存 |
| 底图(CDN) | `max-age=31536000` | 不变,浏览器/代理缓存 |

GitHub 图片代理会缓存,任何给真实计数 SVG 加 `max-age` 的"优化"都会
让计数永久卡死。

## 上传通道安全(铁律 4)

`POST /api/themes`、`POST /api/backgrounds`:

- 服务端解码后按白名单格式(`gif/png/webp`)重编码再存,防图片马
- `Content-Type` / 文件后缀都不作为格式判定的唯一依据
- 校验:命名保留字、尺寸上限、体积上限、每 IP 配额
- 上传接口独立限流,不复用计数路径的限流配额

## 技术选型

| 层 | 选型 | 理由 |
|---|---|---|
| HTTP 框架 | Fiber v3 | 高性能,API 贴近 Express |
| SQLite 驱动 | `modernc.org/sqlite` | 纯 Go,免 CGO,交叉编译友好 |
| 日志 | zerolog | 结构化、零分配 |
| 参数校验 | go-playground/validator | struct tag 校验 |
| 配置 | envconfig | 环境变量驱动 |
| 前端 | Nuxt 3 SSG | 静态生成,embed 进二进制 |
| CSS | UnoCSS | 原子化,体积小 |
| 动画 | GSAP | 数字滚动、过渡 |
| 包管理 | pnpm | 前端依赖 |

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
    record.go        /record/@:name JSON
    api.go           /api/themes /api/fthemes
    params.go        QueryParams + validator
    middleware.go    cors / ipRateLimit
    frontend.go      embed 前端 dist
  counter/           内存 Buffer + 定时批量落库
  store/             SQLite repository(Repository 接口)
  theme/             主题注册与渲染(帧图模型)
  ftheme/            字体样式主题
  bg/                底图叠加
  ratelimit/         IP / name 限流
  assets/            embed.FS 挂载
web/                  Nuxt 3 前端(SSG)
  app/
    pages/           页面
    components/      组件
    composables/     useApi / useLoli / useI18n
    i18n/            locale 字典
assets/
  theme/             卡片主题(帧图)
  character/         立绘主题
  f-theme/           字体样式
  bg/                底图元数据
scripts/             主题校验 / 生成脚本
.github/workflows/   CI/CD
docs/                文档
```
