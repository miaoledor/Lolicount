# Lolicount npm 包支持设计(未来提案)

> 状态:设计稿,未实施。实施时按第 10 节拆 milestone,每 commit 单一功能。
> 契约基准:`docs/projectDesign.md`「接口调用文档」;正确性基准:AGENTS.md 铁律 1/3/5。

## 1. 定位与目标

Lolicount 是服务端渲染 SVG 计数器;npm 包是它的**客户端分发层**,让开发者在自己的
项目(Vue / Nuxt / React / vanilla / SSG / Markdown 站点)里 `npm install` 后直接嵌入
计数器,而不是手写 URL 或粘贴 script 标签。

核心原则:**npm 包是薄客户端,服务端是唯一事实来源。**

- 客户端只做三件事:构造 URL、注入 `<img>`、按需调用已有 JSON 接口。
- v1 **不新增任何服务端 endpoint,后端零改动**。
- 客户端不计算、不缓存、不优化计数(铁律 1:真实计数 SVG 必须使用完整 no-cache 组合头保持实时)。
- 字段名与参数语义和「接口调用文档」1:1 对齐,改一边必查另一边(工程原则 6)。
- 计数增量永远来自真实 HTTP 请求(浏览器加载 `<img>` 或调用 `/api/count`),
  绝不碰 buffer/SQLite 路径(铁律 5)。

## 2. 包矩阵

| npm 名 | 内容 | 依赖 |
|---|---|---|
| `lolicount` | 核心:url builder + vanilla 运行时(data-* 自动初始化、img 挂载、count/read)+ 类型。ESM/CJS/IIFE 三种构建 | 零依赖 |
| `lolicount-vue` | Vue 3 `<Lolicount>` 组件 | peer: `vue ^3` |
| `lolicount-react` | React `<Lolicount>` 组件 | peer: `react ^18` |
| `lolicount-nuxt` | Nuxt 模块(自动导入组件 + runtime config `baseUrl`) | peer: `nuxt ^4`,dep: `lolicount-vue` |

命名前置工作:实施前先查 npm 上 `lolicount` 及各 `-vue/-react/-nuxt` 名字是否被占用;
被占则改 scope(`@lolicount/core` 等)或改名,一次定死。所有包统一 `license: AGPL-3.0`
(npm 页面会展示 copyleft 声明,属预期)。

## 3. 仓库布局

现状:根 `package.json`(`lolicount-dev`,concurrently/sharp)与 `web/` 已由
`pnpm-workspace.yaml` 纳入同一 workspace 并共用锁文件。实施 npm 包时再扩展
`packages/*`:

```
pnpm-workspace.yaml
  packages:
    - web
    - packages/*

packages/
  lolicount/               # 核心包
    src/
      types.ts             # CounterOptions / CountResponse,字段与接口文档 1:1
      url.ts               # buildCounterUrl(opts):校验 + 编码,纯函数
      img.ts               # mount(el | selector):data-* → options → <img>
      auto.ts              # init():全页扫描 [data-lolicount]
      fetch.ts             # count(name)/read(name)
      index.ts             # ESM 入口
    dist/                  # esm / cjs / iife(lolicount.min.js)
  lolicount-vue/src/
  lolicount-react/src/
  lolicount-nuxt/src/      # module.ts + runtime/

examples/
  vanilla/  react/  nuxt/  # 最小可跑示例,人工验证入口
```

工具链:

- 构建 `tsup`(esm + cjs + d.ts + iife),版本管理 `@changesets/cli`(四包 lockstep 同版本号)。
- 测试 `vitest` + `@vue/test-utils` + `@testing-library/react`(fetch 用 vitest mock,不引 MSW)。
- 包体检 CI:`publint` + `attw --pack` 防 exports/types 配置错误。

## 4. 核心 API 设计

### 4.1 选项类型(与查询参数 1:1)

```ts
interface CounterOptions {
  baseUrl?: string    // 默认 https://lolicount.top,自部署可覆盖
  name: string        // 非空;demo 为服务端保留值
  theme?: string      // 白名单 | random
  ftheme?: string     // 白名单 | random
  number?: number     // 0~999999,仅预览,不落库不 +1
  fsize?: number      // 0~500
  scale?: number      // 0.1~4
  unshowf?: boolean
  x?: number; y?: number    // 像素,与 rx/ry 二选一
  rx?: number; ry?: number  // 比例 0~1,与 x/y 二选一
  text?: string       // 最长 64 字符;{n} 替换为计数值
}
```

校验策略(原则 10,不过度防御):只拦「必然错误」——空 `name`、`number` 越界、
`x/y` 与 `rx/ry` 同时给;其余越界值 dev 模式 `console.warn`,最终由服务端 400 兜底。

### 4.2 URL 构造器(SSG / Markdown 安全,零 DOM 依赖)

```ts
import { url } from 'lolicount'

url({ name: 'miaoledor', theme: 'lian', fsize: 16, scale: 1.2 })
// => https://lolicount.top/@miaoledor?theme=lian&fsize=16&scale=1.2

url.markdown({ name: 'miaoledor' })
// => ![miaoledor](https://lolicount.top/@miaoledor)
```

