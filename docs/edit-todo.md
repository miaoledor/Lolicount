# 图片编辑工作台 — 任务计划清单

> 基于 `docs/edit-design.md` 整理，已剔除当前架构无法实现或不在本次范围内的内容。
> 后端代码放入 `internal/imgcore` 和 `internal/server`；前端为新 Nuxt 页面。
> 编辑内容暂存在前端 localStorage，不需要后端草稿持久化。

## 需求摘要

在线编辑多图层主题，实时预览 SVG 渲染效果，编辑内容暂存 localStorage，支持导出为标准主题包。

## 已剔除的需求

以下需求因技术约束或不在本次范围，不列入任务：

- **文字弧度（arc text）**：弧度文字编辑投入产出比低，不实现；相关未接入的工具函数（原 `textpath.go` 的 `ArcPath`/`RenderArcText`）已作为死代码删除。
- **评论功能**：超出本次范围。
- **"推荐添加卡片主题"前端展示**：非核心功能，后续按需补充。
- **草稿持久化与审核流程**：编辑内容暂存前端 localStorage，不需要后端草稿表、投票审核、密钥通过等。后续如需审核流程再单独设计。
- **Runtime Registry 双来源热加载**：无草稿写入 `data/themes/` 的流程，不需要 runtime registry 层。

## 任务清单

### 阶段 1：后端

- [x] **T1.1 预览渲染端点**
  - `POST /api/editor/preview`：接收图层 JSON → 构建 `*theme.Theme` → `composer.Compose` → 返回 SVG
  - 复用 `buildThemeLayers` 逻辑，图层来自请求体而非 registry
  - 不落库，纯内存渲染
  - 校验：图层数 ≥ 1、图片可解码、尺寸合规（不严格校验，返回错误提示）
  - 独立限流，不复用计数路径配额

- [x] **T1.2 导出端点**
  - `POST /api/editor/export`：接收图层 JSON + 主题名称 → 服务端重编码图片为 WebP → 打包 `<name>.zip` 返回
  - 卡片：`0..n-1.webp` + 可选 `meta.json`
  - 立绘：`ren.json` + `config.json` + 可选 `display.json` + `ren/<layer_id>.webp`
  - 图片统一 `asset.ReEncodeImage(raw, EncodeWebP)` 重编码（铁律 4，不信任原格式）
  - 导出包通过 `cmd/check-theme` 全部校验规则
  - 校验：名称保留字 `demo`/`random`、不与 builtin 冲突、ASCII 字母/数字/连字符

### 阶段 2：前端编辑工作台

- [x] **T2.1 新页面与导航**
  - 新建 `web/app/pages/editor.vue`（单一根元素，与 `index.vue` 平级）
  - `NavBar.vue` 添加跳转链接到 `/editor`
  - 简约风格，兼容电脑和手机端

- [x] **T2.2 画布预览组件**
  - `web/app/components/editor/EditorCanvas.vue`
  - 调 `POST /api/editor/preview` 实时渲染 SVG
  - 显示画布尺寸、当前图层数
  - 文字层预览（可编辑文字内容、展示计数器）

- [x] **T2.3 图层面板组件**
  - `web/app/components/editor/LayerPanel.vue` + `LayerItem.vue`
  - 图层列表，按 ZIndex 排序，显示当前第几层
  - 展开/折叠动画（GSAP）
  - 创建图层、删除图层（文字层和底图层不可删除）
  - 拖拽排序调整图层顺序
  - 最上层自动添加文字图层（特殊标记，不可删除）
  - 最下层自动添加底图层（不可删除）

- [x] **T2.4 图层编辑控件**
  - 上传图片到当前图层（图片自动压缩，WebP 无损）
  - 编辑图片位置（拖动动态查看效果）
  - 编辑图片旋转和缩放（滑块控件）
  - 图层分类选择（lass/brow/eye/mouth/face）
  - 多图上传：同一图层多张图，运行时随机选一

