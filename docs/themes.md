# 主题文档

> 主题模型、目录约定与使用方式。贡献流程见 [主题贡献指南](./contributing-themes.md)。

Lolicount 的「主题」分为**图片主题**与**文字风格主题**两类,各自独立选择。

## 图片主题

所有图片主题统一为有序图层栈,计数文字作为其中一个图层。按素材组织方式
分两种(加载时按目录结构自动分派,不作为架构分支):

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
- 每次请求从帧集合中随机抽取一张展示(多帧主题不再按计数顺序轮播)。

内置单图层主题数量较多,例如 `lian`、`kuon`、`umi-1`;完整实时列表以
`GET /api/themes` 或 `assets/themes.json` 为准。

### 多图层主题(原立绘,character)

由多个**透明分层**图片组成,每次请求随机重新组合服装、表情等
(类似 galgame 立绘):

```
assets/theme/<your-theme>/
  ren.json          图层清单(坐标、尺寸、layer_id)
  config.json       画布尺寸与分类范围
  display.json      输出尺寸与裁剪框(可选)
  ren/              分层图片目录
    <layer_id>.<ext>
    ...
```

- 每个分类范围在请求时随机选择一个候选图层;分类按 `first` 索引升序叠放。
- 分层坐标与命名由后端 `asset.LoadCharacterTheme` 解析(见 `internal/imgcore/asset/character_loader.go`)。多图层主题与单图层主题统一存放在 `assets/theme/` 下。
- 新增多图层主题需确保 `composer.NewThemeRegistry()` 能扫描到 `assets/theme/`。

内置多图层主题例如 `lian-ren`、`furi`、`cafestella-yuna`;主题名不要求
`-ren` 后缀,加载器以目录中是否存在 `ren.json` 判断类型。

#### 立绘主题 JSON 文档

立绘主题目录下有两个必需 JSON 文件与一个可选 JSON 文件,各自承担不同职责:

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
- `layer_id` 可以不连续;`config.json` 的 `ranges` 决定哪些数组项参与随机选择。
- `visible` 仅保留 PSD 元数据,运行时不据此过滤;图层是否参与渲染完全由
  `config.json` 的范围决定。
- 活动范围内的图层必须有正数 `layer_id`、正数宽高以及对应的
  `ren/<layer_id>.<ext>` 图片。范围外允许保留 `layer_id: 0` 的兼容占位项。

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

- 代码按各范围的 `first` 从小到大排序渲染,较小索引在底层。
- 每个部位从 `[first, last]` 范围内随机选一个图层,多个部位叠成完整立绘。
- 部位名可扩展;`lass`(身体)、`eye`(眼睛)、`brow`(眉毛)、`mouth`(嘴)、
  `face`(脸)是当前工具使用的常见约定,不是运行时白名单。
- `first`/`last` 是 `ren.json` 数组下标,不是 `layer_id`。

##### 互斥套装:`sets`(可选)

一个主题里的身体若分属差异较大的造型(例如自带围巾的蒙面装与露脸的常服),
扁平的 `ranges` 会让各部位独立随机:正脸的嘴/腮红会被画到围巾上,
不同朝向的身体也会抽到不配套的五官。此时可用 `sets` 把图层划成若干
**互斥套装**——渲染时先随机选一个套装,再只在该套装的范围内挑部位:

```json
{
  "canvasW": 1703,
  "canvasH": 2413,
  "sets": [
    {
      "name": "casual",
      "canvasW": 1703,
      "canvasH": 2413,
      "ranges": {
        "lass": { "first": 1,  "last": 7  },
        "eye":  { "first": 13, "last": 96 },
        "face": { "first": 97, "last": 98 }
      }
    },
    {
      "name": "ninja",
      "canvasW": 1703,
      "canvasH": 2413,
      "ranges": {
        "lass": { "first": 8,  "last": 12 },
        "eye":  { "first": 13, "last": 96 }
      }
    }
  ]
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `sets` | array | 非空时优先于顶层 `ranges`;`check-theme` 逐套装校验 |
| `sets[].name` | string | 仅用于报错定位与测试契约 |
| `sets[].canvasW` / `canvasH` | int | 该套装的原始画布尺寸,必须为正 |
| `sets[].ranges` | object | 该套装专属的部位范围,可以缺少某些部位 |

- 同一 `ren.json` 索引允许出现在不同套装里(如上例共用的 `eye`),
  但**同一套装内部**的部位范围不得重叠。
- 变化数是各套装变化数之**和**而非乘积,因为套装互斥。
- 缺少的部位直接不渲染:蒙面套装不提供 `face`,围巾上就不会多出一张嘴。
- 实例:`assets/theme/senren-mako/config.json`、
  `assets/theme/senren-mako-b/config.json`(千恋＊万花 / 常陸茉子)。

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
- `demo` 为保留值:固定渲染 `0123456789`,不落库。单帧主题(包括显式
  `number>0`)长缓存;多帧/随机组合主题即使指定 `number` 也使用 no-cache
  组合头,因为背景仍会变化。
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
| `theme` | 图片主题 | 主题名,或 `random` |
| ~~`mode`~~ | ~~已移除~~ | ~~所有主题统一随机~~ |
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
node scripts/gen-themes-json.js       # 重新生成全部静态主题的 themes.json
```

CI 在 PR 改动 `assets/theme/**` 时自动运行
`theme-check.yml`。

## 自动修正帧序号

如果卡片主题的帧文件序号不是连续的 `0,1,2,...,size-1`(例如贡献者放了
`1,3,5.png` 或顺序错乱),可用 `cmd/fix-theme` 自动重命名修正:

```bash
go run ./cmd/fix-theme --dry-run   # 预览将重命名的文件(不改文件)
go run ./cmd/fix-theme             # 实际重命名,使序号连续从 0 开始
```

- 只作用于磁盘上的 `assets/theme/`(`embed.FS` 只读,无法运行时改名)。
- 多图层主题(`assets/theme/`,含 `ren.json`)自动跳过,不重排分层 id。
- `--dry-run` 发现需修正时会以非零退出码结束,可在 CI 中用作门禁。

## 主题清单

`assets/themes.json` 由 `scripts/gen-themes-json.js` 自动生成,记录所有内置
主题的名称、图片数量、扩展名与 meta 字段。单图层主题统计根目录帧图,
多图层主题统计 `ren/` 目录下的图层图。运行时 `/api/themes` 由
`ThemeRegistry` 直接生成,清单主要供构建与主题工具使用。
**不要手动编辑** `themes.json`,改主题后重跑生成脚本。
