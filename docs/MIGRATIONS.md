# 迁移治理

本模板使用 repo-local forward-only SQL runner：`go run cmd/migrate/main.go`。迁移按文件名排序执行，已执行版本记录在 `schema_migrations`。

模板定位是 MySQL/GORM schema-first：SQL migration 和 live MySQL schema 是数据结构事实来源，`internal/repository/mysql/{dao,model}` 只由 `gormgen` 从 live schema 生成。不要手改 generated DAO/Model；字段变化应先新增 migration，应用到 MySQL 后再重新生成。

## 规则

- 迁移文件只追加，不修改已合并到主干的 SQL。
- `schema_migrations` 记录 `version`、SHA-256 `checksum` 和 `rollback_policy`；重复执行时会校验当前文件内容是否与已执行记录一致。
- 非 bootstrap 迁移必须以 `-- rollback-policy: forward-only|manual|reversible-by-new-migration` 声明回滚策略。
- 如果 checksum mismatch，说明已执行迁移被修改，应新增修正迁移，不要覆盖历史 SQL。
- 当前 runner 不支持自动 down migration。需要回滚时，新增明确的 forward 修复迁移，例如 `000010_revert_xxx.sql`。
- bootstrap SQL 必须幂等，使用 `ON DUPLICATE KEY UPDATE` 或等价策略。

## 本地验证

生成新迁移草稿：

```bash
make new-migration TABLE=admin
```

补齐 SQL 后再运行迁移验证：

```bash
make migration-smoke
```

该命令会执行两次 `-bootstrap`，验证迁移和 seed 可重复执行。若本地 MySQL 用户/密码不同，请通过 `.env` 或命令行覆盖 `MYSQL_*` 变量。

注意：当前 `migration-smoke` 依赖 `python3` 的 `bcrypt` 包来生成 bootstrap 测试密码哈希；如果本机缺少该依赖，命令会直接失败并提示补齐环境。

## CI 策略

CI 使用 MySQL service 执行迁移 smoke，并在 live schema 上重新生成 DAO/Model 后检查 drift。这样可以同时发现：

- SQL 语法错误
- 非幂等 bootstrap
- 已执行迁移被修改
- migration 与 generated DAO/model 不一致
