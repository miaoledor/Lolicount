# 主题文档

> 主题类型、目录约定与使用方式。贡献流程见 [主题贡献指南](./contributing-themes.md)。

Lolicount 的「主题」分为**图片主题**与**文字风格主题**两类,各自独立选择。

## 图片主题

所有图片主题统一为有序图层栈,计数文字作为其中一个图层。按素材组织方式
分两种(运行时由 `IsCardTheme()` 推断,不作为架构分支):

### 单图层主题(原卡片,frame)

一个目录内含若干**帧图**,文件名即帧索引:

```
assets/theme/<your-theme>/
  0.webp
  1.webp
  2.webp
  ...
  meta.json   (可选)
```

- 帧索引从 `0` 开始,**必须连续递增**(`0, 1, 2, ...`)。
- 支持 `gif` / `png` / `webp`,一个主题内可混用但建议统一。
- 访问计数按 `(count+1) % 帧数` 轮播展示。
- `mode=seq`(默认):顺序轮播;`mode=random`:每次请求随机抽帧。`mode` 对所有主题生效。

内置单图层主题:`lian`、`kuon`。

### 多图层主题(原立绘,character)

由多个**透明分层**图片组成,每次请求随机重新组合服装、表情等
(类似 galgame 立绘):

```
assets/character/<your-theme>/
  ren.json          分层配置(坐标、层级、可选项)
  ren/              分层图片目录
    <layer_id>.<ext>
    ...
```

- 多图层主题也支持 `mode` 参数(与单图层主题统一)。
- 分层坐标与命名遵循 `useLoli` 的约定(见 `web/app/composables/useLoli.ts`)。
- 新增多图层主题需确保 `composer.NewThemeRegistry()` 能扫描到 `assets/character/`。

内置多图层主题:`lian-ren`。

#### 立绘主题 JSON 文档

立绘主题目录下有三个 JSON 文件,各自承担不同职责:

##### `ren.json` — 图层清单

PSD 导出的图层坐标清单,是一个数组,每个元素描述原图中的一层:

