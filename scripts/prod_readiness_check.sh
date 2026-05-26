#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CFG="${1:-$ROOT/configs/pro_configs.toml}"

if [[ ! -f "$CFG" ]]; then
  echo "[FAIL] config file not found: $CFG" >&2
  exit 1
fi

fail=0
warn=0

check_fail() {
  local message="$1"
  echo "[FAIL] $message" >&2
  fail=$((fail + 1))
}

check_warn() {
  local message="$1"
  echo "[WARN] $message" >&2
  warn=$((warn + 1))
}

check_ok() {
  local message="$1"
  echo "[ OK ] $message"
}

require_env() {
  local key="$1"
  if [[ -z "${!key:-}" ]]; then
    check_fail "$key must be set for production readiness"
  else
    check_ok "$key is set"
  fi
}

require_not_true() {
  local key="$1"
  local value="$(printf '%s' "${!key:-}" | tr '[:upper:]' '[:lower:]')"
  if [[ "$value" == "true" || "$value" == "1" || "$value" == "yes" || "$value" == "on" ]]; then
    check_fail "$key must not be enabled in production"
  else
    check_ok "$key is not enabled"
  fi
}

is_true() {
  local value="$(printf '%s' "${1:-}" | tr '[:upper:]' '[:lower:]')"
  [[ "$value" == "true" || "$value" == "1" || "$value" == "yes" || "$value" == "on" ]]
}

