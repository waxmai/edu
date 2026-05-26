# 可观测性指南

当前项目已具备公开运营前所需的基础可观测能力，并新增了面向正式 SaaS 的日志保留收口：

- Trace ID，全链路请求关联
- Metrics，暴露 `/metrics` 供 Prometheus 抓取
- 结构化日志，记录环境、trace、请求与错误上下文
- 告警落盘，告警事件写入 `./logs/alerts.ndjson`
- 恢复链路告警落盘，恢复发送失败事件写入 `./logs/recovery-alerts.ndjson`
- 恢复失败平台查询，平台管理员可通过 `/platform/recovery-alerts` 做受控排查
- 标准输出告警，默认输出 `ALERT {...}` 结构化事件
- 告警阈值可通过环境变量配置，例如 `ALERT_HTTP_LATENCY_SECONDS`、`ALERT_DB_LATENCY_SECONDS`、`ALERT_REDIS_LATENCY_SECONDS`、`ALERT_EXTERNAL_LATENCY_SECONDS`、`ALERT_LOGIN_FAILURES`
- 告警与恢复日志默认启用保留/轮转保护，超过大小上限时自动截尾归档，并保留最近若干份归档文件

## 指标约定

- HTTP 指标必须包含 method、path、status、business code 等低基数字段
- 动态路由必须使用 alias，避免把真实 ID 打进指标标签
- 新模块至少关注四类信号：请求量、错误数、耗时、饱和度

## 当前告警覆盖

当前已接入：

- panic / 5xx 级错误告警
- 高风险认证事件告警基础通道
- 告警事件文件留档
- 恢复发送失败事件单独落盘
- 平台恢复失败告警页，支持按用户名、通道、challengeId、目标掩码查询
- 恢复失败按 channel / error 分类聚合、近 24h / 7d 趋势与 Top 错误概览
- `alerts.ndjson` 与 `recovery-alerts.ndjson` 默认执行 30 天保留、4MB 截尾轮转、最近 5 份归档保留

建议生产环境继续接入：

- 登录失败异常阈值告警
- HTTP 5xx 比例阈值告警
- p95 / p99 接口耗时告警
- MySQL / Redis readiness 连续失败告警
- 外部依赖请求超时告警
- 恢复发送失败事件文件的收集与集中告警
- recovery failure 按 channel / error 分类聚合与趋势展示

## Prometheus 建议

当前默认 job 名称：`edu-schedule-system`

建议至少接入以下告警规则：

1. `up == 0`
2. `/system/ready` 持续失败
3. 5xx 比例超过阈值
4. p95 latency 超过阈值
5. 数据库与 Redis 连续探活失败

## 5. 生产前置检查

在 UAT/生产发布前，先在真实环境变量下运行静态前置检查：

```bash
./scripts/prod_readiness_check.sh configs/pro_configs.toml
```

该脚本会拦截以下高风险配置：

- `JWT_SECRET` 为空、过短或仍使用示例/默认片段
- `AUTH_MODE=disabled`
- `AUTH_TEMPLATE_TOKEN_ENABLED=true`
- `SWAGGER_ENABLED=true` / `PPROF_ENABLED=true`
- `CORS_ALLOWED_ORIGINS` 未显式配置或使用 `*`
- 恢复通道仍为 `placeholder` 或邮件模式缺少 SMTP 必填项
- MySQL / Redis 关键连接变量缺失

> 说明：脚本不替代真实冒烟、备份恢复演练和压测，只负责把常见生产误配置前置失败。

## 6. OpenTelemetry 迁移路径

项目暂未强制引入 OpenTelemetry 依赖，避免部署复杂度无谓上升。若后续需要接入，建议按以下顺序演进：

1. 保留现有 `TRACE-ID` 响应头，作为旧日志体系兼容字段
2. 在请求中间件创建 server span
3. 为 MySQL、Redis、HTTP client 分别接入 instrumentation
4. 让日志中的 `trace_id` 与 OTel trace id 并行输出一段时间
5. 最后再统一到 OTLP 采集链路

建议环境变量：

```bash
OTEL_ENABLED=false
OTEL_SERVICE_NAME=edu-schedule-system
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=0.1
```
