# edu-schedule-system 多租户商用功能实施清单

## 目标

将当前“受控试运行版”升级为可逐步承接正式收费 SaaS 的多租户商用版本。

本文按两个阶段拆解：

1. **第一阶段：受控商用前，2-4 周**
   目标是补齐租户隔离、安全、通知、备份、告警、CI 前置检查，确保少量真实客户可控上线。

2. **第二阶段：正式收费 SaaS，4-8 周**
   目标是补齐套餐、订单、续费、租户生命周期、客户成功、批量数据与经营报表，形成收费闭环。

---

# 第一阶段：受控商用前，2-4 周

## 阶段目标

上线口径：

- 支持少量真实客户受控商用
- 平台人工开通租户
- 平台人工维护套餐与订阅状态
- 有基础告警和值守能力
- 有自动备份与恢复演练
- 有租户隔离测试兜底
- 有生产安全配置检查拦截

不做或暂缓：

- 暂不开放公网自助注册
- 暂不做完整在线支付闭环
- 暂不做复杂自定义角色体系
- 暂不承诺大规模高可用 SLA

---

## 1. 租户隔离专项测试与 Scope 统一治理

### 1.1 统一租户 Scope 能力

**目标**
避免每个 Service 手写租户过滤条件导致遗漏，形成统一的数据访问边界。

**功能清单**

- [x] 新增统一 Scope 工具包，例如 `internal/service/serviceutil/scope.go`
  - 已实现：`EnsureSameOrganization`、`EnsureSameCampus`、`EnsureTenantAccess` 以及 Student/Course/LessonPackage/PaymentRecord/Schedule/LessonRecord/RescheduleRecord 专用访问守卫。
- [ ] 定义租户 Scope 输入结构：
  - actor
  - organizationId
  - campusId
  - dataScope
  - allowPlatformAll
- [x] 提供通用方法：
  - `ApplyOrganizationScope(db, actor)`
  - `ApplyCampusScope(db, actor)`
  - `EnsureSameOrganization(ctx, actor, organizationID)`
  - `EnsureSameCampus(ctx, actor, campusID)`
  - `CanAccessOrganization(actor, organizationID)`
  - `CanAccessCampus(actor, campusID)`
  - 当前阶段先落地 `Ensure*` 写入/详情守卫；列表仍沿用既有 `ScopeStudents/ScopeCourses/...` Gen 条件封装，下一步再统一命名到 `serviceutil`。
- [ ] 所有核心业务列表统一接入 Scope：
  - 学员
  - 课程
  - 课包
  - 缴费
  - 排课
  - 上课记录
  - 调补课
  - 系统用户
  - 校区
- [x] 所有详情接口统一接入 Scope
  - 读详情：现有列表/详情查询已通过 `ScopeStudents/ScopeCourses/ScopeLessonPackages/ScopePaymentRecords/ScopeSchedules/ScopeLessonRecords/ScopeRescheduleRecords` 过滤。
- [x] 所有更新 / 删除 / 状态变更接口统一接入 Scope
  - 已覆盖：student、course、lesson_package、payment_record、schedule、lesson_record、reschedule_record 的更新/删除/状态变更关键入口。
- [ ] 平台管理员访问业务数据时保留显式 bypass 标记，避免误认为普通租户访问

**建议接口/代码改造点**

- `internal/service/student`
- `internal/service/course`
- `internal/service/lesson_package`
- `internal/service/payment_record`
- `internal/service/schedule`
- `internal/service/lesson_record`
- `internal/service/reschedule_record`
- `internal/service/user`
- `internal/service/tenant`

**验收标准**

- [ ] org A 用户无法查看 org B 的任何业务列表数据
- [ ] org A 用户无法通过猜 ID 查看 org B 的详情
- [x] org A 用户无法更新 / 删除 org B 数据
- [x] campus A 用户无法访问 campus B 数据，除非角色具备机构级权限
- [x] teacher 只能访问自己数据范围内的排课、上课记录等数据
- [ ] platform_admin 可以访问全量，但关键操作有审计

---

### 1.2 租户隔离专项自动化测试

**目标**
把“跨租户不可访问”固化成测试，避免后续迭代回归。

**功能清单**

- [x] 新增测试数据构造器：
  - org A
  - org B
  - campus A1
  - campus B1
  - org admin A
  - org admin B
  - campus admin A1
  - teacher A1
  - teacher B1
  - 已在 `internal/service/serviceutil/tenant_scope_integration_test.go` 中构造核心角色/资源矩阵；后续接真实 DB/API fixture 时复用该命名。
- [ ] 为每个核心模块补跨租户测试：
  - list
  - get
  - create
  - update
  - delete/cancel/status
- [x] 增加“猜 ID 攻击”测试：
  - 使用 org A token 请求 org B 的资源 ID
  - 预期 403 或 404，推荐统一 403/404 策略
  - 当前覆盖 serviceutil 守卫层的跨组织/跨校区/教师 self scope 猜 ID 访问。
- [ ] 增加“平台管理员可访问但审计留痕”测试
- [x] 增加“校区管理员不可越校区”测试
- [x] 增加“教师 self scope”测试

**建议测试文件**

- `internal/router/tenant_scope_integration_test.go`
- `internal/router/student_scope_integration_test.go`
- `internal/router/course_scope_integration_test.go`
- `internal/router/lesson_package_scope_integration_test.go`
- `internal/router/payment_scope_integration_test.go`
- `internal/router/schedule_scope_integration_test.go`
- `internal/router/user_scope_integration_test.go`

**验收标准**

- [ ] `go test ./internal/router/...` 通过
- [x] 至少覆盖 6 个核心业务模块的跨租户访问测试
- [ ] 每个模块至少覆盖 list/get/update 三类边界
- [x] 测试中明确包含 org A / org B 双租户数据

---

### 1.3 数据库唯一索引租户化复核

**目标**
避免多租户下唯一键冲突或业务唯一性不清晰。

**功能清单**

- [ ] 复核 `student` 唯一字段：手机号、姓名等是否应机构内唯一
- [ ] 复核 `course` 是否机构内课程名唯一
- [ ] 复核 `sys_user.username` 是否全平台唯一还是机构内唯一
- [ ] 复核 `campus_code` 是否机构内唯一
- [ ] 复核排课冲突唯一性是否包含：
  - organization_id
  - campus_id
  - teacher_id
  - class_date
  - start_time
  - end_time
- [ ] 输出索引调整 migration
- [ ] 补充索引变更说明文档

**验收标准**

- [ ] 不同机构允许存在业务上合理重复的数据
- [ ] 同一机构内仍能约束应唯一的数据
- [ ] migration 可在已有数据上平滑执行

---

