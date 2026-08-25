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

## 主题帧序号修复

`cmd/fix-theme` 扫描 `assets/theme/` 下每个卡片主题目录,把帧图片重命名为
连续的 `0.<ext> 1.<ext> ... n-1.<ext>`。已封装为 pnpm 命令,三端
(Windows / macOS / Linux)通用。适用于:帧序号有缺口(如 `1,3,5.png`)、
顺序错乱、或文件名带非数字前缀导致不连续的情况。

**立绘主题自动跳过**(`assets/character/` 下的 `ren.json` + 图层目录),
因为立绘的图层 id 不是帧序列,不能被重排。

### 两个命令

```bash
pnpm fix-theme:dry    # 预览:只扫描、只打印会怎么改,不动任何文件
pnpm fix-theme        # 执行:真正重命名文件
```

### 典型流程

先 dry-run 预览(养成习惯,先看再改):

```bash
pnpm fix-theme:dry
```

一切正常时:

```
OK    kuon (7 frames, already contiguous)
OK    lian (12 frames, already contiguous)
OK    umi-1 (18 frames, already contiguous)

(dry-run) 0 theme(s) would change, 0 skipped
```

有问题时(退出码为 1,CI 可据此拦截不合规 PR):

```
DRY   my-theme (5 frames, 2 renames):
        3.png -> 1.png
        5.png -> 2.png

(dry-run) 1 theme(s) would change, 0 skipped
```

确认无误后执行修复:

```bash
pnpm fix-theme
```

再跑一次 dry-run 确认已连续:

```bash
pnpm fix-theme:dry
```

应全部显示 `OK (already contiguous)`。

### 支持的命名场景

`fix-theme` 对卡片主题目录支持两种帧命名,均可自动修正为连续 `0..n-1`:

1. **数字命名(可能不连续)**:`0.png 3.png 5.png` → 补缺口为 `0 1 2`。
2. **任意命名(非数字)**:`bs1_um010101_1-1.png bs1_um020101_1-1.png` →
   按文件名升序映射为 `0 1 2`。适合直接从素材包导入、尚未规范命名的主题。

> 混合约定时不混排:若目录里**同时**存在数字名与非数字名图片,只处理数字
> 名帧,非数字名图片被忽略(避免把两种命名约定混在一起重排)。请保持一个
> 主题目录内命名风格统一。

> 例:`umi-1` 主题原始文件名为 `bs1_um*.png`,运行 `pnpm fix-theme` 后
> 自动重命名为 `0.png..17.png`,再 dry 显示 `OK (18 frames, already
> contiguous)`。

### 进阶选项

```bash
# 扫描非默认目录
pnpm fix-theme -- --root assets/theme

# 直接用 go(不用 pnpm)
go run ./cmd/fix-theme --dry-run
go run ./cmd/fix-theme
```

修复后建议一并跑下文「校验流程」中的 `check-theme` / `gen-themes-json`
等命令,确保主题元数据与 manifest 同步。

## 校验流程

提 PR 前,本地跑:

```bash
pnpm fix-theme:dry
go run ./cmd/check-theme
node scripts/validate-theme-meta.js
node scripts/gen-themes-json.js
```

- `pnpm fix-theme:dry`:预览卡片主题帧序号是否连续(`0..n-1`),不改动文件;若有不连续会以非零退出码提示
- `pnpm fix-theme`:执行修复,把不连续的帧图重命名为 `0..n-1`(立绘主题自动跳过)
- `cmd/check-theme`:校验目录名、帧完整性、格式与尺寸
- `scripts/validate-theme-meta.js`:校验 `meta.json` schema
- `scripts/gen-themes-json.js`:校验 `assets/themes.json` 已同步

提交主题前建议一并跑 `pnpm optimize:images:check` 确认 PNG 已无损优化
(详见下文「图片无损优化」),体积过大时跑 `pnpm optimize:images`。

CI 会在 PR 改动 `assets/theme/**` 或 `assets/character/**` 时自动运行
`theme-check.yml`,无需手动触发。

## 图片无损优化

内置主题图经 `embed.FS` 打包进二进制,PNG 体积直接影响最终产物大小。
`scripts/optimize-images.mjs` 封装 [oxipng](https://github.com/shssoichiro/oxipng)
(预编译二进制,经 `oxipng-bin` npm 包分发,无需系统依赖),对
`assets/theme/**/*.png` 与 `assets/character/**/*.png` 做**严格无损**压缩
——只重写 DEFLATE 压缩流与 PNG filter 策略,不改动任何像素,适合对已优化的
PNG 再挤出 10–20% 体积。

> 不要用 sharp / imagemin 的有损减色(`palette: true`)优化主题图:会改变
> 像素,破坏主题视觉一致性。oxipng 是唯一保证像素逐字节不变的无损方案。

### 两个命令

```bash
pnpm optimize:images:check    # 预览:只报告可省多少,不改文件(可优化时退出码 1)
pnpm optimize:images          # 执行:原地无损压缩
```

### 典型流程

先 check 预览(养成习惯,先看再改):

```bash
pnpm optimize:images:check
```

有空间时会输出:

```
check (dry-run): 155 PNG files, 82.0 MiB total, oxipng -o 2
shrinkable: ~13176 KiB (15.7%) could be saved
run `pnpm optimize:images` to apply
```

执行优化:

```bash
pnpm optimize:images
```

再跑一次 check 确认已到位:

```bash
pnpm optimize:images:check
```

应显示 `all PNGs already optimal; nothing to do`。

### 选项

```bash
# 更高压缩级别(0-6,默认 2;越高越小越慢)
pnpm optimize:images --level 4

# 显示每文件详情
pnpm optimize:images --verbose
```

### 范围与边界

- 只处理 `assets/theme` 与 `assets/character` 下的 PNG
- 跳过 `assets/dist`(Nuxt SSG 产物,由 `pnpm generate` 重新生成)
- 跳过 `assets/f-theme`(webp / JSON,非 PNG)
- `--strip safe` 仅移除冗余 metadata,保留色彩配置
- 不影响上传通道(铁律 4 的服务端重编码针对 `/api/themes`、`/api/backgrounds`,
  与本脚本无关)

### 验证无损

优化后建议跑一次 `check-theme` 确认所有主题仍可正常解码:

```bash
go run ./cmd/check-theme
```

如需逐像素验证,可用 Go `image/png` 解码优化前后文件对比像素(项目已有
`image/png` 依赖,无需引入新库):decode 两张图,比较 bounds 与 raw RGBA
字节是否完全一致。

## Web 上传通道

除 PR 外,也可访问 `/upload` 页面上传帧图片,立即可用(服务端重编码,
防图片马)。上传受每 IP 配额限制。

> 安全提示:服务端会对上传图片**解码后按白名单格式重编码**,
> `Content-Type` 与文件后缀都不作为格式判定的唯一依据。

---

下一步:→ [功能贡献指南](./contributing-code.md)
