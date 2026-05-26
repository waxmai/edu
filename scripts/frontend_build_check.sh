#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="${FRONTEND_DIR:-$ROOT/frontend/web}"

if [[ ! -f "$FRONTEND_DIR/package.json" ]]; then
  echo "[SKIP] frontend package.json not found: $FRONTEND_DIR/package.json"
  exit 0
fi

cd "$FRONTEND_DIR"
if [[ -f package-lock.json ]]; then
  npm ci
else
  npm install
fi
npm run build