## 2. 权限点再拆细，至少平台运营权限拆开

### 2.1 平台运营权限拆分

**目标**
避免平台管理员权限过粗，为后续平台运营、客服、财务、运维分权打基础。

**新增权限点建议**

租户管理：

- [x] `platform:tenant:list`
- [x] `platform:tenant:get`
- [x] `platform:tenant:create`
- [x] `platform:tenant:update`
- [x] `platform:tenant:disable`
- [x] `platform:tenant:restore`

校区管理：

- [x] `platform:campus:list`
- [x] `platform:campus:create`
- [x] `platform:campus:update`
- [x] `platform:campus:disable`

订阅管理：

- [x] `platform:subscription:list`
- [x] `platform:subscription:get`
- [x] `platform:subscription:update`
- [x] `platform:subscription:adjust-quota`
- [x] `platform:subscription:extend-trial`
- [x] `platform:subscription:suspend`
- [x] `platform:subscription:resume`

客户成功：

- [x] `platform:cs:list`
- [x] `platform:cs:update-owner`
- [x] `platform:cs:add-followup`
- [x] `platform:cs:update-status`

审计与恢复：

- [x] `platform:audit:list`
- [x] `platform:audit:export`
- [x] `platform:recovery-alert:list`
- [x] `platform:recovery-alert:resolve`

平台报表：

- [x] `platform:report:view`
- [x] `platform:report:export`

**代码改造点**

- [x] 更新 `internal/proposal/permission.go`
- [x] 更新权限矩阵文档 `docs/PERMISSION_MATRIX.md`
- [x] 更新后端路由权限校验
- [x] 更新前端路由 meta 权限
- [x] 更新菜单显示权限
- [x] 更新按钮级权限控制

**验收标准**

- [x] 平台运营账号可以看租户但不能改套餐
- [x] 平台财务账号可以改订阅但不能导出审计
- [x] 平台审计账号可以查看 / 导出审计但不能改租户
- [x] 无权限访问接口返回 403
- [x] 无权限菜单前端不可见

---

### 2.2 平台预置角色

**目标**
先不做完整自定义角色，也能满足第一阶段平台内部分权。

**预置角色建议**

- [x] `platform_admin`：全权限
- [x] `platform_ops`：租户、校区、客户成功查看与跟进
- [x] `platform_finance`：订阅、订单、续费相关权限
- [x] `platform_auditor`：审计、恢复告警只读与导出
- [x] `platform_support`：客户排障、恢复告警处理、有限用户查看

**功能清单**

- [x] 扩展角色常量
- [x] 扩展 `BuildAccessProfile`
- [ ] bootstrap 增加可选平台运营账号
- [x] 前端按角色展示不同平台菜单
- [x] 接口补角色权限测试

**验收标准**

- [x] 5 类平台角色权限边界清晰
- [x] 测试覆盖不同平台角色访问差异

---

## 3. 正式短信 / 邮件通知能力闭环

### 3.1 通知通道抽象

**目标**
将密码恢复发送能力升级为通用通知中心雏形。

**功能清单**

- [x] 新增通知模块：`internal/service/notification`
- [x] 定义通用消息模型：
  - channel：email/sms
  - target
  - templateCode
  - variables
  - organizationID
  - scene
  - idempotencyKey
- [x] 定义 Sender 接口：
  - `Send(ctx, message) (*SendResult, error)`
- [x] 支持 Email Sender
- [x] 支持 SMS Sender
- [x] 支持 Disabled Sender
- [x] 支持 Placeholder Sender，仅 dev/fat 可用
- [x] 发送失败统一分类：
  - timeout
  - auth_failed
  - provider_rejected
  - rate_limited
  - invalid_target
  - unknown

**验收标准**

- [x] 密码恢复可以通过通用通知模块发送
- [x] 邮件 / 短信通道配置缺失时启动前置检查失败
- [x] pro 环境禁止 placeholder

---

### 3.2 邮件正式通道

**功能清单**

- [x] SMTP 配置项完整校验：
  - host
  - port
  - username
  - password
  - from
  - timeout
  - maxAttempts
- [x] 支持 TLS / STARTTLS
- [x] 支持发送超时
- [x] 支持重试
- [x] 支持目标邮箱脱敏记录
- [x] 支持邮件模板变量
- [x] 支持发送结果落库或落日志

**验收标准**

- [ ] UAT 能真实收到恢复邮件
- [x] SMTP 密码错误时有明确错误分类
- [x] SMTP 超时时有失败记录和告警

---

### 3.3 短信正式通道

**功能清单**

- [ ] 选择短信供应商，建议先支持一个主供应商
- [ ] 新增短信配置：
  - provider
  - accessKey
  - accessSecret
  - signName
  - templateCode
  - endpoint
  - timeout
- [ ] 实现正式 SMS Sender
- [x] 支持短信模板变量
- [x] 支持手机号脱敏
- [x] 支持供应商返回码归一化
- [ ] 支持发送频控：
  - 同手机号每分钟限制
  - 同手机号每天限制
  - 同 IP 每小时限制
- [ ] 支持失败重试，避免对不可重试错误反复重试

**验收标准**

- [ ] UAT 能真实收到短信验证码
- [ ] 同手机号频繁请求被限制
- [ ] 供应商异常时有失败告警

---

### 3.4 通知记录与告警

**功能清单**

- [x] 新增 `notification_record` 表
- [x] 已新增结构化发送记录模型 `notification.DeliveryRecord`，当前已支持正式表落库，日志 / 恢复审计仍作为补充链路
- [ ] 字段建议：
  - id
  - organization_id
  - channel
  - scene
  - template_code
  - target_masked
  - provider
  - status
  - error_code
  - error_message
  - request_id
  - attempts
  - sent_at
  - created_at
- [x] 平台侧提供通知记录查询接口
- [x] 平台侧提供失败通知筛选
- [x] 通知失败达到阈值后触发生产告警

**验收标准**

- [x] 平台可以查到密码恢复邮件 / 短信发送记录
- [x] 不落明文手机号、邮箱完整地址或验证码
- [x] 通知失败会触发告警

---

## 4. 自动备份、备份告警、恢复演练

### 4.1 自动备份任务

**目标**
从手动脚本升级为可执行、可监控、可告警的自动备份。

**功能清单**

- [x] 新增备份配置：
  - enabled
  - schedule
  - retentionDays
  - localDir
  - objectStorageEnabled
  - objectStorageBucket
  - encryptEnabled
  - encryptKey
- [x] 支持每日自动 MySQL 备份
- [x] 备份文件命名包含：
  - env
  - database
  - timestamp
  - git commit 可选
