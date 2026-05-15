param(
  [switch]$WithDevSeed
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$argsList = @()
if ($WithDevSeed) {
  $argsList += "-WithDevSeed"
}

Write-Host "[check-db-bootstrap] Bootstrapping curated database..."
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 @argsList | Out-Host

Write-Host "[check-db-bootstrap] Checking dictionary coverage..."
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1 | Out-Host

Write-Host "[check-db-bootstrap] Checking migration ledger..."
$ledger = docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT version FROM public.schema_migrations WHERE version = 'baseline-2026-05-14';"
if (($ledger | Out-String).Trim() -ne "baseline-2026-05-14") {
  throw "[check-db-bootstrap] baseline ledger marker missing"
}

Write-Host "[check-db-bootstrap] Curated DB bootstrap checks passed."
