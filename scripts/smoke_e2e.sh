#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://127.0.0.1:9999/api/v1}"
SMOKE_PREPARE_USERS="${SMOKE_PREPARE_USERS:-none}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required" >&2
  exit 1
fi

cat <<'EOF' >&2
[smoke-e2e] tip:
- 默认会尝试使用 platform_admin / teacher01 与你当前提供的账号密码登录
- 如果当前 dev 库密码已轮换，请显式传入：
  - SMOKE_ADMIN_TOKEN / SMOKE_ADMIN_REFRESH_TOKEN
  - 或 SMOKE_ADMIN_USERNAME / SMOKE_ADMIN_PASSWORD
  - 同理可传 SMOKE_TEACHER_TOKEN 或 teacher 账号密码
- 如果是在本地受控 dev 环境，且明确允许重置 bootstrap 账号口令，可设置：
  - SMOKE_PREPARE_USERS=bootstrap
  - 优先提供 4 个 BOOTSTRAP_*_PASSWORD
  - 或兼容提供 4 个 BOOTSTRAP_*_PASSWORD_HASH
  这会先执行一次 `make bootstrap`，把种子账号恢复到你当前显式注入的初始状态。
EOF

if [[ "$SMOKE_PREPARE_USERS" == "bootstrap" ]]; then
  if [[ -z "${BOOTSTRAP_PLATFORM_ADMIN_PASSWORD:-}" && -z "${BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH:-}" ]]; then
    echo "BOOTSTRAP_PLATFORM_ADMIN_PASSWORD or BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH is required when SMOKE_PREPARE_USERS=bootstrap" >&2
    exit 1
  fi
  if [[ -z "${BOOTSTRAP_ORG_ADMIN_PASSWORD:-}" && -z "${BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH:-}" ]]; then
    echo "BOOTSTRAP_ORG_ADMIN_PASSWORD or BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH is required when SMOKE_PREPARE_USERS=bootstrap" >&2
    exit 1
  fi
  if [[ -z "${BOOTSTRAP_CAMPUS_ADMIN_PASSWORD:-}" && -z "${BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH:-}" ]]; then
    echo "BOOTSTRAP_CAMPUS_ADMIN_PASSWORD or BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH is required when SMOKE_PREPARE_USERS=bootstrap" >&2
    exit 1
  fi
  if [[ -z "${BOOTSTRAP_TEACHER_PASSWORD:-}" && -z "${BOOTSTRAP_TEACHER_PASSWORD_HASH:-}" ]]; then
    echo "BOOTSTRAP_TEACHER_PASSWORD or BOOTSTRAP_TEACHER_PASSWORD_HASH is required when SMOKE_PREPARE_USERS=bootstrap" >&2
    exit 1
  fi
  echo "[smoke-e2e] preparing bootstrap users via make bootstrap ..." >&2
  make bootstrap \
    BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH="$BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH" \
    BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH="$BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH" \
    BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH="$BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH" \
    BOOTSTRAP_TEACHER_PASSWORD_HASH="$BOOTSTRAP_TEACHER_PASSWORD_HASH" \
    BOOTSTRAP_PLATFORM_ADMIN_PASSWORD="$BOOTSTRAP_PLATFORM_ADMIN_PASSWORD" \
    BOOTSTRAP_ORG_ADMIN_PASSWORD="$BOOTSTRAP_ORG_ADMIN_PASSWORD" \
    BOOTSTRAP_CAMPUS_ADMIN_PASSWORD="$BOOTSTRAP_CAMPUS_ADMIN_PASSWORD" \
    BOOTSTRAP_TEACHER_PASSWORD="$BOOTSTRAP_TEACHER_PASSWORD" >&2
fi

API_BASE_URL="$API_BASE_URL" python3 "$(dirname "$0")/smoke_e2e.py"
