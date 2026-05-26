# 权限矩阵（平台运营拆分版）

## 目标

当前 SaaS 权限体系拆成三层：

1. `permission`：动作权限，决定“能不能做”。
2. `menuPermission`：菜单入口，决定“能不能看到入口”。
3. `featureFlags`：套餐/租户能力开关，决定“当前套餐/租户有没有这类能力”。

## 当前角色口径

租户侧：

- `org_admin`
- `campus_admin`
- `teacher`

平台侧预置角色：

- `platform_admin`：平台全权限。
- `platform_ops`：租户、校区、客户成功查看与跟进。
- `platform_finance`：订阅、续费、配额与平台经营报表。
- `platform_auditor`：审计日志与恢复告警只读/导出。
- `platform_support`：客户排障、恢复告警处理、有限业务数据查看。

## 平台治理权限点

| 能力 | permission | menuPermission | featureFlag | 默认角色 |
|---|---|---|---|---|
| 租户列表 | `platform:tenant:list` | `menu:platform:view` / `menu:platform:orgs:view` | `platform_management` | admin, ops, finance, support |
| 租户详情 | `platform:tenant:get` | `menu:platform:view` / `menu:platform:orgs:view` | `platform_management` | admin, ops, finance, support |
| 租户创建 | `platform:tenant:create` | `menu:platform:view` / `menu:platform:orgs:view` | `platform_management` | admin |
| 租户更新 | `platform:tenant:update` | `menu:platform:view` / `menu:platform:orgs:view` | `platform_management` | admin |
| 租户停用 | `platform:tenant:disable` | `menu:platform:view` / `menu:platform:orgs:view` | `platform_management` | admin |
| 租户恢复 | `platform:tenant:restore` | `menu:platform:view` / `menu:platform:orgs:view` | `platform_management` | admin |
| 平台校区查看 | `platform:campus:list` | `menu:platform:view` | `platform_management` | admin, ops, support |
| 平台校区创建 | `platform:campus:create` | `menu:platform:view` | `platform_management` | admin |
| 平台校区更新 | `platform:campus:update` | `menu:platform:view` | `platform_management` | admin |
| 平台校区停用 | `platform:campus:disable` | `menu:platform:view` | `platform_management` | admin |
| 订阅列表 | `platform:subscription:list` | `menu:platform:view` | `subscription_center` | admin, ops, finance |
| 订阅详情 | `platform:subscription:get` | `menu:platform:view` | `subscription_center` | admin, ops, finance |
| 订阅更新 | `platform:subscription:update` | `menu:platform:view` | `subscription_center` | admin, finance |
| 调整配额 | `platform:subscription:adjust-quota` | `menu:platform:view` | `subscription_center` | admin, finance |
| 延长试用 | `platform:subscription:extend-trial` | `menu:platform:view` | `subscription_center` | admin, finance |
| 暂停订阅 | `platform:subscription:suspend` | `menu:platform:view` | `subscription_center` | admin, finance |
| 恢复订阅 | `platform:subscription:resume` | `menu:platform:view` | `subscription_center` | admin, finance |
| 客户成功列表 | `platform:cs:list` | `menu:platform:view` | `customer_success` | admin, ops |
| 更新负责人 | `platform:cs:update-owner` | `menu:platform:view` | `customer_success` | admin, ops |
| 添加跟进记录 | `platform:cs:add-followup` | `menu:platform:view` | `customer_success` | admin, ops |
| 更新跟进状态 | `platform:cs:update-status` | `menu:platform:view` | `customer_success` | admin, ops |
| 审计日志查看 | `platform:audit:list` | `menu:platform:view` | `audit_export` | admin, auditor |
| 审计日志导出 | `platform:audit:export` | `menu:platform:view` | `audit_export` | admin, auditor |
| 恢复告警查看 | `platform:recovery-alert:list` | `menu:platform:view` | `recovery_ops` | admin, auditor, support |
| 恢复告警处理 | `platform:recovery-alert:resolve` | `menu:platform:view` | `recovery_ops` | admin, support |
| 平台报表查看 | `platform:report:view` | `menu:platform:view` | `platform_report` | admin, ops, finance |
| 平台报表导出 | `platform:report:export` | `menu:platform:view` | `platform_report` | admin, finance |

兼容说明：代码中保留旧常量别名：

- `PermPlatformOrgList = PermPlatformTenantList`
- `PermPlatformOrgCreate = PermPlatformTenantCreate`
- `PermPlatformOrgUpdate = PermPlatformTenantUpdate`
- `PermPlatformAuditView = PermPlatformAuditList`
- `PermPlatformRecoveryView = PermPlatformRecoveryAlertList`
- `PermPlatformRecoveryManage = PermPlatformRecoveryAlertResolve`

## 机构侧能力

| 能力 | permission | menuPermission | featureFlag | 默认角色 |
|---|---|---|---|---|
| 订阅中心 | `auth:me:view` | `menu:dashboard:view` | `subscription_center` | org_admin, campus_admin |
| 机构设置查看 | `tenant:settings:view` | `menu:dashboard:view` | 无 | org_admin |
| 机构设置更新 | `tenant:settings:update` | `menu:dashboard:view` | 无 | org_admin |
| 用户管理 | `system:user:list` 及相关动作 | `menu:system:user:view` | `user_management` | org_admin, campus_admin |
| 校区管理 | `platform:campus:list` 及相关动作 | `menu:platform:view` | 无 | org_admin |
| 学员/课程/课包/排课等业务 | 对应模块 `list/get/create/update/delete` | 对应业务菜单 | 对应模块 feature | org_admin, campus_admin, teacher 按数据范围 |

## 后端路由校验现状

- 平台路由先校验平台预置角色，再校验具体 `permission`。
- 审计导出单独要求 `platform:audit:export`，不能只靠 `platform:audit:list`。
- 平台租户/校区/订阅/客户成功接口已按读写权限分组。
- 租户侧组织设置和当前订阅接口仍挂在租户 admin 路由下。

## 后续原则

- 新页面入口先补 `menuPermission`。
- 新动作接口必须补独立 `permission`，不能复用过宽权限。
- 套餐/交付差异优先补 `featureFlags`。
- 平台侧新增模块必须明确归属：ops / finance / auditor / support / admin。
- 所有无权限访问必须返回 403。
