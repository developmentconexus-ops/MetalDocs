param(
  [string]$OutputFile = "migrations_baseline/0001_baseline_2026_05.sql"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$outDir = Split-Path -Parent $OutputFile
if (-not (Test-Path $outDir)) {
  New-Item -ItemType Directory -Force -Path $outDir | Out-Null
}

docker exec metaldocs-postgres pg_dump `
  -U metaldocs_app `
  -d metaldocs `
  --schema-only `
  --no-owner `
  --no-privileges `
  > $OutputFile

Write-Host "[export-schema-baseline] Wrote baseline schema to $OutputFile"