- [x] 支持备份压缩
- [x] 支持备份加密
- [x] 支持备份上传对象存储
- [x] 支持备份完成后校验文件大小
- [x] 支持备份结果记录：成功 / 失败 / 耗时 / 文件大小

**实现建议**

- 第一阶段可先使用宿主机 cron + `scripts/db_backup.sh`
- 同时新增 `backup_history` 或 ndjson 记录
- 后续再演进为平台内任务调度

**验收标准**

- [x] 每天自动生成备份文件
- [x] 备份失败会告警
- [x] 备份文件大小异常会告警
- [x] 至少保留最近 7-30 天备份

---

### 4.2 备份告警

**功能清单**

- [x] 备份失败告警
- [x] 备份超时告警
- [x] 备份文件过小告警
- [x] 连续 N 天无成功备份告警
- [x] 对象存储上传失败告警
- [x] 本地磁盘空间不足告警

**验收标准**

- [x] 人为配置错误时能收到告警
- [ ] 备份成功恢复后告警自动恢复或可关闭

---

### 4.3 恢复演练 SOP

**功能清单**

- [x] 新增恢复演练文档：`docs/DB_RESTORE_DRILL.md`
- [x] 明确恢复演练频率：至少每月一次
- [x] 明确恢复目标环境：独立演练库，禁止覆盖生产
- [x] 明确恢复步骤：
  - 下载备份
  - 解密
  - 解压
  - 恢复到演练库
  - 执行 migration check
  - 执行 smoke test
- [x] 明确恢复验收项：
  - 表数量
  - 核心数据量
  - 登录可用
  - 核心业务可查
- [x] 记录演练结果：
  - 演练时间
  - 备份文件
  - 恢复耗时
  - 是否成功
  - 问题记录

**验收标准**

- [ ] 至少完成一次 UAT 恢复演练
- [ ] 有演练记录
- [ ] 恢复后 smoke test 通过

---

## 5. 生产告警接飞书 / 企业微信

### 5.1 告警通道抽象

**目标**
将当前落盘告警升级为可主动通知值班人的生产告警。

**功能清单**

- [x] 新增告警通知接口：`AlertNotifier`
- [x] 支持 Feishu webhook
- [x] 支持企业微信 webhook
- [x] 支持 disabled
- [x] 支持 best-effort，不影响主业务请求
- [x] 支持告警限流与冷却时间
- [x] 支持告警分级：
  - info
  - warning
  - critical
- [x] 支持告警去重 key

**验收标准**

- [x] 手动触发测试告警可发送到飞书 / 企业微信群
- [x] 告警失败不影响业务接口
- [x] 相同告警不会刷屏

---

### 5.2 第一阶段告警规则

**必须接入的告警**

- [x] 服务不可用：`/system/ready` 连续失败
- [x] HTTP 5xx 比例超过阈值
- [x] p95 接口耗时超过阈值
- [x] MySQL 探活失败
- [x] Redis 探活失败
- [x] 登录失败次数异常升高
- [x] 密码恢复发送失败异常升高
- [x] 通知发送失败异常升高
- [x] 备份失败
- [x] 磁盘空间不足
- [x] panic / recover 事件

**验收标准**

- [x] 每类告警都有测试触发方式
- [x] 告警消息包含环境、服务名、时间、trace/request 信息、建议处理动作
- [x] 告警文案可读，不只是技术堆栈

---

## 6. 生产安全配置核查纳入 CI/CD

### 6.1 prod readiness check 强化

**目标**
生产配置错误直接阻断发布。

**已有基础**
项目已有 `scripts/prod_readiness_check.sh`，需要纳入 CI/CD 并扩展检查项。

**检查项清单**

认证安全：

- [x] `AUTH_MODE` 必须为 `required`
- [x] `AUTH_TEMPLATE_TOKEN_ENABLED` 必须为 `false`
- [x] `JWT_SECRET` 不为空、不短、不含默认值
- [x] refresh token TTL 合理
- [x] 密码策略符合生产要求

调试暴露：

- [x] `SWAGGER_ENABLED=false`
- [x] `PPROF_ENABLED=false`
- [x] debug log 不应在生产开启

跨域：

- [x] CORS 不能是 `*`
- [x] CORS 必须是明确 HTTPS 域名
- [x] 生产不允许 localhost origin

恢复与通知：

- [x] `AUTH_RECOVERY_DELIVERY_MODE` 不能是 placeholder
- [x] 邮件模式必须配置 SMTP 必填项
- [x] 短信模式必须配置供应商密钥
- [x] 不允许恢复验证码 preview

数据库与缓存：

- [x] MySQL 地址、用户、密码必须配置
- [x] Redis 生产必须按预期启用或明确豁免
- [x] 数据库密码不能是 app/root/123456 等弱口令

备份与告警：

- [x] 生产必须启用备份
- [x] 生产必须配置告警 webhook
- [x] 生产必须配置日志目录

**验收标准**

- [x] CI 中执行 `scripts/prod_readiness_check.sh configs/pro_configs.toml`
- [x] 任一高危配置不合规时 CI 失败
- [x] CI 输出明确指出失败原因

---

### 6.2 CI/CD Gate

**功能清单**

- [x] CI 必跑：`go test ./...`
- [x] CI 必跑：`go build ./...`
- [x] CI 必跑：`make generated-check`
- [x] CI 必跑：前端 `npm run build`
- [x] CI 必跑：生产配置检查
- [x] CI 建议跑：依赖漏洞检查
- [x] CI 建议跑：migration lint / smoke
- [x] 发布前需要人工确认备份成功

**验收标准**

- [x] 任一检查失败不可发布生产
- [x] 发布记录可追溯到 commit、构建产物、配置检查结果

---

# 第二阶段：正式收费 SaaS，4-8 周

## 阶段目标

上线口径：

- 支持正式套餐售卖
- 支持订单、续费、欠费、宽限期
- 支持租户完整生命周期
- 支持客户成功运营
- 支持批量导入导出
- 支持第一版经营报表

---

## 1. 套餐中心

### 1.1 套餐模型

**目标**
将当前写在代码或订阅记录里的 plan 能力，升级为可管理的套餐中心。

**新增表建议：`plan`**

字段建议：

- id
- plan_code
- plan_name
- status：active/inactive
- billing_cycle：monthly/yearly/custom
- price_amount
- currency
- max_campuses
- max_users
- max_students
- max_courses
- max_storage_mb
- feature_flags JSON
- trial_days
- description
- sort_order
- created_at
- updated_at

**功能清单**

- [ ] 平台套餐列表
- [ ] 平台套餐详情
- [ ] 创建套餐
- [ ] 编辑套餐
- [ ] 启用 / 停用套餐
- [ ] 套餐功能配置
- [ ] 套餐配额配置
- [ ] 套餐价格配置
- [ ] 套餐排序

