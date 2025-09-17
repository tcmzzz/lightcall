# AI CONVENTIONS

本文件用于指导各个AI工具在本项目的开发方式, 包括不限于: Claude Code (claude.ai/code), Gemini CLI, Qwen CLI, Crush.
通过 @ARCH.md 了解项目结构和数据关系

## WorkFlow

**ALWAYS follow these instructions while makeing plans:**
1. 如果觉得缺少必要信息或者不确定时, 请先询问用户
2. 不要过度设计, 优先考虑简单直接的方案
3. 使用三方代码库或者组件时, 优先查阅使用文档. (通过 mcp context7)


**ALWAYS follow these instructions while coding:**
1. 不要尝试启动服务, 开发容器环境已经由研发人员手工启动
2. 查看前端`pnpm dev`日志可以使用 `docker compose -f example/dev/docker-compose.yaml logs --tail 30 frontend`
3. 前端代码变更后, lint使用 `make lintfront`
4. 服务端代码变更后, lint使用 `make lintgo`
5. 执行npm命令时要去容器环境执行. 假设要执行`pnpm install`, 命令为 `docker compose -f example/dev/docker-compose.yaml exec frontend pnpm install`
6. 前端布局使用flex,
7. 前端CSS使用tailwind css

## Code Style

- **Conciseness**: Write clean, minimal code; fewer lines is better
- **Comments**: Only include comments that are essential to understanding functionality or convey non-obvious information
- **Go**: Use standard Go error handling with detailed error messages
- **Error Handling**: Be explicit but concise about error cases
- **Go Resources**: Always use `defer` for resource cleanup like `rows.Close()` (sqlclosecheck)
- **Go Defer**: Avoid using `defer` inside loops (revive) - use IIFE or scope properly
