# 慢 SQL 与索引治理记录

## 当前优先排查对象

### 1. 租户治理相关统计
- `organization`
- `subscription`
- `campus`
- `sys_user`

关注查询：
- 机构列表聚合
- 校区列表聚合
- 续费跟进台统计

### 2. 教务核心表
- `student`
- `lesson_package`
- `payment_record`
- `schedule`
- `reschedule_record`

关注查询：
- 列表页筛选 + 分页
- 统计聚合
- 导出

## 当前初步结论

### 已有正向信号
- 多数核心业务表已具备 org/campus 维度索引
- 多租户核心查询普遍带 `organization_id` / `campus_id` 条件
- 平台治理页目前聚合量还不算过重

### 当前仍需重点验证的慢点
- 平台租户/校区页中多次 `count` 聚合查询
- 续费跟进台的机构、账号、校区聚合统计
- 审计日志页在日志量增大后的过滤与导出表现

## 当前已留档的真实样本

### 1. 压测留档现状
当前已沉淀 3 份有效报告：
- `output/loadtest-reports/loadtest-20260516-145733.jsonl`
- `output/loadtest-reports/loadtest-20260516-145754.jsonl`
- `output/loadtest-reports/loadtest-campus-success-20260516.jsonl`

### 2. 三份样本分别说明了什么

#### 样本 A，首次改密前阻断
`loadtest-20260516-145733.jsonl`
- 三个 case 都已实际发出请求，`count=80`
- 但全部被 `首次登录后请先修改密码` 拦截为 `403`

结论：
- 压测链路已真正打到 API，不再是空样本
- 首次改密策略会在业务接口层统一阻断未完成改密的账号

#### 样本 B，平台超管写链权限边界
`loadtest-20260516-145754.jsonl`
- 三个 case 仍为 `count=80`
- 全部被 `无权限访问该接口` 拦截为 `403`

结论：
- `platform_admin` 并不是教务写链路压测的合适账号
- 平台角色和机构/校区业务写角色边界是明确生效的

#### 样本 C，校区管理员真实写链路样本
`loadtest-campus-success-20260516.jsonl`
- `schedule_create`: `ok=1`, `conflicts=72`, `server_errors=7`
- `payment_update`: `404=80`
- `schedule_complete_toggle`: `ok=40`, `conflicts=40`

结论：
- 已拿到真正包含 `2xx / 409 / 404 / 500` 的业务写压测样本
- 排课写链在高并发下主要表现为“冲突保护生效”，这是正向信号
- 课时完成切换链路出现大量 `409 lesson package lessons are insufficient...`，说明业务保护也在生效
- `payment_update` 全部 `404`，说明当前默认测试数据口径下 `PAYMENT_ID=1` 并不是有效样本目标，压测脚本默认数据依赖仍需显式准备
- `schedule_create` 出现少量 `500`，说明除了冲突保护外，还存在需要继续排查的并发异常点

## 当前环境/脚本前置问题结论

### 1. 本地 auth 限流与 Redis 依赖
已验证：
- 当 `REDIS_ENABLED=false` 且 auth endpoint rate limit 仍默认开启时
- `POST /api/v1/auth/login` 会直接返回 `503`

结论：
- 本地无 Redis 模式下，登录压测和需要自动登录的 smoke/loadtest 链路会被直接阻断
- 这不是账号错误，而是本地运行前提问题

### 2. bootstrap hash 传参链问题
已验证：
- 通过 `make bootstrap BOOTSTRAP_*_PASSWORD_HASH=...` 传 bcrypt hash 时，如果 shell/make 传递链处理不当，hash 中的 `$` 可能在进入进程前被展开或截断
- 现在更稳的本地口径是直接传 `BOOTSTRAP_*_PASSWORD=...`，由 migrate 进程内生成 bcrypt

结论：
- 当前 bootstrap 口令注入链路对 bcrypt hash 不够稳
- 这会影响本地 smoke、loadtest、受控试发布前准备效率

## 已补的 EXPLAIN 留档

### 平台治理聚合相关
本地 MySQL 当前 EXPLAIN 结果：

```sql
EXPLAIN SELECT id, org_code, org_name, status, subscription_status, edition_code
FROM organization
ORDER BY id ASC;
-- -> Index scan on organization using PRIMARY

EXPLAIN SELECT organization_id, status, plan_code, ends_at, max_campuses, max_users,
               follow_up_status, follow_up_owner, follow_up_note, last_contact_at
FROM subscription;
-- -> Table scan on subscription

EXPLAIN SELECT id, organization_id, status
FROM campus;
-- -> Covering index scan on campus using idx_campus_org_status

EXPLAIN SELECT id, organization_id, campus_id, status
FROM sys_user;
-- -> Table scan on sys_user
```

### 当前 EXPLAIN 直接结论
- `organization` 当前走主键扫描，问题不大
- `campus` 已命中 `idx_campus_org_status`，方向是对的
- `subscription` 当前仍是全表扫描
- `sys_user` 当前仍是全表扫描

这与代码层面的风险判断一致：
- 平台治理页和续费跟进页最先需要盯的，不是单个业务 SQL 的 where 条件遗漏
- 而是 `subscription` / `sys_user` 被整表读取后再做内存聚合

## 基于代码实现的当前查询风险判断

### 1. `ListOrganizations` 当前模式
当前实现是：
- 全表查 `organization`
- 全表查 `subscription`
- 全表查 `campus`
- 全表查 `sys_user`
- 再在内存中按 `organization_id` 聚合