**接口建议**

- [ ] `GET /api/v1/platform/plans`
- [ ] `GET /api/v1/platform/plans/{id}`
- [ ] `POST /api/v1/platform/plans`
- [ ] `PUT /api/v1/platform/plans/{id}`
- [ ] `PATCH /api/v1/platform/plans/{id}/status`

**验收标准**

- [ ] 新租户可选择套餐创建订阅
- [ ] 修改套餐不会直接破坏历史订阅，历史订阅应保留快照或明确同步策略
- [ ] 停用套餐后不可新购，但历史订阅仍可正常展示

---

### 1.2 套餐权益快照

**目标**
避免套餐后续调价或改权益影响历史订单和历史订阅解释。

**功能清单**

- [ ] subscription 中保存套餐快照：
  - plan_name_snapshot
  - price_snapshot
  - feature_flags_snapshot
  - quota_snapshot
- [ ] order 中保存套餐购买快照
- [ ] 套餐变更时不自动覆盖历史订单
- [ ] 支持手动同步订阅权益

**验收标准**

- [ ] 套餐改价后，历史订单金额不变
- [ ] 历史订阅可以解释当时购买的权益

---

## 2. 订单 / 续费 / 欠费宽限期

### 2.1 订单模型

**新增表建议：`billing_order`**

字段建议：

- id
- order_no
- organization_id
- plan_id
- plan_code_snapshot
- plan_name_snapshot
- order_type：new/renew/upgrade/downgrade/manual_adjust
- billing_cycle
- amount
- currency
- status：pending/paid/cancelled/refunded/closed
- pay_method：manual/bank_transfer/wechat/alipay/stripe/other
- paid_at
- starts_at
- ends_at
- created_by
- remark
- created_at
- updated_at

**功能清单**

- [ ] 平台订单列表
- [ ] 创建订单
- [ ] 确认收款
- [ ] 取消订单
- [ ] 退款标记
- [ ] 订单详情
- [ ] 按租户查看订单
- [ ] 订单导出

**接口建议**

- [ ] `GET /api/v1/platform/billing/orders`
- [ ] `POST /api/v1/platform/billing/orders`
- [ ] `GET /api/v1/platform/billing/orders/{id}`
- [ ] `POST /api/v1/platform/billing/orders/{id}/confirm-paid`
- [ ] `POST /api/v1/platform/billing/orders/{id}/cancel`
- [ ] `POST /api/v1/platform/billing/orders/{id}/refund`

**验收标准**

- [ ] 订单确认支付后自动更新 subscription
- [ ] 重复确认支付需要幂等
- [ ] 金额、周期、套餐快照不可丢失
- [ ] 所有财务动作写审计

---

### 2.2 续费逻辑

**功能清单**

- [ ] 支持按当前订阅续费
- [ ] 支持续费指定套餐
- [ ] 支持续费指定周期
- [ ] 支持续费后延长 ends_at
- [ ] 已过期订阅续费后恢复 active
- [ ] suspended 订阅需平台手动恢复或支付后自动恢复，策略明确
- [ ] 支持续费备注

**业务规则建议**

- 当前订阅未过期：新周期从当前 `ends_at` 后开始累计
- 当前订阅已过期：新周期从支付确认时间开始累计
- past_due 支付后恢复 active
- suspended 是否自动恢复由平台配置决定

**验收标准**

- [ ] 未过期续费不会损失剩余天数
- [ ] 过期续费后可以恢复关键写操作
- [ ] 续费订单、订阅变更、审计三者一致

---

### 2.3 欠费宽限期

**功能清单**

- [ ] subscription 增加字段：
  - grace_period_days
  - grace_ends_at
  - past_due_at
  - suspended_at
- [ ] 每日任务检查到期订阅
- [ ] 到期后进入 `past_due`
- [ ] 宽限期内允许部分或全部功能，策略可配置
- [ ] 宽限期结束后进入 `suspended` 或 `expired`
- [ ] 状态变化写审计
- [ ] 状态变化发通知

**建议策略**

- `active/trial`：全部功能可用
- `past_due`：允许读，限制部分关键写，提示续费
- `suspended`：只允许登录、查看账单、续费，不允许业务写操作
- `expired`：长期过期，平台可归档

**验收标准**

- [ ] 到期后自动转 past_due
- [ ] 宽限期结束自动转 suspended/expired
- [ ] 状态变化后业务写操作符合限制策略
- [ ] 前端有明确到期提示

---

## 3. 租户生命周期管理

### 3.1 租户开通向导

**功能清单**

- [ ] 创建机构
- [ ] 创建默认校区
- [ ] 选择套餐
- [ ] 创建订阅
- [ ] 创建机构管理员
- [ ] 生成初始密码或邀请链接
- [ ] 发送开通通知
- [ ] 初始化默认设置
- [ ] 初始化默认角色 / 权限
- [ ] 初始化演示数据可选

**接口建议**

- [ ] `POST /api/v1/platform/tenants/provision`

**验收标准**

- [ ] 一次提交可完整开通可登录租户
- [ ] 开通过程事务化，失败可回滚或可补偿
- [ ] 初始账号必须强制改密

---

### 3.2 租户状态流转

**状态建议**

- `trial`：试用中
- `active`：正常
- `past_due`：欠费宽限期
- `suspended`：暂停服务
- `inactive`：人工停用
- `archived`：归档

**功能清单**

- [ ] 租户停用
- [ ] 租户恢复
- [ ] 租户归档
- [ ] 租户状态变更原因
- [ ] 租户状态变更审计
- [ ] 状态变更通知
- [ ] 停用后登录 / 业务操作策略

**验收标准**

- [ ] inactive/suspended 租户不能进行关键业务写操作
- [ ] archived 租户默认不可登录
- [ ] platform_admin 可恢复租户
- [ ] 状态变化前端有清晰提示

---

### 3.3 租户数据导出与交付

**功能清单**

- [ ] 平台按租户导出基础数据
- [ ] 支持导出：
  - 学员
  - 课程
  - 教师
  - 课包
  - 缴费
  - 排课
  - 上课记录
- [ ] 导出任务异步化
- [ ] 导出文件过期清理
- [ ] 导出操作审计
- [ ] 敏感字段脱敏策略

**验收标准**

- [ ] 只能导出指定租户数据
- [ ] 大数据量导出不阻塞请求
- [ ] 导出行为可审计

---

## 4. 客户成功跟进工具

### 4.1 客户成功基础模型

**新增表建议：`customer_success_record`**

字段建议：

- id
- organization_id
- owner_user_id
- owner_name
- stage：trial/onboarding/active/risk/churned
- health_level：healthy/medium/high
- last_contact_at
- next_follow_up_at
- contact_method
- note
- created_by
- created_at
- updated_at

