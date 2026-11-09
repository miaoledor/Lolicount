# 运行内存优化记录（M18）

> 分支 `feat/memory-optimization`，测量日期 2026-11-09，macOS / Apple M2 / Go 1.26.5。
> 本文记录已落地的改动、实测数据与验证方式；未落地候选列在最后一节。
> 主题模型与铁律见 [架构文档](./architecture.md)、[主题文档](./themes.md) 与 `AGENTS.md`。

## 1. 测量口径与资产规模

所有数字来自 `internal/imgcore/composer/memory_bench_test.go`，不是估算：

- `TestMemoryBaseline`：在强制 GC 前后读 `runtime.MemStats`，报告 `TotalAlloc` 差值（含短命分配）、GC 后存活的 `HeapAlloc`、`HeapSys`、`Sys`。
- `TestMemoryResidentOnly`：**单独进程**运行，避免同一二进制里前一个全量装配阶段抬高 `HeapSys`。
- `go test -bench ... -benchtime=3x`：`B/op`、`allocs/op`、`ns/op`。
- `internal/imgcore/asset/header_test.go`：遍历全部主题图片，逐张把头部解析结果与 `image.DecodeConfig` 对比。

复现命令：

```bash
go test ./internal/imgcore/composer -run TestMemoryBaseline -v -count=1
go test ./internal/imgcore/composer -run TestMemoryResidentOnly -v -count=1
go test ./internal/imgcore/composer -run '^$' -bench 'Benchmark(LoadThemeCatalogs|NewThemeRegistry|NewThemeRegistryWithResident)$' -benchtime=3x -count=1
go test ./internal/imgcore/asset -run TestHeaderDimsMatchesDecodeConfig -count=1
```

本次测量时的资产规模：**1203 个主题**（460 帧图 / 743 立绘）、**43469 张图**、图片合计 **1.03 GiB**（单图中位 3.3 KB、均值 25 KB、p95 143 KB、最大 0.69 MB）；`go:embed` 目录合计约 1.8 GB（theme 1.2 G + live2d 599 M + unity 36 M + dist 41 M）。`docs/architecture.md` 里 284 主题 / 8294 张图是 2026-09-30 的旧快照，已不能代表当前规模。

## 2. 三层内存构成与本次命中

| 层 | 优化前状况 | 本次是否命中 |
|---|---|---|
| ① 嵌入 rodata 触页 | 启动 `headerDims` 对每张图 `fs.ReadFile` 全文件，把 1.03 GiB 图片页全部触达，计入 RSS（file-backed、内核可回收） | **命中**：只读前 32 字节 |
| ② 常驻装配堆 | `NewThemeRegistry` 给全部 1203 个主题各装配一份 base64 图层栈，GC 后存活 308.68 MB | **命中**：只常驻热门 10 个 |
| ③ 请求期瞬时分配 | `DecodeImage` → `Compose` → `SendString` → gzip，每请求复制 3~4 份图字节 | 未做（见第 7 节） |

只做「砍掉冷主题」而不做 ① 的话，启动仍会扫描全量文件头，RSS 几乎不变——这是先做 ① 的原因。

## 3. 优化一：元数据只读文件头

**问题**：`asset.LoadThemeCatalogs` 为每张候选图调用 `headerDims`，旧实现是 `fs.ReadFile`（把整个文件复制进堆）再 `image.DecodeConfig`，只为拿宽和高。43469 次调用累计分配 1.45 GiB，并把嵌入 rodata 整片触页。

**改动**：新增 `internal/imgcore/asset/header.go`，`lazy.go` 删除旧实现。

- `fs.Open` + `io.ReadFull` 只读前 **32 字节**（栈上数组）。
- `dimsFromHeader` 直接解析 PNG `IHDR`、GIF 逻辑屏幕描述符、WebP `VP8` / `VP8L` / `VP8X` 的宽高位；这三种格式覆盖仓库全部资产。
- 未识别容器回退到已注册的图像解码器：用 `io.MultiReader` 把读走的前缀拼回文件流，`image.DecodeConfig` 仍从头看到完整头部，正确性不打折。
- 维度语义与 `image.DecodeConfig` 完全一致，`FrameDims` / 画布稳定性逻辑不变。

