# Lolicount TODO List

> 基于 `docs/projectDesign.md` 拆分,标注当前状态。状态:`[ ]` 待办 / `[~]` 进行中 / `[x]` 完成。

---

## M1:项目骨架

- [x] 初始化 go module + 目录结构 + `.gitignore` + `.env.example`
- [x] 实现 `internal/config`:环境变量加载与默认值
- [x] 实现 `internal/logger`:zerolog 封装
- [x] 实现 `cmd/server/main.go`:Fiber v3 启动 + `/heart-beat` 健康检查
- [x] 实现 `internal/assets/embed.go`:`embed.FS` 挂载 `assets/`
- [x] 本地验证:启动服务,`curl /heart-beat` 返回 alive

## M2:主题系统(内置)

- [ ] 定义 `theme.Theme` / `ThemeChar` 模型
- [ ] 实现 `theme.Registry` 接口 + `builtinRegistry`(扫描 `embed.FS` 的 `assets/theme/*`)
- [ ] 实现图片解码:读宽高 + 转 data URI(gif/png/webp)
- [ ] 实现 `theme.Render`:数字逐位查表拼 SVG(移植 `themify.js`)
- [ ] 支持参数:`theme/padding/offset/align/scale/fsize/pixelated/darkmode/prefix/num`
- [ ] 处理 `demo` 特例(返回固定 `0123456789`)+ `random` 主题
- [ ] 用 ImageGen 生成 `loli` 主题(`0~9` + `_start/_end`)
- [ ] 验证:`curl /@demo?theme=loli` 返回正确 SVG

## M3:存储与计数

> 单一存储路径:请求 → 内存 Buffer → 定时批量写 → SQLite(见 AGENTS.md 铁律 5)。

- [ ] 定义 `store.Repository` 接口(`Get/GetAll/Set/SetMulti`)+ `Counter` 结构体
- [ ] 实现 `store.sqliteRepo`:`modernc.org/sqlite`,`tb_count` 建表 + `SetMulti` 事务批量 upsert
- [ ] 实现 `counter.Buffer`:内存自增 + `time.Ticker` 按 `DB_INTERVAL` 批量落库
- [ ] 实现 `counter.flush`:快照 cache → `SetMulti` → 换新 map(修 Moe-Counter flush 期间丢增量)+ 失败回合并
- [ ] 缓冲上限守护:`len(cache) > 10000` 降级只读 + 日志告警
- [ ] 实现 `server/counter.go`:`GET /@:name` 自增 + 返回 SVG,`GET /get/@:name` 兼容
- [ ] 实现 `server/record.go`:`GET /record/@:name` 返回 JSON
- [ ] 验证:多次请求 `/@test` 计数递增;`/record/@test` 返回 JSON;崩溃后缓冲内增量丢失符合预期

## M4:限流与安全

- [ ] 实现 `ratelimit.ip`:IP 级令牌桶(`10/s, 300/min`),超限返 429
- [ ] 实现 `ratelimit.name`:name 级限流(`5/s`),超限降级只读不 +1(铁律 3)
- [ ] 接入 `go-playground/validator` 校验路由参数
- [ ] 实现 CORS 中间件(仅 `/api/*`)
- [ ] 设置 `Cache-Control: no-store`(非 demo),`demo` 长缓存(铁律 1)
- [ ] 实现 `server/params.go`:`QueryParams` 结构体 + validator 标签 + `applyDefaults()`
- [ ] 验证:压测超限返回 429 / 降级;参数非法返回 400

## M5:底图叠加(方案 C)

- [ ] 定义 `bg.Background` 模型 + `bg.Registry` 接口
- [ ] 实现 `bg.builtinRegistry`:加载 `assets/bg/*.json`(URL + 宽高 + 元数据)
- [ ] 实现 `theme.RenderWithBg`:底图 `<image href="url">` + 数字 `<image>`(data URI)叠加(铁律 2)
- [ ] 扩展 `GET /@:name` 支持 `bg/x/y/align/fsize/scale` 参数(不传走纯数字模式)
- [ ] 准备示例底图元数据 `assets/bg/loli-stand.json`(指向 CDN URL)
- [ ] 验证:`curl /@demo?bg=loli-stand&x=20&y=180&fsize=40` 返回带底图 SVG