**功能清单**

- [ ] 客户列表显示负责人
- [ ] 分配负责人
- [ ] 修改客户阶段
- [ ] 添加跟进记录
- [ ] 设置下次跟进时间
- [ ] 标记风险客户
- [ ] 风险原因分类
- [ ] 客户备注

**接口建议**

- [ ] `GET /api/v1/platform/customer-success/customers`
- [ ] `POST /api/v1/platform/customer-success/customers/{orgId}/assign-owner`
- [ ] `POST /api/v1/platform/customer-success/customers/{orgId}/followups`
- [ ] `PATCH /api/v1/platform/customer-success/customers/{orgId}/stage`
- [ ] `PATCH /api/v1/platform/customer-success/customers/{orgId}/health-level`

**验收标准**

- [ ] 平台运营能看到自己负责的客户
- [ ] 高风险客户能筛选
- [ ] 逾期未跟进客户能筛选
- [ ] 跟进记录不可无痕删除，至少保留审计

---

### 4.2 客户健康度

**健康指标建议**

- [ ] 订阅状态
- [ ] 剩余天数
- [ ] 近 7/30 天登录次数
- [ ] 近 7/30 天排课数量
- [ ] 近 7/30 天课消数量
- [ ] 账号使用率
- [ ] 校区使用率
- [ ] 恢复失败次数
- [ ] 最近跟进时间
- [ ] 是否临近到期

**功能清单**

- [ ] 计算健康分
- [ ] 健康等级：healthy/medium/high
- [ ] 高风险原因解释
- [ ] 高风险客户列表
- [ ] 未分配负责人列表
- [ ] 近期到期客户列表

**验收标准**

- [ ] 客户健康等级可解释
- [ ] 运营人员可以按健康等级筛选
- [ ] 首页或续费台能看到高风险客户

---

## 5. 批量导入导出

### 5.1 批量导入框架

**目标**
统一处理导入任务，避免每个模块重复造轮子。

**新增表建议：`import_task`**

字段建议：

- id
- organization_id
- campus_id
- module
- file_name
- file_url/path
- status：pending/processing/success/partial_failed/failed
- total_count
- success_count
- failed_count
- error_file_url/path
- created_by
- created_at
- finished_at

**功能清单**

- [ ] 下载导入模板
- [ ] 上传 Excel/CSV
- [ ] 预校验
- [ ] 执行导入
- [ ] 错误行导出
- [ ] 导入任务列表
- [ ] 导入结果详情
- [ ] 导入操作审计

**验收标准**

- [ ] 错误数据不会导致整批全部失败，除非选择严格模式
- [ ] 错误原因能精确到行和字段
- [ ] 导入任务只能访问本租户数据

---

### 5.2 第一批导入模块

优先级建议：

1. 学员导入
2. 教师 / 用户导入
3. 课程导入
4. 课包导入

**学员导入字段**

- [ ] 学员姓名
- [ ] 联系电话
- [ ] 家长姓名
- [ ] 校区
- [ ] 备注

**教师 / 用户导入字段**

- [ ] 用户名
- [ ] 昵称
- [ ] 手机号
- [ ] 邮箱
- [ ] 角色
- [ ] 校区
- [ ] 初始密码策略

**课程导入字段**

- [ ] 课程名称
- [ ] 课程类型
- [ ] 单课时价格
- [ ] 状态
- [ ] 备注

**课包导入字段**

- [ ] 学员
- [ ] 课程
- [ ] 总课时
- [ ] 已付金额
- [ ] 总金额
- [ ] 有效期

**验收标准**

- [ ] 每个导入模板都有示例数据
- [ ] 必填项、格式错误、关联数据不存在均能明确提示
- [ ] 导入成功后数据归属正确 organization/campus

---

### 5.3 批量导出

**功能清单**

- [ ] 学员导出
- [ ] 教师 / 用户导出
- [ ] 课程导出
- [ ] 课包导出
- [ ] 缴费导出
- [ ] 排课导出
- [ ] 上课记录导出
- [ ] 导出任务异步化
- [ ] 导出文件过期清理
- [ ] 导出权限控制
- [ ] 导出审计

**验收标准**

- [ ] 导出数据不越租户
- [ ] teacher 不允许导出敏感财务数据
- [ ] 导出大列表不会导致接口超时

---

## 6. 商业报表第一版

### 6.1 机构经营看板

**目标用户**
机构管理员、校区管理员。

**指标清单**

收入类：

- [ ] 今日收款
- [ ] 本月收款
- [ ] 按支付方式统计
- [ ] 按课程统计收入
- [ ] 欠费学员数
- [ ] 欠费金额

课消类：

- [ ] 今日课消
- [ ] 本月课消
- [ ] 按教师统计课时
- [ ] 按课程统计课时
- [ ] 待上课数量
- [ ] 已完成课程数量

学员类：

- [ ] 新增学员数
- [ ] 活跃学员数
- [ ] 低课时学员数
- [ ] 即将到期课包数
- [ ] 停课 / 流失风险学员数

排课类：

- [ ] 今日排课
- [ ] 本周排课
- [ ] 请假次数
- [ ] 调课次数
- [ ] 补课次数

**接口建议**

- [ ] `GET /api/v1/reports/org/overview`
- [ ] `GET /api/v1/reports/org/revenue`
- [ ] `GET /api/v1/reports/org/lesson-consumption`
- [ ] `GET /api/v1/reports/org/student-risk`

**验收标准**

- [ ] org_admin 只能看本机构
- [ ] campus_admin 只能看本校区或授权范围
- [ ] teacher 只能看自己的课时相关数据
- [ ] 金额统计与缴费记录口径一致，只统计 paid

---

### 6.2 平台 SaaS 经营看板

**目标用户**
平台管理员、运营、财务。

**指标清单**

租户类：

- [ ] 总租户数
- [ ] 新增租户数
- [ ] 试用租户数
- [ ] 付费租户数
- [ ] 欠费租户数
- [ ] 暂停租户数
- [ ] 即将到期租户数

收入类：

- [ ] 本月确认收入
- [ ] 待收款订单数
- [ ] 待收款金额
- [ ] 续费金额
- [ ] 新购金额
- [ ] 退款金额

运营类：

- [ ] 未分配负责人客户数
- [ ] 高风险客户数
- [ ] 逾期未跟进客户数
- [ ] 近 7 天活跃租户数
- [ ] 近 30 天活跃租户数

套餐类：

- [ ] 各套餐租户分布
- [ ] 各套餐收入分布
- [ ] 配额逼近上限租户数

**接口建议**