- 纯函数,可在 SSG 构建期执行 → 静态站点 / README 直接生成 `<img>` 或 Markdown。
- `name`/`theme` 走 `encodeURIComponent`;缺省值(theme=lian、scale=1、fsize=0 等)
  不拼进 URL,保持与现有三种嵌入方式一致的干净 URL。
- `baseUrl` 尾部斜杠归一化。

### 4.3 Vanilla 运行时(无构建器场景)

```html
<!-- 方式 A:ESM,一行 -->
<div data-lolicount="miaoledor" data-theme="lian" data-fsize="16"></div>
<script type="module">
  import('https://cdn.jsdelivr.net/npm/lolicount@1/lolicount').then(m => m.init())
</script>

<!-- 方式 B:IIFE 零配置(与现有 /widget/widget.js 约定一致) -->
<div data-lolicount="miaoledor" data-theme="lian"></div>
<script src="https://cdn.jsdelivr.net/npm/lolicount@1.2.3/dist/lolicount.min.js" defer></script>
```

- `init()` 扫描 `[data-lolicount]`,向每个元素注入
  `<img alt="<name>" loading="lazy" src="...">`;计数增量由浏览器图片请求天然触发,
  与现有「SVG address / Img tag / Markdown」三种嵌入方式语义完全一致。
- `data-*` 属性名与查询参数同名(`data-theme`、`data-fsize`、`data-unshowf`…),
  与 emote widget 的肌肉记忆一致;布尔/数字/字符串按参数类型解析。
- `data-base-url` 覆盖服务端地址(自部署)。
- 同页多实例相互独立,无全局状态。

### 4.4 JSON 访问(自定义 UI 用,如挂件扩展)

```ts
const { name, num } = await lolicount.count('miaoledor') // GET /api/count/@:name,真实 name 自增
const { name, num } = await lolicount.read('miaoledor')  // GET /record/@:name,只读
```

- 响应体 `{ "name", "num" }` 与现有接口 1:1,不新增契约。
- 文档显著警告:`count()` **会自增**(name 级限流超限降级只读,铁律 3);
  纯展示一律用 `read()`。
- `demo` / `number>0` 语义跟随服务端,客户端不做特判。

### 4.5 框架组件

```vue
<Lolicount name="miaoledor" theme="lian" :fsize="16" scale="1.2" class="w-40" />
```

```tsx
<Lolicount name="miaoledor" theme="lian" fsize={16} scale={1.2} width={160} />
```

- 内部就是渲染一个 `<img src={url(opts)} alt={name}>`。
- **SSR/SSG 安全**:组件没有任何 fetch/自增副作用,只产出预计算 src 的 img 标签;
  真实增量发生在访客浏览器加载图片时。SSG 构建期只生成 HTML 字符串,不计数。
- props = `CounterOptions`(去 `baseUrl`)+ 透传 `className`/`style`/`width`。
- `number` prop 用于站点编辑器 WYSIWYG 预览(不计数)。

### 4.6 Nuxt 模块(`lolicount-nuxt`)

```ts
// nuxt.config.ts
export default defineNuxtConfig({
  modules: ['lolicount-nuxt'],
  runtimeConfig: { public: { lolicount: { baseUrl: 'https://lolicount.top' } } },
})
```

- 自动导入 `Lolicount` 组件,组件内从 `useRuntimeConfig()` 取 `baseUrl`。
- 本仓库 `web/`(Nuxt 4)是现成的 dogfooding 样本。

## 5. 与铁律的关系(正确性约束)

1. **客户端零缓存**:组件/运行时不得为计数请求加任何缓存(内存、localStorage、
   max-age)。只有确定性的单帧 demo 会由服务端长缓存,客户端无感知。
2. **客户端不裁决**:主题白名单、name 级降级、IP 级 429 全归服务端;客户端只透传
   400/429 错误状态,不自行重试计数请求(重试会重复 +1)。
3. **自部署警告必须写进文档**:把计数器放在 CDN 边缘/代理后面会永久卡住计数
   (铁律 1,camo + Fastly 实测);`baseUrl` 指向自部署时同样适用。
4. **契约单源**:选项字段名 = 查询参数名。接口变更顺序固定:
   `projectDesign.md` 契约节 → Go handler → TS 类型,三处同 commit 周期内完成。
5. **不碰存储**:npm 包不引入任何写入路径,不存在与 `counter.Buffer` 的交互。

## 6. 构建与打包

- `tsup` 三种格式:`esm`(主)、`cjs`、`iife`(挂 `globalThis.lolicount`)。
- exports 地图:

```json
{
  "exports": {
    ".": { "types": "./dist/index.d.ts", "import": "./dist/index.js", "require": "./dist/index.cjs" },
    "./auto": { "types": "./dist/auto.d.ts", "import": "./dist/auto.js", "require": "./dist/auto.cjs" },
    "./package.json": "./package.json"
  },
  "unpkg": "dist/lolicount.min.js",
  "jsdelivr": "dist/lolicount.min.js",
  "sideEffects": false
}
```