## M6:Web 上传通道

- [ ] 实现 `bg.storage`:R2/S3 客户端(`aws-sdk-go-v2`),上传底图文件
- [ ] 实现 `bg.userRegistry`:Redis 存元数据 + 合并 builtin
- [ ] 实现 `server/bg_api.go`:`GET/POST/DELETE /api/backgrounds`
- [ ] 实现 `theme.userRegistry` + `server/theme_api.go`:`GET/POST/DELETE /api/themes`
- [ ] 上传校验:命名保留字、格式白名单、服务端重编码(铁律 4)、尺寸/体积上限、配额
- [ ] 上传接口独立限流(如 5 次/小时/IP)
- [ ] 验证:上传主题/底图 → 立即在 `?theme=` / `?bg=` 可用

## M7:前端(Vue → Nuxt 3 SSG)

- [ ] 初始化 `web/`:Nuxt 3 + pnpm + UnoCSS
- [ ] 实现 `composables/useApi.ts`:封装后端 API 调用
- [ ] 实现首页 `pages/index.vue`:主题市场网格 + 三种嵌入格式说明
- [ ] 实现示例调整页 `pages/playground.vue`:选主题/底图 → 拖拽调 x/y → 调 fsize/scale → 实时预览 → 生成链接
- [ ] 实现 `components/BgPreview.vue`:底图背景 + 可拖拽数字块浮层,实时算 x/y
- [ ] 实现 `components/ParamPanel.vue`:参数控件(theme/bg/fsize/scale/align/padding/offset/darkmode)
- [ ] 实现 `components/LinkOutput.vue`:同一 URL 派生 SVG address / Img tag / Markdown + 复制
- [ ] 实现 `composables/useDragPosition.ts`:pointer 事件 → 坐标换算 → 防抖回调
- [ ] 实现主题画廊 `pages/themes.vue` + 上传页 `pages/upload.vue`
- [ ] 接入 GSAP 动画:数字滚动、主题切换过渡、撒花
- [ ] `nuxi generate` 构建静态产物 → `embed` 进 Go 二进制或部署 CDN
- [ ] 验证:首页可访问,playground 拖拽/预览/复制/上传全功能

## M8:CI/CD 与部署

- [ ] 实现 `cmd/check-theme`:校验主题完整性(目录名/0~9 齐全/格式/尺寸)
- [ ] 实现 `scripts/validate-theme-meta.js`:`meta.json` schema 校验
- [ ] 实现 `scripts/gen-themes-json.js`:生成 `assets/themes.json`
- [ ] 编写 `.github/workflows/ci.yml`:go vet + test -race + Nuxt build
- [ ] 编写 `.github/workflows/theme-check.yml`:PR 改动 `assets/theme|bg/**` 触发校验
- [ ] 编写 `.github/workflows/release.yml`:tag `v*` 构建 Docker + Release
- [ ] 编写 `.github/workflows/rebuild-frontend.yml`:主题变更触发 SSG 重建
- [ ] 编写 `Dockerfile`:多阶段(builder 编 Nuxt+Go → alpine 运行)
- [ ] 编写 `docker-compose.yml`:app(单实例 SQLite,无需 redis)
- [ ] 完善 `README.md`:用法、API、主题/底图贡献指南

---

## 已完成的设计文档

- [x] `docs/detail.md`:技术细节(架构/接口/数据模型)
- [x] `docs/projectDesign.md`:项目设计(存储/接口/结构/计划)
- [x] `AGENTS.md`:AI agent 指南(铁律 + 工程原则)
- [x] `README.md`:项目介绍 + 三种嵌入格式 + 参数表
- [x] `docs/TODOlist.md`:本文件
