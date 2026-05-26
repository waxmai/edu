# 新模块 Checklist

复制模板新增业务模块时，按以下清单逐项完成，避免遗漏生成代码、路由、Wire 和测试。

如果是从模板创建全新项目，请先完成 `docs/TEMPLATE_ADOPTION.md` 的项目初始化 Checklist，再使用本清单新增业务模块。

## 数据层

- [ ] 可先运行 `make new-migration TABLE={table}` 生成迁移草稿，再补齐 SQL。
- [ ] 新增 `migrations/{version}_create_{table}.sql`，只提交可重复执行或明确幂等的 SQL。
- [ ] 迁移文件包含 `-- rollback-policy: ...` 头；已合并迁移不得修改，修正时新增 forward migration。
- [ ] 本地执行 `make migration-smoke`，确认基础迁移和 bootstrap 可重复执行。
- [ ] 使用 live schema 生成 DAO/Model：`make gen-dao DSN='app:app@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local' TABLES='{table}'`。
- [ ] 不手改 `internal/repository/mysql/{dao,model}/*.gen.go`。

## 服务层

- [ ] 在 `internal/service/dto/` 定义请求 DTO 和响应 DTO。响应 DTO 不直接引用 generated DB model。
- [ ] 在 `internal/service/{module}/service.go` 定义 `Service` 接口和实现。
- [ ] Service 方法使用 `context.Context`，不导入 Gin 或 `internal/pkg/core`。
- [ ] 读操作返回响应 DTO，内部再从 generated model 映射。
- [ ] 更新接口必须做字段白名单，避免 mass assignment。
- [ ] 权限/归属检查放在 Service 业务规则中，权限拒绝返回 `apperr.Forbidden(...)` 并覆盖测试。
- [ ] 可预期业务失败返回 `apperr.InvalidArgument/NotFound/Conflict/Forbidden/DependencyFailed`。
- [ ] 添加 service 单元测试，覆盖校验、事务、白名单、not found/conflict 等业务分支。

## API 层

- [ ] 运行 `make gen-handler TABLE={table}` 生成 HTTP 适配层。
- [ ] 在 `internal/router/router.go` 注册 `RegisterGenerated{Module}Routes`。
- [ ] 需要鉴权的路由挂载到 `securedGroup`；公开路由必须在文档中明确说明。
- [ ] 在 `cmd/server/wire.go` 注册 service 和 API handler provider。
- [ ] 运行 `make wire`。
- [ ] 运行 `make swagger`。
- [ ] 不手改 `internal/api/*/*.gen.go`、`cmd/server/wire_gen.go` 或 Swagger 产物。

也可以在 SQL 已补齐、MySQL 已启动后运行 `make gen-module TABLE={table}` 串联 migration/bootstrap、DAO/Model、handler/router、Wire 和 Swagger。该命令不替代 DTO、Service、路由挂载、Wire provider 注册和测试实现。

## 验证

- [ ] 添加或更新 router/handler 集成测试。
- [ ] 运行 `make ci`。
- [ ] 运行 `make generated-check`，确认 Wire/Swagger 产物已提交。
- [ ] 有 schema 变化时在 MySQL 环境运行 DAO/model drift check。
- [ ] 自定义动态路由使用稳定路径或 `core.AliasForRecordMetrics`，避免高基数指标标签。
- [ ] 运行 `make vulncheck` 或确认 CI vulnerability scan 通过。
