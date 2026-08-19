# 常见问题与故障排查

> 使用、部署、贡献过程中常见的问题。找不到答案可提 [issue](https://github.com/miaoledor/Lolicount/issues)。

## 使用

### 计数为什么不增长 / 卡在同一个数字?

最常见原因:计数器 SVG 被某层缓存加了 `max-age`。Lolicount 对非 `demo` 的
计数 SVG 一律返回 `Cache-Control: no-store`(铁律 1),但 GitHub 图片代理、
CDN 或浏览器扩展可能仍缓存。排查:

- 确认请求 URL 不是 `demo`(demo 固定返回 `0123456789`,不计数)。
- 直接 `curl -I https://lolicount.top/@your-name` 检查 `Cache-Control` 是否为 `no-store`。
- 若走了自有 CDN,确认其未对 `image/svg+xml` 强制加长缓存。

### 同一个 name 被刷量怎么办?

name 级限流(`5/s`)会在超限时**降级只读**:返回当前值但不 +1,让正常嵌入
不被一次性刷量打挂。这是设计行为,不是 bug。若需更严格控制,用不同 name
分流,或在反向代理层加额外限流。

### 图片不显示 / 显示 broken?

- 确认 `theme` 名称存在(参考 `assets/themes.json` 或 `/api/themes`)。
- 立绘主题名以 `-ren` 结尾(如 `lian-ren`),不要和卡片主题混淆。
- 嵌入平台需支持外部 SVG 图片。部分平台会过滤 `<img>`,改用 Markdown 形式。
- `theme=random` 时若 builtin 列表为空会报错,确认 `assets/theme/` 非空。

### 在鲲 Galgame 论坛等平台发布后图片不显示?

部分 markdown 编辑器(milkdown/remark)在源码模式会把 `&` 转义成 `\&`
以防 HTML 实体解析。后端 goldmark 渲染时保留字面反斜杠,产出非法 URL
如 `?theme=lian\&fsize=16\&...`,导致 lolicount 收到 `theme=lian\`
而返回 400。

v0.2.3 起 lolicount 在 counter 路由前加 `sanitizeBackslashEscape` 中间件,
自动把 query 里的 `\&` 还原成 `&`,兼容这类脏 URL。无需用户侧改动,
历史已发布内容刷新即恢复。

### `scale` 和 `fsize` 有什么区别?

`scale` 控制图片大小(基于统一 400px 最长边),`fsize` 控制计数文字字号,
两者完全独立。`scale` 省略时所有主题图按 400px 最长边等比缩放。

### `x/y` 和 `rx/ry` 怎么选?

- `x`/`y`:像素坐标,精确到具体位置,适合固定尺寸的底图。
- `rx`/`ry`:比例坐标(0~1),相对图片宽高,适合不同尺寸主题的相对定位。
- 二选一,不要同时传。都不传时文字默认在图片正下方居中。

### `number` 参数和真实计数有什么区别?

`number` 用于**预览**:指定数字直接展示,不落库、不 +1。真实计数走 `/@:name`
正常流程。`demo` 是保留 name,固定返回 `0123456789`,也不落库。

## 部署

### Docker 换域名要不要重新 build?

不需要。后端在 serve `index.html` 时会**运行时**把 `BASE_URL` 注入 SSG 的
`__NUXT__` payload。改 `BASE_URL` 环境变量后 `docker compose up -d` 重启即可
(build once, configure per env)。

### 端口被占用怎么办?

`pnpm dev:server` 会先跑 `node scripts/preflight-port.js` 清理 9721 占用。
生产环境改 `PORT` 环境变量,或用反向代理转发。

### 数据存在哪?怎么备份?

SQLite 文件在 `DB_PATH`(默认 `data/count.db`)。备份:停服或低峰期复制该
文件即可。Docker 部署用 `lolicount-data` 卷持久化。

### 升级后 schema 变了要迁移吗?

`tb_count` 表结构变更时,任务结尾会显式告知是否需要迁移。常规升级前建议备份
`data/count.db`。当前 schema 见 [架构文档](./architecture.md#tb_count-表)。

### 可以多实例水平扩展吗?

**当前不支持**。`counter.Buffer` 是进程私有内存,多实例会互相吞计数(铁律 5)。
严格单实例是有意识的权衡,水平扩展是未来需求。

## 开发

### `go test` 报 embed dist 为空?

CI 会先 `pnpm generate` 构建前端再跑测试。本地需先 `pnpm build:web` 再
`go test`,否则 `assets/dist` 只有 `.gitkeep`,前端相关测试会 404。

### 测试为什么用 `-count=1 -p 1`?

共享库集成测试涉及 SQLite 单写者,`-p 1` 串行避免并行写冲突,`-count=1`
禁用缓存确保每次实跑。测试用独立 DSN,不从 `.env` 取。

### 主题相关测试报错?

主题测试不应依赖具体主题内容或数量。用 fixture / mock 的 Registry,避免
增删主题导致测试 flaky。

### `check-theme` 报帧不连续?

帧索引必须从 `0` 开始连续递增(`0, 1, 2, ...`),不能跳号。支持
`gif` / `png` / `webp`。详见 [主题文档](./themes.md)。

## 贡献

### 可以修改 `AGENTS.md` / `projectDesign.md` / `TODOlist.md` 吗?

这三个文件标注**禁止修改描述内容**(只允许更新 TODOlist 的任务状态)。
如需同步描述,请在 issue 里提出,由维护者处理。

### commit message 有什么要求?

英文,Conventional Commits(`feat:` / `fix:` / `docs:` / `refactor:` / `test:`
等)。正文可附中文说明。一个 commit 只做一件事,保持功能单一。只在本地 commit,
不要 push。
