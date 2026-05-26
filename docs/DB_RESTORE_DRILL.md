# 数据库恢复演练 SOP

> 目标：验证生产备份可用，确保发生误删、迁移失败、主库故障时能在可接受时间内恢复核心数据。

## 1. 演练频率

- UAT / 试运行期：每次正式上线前至少一次。
- 商业化后：至少每月一次。
- 高风险变更前：涉及迁移、批量导入、订阅/租户生命周期脚本时，发布前追加一次。

## 2. 红线

- 禁止直接恢复到生产库。
- 禁止在生产库执行破坏性验证 SQL。
- 恢复目标必须是独立演练库，例如：`edu_schedule_restore_drill_YYYYMMDD`。
- 演练完成后，确认演练库无敏感外发配置，必要时销毁。

## 3. 准备项

- 一份最近 24 小时内的备份文件：`.sql`、`.sql.gz` 或 `.sql.gz.enc`。
- 恢复目标 MySQL 实例和独立数据库。
- 与备份匹配的 `BACKUP_ENCRYPT_KEY`，如启用了加密。
- 当前代码版本和 migration 工具可运行。

## 4. 恢复步骤

### 4.1 获取备份

```bash
ls -lh ./output/db-backups/
```

如备份在对象存储，先下载到本地安全目录。

### 4.2 创建演练库

示例：

```bash
mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u root -p \
  -e 'CREATE DATABASE IF NOT EXISTS edu_schedule_restore_drill DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;'
```

### 4.3 恢复备份到演练库

```bash
MYSQL_DB=edu_schedule_restore_drill \
MYSQL_USER=app \
MYSQL_PASSWORD='***' \
BACKUP_ENCRYPT_KEY='***' \
./scripts/db_restore.sh ./output/db-backups/<backup-file>.sql.gz.enc
```

未加密备份可省略 `BACKUP_ENCRYPT_KEY`。

### 4.4 执行 migration check

```bash
MYSQL_READ_NAME=edu_schedule_restore_drill \
MYSQL_WRITE_NAME=edu_schedule_restore_drill \
make migration-check
```

如当前配置不支持 `MYSQL_READ_NAME/MYSQL_WRITE_NAME` 环境覆盖，则使用演练专用 config 文件执行迁移检查。

### 4.5 执行 smoke test

至少验证：

- `/system/ready` 可用。
- 平台管理员登录可用。
- 租户列表可查。
- 学员、课程、课包、排课列表可查。
- 订阅状态查询可用。

推荐：

```bash
API_BASE_URL=http://127.0.0.1:8080 make smoke-e2e
```

## 5. 验收项

记录以下结果：

- 表数量是否符合预期。
- 核心表数据量是否合理：
  - `organization`
  - `campus`
  - `sys_user`
  - `student`
  - `course`
  - `lesson_package`
  - `schedule`
  - `payment_record`
- 登录是否可用。
- 核心列表查询是否可用。
- migration check 是否通过。
- smoke test 是否通过。
- 恢复耗时是否在 RTO 目标内。

## 6. 演练记录模板

```markdown
## 恢复演练记录

- 演练时间：
- 操作者：
- 代码版本 / commit：
- 备份文件：
- 备份文件大小：
- 恢复目标库：
- 恢复开始时间：
- 恢复结束时间：
- 恢复耗时：
- migration check：通过 / 失败
- smoke test：通过 / 失败
- 核心表数量核对：通过 / 失败
- 问题记录：
- 结论：通过 / 不通过
```

## 7. 常见失败处理

- 解密失败：确认 `BACKUP_ENCRYPT_KEY` 与备份生成时一致。
- gzip 解压失败：确认文件完整下载，检查文件大小与备份历史记录一致。
- restore 中断：检查 MySQL 权限、连接、目标库是否存在。
- migration check 失败：确认恢复备份对应的代码版本和 migration 状态。
- smoke test 失败：优先检查配置是否仍指向演练库，避免误连生产。