- 体积目标:核心 dist < 4 KB(min+gzip),纯 URL 构造 + DOM 注入,零依赖。
- 环境目标:Node >= 22、evergreen 浏览器。
- 发布文件:`dist/` + `README.md` + `LICENSE`;`src`/测试不发布。
- CI 体检:`publint` + `attw --pack`,防止条件导出/类型解析错误。

## 7. CI / 发布

- 新 workflow `npm-publish.yml`,**仅由 `npm-v*` tag 触发**——刻意与服务端 `v*` tag
  区分,避免误跑 `release.yml` 的 docker 构建推送。
- 流程:`pnpm install --frozen-lockfile` → 构建全部包 → `vitest` 全绿 →
  `publint` + `attw` → `npm publish --provenance`(OIDC trusted publishing,
  无长期 token;备选 `NPM_TOKEN` secret)。
- 版本:changesets,四包 lockstep。
- script tag 通道版本固定:官方文档推荐经 jsDelivr/unpkg 锁版本
  (`lolicount@1.2.3`)。**服务端不托管 npm dist 产物**(保持单二进制,不新增静态文件;
  现有 `/widget/widget.js` 是独立机制,不动)。

## 8. 测试

- 单元测试(vitest):
  - url builder:全参数组合;缺省值不拼 URL;越界 warn;保留值 `demo`/`random` 放行;
    中文/特殊字符编码;`baseUrl` 尾斜杠归一化。
  - data-* 解析:`unshowf`(bool)、`fsize`/`scale`(number)、`theme`(string)类型转换;
    非法值 fallback + warn。
  - `count`/`read`:mock fetch,断言请求路径(`/api/count/@x` vs `/record/@x`)、
    响应解析、400/429 错误映射、不做重试。
- 组件测试:`@vue/test-utils` / `@testing-library/react`,断言渲染出的
  `img src` / `alt` 与 props 映射正确,SSR 渲染(`renderToString`)输出稳定。
- 「主题不影响测试」(AGENTS.md test 节):url/运行时测试把 `theme` 当不透明字符串
  透传,不依赖任何具体主题存在;默认值断言只引用 `types.ts` 中的常量,不散落硬编码。
- E2E(可选,后置):真实 Go 服务器(独立测试 DSN)+ Playwright,断言示例页
  img 为合法 SVG 且刷新后计数 +1。

## 9. 文档与人工检验

- 每包 `README.md`(安装/组件用法/script tag 用法/参数对照表,回链
  `projectDesign.md` 契约节);根 README 增加「npm 使用」一节(组件 / script tag /
  SSG 字符串三种嵌入快览)。
- `docs/TODOlist.md` 增加独立的未来里程碑,随 commit 更新任务状态。
- 人工检验示例(每阶段完成后执行):
  1. `pnpm --filter lolicount test` 全绿;`pnpm -r build` 后 `publint`/`attw` 无告警。
  2. `examples/nuxt` 起 dev server:刷新页面两次,计数各 +1;切换 `theme`/`fsize`
     props,URL 与渲染随之变化;DevTools 确认请求头为 `no-store` 系列。
  3. `examples/vanilla` 双击 `index.html`:script tag 加载,计数图正常显示。
  4. 干净目录 `npm i lolicount`,拷示例进一个最小 Vite 项目,`build` 后验证产物
     与页面效果,确认无 peer 依赖告警。
  5. `curl -I https://cdn.jsdelivr.net/npm/lolicount@<ver>/dist/lolicount.min.js`
     确认 CDN 可取(发布后)。

## 10. Milestone 拆分(一 commit 一功能)

- [ ] **P0 前置**:npm 名字可用性确认 + 账号/scope 决策;workspace 纳入
      `packages/*`——需人工确认 lockfile 迁移无副作用
- [ ] **P1** `packages/lolicount` 核心:`types.ts` + `url.ts` + vitest + tsup 构建配置
- [ ] **P2** 核心运行时:`img.ts`/`auto.ts`/`fetch.ts` + IIFE 构建 + 对应测试
- [ ] **P3** `lolicount-vue` + `lolicount-react`(含组件测试)
- [ ] **P4** `lolicount-nuxt` 模块
- [ ] **P5** `npm-publish.yml` + changesets + 首次发布
- [ ] **P6** 文档:各包 README、根 README 章节、TODOlist 状态回填

## 11. 待用户决策

1. 包名:`lolicount` 在 npm 是否可用?可用则用平名 + `-vue/-react/-nuxt` 后缀;
   否则改 scope `@lolicount/*`。
2. v1 框架范围:Vue + Nuxt + React 全上,还是 Vue/Nuxt 先行(项目本体是 Vue 生态),
   React 后置?
3. 发布账号:挂 miaoledor 个人账号还是新建 org?影响 scope 与权限管理。
4. IIFE script tag 通道是否 v1 保留(建议保留:成本低,覆盖无构建器用户)。
