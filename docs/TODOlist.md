# Lolicount TODO List
默认不要修改该文件的任务描述,只更新完成状态;用户或维护者明确授权文档同步时例外。
完成任务后尝试填充todolist的任务状态
> 基于 `docs/projectDesign.md` 拆分,标注当前状态。状态:`[ ]` 待办 / `[~]` 进行中 / `[x]` 完成。
> M1～M15 保留历史演进记录,其中旧参数、页面或目录不代表当前公开契约;当前行为以
> `AGENTS.md`、`docs/projectDesign.md` 与较新的里程碑为准。

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

**状态:已完成(代码、验证与文档已同步)。**

- [x] 重写 theme 包:Theme=帧集合,Render=选帧图 + 计数文字叠加
- [x] 调整早期 server 参数为帧图模型;后续里程碑重新引入 `fsize`、`scale` 与定位参数
- [x] 保留 assets/theme/loli 旧字形图作为 10 帧占位(帧图模型兼容)
- [x] 验证 curl /@demo?theme=loli 返回单帧图 + 文字 SVG
- [x] AGENTS.md 渲染段同步为统一图层栈与帧图/立绘模型
- [x] M5 底图叠加路线已按统一图层栈重新实现

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

- [x] 实现 `ratelimit.ip`:IP 级令牌桶(程序默认 `60/s, 3000/min`),超限返 429
- [x] 实现 `ratelimit.name`:name 级限流(程序默认 `20/s`),超限降级只读不 +1(铁律 3)
- [x] 接入 `go-playground/validator` 校验路由参数
- [x] 实现 CORS 中间件(仅 `/api/*`)
- [x] 设置真实 SVG 的完整 no-cache 组合头;确定性 `demo` 长缓存、随机 `demo` 不缓存(铁律 1)
- [x] AGENTS.md 铁律1 与缓存表同步为 no-cache 组合头(`no-store, no-cache, max-age=0, must-revalidate`,GitHub camo/Fastly 边缘无视裸 no-store;经用户一次性授权修改)
- [x] `docs/projectDesign.md` 缓存契约同步为完整 no-cache 组合头
- [x] `docs/projectDesign.md` 补充 `GET /api/themes/hot` 与 `tb_theme_usage` 契约
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

## M7:前端(Vue → Nuxt 4 SSG)

- [x] 初始化 `web`:Nuxt 4 + pnpm + UnoCSS
- [x] 实现 `composables/useApi.ts`:封装后端 API 调用
- [x] 实现首页 `pages/index.vue`:主题市场网格 + 三种嵌入格式说明
- [x] 实现示例调整页 `pages/playground.vue`:选主题/底图 → 拖拽调 x/y → 调 fsize/scale → 实时预览 → 生成链接
- [x] 实现 `components/BgPreview.vue`:底图背景 + 可拖拽数字块浮层,实时算 x/y
- [x] 实现 `components/ParamPanel.vue`:参数控件(theme/bg/fsize/scale/align/padding/offset/darkmode)
- [x] 实现 `components/LinkOutput.vue`:同一 URL 派生 SVG address / Img tag / Markdown + 复制
- [x] 实现 `composables/useDragPosition.ts`:pointer 事件 → 坐标换算 → 防抖回调
- [x] 实现早期主题画廊并在后续收敛到首页
- [x] 接入 GSAP 动画:数字滚动、主题切换过渡、撒花
- [x] `nuxi generate` 构建静态产物 → `embed` 进 Go 二进制或部署 CDN
- [x] 验证:首页可访问,playground 拖拽/预览/复制

## M7.5:前端细节增强 
- [x] 所有的内容都需要在同一个页面
- [x] 在右下角添加一个跳转到最上的功能
- [x] 主题画廊选中主题同步到 URL 查询参数(`/themes?theme=名称`),带参进入/刷新可恢复选中,面板切换与软导航(含参数移除回退默认)均保持一致


## M8:CI/CD 与部署

