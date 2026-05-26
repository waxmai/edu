#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=alert_notify.sh
source "$SCRIPT_DIR/alert_notify.sh"

HISTORY_FILE="${BACKUP_HISTORY_FILE:-./logs/backup-history.ndjson}"
ALERT_FILE="${BACKUP_ALERT_FILE:-./logs/backup-alerts.ndjson}"
MAX_HOURS="${BACKUP_MAX_HOURS_WITHOUT_SUCCESS:-26}"
MIN_FREE_MB="${BACKUP_MIN_FREE_MB:-1024}"
BACKUP_DIR="${BACKUP_DIR:-./output/db-backups}"

mkdir -p "$(dirname "$ALERT_FILE")" "$BACKUP_DIR"

json_escape() {
  python3 -c 'import json,sys; print(json.dumps(sys.argv[1], ensure_ascii=False)[1:-1])' "$1"
}

alert() {
  local kind="$1"
  local msg="$2"
  local now
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '{"recordedAt":"%s","type":"%s","message":"%s"}\n' "$now" "$(json_escape "$kind")" "$(json_escape "$msg")" >> "$ALERT_FILE"
  send_alert "critical" "数据库备份健康检查告警" "$kind: $msg" "请检查备份历史、定时任务、磁盘空间和对象存储同步状态。" || true
  echo "ALERT $kind: $msg" >&2
}

status="$(python3 - "$HISTORY_FILE" "$MAX_HOURS" <<'PY'
import json, pathlib, sys, time
path = pathlib.Path(sys.argv[1])
max_hours = float(sys.argv[2])
if not path.exists():
    print('NO_SUCCESS')
    raise SystemExit(0)
latest = None
for line in path.read_text().splitlines():
    try:
        row = json.loads(line)
    except Exception:
        continue
    if row.get('status') != 'success':
        continue
    value = row.get('recordedAt', '')
    try:
        ts = time.mktime(time.strptime(value, '%Y-%m-%dT%H:%M:%SZ'))
    except Exception:
        continue
    latest = max(latest or ts, ts)
if latest is None:
    print('NO_SUCCESS')
elif time.time() - latest > max_hours * 3600:
    print('STALE_SUCCESS')
else:
    print('OK')
PY
)"
case "$status" in
  OK) ;;
  NO_SUCCESS) alert "backup_no_success" "no successful backup history found" ;;
  STALE_SUCCESS) alert "backup_stale_success" "no successful backup within ${MAX_HOURS} hours" ;;
esac

free_mb="$(df -Pm "$BACKUP_DIR" | awk 'NR==2 {print $4}')"
if [[ "$free_mb" =~ ^[0-9]+$ ]] && [[ "$free_mb" -lt "$MIN_FREE_MB" ]]; then
  alert "backup_disk_space_low" "backup directory free space ${free_mb}MB < ${MIN_FREE_MB}MB"
fi
