#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT/scripts/prod_readiness_check.sh"
CFG="$ROOT/configs/pro_configs.toml"

run_expect_fail() {
  local name="$1"
  shift
  local out
  set +e
  out="$($SCRIPT "$CFG" 2>&1)"
  local code=$?
  set -e
  if [[ "$code" -eq 0 ]]; then
    echo "[FAIL] $name: expected failure" >&2
    echo "$out" >&2
    exit 1
  fi
  for expected in "$@"; do
    if ! grep -Fq "$expected" <<< "$out"; then
      echo "[FAIL] $name: missing expected output: $expected" >&2
      echo "$out" >&2
      exit 1
    fi
  done
  echo "[ OK ] $name"
}

base_env() {
  export AUTH_MODE=required
  export AUTH_TEMPLATE_TOKEN_ENABLED=false
  export JWT_SECRET='prod-readiness-secret-value-with-more-than-32-chars'
  export AUTH_REFRESH_TOKEN_TTL_SECONDS=604800
  export AUTH_MIN_PASSWORD_LENGTH=8
  export AUTH_PASSWORD_COMPLEXITY=strong
  export SWAGGER_ENABLED=false
  export PPROF_ENABLED=false
  export LOG_LEVEL=info
  export DEBUG=false
  export CORS_ALLOWED_ORIGINS='https://edu.example.com'
  export AUTH_RECOVERY_DELIVERY_MODE=email
  export AUTH_RECOVERY_DELIVERY_EMAIL_HOST=smtp.example.com
  export AUTH_RECOVERY_DELIVERY_EMAIL_USERNAME=mailer
  export AUTH_RECOVERY_DELIVERY_EMAIL_PASSWORD='smtp-password-value'
  export AUTH_RECOVERY_DELIVERY_EMAIL_FROM=noreply@example.com
  export AUTH_RECOVERY_PREVIEW_ENABLED=false
  export MYSQL_READ_ADDR=prod-read:3306
  export MYSQL_WRITE_ADDR=prod-write:3306
  export MYSQL_READ_USER=edu_read
  export MYSQL_WRITE_USER=edu_write
  export MYSQL_READ_NAME=edu_schedule
  export MYSQL_WRITE_NAME=edu_schedule
  export MYSQL_READ_PASS='read-password-strong'
  export MYSQL_WRITE_PASS='write-password-strong'
  export REDIS_ENABLED=true
  export REDIS_ADDR=redis:6379
  export BACKUP_ENABLED=true
  export BACKUP_RETENTION_DAYS=30
  export BACKUP_DIR=/var/backups/edu-schedule-system
  export LOG_DIR=/var/log/edu-schedule-system
  export ALERT_CHANNEL=feishu
  export ALERT_FEISHU_WEBHOOK='https://open.feishu.cn/open-apis/bot/v2/hook/example'
}

base_env
$SCRIPT "$CFG" >/dev/null
printf '[ OK ] valid production readiness env passes\n'

base_env
export CORS_ALLOWED_ORIGINS='http://localhost:5173'
run_expect_fail "reject localhost cors" "CORS_ALLOWED_ORIGINS must not contain localhost origin"

base_env
export ALERT_CHANNEL=stdout
run_expect_fail "reject stdout alert channel" "production must configure active ALERT_CHANNEL webhook"

base_env
export MYSQL_WRITE_PASS=app
run_expect_fail "reject weak mysql password" "MYSQL_WRITE_PASS must not be empty/weak/default"

base_env
export AUTH_RECOVERY_DELIVERY_MODE=placeholder
run_expect_fail "reject recovery placeholder" "AUTH_RECOVERY_DELIVERY_MODE=placeholder must not be used in production"
