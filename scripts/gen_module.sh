#!/bin/bash
set -euo pipefail

table="${1:-}"
dsn="${2:-}"
tables="${3:-}"
apply_migrations="${4:-true}"

if [[ -z "$table" ]]; then
  printf "Usage: %s <table> [dsn] [tables] [apply_migrations]\n" "$0" >&2
  exit 1
fi

if [[ -z "$tables" ]]; then
  tables="$table"
fi

case "$apply_migrations" in
  true|1|yes|y)
    printf "Applying migrations and bootstrap data...\n"
    go run cmd/migrate/main.go -env dev -bootstrap
    ;;
  false|0|no|n)
    printf "Skipping migrations because APPLY_MIGRATIONS=%s\n" "$apply_migrations"
    ;;
  *)
    printf "APPLY_MIGRATIONS must be true or false, got %s\n" "$apply_migrations" >&2
    exit 1
    ;;
esac

if [[ -n "$dsn" ]]; then
  printf "Generating DAO/model for tables: %s\n" "$tables"
  go run cmd/gormgen/main.go -dsn "$dsn" -tables "$tables"
else
  printf "Skipping DAO/model generation because DSN is empty.\n"
fi

printf "Generating HTTP handler/router for table: %s\n" "$table"
go run cmd/handlergen/main.go -table "$table"

printf "Regenerating Wire graph...\n"
go run github.com/google/wire/cmd/wire ./cmd/server

printf "Regenerating Swagger docs...\n"
./scripts/swagger.sh

printf "Module generation chain completed for %s.\n" "$table"
printf "Review router registration, Wire providers, service implementation, DTOs, and tests before committing.\n"
