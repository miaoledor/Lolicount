# 主题贡献指南

> [功能贡献指南](./contributing-code.md) · [返回贡献总览](../CONTRIBUTING.md)

Lolicount 有两种主题类型。请按对应格式准备素材。

## 1. 卡片主题(frame)

每个主题是一个目录,内含若干**帧图片**,文件名即帧索引:

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
- 访问计数按 `(count+1) % 帧数` 轮播展示;`mode=random` 时每次请求随机抽帧。
- 命名保留字:`demo`、`random`,以及已存在的 builtin 主题名,不可冲突。

## 2. 立绘主题(character)

立绘主题由多个**透明分层**图片组成,每次请求随机组合服装、表情等:

```
assets/character/<your-theme>/
  <layers...>
```

- 立绘主题**固定随机模式**,不支持顺序模式。
- 分层坐标与命名遵循 `useLoli` 的约定(见 `web/app/composables/useLoli.ts`)。
- 新增立绘主题需同步更新 `internal/theme` 的 character registry 扫描路径。

## meta.json

可选,用于描述主题元信息:

```json
{
  "name": "lian",
  "author": "yourname",
  "description": "Loli-style digit frames",
  "tags": ["cute", "anime"],
  "version": "1.0.0"
}
```

## 3. 文字风格主题(f-theme)

文字风格主题是一个 JSON 文件,定义计数文字的字体/颜色/粗细,与图片主题解耦:

```
assets/f-theme/<your-ftheme>.json
```

字段:

```json
{
  "name": "neon",
  "family": "monospace",
  "color": "#0ff",
  "weight": "bold"
}
```

- `name`:必须与文件名(不含 `.json`)一致
- `family`:CSS `font-family`,建议用通用字体族(`monospace` / `serif` / `sans-serif`)
- `color`:CSS 颜色值
- `weight`:CSS `font-weight`(`normal` / `bold` / 数字)

使用时通过 `?ftheme=<name>` 参数引用,`random` 随机选一个。

---

## 校验流程

提 PR 前,本地跑:

```bash
go run ./cmd/check-theme
node scripts/validate-theme-meta.js
node scripts/gen-themes-json.js
```

- `cmd/check-theme`:校验目录名、帧完整性、格式与尺寸
- `scripts/validate-theme-meta.js`:校验 `meta.json` schema
- `scripts/gen-themes-json.js`:校验 `assets/themes.json` 已同步

CI 会在 PR 改动 `assets/theme/**` 或 `assets/character/**` 时自动运行
`theme-check.yml`,无需手动触发。

## Web 上传通道

除 PR 外,也可访问 `/upload` 页面上传帧图片,立即可用(服务端重编码,
防图片马)。上传受每 IP 配额限制。

> 安全提示:服务端会对上传图片**解码后按白名单格式重编码**,
> `Content-Type` 与文件后缀都不作为格式判定的唯一依据。

---

下一步:→ [功能贡献指南](./contributing-code.md)
