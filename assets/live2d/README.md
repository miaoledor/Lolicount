# assets/live2d/ — Live2D (Cubism) 动态立绘模型目录

此目录存放 Live2D 计数器使用的角色模型。设计文档见
[docs/live2d-widget.md](../../docs/live2d-widget.md)。

## 放置约定

支持**两种** Cubism 格式,引擎按清单自动选择运行时:

- **Cubism 3**:`model3.json` / `<name>.model3.json` 清单 + 标准 `.moc3` 二进制;
- **Cubism 2(legacy)**:`model.json` / `<name>.model.json` 清单 + 标准 `.moc` 二进制
  —— BanG Dream! 等游戏提取物走这条路径,由随附的 `live2d-legacy.min.js`(官方
  Cubism 2 Web 运行时)渲染。

```
assets/live2d/
  <model-name>/                    # 仅字母/数字/连字符(见 live2dModelPattern)
    model3.json                    # Cubism 3 模型清单(入口,二选一)
    model.json                     # Cubism 2 / BanG Dream 模型清单(入口,二选一)
    <name>.moc3                    # Cubism 3 骨骼/网格二进制(被清单引用)
    <name>.moc                     # Cubism 2 骨骼/网格二进制(被清单引用)
    texture_00.png                 # 纹理(必需;被清单引用)
    physics3.json / physics.json   # 物理(可选)
    pose3.json                     # 遮挡(pose)(可选)
    <motion>.mtn                   # 动作(可选,可多个;untitled 引擎读取清单 Motions)
    <expr>.exp3.json               # 表情(可选)
```

- **入口文件**:必须是 `model3.json` / `<name>.model3.json`(Cubism 3)或
  `model.json` / `<name>.model.json`(Cubism 2 / BanG Dream)。`live2dHasManifest`
  以「目录里存在上述任一清单文件」判定一个模型目录有效;
- 清单内的文件引用用**相对路径**引用 moc/moc3 / 纹理 / 动作等,文件须与清单
  同目录(本服务按「模型目录 + 相对路径」取字节,见 `live2dModelHandler`);
- 本特性**只有交互(`<iframe>`)路径**:角色由 `live2d-player.html` 在浏览器里实时
  WebGL 渲染。没有免 JS 的 `<img>` 路径(Live2D 需 WebGL + JS,预渲染动画图又太大,
  见 `docs/live2d-widget.md` 第 1 节)。

## 交互行为(点击切换动作,无鼠标跟踪)

`live2d-player.html` 的交互约定:

- **取景**:渲染时按「不透明像素包围盒」做 contain-fit,把整个角色(头到脚 / 头到腰,
  取决于模型本身是全身体还是半身像)完整放进画布并留少量边距——不做半身裁剪、
  也不上移偏移,全身体与半身模型都适用;
- **无鼠标/眼动跟踪**:关闭引擎 Automator 的 `autoFocus`/`autoHitTest`,角色眼睛与
  头部**不**跟随鼠标;
- **点一下换一个动作**:仿 PSB/emote 播放器,收集清单里的 motion 分组(跳过以 `-`
  开头的分隔行与「初期化 / 視線追従」),每次点击按序切换到下一个动作。若模型未带
  `.mtn` 动作文件(见下文 BanG Dream 说明),点击无效果、保持静态 idle。

## BanG Dream 模型需要自带动作文件

BanG Dream! 角色的**完整模型包**通常附带 `.mtn` 动作与 `.json`/`.exp3.json` 表情
文件,并在 `model.json` 的 `motions` / `expressions` 字段列出。只有 `.moc` + 贴图、
而 `model.json` 里 `motions` 为空的「裸」模型(仅骨骼网格)加载后是**静态**的:
取景与无眼动照常工作,但「点击换动作」没有可切换的内容。要演示动作切换,请把带
`.mtn` 文件、且 `model.json` 已列出这些动作的**完整**角色包放入 `assets/live2d/<name>/`。

## 本地已放置的 BanG Dream 角色(17 名)

全部选**演出服(live)变体**——短裙/短裤款式,像游戏剧情画面一样能看到大腿;`casual`
是半身舞台模型、`furisode` 是长袍盖腿,都不符合需求。经典 12 名取自
`KitsuneX07/bangdream-live2d-models`,MyGO 5 名取自 **Bestdori CDN**(上游源)。

**MyGO 关键坑**:bytehunter/U1s1-king 等提取仓库只带了 `texture_01.png`,而
buildData 清单显示 live 模型是**双纹理**——`texture_00.png` 在 `<id>_general`
bundle、`texture_01.png` 在服装 bundle。缺 slot-0 纹理时 legacy 渲染器会对每个
drawable 报 `texParameter: no texture bound` 并渲染全空白。**两张都要从
`bestdori.com/assets/jp/live2d/chara/<bundle>_rip/` 下载**(buildData.asset 里
`textures[].bundleName/fileName` 写明出处)。MyGO 的 moc 不带 `.mtn`,动作复用
kasumi 的(BD 全系共用标准 Cubism 2 参数 ID,呼吸/眨眼/摆动直接生效)。

