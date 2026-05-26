# 开发文档

> 适用范围：edu-schedule-system 后端模板（Go + Gin + MySQL/GORM schema-first）
> 更新时间：2026-01-30

## 1. 快速开始

### 1.1 环境要求
- Go 1.24.x（见 `go.mod` / `toolchain go1.24.11`）。项目按 Go minor 版本管理，补丁级 toolchain 可升级；不要在功能开发中降级 Go 或擅自跨 minor 升级。
- MySQL、Redis（本地可通过 `REDIS_ENABLED=false` 关闭 Redis 依赖）

### 1.2 配置加载规则
- 默认按 `-env` 选择：`configs/{env}_configs.toml`（env 取值：`dev|fat|uat|pro`）
- 可通过 `-config` 或环境变量 `CONFIG_PATH` 指定配置路径
- 未传 `-env` 时默认使用 `fat`
- `jwt.secret` 必填；仓库内 TOML 默认留空，开发时用 `JWT_SECRET=dev-secret` 或本地未提交配置覆盖。
- 本仓库默认 MySQL db/user 为 `app`；Makefile 本地命令默认使用 MySQL 密码 `app`，可通过 `MYSQL_READ_PASS=` 和 `MYSQL_WRITE_PASS=` 覆盖为空。
- `auth.mode` 支持 `auto|required|disabled`；`auto` 保持默认的仅生产环境鉴权。
- `auth.templateTokenEnabled` 控制模板 Token 端点，建议默认关闭；仅在明确的本地调试场景临时开启，可用 `AUTH_TEMPLATE_TOKEN_ENABLED=true|false` 覆盖。未显式开启时，`POST /api/v1/auth/token` 不注册，即使在 `dev` 也默认关闭。
- `auth.recoveryDelivery.mode` 控制找回密码验证码发送通道，支持 `placeholder|disabled|email|sms`；
  - `email` 需提供 SMTP host/port/username/password/from 配置
  - `sms` 当前为 stub/provider 抽象预留模式，可通过 `AUTH_RECOVERY_DELIVERY_SMS_PROVIDER`、`AUTH_RECOVERY_DELIVERY_SMS_SIGN_NAME` 配置 provider 元信息
- `redis.enabled=false` 时使用 no-op cache，不连接 Redis，readiness 中 Redis 仍为 OK。

### 1.3 启动与调试
```bash
make local-run
```
- Swagger（需 `SWAGGER_ENABLED=true` 且非 `pro`）：`/swagger/index.html`
- 健康检查：`/system/health`
- 就绪检查：`/system/ready`

## 2. 架构与目录职责

分层：`API(Handler) -> Service -> Repository`

```
cmd/                 入口 & 代码生成
configs/             配置文件（toml）
internal/api/        HTTP Handler & Router
internal/service/    业务逻辑
internal/repository/ 数据访问（MySQL/Redis/Mongo）
internal/pkg/        基础设施（logger/trace/errors/...)
internal/code/       业务错误码
docs/                Swagger 输出
```

## 3. 新项目采用与业务开发流程

如果是从模板创建新项目，先完成 `docs/TEMPLATE_ADOPTION.md` 中的初始化 Checklist：模块名、项目名、JWT、MySQL、Redis、CORS、Auth、Demo 模块和模板 Token 都应在业务开发前明确。不要把本模板包装成通用框架；默认开发路径是 MySQL migration -> live schema -> `gormgen` -> Service -> generated Handler/Router。

## 4. 业务开发流程（从 0 到 1）

可复制的 `task` 模块完整示例见 `docs/MODULE_EXAMPLE.md`。
逐项开发检查见 `docs/MODULE_CHECKLIST.md`。
迁移治理见 `docs/MIGRATIONS.md`，认证/权限接入见 `docs/AUTHORIZATION.md`，可观测性约定见 `docs/OBSERVABILITY.md`。

