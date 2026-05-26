#!/usr/bin/env bash
set -euo pipefail

kind="${1:-manual}"
case "$kind" in
  manual)
    ./scripts/alert_notify.sh warning "手动测试告警" "manual alert test" "请确认飞书/企微告警通道是否收到消息。"
    ;;
  backup)
    ./scripts/alert_notify.sh critical "数据库备份失败" "manual backup failure alert test" "请检查备份脚本、MySQL 连接、磁盘空间和对象存储配置。"
    ;;
  disk)
    ./scripts/alert_notify.sh critical "磁盘空间不足" "manual disk space alert test" "请清理日志/备份归档，确认磁盘扩容和日志轮转策略。"
    ;;
  notification)
    ./scripts/alert_notify.sh warning "通知发送失败异常升高" "manual notification failure spike alert test" "请检查短信/邮件供应商状态、凭据、限流和模板审核状态。"
    ;;
  recovery)
    ./scripts/alert_notify.sh warning "密码恢复发送失败异常升高" "manual recovery delivery failure spike alert test" "请检查密码恢复通道配置、供应商限流、模板和目标地址有效性。"
    ;;
  ready)
    ./scripts/alert_notify.sh critical "服务就绪检查连续失败" "manual /system/ready failure alert test" "请检查 MySQL/Redis 探活、网络、凭据和最近发布变更。"
    ;;
  http5xx)
    ./scripts/alert_notify.sh critical "HTTP 5xx 比例超过阈值" "manual HTTP 5xx ratio alert test" "请检查应用错误日志、panic/recover 告警、数据库/缓存探活和最近发布。"
    ;;
  p95)
    ./scripts/alert_notify.sh warning "p95 接口耗时超过阈值" "manual p95 latency alert test" "请检查慢 SQL、Redis 延迟、外部供应商耗时和实例负载。"
    ;;
  login)
    ./scripts/alert_notify.sh warning "登录失败次数异常升高" "manual login failure spike alert test" "请检查是否存在撞库/暴力破解、异常 IP、账号锁定策略和认证服务日志。"
    ;;
  panic)
    ./scripts/alert_notify.sh critical "服务 panic / recover 事件" "manual panic/recover alert test" "请立即查看错误堆栈、请求参数和最近发布变更。"
    ;;
  *)
    echo "usage: $0 {manual|backup|disk|notification|recovery|ready|http5xx|p95|login|panic}" >&2
    exit 1
    ;;
esac
