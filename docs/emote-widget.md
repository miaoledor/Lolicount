# Emote 挂件（PSB 动图计数器）设计文档

> 状态：POC 实现中（分支 `feat/psb-widget`）
> 关联调研结论与资源验证记录见本文末尾「附录：链路验证记录」。

## 1. 背景与目标

Lolicount 当前的输出契约是**静态 SVG 图片**（`<img src="https://lolicount.top/@name">`），
浏览器在 `<img>` 中禁用脚本，因此无法承载需要 WebGL 实时渲染的内容。

PSB（Packaged Struct Binary）是 M2 公司 **E-mote 引擎**的角色动画格式（galgame 看板娘、
动态立绘常用），一个文件内含多层纹理与**多个命名动作（motion / mainTimeline）**。

本功能目标：让第三方网页通过一段 `<script>` 挂件引用计数器——

1. 页面打开时**自动计数 +1**（与 `<img>` 引用 `/@name` 等价）；
2. 渲染 E-mote 角色的 WebGL 动画；
3. **每次打开随机播放一个动作**（`mainTimelineLabels` 随机选取）。

**明确的边界**：GitHub README 只允许 `<img>`，无法运行脚本，PSB 动画在 README 场景
不可用；README 用户继续使用现有静态 SVG 主题，本功能不做额外降级。

## 2. 整体架构

```
第三方网页
  │ <div data-lolicount="name" data-model="azuki" data-text="...{n}..."></div>
  │ <script src="https://host/widget/widget.js" defer></script>
  ▼
widget.js（自研，原生 JS）
  ├─ 动态注入 FreeMoteDriver.js + emoteplayer.js（同源，vendor 自 FreeMote-SDK）
  ├─ fetch GET /psb/azuki        ──► Go 后端：embed.FS 直接回字节（immutable 长缓存）
  ├─ EmotePlayer 初始化（WebGL canvas）
  ├─ mainTimelineLabels 随机选一个动作播放
  └─ fetch GET /api/count/@name  ──► Go 后端：自增计数，返回 JSON（no-store，CORS）
       └─ 渲染计数文字（{n} 模板），默认画布正下方居中
```

与现有系统的关系：

- **不触碰 imgcore 渲染管线**。PSB 模型不是 `imgcore` 的一种主题类型，而是与
  `assets/theme/` 平行的资产类别（`assets/psb/`），挂件路径与 SVG 路径完全解耦。
- **计数语义完全复用** `counter.Buffer`：`/api/count/@name` 与 `counterHandler` 走同一
  套 `incrementOrDegrade`（name 级限流降级只读）与 `demo` / `number` 特例，只是把
  「渲染 SVG」换成「返回 JSON」。
- **缓存铁律不变**：真实计数一律 `no-store`；模型文件是构建期嵌入的不可变字节，
  可以 `max-age=31536000, immutable`。

## 3. 后端接口（Go / Fiber v3）

### 3.1 `GET /api/count/@:name`

挂在 `/api` 前缀下，自动继承现有 `cors()` 中间件（reflect-origin），跨域挂件可直接
`fetch`。中间件链与 SVG 路径一致：`sanitizeBackslashEscape` → `ipRateLimit`。

| 行为 | 说明 |
|---|---|
| 真实 name | `incrementOrDegrade`（自增；name 级限流超限降级只读，不 429） |
| `name=demo` | 不自增，返回固定 `0123456789`（与 SVG 路径特例对齐） |
| `number>0` | 不自增，直接返回该值 |
| 响应 | `{"name":"...","num":123}`，`Cache-Control: no-store` |

### 3.2 `GET /api/psb/models`

列出 `assets/psb/` 下的可用模型：`{"models":[{"name":"azuki"}]}`，
`Cache-Control: public, max-age=60`（与其他 `/api` 列表一致）。目录为空时返回空列表，
不报错。

### 3.3 `GET /psb/:model`

