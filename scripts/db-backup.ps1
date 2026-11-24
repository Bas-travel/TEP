<#
 MySQL backup for the TEP docker compose stack (Windows). Writes backups\tep-<timestamp>.sql.gz
 Usage:  powershell -ExecutionPolicy Bypass -File scripts\db-backup.ps1
 Needs .env (run scripts\setup-env.ps1 first) and `docker compose up` running.
#>
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path -Parent $PSScriptRoot)
$envMap = @{}
Get-Content .env | Where-Object { $_ -match '^\s*[A-Za-z_][A-Za-z0-9_]*=' } | ForEach-Object {
  $k, $v = $_ -split '=', 2; $envMap[$k.Trim()] = ($v -replace '\s+#.*$', '').Trim() }
$db = if ($envMap.MYSQL_DATABASE) { $envMap.MYSQL_DATABASE } else { 'tep' }
New-Item -ItemType Directory -Force backups | Out-Null
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$file = "backups\$db-$stamp.sql.gz"
# Dump + gzip inside the container so no binary data passes through the PowerShell pipeline.
docker compose exec -T -e "MYSQL_PWD=$($envMap.MYSQL_ROOT_PASSWORD)" mysql sh -c "mysqldump -uroot --single-transaction --routines --triggers --events --hex-blob --set-gtid-purged=OFF --no-tablespaces $db | gzip -9 > /tmp/backup.sql.gz"
if ($LASTEXITCODE -ne 0) { throw 'mysqldump failed' }
docker compose cp mysql:/tmp/backup.sql.gz $file
docker compose exec -T mysql rm -f /tmp/backup.sql.gz
if ((Get-Item $file).Length -lt 1024) { throw "backup suspiciously small: $file" }
"backup ok: $file ($((Get-Item $file).Length) bytes)"
