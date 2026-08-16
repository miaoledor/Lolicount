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