风险：
- 在租户规模继续增长后，平台页会出现“为一个列表页拉全量从表再做内存聚合”的放大成本
- 这类查询未必先表现为单条 SQL 很慢，更可能先表现为读取行数偏大、接口响应变长、内存占用抬升

### 2. `ListSubscriptionPipeline` 当前模式
当前实现同样是：
- 全表查 `organization`
- 全表查 `subscription`
- 全表查 `sys_user`
- 全表查 `campus`
- 在内存中计算 `usedUsers` / `usedCampuses`

风险：
- 在平台续费跟进页，最先需要盯的是 `sys_user` 与 `campus` 的全量扫描成本
- 如果机构数、账号数、校区数增长明显，优先改造方向可能不是“继续补单列索引”，而是把按机构聚合改成 SQL 聚合、汇总表或缓存快照

### 3. 审计日志页风险点
当前代码与库结构现状表明：
- 当前数据库内并不存在 `audit_log` 物理表
- 因此审计页的性能瓶颈不能直接按“现成 MySQL audit_log 表”假设来写死

结论：
- 审计日志性能分析需要先确认真实存储介质和查询路径
- 在未确认之前，不应把 `audit_log` 的 EXPLAIN 当作已存在前提

## 当前建议索引复核方向
- `subscription(organization_id, status, ends_at)`
- `campus(organization_id, status)`
- `sys_user(organization_id, campus_id, status)`
- `lesson_package(organization_id, campus_id, student_id)`
- `reschedule_record(organization_id, campus_id)`

## 当前阶段结论

### 可以确认的
- 压测脚本已从“失败可留档”继续推进到“能稳定沉淀真实业务写样本”
- 平台治理相关查询的首要风险，已被代码与 EXPLAIN 双重指向为 **全表读取 + 内存聚合**
- `subscription` / `sys_user` 是当前平台治理聚合里最值得优先治理的两张表
- 当前压测前提里还存在 2 个工程问题需要单独修：
  - 本地无 Redis 时 auth 登录限流直接阻断
  - bootstrap bcrypt hash 传参链会损坏 hash
- 教务写链并发保护已有正向表现：
  - 排课冲突保护能稳定产出大量 `409`
  - 课时不足保护能稳定产出 `409`
- `payment_update` 默认样本 404 的直接原因已确认，不是接口坏，而是当前库内不存在 `payment_record.id=1`
- `schedule_create` 的少量 `500` 已收敛为预期 `409 conflict`
- 平台超管两条最重的聚合路径已开始收敛：
  - `ListOrganizations`
  - `ListSubscriptionPipeline`
- 这两条链路已从“整表读 `sys_user/campus` 后内存聚合”改为“按机构 SQL 聚合后回填”，属于实质性减压
- `ListCampuses` 中原本逐校区执行的 4 类 count 也已收敛为批量聚合：
  - 活跃账号数
  - 低课时学员数
  - 欠费学员数
  - 调课记录数
- 针对当前聚合路径的索引配套已补齐一轮：
  - `subscription(organization_id, status, ends_at)`
  - `sys_user(campus_id, status)`
  - `lesson_package(campus_id, student_id)`

### 仍未最终确认的
- 是否需要把平台治理聚合进一步收敛为单条 join SQL 或汇总表
- 审计日志页的真实存储与查询性能瓶颈

## 已实施的聚合优化
- `ListOrganizations`：`usedUsers/usedCampuses` 已改为按机构聚合 SQL 结果回填
- `ListSubscriptionPipeline`：`usedUsers/usedCampuses` 已改为按机构聚合 SQL 结果回填
- `ListCampuses`：逐校区 4 类统计已改为批量聚合查询回填
  - 活跃账号数
  - 低课时学员数
  - 欠费学员数
  - 调课记录数

### 当前收益
- 避免平台页每次都把 `sys_user` / `campus` 全量加载到应用层再做 map 聚合
- 避免 `ListCampuses` 对每个校区重复发起多组 count 查询，降低 N*4 型放大
- 新增索引后，`sys_user(campus_id,status)` 与 `lesson_package(campus_id,student_id)` 已开始被 EXPLAIN 命中
- 在租户、账号、校区规模上升时，这类查询的内存放大和读取放大都会更可控
- 改动对前端返回口径基本无侵入，属于后端实现侧优化

## 已实施的索引补强
- 新增 migration：`migrations/000016_add_platform_aggregate_indexes.sql`
- 当前已补索引：
  - `subscription(organization_id, status, ends_at)`
  - `sys_user(campus_id, status)`
  - `lesson_package(campus_id, student_id)`

### 当前 EXPLAIN 观察
- `sys_user` 的 campus 风险聚合已命中 `idx_sys_user_campus_status`
- `lesson_package` 的 campus 学员聚合已命中 `idx_lesson_package_campus_student`
- `subscription` 在按 `organization_id + status + ends_at` 的条件/排序口径下已具备复合索引支撑

## 后续补强口径
- `scripts/loadtest_core_api.sh` 的默认样本 ID 应与当前 bootstrap/默认库状态一致，避免 `PAYMENT_ID=1` 这类天然失效口径
- 压测前应先做样本自检，至少确认 `schedule` / `payment_record` 的目标 ID 在当前环境真实存在
- 排课创建并发路径应把数据库唯一键冲突尽量收敛为业务层 `409 conflict`，避免把预期冲突噪声暴露成 `500`
- 若平台治理聚合继续扩大，下一步优先评估：
  - `sys_user(organization_id, status)` 是否值得继续补齐
  - `lesson_package(campus_id, remain_lessons, student_id)` / `lesson_package(campus_id, paid_amount, total_amount, student_id)` 是否值得为风险聚合再细化
  - 是否引入机构级聚合快照表