- [x] 实现 `cmd/check-theme`:校验主题完整性(目录名/0~9 齐全/格式/尺寸)
- [x] 实现 `scripts/validate-theme-meta.js`:`meta.json` schema 校验
- [x] 实现 `scripts/gen-themes-json.js`:生成 `assets/themes.json`
- [x] 编写 `.github/workflows/ci.yml`:go vet + test -race + Nuxt build
- [x] 更新 CI/CD：固定 pnpm 与 Ubuntu 运行器、升级相关 Actions；前端测试与 SSG 产物在 Go 任务中复用，Docker 统一测试与发布构建；补充完整本地验证与 GitHub 账单锁定排查说明。远端 CI 须待账号解除锁定后验证。
- [x] 编写 `.github/workflows/theme-check.yml`:PR 改动 `assets/theme|bg/**` 触发校验
- [x] 编写 `.github/workflows/release.yml`:tag `v*` 构建 Docker + Release
- [x] 编写 `.github/workflows/rebuild-frontend.yml`:主题变更触发 SSG 重建
- [x] 编写 `Dockerfile`:多阶段(builder 编 Nuxt+Go → alpine 运行)
- [x] 编写 `docker-compose.yml`:app(单实例 SQLite,无需 redis)
- [x] 完善 `README.md`:用法、API、主题/底图贡献指南

## M9:功能增强
- [x] 实现随机图片展示
- [x] 参考kungal-forum中设置页面莲的样式，实现gal图片立绘渲染
- [x] 真实计数响应缓存头改为 `no-store, no-cache, max-age=0, must-revalidate`:GitHub camo/Fastly 边缘无视裸 `no-store`(MISS→HIT 实测),导致 README 嵌入计数冻结
- [x] 将主题分为两种类型
  一种为帧图主题(png0 png1 png2...)，每次请求随机抽帧
  一种为角色立绘主题 例如莲的角色立绘，每次获取该图片都会随机抽取服装和表情等
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
  
## M11后续功能规划安排
-[x] 启动时进行扫描，如果发现卡片主题的文件序号不为0,1,2...size-1 则进行自动修改
  实现 `cmd/fix-theme`:扫描磁盘 `assets/theme/`,帧序号不连续时自动重命名为 0..n-1;支持 `--dry-run`;立绘主题自动跳过

-[x] 取消启动时到秒修改，改为使用pnpm运行脚本单独进行扫描修改
  如果发现卡片主题的文件序号不为0,1,2...size-1 则进行自动修改
  实现 `cmd/fix-theme`:扫描磁盘 `assets/theme/`,帧序号不连续时自动重命名为 0..n-1;支持 `--dry-run`;立绘主题自动跳过
  新增 `pnpm fix-theme` / `pnpm fix-theme:dry` 脚本入口(三端兼容),并在主题贡献文档校验流程中接入

## M12 后续大体规划任务
图片上面可以添加文字，文字可以选择使用普通文字，计数器，参考sekai-stickers可以动态调整文字


丰富计数样式：
  进度条 SVG:?style=progress&goal=1000 渲染 █████░░░ 623/1000。适合目标进度展示(粉丝数、下载量目标)。
  shields.io 风格 badge**:返回标准 SVG badge https://img.shields.io/badge/visits-12345-blue,尺寸小、社区生态熟悉,可直接贴 README
  emoji 计数:?style=emoji 用 🔟 表情或 🎉 堆叠表示量级(每 100 一个),趣味场景。
图标
更多主题
支持外部图片源
支持外部插件-通过restful请求调用外部插件-兼容任何语言
支持仪表盘，查看访问排名和总访问次数
- [x] 服务器捐赠内容添加：新增 `/support`，提供 GitHub Sponsors 与 Star / 主题 / 代码贡献入口；原创 Three.js 爱心罐支持投放、堆叠、旋转、多语言、深浅色、静态降级与减少动态效果，详见 [Support 页面](./support-page.md)。
图片大小优化脚本
图片上面支持添加文字


---

## M13:Emote 挂件(PSB 动图计数器,方案 A POC)

> 设计文档:`docs/emote-widget.md`。E-mote (PSB) 角色 WebGL 实时渲染,每次打开随机动作,页面加载自动计数,第三方网页经 `<script>` 挂件引用。

