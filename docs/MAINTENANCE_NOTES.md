# 项目维护记录

## 当前接管基线（2026-05-20）

仓库：`edu-schedule-system`
分支：`dev`
本地路径：`/home/administrator/.openclaw/workspace/projects/edu-schedule-system`

### 技术栈

- 后端：Go，入口 `main.go`，Makefile 维护常用命令
- 前端：`frontend/web`，Vue 3 + Vite + Element Plus
- 数据库：MySQL 为主，Redis 可本地关闭

### 常用验证命令

后端完整基线：

```bash
make ci
```

前端构建：

```bash
cd frontend/web
npm ci
npm run build
```

配置检查：

```bash
make env-check
```

### 2026-05-20 首轮修复

接管后先跑基线，发现两个后端测试失败：

1. `internal/pkg/core` 直接依赖 `internal/service`，违反导入边界。
2. `internal/api/auth` 的限流 fail-open 测试失败：测试通过 `t.Setenv("AUTH_ENDPOINT_RATE_LIMIT_FAIL_OPEN", "")` 模拟未配置，但实现把空字符串视为显式配置，从而跳过本地 Redis 关闭时的 fail-open 逻辑。

已完成修复：

- 将 session actor context key 下沉到 `internal/pkg/core`，`internal/service` 继续通过兼容函数 `WithActor` / `ActorFromContext` 使用，消除 `core -> service` 反向依赖。
- 调整认证端点限流 fail-open 逻辑：空字符串不作为显式开关；Redis 关闭且非生产环境时默认 fail-open，显式 `false` 仍可关闭。

### 已通过验证

- `go test ./internal ./internal/api/auth ./internal/service/dashboard`
- `make ci`（包含 `go vet ./...`、`go test ./... -cover`、`go build ./...`）
- `frontend/web && npm run build`

### 后续维护建议

- 每次功能开发前先确认 `git status --short`，避免混入无关改动。
- 涉及 Handler/DAO/Swagger/Wire 生成文件时，按 README/AGENTS 指令走生成命令，不手改生成文件。
- 后端改动至少跑相关包测试；合并前跑 `make ci`。
- 前端改动至少跑 `npm run build`。
- 需要本地联调时，优先使用 README 中的 `make mysql-up`、`make bootstrap`、`make local-run` 流程。