weak_secret() {
  local value="$1"
  local min_len="${2:-32}"
  local lower="$(printf '%s' "$value" | tr '[:upper:]' '[:lower:]')"
  [[ ${#value} -lt $min_len ]] && return 0
  case "$lower" in
    app|root|admin|password|123456|12345678|changeme|change-me|secret|default|example|test|demo) return 0 ;;
    *dev-secret*|*local-only*|*change-me*|*changeme*|*example*|*template*|*secret-local*|*please-change*|*default*) return 0 ;;
  esac
  return 1
}

weak_db_password() {
  local value="$1"
  local lower="$(printf '%s' "$value" | tr '[:upper:]' '[:lower:]')"
  [[ -z "$value" ]] && return 0
  case "$lower" in
    app|root|admin|password|123456|12345678|mysql|edu|test|demo|changeme|change-me) return 0 ;;
  esac
  [[ ${#value} -lt 12 ]] && return 0
  return 1
}

require_https_origins() {
  local raw="$1"
  IFS=',' read -ra origins <<< "$raw"
  for origin in "${origins[@]}"; do
    origin="$(printf '%s' "$origin" | xargs)"
    [[ -z "$origin" ]] && continue
    if [[ "$origin" == "*" ]]; then
      check_fail "CORS_ALLOWED_ORIGINS must not contain '*' in production"
    elif [[ "$origin" == http://localhost* || "$origin" == http://127.0.0.1* || "$origin" == http://0.0.0.0* || "$origin" == *localhost* ]]; then
      check_fail "CORS_ALLOWED_ORIGINS must not contain localhost origin: $origin"
    elif [[ "$origin" != https://* ]]; then
      check_fail "CORS_ALLOWED_ORIGINS must use HTTPS origin: $origin"
    fi
  done
}

require_positive_int() {
  local key="$1"
  local min="$2"
  local max="$3"
  local value="${!key:-}"
  if [[ -z "$value" ]]; then
    check_warn "$key is not set; using application default"
    return
  fi
  if ! [[ "$value" =~ ^[0-9]+$ ]]; then
    check_fail "$key must be an integer"
    return
  fi
  if (( value < min || value > max )); then
    check_fail "$key must be between $min and $max, got $value"
  else
    check_ok "$key is within expected range"
  fi
}

require_one_of() {
  local key="$1"
  shift
  local value="${!key:-}"
  local allowed
  for allowed in "$@"; do
    if [[ "$value" == "$allowed" ]]; then
      check_ok "$key=$value"
      return
    fi
  done
  check_fail "$key must be one of: $*; got '${value:-<empty>}'"
}

echo "Production readiness check"
echo "Config: $CFG"
echo

# Authentication security.
require_one_of AUTH_MODE required
require_not_true AUTH_TEMPLATE_TOKEN_ENABLED
require_env JWT_SECRET
if [[ -n "${JWT_SECRET:-}" ]]; then
  if weak_secret "$JWT_SECRET" 32; then
    check_fail "JWT_SECRET must be at least 32 chars and not contain default/example fragments"
  else
    check_ok "JWT_SECRET looks non-default"
  fi
fi
require_positive_int AUTH_REFRESH_TOKEN_TTL_SECONDS 3600 2592000
require_positive_int AUTH_MIN_PASSWORD_LENGTH 8 128
if [[ -n "${AUTH_PASSWORD_COMPLEXITY:-}" && "${AUTH_PASSWORD_COMPLEXITY}" != "strong" ]]; then
  check_fail "AUTH_PASSWORD_COMPLEXITY must be strong in production"
elif [[ -n "${AUTH_PASSWORD_COMPLEXITY:-}" ]]; then
  check_ok "AUTH_PASSWORD_COMPLEXITY=strong"
else
  check_warn "AUTH_PASSWORD_COMPLEXITY not set; ensure application password policy is strong"
fi

# Debug exposure.
require_not_true SWAGGER_ENABLED
require_not_true PPROF_ENABLED
if [[ "${LOG_LEVEL:-info}" == "debug" || "${DEBUG:-false}" == "true" ]]; then
  check_fail "debug log/mode must not be enabled in production"
else
  check_ok "debug log/mode is not enabled"
fi

# CORS.
if [[ -z "${CORS_ALLOWED_ORIGINS:-}" ]]; then
  check_fail "CORS_ALLOWED_ORIGINS must be explicitly set to production HTTPS frontend origin(s)"
else
  require_https_origins "$CORS_ALLOWED_ORIGINS"
  check_ok "CORS_ALLOWED_ORIGINS checked"
fi

# Recovery and notification.
case "${AUTH_RECOVERY_DELIVERY_MODE:-}" in
  email)
    require_env AUTH_RECOVERY_DELIVERY_EMAIL_HOST
    require_env AUTH_RECOVERY_DELIVERY_EMAIL_USERNAME
    require_env AUTH_RECOVERY_DELIVERY_EMAIL_PASSWORD
    require_env AUTH_RECOVERY_DELIVERY_EMAIL_FROM
    ;;
  sms)
    require_env AUTH_RECOVERY_DELIVERY_SMS_PROVIDER
    ;;
  disabled)
    check_warn "AUTH_RECOVERY_DELIVERY_MODE=disabled; account recovery delivery will be unavailable"
    ;;
  "")
    check_fail "AUTH_RECOVERY_DELIVERY_MODE must be set to email, sms, or disabled explicitly"
    ;;
  placeholder)
    check_fail "AUTH_RECOVERY_DELIVERY_MODE=placeholder must not be used in production"
    ;;
  *)
    check_fail "AUTH_RECOVERY_DELIVERY_MODE must be email, sms, or disabled, got '${AUTH_RECOVERY_DELIVERY_MODE}'"
    ;;
esac
require_not_true AUTH_RECOVERY_PREVIEW_ENABLED

# Database and cache.
require_env MYSQL_READ_ADDR
require_env MYSQL_WRITE_ADDR
require_env MYSQL_READ_USER
require_env MYSQL_WRITE_USER
require_env MYSQL_READ_NAME
require_env MYSQL_WRITE_NAME
require_env MYSQL_READ_PASS
require_env MYSQL_WRITE_PASS
if [[ -n "${MYSQL_READ_PASS:-}" ]] && weak_db_password "$MYSQL_READ_PASS"; then
  check_fail "MYSQL_READ_PASS must not be empty/weak/default"
fi
if [[ -n "${MYSQL_WRITE_PASS:-}" ]] && weak_db_password "$MYSQL_WRITE_PASS"; then
  check_fail "MYSQL_WRITE_PASS must not be empty/weak/default"
fi
if [[ "${MYSQL_READ_USER:-}" == "root" || "${MYSQL_WRITE_USER:-}" == "root" ]]; then
  check_fail "MySQL production user must not be root"
fi
if [[ "${REDIS_ENABLED:-true}" == "false" ]]; then
  check_warn "REDIS_ENABLED=false; production must have an explicit architecture exemption"
else
  require_env REDIS_ADDR
fi

# Backup and alerting.
if ! is_true "${BACKUP_ENABLED:-}"; then
  check_fail "BACKUP_ENABLED=true is required in production"
else
  check_ok "BACKUP_ENABLED=true"
fi
require_positive_int BACKUP_RETENTION_DAYS 7 366
require_env BACKUP_DIR
require_env LOG_DIR
case "${ALERT_CHANNEL:-}" in
  feishu|lark)
    require_env ALERT_FEISHU_WEBHOOK
    ;;
  wecom|wechat_work)
    require_env ALERT_WECOM_WEBHOOK
    ;;
  webhook)
    require_env ALERT_WEBHOOK_URL
    ;;
  ""|disabled|none|off|stdout)
    check_fail "production must configure active ALERT_CHANNEL webhook, got '${ALERT_CHANNEL:-<empty>}'"
    ;;
  *)
    check_fail "unsupported ALERT_CHANNEL='${ALERT_CHANNEL}'"
    ;;
