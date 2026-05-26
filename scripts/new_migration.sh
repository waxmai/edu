#!/bin/bash
set -euo pipefail

table="${1:-}"
name="${2:-}"

if [[ -z "$table" ]]; then
  printf "Usage: %s <table> [migration_name]\n" "$0" >&2
  exit 1
fi

slug_source="${name:-create_${table}}"
slug="$(printf "%s" "$slug_source" | tr '[:upper:]' '[:lower:]' | tr -c 'a-z0-9_' '_')"
slug="${slug##_}"
slug="${slug%%_}"
version="$(date +%Y%m%d%H%M%S)"
path="migrations/${version}_${slug}.sql"

cat > "$path" <<EOF
-- rollback-policy: forward-only

-- TODO: define schema changes for ${table}.
-- Example:
-- CREATE TABLE IF NOT EXISTS ${table} (
--   id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
--   created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
--   updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
EOF

printf "Created %s\n" "$path"
printf "Fill in the SQL, then run: make gen-module TABLE=%s\n" "$table"
