# 主题上传 + 审核投票 — 设计方案

> 基于 `AGENTS.md` 铁律 4（上传必须服务端重编码）和 M6 上传通道预留设计。
> 本文件为方案分析，不含实现代码。

## 1. 需求概述

用户可通过 Web 上传自定义主题。上传流程分三阶段：

1. **交检（Validation）**：上传时服务端校验大小、尺寸、格式、命名、配额
2. **审核（Review）**：通过交检的主题进入"待审核"状态，展示在审核队列
3. **投票（Voting）**：审核期间社区可对待审核主题投票（赞成/反对）
4. **采纳（Approve）**：管理员手动采纳或投票达标后自动采纳，主题进入正式 registry

## 2. 架构约束（来自 AGENTS.md）

| 约束 | 说明 |
|---|---|
| 服务端重编码 | 不信任客户端格式，解码后按 gif/png/webp 白名单重编码（铁律 4） |
| 独立限流 | 上传接口独立限流 `RATE_LIMIT_UPLOAD_PER_HOUR`（默认 10/h/IP），不复用计数配额 |
| 命名保留字 | 不能与 builtin 主题 / `demo` / `random` 冲突 |
| 尺寸上限 | 帧图 ≤ 2048px，立绘图层 ≤ 4096px |
| 体积上限 | 单文件 ≤ 4 MiB |
| 存储路径 | 当前只有 SQLite（`tb_count`），无对象存储（R2/S3 为 M6 预留） |
| 单二进制 | 主题图 + 前端 dist 经 `embed.FS` 打包进 Go 二进制 |

## 3. 数据模型

新增两张 SQLite 表（与 `tb_count` 同库 `data/count.db`）：

### `tb_theme_submission` — 主题提交记录

```sql
CREATE TABLE IF NOT EXISTS tb_theme_submission (
    id          INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    name        VARCHAR(32) NOT NULL,              -- 主题名（ASCII 字母/数字/连字符）
    submitter   VARCHAR(64) NOT NULL,              -- 提交者标识（IP hash）
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending / approved / rejected
    theme_type  VARCHAR(8)  NOT NULL,              -- card / character
    canvas_w    INTEGER     NOT NULL,
    canvas_h    INTEGER     NOT NULL,
    file_count  INTEGER     NOT NULL,              -- 图片数量
    upvotes     INTEGER     NOT NULL DEFAULT 0,    -- 赞成票数
    downvotes   INTEGER     NOT NULL DEFAULT 0,    -- 反对票数
    created_at  INTEGER     NOT NULL DEFAULT 0,    -- Unix timestamp
    reviewed_at INTEGER     NOT NULL DEFAULT 0,    -- 审核完成时间
    review_note VARCHAR(256) NOT NULL DEFAULT ''   -- 审核备注
);
```

### `tb_theme_vote` — 投票记录（防重复投票）

```sql
CREATE TABLE IF NOT EXISTS tb_theme_vote (
    id            INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    submission_id INTEGER     NOT NULL,
    voter         VARCHAR(64) NOT NULL,             -- 投票者标识（IP hash）
    vote          INTEGER     NOT NULL,             -- +1 赞成 / -1 反对
    created_at    INTEGER     NOT NULL DEFAULT 0,
    FOREIGN KEY (submission_id) REFERENCES tb_theme_submission(id)
);
```

- `voter` = SHA256(IP + User-Agent 截断)，用于防同一标识重复投票（不要求登录）
- `tb_theme_vote` 的 `(submission_id, voter)` 应加唯一索引防重复

### 迁移脚本

```bash
# 新增表不影响 tb_count，无需迁移已有数据
sqlite3 data/count.db < migrations/002_theme_submission.sql
```

任务结尾需提醒用户执行此迁移。

## 4. 文件存储策略

当前无对象存储（R2/S3 为 M6 预留）。待审核主题的图片暂存在本地磁盘：

```
data/submissions/
  <submission_id>/
    0.webp          # 卡片主题：0..n-1.webp
    ren.json        # 立绘主题：ren.json + config.json + ren/*.webp
    config.json
    ren/
      0.webp
      ...
```

- 上传时服务端解码 + 重编码为 WebP 后写入 `data/submissions/<id>/`
- 采纳后由管理员或 CI 将目录移动到 `assets/theme/<name>/` 并重建 registry
- **不打包进 embed.FS**：embed.FS 是编译期静态打包，运行时上传的主题无法进入。采纳的主题需要 runtime 动态加载（见第 7 节）

## 5. API 设计

### 5.1 上传主题

```
POST /api/themes/upload
Content-Type: multipart/form-data

Body:
  name: string           -- 主题名
  type: string           -- "card" | "character"
  canvas_w: int
  canvas_h: int
  files: []File          -- 图片文件（多文件上传）
  manifest: string       -- 立绘主题的 ren.json 内容（JSON string，可选）
  config: string         -- 立绘主题的 config.json 内容（JSON string，可选）
```

**交检规则（按顺序）：**

1. IP 配额检查（`RATE_LIMIT_UPLOAD_PER_HOUR`，默认 10/h）
2. 名称校验：ASCII 字母/数字/连字符、非保留字、不与 builtin 冲突
3. 文件数量：1~20 张
4. 每文件体积 ≤ 4 MiB
5. 每文件解码成功（gif/png/webp），服务端重编码为 WebP
6. 尺寸校验：帧图 ≤ 2048px，立绘图层 ≤ 4096px
7. 立绘主题：manifest + config JSON 合法性