- [x] 调研 PSB/E-mote 格式与浏览器播放方案(FreeMote-SDK WebGL 驱动)
- [x] `GET /api/count/@:name`:计数 JSON 接口(自增/demo/number 特例/CORS/no-cache 组合头)+ 单测
- [x] `assets/psb/` 模型目录 + `GET /api/psb/models`、`GET /psb/:model`(immutable 长缓存/名称白名单)+ 单测
- [x] vendor WebGL 驱动到 `web/public/widget/`(附 CC BY-NC-SA 第三方声明)
- [x] `widget.js` 挂件:随机动作播放 + 画布像素静止检测自动换动作 + `{n}` 计数模板渲染
- [x] Playground 集成:`/api/themes` 追加 `animated` 标记,主题下拉显示动图模型,生成后
  WebGL 预览 + 挂件 HTML 嵌入代码(无独立页面,复用 `EmotePreview` 组件)
- [x] 修复 frontend.go 预渲染子页路由(目录 → index.html,当时的 `/editor`、`/about`、`/emote` 不再回落首页)+ 单测;Editor 后续已拆出主仓
- [x] 端到端验证:核心单二进制 + `PSB_DIR` 外置模型 + 跨域第三方页面挂件
- [ ] psb_converter(Python)归一化逻辑移植为 Go 包,支持直接投放原始 .psb
- [ ] 挂件交互增强(点击反应、视线跟随)

---

## M14:Spine 动态立绘(骨骼动画计数器)

> 设计文档:`docs/spine-widget.md`。Spine 3.8 WebGL 实时渲染,随机动作,页面加载自动
> 计数;**额外提供免 JS 路径**——构建期预渲染 looping 动画 WebP,第三方裸 `<img>` 嵌入
> 即可播放(无脚本,适配 GitHub README)。

- [x] `assets/spine/` 模型目录 + `GET /api/spine/models`、`GET /spine/models/:name/:file`、
  `GET /spine/anim/:name/:file`(immutable 长缓存/名称+文件名白名单)
- [x] vendor Spine 3.8 WebGL 运行时至 `web/public/spine/spine-3.8.js`
- [x] `spine-player.html` 交互页:AssetManager 载 atlas+skel/json,随机动画,
  `getBounds` 自适应取景,`{n}` 计数 overlay
- [x] `scripts/render-spine-anim.mjs` 免 JS WebP 渲染器(headless Chrome + CDP + ffmpeg)
- [x] 主题列表集成:`/api/themes` 追加 `spineModelNames` 带 `animated:true, kind:"spine"`;
  `ThemeKind`/`ThemeInfo` 增加 `spine`
- [ ] 交互增强(动作切换、点击反应)
- [ ] 4.x 运行时兼容(过新模型需换 4.x runtime)

---

## M15:Live2D 动态立绘(Cubism 计数器)

> 设计文档:`docs/live2d-widget.md`。Cubism 3 / Cubism 2(BanG Dream)模型经
> untitled-pixi-live2d-engine (PixiJS v8) 实时渲染,**点击切换动作**(无鼠标眼动跟踪),
> 页面加载自动计数。
> **仅交互(`<iframe>`)路径**:Live2D 需 WebGL + JS,无免 JS 的 `<img>` 路径
> (预渲染动画图太大、实时渲染太慢,两害相权放弃,见设计文档第 1 节)。

- [x] `assets/live2d/` 模型目录 + `GET /api/live2d/models`、`GET /live2d/models/:name/:file`
  (immutable 长缓存/名称+文件名白名单)
- [x] vendor SDK 四件套到 `web/public/live2d/`:pixi.min.js、live2dcubismcore.min.js、
  live2d-legacy.min.js、live2d-engine.js(依序加载为硬约束)
- [x] `live2d-player.html` 交互页:`Live2DModel.from(<清单>)`(按清单自动选 Cubism 3 / 2
  运行时),**点击画布按序切换 motion**(仿 emote 播放器),`{n}` 计数 overlay
- [x] `gen-live2d-model3.mjs` 为裸 Cubism 3 模型目录生成 `model3.json` 清单
- [x] 主题列表集成:`/api/themes` 追加 `live2dModelNames` 带 `animated:true, kind:"live2d"`;
  `ThemeKind`/`ThemeInfo` 增加 `live2d`