### 4.1 新增数据表 & DAO/Model 生成
1) 先通过 migration 创建或变更 MySQL 表，并应用到 live schema。
2) 从 live MySQL schema 生成 model/dao：
```bash
go run cmd/gormgen/main.go -dsn "user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local" -tables "your_table"
```
生成位置：
- `internal/repository/mysql/model/*.gen.go`
- `internal/repository/mysql/dao/*.gen.go`

### 4.2 DTO 定义
在 `internal/service/dto/` 定义请求/响应结构体：
- 请求结构体添加 `json`/`form`/`uri` tag
- 校验使用 `binding`/`validate` tag（由 Gin validator 触发）

### 4.3 Service 实现
在 `internal/service/{domain}/service.go`：
- **先定义接口，再实现**
- 方法签名统一 `func(ctx context.Context, ...)`
- 读写分离：`dao.Use(db.GetDbR())` / `dao.Use(db.GetDbW())`
- Service 负责业务规则：必填字段、正 ID、字段白名单、状态流转、幂等和存在性检查都应在这里处理，并配套单元测试。
- 事务可用：
  ```go
  writeDB := dao.Use(db.GetDbW())
  err := writeDB.Transaction(func(tx *dao.Query) error {
      // 使用 tx.WithContext(ctx) 做存在性检查和写入
      return nil
  })
  ```

### 4.4 Handler/Router 生成与补充
使用生成器快速创建 Handler + Router + Swagger 注释：
```bash
go run cmd/handlergen/main.go -table "your_table"
# 或使用 Makefile
make gen-handler TABLE=your_table
```
生成位置：
- `internal/api/{table}/*_handler.gen.go`
- `internal/api/{table}/*_routers.gen.go`

生成的 Handler 只做 HTTP 适配：参数绑定、正 ID 校验、`gorm.ErrRecordNotFound` 到 404 的映射、`core.BusinessError` 透传和统一响应。不要在生成文件中补业务逻辑；需要调整生成行为时修改 `cmd/handlergen/*_template.go.tpl` 后重新生成。

### 4.5 路由注册
在 `internal/router/router.go` 注册新路由（默认前缀 `/api/v1`）：
```go
apiV1Group := mux.Group("/api/v1")
securedGroup := apiV1Group
your.RegisterGeneratedYourRoutes(yourHandler, securedGroup)
```
如果调整路由前缀，请同步更新 Swagger 注释中的 `@Router` 路径。

### 4.6 依赖注入（Wire）
新增 Service/Handler 后，更新 `cmd/server/wire.go`，并重新生成：
```bash
wire ./cmd/server
# 或使用 Makefile
make wire
```
注意：`cmd/server/wire_gen.go` 为自动生成文件，不要手改。

### 4.7 Swagger 生成
```bash
scripts/swagger.bat
# 或
scripts/swagger.sh
```
输出：`docs/docs.go`、`docs/swagger.json`、`docs/swagger.yaml`

## 5. 开发规则（必须遵守）

### 5.1 分层与依赖方向
- Handler 只做参数解析、调用 Service、返回统一响应
- Service 承担业务规则、事务、聚合
- Repository 只做 CRUD，不写业务判断
- `internal/pkg` 不应依赖 `service` 或业务逻辑；基础设施例外需保持局部且有明确理由（如限流复用 Redis 客户端）

### 5.2 请求与响应
- 统一使用 `core.Context`（不要直接依赖 `gin.Context`）
- 参数解析：`ShouldBindJSON / ShouldBindQuery / ShouldBindURI`
- 成功返回：`ctx.Payload(data)`（统一 Response 包装）
- 错误返回：`ctx.AbortWithError(core.Error(httpCode, businessCode, message))`
- 不直接对外暴露内部错误栈；必要时仅记录日志

