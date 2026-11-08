# Support 页面

访问 `/support`。顶部导航、手机展开菜单和页脚均提供入口，`/support/` 同样可直达。
页面使用现有中、英、日文切换和深浅配色，赞助入口沿用项目已有的
[GitHub Sponsors](https://github.com/sponsors/miaoledor)。

## 设计与交互

布局参考 [chanfoo 的 Support 页面](https://chanfoo.net/en/support)：大标题、
玻璃爱心罐、独立的爱心按钮和下方支持入口。罐体、金属扣件、打开的罐盖和
立体爱心由 Three.js 几何体原创生成，不加载参考站模型或外部模型资源。

- 点击罐体或下方爱心按钮，投入一颗爱心；爱心受重力影响，与罐壁、罐底和其他爱心碰撞。
- 左右拖动罐体旋转视角；聚焦罐体后，回车或空格投放爱心，左右方向键旋转。
- 初始 20 颗，最多保留 48 颗；继续点击时移除最早的一颗，控制渲染和碰撞开销。
- 致谢数量只记录当前页面的互动，刷新或离开页面后重置，不代表捐款、捐赠人数或服务器访问计数。
- 赞助卡跳转 GitHub Sponsors；其他卡片分别跳转仓库、主题贡献指南和代码贡献指南。

## 运行与降级

`web/app/pages/support.vue` 是单根元素页面，SSG 显式生成 `/support/index.html`，
随现有静态资源一起嵌入 Go 二进制。Three.js 场景只在该页面挂载后动态导入，
首页和其他路由无需初始化 WebGL。

场景代码位于 `web/app/utils/support/`，分别负责模型、渲染生命周期和球形接触近似。
像素比限制为 1.75；静止、离开视口或切换后台后停止连续渲染。
组件卸载时清理观察器、事件监听、动画帧、几何体、材质、环境纹理和 WebGL 上下文。

服务端预渲染先显示静态爱心罐；WebGL 初始化失败或上下文丢失时保留静态展示，
赞助与贡献链接始终可用。`prefers-reduced-motion` 下投放直接呈现稳定堆叠结果，
键盘和拖动旋转直接更新视图，并取消爱心按钮的弹跳动画。

本功能不增加 API、不修改计数 SVG、缓存、限流或 SQLite schema。
**无需数据库迁移，也无需对 `data/count.db` 执行命令。**

## 自动验证

需要 Node.js 22.6+（项目开发环境建议当前 Node.js 22 LTS）：

```sh
pnpm test:support
pnpm --dir web generate
```

物理测试覆盖重力、罐底与罐壁接触、重叠初态、长时间步、满罐长期稳定、
新心形加入堆叠，以及三种语言的文案完整性。测试不依赖主题内容或数量。

构建后，按项目现有发布流程将 `web/dist/` 内容复制到 `assets/dist/`，再运行：

```sh
go test -race -count=1 -p 1 ./internal/server ./internal/counter ./internal/store ./internal/ratelimit
```

`TestFrontendServesSupportPage` 校验 `/support` 和 `/support/` 均返回独立页面、
静态降级内容和正确的 HTML 缓存头；缺少构建产物时会显式跳过。

## 人工检验

1. 运行 `pnpm dev`，访问 `http://localhost:3721/support`。
2. 点击罐体与爱心按钮各几次，观察心形下落和堆叠；拖动旋转；用 Tab 聚焦罐体，
   按 Enter、空格、左右方向键，检查相同操作是否可用。
3. 连续投放超过 48 次，确认仍可互动、罐内不会无限增加对象；切换到关于页再返回，
   确认模型重新出现且数量重置。
4. 切换中文、英文、日文及深浅模式，在 320px、390px、768px 和桌面宽度检查
   标题、模型、赞助卡和贡献卡，无水平溢出或不可点击的链接。
5. 开启系统或浏览器的「减少动态效果」，确认投放直接显示结果，内容和键盘操作仍可使用。
   在禁用 WebGL 的浏览器中打开页面，确认展示静态爱心罐，支持链接仍可用。
6. 本地构建 Go 服务后直达或刷新 `/support`、`/support/`，确认不会回落到首页。

本次 SSG 构建、12 项物理与多语言测试，以及 Go server / counter / store /
ratelimit 的 race 回归均通过。额外运行全仓 Vue 类型检查时，旧的
BackToTop、EmotePreview、HomeHero、LangSwitch、NavBar、index 等文件
仍存在已有类型错误；新增 Support 页面、组件和场景文件没有报告类型错误。
开发预览还会报告现有 `app.vue` 的空 `noscript` 节点水合警告，未纳入此次功能修改。