**正确性验证**：`TestHeaderDimsMatchesDecodeConfig` 遍历全部 43469 张图，逐张比对宽高，**全部一致**；另有 PNG 头、未知 WebP chunk 回退、超长 payload 短读三组单元测试。

**收益**：

| 指标 | 优化前 | 优化后 | 变化 |
|---|---|---|---|
| catalog 扫描累计分配 | 1448.72 MB | 31.69 MB | **−97.8%** |
| catalog 扫描 `B/op` | 1,518,758,472 | 33,176,613 | **−97.8%** |
| catalog 扫描 `allocs/op` | 847,503 | 437,229 | −48.4% |
| catalog 扫描耗时（热态 `ns/op`） | 374,098,861 | 106,146,292 | −71.6% |
| 冷启动 catalog 阶段（`ms`，首次触页） | 4910.5 | 1594.9 | −67.5% |
| rodata 启动触页 | 整文件约 1.03 GiB | 每文件约 1 个 4 KiB 页，估算 ~170 MiB | 估算 −84% |

## 4. 优化二：只让热门主题常驻

**关键区分**：这是「只**常驻** 10 个主题」，不是「只**支持** 10 个主题」。1203 个主题仍全部可访问；`go:embed` 全量打包、单二进制形态也没动。字面砍掉 1193 个主题会破坏 `/api/themes`、`random`、分享链接与 Playground，而且 rodata 仍在二进制里，RSS 并不会因此下降。

**改动**（`internal/imgcore/composer/registry.go`）：

- `NewThemeRegistryWithResident(limit)`：加载**全量 catalog**（路径/坐标/尺寸元数据，约 5.5 MB），只预装配前 `limit` 个（按名排序）。`NewThemeRegistry()` 保持全量装配，供测试与 `cmd/gen-theme-thumbs` 使用。
- `Warm(names)`：新接口 `ResidentController`。`Server.refreshHotThemes()` 在启动与每小时用 `tb_theme_usage` 的 Top-N 替换常驻集，离开集合的主题立即释放其 base64 装配。
- 冷主题：`Get()` / `PrepareRender()` 用**名字播种的 PRNG** 现装一份私有主题，用完即弃，不写入注册表。因为每次都是全新 `NewPRNG(name)`，装配结果与旧版「启动时装配」逐字节一致，**画廊缩略图不会变**。
- 新接口 `ThemeInfo.IsMultiFrame(name)`：`themeIsMultiFrame` 改读 catalog 元数据判定缓存头，不再为判定缓存策略而装配一个冷主题。
- `compose()` 去掉了「先 `Get()` 再被 `PrepareRender()` 立刻覆盖」的那次多余取用，冷主题每请求只装配一次。
- 常驻数量可通过启动参数 `-theme-resident-limit` 或环境变量 `THEME_RESIDENT_LIMIT` 配置，默认 **10**；详见「启动参数」小节。`/api/themes/hot` 对外仍固定返回 10 个，配置值只决定常驻集宽度。

**收益**（生产路径 = 优化一 + 常驻 10 个）：

| 指标 | 优化前 | 优化后 | 变化 |
|---|---|---|---|
| 启动累计分配（`B/op`） | 2,731,320,821 | 39,435,306 | **−98.6%** |
| 启动耗时（热态 `ns/op`） | 468,011,389 | 103,607,389 | −77.9% |
| 常驻堆 `HeapAlloc`（GC 后） | 308.68 MB | **7.10 MB** | **−97.7%** |
| `HeapSys`（隔离进程） | 479.41 MB | 19.50 MB | −95.9% |
| `Sys`（隔离进程） | 488.86 MB | 25.99 MB | −94.7% |
| 常驻装配主题数 | 1203 | 10 | −99.2% |
| 可服务主题数 | 1203 | **1203** | 0 |

两层优化各自的贡献（同一 `BenchmarkNewThemeRegistry` 口径）：

