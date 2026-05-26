#!/usr/bin/env bash
set -euo pipefail

LOG_FILE="${LOG_FILE:-/tmp/edu-schedule-system-server.log}"
PORT="${PORT:-9999}"

./edu-schedule-system -env dev >"${LOG_FILE}" 2>&1 &
PID=$!
trap 'kill ${PID} >/dev/null 2>&1 || true' EXIT

for _ in $(seq 1 50); do
  if curl -fsS "http://127.0.0.1:${PORT}/system/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.2
done

curl -fsS "http://127.0.0.1:${PORT}/system/health" >/dev/null
curl -fsS "http://127.0.0.1:${PORT}/system/ready" >/dev/null

echo "server smoke passed"