- [x] 修复 `extOf` 双点 bug(Spine/Live2D Content-Type 映射 miss→octet-stream)
- [x] 修复交互页 base 路径(按尾段匹配,避免 `/live2d-player.html/` 坏路径)
- [x] 修复 Pixi v8 不绑定传入 canvas 的 bug(改用 `app.canvas` 替换 `#c`)+ 两遍式
  不透明像素框 **contain-fit 整框取景**(全身/半身都完整显示,不裁剪不上移)
- [x] **去除眼动/鼠标跟踪**:关闭引擎 `Automator` 的 `autoFocus`/`autoHitTest`,
  角色眼睛与头部不再跟随鼠标
- [x] **点击切换动作**:收集 `model.internalModel.motionManager.definitions` 的
  motion 分组(跳过 `-` 分隔行与「初期化 / 視線追従」),点击画布按序
  `startMotion(group, index)` 换下一个(公开 `Live2DModel` 无 `motionManager`,
  MotionManager API 是 `startMotion` 不是 `playMotion`);裸模型(无 `.mtn`)
  时点击无效果、保持静态 idle
- [x] **修复清单探测 HEAD→GET**:后端 `/live2d/models/` 路由只注册 GET,HEAD 落到
  SPA catch-all 误报导致模型不加载(白屏);改 GET + 校验 JSON manifest
- [x] **BanG Dream 角色导入**:BD 提取物是**标准 Cubism 2 moc**,`live2d-legacy.min.js`
  官方 Cubism 2 运行时可直接渲染(引擎按 `model.json` 清单自动选 legacy 运行时)

---

## M16:内存策略、界面打磨与主题修正

- [x] 新增公主连结剧情立绘静态主题：保留 1657 套有可见差异且变化数大于 1 的服装并整理为 1406 个主题；有名角色按服装独立，无名 NPC 归并为 205 个互斥套装主题；删除未登场剪影以及其余全部变化数为 1 的主题。支持身体/眼睛/嘴巴组合、来源快照、离线导出、多语言搜索与缩略图。详见 [主题说明](./priconne-themes.md)。

- [x] 关于页重构:按设计稿改为「求 star」卡片 + 项目详情网格;移除页脚与轮播,
  首页 How to use 下加入同款求 star 卡片,首页主题展示改宝丽来样式
- [x] 移除顶部导航「主题展示」入口(画廊由 Playground 承担)
- [x] How / About 页精简为项目简介、醒目的主题总数与单主题最大变化数、三步使用说明及 Markdown 示例；统计直接读取主题接口，保留中英日文与移动端适配，移除个人介绍、技术栈、社交及求 Star 模块。
- [x] 主题选择写入路由(`?theme=`),分享链接直达所选主题
- [x] 主题预览下新增「下载当前帧」:ACAO 头 + `crossorigin` 打通跨域 canvas
  导出,同步 `toDataURL` 规避 Safari 手势窗口限制
- [x] Playground 主题按来源分组折叠(默认折叠,搜索/类型筛选时自动展开),
  移除游戏下拉筛选;新增基于真实使用量的热门主题组(固定 10 个,启动 +
  每小时重算:`tb_theme_usage` 表 + `GET /api/themes/hot`,第二个
  `counter.Buffer` 复用铁律 5 批处理语义)
- [x] **图片内存策略:每槽位常驻一张的懒加载**。启动只构建元数据目录(路径/
  摆放/文件头尺寸),请求时 `PrepareRender` 换装并返回深拷贝快照;常驻匿名堆
  ≈0.13 GiB(原全量预加载 ≈0.94 GiB),PSB 保持磁盘流式
- [x] **多帧主题画布跨换帧稳定**:`RandomPickLayer` 携带全候选帧尺寸元数据
  (`FrameDims`),画布取所有帧按显示缩放后的最大宽高,随机换帧不再跳变
- [x] **修正星光咖啡馆七个主题的角色标注**:目录名来自贡献者解包的 xp3 内部
  目录,与实际立绘不符;按图层名/表情瞳色/官方立绘重正:nova=四季ナツメ,
  yuna/riona/shuna=明月栞那,ai/rokuro=汐山涼音,ketsie=火打谷愛衣。
  目录名保持不变,已嵌入计数 URL 不受影响