- [ ] `GET /api/v1/platform/reports/overview`
- [ ] `GET /api/v1/platform/reports/revenue`
- [ ] `GET /api/v1/platform/reports/tenant-health`
- [ ] `GET /api/v1/platform/reports/plan-distribution`

**验收标准**

- [ ] 平台报表只允许平台相关权限访问
- [ ] 收入统计与订单 paid 状态一致
- [ ] 租户健康数据与客户成功列表一致

---

# 建议里程碑拆分

## 第一阶段里程碑

### M1：租户隔离与权限硬化

交付内容：

- Scope 工具
- 核心模块 Scope 改造
- 跨租户自动化测试
- 平台权限点拆分
- 平台预置角色

验收命令：

```bash
go test ./internal/router/...
go test ./internal/service/...
```

---

### M2：通知、备份、告警

交付内容：

- 通知中心雏形
- 邮件正式通道
- 短信正式通道
- 通知记录
- 自动备份
- 备份告警
- 飞书 / 企业微信告警

验收方式：

- UAT 实收邮件
- UAT 实收短信
- 手动触发告警到群
- 完成一次备份恢复演练

---

### M3：生产发布 Gate

交付内容：

- prod readiness check 扩展
- CI/CD 接入
- 生产发布检查清单

验收方式：

- 高危配置可阻断 CI
- 正确配置可通过 CI

---

## 第二阶段里程碑

### M4：套餐与订单

交付内容：

- 套餐中心
- 套餐权益快照
- 订单管理
- 确认收款
- 续费逻辑
- 欠费宽限期

验收方式：

- 创建租户选择套餐
- 创建续费订单并确认收款
- 订阅自动延长
- 到期自动进入 past_due/suspended

---

### M5：租户生命周期与客户成功

交付内容：

- 租户开通向导
- 租户停用 / 恢复 / 归档
- 客户成功负责人
- 跟进记录
- 健康度评分

验收方式：

- 一键开通可登录租户
- 停用租户后关键操作受限
- 高风险客户列表可用

---

### M6：批量数据与商业报表

交付内容：

- 导入任务框架
- 学员 / 用户 / 课程 / 课包导入
- 核心模块导出
- 机构经营看板
- 平台 SaaS 经营看板

验收方式：

- Excel 导入错误行可回传
- 导出不越租户
- 报表金额与业务流水一致

---

# 推荐实施顺序

1. Scope 统一治理
2. 跨租户自动化测试
3. 平台权限拆分
4. 通知中心与正式短信 / 邮件
5. 自动备份与恢复演练
6. 生产告警接入
7. CI/CD 生产配置 Gate
8. 套餐中心
9. 订单与续费
10. 欠费宽限期
11. 租户开通向导
12. 客户成功跟进
13. 批量导入导出
14. 商业报表第一版

---

# 第一阶段完成后的上线判断

满足以下条件后，可以进入受控商用：

- [ ] 核心模块跨租户测试通过
- [ ] 平台权限已拆分，不再只有超管一把梭
- [ ] 正式邮件 / 短信至少一种通道可用，密码恢复闭环
- [ ] 自动备份可用，并完成至少一次恢复演练
- [ ] 生产告警可发送到飞书 / 企业微信群
- [ ] 生产配置检查纳入 CI/CD
- [ ] Swagger / pprof / template token / placeholder recovery 在生产被拦截
- [ ] UAT 环境完整 smoke 通过

---

# 第二阶段完成后的上线判断

满足以下条件后，可以定义为正式收费 SaaS 第一版：

- [ ] 套餐可配置
- [ ] 订单可创建、确认收款、续费
- [ ] 欠费宽限期自动流转
- [ ] 租户可一键开通、停用、恢复、归档
- [ ] 客户成功跟进可运营
- [ ] 支持核心数据批量导入导出
- [ ] 机构经营看板可用
- [ ] 平台 SaaS 经营看板可用
- [ ] 财务、订阅、租户状态、业务限制四者口径一致

---

# 第三阶段：规模化 SaaS 运营与增长，8-12 周

## 阶段目标

上线口径：

- 支持更精细的计费、套餐变更和商业审计
- 支持平台级运营增长、线索转化和客户分层
- 支持企业级安全、审计、合规与数据保留
- 支持多环境、多租户运维自动化
- 支持更完整的开放 API / Webhook / 第三方集成
- 支持平台经营指标持续追踪和异常预警

---

## 1. 套餐升级 / 降级 / 变更计费

### 1.1 套餐变更模型

**目标**
支持租户在订阅周期内升级、降级、加购权益，并保留完整变更记录。

**新增表建议：`subscription_change`**

字段建议：

- id
- organization_id
- subscription_id
- change_no
- change_type：upgrade/downgrade/quota_addon/cycle_change/manual_adjust
- from_plan_id
- to_plan_id
- from_plan_snapshot JSON
- to_plan_snapshot JSON
- effective_mode：immediate/next_cycle/custom_date
- effective_at
- prorated_amount
- currency
- status：pending/confirmed/cancelled/applied/failed
- requested_by
- approved_by
- applied_at
- remark
- created_at
- updated_at

**功能清单**

- [ ] 创建套餐变更申请
- [ ] 计算升级补差价
- [ ] 支持下周期生效的降级
- [ ] 支持人工调整权益
- [ ] 支持变更审批
- [ ] 支持变更取消
- [ ] 支持变更执行失败重试
- [ ] 订阅详情展示变更历史

**接口建议**

- [ ] `GET /api/v1/platform/subscription-changes`
- [ ] `POST /api/v1/platform/subscription-changes`
- [ ] `GET /api/v1/platform/subscription-changes/{id}`
- [ ] `POST /api/v1/platform/subscription-changes/{id}/confirm`
- [ ] `POST /api/v1/platform/subscription-changes/{id}/cancel`
- [ ] `POST /api/v1/platform/subscription-changes/{id}/apply`

**代码改造点**

- [ ] 新增套餐变更 Service
- [ ] 订阅续费逻辑支持 pending change
- [ ] 权益校验读取当前生效权益，而不是只读 plan_code
- [ ] 增加金额计算工具，统一处理补差价、退款、折扣
- [ ] 增加套餐变更事务边界，避免订单已付但订阅未更新

**验收标准**

- [ ] 月付租户可从基础版升级到专业版，并立即扩容
- [ ] 年付租户升级可生成补差价订单
- [ ] 降级选择下周期生效，不影响当前周期权益
- [ ] 所有变更均有可追溯记录

---

### 1.2 加购包 / 资源包

**目标**
支持在基础套餐之外按需加购短信、存储、校区、账号、学员额度等资源。

**新增表建议：`addon_package`**

字段建议：

