# 认证与权限接入指南

当前项目已完成面向公开运营前的认证与权限升级，包含：

- 正式账号密码登录
- 首次登录强制改密
- access token + refresh token
- 设备级会话治理
- 会话列表、单设备下线、全端强退
- 异常会话识别
- 找回密码与二次验证
- 恢复审计
- 细粒度菜单权限、动作权限、数据范围权限

当前恢复链路说明：
- 生产口径下验证码必须走正式短信/邮件通道
- 默认不会通过接口直接回传验证码
- 当前后端已抽象恢复发送通道接口 `RecoveryDeliverySender`
- 当前支持的恢复发送模式：
  - `placeholder`：占位审计 sender，适合受控开发/联调
  - `disabled`：关闭恢复发送能力
  - `email`：启用 SMTP 邮件发送
- 启用 `email` 时需要配置 SMTP host/port/username/password/from
- 当前 SMTP sender 已具备基础 timeout 控制与成功/失败审计字段
- 仅在 `dev` 且显式设置 `AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED=true` 时，才允许本地调试预览模式留痕启用

## 当前认证行为

- `AUTH_MODE=required`：发布默认值，所有业务接口必须鉴权
- `AUTH_MODE=disabled`：仅本地开发或特殊测试使用
- `AUTH_TEMPLATE_TOKEN_ENABLED=false`：共享环境与生产环境保持关闭

## 已落地正式接口

### 登录
`POST /api/v1/auth/login`

返回：
- access token
- refresh token
- 当前用户信息
- 当前登录会话信息

### 刷新令牌
`POST /api/v1/auth/refresh`

已支持：
- refresh token rotation
- 会话级刷新跟踪
- 异常会话风险标记

### 当前用户
`GET /api/v1/auth/me`

当前返回包含：
- `permissions`
- `menuPermissions`
- `dataScope`

### 修改密码
`POST /api/v1/auth/change-password`

密码要求：
- 至少 12 位
- 必须同时包含大写字母、小写字母、数字、特殊字符

修改成功后：
- 清除必须改密标记
- 更新密码修改时间
- 收敛历史风险状态

### 退出登录
`POST /api/v1/auth/logout`

### 会话列表
`GET /api/v1/auth/sessions`

### 单设备下线
`DELETE /api/v1/auth/sessions/{id}`

### 全端强退
`POST /api/v1/auth/logout-all`

### 发起密码恢复
`POST /api/v1/auth/password-recovery/start`

支持：
- email 主验证
- sms 主验证
- email / sms 二次验证
- 风险标记

### 通过恢复链路重置密码
`POST /api/v1/auth/password-recovery/reset`

恢复成功后：
- 新密码需满足至少 12 位，且包含大写字母、小写字母、数字、特殊字符
- 重置密码
- 注销全部活跃会话
- 写恢复审计

## 细粒度权限模型

### 菜单权限
例如：
- `menu:dashboard:view`
- `menu:student:view`
- `menu:payment:view`
- `menu:system:user:view`

### 动作权限
例如：
- `student:list`
- `student:create`
- `schedule:reschedule`
- `payment:update`
- `system:user:reset-password`

### 数据范围权限
- `all`
- `self`

当前策略：
- `admin` 默认 `all`
- `teacher` 默认 `self`

## 前端接入约定

前端现已按以下方式接入：

- 登录后本地持久化 access token / refresh token
- 应用启动时调用 `/auth/me` 恢复当前用户
- 路由按 `permission` / `menuPermission` 做访问控制
- 左侧菜单按 `menuPermissions` 动态展示
- 特定 SaaS 能力按 `featureFlags` 做可见性收口，例如订阅中心、恢复治理、审计导出
- 会话治理与密码恢复接口已可供前端继续扩展 UI

## 当前新增的 SaaS 功能开关口径

当前已约定的 feature flags 包括：

- `subscription_center`：订阅中心与续费相关能力
- `audit_export`：审计导出与平台审计查看能力
- `recovery_ops`：恢复失败治理与平台恢复排障能力

当前租户设置中心第一版已覆盖：

- `timezone`
- `brandName`
- `notificationEmail`
- `securityPolicy`
- `remark`

建议后续新增设置项时，优先遵循：
- 与订阅/套餐相关的能力走 `featureFlags`
- 与运营配置相关的静态项走 tenant settings
- 与权限相关的控制继续走 permission / menuPermission

## 恢复与安全建议

公开运营前建议继续保持：

- 正式环境关闭模板 token
- 恢复验证码仅通过正式通道发送
- `pro` 环境不得使用 `placeholder` 恢复发送模式
- 恢复预览仅允许 `dev` 环境在显式设置 `AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED=true` 时本地调试使用
- 风险登录与恢复事件接入外部告警平台
- 周期性审计异常会话与恢复记录
