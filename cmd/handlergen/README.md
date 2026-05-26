## 生成 Handler、Routers 和 Swagger 接口文档工具

### 概述

本工具旨在自动化生成 Handler、Routers 和 Swagger 接口文档，简化开发流程，确保代码的一致性和可维护性。生成的 Handler 只负责 HTTP 适配，会调用 `internal/service/{table}` 中的业务服务，不会直接访问 DAO。

### 使用方法

要查看工具的详细使用说明，可以执行以下命令：

```shell
go run cmd/handlergen/main.go -h
```

这将显示工具的用法和可用选项。

#### 选项说明

- `-table`：此选项用于指定需要生成代码的数据库表名。

### 示例用法

以下是如何使用该工具生成名为 `admin` 的表的 Handler、Routers 和 Swagger 接口文档的示例。`admin` 是仓库内置 demo/reference module，真实项目请替换为自己的业务表名：

```shell
# 在项目根目录下执行
go run cmd/handlergen/main.go -table "admin"
```

这行命令会根据指定的表生成相应的 Handler、Routers 和 Swagger 文档，帮助你快速启动和维护项目中的 API 开发。

### 注意事项

- 确保在执行命令时，你位于项目的根目录下。
- 替换 `admin` 为你实际需要生成文档的表名；不要把 demo module 当作正式账号或权限系统。
- 该工具假设你的 DTO、Service、Wire 和 Router 注册会按模板边界补齐。
- 生成代码依赖对应的服务和 DTO 已存在，例如 `internal/service/admin.Service`、`dto.AdminCreateRequest` 和 `dto.AdminUpdateRequest`。新增表时请先补齐或同步生成 service/dto，再在 `cmd/server/wire.go` 注册 service 与 handler provider，并在 `internal/router/router.go` 挂载路由。
- 生成文件属于边界适配层，不要手改 `internal/api/{table}/*_handler.gen.go` 或 `*_routers.gen.go`；需要改变通用行为时修改本目录模板后重新生成。
- 当前模板会校验 path ID 为正整数，使用 `errors.Is` 识别 `gorm.ErrRecordNotFound`，将服务返回的 `apperr.Error` 映射为 4xx，并且写操作响应只暴露 `rows_affected`。
- 业务校验、字段白名单、事务和服务错误类型应在 `internal/service/{table}` 中实现并测试，Handler 只负责 HTTP 绑定和错误码映射。
