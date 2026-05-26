#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <backup.sql|backup.sql.gz|backup.sql.gz.enc>" >&2
  exit 1
fi

if [[ -z "${MYSQL_DB:-}" || -z "${MYSQL_USER:-}" ]]; then
  echo "MYSQL_DB and MYSQL_USER are required" >&2
  exit 1
fi

export MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
export MYSQL_PORT="${MYSQL_PORT:-3306}"
export MYSQL_CONTAINER="${MYSQL_CONTAINER:-edu-schedule-system-mysql-1}"

SRC="$1"
WORK_FILE="$SRC"
TMP_DIR=""

cleanup() {
  if [[ -n "$TMP_DIR" && -d "$TMP_DIR" ]]; then
    rm -rf "$TMP_DIR"
  fi
}
trap cleanup EXIT

if [[ "$SRC" == *.enc ]]; then
  if [[ -z "${BACKUP_ENCRYPT_KEY:-}" ]]; then
    echo "BACKUP_ENCRYPT_KEY is required to restore encrypted backup" >&2
    exit 1
  fi
  TMP_DIR="$(mktemp -d)"
  DEC_FILE="$TMP_DIR/$(basename "${SRC%.enc}")"
  openssl enc -d -aes-256-cbc -pbkdf2 -pass "pass:${BACKUP_ENCRYPT_KEY}" -in "$SRC" -out "$DEC_FILE"
  WORK_FILE="$DEC_FILE"
fi

if [[ "$WORK_FILE" == *.gz ]]; then
  if [[ -z "$TMP_DIR" ]]; then
    TMP_DIR="$(mktemp -d)"
  fi
  SQL_FILE="$TMP_DIR/$(basename "${WORK_FILE%.gz}")"
  gzip -dc "$WORK_FILE" > "$SQL_FILE"
  WORK_FILE="$SQL_FILE"
fi

python3 "$(dirname "$0")/mysql_backup_restore.py" restore "$WORK_FILE"

echo "restore completed from: $SRC"
