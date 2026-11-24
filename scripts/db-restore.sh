#!/usr/bin/env bash
# Restore a dump created by scripts/db-backup.sh.
#   scripts/db-restore.sh backups/tep-YYYYmmdd-HHMMSS.sql.gz tep_restore_check          # into a NEW/scratch database (safe)
#   scripts/db-restore.sh backups/tep-....sql.gz tep --yes                                # overwrite the live database
# Restoring over an existing database that has tables requires --yes. Refused when APP_ENV=production.
set -euo pipefail
cd "$(dirname "$0")/.."
[ -f .env ] && { set -a; . ./.env; set +a; }
FILE="${1:?usage: db-restore.sh <dump.sql.gz> <target-database> [--yes]}"; TARGET="${2:?target database required}"; YES="${3:-}"
[ -f "$FILE" ] || { echo "no such file: $FILE" >&2; exit 1; }
[[ "$TARGET" =~ ^[A-Za-z0-9_]+$ ]] || { echo "invalid database name" >&2; exit 1; }
[ "${APP_ENV:-dev}" != production ] || { echo "refusing to restore while APP_ENV=production" >&2; exit 1; }
gzip -t "$FILE"
MODE="${DB_MODE:-compose}"
run_sql() { # reads SQL from stdin
  if [ "$MODE" = compose ]; then
    docker compose exec -T -e MYSQL_PWD="${MYSQL_ROOT_PASSWORD:?}" mysql mysql -uroot "$@"
  else
    MYSQL_PWD="${MYSQL_PASSWORD:-}" mysql -h"${MYSQL_HOST:-127.0.0.1}" -P"${MYSQL_PORT:-3306}" -u"${MYSQL_USER:-root}" "$@"
  fi
}
EXISTING=$(echo "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$TARGET'" | run_sql -N)
if [ "$EXISTING" -gt 0 ] && [ "$YES" != --yes ]; then
  echo "database '$TARGET' already has $EXISTING tables; re-run with --yes to overwrite it" >&2; exit 1
fi
echo "CREATE DATABASE IF NOT EXISTS \`$TARGET\` CHARACTER SET utf8mb4" | run_sql
gunzip -c "$FILE" | run_sql "$TARGET"
echo "restore ok: $FILE -> $TARGET ($(echo "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$TARGET'" | run_sql -N) tables)"
