<#
 Restore a dump made by db-backup.ps1 into a database of the compose MySQL.
 Usage:  scripts\db-restore.ps1 -File backups\tep-....sql.gz -Target tep_restore_check        # scratch DB (safe)
         scripts\db-restore.ps1 -File backups\tep-....sql.gz -Target tep -Yes                  # overwrite live DB
#>
param([Parameter(Mandatory)][string]$File, [Parameter(Mandatory)][string]$Target, [switch]$Yes)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path -Parent $PSScriptRoot)
if ($Target -notmatch '^[A-Za-z0-9_]+$') { throw 'invalid database name' }
if (-not (Test-Path $File)) { throw "no such file: $File" }
$envMap = @{}
Get-Content .env | Where-Object { $_ -match '^\s*[A-Za-z_][A-Za-z0-9_]*=' } | ForEach-Object {
  $k, $v = $_ -split '=', 2; $envMap[$k.Trim()] = ($v -replace '\s+#.*$', '').Trim() }
if ($envMap.APP_ENV -eq 'production') { throw 'refusing to restore while APP_ENV=production' }
$pw = "MYSQL_PWD=$($envMap.MYSQL_ROOT_PASSWORD)"
$n = [int](docker compose exec -T -e $pw mysql mysql -uroot -N -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$Target'")
if ($n -gt 0 -and -not $Yes) { throw "database '$Target' already has $n tables; re-run with -Yes to overwrite it" }
docker compose cp $File mysql:/tmp/restore.sql.gz
docker compose exec -T -e $pw mysql mysql -uroot -e "CREATE DATABASE IF NOT EXISTS ``$Target`` CHARACTER SET utf8mb4"
docker compose exec -T -e $pw mysql sh -c "gunzip -c /tmp/restore.sql.gz | mysql -uroot $Target"
if ($LASTEXITCODE -ne 0) { throw 'restore failed' }
docker compose exec -T mysql rm -f /tmp/restore.sql.gz
"restore ok: $File -> $Target"
