# 历史噪音改动分拣记录

## 时间
- 2026-05-10

## 目的
对当前工作区里未提交的历史改动做风险分拣，避免在交付前误删仍然有效的业务修复。

---

## 一、可确认属于本轮之外、但不建议直接回滚的改动

这些文件当前仍有未提交 diff，但内容明显影响真实业务链路、安全口径或种子账号可用性。由于已经参与了本轮联调/回归结果，不建议在未做专项回归前直接清掉。

### 1. 认证与安全相关
- `internal/api/auth/handler.go`
- `internal/service/auth/service.go`
- `internal/service/auth/service_change_password_test.go`
- `migrations/bootstrap/000002_seed_sys_user.sql`

判断：
- 涉及改密后 token 处理、默认种子密码、登录安全与会话口径
- 已参与本轮 admin / teacher 登录、改密、回归
- **建议保留，单独走安全改动复核**

### 2. 数据权限与只读范围相关
- `internal/api/course/course_handler.gen.go`
- `internal/api/lesson_package/lesson_package_handler.gen.go`
- `internal/api/payment_record/business_handler.go`
- `internal/api/reschedule_record/business_handler.go`
- `internal/api/student/student_handler.gen.go`

判断：
- 涉及 teacher / self data scope 返回口径
- 已参与本轮 teacher 权限回归
- **建议保留，单独走 RBAC/数据范围复核**

### 3. 删除保护与业务约束相关
- `internal/service/lesson_package/service.go`
- `internal/service/schedule/service.go`
- `internal/service/student/service.go`
- `internal/service/user/service.go`

判断：
- 涉及“有关联业务数据时禁止误删”等保护逻辑
- 已被本轮接口回归实际验证命中
- **建议保留，单独走业务约束复核**

---

## 二、明显偏辅助/留痕的改动

### 1. 学习记录
- `.learnings/ERRORS.md`

判断：
- 属于过程留痕，不影响业务发布
- 可保留，也可在正式交付前移出主交付范围

### 2. API 文档口令示例修正
- `docs/api/接口设计文档.md`

判断：
- 将旧示例密码调整为当前真实默认口令
- 文档口径是对的，不建议回滚

### 3. 自动生成前端声明
- `frontend/web/src/components.d.ts`

判断：
- 多半是组件按需加载调整后的自然漂移
- 不构成风险，可随最终前端交付一起提交或重生成后再确认

---

## 三、当前建议

### 建议立即做的事
1. 不直接回滚上述后端历史改动
2. 把它们视为“待专项复核的有效改动”
3. 正式交付前单开一次分支或 MR，只审这批 backend/security/rbac 变更

### 不建议现在做的事
1. 不带回归直接 `git checkout --` 清空这些改动
2. 不在当前已经稳定可试运营的状态下做大规模后端回滚

---

## 四、当前结论

当前未提交的历史改动里，真正意义上的“纯噪音”已经基本处理到：
- 临时脚本已移出主工作目录
- 前端高频交互改造已独立提交
- 交付文档已独立提交

剩余未提交项，更多是：
- 安全口径
- 数据范围口径
- 删除保护口径
- 文档/生成产物留痕

它们不适合定义为“随手清掉的噪音”，更适合定义为：
- **待专项复核的历史有效改动**