- id
- addon_code
- addon_name
- addon_type：sms/storage/campus/user/student/custom
- quota_key
- quota_value
- billing_mode：one_time/monthly/yearly
- price_amount
- currency
- status
- description
- created_at
- updated_at

**新增表建议：`tenant_addon`**

字段建议：

- id
- organization_id
- subscription_id
- addon_package_id
- addon_snapshot JSON
- quantity
- starts_at
- ends_at
- status：active/expired/cancelled
- source_order_id
- created_at
- updated_at

**功能清单**

- [ ] 平台配置加购包
- [ ] 租户购买加购包
- [ ] 订阅权益合并套餐与加购包
- [ ] 加购包到期自动失效
- [ ] 加购包使用量展示
- [ ] 支持手动赠送资源包

**验收标准**

- [ ] 租户购买短信包后短信额度增加
- [ ] 加购存储包后上传限制同步变更
- [ ] 加购包到期后权益自动回收
- [ ] 套餐权益和加购权益计算口径一致

---

## 2. 商业审计与财务对账

### 2.1 订单收款审计

**目标**
让订单、收款、订阅变更、发票/凭证之间可追溯，降低人工对账风险。

**新增表建议：`payment_record`**

字段建议：

- id
- payment_no
- order_id
- organization_id
- amount
- currency
- pay_method
- pay_channel_trade_no
- paid_at
- confirmed_by
- confirm_source：manual/callback/import
- receipt_file_id
- status：confirmed/reversed/invalid
- remark
- created_at
- updated_at

**功能清单**

- [ ] 收款确认记录独立保存
- [ ] 支持上传付款凭证
- [ ] 支持冲正 / 作废收款记录
- [ ] 支持按租户、订单、日期筛选收款
- [ ] 支持导出收款明细
- [ ] 订单金额、收款金额、订阅生效结果联动校验

**接口建议**

- [ ] `GET /api/v1/platform/payments`
- [ ] `POST /api/v1/platform/orders/{id}/payments`
- [ ] `POST /api/v1/platform/payments/{id}/reverse`
- [ ] `GET /api/v1/platform/payments/export`

**验收标准**

- [ ] 每一笔 paid 订单至少存在一条有效收款记录
- [ ] 作废收款后订单状态和订阅状态不出现矛盾
- [ ] 财务可按月导出收款明细

---

### 2.2 商业操作审计日志

**目标**
记录所有影响套餐、订单、订阅、租户状态、权限和数据导出的关键操作。

**新增表建议：`platform_audit_log`**

字段建议：

- id
- actor_user_id
- actor_name_snapshot
- actor_role_snapshot
- organization_id
- target_type
- target_id
- action
- before_snapshot JSON
- after_snapshot JSON
- ip
- user_agent
- request_id
- risk_level：low/medium/high/critical
- created_at

**功能清单**

- [ ] 套餐增删改审计
- [ ] 订单确认 / 作废审计
- [ ] 订阅延期 / 暂停 / 恢复审计
- [ ] 租户停用 / 归档 / 导出审计
- [ ] 平台权限变更审计
- [ ] 高风险操作二次确认
- [ ] 审计日志只读，不允许普通后台删除

**验收标准**

- [ ] 任意租户状态变更可追溯到操作者
- [ ] 任意套餐价格变更可查看前后差异
- [ ] 数据导出操作有完整记录

---

## 3. 试用、线索与转化漏斗

### 3.1 试用租户管理

**目标**
支持从线索创建试用租户，跟踪试用行为，并推动转付费。

**新增表建议：`trial_application`**

字段建议：

- id
- lead_name
- contact_name
- contact_phone
- contact_email
- organization_name
- expected_student_count
- source_channel
- status：new/contacted/trial_opened/converted/lost
- trial_organization_id
- assigned_to
- trial_starts_at
- trial_ends_at
- lost_reason
- created_at
- updated_at

**功能清单**

- [ ] 创建试用申请
- [ ] 分配销售 / 客户成功负责人
- [ ] 一键开通试用租户
- [ ] 试用期自动到期提醒
- [ ] 试用转正式订阅
- [ ] 试用失败原因记录
- [ ] 试用行为摘要：登录次数、导入数据量、排课次数、邀请账号数

**接口建议**

- [ ] `GET /api/v1/platform/trials`
- [ ] `POST /api/v1/platform/trials`
- [ ] `POST /api/v1/platform/trials/{id}/open-tenant`
- [ ] `POST /api/v1/platform/trials/{id}/convert`
- [ ] `POST /api/v1/platform/trials/{id}/mark-lost`

**验收标准**

- [ ] 可从试用申请一键开通租户
- [ ] 试用到期前自动提醒负责人
- [ ] 试用转付费后订阅、订单、租户状态一致

---

### 3.2 转化漏斗看板

**目标**
让平台能看到线索到试用、试用到付费、付费到续费的转化效率。

**功能清单**

- [ ] 线索数
- [ ] 已联系数
- [ ] 已开通试用数
- [ ] 试用活跃数
- [ ] 试用转付费数
- [ ] 首购金额
- [ ] 续费金额
- [ ] 流失租户数
- [ ] 按来源渠道统计转化率
- [ ] 按负责人统计转化率

**接口建议**

- [ ] `GET /api/v1/platform/reports/conversion-funnel`
- [ ] `GET /api/v1/platform/reports/trial-activity`

**验收标准**

- [ ] 可以按日期范围查看漏斗
- [ ] 可以定位转化低的渠道和负责人
- [ ] 漏斗口径与试用、订单、订阅数据一致

---

## 4. 企业级安全与合规

### 4.1 数据保留与删除策略

**目标**
对停用、归档、注销租户的数据保留周期和删除流程形成产品化能力。

**新增表建议：`tenant_data_retention_policy`**

字段建议：

- id
- organization_id
- policy_type：default/custom
- retention_days_after_suspend
- retention_days_after_archive
- allow_hard_delete
- hard_delete_approved_by
- created_at
- updated_at

**功能清单**

- [ ] 配置默认数据保留周期
- [ ] 租户归档后进入只读保留期
- [ ] 到期前通知平台负责人
- [ ] 支持人工审批硬删除
- [ ] 删除前生成最终导出包
- [ ] 删除后保留删除证明与审计日志

**验收标准**

- [ ] 归档租户不会被立即物理删除
- [ ] 硬删除必须经过审批
- [ ] 删除后业务表不可再查询该租户数据，审计记录仍可追踪

---

### 4.2 敏感操作二次验证

**目标**
降低平台误操作和账号被盗后的高危风险。

**高危操作范围**

- 停用 / 恢复 / 归档租户
- 确认大额收款
- 作废订单或收款
- 导出租户全量数据
- 修改套餐价格
- 修改平台角色权限
- 手动延长订阅

