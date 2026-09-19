# AGENTS.md

## Digitalway Core（AI Skill）

本项目依赖 `github.com/digitalwayhk/core`。开发或修改其后端 API 前：

1. 确认存在 `.codex/skills/use-digitalway-core/SKILL.md`；缺失时在本仓库根执行:
   `DIGITALWAY_CORE_PATH=<core源码路径> <core>/scripts/link-consumer-skill.sh --target . --write-agents`
   或仅有模块依赖时:
   `CORE=$(go list -m -f '{{.Dir}}' github.com/digitalwayhk/core) && "$CORE/scripts/link-consumer-skill.sh" --target . --write-agents`
2. Codex / 兼容 Agent 必须阅读 `.codex/skills/use-digitalway-core/SKILL.md`。它是指针，正文在 `.codex/skills/core-skill/`（本仓库）或 core 仓库 `docs/ai/core-skill/`；按其「主题分片索引」按需读取分片，并结合 core 仓库 `docs/codex`、`examples` 执行。
3. GitHub Copilot 阅读 `.github/copilot/skills/core-backend-api.md`（若已安装），同样按指针进入权威源。
4. 当指南与当前代码、测试或公开契约不一致时，以代码、测试和契约为准。
5. Core 文档与示例路径：优先 `go list -m -f '{{.Dir}}' github.com/digitalwayhk/core`，或 `go.mod` 的 `replace` / 环境变量 `DIGITALWAY_CORE_PATH`。

完整说明: core 仓库 `docs/codex/CONSUMER_AI_SKILL_SETUP.md`。
