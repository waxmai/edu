#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HISTORY_FILE="${BACKUP_HISTORY_FILE:-$ROOT/logs/backup-history.ndjson}"
MAX_HOURS="${BACKUP_RELEASE_MAX_HOURS:-26}"
REQUIRE_OBJECT_STORAGE="${BACKUP_RELEASE_REQUIRE_OBJECT_STORAGE:-false}"

if [[ ! -f "$HISTORY_FILE" ]]; then
  echo "[FAIL] backup history not found: $HISTORY_FILE" >&2
  exit 1
fi

python3 - "$HISTORY_FILE" "$MAX_HOURS" "$REQUIRE_OBJECT_STORAGE" <<'PY'
import json, pathlib, sys, time
path = pathlib.Path(sys.argv[1])
max_hours = float(sys.argv[2])
require_object_storage = sys.argv[3].lower() in {"1", "true", "yes", "on"}
latest = None
latest_row = None
for line in path.read_text().splitlines():
    try:
        row = json.loads(line)
    except Exception:
        continue
    if row.get("status") != "success":
        continue
    value = row.get("recordedAt", "")
    try:
        ts = time.mktime(time.strptime(value, "%Y-%m-%dT%H:%M:%SZ"))
    except Exception:
        try:
            ts = time.mktime(time.strptime(value[:19] + "Z", "%Y-%m-%dT%H:%M:%SZ"))
        except Exception:
            continue
    if latest is None or ts > latest:
        latest = ts
        latest_row = row
if latest is None:
    print("[FAIL] no successful backup record found", file=sys.stderr)
    raise SystemExit(1)
age_hours = (time.time() - latest) / 3600
if age_hours > max_hours:
    print(f"[FAIL] latest successful backup is too old: {age_hours:.2f}h > {max_hours:.2f}h", file=sys.stderr)
    raise SystemExit(1)
file_path = latest_row.get("file", "")
size = int(latest_row.get("sizeBytes") or 0)
if size <= 0:
    print("[FAIL] latest successful backup has invalid size", file=sys.stderr)
    raise SystemExit(1)
if require_object_storage and not latest_row.get("objectStorageUploaded"):
    print("[FAIL] latest successful backup has no object storage confirmation", file=sys.stderr)
    raise SystemExit(1)
print(f"[ OK ] latest successful backup age={age_hours:.2f}h size={size} file={file_path}")
PY
