# Unity UGUI 动态立绘计数器

## 1. 范围

本功能把游戏提取的 Unity AssetBundle 中的 UGUI 立绘转换成浏览器可直接播放的
Lolicount 动态主题。当前资源来自舞萌DX `AssetBundleImages/navichara`，包含 25 个
导航角色（编号 01–26，原包无 07，并排除静态联动角色 17、23），共 23 个。

浏览器不能直接执行 `.ab`。生产路径不加载 Unity 引擎，也不在请求时解析资源：

```text
AssetBundle（构建机）
  → scripts/extract-unity-ui.py + UnityPy
  → model.json + lossless WebP layers
  → embed.FS
  → unity-player.html（DOM/CSS + requestAnimationFrame）
```

这条路径保留资源实际使用的层级、RectTransform 局部坐标、旋转、缩放、显隐、颜色、
`MultipleImage` 下标换图和 `m_Sprite` 引用换图曲线。网页端只下载选中角色的模型，
不需要约 800 MB 的 Unity WebGL runtime。

## 2. 资源转换

安装构建期依赖：

```powershell
python -m pip install -r scripts/requirements-unity.txt
python -m playwright install chromium
```

从提取包重新生成：

```powershell
python scripts/extract-unity-ui.py 'E:\Downloads\AssetBundleImages'
python scripts/verify-unity-ui.py 'E:\Downloads\AssetBundleImages'
```

启动本地服务后，从真实播放器的固定首帧批量生成图库缩略图：

```powershell
python scripts/capture-unity-thumbnails.py
python scripts/capture-unity-thumbnails.py --check
```

默认输出：

- `assets/unity/maimai-navichara-XX/model.json`：浏览器模型；
- 同目录 `*.webp`：从 Sprite 无损转换的透明分层图；
- `web/public/images/emote-thumbs/maimai-navichara-XX.webp`：由播放器合成全部图层后截取的主题图库缩略图。

脚本按 AssetBundle 内部 CAB 名解析跨包 PPtr，不依赖来源机器 manifest 中的绝对
Jenkins 路径。AnimationClip 的 StreamedClip 使用 Unity 的小端多项式曲线格式
解码；DenseClip 与 ConstantClip 也会合并到同一标量轨道模型中。验证脚本会把每个
导出节点重新与 prefab 的 `m_LocalPosition` 对照，防止 anchored/local 坐标混用。

`model.json` 的稳定字段：

- `bounds`：初始姿势可见内容框，播放器据此 contain-fit；
- `nodes`：父节点索引、初始 position/rotation/scale/size/pivot、图片和备选图片；
- `animations`：动作名、时长、循环标记和标量 tracks；
- `tracks[].keys`：线性键或 `[time,value,a,b,c]` 三次多项式键。

`BlendImage` 的 Additive 图层会在构建期从黑底加色纹理转换为 straight-alpha WebP；
显隐与图片下标轨道标记为离散值，播放器不会在两个关键帧间错误插值。
动画中的 `m_AnchoredPosition` 会加上 prefab 的锚点参考偏移，转换成与 DOM 父子变换
一致的 local position；带 `preserveAspect` 的 Image 保持 Sprite 原始宽高比。

## 3. HTTP 接口

| 路径 | 作用 | 缓存 |
|---|---|---|
| `GET /api/unity/models` | 返回已嵌入模型名 | `public, max-age=60` |
| `GET /unity/models/:name/model.json` | 模型清单 | 一年 immutable |
| `GET /unity/models/:name/:sprite.webp` | 透明分层图 | 一年 immutable |
| `GET /unity-player.html?...` | 自包含 iframe 播放页 | 前端静态文件策略 |

模型名和文件名均为白名单匹配，不能借路由访问模型目录之外的文件。模型文件固定在
构建期嵌入，因此 immutable 与现有 Spine/Live2D 资源接口一致。

## 4. 播放器与嵌入

播放器查询参数：

| 参数 | 含义 |
|---|---|
| `model` | `maimai-navichara-XX`，必填 |
| `name` | 计数器名；缺省为 `demo` |
| `text` | `{n}` 计数模板 |
| `motion` | 可选动作名；缺省从 `Navi_Default` 待机动作开始 |
| `thumbnail=1&time=秒数` | 构建缩略图用的无计数器固定帧模式 |

示例：

```html
<iframe
  src="https://lolicount.top/unity-player.html?model=maimai-navichara-01&name=me&text=%E7%AC%AC%20%7Bn%7D%20%E4%BD%8D%E8%AE%BF%E5%AE%A2"
  width="360"
  height="520"
  frameborder="0">
</iframe>
```

页面打开后调用 `GET /api/count/@:name`，因此真实名称仍实时自增并走 `no-store`；
模型/图片本身才使用 immutable。点击角色会随机切换到另一个非占位动作；指定
`motion` 主要用于调试、截图与可复现预览。
`prefers-reduced-motion: reduce` 下只渲染首帧，不启动动画循环。

## 5. 前端集成

`GET /api/themes` 将这些模型标记为 `animated:true, kind:"unity"`。主题图库把
`unity` 作为独立类型，并通过 `ModelIframePreview` 选择 `unity-player.html`。
同一组件也修正了旧逻辑中 Spine 被错误传给 E-mote `EmotePreview` 的问题；
Spine、Live2D、Unity 现在分别进入自己的播放器页面。

## 6. 限制

- 这是 UGUI 2D 转换器，不是通用 Unity 场景或 3D 模型播放器；
- RGB 颜色曲线在 DOM 中按平均亮度近似，alpha 保持精确；
- 当前只接入 `navichara` 动态立绘。`partner` 是 128×128 静态头像，未伪装成动态主题；
- 原始游戏素材的版权归其权利人所有，部署者应自行确认使用与分发授权。
