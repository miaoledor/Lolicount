# Lolicount 文档导航

> 所有文档的入口。按需选择,也可从 [README](../README.md) 跳转。

## 快速入口

| 文档 | 内容 | 适合谁 |
|---|---|---|
| [README](../README.md) | 项目介绍、快速开始、参数、API | 所有人 |
| [部署文档](./deployment.md) | 使用、开发模式、生产部署(Win/Mac/Linux) | 使用者 / 运维 |
| [架构文档](./architecture.md) | 总体架构、数据存储、渲染模型、限流、缓存、目录结构 | 贡献者 |
| [技术选型](./tech-stack.md) | 各层依赖、版本、选型理由 | 贡献者 |
| [主题文档](./themes.md) | 主题模型、目录约定、参数关系 | 主题作者 / 贡献者 |
| [项目设计](./projectDesign.md) | 接口契约、数据模型、项目结构(只读) | 贡献者 |
| [常见问题](./faq.md) | 使用 / 部署 / 开发 / 贡献常见问题 | 所有人 |
| [TODOlist](./TODOlist.md) | 里程碑与任务状态(只读描述) | 维护者 |

## 贡献指南

| 文档 | 内容 |
|---|---|
| [CONTRIBUTING.md](../CONTRIBUTING.md) | 贡献总览、通用约定、行为准则 |
| [主题贡献指南](./contributing-themes.md) | 卡片 / 立绘 / 文字风格主题贡献流程 |
| [功能贡献指南](./contributing-code.md) | Go / 前端 / CI / 文档贡献流程 |

## AI 协作

| 文档 | 内容 |
|---|---|
| [AGENTS.md](../AGENTS.md) | AI agent 指南:铁律、工程原则、项目结构(只读) |

## 按角色

**只想用**:README → 部署文档 → 常见问题

**想贡献主题**:README → 主题文档 → 主题贡献指南

**想贡献代码**:README → 架构文档 → 技术选型 → 功能贡献指南 → AGENTS.md

**想部署**:README → 部署文档 → 架构文档(存储/限流部分)

## 文件约定

- `AGENTS.md`、`docs/projectDesign.md`、`docs/TODOlist.md` 标注**禁止修改
  描述内容**(TODOlist 仅允许更新任务状态)。需同步描述请提 issue。
- `.env` 含密钥,**永远不要提交**,只提交 `.env.example`。
- `assets/themes.json` 由脚本生成,**不要手动编辑**。
