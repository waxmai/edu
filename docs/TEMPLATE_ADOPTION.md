# 模板采用指南

本文面向把本仓库复制成新后端服务的场景。模板定位是 Go + Gin + MySQL/GORM schema-first 后端基座，不是数据库无关的通用框架。

## 1. 定位确认

采用前先确认团队接受以下约束：

- MySQL 是默认主数据源，应用启动会初始化 MySQL read/write 连接。
- 表结构以 SQL migration 和 live MySQL schema 为准，`internal/repository/mysql/{dao,model}` 由 `gormgen` 生成。
- Handler/Router 由 `handlergen` 生成，业务规则写在 `internal/service/{module}`。
- Redis 是可选缓存/基础设施依赖，本地可以通过 `REDIS_ENABLED=false` 关闭。
- 模板自带的 `admin` 和模板 Token 只用于示例，不代表正式账号、登录或权限模型。

如果新服务需要 PostgreSQL、SQLite、Mongo-first、事件驱动或强 repository interface 抽象，应先评估是否要 fork 出独立模板，而不是在业务项目里绕开这些默认假设。

## 2. 新项目初始化 Checklist

- [ ] 复制仓库后先创建业务分支，避免在模板主干上直接开发业务。
- [ ] 修改 `go.mod`：`go mod edit -module github.com/<org>/<service>`。
- [ ] 批量替换 import path：把 `edu-schedule-system/...` 替换为新 module path，例如 `github.com/<org>/<service>/...`。
- [ ] 运行 `go mod tidy`，确认没有旧 module path 残留。
- [ ] 修改项目名：`Makefile` 的 `PROJECT_NAME`、Docker Compose service/network/volume 名、Docker image tag、Prometheus job、Swagger 标题、`configs/constants.go` 的 `ProjectName`。
- [ ] 修改文档品牌：README、`docs/DEVELOPMENT.md`、`docs/OBSERVABILITY.md` 和部署文档中的 `edu-schedule-system`、端口、服务名。
- [ ] 准备本地 `.env`：从 `.env.example` 复制，不提交真实密钥。
- [ ] 配置 `JWT_SECRET`，并按业务需要调整 issuer/audience/leeway。
- [ ] 配置 `MYSQL_READ_ADDR`、`MYSQL_READ_USER`、`MYSQL_READ_PASS`、`MYSQL_READ_NAME`。
- [ ] 配置 `MYSQL_WRITE_ADDR`、`MYSQL_WRITE_USER`、`MYSQL_WRITE_PASS`、`MYSQL_WRITE_NAME`。
- [ ] 决定 Redis 策略：本地 no-op 使用 `REDIS_ENABLED=false`；启用时配置 `REDIS_ADDR`、`REDIS_PASS`、`REDIS_DB`。
- [ ] 配置 CORS：生产只允许明确域名，按需设置 `CORS_ALLOWED_ORIGINS` 和 `CORS_ALLOW_CREDENTIALS`。
- [ ] 配置认证：本地可用 `AUTH_MODE=disabled|auto`，联调和生产建议显式设置 `AUTH_MODE=required`。
- [ ] 关闭模板 Token：所有环境默认保持 `AUTH_TEMPLATE_TOKEN_ENABLED=false`；只有明确的本地调试窗口才临时改为 `true`，正式登录完成后本地也应关闭或移除。
- [ ] 按部署入口配置 Swagger、pprof、metrics：生产不要开启 Swagger/pprof，`METRICS_ENABLED` 只暴露给可信监控网络。
- [ ] 运行 `make env-check` 验证有效配置。
- [ ] 运行 `make ci` 验证编译、vet 和测试。
- [ ] 运行 `make generated-check` 验证 Wire/Swagger 产物一致。
- [ ] 如果改了 migration 或 generated DAO/Model，运行 `make migration-smoke` 并确认 DAO/model drift check 通过。

## 3. Demo 模块处理

内置 `admin` 模块覆盖了模板的主要开发路径，适合当参考代码：

- `migrations/000001_create_admin.sql`
- `migrations/bootstrap/000001_seed_admin.sql`
- `internal/repository/mysql/model/admin.gen.go`
- `internal/repository/mysql/dao/admin.gen.go`
- `internal/repository/mysql/dao/gen.go` 中的 `Admin` root DAO 引用
- `internal/service/dto/admin.go`
- `internal/service/admin/`
- `internal/api/admin/`
- `internal/router/admin_integration_test.go`
- `internal/router/router_test.go` 中依赖 fake admin handler/service 的认证和 token 测试
- `internal/service/admin/service_test.go`
- `docs/docs.go`、`docs/swagger.json`、`docs/swagger.yaml` 中的 admin paths/definitions
- `cmd/server/wire.go` 中的 admin provider
- `internal/router/router.go` 中的 admin handler 注入和 route registration

保留它时，应在项目文档中标注为 demo/reference module。移除它时，先接入至少一个真实业务模块，再删除上述文件和注册点；generated DAO root、Wire 和 Swagger 不应长期手改，优先通过 live schema 重新运行 `make gen-dao`、`make wire` 和 `make swagger` 刷新产物。随后运行：

```bash
make wire
make swagger
go mod tidy
make ci
make generated-check
```

如果只想隔离而不是删除，可以把 admin 路由挂到内部前缀或只在 `dev` 注册，但不要把它包装成正式权限系统。

## 4. 模板 Token 处理

`POST /api/v1/auth/token` 由 `internal/router/router.go` 注册，仅用于本地显式开启后的测试 JWT 生成。它不校验密码、验证码、第三方身份或真实权限。

正式项目应：

- [ ] 新增真实登录接口。
- [ ] 明确 access token / refresh token 生命周期。
- [ ] 实现 token revoke/rotation 策略。
- [ ] 在 Service 层实现角色、权限点和资源归属判断。
- [ ] 设置 `AUTH_TEMPLATE_TOKEN_ENABLED=false`，仅在明确本地调试时临时开启。
- [ ] 删除或隔离 `registerTemplateAuthRoutes` 相关逻辑，或保留为仅本地开发工具。

## 5. Schema-first 开发流程

新增业务模块时遵循固定顺序：

1. 编写 SQL migration。
2. 在本地或 CI MySQL 上应用 migration。
3. 从 live schema 运行 `make gen-dao DSN=... TABLES=...`。
4. 编写 DTO 和 Service，业务规则、事务、字段白名单都放在 Service。
5. 运行 `make gen-handler TABLE=...` 生成 HTTP 适配层。
6. 在 `internal/router/router.go` 注册路由，在 `cmd/server/wire.go` 注册 provider。
7. 运行 `make wire`、`make swagger`、`make ci`。

不要手改 generated DAO/Model、generated Handler/Router、Wire 产物或 Swagger 产物。字段变化应通过 migration 和重新生成代码完成。