esac


# Optional live endpoint checks. These are required for target/CI evidence when PROD_BASE_URL is supplied.
if [[ -n "${PROD_BASE_URL:-}" ]]; then
  base="${PROD_BASE_URL%/}"
  if [[ "$base" != https://* && "$base" != http://127.0.0.1* && "$base" != http://localhost* ]]; then
    check_fail "PROD_BASE_URL must be HTTPS for non-local target checks"
  fi
  if ! command -v curl >/dev/null 2>&1; then
    check_fail "curl is required for PROD_BASE_URL live checks"
  else
    tmp_body="$(mktemp)"
    trap 'rm -f "$tmp_body"' EXIT

    check_endpoint_2xx() {
      local path="$1"
      local name="$2"
      local status
      status="$(curl -sS -o "$tmp_body" -w '%{http_code}' --max-time "${PROD_CHECK_TIMEOUT_SECONDS:-10}" "$base$path" || true)"
      if [[ "$status" =~ ^2[0-9][0-9]$ ]]; then
        check_ok "$name returned HTTP $status"
      else
        check_fail "$name must return 2xx, got HTTP ${status:-curl-error}"
      fi
    }

    check_endpoint_rejects() {
      local method="$1"
      local path="$2"
      local name="$3"
      local status
      status="$(curl -sS -o "$tmp_body" -w '%{http_code}' --max-time "${PROD_CHECK_TIMEOUT_SECONDS:-10}" -X "$method" "$base$path" || true)"
      case "$status" in
        401|403|404|405)
          check_ok "$name is not anonymously usable (HTTP $status)"
          ;;
        000)
          check_fail "$name live check failed to connect"
          ;;
        *)
          check_fail "$name must reject anonymous/template access with 401/403/404/405, got HTTP $status"
          ;;
      esac
    }

    check_endpoint_2xx "/system/health" "PROD_BASE_URL /system/health"
    check_endpoint_2xx "/system/ready" "PROD_BASE_URL /system/ready"
    check_endpoint_rejects GET "/api/v1/admins" "PROD_BASE_URL /api/v1/admins"
    check_endpoint_rejects POST "/api/v1/auth/token" "PROD_BASE_URL /api/v1/auth/token template-token endpoint"
  fi
else
  check_warn "PROD_BASE_URL not set; skipping live target endpoint checks"
fi
echo
if (( fail > 0 )); then
  echo "Production readiness check failed: ${fail} failure(s), ${warn} warning(s)." >&2
  exit 1
fi

echo "Production readiness check passed with ${warn} warning(s)."