返回 `assets/psb/<model>/model.psb` 的原始字节：

- `Content-Type: application/octet-stream`
- `Cache-Control: public, max-age=31536000, immutable`（内容随构建固定）
- 模型名做白名单校验（`^[a-z0-9-]+$`，且必须存在于嵌入目录），防路径穿越；
  不存在返回 404。

### 3.4 路由注册顺序

三条路由都在 `registerRoutes()` 中、`registerFrontend()` 之前注册（Fiber 按注册顺序
匹配，无需改 `isDynamicPath`）。

## 4. 模型资产

### 4.1 目录约定

```
assets/psb/
  README.md            # 放置说明（入库）
  azuki/
    model.psb          # pure、spec=ems 的 PSB（不入 git，本地/部署时放置）
```

`assets/embed.go` 增加 `//go:embed all:psb`。模型文件体积大（数 MB～数十 MB）且有
版权风险，**一律不提交进仓库**（`assets/psb/*/model.psb` 进 `.gitignore`），目录里
只提交说明文档；构建时本地放了什么就嵌入什么，没放则模型列表为空、演示页空态。

### 4.2 模型从哪来 / 如何转换（离线一次性流程）

浏览器驱动（E-mote 3.9 WebGL 版）只能吃 **pure、`spec=ems` 的 PSB 原始字节**。
常见的游戏分发 PSB 有多种包装（LZ4/MDF/PSZ 壳、加密、`spec=win` 纹理布局），
需要离线归一化：

1. **来源**：使用你有权使用的模型（自制 / 有授权）。商用游戏里提取的模型只能本地
   测试，不可公开分发。
