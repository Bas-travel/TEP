#!/usr/bin/env bash
# Creates .env from .env.example with cryptographically random secrets.
# Usage: scripts/setup-env.sh [--force]      (refuses to overwrite an existing .env without --force)
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ -f .env && "${1:-}" != "--force" ]]; then
  echo ".env already exists — nothing changed (use --force to regenerate; this also invalidates existing DB volumes' passwords)." >&2
  exit 1
fi

# Alphanumeric only: safe inside the MySQL DSN, docker compose interpolation and shell quoting.
gen() { LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c "$1"; }

umask 077
sed \
  -e "s/^JWT_SECRET=__GENERATE__/JWT_SECRET=$(gen 64)/" \
  -e "s/^MYSQL_PASSWORD=__GENERATE__/MYSQL_PASSWORD=$(gen 32)/" \
  -e "s/^MYSQL_ROOT_PASSWORD=__GENERATE__/MYSQL_ROOT_PASSWORD=$(gen 32)/" \
  -e "s/^MINIO_ROOT_USER=__GENERATE__/MINIO_ROOT_USER=tep$(gen 10)/" \
  -e "s/^MINIO_ROOT_PASSWORD=__GENERATE__/MINIO_ROOT_PASSWORD=$(gen 32)/" \
  .env.example > .env

echo "Created .env with generated secrets (git-ignored)."
echo "Next:  docker compose up --build   then   curl http://localhost:8080/health/ready"