- [x] 修复主题搜索框右侧越界:未加载全局 CSS reset,`<input>` 保持 UA
  content-box,`w-full + px + border` 溢出面板 26px;搜索框加 `box-border`
- [x] 移除「隐藏字体 (unshowf)」面板开关:行为冻结为原默认(生成链接仍带
  `unshowf=true`),服务端 `?unshowf=` 参数保留可用
- [x] 生成按钮烟花动画:重写 StarBurst 为 fixed 全视口烟花(主爆点 + 两个
  错峰卫星爆点、白闪、重力弧线粒子);节点 setTimeout 兜底清理、排程改定时器,
  不再依赖 rAF 心跳(旧版被按钮行容器裁剪且冻结标签页会泄漏节点)
- [x] 画廊静态缩略图管线 `scripts/gen-theme-thumbs-webp.mjs`(修复限流图床
  卡片图加载失败)与图片无损压缩脚本 `scripts/optimize-images.mjs`

---

## M17:Unity UGUI 动态立绘(舞萌DX)

> 设计文档:`docs/unity-widget.md`。构建期把 Unity AssetBundle 中的 UGUI prefab、
> Sprite 与 AnimationClip 转换成浏览器模型,运行时用轻量 DOM/CSS 播放器渲染。

- [x] `scripts/extract-unity-ui.py`:解析跨包 PPtr、层级与 Streamed/Dense/Constant 曲线
- [x] 导入 23 个 `maimai-navichara-XX` 模型及图库缩略图(排除缺失的 07 与联动静态角色 17、23)
- [x] `GET /api/unity/models`、`GET /unity/models/:name/:file` 白名单资源接口
- [x] `unity-player.html`:随机动作、点击换动作、contain-fit、`{n}` 计数 overlay
- [x] 主题列表/筛选/搜索集成 `kind:"unity"` 与舞萌DX游戏分类
- [x] 修复 Spine 动态主题误用 E-mote 播放器及非 PSB 嵌入代码错误
- [x] `prefers-reduced-motion` 首帧静态降级

---

## 已完成的设计文档

- [x] `docs/architecture.md`:架构文档(总体架构/存储/渲染/限流/缓存)
- [x] `docs/projectDesign.md` 中的接口契约与 `internal/server/params.go` 对齐
- [x] `docs/projectDesign.md`:项目设计(存储/接口/结构/计划)
- [x] `AGENTS.md`:AI agent 指南(铁律 + 工程原则)
- [x] `README.md`:项目介绍 + 三种嵌入格式 + 参数表
- [x] `docs/TODOlist.md`:本文件

---

## M17:Editor、提取 skill 拆分与主仓收敛

- [x] Editor 前后端迁移到独立仓库,主仓移除 `/editor` 页面、Editor API 与专用实现
- [x] 立绘提取 skill 迁移到独立仓库,主仓文档改为外部工具入口
- [x] 增加拆分路由回归测试,确保已迁移入口不会被主仓静态路由误接管
- [x] 删除 `imgcore` 中仅供 Editor 使用的编码死代码,保留运行时合成职责
- [x] 修复 `ren.json` 图层尺寸契约,覆盖目录加载、首次装配与后续换装回归测试
- [x] 强化 `cmd/check-theme` 的分层主题校验并清理不合规资产
- [x] 审查并同步项目设计、架构、部署、主题、贡献、PSB 与 TODO 文档;
  `AGENTS.md` 按保护规则保持与主分支一致
- [x] 取消主题上传与在线编辑路线,主题贡献统一收敛为仓库 Pull Request
- [x] 修复动态预览因 Fiber 默认 4 KiB 请求头缓冲触发的 431;提升为有界 32 KiB,
  并以真实 TCP + 8 KiB Cookie 回归测试覆盖
- [x] 固化 PSB 按需生命周期:启动不预读,请求时磁盘流式读取,响应结束关闭文件流;
  浏览器省略 Cookie,预览与下载均使用 private immutable 本地缓存,共享代理不保存
- [x] 合并前清理上传限流/Admin 占位设施、旧立绘全量加载器与渲染时候选分支,
  并完成 Go 全量测试、vet、主题校验和前端 SSG 构建

---

## M18:运行内存优化

