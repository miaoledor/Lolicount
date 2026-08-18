# Lolicount TODO List
agent铁律-不要修改该文件的任何描述内容,至允许修改当前任务的完成状态
完成任务后尝试填充todolist的任务状态
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

- [x] 定义 `theme.Theme` / `ThemeChar` 模型
- [x] 实现 `theme.Registry` 接口 + `builtinRegistry`(扫描 `embed.FS` 的 `assets/theme/*`)
- [x] 实现图片解码:读宽高 + 转 data URI(gif/png/webp)
- [x] 实现 `theme.Render`:数字逐位查表拼 SVG(移植 `themify.js`)
- [x] 支持参数:`theme/padding/offset/align/scale/fsize/pixelated/darkmode/prefix/num`
- [x] 处理 `demo` 特例(返回固定 `0123456789`)+ `random` 主题
- [~] 用 ImageGen 生成 `loli` 主题(`0~9` + `_start/_end`)
- [x] 验证:`curl /@demo?theme=loli` 返回正确 SVG

## M2.5 Refactor
修正路线
本项目只使用一张图片进行展示
比如default-theme[0.png 1.png 2.png 3.png ... size-1.png]
则每次显示(count+1)%size,count++
默认状态下count作为文字在图片正下方正中央进行展示
可接受参数number=选择数字进行展示，默认我为0
后续将增加修改位置选项

**状态:已完成(代码与验证)。AGENTS.md 的「Rendering/铁律2」仍描述 M2 字形图模型,因该文件标注禁止修改,待用户同步。**

- [x] 重写 theme 包:Theme=帧集合,Render=选帧图 + 计数文字叠加
- [x] 调整 server 参数:number 选帧(默认 0),移除 padding/prefix/fsize/scale/align/offset/pixelated/darkmode
- [x] 保留 assets/theme/loli 旧字形图作为 10 帧占位(帧图模型兼容)
- [x] 验证 curl /@demo?theme=loli 返回单帧图 + 文字 SVG
- [ ] AGENTS.md 渲染段/铁律2 同步为帧图模型(文件禁止修改,待用户处理)
- [ ] 后续 M5 底图叠加路线待按帧图模型重新评估

## M3:存储与计数

> 单一存储路径:请求 → 内存 Buffer → 定时批量写 → SQLite(见 AGENTS.md 铁律 5)。

- [x] 定义 `store.Repository` 接口(`Get/GetAll/Set/SetMulti`)+ `Counter` 结构体
- [x] 实现 `store.sqliteRepo`:`modernc.org/sqlite`,`tb_count` 建表 + `SetMulti` 事务批量 upsert
- [x] 实现 `counter.Buffer`:内存自增 + `time.Ticker` 按 `DB_INTERVAL` 批量落库
- [x] 实现 `counter.flush`:快照 cache → `SetMulti`(绝对值覆盖,不换 map,保留基线避免丢增量)+ 失败下次重试
- [x] 缓冲上限守护:`len(cache) > 10000` 降级只读 + 日志告警
- [x] 实现 `server/counter.go`:`GET /@:name` 自增 + 返回 SVG,`GET /get/@:name` 兼容
- [x] 实现 `server/record.go`:`GET /record/@:name` 返回 JSON
- [x] 验证:多次请求 `/@test` 计数递增;`/record/@test` 返回 JSON;崩溃后缓冲内增量丢失符合预期

> flush 设计说明:cache 持有绝对值而非增量,`SetMulti` 每次覆盖写整个快照,
> 因此 flush **不换新 map**——换 map 会丢失基线,导致下次 Incr 从 0 重新计数
> (Moe-Counter 的同类 bug)。flush 期间的新增 Incr 写入当前 map,下次 flush
> 一并落库,不丢失。进程崩溃时 `DB_INTERVAL` 窗口内的内存增量丢失,符合预期。

## M4:限流与安全

