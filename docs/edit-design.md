现在我要给lolicount添加在线编辑多图层立绘主题页面，编辑时新增或修改json
其中lass，brow，eye，mouth，face分别作为一个图层文件夹，
每个图层文件夹中只会随机显示文件夹中的一张图片，
用户可以在前端创建图层并编辑该图层显示在第几层
用户可以给该图层上传图片，并编辑该图片的位置，可以拖动动态查看该图片位置的效果
用户可以调整该图层的图片等选转和放缩
用户可以点击合并图层查看效果
用户可以点击展开图层，并选择想要编辑的图层
图层被展开后按图层排在第几层排序
显示当前在第几层图层
编辑完成后显示在审核中，并且其他用户可以点击赞同或者否定
其他用户也可以添加评论
需要修改当前的json结构，使用英文作为变量名
需要调节当前的目录结构，lass，brow，eye，mouth，face相关的图片要分别放在一个文件夹中
且在当前图层文件夹中重新从0，1，2...开始命名
传入文件自动进行压缩，尽量采用无损，采用webp
将该功能命名为图片编辑工作台
将图片编辑工作台后端部分放入imgcore
注意前端样式，要兼容电脑和手机端
图片编辑工作台采用一个新的页面，不在原来的page
在原上方栏目中跳转到新页面
图片编辑工作台采用较为简约的页面和较为优雅的动画效果
在展开图层选择中采用较为优雅的动画效果
[    ]叠加效果
[    ]


三个图层的展开效果
[      ]
[  {   ]  }
   {  [   }   ]
      [       ]

在点击展开时要有较为优雅的动画效果展开

选择好图层后出现该图层是第几层
可以对该图层选择上传图片
最上层有个文字层
可以编辑文字和展示计数器
检测图层数，一层（不包文字层）则归入卡片主题
多层则归入立绘主题
保存后则进入审核阶段
审核阶段可以采用投票功能，表述赞同或者否定，只能投票一次，暂时采用本地存储投票信息方案
在保存时可以编辑主题名称
在前端页面展示推荐添加卡片主题
在修改的时候不严格交检内容，直到要保存时，才检查内容是否合法，或者在预览时也要检查内容是否合法，并展示不合法的提示
在投票达到50的时候该主题会自动通过审核，或者由人工使用密钥通过审核
在最上层会自动添加文字图层，有特别的标记表示为文字图层，无法删除
在最下方自动添加底图层，无法删除
在文字图层会有一系列修改内容，包括文字的移动，加入计数器，文字颜色主题，文字弧度，文字旋转
使用拖动栏实时修改文字位置 右边和下面有拖动栏可以拖动文字
【     】｜
【     】｜<
 ---^---



需要重构绘画核心以应对大版本更新
文字图层不做保存，是每个用户使用的时候的编辑内容，在编辑工作台编辑的内容仅可预览
---

## 主题上传方案决策

### 采用方案 B：运行时主题目录（双 Registry）

在 `embed.FS` 之外新增运行时目录 `data/themes/<name>/`，`ThemeRegistry` 同时扫描两个来源。审核通过的主题写入 `data/themes/`，即时生效，无需重启或重建。

- **草稿阶段**：草稿数据（含图片 base64、图层 JSON、投票信息）存 SQLite 独立表 `tb_theme_draft`，不碰 `tb_count`，不违背铁律 5。
- **审核通过**：从草稿提取图片，经服务端重编码（WebP，铁律 4）后写入 `data/themes/<name>/`（按新 schema 目录结构 `lass/0.webp` 等），同时更新草稿状态为 `approved`。
- **Registry 改造**：`unifiedRegistry` 持有一个额外的 `runtimeRegistry`（读 `os.DirFS("data/themes")`），`Get`/`List` 先查 builtin（embed.FS）再查 runtime。`theme_loader.go`/`character_loader.go` 已接受 `fs.FS` 接口，改为 `os.DirFS` 几乎零成本。
- **持久化**：Docker volume `/app/data` 已挂载，`data/themes/` 天然持久化。
- **单实例约束**：与铁律 5 的单实例约束一致，不预设水平扩展。

### 导出与官方静态库主题合并

审核通过的主题支持导出为标准主题包，便于合并到官方静态库（`assets/theme/`）：