### 5.3 业务错误码
统一维护在 `internal/code/code.go`：
- 常量定义 + `zhCNText`/`enUSText` 映射
- 业务码需唯一，避免复用
- Handler 使用 `core.Error(httpCode, businessCode, message)` 返回错误；生成模板会透传 `core.BusinessError`，因此新增业务错误时应先补错误码和文案，再在服务边界返回可映射的错误。
- Service 中可预期失败优先返回 `internal/service/apperr`：`InvalidArgument` 映射 400，`NotFound` 映射 404，`Conflict` 映射 409，`Forbidden` 映射 403，`DependencyFailed` 映射 503。兼容历史 GORM 查询的 `gorm.ErrRecordNotFound` 仍由生成 Handler 映射为 404。

### 5.4 日志、Trace 与指标
- 使用 `ctx.RequestContext()` 透传 `Trace`/`Logger`
- 需要控制日志体量时：
  - `TRACE_LOG_BODY` 控制是否记录 body
  - `TRACE_BODY_MAX_BYTES` 控制 body 最大记录字节数
- 指标采集：
  - `METRICS_ENABLED` 开启 `/metrics`
  - 动态路由建议加 `core.AliasForRecordMetrics("/path/:id")`

### 5.5 认证与授权
- 统一在 Router 层挂载中间件（如 `interceptor.JWTokenAuthVerify`）
- `AUTH_MODE=auto|required|disabled` 可覆盖默认认证策略；默认策略为仅 `pro` 环境鉴权。
- `AUTH_TEMPLATE_TOKEN_ENABLED` 只用于本地模板 Token 端点，建议默认关闭；正式业务应使用真实登录/刷新流程，并按 `docs/TEMPLATE_ADOPTION.md` 移除或隔离模板 Token。未显式设置为 `true` 时不应依赖该端点。
- `AUTH_RECOVERY_DELIVERY_MODE` 控制恢复验证码发送实现：
  - `placeholder`：占位审计 sender
  - `disabled`：关闭恢复发送
  - `email`：启用 SMTP 邮件发送
- `AUTH_RECOVERY_DELIVERY_EMAIL_HOST`、`AUTH_RECOVERY_DELIVERY_EMAIL_PORT`、`AUTH_RECOVERY_DELIVERY_EMAIL_USERNAME`、`AUTH_RECOVERY_DELIVERY_EMAIL_PASSWORD`、`AUTH_RECOVERY_DELIVERY_EMAIL_FROM` 仅在 `AUTH_RECOVERY_DELIVERY_MODE=email` 时必填。
- `AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED` 仅用于 `dev` 下的本地调试预览，不得替代正式邮件/短信通道。
- 认证成功后通过 `ctx.SessionUserInfo()` 读取用户信息

### 5.6 数据访问约束
- 使用生成的 `dao`，避免随意写 `db.Raw()`
- 禁止把 `*gorm.DB` 暴露给 Service 之外的层

### 5.7 生成文件不可编辑
以下文件为自动生成，**请勿手动修改**：
- `internal/repository/mysql/model/*.gen.go`
- `internal/repository/mysql/dao/*.gen.go`
- `internal/api/*/*_handler.gen.go`
- `internal/api/*/*_routers.gen.go`
- `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`
- `cmd/server/wire_gen.go`

### 5.8 代码格式与质量
- 格式化：`go fmt ./...` 或 `go run cmd/mfmt/main.go`
- 测试：`go test ./...`
- 本地 CI 检查：`make ci`（执行 `go vet ./...`、`go test ./... -cover`、`go build ./...`）
- 建议本地开启 `golangci-lint`（配置见 `.golangci.yml`）

## 6. 常用命令速查
```bash
# 运行
make local-run

# 生成 Handler/Router
make gen-handler TABLE=your_table

# 生成 DAO/Model
make gen-dao DSN="user:pass@tcp(127.0.0.1:3306)/db" TABLES="your_table"

# 生成模块适配层，并在 DSN 存在时先生成 DAO/Model
make gen-module TABLE=your_table DSN="user:pass@tcp(127.0.0.1:3306)/db" TABLES="your_table"

# Swagger
scripts/swagger.bat
# 或 scripts/swagger.sh

# 格式化 / 测试
go fmt ./...
go test ./...
make ci
```
