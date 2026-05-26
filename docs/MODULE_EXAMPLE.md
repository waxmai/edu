# 新模块示例：task

本文用 `task` 表演示从建表到测试的完整流程。示例遵守当前项目边界：HTTP 适配层由 `handlergen` 生成，业务规则写在 `internal/service/task`，数据库访问使用 `gormgen` 从 live MySQL schema 生成的 DAO/Model，生成文件不手改。仓库内置的 `admin` 是 demo/reference module，可参考但不代表正式账号或权限系统。实际开发可同步参考 `docs/MODULE_CHECKLIST.md` 逐项确认。

## 1. 建表与迁移

新增迁移文件，例如 `migrations/000002_create_task.sql`：

```sql
-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
CREATE TABLE IF NOT EXISTS task (
  id INT NOT NULL AUTO_INCREMENT COMMENT '主键',
  title VARCHAR(128) NOT NULL COMMENT '标题',
  status VARCHAR(32) NOT NULL DEFAULT 'todo' COMMENT '状态',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务表';
```

本地有 MySQL 时执行迁移：

```bash
JWT_SECRET=dev-secret MYSQL_READ_PASS=app MYSQL_WRITE_PASS=app make migrate
```

## 2. 生成 DAO/Model

从 live MySQL schema 生成持久化代码：

```bash
make gen-dao \
  DSN='app:app@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local' \
  TABLES='task'
```

生成文件位于：

- `internal/repository/mysql/model/task.gen.go`
- `internal/repository/mysql/dao/task.gen.go`

这些文件由 `gormgen` 维护，不要手改。字段变化时改迁移和数据库 schema 后重新生成。

## 3. 定义 DTO

新增 `internal/service/dto/task.go`：

```go
package dto

type TaskCreateRequest struct {
    Title  string `json:"title" binding:"required"`
    Status string `json:"status"`
}

type TaskUpdateRequest map[string]interface{}

type TaskResponse struct {
    ID     int32  `json:"id"`
    Title  string `json:"title"`
    Status string `json:"status"`
}
```

DTO 只描述 API 边界。默认值、字段白名单、状态流转等业务规则放在 Service。响应 DTO 不直接引用 generated DB model。

## 4. 实现 Service

新增 `internal/service/task/service.go`，模式参考 `internal/service/admin/service.go`：

```go
package task

import (
    "context"
    "strings"

    "edu-schedule-system/internal/repository/mysql"
    "edu-schedule-system/internal/repository/mysql/dao"
    "edu-schedule-system/internal/repository/mysql/model"
    "edu-schedule-system/internal/service/apperr"
    "edu-schedule-system/internal/service/dto"
)

type Service interface {
    Create(ctx context.Context, req *dto.TaskCreateRequest) (int32, error)
    List(ctx context.Context) ([]dto.TaskResponse, error)
    GetByID(ctx context.Context, id int32) (*dto.TaskResponse, error)
    DeleteByID(ctx context.Context, id int32) (int64, error)
    UpdateByID(ctx context.Context, id int32, req dto.TaskUpdateRequest) (int64, error)
}

type service struct{ db mysql.Repo }

func New(db mysql.Repo) Service { return &service{db: db} }

func (s *service) Create(ctx context.Context, req *dto.TaskCreateRequest) (int32, error) {
    if req == nil {
        return 0, apperr.InvalidArgument("task create request is required")
    }
    title := strings.TrimSpace(req.Title)
    if title == "" {
        return 0, apperr.InvalidArgument("title is required")
    }
    status := strings.TrimSpace(req.Status)
    if status == "" {
        status = "todo"
    }

    item := &model.Task{Title: title, Status: status}
    writeDB := dao.Use(s.db.GetDbW())
    if err := writeDB.WithContext(ctx).Task.Create(item); err != nil {
        return 0, err
    }
    return item.ID, nil
}
```

继续按 admin 模块补齐 `List/GetByID/DeleteByID/UpdateByID`。更新接口必须做字段白名单，避免客户端覆盖未授权字段。
读接口从 generated model 查询后映射为响应 DTO 再返回，避免数据库 schema 直接泄漏到 API。

## 5. 生成 Handler/Router

Service 和 DTO 就绪后生成 HTTP 适配层：

```bash
make gen-handler TABLE=task
```

生成文件位于 `internal/api/task/`，只负责绑定参数、调用 Service、统一响应和错误映射。

## 6. 注册 Router 和 Wire

在 `internal/router/router.go` 注入并挂载新 handler：

```go
func NewHTTPMux(logger *zap.Logger, db mysql.Repo, cache redis.Repo, adminHandler *admin.Handler, taskHandler *task.Handler) (core.Mux, error) {
    // ...
    admin.RegisterGeneratedAdminRoutes(adminHandler, securedGroup)
    task.RegisterGeneratedTaskRoutes(taskHandler, securedGroup)
    return mux, nil
}
```

在 `cmd/server/wire.go` 增加 provider：

```go
import (
    taskAPI "edu-schedule-system/internal/api/task"
    taskService "edu-schedule-system/internal/service/task"
)

wire.Build(
    // ...
    taskAPI.New,
    taskService.New,
)
```

然后重新生成 Wire：

```bash
make wire
```

也可以用组合入口执行 DAO/Model 生成、handler 生成和 Wire：

```bash
make gen-module TABLE=task
```

如果传入 `DSN`，`gen-module` 会先生成 DAO/Model；不传 `DSN` 时只生成 Handler 并运行 Wire，适合 schema 已经生成过的场景。注意：`gen-module` 不会自动修改 `internal/router/router.go` 或 `cmd/server/wire.go`，新增模块仍需先按上文手动注册路由和 provider，否则 Wire 或编译会失败。

## 7. 测试建议

Service 测试使用 `sqlmock`，避免依赖 live MySQL。参考 `internal/service/admin/service_test.go`：

- 校验必填字段、正 ID、字段白名单等业务规则。
- 使用 `sqlmock` 断言关键 SQL、事务和 rows affected。
- 对 `gorm.ErrRecordNotFound` 保持原样返回，或返回 `apperr.NotFound(...)`，生成 Handler 都会映射为 404。

Router/Handler 集成测试使用 fake service，不连 MySQL/Redis。参考 `internal/router/admin_integration_test.go`：

- 通过 `NewHTTPMux` 真实挂载路由。
- 使用 `httptest` 请求 `/api/v1/{module}`。
- 断言 200、400、404 等 HTTP 语义。
- 子进程隔离全局配置，避免不同 env/config 测试互相污染。

## 8. 提交前检查

本地和 CI 使用同一入口：

```bash
make ci
```

需要查看 Swagger 或 pprof 时本地启动：

```bash
JWT_SECRET=dev-secret SWAGGER_ENABLED=true PPROF_ENABLED=true METRICS_ENABLED=true make local-run
```
