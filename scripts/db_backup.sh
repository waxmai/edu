#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=alert_notify.sh
source "$SCRIPT_DIR/alert_notify.sh"

if [[ -z "${MYSQL_DB:-}" || -z "${MYSQL_USER:-}" ]]; then
  echo "MYSQL_DB and MYSQL_USER are required" >&2
  exit 1
fi

export MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
export MYSQL_PORT="${MYSQL_PORT:-3306}"
export MYSQL_CONTAINER="${MYSQL_CONTAINER:-edu-schedule-system-mysql-1}"

ENV_NAME="${ENV:-${APP_ENV:-dev}}"
BACKUP_DIR="${BACKUP_DIR:-./output/db-backups}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
BACKUP_MIN_BYTES="${BACKUP_MIN_BYTES:-1024}"
BACKUP_COMPRESS="${BACKUP_COMPRESS:-true}"
BACKUP_ENCRYPT="${BACKUP_ENCRYPT:-false}"
BACKUP_ENCRYPT_KEY="${BACKUP_ENCRYPT_KEY:-}"
BACKUP_HISTORY_FILE="${BACKUP_HISTORY_FILE:-./logs/backup-history.ndjson}"
BACKUP_ALERT_FILE="${BACKUP_ALERT_FILE:-./logs/backup-alerts.ndjson}"
BACKUP_OBJECT_STORAGE_ENABLED="${BACKUP_OBJECT_STORAGE_ENABLED:-false}"
BACKUP_OBJECT_STORAGE_URI="${BACKUP_OBJECT_STORAGE_URI:-}"

START_TS="$(date +%s)"
STAMP="$(date +%Y%m%d-%H%M%S)"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || true)}"
mkdir -p "$BACKUP_DIR" "$(dirname "$BACKUP_HISTORY_FILE")" "$(dirname "$BACKUP_ALERT_FILE")"

BASENAME="${ENV_NAME}-${MYSQL_DB}-${STAMP}"
if [[ -n "$GIT_COMMIT" ]]; then
  BASENAME="${BASENAME}-${GIT_COMMIT}"
fi
RAW_OUT="$BACKUP_DIR/${BASENAME}.sql"
FINAL_OUT="$RAW_OUT"
STATUS="success"
ERROR_MSG=""
SIZE_BYTES=0

json_escape() {
  python3 -c 'import json,sys; print(json.dumps(sys.argv[1], ensure_ascii=False)[1:-1])' "$1"
}

record_history() {
  local ended_ts duration now file_arg
  ended_ts="$(date +%s)"
  duration=$((ended_ts - START_TS))
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  file_arg="${FINAL_OUT:-}"
  if [[ -f "$file_arg" ]]; then
    SIZE_BYTES="$(wc -c < "$file_arg" | tr -d ' ')"
  fi
  printf '{"recordedAt":"%s","status":"%s","env":"%s","database":"%s","file":"%s","sizeBytes":%s,"durationSeconds":%s,"error":"%s"}\n' \
    "$now" "$STATUS" "$(json_escape "$ENV_NAME")" "$(json_escape "$MYSQL_DB")" "$(json_escape "$file_arg")" "${SIZE_BYTES:-0}" "$duration" "$(json_escape "$ERROR_MSG")" >> "$BACKUP_HISTORY_FILE"
}

alert_backup() {
  local kind="$1"
  local msg="$2"
  local now
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '{"recordedAt":"%s","type":"%s","env":"%s","database":"%s","file":"%s","message":"%s"}\n' \
    "$now" "$(json_escape "$kind")" "$(json_escape "$ENV_NAME")" "$(json_escape "$MYSQL_DB")" "$(json_escape "${FINAL_OUT:-}")" "$(json_escape "$msg")" >> "$BACKUP_ALERT_FILE"
  send_alert "critical" "数据库备份告警" "$kind: $msg" "请检查备份脚本、MySQL 连接、磁盘空间、对象存储配置和最近一次成功备份。" || true
}

cleanup_old_backups() {
  if [[ "$BACKUP_RETENTION_DAYS" =~ ^[0-9]+$ ]] && [[ "$BACKUP_RETENTION_DAYS" -gt 0 ]]; then
    find "$BACKUP_DIR" -type f \( -name '*.sql' -o -name '*.sql.gz' -o -name '*.sql.gz.enc' -o -name '*.sql.enc' \) -mtime +"$BACKUP_RETENTION_DAYS" -print -delete >/dev/null 2>&1 || true
  fi
}

trap 'STATUS="failed"; ERROR_MSG="${ERROR_MSG:-backup command failed}"; record_history; alert_backup "backup_failed" "$ERROR_MSG"' ERR

python3 "$(dirname "$0")/mysql_backup_restore.py" backup "$RAW_OUT"

if [[ ! -s "$RAW_OUT" ]]; then
  ERROR_MSG="backup file is empty or missing"
  false
fi

FINAL_OUT="$RAW_OUT"
if [[ "$BACKUP_COMPRESS" == "true" ]]; then
  gzip -f "$RAW_OUT"
  FINAL_OUT="${RAW_OUT}.gz"
fi

if [[ "$BACKUP_ENCRYPT" == "true" ]]; then
  if [[ -z "$BACKUP_ENCRYPT_KEY" ]]; then
    ERROR_MSG="BACKUP_ENCRYPT_KEY is required when BACKUP_ENCRYPT=true"
    false
  fi
  openssl enc -aes-256-cbc -salt -pbkdf2 -pass "pass:${BACKUP_ENCRYPT_KEY}" -in "$FINAL_OUT" -out "${FINAL_OUT}.enc"
  rm -f "$FINAL_OUT"
  FINAL_OUT="${FINAL_OUT}.enc"
fi

SIZE_BYTES="$(wc -c < "$FINAL_OUT" | tr -d ' ')"
if [[ "$SIZE_BYTES" -lt "$BACKUP_MIN_BYTES" ]]; then
  ERROR_MSG="backup file too small: ${SIZE_BYTES} bytes < ${BACKUP_MIN_BYTES} bytes"
  alert_backup "backup_file_too_small" "$ERROR_MSG"
  false
fi

if [[ "$BACKUP_OBJECT_STORAGE_ENABLED" == "true" ]]; then
  if [[ -z "$BACKUP_OBJECT_STORAGE_URI" ]]; then
    ERROR_MSG="BACKUP_OBJECT_STORAGE_URI is required when BACKUP_OBJECT_STORAGE_ENABLED=true"
    alert_backup "backup_upload_failed" "$ERROR_MSG"
    false
  fi
  if command -v aws >/dev/null 2>&1 && [[ "$BACKUP_OBJECT_STORAGE_URI" == s3://* ]]; then
    aws s3 cp "$FINAL_OUT" "$BACKUP_OBJECT_STORAGE_URI/"
  elif command -v rclone >/dev/null 2>&1; then
    rclone copy "$FINAL_OUT" "$BACKUP_OBJECT_STORAGE_URI"
  else
    ERROR_MSG="object storage upload requires aws cli for s3:// or rclone"
    alert_backup "backup_upload_failed" "$ERROR_MSG"
    false
  fi
fi

cleanup_old_backups
record_history

echo "backup created: $FINAL_OUT"
