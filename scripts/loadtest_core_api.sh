#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://127.0.0.1:9999/api/v1}"
LOGIN_USERNAME="${LOGIN_USERNAME:-admin}"
LOGIN_PASSWORD="${LOGIN_PASSWORD:-}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required" >&2
  exit 1
fi

if [[ -z "${AUTH_TOKEN:-}" ]]; then
  if ! AUTH_TOKEN="$(API_BASE_URL="$API_BASE_URL" LOGIN_USERNAME="$LOGIN_USERNAME" LOGIN_PASSWORD="$LOGIN_PASSWORD" python3 - <<'PY'
import json
import os
import sys
import urllib.request
import urllib.error

base = os.environ['API_BASE_URL'].rstrip('/')
username = os.environ['LOGIN_USERNAME']
password = os.environ['LOGIN_PASSWORD']
url = base + '/auth/login'
payload = json.dumps({'username': username, 'password': password}).encode('utf-8')
req = urllib.request.Request(url, data=payload, method='POST', headers={'Content-Type': 'application/json'})
try:
    with urllib.request.urlopen(req, timeout=10) as resp:
        body = json.loads(resp.read().decode('utf-8'))
except urllib.error.HTTPError as e:
    detail = e.read().decode('utf-8')
    sys.stderr.write(detail + '\n')
    raise SystemExit(
        'auto login failed, set AUTH_TOKEN directly or provide LOGIN_USERNAME/LOGIN_PASSWORD for the current dev database state'
    )

data = body.get('data') or {}
token = data.get('token')
if not token:
    sys.stderr.write(json.dumps(body, ensure_ascii=False) + '\n')
    raise SystemExit('login response missing token; set AUTH_TOKEN directly if this environment requires a pre-rotated account')
print(token)
PY
)"; then
    exit 1
  fi
fi

if [[ -z "${AUTH_TOKEN:-}" ]]; then
  echo "AUTH_TOKEN is required" >&2
  exit 1
fi

sample_check() {
  python3 - <<'PY'
import json
import os
import sys
import urllib.request
import urllib.error

base = os.environ['API_BASE_URL'].rstrip('/')
headers = {
    'Authorization': f"Bearer {os.environ['AUTH_TOKEN'].strip()}",
    'Content-Type': 'application/json',
}
checks = [
    (f"/payment-records/{os.environ['LOADTEST_PAYMENT_ID']}", 'payment'),
    (f"/schedules/{os.environ['LOADTEST_SCHEDULE_ID']}", 'schedule'),
]
failed = []
for path, label in checks:
    req = urllib.request.Request(base + path, method='GET', headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            if resp.status != 200:
                failed.append(f'{label}:{resp.status}')
    except urllib.error.HTTPError as e:
        failed.append(f'{label}:{e.code}')
    except Exception as e:
        failed.append(f'{label}:{e}')
if failed:
    sys.stderr.write('loadtest sample check failed: ' + ', '.join(failed) + '\n')
    raise SystemExit(1)
PY
}

export API_BASE_URL AUTH_TOKEN
export LOADTEST_THREADS="${LOADTEST_THREADS:-8}"
export LOADTEST_ITERATIONS="${LOADTEST_ITERATIONS:-10}"
export LOADTEST_SCHEDULE_ID="${LOADTEST_SCHEDULE_ID:-2}"
export LOADTEST_PAYMENT_ID="${LOADTEST_PAYMENT_ID:-2}"
export LOADTEST_LESSON_PACKAGE_ID="${LOADTEST_LESSON_PACKAGE_ID:-1}"
export LOADTEST_STUDENT_ID="${LOADTEST_STUDENT_ID:-1}"
export LOADTEST_COURSE_ID="${LOADTEST_COURSE_ID:-1}"
export LOADTEST_TEACHER_ID="${LOADTEST_TEACHER_ID:-1}"

sample_check

python3 "$(dirname "$0")/loadtest_core_api.py"
