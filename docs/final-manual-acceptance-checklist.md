# 最终验收结果记录

## 验收日期
- 2026-05-08
- 2026-05-10（补充回归与运营体验验收）

## 验收范围
- 登录 / 首次改密 / 登出
- RBAC 与未改密访问限制
- 学员、课包、排课、缴费、调补课主链
- 并发保护与课包扣减一致性
- 发布验证（测试、前端构建、Swagger）
- 管理员 / teacher 双角色业务回归
- 前端高频页面运营友好性改造验收

## 验收结果

### 1. 认证与权限
- [x] admin 首次登录返回 `mustChangePassword=true`
- [x] teacher 首次登录返回 `mustChangePassword=true`
- [x] 未改密用户访问业务接口返回 403
- [x] 改密成功后可进入业务页
- [x] teacher 访问 `/users` 返回 403
- [x] teacher 创建课程返回 403
- [x] teacher 新增缴费记录返回 403
- [x] logout 后 token 失效，再访问 `/auth/me` 返回 401

### 2. 业务主链
- [x] `GET /lesson-packages` 正常
- [x] `GET /payment-records` 正常
- [x] `GET /reschedule-records` 正常
- [x] `GET /users`（admin）正常
- [x] `POST /courses` 正常
- [x] `POST /student` 正常
- [x] `POST /lesson-packages` 正常
- [x] `POST /payment-records` 正常
- [x] `POST /schedules` 正常
- [x] `POST /schedules/{id}/leave` 正常
- [x] `POST /makeup-schedules` 正常
- [x] `POST /lesson-records` 正常
- [x] 删除与约束校验符合当前业务规则（已完成排课、有关联课时包/学员/课程时禁止误删）

### 3. 并发与一致性
- [x] `payment_update` 并发压测通过
- [x] `schedule_complete_toggle` 并发压测通过
- [x] `schedule_create` 并发压测出现成功创建与 409 冲突，说明冲突保护生效
- [x] 课包扣减在完成课次切换中保持一致

### 4. 发布验证
- [x] `go test ./...` 通过
- [x] `frontend/web npm run build` 多轮回归通过
- [x] `make ci` 通过
- [x] Swagger 重生成通过
- [x] `make server-smoke` 通过
- [x] `make smoke-e2e` 已具备正式入口，并已明确默认种子口令与显式 token / 密码两种使用口径

### 5. 数据与迁移
- [x] `000009_harden_constraints_and_keys.sql` 已在真实 dev 库跑通
- [x] decimal / nullable 适配已完成
- [x] 外键/唯一约束/检查约束已落地
- [x] 本轮已完成清库、bootstrap、服务重启、健康检查回归

### 6. 审计与运维材料
- [x] 已补审计日志基础能力
- [x] 已补发布/部署/回滚手册
- [x] 已补压测脚本
- [x] 已补备份恢复脚本骨架
- [x] 已完成一次事务回滚能力探针，结果正常

### 7. 前端运营体验升级
- [x] 排课、缴费、课时包高频弹窗由手填 ID 改为选择式交互
- [x] 排课管理页筛选由手填 ID 改为选择式交互
- [x] 缴费管理页筛选由手填 ID 改为选择式交互
- [x] 课时包管理页筛选由手填 ID 改为选择式交互
- [x] 上课记录页筛选由手填 ID 改为选择式交互
- [x] 排课、缴费、上课记录等列表与详情展示改为业务名称优先，缺失时再回退显示 ID
- [x] 空状态与部分提示文案已完成一轮产品化整理

## 已知边界
- 前端包体已完成一轮拆包优化，当前不再存在阻断发布的单一 Element Plus 超大包；后续仍可继续按实际访问热点做更细粒度优化。
- 备份恢复目前完成的是回滚能力探针与脚本落地，若按严格生产标准，仍建议补一次完整的导出后恢复演练留档。
- 审计日志已覆盖核心业务动作，但如需更严格合规，可继续补用户管理等更多操作面。
- 仍有少量后端错误提示文案偏技术风格，更适合在正式公网推广前继续统一为运营可读文案。
- 仓库中仍存在一批早于本轮验收的历史改动，建议正式交付前单独做一次噪音清理与范围冻结。

## 当前发布判断
- 代码与主链功能：通过
- 联调与基础发布验证：通过
- 可控发布：可以
- 公开演示与受控试运营：可以
- 完全放开公网商用终版：暂不建议直接定义为终版
