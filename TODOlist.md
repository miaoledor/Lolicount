按从上到下的实现顺序,给出完整 TodoList。每项都是可独立完成、可验证的最小单元。

## Lolicount 实现 TodoList

**M1:项目骨架**
1. 初始化 go module + 目录结构 + `.gitignore` + `.env.example`
2. 实现 `internal/config`:环境变量加载与默认值
3. 实现 `internal/logger`:zerolog 封装
4. 实现 `cmd/server/main.go`:Fiber v3 启动 + `/heart-beat` 健康检查
5. 实现 `internal/assets/embed.go`:`embed.FS` 挂载 `assets/`
6. 本地验证:启动服务,`curl /heart-beat` 返回 alive

**M2:主题系统(内置)**
7. 定义 `theme.Theme` / `ThemeChar` 模型
8. 实现 `theme.Registry` 接口 + `builtinRegistry`(扫描 `embed.FS` 的 `assets/theme/*`)
9. 实现图片解码:读宽高 + 转 data URI(gif/png/webp)
10. 实现 `theme.Render`:数字逐位查表拼 SVG(移植 `themify.js`)
11. 支持参数:`theme/padding/offset/align/scale/pixelated/darkmode/prefix/num`
12. 处理 `demo` 特例(返回固定 `0123456789`)+ `random` 主题
13. 用 ImageGen 生成 `loli` 主题(`0~9` + `_start/_end`)
14. 验证:`curl /@demo?theme=loli` 返回正确 SVG

**M3:存储与计数**
15. 定义 `store.Repository` 接口(`Get/Incr/SetMulti`)
16. 实现 `store.memoryRepo`:map + `sync.RWMutex`
17. 实现 `store.redisRepo`:`INCR` / `GET`
18. 实现 `store.sqliteRepo`:`modernc.org/sqlite`,批量 upsert
19. 实现 `counter.Buffer`:内存自增 + `time.Ticker` 批量落库(sqlite 模式)
20. 实现 `server.counter.go`:`GET /@:name` 自增 + 返回 SVG,`GET /get/@:name` 兼容
21. 实现 `server.record.go`:`GET /record/@:name` 返回 JSON
22. 验证:多次请求 `/@test`,计数递增;`/record/@test` 返回 JSON

**M4:限流与安全**
23. 实现 `ratelimit.ip`:IP 级令牌桶(`10/s, 300/min`)
24. 实现 `ratelimit.name`:name 级限流(`5/s`,超限降级只读不 +1)
25. 接入 `go-playground/validator` 校验路由参数
26. 实现 CORS 中间件
27. 设置 `Cache-Control: no-store`(非 demo),`demo` 长缓存
28. 验证:压测超限返回 429 / 降级;参数非法返回 400

**M5:底图叠加(方案 C)**
29. 定义 `bg.Background` 模型 + `bg.Registry` 接口
30. 实现 `bg.builtinRegistry`:加载 `assets/bg/*.json`(URL + 宽高 + 元数据)
31. 实现 `theme.RenderWithBg`:底图 `<image href="url">` + 数字图层叠加
32. 扩展 `GET /@:name` 支持 `bg/x/y/align` 参数(不传走纯数字模式)
33. 准备示例底图元数据 `assets/bg/loli-stand.json`(指向 CDN URL)
34. 验证:`curl /@demo?bg=loli-stand&x=20&y=180` 返回带底图 SVG

**M6:Web 上传通道**
35. 实现 `bg.Storage`:R2/S3 客户端,上传底图文件
36. 实现 `bg.userRegistry`:Redis 存元数据 + 合并 builtin
37. 实现 `server.bg_api.go`:`GET/POST/DELETE /api/backgrounds`
38. 实现 `theme.userRegistry` + `server.theme_api.go`:`GET/POST/DELETE /api/themes`
39. 上传校验:命名保留字、格式白名单、重编码、尺寸/体积上限、配额
40. 上传接口独立限流(如 5 次/小时/IP)
41. 验证:上传主题/底图 → 立即在 `?theme=` / `?bg=` 可用

**M7:前端(Nuxt 3 SSG)**
42. 初始化 `web/`:Nuxt 3 + pnpm + UnoCSS
43. 实现 `useApi.ts`:封装后端 API 调用
44. 实现首页 `pages/index.vue`:主题市场网格 + 参数表单 + 实时预览 + 复制 Markdown
45. 实现 `ThemeCard.vue` / `ParamForm.vue` / `CounterPreview.vue` 组件
46. 实现主题画廊 `pages/themes.vue`
47. 实现上传页 `pages/upload.vue`:主题/底图上传表单
48. 接入 GSAP 动画:数字滚动、主题切换过渡、撒花
49. `nuxi generate` 构建静态产物 → `embed` 进 Go 二进制或部署 CDN
50. 验证:首页可访问,主题预览/参数/复制/上传全功能

**M8:CI/CD 与部署**
51. 实现 `cmd/check-theme`:校验主题完整性(目录名/0~9 齐全/格式/尺寸)
52. 实现 `scripts/validate-theme-meta.js`:meta.json schema 校验
53. 实现 `scripts/gen-themes-json.js`:生成 `assets/themes.json`
54. 编写 `.github/workflows/ci.yml`:go vet + test -race + Nuxt build
55. 编写 `.github/workflows/theme-check.yml`:PR 改动 `assets/theme|bg/**` 触发校验
56. 编写 `.github/workflows/release.yml`:tag `v*` 构建 Docker + Release
57. 编写 `.github/workflows/rebuild-frontend.yml`:主题变更触发 SSG 重建
58. 编写 `Dockerfile`:多阶段(builder 编 Nuxt+Go → alpine 运行)
59. 编写 `docker-compose.yml`:app + redis
60. 完善 `README.md`:用法、API、主题/底图贡献指南

---

共 60 项,按顺序执行。每完成一项我会更新计划状态并继续下一项。要从第 1 项开始吗?