- [x] **T2.5 文字层控件**
  - `web/app/components/editor/TextLayerControls.vue`
  - 文字内容编辑 + 计数器预览
  - 文字位置拖动栏（右边和下边拖动条，实时修改 X/Y）
  - 文字颜色主题选择（复用 f-theme 列表）
  - 文字字号、旋转
  - 文字层不做保存，仅预览

- [x] **T2.6 localStorage 暂存**
  - 编辑内容（图层 JSON + 图片 base64）自动存入 localStorage
  - 页面刷新后恢复编辑状态
  - 支持多个草稿槽位（按主题名区分）
  - 清除草稿功能

- [x] **T2.7 保存与导出 UI**
  - 保存时编辑主题名称
  - 保存时严格校验内容合法性，展示不合法提示
  - 检测图层数：一层（不含文字层）→ 卡片主题，多层 → 立绘主题
  - 导出按钮：调 `POST /api/editor/export` 下载 `<name>.zip`

### 阶段 3：集成与测试

- [x] **T3.1 图层 JSON schema 定义**
  - 英文变量名，字段与后端 `Theme`/`Layer` 结构对齐
  - `name` / `canvas` / `display` / `layers[]` / `layers[].category` / `layers[].zIndex` / `layers[].images[]`
  - 前后端数据契约对齐（AGENTS.md：改一边必查另一边）

- [x] **T3.2 端到端测试**
  - 创建卡片主题（单图层）→ 预览 → 导出
  - 创建立绘主题（多图层）→ 预览 → 导出
  - 导出包通过 `go run ./cmd/check-theme`
  - 主题数量变动不影响测试结果（AGENTS.md：容易变动的主题内容和数量不应该影响测试）

- [x] **T3.3 人工检验**
  - 电脑端：图层展开动画、拖拽排序、图片拖动定位
  - 手机端：图层面板底部抽屉、画布自适应、触摸拖拽
  - localStorage：刷新恢复编辑状态、多草稿切换、清除草稿
  - 导出包：解压到 `assets/theme/` 可被 Registry 加载

## 图层 JSON Schema

```json
{
  "name": "my-theme",
  "canvas": { "width": 2362, "height": 4134 },
  "display": { "size": 400, "crop": null },
  "layers": [
    {
      "id": 1,
      "category": "lass",
      "zIndex": 0,
      "fixed": false,
      "images": [
        { "src": "[image omitted]", "left": 116, "top": 348, "width": 2213, "height": 3646 }
      ]
    }
  ]
}
```

- `category`：`lass`/`brow`/`eye`/`mouth`/`face`，导出时映射到 `config.json` 的 `ranges`
- `images[]`：同一图层的多张候选图，运行时随机选一（对应 `RandomPickLayer`）
- 文字层不保存，仅预览时注入

## 导出标准

导出包解压到 `assets/theme/<name>/` 即可直接使用，通过 `cmd/check-theme` 全部校验。

### 卡片主题（单图层）

```
<name>/
  0.webp
  1.webp
  ...
  n-1.webp
  meta.json   (可选)
```

### 立绘主题（多图层）

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

- `ren.json`：图层清单（`name`/`left`/`top`/`width`/`height`/`visible`/`layer_id`/`group_layer_id`）
- `config.json`：画布尺寸 + 分类区间（`canvasW`/`canvasH`/`ranges`，1-based 闭区间）
- `display.json`：输出尺寸 + 裁剪（`size`/`crop.left`/`crop.top`/`crop.width`/`crop.height`）

## 通用规则

- 目录名：ASCII 字母/数字/连字符，不能为 `demo`/`random`，不与 builtin 冲突
- 图片格式：统一 WebP，服务端重编码（铁律 4）
- 单文件体积：≤ 4 MiB
- 帧图宽高 ≤ 2048px，立绘图层宽高 ≤ 4096px
- 上传/导出接口独立限流，不复用计数路径配额
- 编辑内容暂存 localStorage，不落库，不碰 `tb_count`（铁律 5）
