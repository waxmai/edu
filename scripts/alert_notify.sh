#!/usr/bin/env bash
set -euo pipefail

PROJECT_NAME="${PROJECT_NAME:-edu-schedule-system}"
APP_ENV="${APP_ENV:-${ENV:-unknown}}"
ALERT_CHANNEL="${ALERT_CHANNEL:-disabled}"
ALERT_WEBHOOK_TIMEOUT_SECONDS="${ALERT_WEBHOOK_TIMEOUT_SECONDS:-3}"

json_escape() {
  python3 -c 'import json,sys; print(json.dumps(sys.argv[1], ensure_ascii=False)[1:-1])' "$1"
}

render_text() {
  local severity="$1"
  local title="$2"
  local message="$3"
  local suggestion="$4"
  local now
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '[%s] %s\n项目：%s\n环境：%s\n时间：%s\n内容：%s\n建议：%s' \
    "$(printf '%s' "$severity" | tr '[:lower:]' '[:upper:]')" "$title" "$PROJECT_NAME" "$APP_ENV" "$now" "$message" "$suggestion"
}

send_payload() {
  local payload="$1"
  local url="$2"
  if [[ -z "$url" ]]; then
    return 0
  fi
  curl -fsS --max-time "$ALERT_WEBHOOK_TIMEOUT_SECONDS" -H 'Content-Type: application/json' -d "$payload" "$url" >/dev/null
}

send_alert() {
  local severity="$1"
  local title="$2"
  local message="$3"
  local suggestion="$4"
  local text payload
  text="$(render_text "$severity" "$title" "$message" "$suggestion")"
  case "$ALERT_CHANNEL" in
    disabled|none|off|'')
      return 0
      ;;
    stdout)
      printf 'ALERT %s\n' "$text"
      ;;
    feishu|lark)
      payload="{\"msg_type\":\"text\",\"content\":{\"text\":\"$(json_escape "$text")\"}}"
      send_payload "$payload" "${ALERT_FEISHU_WEBHOOK:-}"
      ;;
    wecom|wechat_work)
      payload="{\"msgtype\":\"text\",\"text\":{\"content\":\"$(json_escape "$text")\"}}"
      send_payload "$payload" "${ALERT_WECOM_WEBHOOK:-}"
      ;;
    webhook)
      payload="{\"severity\":\"$(json_escape "$severity")\",\"title\":\"$(json_escape "$title")\",\"message\":\"$(json_escape "$message")\",\"text\":\"$(json_escape "$text")\"}"
      send_payload "$payload" "${ALERT_WEBHOOK_URL:-}"
      ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  send_alert "${1:-warning}" "${2:-测试告警}" "${3:-manual alert test}" "${4:-请确认告警通道配置和接收群。}"
fi