2. **转换工具**：[703519523/Emote_Widget](https://github.com/703519523/Emote_Widget)
   的 `emote_widget/utils/psb_converter`（纯 Python 标准库，约 850 行），流程为
   「解壳 → 解密 → Win→EMS 纹理字节序适配（仅 RGBA8）」：

   ```python
   from psb_converter import PsbNormalizer, adapt_win_psb_to_ems
   result = PsbNormalizer(src, require_win_spec=False).normalize_with_summary()
   data = adapt_win_psb_to_ems(result.data) if result.summary["spec"] == "win" else result.data
   ```

   本地验证过的转换结果见附录。
3. 放入 `assets/psb/<name>/model.psb`（目录名仅小写字母/数字/连字符），重启服务生效。

> 未来增强（不在本次范围）：把 psb_converter 的归一化逻辑移植为 Go 包
> （它只用 zlib/struct，无第三方依赖），实现「用户直接丢原始 .psb，服务端启动时
> 自动归一化」。DXT 压缩纹理的 win 模型目前 Python 工具也不支持，维持离线转换
> 是刻意收窄的 POC 边界。

## 5. 挂件脚本（`web/public/widget/`）

| 文件 | 来源 | 说明 |
|---|---|---|
| `FreeMoteDriver.js` | vendor（FreeMote-SDK，经 Emote_Widget 仓库） | E-mote 3.9 引擎的 emscripten 编译产物，约 1.2MB |
| `emoteplayer.js` | vendor（同上） | `EmotePlayer` 封装类（加载/动作/变量/渲染循环） |
| `widget.js` | 自研 | 挂件入口，原生 JS IIFE，无构建依赖 |
| `THIRD-PARTY-NOTICE.md` | 自建 | CC BY-NC-SA 4.0 声明与出处 |

放入 Nuxt `web/public/` 后随 dist 构建进入 `embed.FS`，由现有 frontend catch-all
自动 serve（`/widget/widget.js` 等），**后端零改动**。

### widget.js 职责

1. 扫描 `document.querySelectorAll('[data-lolicount]')`（支持多个挂件实例）；
2. 读取属性：`data-lolicount`（计数名）、`data-model`（模型名，默认取模型列表第一个）、
   `data-text`（`{n}` 模板，缺省为纯数字）、`data-width`/`data-height`（画布尺寸，
   默认 320×480）；
3. 依序动态注入同源 `FreeMoteDriver.js` → `emoteplayer.js`（一次加载全局共享）；
4. `EmotePlayer.createRenderCanvas` + `new EmotePlayer(canvas)`，从 `/psb/<model>`
   拉取字节加载模型，按 `charaBounds` 自适应居中缩放；
5. `mainTimelineLabels` 随机取一个动作赋给 `mainTimelineLabel` 播放；
6. `fetch('/api/count/@name')` 取计数值，`{n}` 替换后渲染在画布下方；
7. WebGL 不可用 / 模型加载失败时，在挂件容器内渲染文字错误提示，不抛白屏。

## 6. 演示页（`web/app/pages/emote.vue`）

- 模型选择器（数据来自 `/api/psb/models`，空态显示放置模型指引）；
- WebGL 实时预览 + 「换个动作」按钮（重新随机）+ 当前动作名展示；
- 生成可复制的挂件引用代码（复用现有复制交互模式）；
- zh / en / ja 三语言文案；首页加入口链接。

## 7. 缓存与限流对照（对齐 AGENTS.md 铁律）

| 资源 | Cache-Control | 理由 |
|---|---|---|
| `/api/count/@name`（真实计数） | `no-store` | 铁律 1：真实计数绝不缓存 |
| `/api/count/@demo` 等 | `no-store` | 与 SVG 路径 demo 特例行为一致（不引入长缓存分支） |
| `/api/psb/models` | `public, max-age=60` | 短缓存，对齐其他 `/api` 列表 |
| `/psb/:model` | `public, max-age=31536000, immutable` | 构建期嵌入，字节不可变 |

限流：`/api/count/@name` 复用 `ipRateLimit` + name 级 `incrementOrDegrade`，语义与
SVG 路径完全一致，不新增阈值。

## 8. 许可与风险

- **FreeMoteDriver.js / emoteplayer.js**：CC BY-NC-SA 4.0（禁商用），来自
  Project-AZUSA/FreeMote-SDK。已附 `THIRD-PARTY-NOTICE.md`。Lolicount 当前为非商用
  免费服务，不冲突；**若未来商业化，必须替换为 M2 官方 WebGL SDK 授权**。
- 引擎为 E-mote 3.9（2018）编译版，过新的模型可能加载失败。
- 模型素材版权归各自作者：仓库只收录放置流程，不收录任何模型文件；公开部署时
  运营者需自行确保模型授权。
- 未知交互（点击反应、视线跟随等）暂不实现，POC 只做「随机动作 + 计数」。

## 附录：链路验证记录（2026-09-04）

- 驱动加载验证：Emote_Widget 的 web_frontend 即同一套
  `FreeMoteDriver.js` + `emoteplayer.js`，`EmotePlayer_Initialize(Uint8Array[])`
  直接接受 pure ems PSB 字节（`promiseLoadDataFromURL` → `getBinaryAsync` →
  XHR arraybuffer → `loadData`）。
- 模型转换验证（工具：psb_converter，本机 Python 3.12）：
  - `chara.psb`（Emote_Widget 自带示例，18MB）：raw `PSB\0`、`spec=ems`，**无需转换**；
  - `dx_e-moteアズキ私服a.psb`（Nekopara DX，3.1MB）：LZ4 壳 + `spec=win`，
    解壳后 18.4MB，Win→EMS 适配成功（RGBA8 纹理）；
  - `dx_e-moteアズキ私服a大トロ.psd`（同上变体，3.0MB）：同上，17.3MB 转换成功。
  - 转换产物：`E:\psb-out\*.ems.psb`（本地测试用，不入仓库）。
- 随机动作：`player.mainTimelineLabels` 为动作名数组，随机索引赋给
  `player.mainTimelineLabel` 即播放（Emote_Widget 的 core_renderer.js 同样从
  `mainTimelineLabels` 取列表，验证了该 API 可用性）。