> 量化数据与人工检验见 `docs/memory-optimization.md`。

- [x] **图片尺寸只读文件头**:`asset/header.go` 只读取前 32 字节解析
  PNG/GIF/WebP 头,未识别容器回退 `image.DecodeConfig`;43469 张图逐张
  与全文件解码器比对尺寸全部一致。catalog 扫描分配 1.52 GB/op →
  33.2 MB/op
- [x] **只常驻热门主题**:`NewThemeRegistryWithResident` + `Warm()` 由
  `tb_theme_usage` Top-N 维护常驻集,冷主题按请求现装现放;
  `THEME_RESIDENT_LIMIT` 默认 10。常驻堆 308.68 MB → 7.10 MB,
  启动累计分配 2.73 GB/op → 39.4 MB/op
- [x] **冷路径配套**:`compose()` 去掉被覆盖的 `Get()`,
  `ThemeInfo.IsMultiFrame` 用元数据判定缓存头,冷装配保持名字播种的
  确定性(画廊缩略图不变)
- [x] **启动参数可配常驻数量**:`-theme-resident-limit` 默认 10,优先级
  命令行 > `THEME_RESIDENT_LIMIT` > 内置默认;`0` = 全量常驻,非法值
  启动即失败,`-h` 打印用法。启动日志带 `theme_resident_limit`,
  `cmd/server/main_test.go` 覆盖优先级与校验
- [x] 新增内存基准 harness、头部全量一致性测试、常驻池单测与配置单测;
  `go test ./... -count=1` 全绿
- [x] 同步 `docs/memory-optimization.md`、`docs/architecture.md`、
  `docs/deployment.md` 与 `.env.example`
- [ ] 后续候选:请求链路零拷贝、套装主题增量换装、大模型 handler 流式、
  构建期 catalog 清单、冷资产外置、热度时间衰减

---

## M19:预览卡片刷新冷却与 M3 加载器

