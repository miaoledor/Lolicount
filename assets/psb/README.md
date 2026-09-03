# assets/psb/ — E-mote (PSB) 模型目录

此目录存放 Emote 挂件使用的 E-mote 角色模型。设计文档见
[docs/emote-widget.md](../../docs/emote-widget.md)。

## 放置约定

```
assets/psb/
  <model-name>/          # 仅小写字母/数字/连字符
    model.psb            # pure、spec=ems 的 PSB 文件
```

- 模型文件**不提交进仓库**（体积大且有版权风险，已在 .gitignore 排除）；
  本地/部署机构建时放入即可被 `embed.FS` 打包。
- 放入或更换模型后重启服务生效。
- `/psb/:model` 响应是 `immutable` 长缓存：**更换某目录内的模型内容时请同时更换
  目录名**（或递增版本号），否则浏览器会继续使用旧缓存。

## 模型从哪来

浏览器端 WebGL 驱动（E-mote 3.9）只能加载 **pure、`spec=ems`** 的 PSB。
游戏分发的模型通常带壳/加密/`spec=win`，需要离线归一化，转换方法与工具见
设计文档「模型从哪来 / 如何转换」一节。请只使用你有权分发/展示的模型。
