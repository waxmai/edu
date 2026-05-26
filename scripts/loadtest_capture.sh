#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

STAMP="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="output/loadtest-reports"
OUT_FILE="$OUT_DIR/loadtest-$STAMP.jsonl"
mkdir -p "$OUT_DIR"

TMP_OUTPUT="$(mktemp)"
cleanup() {
  rm -f "$TMP_OUTPUT"
}
trap cleanup EXIT

./scripts/loadtest_core_api.sh | tee "$TMP_OUTPUT"

python3 - <<'PY' "$TMP_OUTPUT" "$OUT_FILE"
import json, sys, pathlib, datetime
src = pathlib.Path(sys.argv[1])
out = pathlib.Path(sys.argv[2])
lines = []
for raw in src.read_text().splitlines():
    raw = raw.strip()
    if not raw or not raw.startswith('{'):
        continue
    obj = json.loads(raw)
    obj['capturedAt'] = datetime.datetime.now().isoformat()
    lines.append(obj)
out.write_text("\n".join(json.dumps(x, ensure_ascii=False) for x in lines) + ("\n" if lines else ""))
print(out)
PY

if grep -q '"status": 0' "$OUT_FILE"; then
  echo "Loadtest report written to $OUT_FILE (contains transport failures, check API availability)"
  exit 1
fi

echo "Loadtest report written to $OUT_FILE"
