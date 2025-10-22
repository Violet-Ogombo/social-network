#!/usr/bin/env bash
set -euo pipefail

DB_FILE=${DB_PATH:-/app/backend/socialnetwork.db}
MIGRATIONS_DIR=/migrations

echo "Entrypoint: ensuring DB file exists at $DB_FILE"
mkdir -p "$(dirname "$DB_FILE")"
if [ ! -f "$DB_FILE" ]; then
  echo "Creating new SQLite DB file"
  sqlite3 "$DB_FILE" ".databases"
fi

echo "Applying migrations from $MIGRATIONS_DIR"
if [ -d "$MIGRATIONS_DIR" ]; then
  for f in $(ls -1 "$MIGRATIONS_DIR"/*up.sql 2>/dev/null | sort); do
    echo "Applying migration: $f"
    sqlite3 "$DB_FILE" < "$f" || { echo "Migration failed: $f"; exit 1; }
  done
else
  echo "No migrations directory found at $MIGRATIONS_DIR"
fi

echo "Starting application"
exec "$@"
