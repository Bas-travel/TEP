<#
.SYNOPSIS
  Creates .env from .env.example with cryptographically random secrets (Windows PowerShell 5.1+ / PowerShell 7).
.EXAMPLE
  powershell -ExecutionPolicy Bypass -File scripts\setup-env.ps1
  powershell -ExecutionPolicy Bypass -File scripts\setup-env.ps1 -Force   # regenerate (existing DB volumes keep the OLD passwords)
#>
param([switch]$Force)
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if ((Test-Path .env) -and -not $Force) {
  Write-Error ".env already exists - nothing changed. Use -Force to regenerate."
  exit 1
}

# Alphanumeric only: safe inside the MySQL DSN and docker compose interpolation.
$alphabet = [char[]](48..57 + 65..90 + 97..122)
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
function New-Secret([int]$Length) {
  $bytes = New-Object byte[] 4
  $out = New-Object System.Text.StringBuilder
  while ($out.Length -lt $Length) {
    $rng.GetBytes($bytes)
    $n = [BitConverter]::ToUInt32($bytes, 0)
    # rejection sampling to avoid modulo bias (62 symbols)
    if ($n -lt 4294967296 - (4294967296 % $alphabet.Length)) { [void]$out.Append($alphabet[$n % $alphabet.Length]) }
  }
  $out.ToString()
}

$map = @{
  'JWT_SECRET'          = New-Secret 64
  'MYSQL_PASSWORD'      = New-Secret 32
  'MYSQL_ROOT_PASSWORD' = New-Secret 32
  'MINIO_ROOT_USER'     = 'tep' + (New-Secret 10)
  'MINIO_ROOT_PASSWORD' = New-Secret 32
}
$lines = Get-Content .env.example | ForEach-Object {
  $line = $_
  foreach ($k in $map.Keys) {
    if ($line -match "^$k=__GENERATE__") { $line = "$k=$($map[$k])" }
  }
  $line
}
# UTF-8 without BOM (docker compose rejects a BOM in .env)
[System.IO.File]::WriteAllLines((Join-Path $root '.env'), $lines, (New-Object System.Text.UTF8Encoding($false)))

Write-Host "Created .env with generated secrets (git-ignored)."
Write-Host "Next:  docker compose up --build   then   curl.exe http://localhost:8080/health/ready"
