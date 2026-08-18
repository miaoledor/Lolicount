# Contributing to Lolicount

[中文](./CONTRIBUTING.md) · [English](./CONTRIBUTING.en.md)

感谢你对 Lolicount 的兴趣!本项目有两种贡献方式,请根据你的目的选择:

- 🎨 **主题贡献** — 新增或改进内置主题(帧图 / 立绘)。
  详见 [主题贡献指南](./docs/contributing-themes.md)
- 💻 **功能贡献** — 修复 bug、新增功能、改进文档或部署。
  详见 [功能贡献指南](./docs/contributing-code.md)

两份指南之间可以互相跳转。无论哪种贡献,都请先阅读下面的通用约定。

## 通用约定

- **Commit message** 用英文,遵循 Conventional Commits(`feat:`、`fix:`、
  `docs:`、`refactor:`、`test:` 等)。正文可附中文说明。
- **代码注释** 用英文,只在必要处写(解释「为什么」,不解释「是什么」)。
- **一个 commit 只做一件事**,保持功能单一,便于 review 与回滚。
- **不要提交 `.env`**。它含密钥(R2/S3)。只提交 `.env.example`。
- **不要修改 `AGENTS.md`、`docs/projectDesign.md`、`docs/TODOlist.md` 的
  描述内容**(这些文件标注禁止修改描述),只允许更新任务状态。
- 涉及数据库 schema 变更时,任务结尾必须告诉用户是否需要迁移。
- PR 前先跑本地校验:
  ```bash
  go vet ./...
  go test -race ./...
  go run ./cmd/check-theme
  pnpm --dir web generate
  ```

## 行为准则

请保持友善与尊重。任何形式的骚扰、人身攻击或歧视性言论均不可接受。
