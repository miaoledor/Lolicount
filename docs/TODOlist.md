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
- [x] AGENTS.md 铁律1 与缓存表同步为 no-cache 组合头(`no-store, no-cache, max-age=0, must-revalidate`,GitHub camo/Fastly 边缘无视裸 no-store;经用户一次性授权修改)
- [ ] `docs/projectDesign.md` 缓存契约仍写裸 `no-store`(文件禁止修改,待用户同步为 no-cache 组合头)
- [ ] `docs/projectDesign.md` 接口文档缺 `GET /api/themes/hot`(热门主题,启动/每小时更新)与 `tb_theme_usage` 表(文件禁止修改,待用户同步)
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
- [x] 实现主题画廊 `pages/themes.vue` + 上传页 `pages/upload.vue`
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
动态在线修改count字段位置 在线查看效果
支持外部图片源
支持外部插件-通过restful请求调用外部插件-兼容任何语言
支持仪表盘，查看访问排名和总访问次数
支持web上传主题
服务器捐赠内容添加
图片大小优化脚本
图片上面支持添加文字
文字支持在线动态调整位置，在线预览效果


---

## M13:Emote 挂件(PSB 动图计数器,方案 A POC)

> 设计文档:`docs/emote-widget.md`。E-mote (PSB) 角色 WebGL 实时渲染,每次打开随机动作,页面加载自动计数,第三方网页经 `<script>` 挂件引用。

- [x] 调研 PSB/E-mote 格式与浏览器播放方案(FreeMote-SDK WebGL 驱动)
- [x] `GET /api/count/@:name`:计数 JSON 接口(自增/demo/number 特例/CORS/no-store)+ 单测
- [x] `assets/psb/` 模型目录 + `GET /api/psb/models`、`GET /psb/:model`(immutable 长缓存/名称白名单)+ 单测
- [x] vendor WebGL 驱动到 `web/public/widget/`(附 CC BY-NC-SA 第三方声明)
- [x] `widget.js` 挂件:随机动作播放 + 画布像素静止检测自动换动作 + `{n}` 计数模板渲染
- [x] Playground 集成:`/api/themes` 追加 `animated` 标记,主题下拉显示动图模型,生成后
  WebGL 预览 + 挂件 HTML 嵌入代码(无独立页面,复用 `EmotePreview` 组件)
- [x] 修复 frontend.go 预渲染子页路由(目录 → index.html,/editor /about /emote 不再回落首页)+ 单测
- [x] 端到端验证:单二进制部署 + 跨域第三方页面挂件(示例模型 azuki/vanilla 随仓库分发)
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
