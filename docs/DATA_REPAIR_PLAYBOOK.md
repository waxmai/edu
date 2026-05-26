# 平台数据修复辅助说明

## 当前定位

当前版本暂未提供直接在线修复数据的危险按钮，而是先提供平台侧修复辅助入口与 SOP 说明，降低误操作风险。

## 建议优先支持的修复场景

### 1. 课包金额异常
- 核验 `payment_record`
- 核验 `lesson_package.paid_amount`
- 对照 `paid_amount <= total_amount`

### 2. 课时扣减异常
- 核验 `schedule` 的 `completed` 状态数量
- 对照 `lesson_package.used_lessons` / `remain_lessons`
- 校验 `used + remain = total`

### 3. 恢复链路异常
- 查 recovery alerts
- 查 recovery audit
- 查 challenge 状态与过期时间

### 4. 账号/租户权限异常
- 核验 `sys_user.role_code`
- 核验 `organization_id` / `campus_id`
- 核验当前 permission / feature flag / menu permission 口径

## 当前页面入口对应关系

### 平台运营页已提供的辅助入口
- 恢复异常排查 → `/platform/recovery-alerts`
- 审计追溯 → `/platform/audit-logs`
- 续费跟进台 → `/platform/subscriptions`
- 租户治理 → `/platform/tenants`

这意味着当前平台运营页已经从“只看状态”进入“可跳转到排查动作”的半工具化阶段。

## 推荐排查 SOP

### 场景 A，恢复验证码发送失败
1. 先进入 `/platform/recovery-alerts`
2. 按用户名 / 通道 / challengeId / 目标掩码过滤
3. 记录失败时间、错误文案、challenge 状态、过期时间
4. 再跳转 `/platform/audit-logs`
5. 用 Trace ID、动作、操作人、时间范围回放链路
6. 最后判断是：
   - 通道问题
   - challenge 过期 / 状态不一致
   - 用户目标信息错误
   - 环境配置问题

### 场景 B，权限异常 / 菜单异常
1. 先确认当前账号 `role_code`
2. 再确认 `organization_id` / `campus_id` 归属是否正确
3. 对照 feature flag 是否启用对应能力
4. 对照 permission / menu permission 是否匹配前端显示口径
5. 必要时通过审计日志回放最近的账号、租户、订阅变更动作

### 场景 C，续费治理与租户健康异常
1. 从 `/platform/subscriptions` 看 follow-up 状态、剩余天数、账号/校区配额
2. 从 `/platform/tenants` 看租户健康级别、配额与租户设置
3. 若发现异常，再进入 `/platform/audit-logs` 回放最近修改链路

## 当前阶段建议

先提供：
- 修复说明
- 排查入口
- 审计与 trace 追踪入口
- 明确的排查 SOP

后续再考虑：
- 受控修复工具页
- 二次确认 + 审批流
- 只允许平台超管执行的修复动作
- 修复前后快照留档
- 修复动作强制写审计日志

## 当前边界

现阶段不建议直接上线：
- 任意 SQL 修复输入框
- 无审批的直接修复按钮
- 批量改课时 / 批量改金额 / 批量改权限这类高风险动作

更稳妥的推进顺序应是：
1. 先把排查入口、Trace 追溯、SOP 做实
2. 再收敛为少量高频、低破坏面的受控修复动作
3. 最后再考虑审批化、工单化、快照化的真正修复工具
