# assets/psb/ — E-mote (PSB) 模型目录

此目录存放 Emote 挂件使用的 E-mote 角色模型。设计文档见
[docs/emote-widget.md](../../docs/emote-widget.md)。

## 放置约定

```
assets/psb/
  <model-name>/          # 仅小写字母/数字/连字符
    model.psb.gz         # 推荐:pure ems 规格 PSB 的 gzip 压缩存储(gzip -9)
    # 或 model.psb       # 未压缩形式,同样支持
```

- 仓库自带示例模型:`azuki`、`vanilla` 等(均为 .gz 存储,单个体积约 2~4MB);
- PSB 内部主要是未压缩的 RGBA 纹理数据,gzip 压缩率约 8 倍——**入库模型请压缩为
  `model.psb.gz`**;服务端直接以 `Content-Encoding: gzip` 下发,浏览器透明解压,
  挂件驱动无需任何改动;
- 新增模型后重启服务即被 `embed.FS` 打包;**请只提交你有权分发/展示的模型素材**,
  版权素材请自行评估;
- `/psb/:model` 响应是 `immutable` 长缓存:**更换某目录内的模型内容时请同时更换
  目录名**(或递增版本号),否则浏览器会继续使用旧缓存。

## 模型从哪来

浏览器端 WebGL 驱动(E-mote 3.9)只能加载 **pure、`spec=ems`** 的 PSB。
游戏分发的模型通常带壳/加密/`spec=win`,需要离线归一化,转换方法与工具见
设计文档「模型从哪来 / 如何转换」一节。