- [x] 实现 `ratelimit.ip`:IP 级令牌桶(`10/s, 300/min`),超限返 429
- [x] 实现 `ratelimit.name`:name 级限流(`5/s`),超限降级只读不 +1(铁律 3)
- [x] 接入 `go-playground/validator` 校验路由参数
- [x] 实现 CORS 中间件(仅 `/api/*`)
- [x] 设置 `Cache-Control: no-store`(非 demo),`demo` 长缓存(铁律 1)
- [x] 实现 `server/params.go`:`QueryParams` 结构体 + validator 标签 + `applyDefaults()`
- [x] 验证:压测超限返回 429 / 降级;参数非法返回 400

## M5:底图叠加(方案 C)

- [x] 定义 `bg.Background` 模型 + `bg.Registry` 接口
- [x] 实现 `bg.builtinRegistry`:加载 `assets/bg/*.json`(URL + 宽高 + 元数据)
- [x] 实现 `theme.RenderWithBg`:底图 `<image href="url">` + 数字 `<image>`(data URI)叠加(铁律 2)
- [x] 扩展 `GET /@:name` 支持 `bg/x/y/align/fsize/scale` 参数(不传走纯数字模式)
- [x] 准备示例底图元数据 `assets/bg/loli-stand.json`(指向 CDN URL)
- [x] 验证:`curl /@demo?bg=loli-stand&x=20&y=180&fsize=40` 返回带底图 SVG

## M5.5 项目方向修复

- [x] theme 作为底部图片的背景,不作为任何数字,仅仅表现风格,不作为任何计数的体现,层级在最下级(0)
- [x] count 作为字体,负责表现出当前的计数,层级在底图上面(1)

## M5.6 项目功能补充
- [x]在不传入图片放缩的时候，所有图片都以同样的放缩大小展示
    例如 1000 * 1000 和 2000 * 1800的图片展示出来的最终大小都是接近于一致的，采用放缩大小对图片进行限制大小，不要采用拉伸之类的方式
- [x]字体默认放置在图片的正下方
  [图片]
  [字体]
- [x]字体可以接受被隐藏 ?unshowf=ture
- [x]启动时自动扫描主题进行注册

## M6 给字体加上不同样式f-theme
- [x] 实现 f-theme 字体样式主题(?ftheme= 选择)
- [x] 内置样式:default/pink/neon/serif
- [x] 启动时自动扫描 assets/f-theme 注册
- [x] 在不添加x,y时默认在下方，添加分为像素x,y和比例x,y

## M7:前端(Vue → Nuxt 3 SSG)

- [x] 初始化 `web/`:Nuxt 3 + pnpm + UnoCSS
- [x] 实现 `composables/useApi.ts`:封装后端 API 调用
- [x] 实现首页 `pages/index.vue`:主题市场网格 + 三种嵌入格式说明
- [x] 实现示例调整页 `pages/playground.vue`:选主题/底图 → 拖拽调 x/y → 调 fsize/scale → 实时预览 → 生成链接
- [x] 实现 `components/BgPreview.vue`:底图背景 + 可拖拽数字块浮层,实时算 x/y
- [x] 实现 `components/ParamPanel.vue`:参数控件(theme/bg/fsize/scale/align/padding/offset/darkmode)
- [x] 实现 `components/LinkOutput.vue`:同一 URL 派生 SVG address / Img tag / Markdown + 复制
- [x] 实现 `composables/useDragPosition.ts`:pointer 事件 → 坐标换算 → 防抖回调
- [x] 实现主题画廊 `pages/themes.vue` + 上传页 `pages/upload.vue`
- [x] 接入 GSAP 动画:数字滚动、主题切换过渡、撒花
- [x] `nuxi generate` 构建静态产物 → `embed` 进 Go 二进制或部署 CDN
- [x] 验证:首页可访问,playground 拖拽/预览/复制

## M7.5:前端细节增强 
- [x] 所有的内容都需要在同一个页面
- [x] 在右下角添加一个跳转到最上的功能


## M8:CI/CD 与部署