```json
{
  "name": "笑１",
  "left": 232,
  "top": 380,
  "width": 78,
  "height": 29,
  "visible": 1,
  "layer_id": 2,
  "group_layer_id": 2923
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | string | PSD 图层名(日文/中文),仅备注,不参与渲染 |
| `left` | int | 该层在 PSD 画布上的左上角 X 坐标(像素) |
| `top` | int | 该层在 PSD 画布上的左上角 Y 坐标(像素) |
| `width` | int | 图层宽度(像素) |
| `height` | int | 图层高度(像素) |
| `visible` | int | PSD 中是否可见(`1`=可见,`0`=隐藏) |
| `layer_id` | int | 图层编号,对应 `ren/` 目录下的文件名(如 `2` → `ren/2.webp`) |
| `group_layer_id` | int | PSD 中的分组 ID(当前代码未使用,保留) |

- 数组下标 ≠ `layer_id`,代码用 `layer_id` 去找 `ren/<layer_id>.<ext>` 文件。
- `layer_id` 可以不连续;`config.json` 的 `ranges` 决定哪些 `layer_id` 参与随机选择。
- `visible=0` 的层仍会被加载解码,但不会被 `config.json` 的范围选中时不会显示。

##### `config.json` — 画布尺寸与部位分组

定义 PSD 原始画布大小,并把 `ren.json` 的图层按部位分成若干闭区间:

```json
{
  "canvasW": 504,
  "canvasH": 925,
  "ranges": {
    "brow":  { "first": 1,  "last": 18  },
    "eye":   { "first": 19, "last": 36  },
    "mouth": { "first": 37, "last": 56  },
    "face":  { "first": 57, "last": 62  },
    "lass":  { "first": 63, "last": 70  }
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `canvasW` | int | PSD 原始画布宽度(像素) |
| `canvasH` | int | PSD 原始画布高度(像素) |
| `ranges` | object | 部位名 → 图层范围的映射 |
| `ranges.<name>.first` | int | 该部位在 `ren.json` 数组中的起始下标(0-based,闭区间) |
| `ranges.<name>.last` | int | 该部位在 `ren.json` 数组中的结束下标(闭区间) |

- 代码固定按 `["lass", "eye", "brow", "mouth", "face"]` 顺序遍历这 5 个 key。
- 每个部位从 `[first, last]` 范围内随机选一个图层,5 个部位叠成完整立绘。
- 部位名是约定值,不是任意名;当前支持的部位:`lass`(身体)、`eye`(眼睛)、`brow`(眉毛)、`mouth`(嘴)、`face`(脸)。
- 主题可以缺少某些部位(如 `hinata` 没有 `brow`),代码会跳过不存在的 key。
- `first`/`last` 是 `ren.json` 数组下标,不是 `layer_id`。

##### `display.json` — 输出尺寸与裁剪框

控制最终输出图片的大小和取景范围:

```json
{
  "size": 400,
  "crop": {
    "left": 137,
    "top": 323,
    "width": 367,
    "height": 602
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `size` | int | 输出图片的目标高度(像素);宽度按裁剪框比例自动计算 |
| `crop` | object | 裁剪框,只显示 PSD 画布中的这块区域(可选) |
| `crop.left` | int | 裁剪框左上角 X(PSD 原始坐标) |
| `crop.top` | int | 裁剪框左上角 Y(PSD 原始坐标) |
| `crop.width` | int | 裁剪框宽度 |
| `crop.height` | int | 裁剪框高度 |

- `size=400` → 输出高度固定 400px,宽度 = `crop.width × (size / crop.height)`。
- `crop` 对应渲染时嵌套 `<svg viewBox="left top width height">`,把该区域映射到输出视口。
- 不写 `display.json` 时,走请求参数 `scale` 按最长边缩放(默认 400px)。
- 裁剪框用于去掉 PSD 画布周围的空白,只保留人物区域。

##### 三者协作关系

```
ren.json       "layer 2 在画布 (232,380) 处,宽 78 高 29"
     ↕ layer_id 关联 ren/2.webp
config.json    "layer 1~18 是眉毛,随机选一个"
     ↕ 画布尺寸 504×925
display.json   "只显示 (137,323) 那块,缩放到高 400"
     ↓
最终输出       243×400 的立绘 SVG + 计数文字
```

### 共同约定

- 命名保留字:`demo`、`random`,以及已存在的 builtin 主题名,不可冲突。
- 主题名只允许字母(大小写均可)、数字、连字符(`-`)。
- `demo` 为保留值:固定返回 `0123456789`,不落库,长缓存。
- `random` 为保留值:每次请求从 builtin 列表随机挑一个。

## 文字风格主题(f-theme)

文字风格主题是一个 JSON 文件,定义计数文字的字体/颜色/粗细,与图片主题
解耦,通过 `?ftheme=<name>` 参数引用:

```
assets/f-theme/<your-ftheme>.json
```

字段:

| 字段 | 说明 | 示例 |
|---|---|---|
| `name` | 名称,须与文件名一致 | `neon` |
| `family` | CSS `font-family` | `monospace` |
| `color` | CSS 颜色值 | `#0ff` |
| `weight` | CSS `font-weight` | `bold` |

内置文字风格:`default`、`neon`、`pink`、`serif`。

## 参数与主题的关系

| 参数 | 作用于 | 说明 |
|---|---|---|
| `theme` | 图片主题 | 卡片或立绘主题名,或 `random` |
| `mode` | 图片主题(仅卡片) | `seq` / `random` |
| `ftheme` | 文字风格 | 字体/颜色/粗细,或 `random` |
| `fsize` | 文字大小 | 字号(px),与 `scale` 独立 |
| `scale` | 图片大小 | 基于统一 400px 最长边的倍数 |
| `unshowf` | 文字显隐 | `true` 隐藏计数文字 |
| `x`/`y` | 文字位置 | 像素坐标(相对图片左上角) |
| `rx`/`ry` | 文字位置 | 比例坐标(0~1,相对图片宽高) |
| `number` | 预览 | 指定数字直接展示,不落库不 +1 |

> `x`/`y` 与 `rx`/`ry` 二选一;都不传时文字默认显示在图片正下方居中。

## 校验

提 PR 前本地跑:

```bash
go run ./cmd/check-theme          # 校验卡片 + 立绘主题完整性
node scripts/validate-theme-meta.js   # 校验 meta.json schema
node scripts/gen-themes-json.js       # 校验 themes.json 同步(卡片主题)
```

CI 在 PR 改动 `assets/theme/**` 或 `assets/character/**` 时自动运行(两个目录均触发校验)
`theme-check.yml`。

## 自动修正帧序号

如果卡片主题的帧文件序号不是连续的 `0,1,2,...,size-1`(例如贡献者放了
`1,3,5.png` 或顺序错乱),可用 `cmd/fix-theme` 自动重命名修正:

```bash
go run ./cmd/fix-theme --dry-run   # 预览将重命名的文件(不改文件)
go run ./cmd/fix-theme             # 实际重命名,使序号连续从 0 开始
```

- 只作用于磁盘上的 `assets/theme/`(`embed.FS` 只读,无法运行时改名)。
- 多图层主题(`assets/character/`,含 `ren.json`)自动跳过,不重排分层 id。
- `--dry-run` 发现需修正时会以非零退出码结束,可在 CI 中用作门禁。

## 主题清单

`assets/themes.json` 由 `scripts/gen-themes-json.js` 自动生成,记录卡片主题
的名称、帧数、扩展名与 meta 字段。前端与 API 消费该清单展示主题列表。
**不要手动编辑** `themes.json`,改主题后重跑生成脚本。
