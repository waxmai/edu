#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATION_DIR="${MIGRATION_DIR:-$ROOT/migrations}"

if [[ ! -d "$MIGRATION_DIR" ]]; then
  echo "[FAIL] migration directory not found: $MIGRATION_DIR" >&2
  exit 1
fi

fail=0
seen=""
prev=""

while IFS= read -r file; do
  base="$(basename "$file")"
  if ! [[ "$base" =~ ^[0-9]{6}_.+\.sql$ ]]; then
    echo "[FAIL] migration filename must match NNNNNN_name.sql: $base" >&2
    fail=$((fail + 1))
    continue
  fi
  seq="${base%%_*}"
  if grep -q " $seq " <<< " $seen "; then
    echo "[FAIL] duplicate migration sequence: $seq" >&2
    fail=$((fail + 1))
  fi
  seen="$seen $seq"
  if [[ -n "$prev" && "$seq" < "$prev" ]]; then
    echo "[FAIL] migration sequence out of order: $base after $prev" >&2
    fail=$((fail + 1))
  fi
  prev="$seq"
  if [[ ! -s "$file" ]]; then
    echo "[FAIL] migration file is empty: $base" >&2
    fail=$((fail + 1))
  fi
  if grep -Eiq 'DROP[[:space:]]+DATABASE|DROP[[:space:]]+SCHEMA|TRUNCATE[[:space:]]+TABLE' "$file"; then
    echo "[FAIL] destructive migration statement requires explicit review: $base" >&2
    fail=$((fail + 1))
  fi
  echo "[ OK ] $base"
done < <(find "$MIGRATION_DIR" -maxdepth 1 -type f -name '*.sql' | sort)

if (( fail > 0 )); then
  echo "migration lint failed: $fail issue(s)" >&2
  exit 1
fi

echo "migration lint passed"