| 阶段 | 启动 `B/op` | 常驻堆 `HeapAlloc` |
|---|---|---|
| 优化前 | 2,731,320,821 | 308.68 MB |
| 仅优化一（全量装配） | 1,245,921,402（−54.4%） | 308.67 MB（≈不变） |
| 优化一 + 优化二（常驻 10） | 39,435,306（−98.6%） | 7.10 MB（−97.7%） |

即：**优化一降的是启动分配与触页，优化二降的是常驻堆**，两者必须同时做才能真正压住运行内存。表中 `HeapSys/Sys` 的「优化后」来自隔离进程，`TestMemoryBaseline` 里 `resident10` 阶段显示 455 MB 是因为同一进程刚跑完全量装配、Go 尚未把堆归还 OS，不能当作常态值。

### 启动参数：配置常驻主题数量

按你要求的「启动参数可配、默认 10 个」，提供三种设置方式，优先级从高到低：

| 方式 | 示例 | 说明 |
|---|---|---|
| 命令行启动参数 | `./lolicount -theme-resident-limit 32` | 最高优先级，覆盖环境变量 |
| 环境变量 | `THEME_RESIDENT_LIMIT=32 ./lolicount` | 未传启动参数时生效 |
| 内置默认值 | `./lolicount` | **10 个** |

- 也接受 `--theme-resident-limit=32` 写法；`0` 表示保持全量主题常驻（供缩略图生成与排障）。
- 负数、非整数、未知参数在启动阶段直接报错退出（退出码 2）；`-h` 打印用法后正常退出。
- 启动日志 `config loaded` 带 `theme_resident_limit` 字段，可确认最终生效值。
- 只影响常驻集宽度，**不改变主题总数、`random` 行为或 `/api/themes/hot` 的固定 10 个**。

实现位于 `cmd/server/main.go` 的 `applyRuntimeFlags`：用标准库 `flag.NewFlagSet` 解析，
把 flag 的默认值直接设成 `config.Load()` 的结果（环境变量或默认 10），因此不传参数时
原值不变、优先级自然成立；解析后再跑一次 `cfg.Validate()`，让命令行值同样接受范围校验。
未引入新依赖。

## 5. 代价与权衡

| 项 | 实测 / 说明 |
|---|---|
| 冷主题每请求装配 | 同一帧主题 `wenders`：常驻路径 519,345 `B/op` / 19 allocs，冷路径 529,304 `B/op` / 21 allocs，**分配几乎相同**（热路径同样要 `DecodeImage` 换帧，冷路径省掉深拷贝与换装锁）。请求级 `ns/op` 在不同 run 间波动较大（77~192 µs），不作为结论依据。多槽立绘主题冷装配需加载全部槽位，代价更高，本次未单独量化。 |
| 热度僵化 | `tb_theme_usage` 存累计绝对值，`Top()` 是历史总榜，新主题较难挤进池；滑动窗口 / 时间衰减列为后续候选。冷主题变热后，每小时刷新会把它 Warm 进池。 |
| 单二进制 rodata | 本次**未**改打包形态。启动不再全量触页，但进程 RSS 仍含二进制映射的文件页，属内核可回收部分。彻底压 RSS 需要把冷资产移出 `go:embed`，那会放弃单二进制部署。 |
| 并发 | 常驻装配的换装仍由 `lazyEntry.mu` 串行化；冷装配返回私有对象，不共享、无锁争用。`Warm` 逐条目加锁，与渲染路径不形成锁环。 |
| 铁律兼容 | 未改任何 `Cache-Control`（铁律 1）；未动 name 级限流降级只读（铁律 3）；存储仍是 `counter.Buffer` → SQLite 单路径（铁律 5），热度仍复用第二个 `counter.Buffer`，无新存储。 |

## 6. 验证与人工检验

自动化：`go test ./... -count=1` 全绿（含 `assets`、`cmd/gen-theme-thumbs`、`internal/server`、`internal/imgcore/*`）。关键新增测试：