> 规范参考:[M3 Loading indicator](https://m3.material.io/components/loading-indicator/overview)
> (M3 Expressive,适用于 200ms–5s 的短等待)。

- [x] 点击预览卡片刷新进入 **2s 冷却**,窗口内连点不再触发新的 SVG 请求;
  门控抽为无框架依赖的 `web/app/utils/reloadCooldown.ts`(可注入时钟),
  `scripts/reload-cooldown.test.mjs` 覆盖窗口边界、自定义时长与重置
- [x] 卡片右下角新增 `web/app/components/M3LoadingIndicator.vue`:contained
  容器(36px、38/48 比例)+ 七形状 morph + 旋转,`prefers-reduced-motion`
  降级为静态圆点,`role="status"` 复用三语 `previewLoading` 文案;
  冷却期间显示,组件本身不感知计时
- [x] 加载器中心显示**整秒倒数**(`2 → 1`),`remainingWholeSeconds()` 向上
  取整并有单元测试;冷却中重复点击弹出三语「2 秒冷却」提示(1.5s,
  连点重置时长);不再改变鼠标指针,改为禁止连点选中预览文字
- [x] `pnpm test:frontend` 纳入新测试(27/27 通过);`pnpm build:web` 完整
  SSG 生成成功(12 个路由预渲染),组件自动导入已在 `components.d.ts` 中确认
- [ ] 人工检验:连点预览卡片时右下角加载器显示约 2s、中心倒数 2→1、
  提示「2 秒冷却」出现且背景不重复换帧;系统开启「减少动态效果」时
  应为静态圆点

---

## M20:页脚版本号与 CD 自动注入

- [x] 页脚改为 `© <year> Lolicount <tag>`(当前 `© 2026 Lolicount v0.4.4`),
  年份保持动态,版本号由构建注入而非写死
- [x] 新增 `scripts/app-version.mjs`:优先级 `NUXT_PUBLIC_APP_VERSION` >
  `APP_VERSION` > `GITHUB_REF_NAME`(tag 格式) > `git describe --tags`;
  `web/nuxt.config.ts` 暴露 `runtimeConfig.public.appVersion`,
  `HomeFooter.vue` 读取,空值时不显示版本
- [x] CD 注入:`release.yml` / `release-assets.yml` 传 `APP_VERSION` build arg,
  `Dockerfile` 加 `ARG/ENV NUXT_PUBLIC_APP_VERSION` 并复制整个 `scripts/`
  (同时修复 `test:frontend` 新增用例未进镜像的问题)
- [x] `scripts/app-version.test.mjs` 覆盖四级优先级与无版本回退,
  `pnpm test:frontend` 32/32 通过;`docs/ci-cd.md` 记录构建入口对照表
- [ ] 人工检验:本地 `pnpm build:web` 后首页页脚显示
  `© 2026 Lolicount v0.4.4`;打新 tag 后由 release workflow 构建的镜像
  页脚应显示新版本号

---

## M21:预览刷新跳动与烟花

- [x] 点击预览卡片刷新时,预览内容用 GSAP 做轻微 squash-stretch 跳动
  (上移 9px → 落地压缩 → `elastic.out` 回弹,`transformOrigin: 50% 100%`),
  对静态图与 psb/live2d/spine/unity 动画预览统一生效;`nextTick` 后执行,
  避开 keyed 组件重挂载时被销毁的旧节点
- [x] 刷新时复用 `StarBurst.trigger(x, y)` 放主烟花 + 两个错峰卫星烟花,
  坐标取点击点,键盘激活回退到卡片中心;冷却期间重复点击只显示提示,
  不播放动效;`prefers-reduced-motion` 下两者均跳过
- [x] 复用已有 `starBurst` 声明(移至动效区共用),`preview-stage` 加
  `ref` 与 `$event`;未引入新依赖,冷却逻辑与缓存头均未改动
- [x] `pnpm test:frontend` 32/32 通过;`pnpm build:web` 完整 SSG 12 路由
  预渲染成功(纯视觉动效,以人工检验为主,未新增单元测试)
- [ ] 人工检验:点击预览卡片应见人物轻微跳动与烟花;冷却期连点只有
  「2 秒冷却」提示且无新动效;开启系统「减少动态效果」后两种动效均不出现

---

## M22:右侧预览卡片高度适配屏幕

> 初版让卡片占满一屏,实测导致卡片过大、下方内容被挤出;
> 以下保留初始方案,最终修正见后两条。

- [x] 初版:桌面端(≥1024px)`Theme Preview` 面板设
  `min-height/max-height: calc(100vh - 6rem)` + `flex` column 占满视口高度
- [x] 初版:`.preview-stage` 用 `flex: 1 1 auto` 与
  `clamp(16rem, 46vh, 34rem)` 吸收高度,`.preview-image` 放宽到
  `clamp(18rem, 60vh, 38rem)`
- [x] **修正(卡片过大 / 下方内容不完整)**:移除面板的 `min-height` 与 flex,
  高度回归内容自适应,仅保留 `max-height: calc(100vh - 6rem)` +
  `overflow-y: auto` 应对极短窗口;stage 去掉 `flex:1`,收小到
  `clamp(14rem, 32vh, 24rem)`,图片同步收小到 `clamp(14rem, 44vh, 28rem)`,
  「下载当前帧」与主题元信息不再被挤出
- [x] **修正(动画留白)**:基础 `.preview-stage` 增加
  `padding-block: 1.75rem`,背景由 `--loli-surface` 改为 `transparent`
  (继承卡片 `--bg-card`),人物上下留出同色空白,跳动动画不再触边裁切;
  移动端同样生效
- [x] `pnpm test:frontend` 34/34 通过;`pnpm build:web` 完整 SSG 12 路由,
  产物 CSS 确认已移除强制 `min-height`、保留 `max-height`、含
  `padding-block` / `background:0 0`(压缩后的 transparent) / 新 clamp
- [ ] 人工检验:桌面 / 手机下卡片下方按钮与元信息完整可见、无内部滚动条;
  人物上下留白与卡片底色一致,点击跳动不被裁切

---

## M23:统计卡片新增「最快请求速度」

- [x] 三语 README 居中卡片由两列扩为三列(`width="33%"`,QQ 徽章行
  `colspan="3"`),新增 `14,000+` 及三语标签,`<sub>` 注明单核渲染基准口径
- [x] `/about` 页新增第三项展示数据:`fastestRendersPerSecond` 静态常量,
  `about.ts` 三语补齐 `fastestRequest` / `fastestNote` 并更新 `statsNote`;
  栅格改 3 列、移动端单列、数字字号相应缩小
- [x] 口径来自 `BenchmarkPrepareRender` 最快实测轮次 68,167 ns/op
  ≈ 14.6k 次/秒(取保守 14,000+),详见 `docs/about-page.md`;非端到端 QPS
- [x] `scripts/readme-stats.test.mjs` 校验三语卡片数字与列数一致,
  `about-statistics.test.mjs` 校验三语新文案键,均纳入 `test:frontend`
- [ ] 人工检验:三语 README 卡片三列对齐、徽章横跨整行;`/about` 三项
  在中/英/日与手机宽度下不溢出

---

## M24:修正千恋＊万花 / 常陸茉子主题的五官错配

- [x] **根因**:`d97cd45d5` 按正身/侧身把两套素材的身体交叉搬运,但眼睛、
  嘴/腮红层仍整组留在原主题且扁平随机;自带围巾的忍者装身体被无条件
  叠上正脸的 `face_96/97`,围巾上多出一张嘴,身体与五官朝向也会串场
- [x] `senren-mako` / `senren-mako-b` 的 `config.json` 改用 `sets`
  互斥套装:`casual` 保留 `face`,`ninja` 不含 `face`;渲染先选套装再挑部位,
  身体与五官不再跨造型混搭(变化数 2016→1596、588→588)
- [x] `internal/imgcore/asset/character_sets_test.go` 锁定配置契约:
  蒙面套装不得有 `face`、两套装身体范围不重叠、所有范围不越界;
  断言只针对配置结构,主题数量与美术内容变化不会打断测试
- [x] 按官方生成器重出两张画廊缩略图(`web/public/images/theme-thumbs/`),
  `docs/themes.md` 补充 `sets` 字段说明与实例
- [x] `go test ./...`、`go vet ./...`、`cmd/check-theme`、
  `pnpm test:frontend`、完整 SSG 均通过;实测冷装配与全量常驻各 16 帧,
  两种套装都能选中且五官无错配
- [ ] 人工检验:网页预览反复刷新两个主题,蒙面装围巾上不得出现第二张嘴,
  常服五官完整;画廊卡片缩略图与新配置一致

---

## M25:生成计数器 UI 优化(自然 / 优雅)

- [x] `ParamPanel.vue`:三列「参数 / 说明 / 值」表格改为柔和表单行
  (标签 + 参数 code 徽章 + 控件),去掉表头与密集网格线;模式条由
  `quick >> fine >> expert` 改为等宽分段控件(`aria-pressed`),
  名字栏改 1px 边框 + focus 光环并移除空值抖动动画;
  控件统一 hover/focus 过渡,移动端标签与控件堆叠
- [x] `index.vue`:结果区与空状态用
  `<Transition name="soft" mode="out-in">` 淡入淡出,空状态改虚线占位 +
  主/次文案层次;生成按钮加柔和阴影与 active 态,文字色改用
  `--bg-card` 兼容深色主题;`prefers-reduced-motion` 下关闭动效
- [x] `LinkOutput.vue`:从零散 utility 类改为 scoped 样式,按格式分组
  (标签 + 代码块 + 复制按钮),复制按钮改描边 + 图标,成功态用
  `--accent-soft` 而非硬绿色
- [x] 修 BUG-08 残留:`themesGallery.generateTitle` 的 en/jp 原为中文,
  改为 `Create a counter` / `カウンターを作成`;新增 `param.nameLabel`
  三语,`namePlaceholder` 改为示例 `my-counter`
- [x] `scripts/generate-form-i18n.test.mjs` 校验三语标题不再回退中文、
  名称标签/示例齐全、embed 文案齐全,纳入 `test:frontend`(37/37 通过);
  完整 SSG 12 路由,产物 CSS 确认新样式生效、旧 shake 与 `>>` 已消失,
  预渲染首页标题为 `Create a counter`
- [ ] 人工检验:中/英/日标题与占位文案;quick/fine/expert 切换;
  生成后结果淡入与复制反馈;手机宽度与深色主题下无溢出