- **导出格式**：导出为符合新 schema 的目录结构压缩包（`<name>.zip`），内含 `ren.json`（或新 schema JSON）、`config.json`、`display.json`、分层图目录（`lass/`、`brow/`、`eye/`、`mouth/`、`face/`，图片从 0 开始命名，WebP 格式）。
- **导出入口**：编辑工作台 / 管理后台提供「导出主题包」按钮，将 `data/themes/<name>/` 打包为 zip 下载。
- **合并到官方库**：导出的 zip 可作为 PR 提交到官方仓库的 `assets/theme/`，经 `cmd/check-theme` 校验通过后合并进 `embed.FS`，成为 builtin 主题。合并后该主题从运行时目录迁移为编译期静态资源，重启后由 `embed.FS` 提供。
- **去重**：合并到官方库后，运行时目录中的同名主题可清理（Registry 优先查 builtin，builtin 命中后不查 runtime）。
- **CI 联动**：PR 改动 `assets/theme/**` 触发 `theme-check.yml` 校验 + `rebuild-frontend.yml` 重建 SSG，与现有 CI 流程一致。

### 导出标准（解压到 assets/ 对应目录即可直接使用）

导出包必须通过 `cmd/check-theme` 的全部校验规则，解压到 `assets/theme/` 目录下即可被 Registry 加载使用，无需任何额外处理。

#### 通用规则

- **目录名**：仅允许 ASCII 字母（大小写）、数字、连字符；不能为保留字 `demo`、`random`；与解压目标目录下已有主题不能重名。
- **图片格式**：仅 `.gif` / `.png` / `.webp`；服务端重编码后统一输出 WebP（铁律 4）。
- **单文件体积**：≤ 4 MiB（`maxFileBytes = 4 * 1024 * 1024`）。
- **文件清理**：导出时不包含 `.DS_Store` 等 dotfile（check-theme 跳过 dotfile，但导出包应保持干净）。
- **可选 meta.json**：卡片主题可附带 `meta.json`（任意合法 JSON），立绘主题不使用 `meta.json`。

#### 卡片主题导出标准（目标：`assets/theme/<name>/`）

单图层主题（编辑工作台中不包含文字层时仅一层）导出为卡片主题。结构：

```
<name>/
  0.webp
  1.webp
  ...
  n-1.webp
  meta.json   (可选)
```

- 帧文件命名为 `<int>.<ext>`，索引从 0 开始连续递增（`0..n-1`），不允许跳号或缺号。
- 同一主题内所有帧使用相同扩展名（混合扩展名会被 check-theme 警告）。
- 每帧图片宽高 ≤ 2048px（`maxFrameSide`）。
- 至少 1 帧图。
- 导出包为 `<name>.zip`，解压后顶层目录即为 `<name>/`，可直接放入 `assets/theme/`。

#### 立绘主题导出标准（目标：`assets/theme/<name>/`）

多图层主题导出为立绘主题。结构（兼容现有 loader + check-theme）：

```
<name>/
  ren.json
  config.json
  display.json   (可选)
  ren/
    1.webp
    2.webp
    ...
    N.webp
```

- **ren.json**：图层清单数组，非空，每个元素描述一个图层的绝对位置。字段：`name`、`left`、`top`、`width`、`height`、`visible`、`layer_id`、`group_layer_id`。导出时按编辑工作台的图层顺序生成，`layer_id` 从 1 开始与 `ren/` 下文件名对应。
- **config.json**：画布尺寸 + 分类区间。字段：`canvasW`、`canvasH`、`ranges`（各分类的 `first`/`last` 闭区间，1-based）。编辑工作台的图层分类（lass/brow/eye/mouth/face 等）映射到 `ranges`。
- **display.json**（可选）：输出尺寸 + 裁剪。字段：`size`、`crop`（`left`/`top`/`width`/`height`）。
- **ren/ 目录**：图层图片，命名为 `<layer_id>.<ext>`（`layer_id` 为正整数，与 `ren.json` 中 `layer_id` 对应）。每张图宽高 ≤ 4096px（`maxCharLayerSide`）。
- 至少 1 张图层图片。
- 导出包为 `<name>.zip`，解压后顶层目录即为 `<name>/`，可直接放入 `assets/theme/`。

#### 导出流程

1. 用户在编辑工作台 / 管理后台点击「导出主题包」。
2. 服务端从 `data/themes/<name>/`（或草稿数据）读取主题内容。
3. 按主题结构（有 `ren.json` 为多图层/立绘包，否则为单图层/卡片包）判定导出格式。
4. 图片统一重编码为 WebP（铁律 4，服务端解码后重编码，不信任原格式）。
5. 按上述标准生成目录结构，打包为 `<name>.zip`（顶层为 `<name>/` 目录）。
6. 用户下载 zip，解压到 `assets/theme/`，本地运行 `go run ./cmd/check-theme` 验证通过后即可提交 PR。