# 发布与回滚手册

## 1. 发布前检查

发布前必须确认：
- `git status --short` 为空
- `make ci` 通过（包含 `go vet ./...`、`go test ./... -cover`、`go build ./...`、`make generated-check`、前端构建、migration lint、生产配置核查）
- `make release-backup-gate` 通过，确认最近一次成功备份未超出 `BACKUP_RELEASE_MAX_HOURS`
- `AUTH_MODE=required`
- `AUTH_TEMPLATE_TOKEN_ENABLED=false`
- 已配置正式 `JWT_SECRET`
- 已设置明确的 `CORS_ALLOWED_ORIGINS`
- 已验证登录失败锁定策略、logout 失效、首次改密
- 已验证设备级会话治理与密码恢复链路
- 如需连库重跑脚本，数据库名统一使用 `edu-schedule-system`

## 2. 推荐发布顺序

1. 备份数据库
2. 发布后端二进制或镜像
3. 执行 migration
4. 验证 `/system/health` 与 `/system/ready`
5. 验证登录、`/auth/me`、会话列表、单设备下线、全端强退、密码恢复主链
6. 发布前端
7. 再次做主流程抽检

## 3. Migration 失败处理

### 3.1 尚未写入 `schema_migrations`
1. 停止继续发布
2. 查看失败 SQL 与数据库当前对象状态
3. 若是脏数据导致，先修复数据再重跑
4. 若是约束半执行，补一条 forward migration 修正，不改历史已发布 SQL

### 3.2 已部分落库但未完成
1. 先盘点 `information_schema.table_constraints`、`statistics`
2. 确认已新增的约束、索引、外键
3. 用新的 forward migration 做幂等补齐，不直接修改已执行历史迁移

## 4. 数据修复手册

### 4.1 课包金额错账
- 以 `payment_status='paid'` 的流水汇总为准
- 修复后重新校验 `paid_amount <= total_amount`

### 4.2 课时错账
- `used_lessons = completed_schedule_count`
- `remain_lessons = total_lessons - used_lessons`
- 修复后校验 `used_lessons + remain_lessons = total_lessons`

### 4.3 排课/调补课引用脏数据
1. 清理不存在的 `original_schedule_id`
2. 清理不存在的 `operator_id`
3. 再补外键或重跑约束 migration

## 5. 用户反馈安全问题处理

### 5.1 会话异常
1. 查 `auth_session`
2. 确认 `risk_flags` / `is_suspicious`
3. 视情况单设备下线或全端强退
4. 记录处置人、时间、原因

### 5.2 找回密码争议
1. 查 `auth_recovery_challenge`
2. 查 `auth_recovery_audit`
3. 核验通道、IP、UA、验证码状态与完成时间
4. 必要时强制重置密码并全端下线

## 6. 回滚策略

本项目 migration 采用 forward-only。

如果发布后发现问题：
- 不回改历史 migration 文件
- 新增一条修正 migration
- 必要时先回滚应用版本，再执行数据修正 migration

## 7. 发布后抽检

至少抽检：
- admin 登录 + 首次改密
- 连续输错密码触发锁定，锁定到期后恢复登录
- teacher 登录 + 权限限制 + 数据范围限制
- 会话列表
- 单设备下线
- 全端强退
- 密码恢复
- 学员列表
- 课包列表
- 排课创建
- 缴费列表与新增
- 调补课记录
- 审计日志输出
- logout 后旧 token 失效

## 8. 备份与恢复演练

优先命令：

```bash
MYSQL_DB=edu-schedule-system MYSQL_USER=app MYSQL_PASSWORD=app ./scripts/db_backup.sh
BACKUP_HISTORY_FILE=./logs/backup-history.ndjson make release-backup-gate
MYSQL_DB=edu-schedule-system MYSQL_USER=app MYSQL_PASSWORD=app ./scripts/db_restore.sh ./output/db-backups/<file>.sql
```

说明：
- 发布前必须先执行备份，并用 `make release-backup-gate` 验证 `logs/backup-history.ndjson` 中存在最近成功备份。
- 可通过 `BACKUP_RELEASE_MAX_HOURS` 设置允许的最新备份时效，默认 26 小时。
- 如生产要求备份已上传对象存储，设置 `BACKUP_RELEASE_REQUIRE_OBJECT_STORAGE=true`。
- 本机若无 `mysql` / `mysqldump`，脚本会自动回退到 Docker 容器内执行。
- 默认容器名为 `edu-schedule-system-mysql-1`，可通过 `MYSQL_CONTAINER` 覆盖。
- 建议每次恢复演练后把日期、库名、操作者、结果写入发布留档。