- `internal/imgcore/asset/header_test.go`：全量图片尺寸一致性 + 头部边界用例。
- `internal/imgcore/composer/resident_test.go`：常驻集边界、`Warm` 替换与释放、冷装配确定性、元数据多帧判定。
- `internal/config/config_test.go`：`THEME_RESIDENT_LIMIT` 默认值、覆盖、非法值。
- `cmd/server/main_test.go`：启动参数优先级（命令行 > 环境变量 > 默认）、两种 flag 写法、`0`、负数/非整数/未知参数报错、`-h`。
- `internal/imgcore/composer/memory_bench_test.go`：可复现基准（只记录不断言绝对数值，主题数量增长不会让测试失败）。

人工检验（AGENTS.md human check），启动后依次执行：

```bash
# 1) 用启动参数指定常驻数量；不传即默认 10
go run ./cmd/server -theme-resident-limit 10
# 查看用法与当前默认值：go run ./cmd/server -h

# 1b) 环境变量与启动参数同时存在时，启动日志应显示 theme_resident_limit=5
THEME_RESIDENT_LIMIT=32 go run ./cmd/server -theme-resident-limit 5
# 非法值应在启动阶段报错退出（退出码 2）
go run ./cmd/server -theme-resident-limit=-1

# 2) 冷门主题（不在 /api/themes/hot 里）仍应正常返回 SVG，且每次背景有变化
curl -s -o /dev/null -w 'cold: %{http_code} %{size_download} bytes\n' \
  'http://127.0.0.1:9721/@demo?theme=pcr-107411'

# 3) 真实计数 SVG 的缓存头必须仍是 no-cache 组合（铁律 1）
curl -sI 'http://127.0.0.1:9721/@my-counter?theme=lian-ren' | grep -i '^cache-control:'

# 4) 主题总数不变、热门列表仍为 10 个
curl -s 'http://127.0.0.1:9721/api/themes' | python3 -c 'import json,sys;print("themes:",len(json.load(sys.stdin)["themes"]))'
curl -s 'http://127.0.0.1:9721/api/themes/hot'

# 5) 缩略图生成器产出与改动前一致（确定性装配）
go run ./cmd/gen-theme-thumbs -out /tmp/theme-thumbs && \
  git status --short web/public/images/theme-thumbs | head

# 6) 内存复测（单独进程，避免前一阶段抬高 HeapSys）
go test ./internal/imgcore/composer -run TestMemoryResidentOnly -v -count=1
```

第 3 步期望 `no-store, no-cache, max-age=0, must-revalidate`；第 5 步期望无新增/修改文件。

## 7. 未落地的候选（按预期收益排序）

| 策略 | 预期收益 | 代价 / 本次未做的原因 |
|---|---|---|
| 请求链路零拷贝（`DecodeImage` 流式 base64、`Compose` 直接写响应、`strings.Builder.Grow`） | 每请求少 2~3 份图字节；当前 `PrepareRender` 约 0.5 MB/op | 需改 `Compose` 返回类型，波及大量测试，属于独立功能 |
| `Sets>0` 套装主题改增量换装 | `PrepareRender` 目前对该分支每次全量 `AssembleTheme` | 会改变随机组合语义，需单独设计回归用例 |
| Live2D / Spine / Unity handler 改流式（`fs.ReadFile` → `fs.Open` + `io.Copy`） | 削掉单个大模型文件的请求峰值 | 独立功能，PSB 已有磁盘流式先例 |
| 构建期生成 catalog 静态清单 | 启动连 32 字节/文件都不用读，catalog 阶段趋近于反序列化一份 JSON | 需接入 `go generate` 与 CI，且要保证与资产树同步 |
| 冷资产外置到磁盘 / `os.DirFS` | 二进制瘦身，rodata RSS 大幅下降 | 放弃「单二进制」形态，Docker 需携带资产目录，属架构决策 |
| `GOMEMLIMIT` + `GOGC` 调优 | 给容器内存设软上限，减少 OOM 风险 | 纯部署配置，不改代码即可做 |
| 热度滑动窗口 / 衰减 | 避免热门榜固化，新主题能进池 | 涉及 `tb_theme_usage` 语义，需评估数据迁移 |
| 前端 1583 张缩略图懒加载 | 降浏览器侧运行内存（约 33 MB 图片） | 前端独立改动 |
