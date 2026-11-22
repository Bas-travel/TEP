# Runbook — MySQL backup & restore

| | Linux / macOS | Windows (PowerShell) |
|---|---|---|
| Backup | `scripts/db-backup.sh` | `powershell -ExecutionPolicy Bypass -File scripts\db-backup.ps1` |
| Restore (safe, scratch DB) | `scripts/db-restore.sh backups/<file>.sql.gz tep_restore_check` | `scripts\db-restore.ps1 -File backups\<file>.sql.gz -Target tep_restore_check` |
| Restore over live DB | add `--yes` | add `-Yes` |

- Dumps are consistent (`--single-transaction`), include routines/triggers/events, gzip-verified and written to `backups/` (git-ignored).
- Restore refuses: an existing non-empty target without `--yes`/`-Yes`, invalid database names, and `APP_ENV=production`.
- Default mode talks to the compose `mysql` service. Set `DB_MODE=direct MYSQL_HOST=… MYSQL_PORT=… MYSQL_USER=… MYSQL_PASSWORD=…` (bash scripts) to reach any other MySQL.
- **Drill (do this before every release):** back up → restore into `tep_restore_check` → compare table row counts → drop the scratch DB.

## Verification status
- Bash scripts: round trip tested on MySQL 8.0 (39 tables / 1,633 rows identical after restore; all three refusal guards exercised).
- PowerShell scripts: **not executed** (no PowerShell/Docker in the authoring environment) — run the drill once on your machine and report any error.
