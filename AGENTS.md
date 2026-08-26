# Lolicount — AI Agent Project Guide

铁律-禁止修改该文件
严格遵守以下内容

萌系可换肤 SVG 访问计数器。后端 = Go 1.25+ / Fiber v3 / SQLite(`modernc.org/sqlite`),前端 = Nuxt 4 SSG。
单二进制部署:主题图 + 前端 dist 经 `embed.FS` 打包进 Go 二进制。

## 铁律 (Iron Rules — non-negotiable; these override every other guideline in this file)

1. **计数器 SVG 必须实时,`demo` 必须长缓存 —— 一条都不许混。** `GET /@:name`(以及 `/get/@:name`)的响应一律 `Cache-Control: no-store`;只有 `name=demo`(固定 `0123456789`,不落库)才能 `max-age=31536000`。GitHub 图片代理会缓存,任何给真实计数 SVG 加 `max-age` 的"优化"都会让计数永久卡死。这是本项目最关键的正确性约束,改动缓存逻辑前先把这条再读一遍。
3. **name 级限流超限是"降级只读",不是 429。** 单 name 超过 `RATE_LIMIT_NAME_PER_SEC`(默认 `20/s`)时,返回当前计数值但不 `+1`(降级),让正常嵌入不被一次性刷量打挂。`429` 是 IP 级限流(`RATE_LIMIT_IP_PER_SEC` 默认 `60/s`、`RATE_LIMIT_IP_PER_MIN` 默认 `3000/min`)的职责。两套阈值、两种响应,别图省事统一成一种。
4. **上传主题必须服务端重编码 —— 不信任客户端格式声明。** Web 上传通道(M6 预留,当前未实现)收到的图片,服务端解码后再按白名单格式重编码(`gif/png/webp`)再存,防图片马。`Content-Type` / 文件后缀都不能作为格式判定的唯一依据。同时校验:命名保留字、尺寸上限、体积上限、每 IP 配额(`RATE_LIMIT_UPLOAD_PER_HOUR`)。
5. **存储只有一条路径:请求 → 内存 Buffer → 定时批量写 → SQLite。** 不要再引入 memory/redis/sqlite 三态切换,也不要把 Redis 当 SQLite 的前置缓存。`counter.Buffer` 在内存自增 + `time.Ticker` 按 `DB_INTERVAL` 批量 upsert,解决 SQLite 单写者问题;`store.Repository` 是接口,`sqliteRepo` 是唯一实现,业务代码只依赖接口。多实例水平扩展是未来需求,届时再评估,当前不预设。

## Core Engineering Principles

> 共享基线,默认值而非教条 —— 用判断力。

1. 所有 commit message 第一句简要使用英文，在接下来使用英文详细描述本次commit 再在下方添加中文简要描述commit。
2. 所有代码注释用英文，尽量在整个函数或者方法，类上添加整段的描述注释，而不是在代码段内添加描述。
3. 单个源文件保持在 ~500 行以内;超过 ~300 行考虑拆分，尽量保证单文件功能单一。
4. Go:错误先 `if err != nil` 显式处理,不用 panic 控流;`context.Context` 作为函数第一个参数传递,不在包级全局存请求作用域状态。
5. 前端:每个函数写成 arrow function;类名用 `cn` 合并;Nuxt 页面/路由根组件必须有**单一真实根元素**(不用 `display: contents`,模板根不要有前导注释/兄弟节点,否则触发 Nuxt 单根节点警告并丢掉页面过渡动画)。
6. 前后端数据契约对齐:字段名、响应格式必须和 `docs/projectDesign.md` 的「接口调用文档」一节一致;改一边必查另一边。
7. 每次改动后检查是否有意外副作用(尤其限流/缓存/存储/主题显示错误三处)。
8. 涉及数据库 schema 变更(`tb_count` 表、SQLite)时,任务结尾必须显式告诉用户:是否需要迁移、跑哪个命令、对哪个库。本项目 SQLite 用 `modernc.org/sqlite`(纯 Go,免 CGO),不要混入 `mattn/go-sqlite3`。
9. 找最贴合项目现状的现代方案;必要时查官方最新文档。
10. 不要为优雅/模块化把代码写得复杂难懂,不写过度防御代码。

## Comments

**默认:但行不写。，在类或者方法的上面添加注释** 大多数注释要么是代码没写好的补丁,要么会随代码腐化。只在以下情况写:非显然的约束(比如上面铁律涉及的那几处缓存/限流分支,在代码里加一行注释指回 AGENTS.md 的对应条目)、外部契约要求、或者绕过了一个真实陷阱。解释"为什么",不解释"是什么"。

