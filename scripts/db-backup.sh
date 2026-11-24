#!/usr/bin/env bash
# MySQL backup for the TEP stack. Writes a consistent, compressed logical dump to ./backups/.
#   scripts/db-backup.sh                 # docker compose stack (service "mysql")
#   DB_MODE=direct MYSQL_HOST=127.0.0.1 MYSQL_PORT=3307 scripts/db-backup.sh   # any reachable MySQL
# Reads MYSQL_DATABASE / MYSQL_USER / MYSQL_PASSWORD / MYSQL_ROOT_PASSWORD from the environment or ./.env.
# Never prints secrets. Read-only against the database.
set -euo pipefail
cd "$(dirname "$0")/.."
[ -f .env ] && { set -a; . ./.env; set +a; }
DB="${MYSQL_DATABASE:-tep}"; MODE="${DB_MODE:-compose}"
OUT_DIR="${BACKUP_DIR:-backups}"; mkdir -p "$OUT_DIR"
FILE="$OUT_DIR/${DB}-$(date +%Y%m%d-%H%M%S).sql.gz"
DUMP_OPTS=(--single-transaction --routines --triggers --events --hex-blob --set-gtid-purged=OFF --no-tablespaces "$DB")
if [ "$MODE" = compose ]; then
  : "${MYSQL_ROOT_PASSWORD:?MYSQL_ROOT_PASSWORD missing (run scripts/setup-env.sh)}"
  docker compose exec -T -e MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysql mysqldump -uroot "${DUMP_OPTS[@]}" | gzip -9 > "$FILE"
else
  MYSQL_PWD="${MYSQL_PASSWORD:-}" mysqldump -h"${MYSQL_HOST:-127.0.0.1}" -P"${MYSQL_PORT:-3306}" -u"${MYSQL_USER:-root}" "${DUMP_OPTS[@]}" | gzip -9 > "$FILE"
fi
gzip -t "$FILE"                       # integrity check of the archive itself
BYTES=$(wc -c < "$FILE")
[ "$BYTES" -gt 1024 ] || { echo "backup suspiciously small ($BYTES bytes): $FILE" >&2; exit 1; }
echo "backup ok: $FILE ($BYTES bytes)"