- [x] 实现 `cmd/check-theme`:校验主题完整性(目录名/0~9 齐全/格式/尺寸)
- [x] 实现 `scripts/validate-theme-meta.js`:`meta.json` schema 校验
- [x] 实现 `scripts/gen-themes-json.js`:生成 `assets/themes.json`
- [x] 编写 `.github/workflows/ci.yml`:go vet + test -race + Nuxt build
- [x] 编写 `.github/workflows/theme-check.yml`:PR 改动 `assets/theme|bg/**` 触发校验
- [x] 编写 `.github/workflows/release.yml`:tag `v*` 构建 Docker + Release
- [x] 编写 `.github/workflows/rebuild-frontend.yml`:主题变更触发 SSG 重建
- [x] 编写 `Dockerfile`:多阶段(builder 编 Nuxt+Go → alpine 运行)
- [x] 编写 `docker-compose.yml`:app(单实例 SQLite,无需 redis)
- [x] 完善 `README.md`:用法、API、主题/底图贡献指南

## M9:功能增强
- [x] 实现随机图片展示
- [x] 参考kungal-forum中设置页面莲的样式，实现gal图片立绘渲染
- [x] 将主题分为两种类型
  一种为原本的主题 支持顺序模式(png0 png1 png2...)和随机模式(png1 png3 png0 ...)
  一种为角色立绘主题 例如莲的角色立绘 只支持随机模式，每次获取该图片都会随机抽取服装和表情等
  两种主题都是作为底，展示层级在0,两种都要支持三种嵌入方式
- [x] 修改原本的卡片主题选择，点击该主题不再作为playground的主题选择，而是重新加载该图片
- [x] 增强前端Playground，根据选择的主题类型（立绘/卡片），渲染出可以选择的参数，其中主题采用下拉栏选主题名
  在选择玩参数后，点击Generate it!在下方展示出渲染出的图片
  并同时在下方展示三种嵌入方式的链接

## M9.5:修复功能
- [x] 卡片主题展示时 点击无法刷新图片 且只有两种卡片主题 但是却有三张卡片
- [x] 在目录结构上，将卡片主题和立绘主题分开，并修复相关影响代码
- [x] 区分主题 和 主题类型
  主题类型为（卡片主题/立绘主题）
  主题先选择完主题类型
  然后再进行选择该主题类型下的主题
  比如选择了卡片主题后，可以选lian kuon
  选择立绘主题后，可以选lian-ren

## M9.6:修复功能
- [x] 卡片主题展示下为空
- [x] Playground无法选择立绘主题，按钮点击后无反应
- [x] 去掉拖动在预览图上拖拽设置像素 x/y的功能
- [x] 多个输入框之间没有间隔
- [x] generate it可以多次点击生成图片

## M10:近一步完善项目结构
-[x] web页面支持中英双文
-[x] 根据现有的todolist projectdesign完善页面内容
-[x] readme支持中，日，英三语
-[x] 在web页面添加用户miaoledor和仓库页面的跳转
-[x] 在页面最下方添加感谢名单，感谢所有的贡献者
-[x] 在web和仓库readme添加捐献通道
-[x] 添加贡献规范，贡献分为两种，一种是主题贡献，一种功能贡献，给两种贡献分别添加文档，两个文档间可以跳转
-[x] 技术内容尽量放在贡献文档，readme偏向于介绍与使用
-[x] readme添加致谢项目：
  https://github.com/KunMoe/kun-galgame-forum
  https://github.com/journey-ad/Moe-Counter
-[x] 为项目其他细节，项目结构，技术选型之类的都添加项目的文档
  添加使用与部署文档
-[x] 使用与部署都要支持win mac linux,例如pnpm dev时的脚本要三端运行没问题
  




---

## 已完成的设计文档

- [x] `docs/detail.md`:技术细节(架构/接口/数据模型)
- [x] `docs/projectDesign.md`:项目设计(存储/接口/结构/计划)
- [x] `AGENTS.md`:AI agent 指南(铁律 + 工程原则)
- [x] `README.md`:项目介绍 + 三种嵌入格式 + 参数表
- [x] `docs/TODOlist.md`:本文件
