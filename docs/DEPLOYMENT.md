# 部署说明

## 环境差异

### dev
- 用于本地开发
- 可开启 Swagger / metrics / 调试能力
- 可临时用本地 `.env`
- 若需无鉴权调试，只允许本机临时设置 `AUTH_MODE=disabled`
- 本地联调数据库名统一使用 `edu-schedule-system`

### fat
- 用于功能联调
- `AUTH_MODE=required`
- `AUTH_TEMPLATE_TOKEN_ENABLED=false`
- 恢复发送通道建议使用 `AUTH_RECOVERY_DELIVERY_MODE=placeholder` 或受控测试邮件通道
- 必须使用正式 JWT secret 注入
- 必须设置明确 CORS 白名单

### uat
- 用于验收测试
- `AUTH_MODE=required`
- `AUTH_TEMPLATE_TOKEN_ENABLED=false`
- 如需验证找回密码链路，建议启用 `AUTH_RECOVERY_DELIVERY_MODE=email` 并使用受控 SMTP 账号与测试邮箱
- 禁止使用默认开发密钥
- 迁移前先做数据库备份
- 验证登录失败锁定、logout 失效、恢复链路、关键审计日志与恢复失败告警查询页

### pro
- 用于正式生产
- `AUTH_MODE=required`
- `AUTH_TEMPLATE_TOKEN_ENABLED=false`
- `AUTH_RECOVERY_DELIVERY_MODE` 不得为 `placeholder`
- 禁止 Swagger / pprof 暴露
- 必须使用正式密钥管理和数据库备份策略
- 必须明确备份恢复方案，允许本机无 mysql 客户端时走 Docker 容器兜底
- 必须接入 Prometheus 与告警平台

## 数据库初始化流程

1. 创建数据库与应用账号，数据库名使用 `edu-schedule-system`
2. 配置 `MYSQL_READ_*` / `MYSQL_WRITE_*`
3. 注入正式 `JWT_SECRET`
4. 执行 migration
5. 如需 bootstrap，仅在明确允许的环境执行
6. 首次登录后立刻完成管理员改密
7. 验证数据库备份脚本在当前宿主机或 Docker 容器路径可用
8. 至少完成一次备份恢复演练留档

## 生产发布顺序

1. 代码封板
2. 备份数据库
3. 发布后端
4. 执行 migration
5. 验证健康检查、登录、会话治理与恢复主链
6. 验证平台审计导出与恢复失败告警查询页
7. 发布前端
8. 完成冒烟与人工验收

## CORS 策略

生产必须显式设置：
- `CORS_ALLOWED_ORIGINS=https://your-admin.example.com`
- `CORS_ALLOW_CREDENTIALS=false`

禁止使用：
- `*`
- 空白默认放开策略

## 密钥策略

以下配置不得使用示例值：
- `JWT_SECRET`
- `AES_SECRET`
- `RSA_PRIVATE_KEY`
- `AUTH_RECOVERY_DELIVERY_EMAIL_PASSWORD`

统一通过部署平台环境变量或密钥管理系统注入。

## 恢复通道部署建议

如果启用 `AUTH_RECOVERY_DELIVERY_MODE=email`，至少应明确：
- SMTP Host / Port / Username / Password / From
- 建议额外设置 `AUTH_RECOVERY_DELIVERY_TIMEOUT_SECONDS`、`AUTH_RECOVERY_DELIVERY_MAX_ATTEMPTS`、`AUTH_RECOVERY_DELIVERY_EMAIL_LOCAL_NAME`
- 使用专门的恢复邮件账号，不与个人邮箱混用
- 在 `fat/uat` 先使用测试邮箱验证收件、超时、认证失败、错误日志与审计链路
- `pro` 环境禁止依赖 `AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED`
- 邮件投递失败时，运营侧应有可追踪日志或告警
- 已将 `logs/recovery-alerts.ndjson` 纳入日志采集、保留与轮转策略
- 平台管理员应能访问恢复失败告警页，作为第一线排障入口

如果启用 `AUTH_RECOVERY_DELIVERY_MODE=sms`，当前建议口径：
- 仅作为 provider 抽象与联调占位模式使用
- 至少配置 `AUTH_RECOVERY_DELIVERY_SMS_PROVIDER`、`AUTH_RECOVERY_DELIVERY_SMS_SIGN_NAME`
- 在正式短信通道接入前，不要把 stub 模式误当作真实生产短信能力