| 目录名 | 角色 ID | 来源模型包 | 动作数 |
|---|---|---|---|
| `kasumi` | 001 | `001_live_sr_01` | 38 |
| `tae` | 002 | `002_live_sr_01` | 23 |
| `rimi` | 003 | `003_live_sr_01` | 25 |
| `saya` | 004 | `004_live_sr_01` | 27 |
| `arisa` | 005 | `005_live_sr_01` | 26 |
| `ran` | 006 | `006_live_sr_01` | 34 |
| `moca` | 007 | `007_live_sr_01` | 40 |
| `himari` | 008 | `008_live_sr_01` | 33 |
| `tomoe` | 009 | `009_live_sr_01` | 34 |
| `tsugumi` | 010 | `010_live_sr_01` | 34 |
| `kokoro` | 011 | `011_live_r_2023` | 38 |
| `kaoru` | 012 | `012_live_default` | 34 |
| `tomori` | 036 | Bestdori `036_live_sr_01` + `036_general` | 38(复用) |
| `anon` | 037 | Bestdori `037_live_sr_01` + `037_general` | 38(复用) |
| `rana` | 038 | Bestdori `038_live_sr_01` + `038_general` | 38(复用) |
| `soyo` | 039 | Bestdori `039_live_sr_01` + `039_general` | 38(复用) |
| `taki` | 040 | Bestdori `040_live_sr_01` + `040_general` | 38(复用) |

放置步骤(以新增一名角色为例):

1. 从上述仓库的 `models/<id>_<variant>/` 取一个**贴图是真实 PNG**(而非 14KB 左右的
   HTML 占位页)、且 `motions/` 里有 `.mtn` 的模型包——各角色的 `live_default` 贴图
   普遍是坏的。要「能看到大腿」选 `<id>_live_sr_01` 等演出服;MyGO 从 Bestdori 下,
   注意双纹理;
2. 把 `moc` / `physics.json` / `textures/*.png` / `motions/*.mtn` /
   `expressions/*.exp.json` **全部扁平化**进 `assets/live2d/<name>/`(模型目录内无同名
   冲突,可直接按文件名铺平);
3. 写一个标准 Cubism 2 `model.json`:`{"version":2,"name":"<name>","model":"<name>.moc",
   "textures":["texture_00.png",…],"physics":"<name>.physics.json",
   "motions":{"motion":[{"file":"…mtn"},…]},"expressions":[{"name":"…","file":"…exp.json"},…]}`;
4. 重新 `go build`,重启服务,`/api/live2d/models` 即出现新名字;交互页用
   `/live2d-player.html?model=<name>` 打开测试。

> 上表是各角色对应的模型包来源,便于本地替换/升级。

## Project SEKAI 角色(26 名,随代码入库)

Project SEKAI 的**剧情演出立绘就是 Live2D 模型**,不是切图立绘:剧情脚本里的
`CostumeType`(如 `01ichika_normal`)直接对应一个模型 bundle,`FacialName`
(如 `face_sad_01`)对应一组表情动作。所以本目录放的是每名角色的**默认演出服**
(`<NN><name>_normal`)+ 该角色共享的动作集,**表情差分(表情切换)就是这里的
`face_*.motion3.json`**——点击角色循环切换,和游戏剧情里换表情是同一套资源。

来源是游戏解包资源的公开镜像 `storage.sekai.best`(`sekai-live2d-assets` 桶),与
BanG Dream 各模型包的性质一致。26 名可操作角色(24 名 + MEIKO / KAITO)全部收录,
目录名 `pjsk-<modelName 去数字前缀>`,如 `pjsk-ichika`。

**坑(都已在入库前处理):**

- **stub 纹理会让整个模型渲染全空白**。`08shizuku_normal` 带两张 2048 图集,其中
  `texture_01.png` 实际只有 14×19 像素的有效内容;把它留在 `Textures` 数组里,
  Cubism 运行时整个模型什么都不画(且不报错)。入库前按 alpha 覆盖率量一遍,低于 2%
  的图集直接丢弃——这和下文 MyGO 的「双纹理」坑是同一个问题的两面:图集**少了**不画,
  **多了空壳**也不画。
- **同一套服装有多个骨骼版本**(`_t01` / `_t02` / …,每个都是完整 moc3 + 贴图)。
  6 名角色(ena / miku / rin / len / luka / kaito)的 `_normal` 是这种情况,取版本号
  最小的一个,保证可复现。
- **表情与身体动作分两个 bundle**:模型目录带 `motions/`,但剧情通用动作在
  `live2d/motion/v1/main/<NN>_<name>/<model>_motion_base/{facial,motion}/`。
  `facial/` 全收(表情差分),`motion/` 只收默认 `normal` 性格预设里的
  nod / tilthead / shakehead / pose 等基础动作,避免每个角色再塞 240 个情境动作。

清单里 `Motions.Idle` 放基础动作(引擎自动循环),`Motions.TapBody` 放全部表情
(点击依次切换)。



官方运行时只能解析**标准格式**:

- Cubism 3 的 `live2dcubismcore.min.js` 只解析 **Cubism Editor 导出的标准 moc3**
  (魔数 `MOC3`、小端 section 表);
- Cubism 2 的 `live2d-legacy.min.js` 只解析**标准 moc**(魔数 `moc`、版本字节
  8–11)—— BanG Dream 提取物正是这种标准 Cubism 2 moc,可直接渲染。

某些非标准提取容器(魔数变体、自定义头部编码)两种运行时都无法解析,放进本目录
后加载会 404 / 报错。因此放进来前请确认 moc/moc3 是标准格式(可用 Live2D Cubism
Viewer / 官方 SDK 打开验证)。`scripts/gen-live2d-model3.mjs` 可基于一个「原始模型
目录」自动生成缺失的 `model3.json` 清单(适用于有文件但无清单的 Cubism 3 模型)。

## 模型资产不入库

Live2D 模型(尤其是游戏提取物,如 BanG Dream 角色)体积大、多为受版权保护的素材,
与本仓库对 PSB / Spine 模型的一致约定相同:**仓库只提交放置流程与代码,不提交模型
二进制**。

- `assets/live2d/` 下默认只有 `.gitkeep`(以及本 README);
- 本地测试时把模型放入 `assets/live2d/<name>/`,重启即被 `embed.FS` 打包;
- 公开部署时,运营者自行把**有权分发/展示**的模型放入该目录再构建。