**功能清单**

- [ ] 高危操作弹窗确认
- [ ] 要求输入操作原因
- [ ] 可选短信 / 邮箱 / TOTP 二次验证
- [ ] 高危操作写入审计日志
- [ ] 连续失败触发告警

**验收标准**

- [ ] 未通过二次验证无法执行高危操作
- [ ] 操作原因和验证结果可审计
- [ ] 高危操作通知可发送给平台管理员

---

## 5. 开放 API 与 Webhook

### 5.1 租户开放 API Key

**目标**
为后续与机构自有系统、低代码平台、数据仓库对接预留标准能力。

**新增表建议：`tenant_api_key`**

字段建议：

- id
- organization_id
- key_name
- key_prefix
- key_hash
- scopes JSON
- status：active/revoked/expired
- last_used_at
- expires_at
- created_by
- revoked_by
- revoked_at
- created_at
- updated_at

**功能清单**

- [ ] 租户创建 API Key
- [ ] API Key 权限 Scope 配置
- [ ] API Key 过期时间配置
- [ ] API Key 吊销
- [ ] API 调用日志
- [ ] API 限流

**接口建议**

- [ ] `GET /api/v1/tenant/api-keys`
- [ ] `POST /api/v1/tenant/api-keys`
- [ ] `DELETE /api/v1/tenant/api-keys/{id}`
- [ ] `GET /api/v1/tenant/api-logs`

**验收标准**

- [ ] API Key 只展示一次明文
- [ ] Key 泄露后可立即吊销
- [ ] API 调用不允许越租户

---

### 5.2 Webhook 事件订阅

**目标**
允许租户或平台对关键业务事件进行外部通知。

**新增表建议：`webhook_endpoint`**

字段建议：

- id
- organization_id nullable
- endpoint_name
- url
- secret
- event_types JSON
- status：active/inactive
- last_success_at
- last_failed_at
- failure_count
- created_at
- updated_at

**新增表建议：`webhook_delivery`**

字段建议：

- id
- endpoint_id
- event_type
- payload JSON
- status：pending/success/failed
- http_status
- response_body
- retry_count
- next_retry_at
- delivered_at
- created_at

**首批事件建议**

- tenant.created
- tenant.suspended
- tenant.restored
- subscription.created
- subscription.renewed
- subscription.past_due
- order.paid
- student.created
- course.created
- schedule.created

**验收标准**

- [ ] Webhook 签名可校验
- [ ] 失败自动重试
- [ ] 可查看投递历史和失败原因
- [ ] 单个租户 Webhook 失败不影响主业务事务

---

## 6. 多租户运维自动化

### 6.1 租户级健康巡检

**目标**
定期发现租户数据异常、配置缺失、订阅异常和用量超限风险。

**巡检项建议**

- [ ] 租户状态与订阅状态是否一致
- [ ] 订阅到期但未进入 past_due 的异常
- [ ] 欠费超过宽限期但仍可操作的异常
- [ ] 套餐配额与实际用量是否超限
- [ ] 关键租户管理员是否为空
- [ ] 通知通道是否配置异常
- [ ] 最近 7 天是否存在大量失败任务
- [ ] 租户数据唯一索引冲突风险

**新增表建议：`tenant_health_check`**

字段建议：

- id
- organization_id
- check_batch_no
- health_status：ok/warning/critical
- issue_count
- critical_count
- result JSON
- checked_at

**接口建议**

- [ ] `GET /api/v1/platform/tenant-health-checks`
- [ ] `POST /api/v1/platform/tenant-health-checks/run`
- [ ] `GET /api/v1/platform/tenant-health-checks/{id}`

**验收标准**

- [ ] 可手动触发全量租户巡检
- [ ] 可查看异常租户列表
- [ ] critical 异常自动发送平台告警

---

### 6.2 后台任务中心

**目标**
统一管理导入、导出、备份、Webhook、通知、巡检等异步任务。

**新增表建议：`background_job`**

字段建议：

- id
- job_type
- organization_id nullable
- status：pending/running/success/failed/cancelled
- progress
- payload JSON
- result JSON
- error_message
- retry_count
- max_retry_count
- next_retry_at
- started_at
- finished_at
- created_by
- created_at
- updated_at

**功能清单**

- [ ] 后台任务列表
- [ ] 后台任务详情
- [ ] 任务进度展示
- [ ] 失败任务重试
- [ ] 可取消任务
- [ ] 任务结果下载
- [ ] 任务失败告警

**验收标准**

- [ ] 导入 / 导出 / Webhook / 通知均可在任务中心追踪
- [ ] 失败任务能看到明确错误原因
- [ ] 重试不会造成重复扣费或重复写入核心业务数据

---

## 7. 第三阶段里程碑

### M7：套餐变更与商业审计

交付内容：

- 套餐升级 / 降级
- 加购包
- 收款记录
- 商业审计日志
- 高危操作二次确认

验收方式：

- 可完成升级补差价订单
- 可完成下周期降级
- 财务可导出收款明细
- 所有关键商业操作可审计

---

### M8：试用增长与客户转化

交付内容：

- 试用申请
- 试用租户开通
- 试用转付费
- 转化漏斗看板

验收方式：

- 从线索到试用到付费链路闭环
- 可按渠道、负责人查看转化率
- 试用到期提醒可用

---

### M9：开放集成与运维自动化

交付内容：

- API Key
- Webhook
- 租户健康巡检
- 后台任务中心
- 数据保留与删除策略

验收方式：

- API Key 不越租户
- Webhook 失败可重试和追踪
- 巡检能发现订阅 / 配额异常
- 异步任务全链路可观测

---

# 第三阶段推荐实施顺序

1. 商业操作审计日志
2. 收款记录与财务对账
3. 套餐升级 / 降级
4. 加购包 / 资源包
5. 高危操作二次验证
6. 试用申请与试用租户
7. 试用转付费
8. 转化漏斗看板
9. 数据保留与删除策略
10. 租户 API Key
11. Webhook 事件订阅
12. 后台任务中心
13. 租户级健康巡检

---

# 第三阶段完成后的上线判断

满足以下条件后，可以进入规模化 SaaS 运营阶段：

- [ ] 套餐升级 / 降级 / 加购可以闭环
- [ ] 商业收款、订单、订阅、审计日志口径一致
- [ ] 高危操作具备二次验证和完整审计
- [ ] 试用到付费转化链路可运营
- [ ] 平台可持续追踪转化率、续费率、流失率
- [ ] 租户数据保留和删除策略明确
- [ ] API Key / Webhook 有基础安全和限流能力
- [ ] 后台任务失败可追踪、可重试、可告警
- [ ] 租户健康巡检可发现关键异常
