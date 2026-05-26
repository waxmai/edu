#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
WEB_DIR="$ROOT_DIR/frontend/web"
PORT="${1:-5173}"

cd "$WEB_DIR"

echo "[frontend-reset] web dir: $WEB_DIR"
echo "[frontend-reset] target port: $PORT"

echo "[frontend-reset] removing generated shadow js files under src/ ..."
find src \( -name '*.js' -o -name '*.js.map' -o -name '*.vue.js' \) -delete

echo "[frontend-reset] clearing vite cache ..."
rm -rf node_modules/.vite dist

echo "[frontend-reset] killing existing vite/node listeners on :$PORT ..."
PIDS="$(lsof -tiTCP:${PORT} -sTCP:LISTEN || true)"
if [[ -n "$PIDS" ]]; then
  kill $PIDS || true
  sleep 1
fi

PIDS_AFTER="$(lsof -tiTCP:${PORT} -sTCP:LISTEN || true)"
if [[ -n "$PIDS_AFTER" ]]; then
  echo "[frontend-reset] force killing remaining listeners on :$PORT ..."
  kill -9 $PIDS_AFTER || true
  sleep 1
fi

if lsof -tiTCP:${PORT} -sTCP:LISTEN >/dev/null 2>&1; then
  echo "[frontend-reset] port $PORT is still occupied, aborting" >&2
  exit 1
fi

echo "[frontend-reset] starting vite on :$PORT ..."
npm run dev -- --host 127.0.0.1 --port "$PORT"