## Project Structure

按职责切包(domain-oriented),不是按技术层切(不用 controller/service/dao)。依赖方向必须单向:`internal/server`(HTTP/编排)→ `counter` / `imgcore`(渲染)→ `store`。`imgcore` 内部三个 drawer(`cardthemedrawer`/`characterthemedrawer`/`fdrawer`)互不 import,仅由 `imgcore/renderer` 合成。一旦出现循环依赖,说明分层错了,先修依赖方向再加功能。

## Data Storage

**唯一存储路径**:请求 → `counter.Buffer`(内存 map 自增)→ `time.Ticker` 按 `DB_INTERVAL` 批量 upsert → SQLite(`data/count.db`)。

- `counter.Buffer` 内存维护当前计数(绝对值),避免每次请求读/写 DB;`time.Ticker` 按 `DB_INTERVAL` 秒触发 `flush()`,快照 `cache` 调 `store.SetMulti`(事务内批量 `INSERT ... ON CONFLICT(name) DO UPDATE`)。
- **flush 不换 map,保留基线**:`cache` 存绝对计数,`flush()` 只快照不清空——`SetMulti` 是绝对值覆盖,下次 flush 重推增长的值即可。`SetMulti` 在飞期间的 `Incr` 直接写当前 map,纳入下次 flush,不丢失;`SetMulti` 失败时快照值仍在 cache(可能已被并发 Incr 覆盖更大),下次 flush 重试。
- **缓冲上限**:`len(cache) > 10000` 时不再接受新 name 的缓冲,降级只读 + 日志告警,防极端流量撑爆内存。
- **数据丢失窗口**:`DB_INTERVAL` 秒内进程崩溃,内存 `cache` 全丢。这是缓冲方案的固有代价,生产建议 `DB_INTERVAL=5~10`,演示可 60。
- **严格单实例**:`cache` 是进程私有 + `SetMulti` 绝对值覆盖,多实例会互相吞计数。当前不支持水平扩展,这是有意识的权衡。

### `tb_count` 表结构

```sql
CREATE TABLE IF NOT EXISTS tb_count (
    id    INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    name  VARCHAR(32) NOT NULL UNIQUE,
    num   BIGINT      NOT NULL DEFAULT 0
);
```

`name` 的 `UNIQUE` 约束**自带唯一索引**(SQLite 自动创建 `sqlite_autoindex_*`),不需要再手动 `CREATE INDEX`。该索引同时是 `ON CONFLICT(name)` upsert 的触发条件,并保证并发 upsert 同一 `name` 不会产生重复行。`num` 用 `BIGINT`(64 位整数),业务从不按 `num` 查询,无需额外索引。

## Rendering (imgcore 两层合成)

渲染核心在 `internal/imgcore`,server 只调 `renderer.Render`。两类主题作底图(layer 0),计数文字作 layer 1:

