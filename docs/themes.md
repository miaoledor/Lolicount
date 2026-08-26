# 主题文档

> 主题类型、目录约定与使用方式。贡献流程见 [主题贡献指南](./contributing-themes.md)。

Lolicount 的「主题」分为**图片主题**与**文字风格主题**两类,各自独立选择。

## 图片主题

图片主题作为底图展示(层级 0),计数文字叠加在其上(层级 1)。按素材组织
方式又分两种:

### 卡片主题(frame)

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
- `mode=seq`(默认):顺序轮播;`mode=random`:每次请求随机抽帧。

内置卡片主题:`lian`、`kuon`。

### 立绘主题(character)

由多个**透明分层**图片组成,**固定随机模式**,每次请求重新组合服装、
表情等(类似 galgame 立绘):

```
assets/character/<your-theme>/
  ren.json          分层配置(坐标、层级、可选项)
  ren/              分层图片目录
    <layer_id>.<ext>
    ...
```

- 立绘主题**不支持顺序模式**,固定随机。
- 分层坐标与命名遵循 `useLoli` 的约定(见 `web/app/composables/useLoli.ts`)。
- 新增立绘主题需确保 `internal/imgcore/characterthemedrawer` 的 character registry 能扫描到。

内置立绘主题:`lian-ren`。

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

CI 在 PR 改动 `assets/theme/**` 或 `assets/character/**` 时自动运行
`theme-check.yml`。

## 自动修正帧序号

如果卡片主题的帧文件序号不是连续的 `0,1,2,...,size-1`(例如贡献者放了
`1,3,5.png` 或顺序错乱),可用 `cmd/fix-theme` 自动重命名修正:

```bash
go run ./cmd/fix-theme --dry-run   # 预览将重命名的文件(不改文件)
go run ./cmd/fix-theme             # 实际重命名,使序号连续从 0 开始
```

- 只作用于磁盘上的 `assets/theme/`(`embed.FS` 只读,无法运行时改名)。
- 立绘主题(`assets/character/`,含 `ren.json`)自动跳过,不重排分层 id。
- `--dry-run` 发现需修正时会以非零退出码结束,可在 CI 中用作门禁。

## 主题清单

`assets/themes.json` 由 `scripts/gen-themes-json.js` 自动生成,记录卡片主题
的名称、帧数、扩展名与 meta 字段。前端与 API 消费该清单展示主题列表。
**不要手动编辑** `themes.json`,改主题后重跑生成脚本。