**响应：**

```json
{
  "id": 42,
  "status": "pending",
  "message": "submission created, pending review"
}
```

**错误码：**
- 429：IP 配额超限
- 400：交检失败（名称/格式/尺寸/体积不合规）
- 413：总体积超限

### 5.2 待审核列表

```
GET /api/themes/submissions?status=pending&page=1&limit=20
```

返回待审核主题列表（含缩略图 URL、投票数）。

### 5.3 投票

```
POST /api/themes/submissions/:id/vote
Content-Type: application/json

Body:
  { "vote": 1 }    -- +1 赞成, -1 反对
```

- 同一 voter 对同一 submission 只能投一次
- 重复投票覆盖原投票（改投）
- 响应更新后的票数

### 5.4 审核操作（管理员）

```
POST /api/themes/submissions/:id/review
Content-Type: application/json

Body:
  { "action": "approve", "note": "looks good" }
  { "action": "reject", "note": "duplicate" }
```

- 需要管理员密钥（`ADMIN_KEY` 环境变量，请求头 `X-Admin-Key`）
- `approve`：将 `data/submissions/<id>/` 移动到 `assets/theme/<name>/`，状态改 `approved`
- `reject`：状态改 `rejected`，文件保留 7 天后清理

### 5.5 自动采纳阈值（可选）

当 `upvotes - downvotes >= AUTO_APPROVE_THRESHOLD`（默认 10）时自动采纳。
可通过环境变量 `AUTO_APPROVE_THRESHOLD` 配置，设为 0 则禁用自动采纳。

## 6. 限流方案

复用现有 `ratelimit` 包，新增上传专用限流器：

```go
// internal/server/server.go
uploadLimiter *ratelimit.IPLimiter  // RATE_LIMIT_UPLOAD_PER_HOUR
```

- 独立于 `ipLimiter`（计数路径）和 `nameLimiter`
- 窗口为 1 小时，阈值 `RATE_LIMIT_UPLOAD_PER_HOUR`
- 超限返回 429

## 7. Runtime 动态主题加载

当前 `embed.FS` 是编译期打包，运行时上传的主题无法进入 embed。采纳的主题需要 runtime 动态加载：

**方案 A（推荐，最小改动）：**
- `assets/theme/` 目录同时包含 embed 打包的 builtin 主题和运行时采纳的主题
- Theme Registry 启动时扫描 `assets/theme/` 目录（embed.FS + 本地磁盘叠加）
- 采纳操作将主题目录从 `data/submissions/` 移到 `assets/theme/`
- 需要给 Registry 增加"本地磁盘扫描"能力（当前只读 embed.FS）

**方案 B（M6 完整方案）：**
- 引入 R2/S3 对象存储，采纳的主题上传到 R2
- Registry 从 R2 拉取主题（需要 `aws-sdk-go-v2`，当前未引入）
- 适合多实例部署，但当前为单实例，投入产出比低

**建议**：当前用方案 A，M6 多实例时再评估方案 B。

## 8. 前端设计

### 8.1 上传页面 `/upload`（或编辑器内集成）

- 复用编辑器的导出 JSON 格式（`EditorRequest`）
- 上传前前端预校验：名称格式、文件数量、体积
- 上传后展示提交 ID 和审核状态

### 8.2 审核队列页面 `/submissions`

- 展示待审核主题列表（缩略图 + 名称 + 投票数）
- 每个主题可预览（复用 `/api/editor/preview` 渲染）
- 投票按钮（赞成/反对）
- 已投票的主题高亮显示当前选择

### 8.3 管理员审核面板（可选）

- `/admin/submissions`：管理员可见的审核操作面板
- 批量采纳/拒绝
- 需要 `X-Admin-Key` 认证

## 9. 安全考量

| 风险 | 措施 |
|---|---|
| 图片马 | 服务端解码 + 重编码（铁律 4），不信 Content-Type |
| 名称注入 | ASCII 白名单 + 保留字检查 |
| 体积攻击 | IP 配额 + 单文件 4MiB + BodyLimit |
| 投票刷票 | IP hash 防重复 + 投票接口独立限流 |
| 路径穿越 | 名称白名单（无 `/` `..` 等），submission_id 为整数 |
| 恶意 manifest | JSON 解析 + 字段校验 |

## 10. 实现优先级

分阶段实现，每阶段可独立交付：

| 阶段 | 内容 | 依赖 |
|---|---|---|
| **P1** | 上传端点 + 交检 + 本地存储 | 无 |
| **P2** | 待审核列表 + 预览 + 投票 | P1 |
| **P3** | 管理员审核 + 采纳 + Runtime 加载 | P1, P2 |
| **P4** | 自动采纳阈值 + 前端上传/投票页面 | P1-P3 |

## 11. 不实现的内容

- **用户注册/登录**：当前用 IP hash 标识，不引入用户系统
- **评论功能**：edit-todo.md 已剔除，不在范围
- **对象存储（R2/S3）**：M6 预留，当前用本地磁盘
- **多实例投票同步**：当前单实例，SQLite 即可