- **卡片主题(frame)**:`assets/theme/<name>/` 下帧图 `0..n-1`(gif/png/webp),`cardthemedrawer.Draw` 把选中帧 base64 内嵌成 data URI `<image>`。显示帧 = `(count+1) % size`(`mode=seq`,默认);`mode=random` 每次请求随机抽帧。
- **立绘主题(character)**:`assets/character/<name>/` 下 `ren.json` + 分层图(ren/*.webp),`characterthemedrawer` 每次请求随机组合分层(类似 galgame 立绘),**固定随机模式**,不支持 `seq`。
- **文字层**:`fdrawer.Draw` 用 `<text>` 渲染计数文字,`ftheme` 控制字体/颜色/粗细。
- **合成**:`renderer.Render` 合并两层,viewBox = `max(bg宽, 文字宽) × (bg高 + 文字高)`;底图水平居中,文字默认在图片正下方居中。
- **`scale`**:控制底图显示大小(基于统一最长边缩放)。`fsize`:控制计数文字字号。两者独立。
- **文字定位**:`x`/`y`(像素)或 `rx`/`ry`(比例 0~1)二选一;都不传时文字默认图片正下方居中。
- **`demo` / `number` 参数特例**:`demo` 固定返回 `0123456789`,不落库,长缓存;`number>0` 直接展示该值,不落库不 +1。这两条在 handler 层 early return,不进 `counter.Buffer`。

## Key Conventions

- **主题加载**:`renderer.NewThemeRegistry()` 启动时扫描 `embed.FS` 的 `assets/theme/*`(卡片)与 `assets/character/*`(立绘),帧图 base64 转 data URI 缓存内存。`renderer.NewFThemeRegistry()` 扫描 `assets/f-theme/*.json`。
- **主题目录约定**:卡片主题格式 gif/png/webp;立绘主题分层图用 webp。
- **`random` 主题**:从 builtin 列表(card + character)随机挑一个,每次请求重选(不走缓存)。
- **CORS**:Web 上传通道(`/api/*`)需要 CORS;计数 SVG 路径(`/@:name`)被 README 嵌入,通常不需要 CORS 头。

## Caching Strategy (do not "optimize" without re-reading Iron Rule 1)

| 资源 | Cache-Control | 理由 |
|---|---|---|
| 计数器 SVG(非 demo) | `no-store` | 计数实时,GitHub 代理场景必需 |
| `demo` 主题 | `max-age=31536000` | 固定值,长缓存 |
| `/api/*` 列表 | `public, max-age=60` | 短缓存,平衡新鲜度与压力 |

## Upload Channel (Web 上传)

Web 上传通道(M6 预留,当前 `POST /api/themes`、`POST /api/backgrounds` 尚未实现)是用户自助上传通道,安全约束:

- 格式白名单:gif/png/webp
- 服务端重编码(防图片马),不信客户端 `Content-Type`
- 尺寸上限、体积上限、每 IP 配额(如 5 次/小时/IP)
- 在本地前端展示最近已上传，不允许重复上传
- 命名保留字:不能与 builtin 主题/`demo`/`random` 冲突
- 上传接口独立限流,不复用计数路径的限流配额

## CI / Theme Contribution

- `cmd/check-theme`:校验主题完整性(目录名合规、帧完整性、格式/尺寸/体积合格)
- `scripts/validate-theme-meta.js`:`meta.json` schema 校验
- `scripts/gen-themes-json.js`:生成 `assets/themes.json`
- PR 改动 `assets/theme/**` 或 `assets/character/**` 触发 `theme-check.yml`
- 主题变更触发 `rebuild-frontend.yml` 重建 SSG

## Database

- SQLite 表 `tb_count`,字段:`name` / `num`。批量 upsert 走 `SetMulti`。
- schema 变更要写迁移脚本并在任务结尾提醒用户是否需要执行。
- DB 测试用独立的测试 DSN,不要从 `.env` 取、不要打印 DSN;共享库的 Go 集成测试跑 `-count=1 -p 1`。

## Dependencies (Tech Stack)

后端:Go 1.25+、Fiber v3、go-playground/validator、envconfig + godotenv、zerolog、modernc.org/sqlite、golang.org/x/image/webp、embed.FS。
前端:Nuxt 4(SSG)、Vue 3、UnoCSS、GSAP、clsx、pnpm。
对象存储(R2/S3)为 M6 预留,当前未引入 aws-sdk-go-v2。
不要擅自引入同类替代库(如把 Fiber 换成 Gin、把 zerolog 换成 slog、把 modernc.org/sqlite 换成 mattn/go-sqlite3)—— 选型已定,换库要先和用户确认。

## Session / Branch Hygiene

一次任务 = 一个 Codex session;一个目标仓库 = 一个分支 = 一个 worktree。别让两个 session 写同一个 checkout。本仓库的 `.env` 含密钥(R2/S3),永远不要提交 `.env`,只提交 `.env.example`。

## Git 
每次commit只进行一个功能的实现，保证每个commit功能的单一
commit时给出本次规范详细的commit message，描述本次commit

### commit message格式要求：
commit message第一句使用规范的英文
如feat(change) fix(change) ...
在下面一段使用英语描述本次的修改
再下一段使用中文描述本次的修改内容

只在本地commit 不要进行任何push

## test
在每次进行功能commit后判断是否需要单元测试，尽量使用单元测试并保存单元测试内容
如果需要，下一次commit 提交测试内容并进行测试
对测试出的错误进行修复commit
再次检查测试，看测试是否完善，数据是否足够覆盖大部分情况，如果数据覆盖率不足，则增加数据强度，再次进行测试
使用增强的数据进行测试，修复测试出来的bug
容易变动的主题内容和数量不应该影响测试的结果


## human check
每次小修改判断是否需要人工检验，若是判断需要则给出人工检测内容和操作示例
一大阶段任务完成后，总结本次工作做了什么，并给出人工检验的方式与示例

## about todolist
完成任务后尝试填充todolist的任务状态