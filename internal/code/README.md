## 错误码规则

- 错误码需在 `code` 包中进行定义。

#### 错误码为 5 位数

| 1 | 01 | 01 |
| :------ | :------ | :------ |
| 服务级错误码 | 模块级错误码 | 具体错误码 |

- 服务级错误码：1 位数进行表示，比如 1 为系统级错误；2 为普通错误，通常是由用户非法操作引起。
- 模块级错误码：2 位数进行表示，比如 01 为用户模块；02 为订单模块。
- 具体的错误码：2 位数进行表示，比如 01 为手机号不合法；02 为验证码输入错误。

#### 使用约定

- 新增错误码时同步维护 `zhCNText` 和 `enUSText`，并确认业务码全局唯一。
- Handler 通过 `core.Error(httpCode, businessCode, message)` 返回统一错误响应；不要直接暴露内部错误栈。
- Service 需要表达可预期业务失败时，优先返回 `internal/service/apperr` 中的服务错误；生成的 Handler 模板会识别 `gorm.ErrRecordNotFound`，并将 `apperr.Error` 映射为对应 HTTP 状态和业务码。Service 不应导入 `internal/pkg/core`。
- 参数绑定、非法 ID 等请求格式问题使用 `ParamBindError`；数据不存在使用 `RecordNotFound`；资源冲突使用 `Conflict`；权限不足使用 `Forbidden`；外部依赖不可用使用 `DependencyFailed`；未分类系统异常使用 `ServerError`。
- Service 层标准错误类型映射：`InvalidArgument -> 400/ParamBindError`，`NotFound -> 404/RecordNotFound`，`Conflict -> 409/Conflict`，`Forbidden -> 403/Forbidden`，`DependencyFailed -> 503/DependencyFailed`。
