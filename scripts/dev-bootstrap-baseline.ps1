param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env",
  [string]$BaselineFile = "migrations_baseline/0001_baseline_2026_05.sql"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $EnvFile)) {
  throw "$EnvFile not found. Copy .env.example to .env before baseline bootstrap."
}

Get-Content $EnvFile | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') {
    return
  }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name, $value, 'Process')
}

if ([string]::IsNullOrWhiteSpace($env:POSTGRES_USER) -or [string]::IsNullOrWhiteSpace($env:POSTGRES_DB)) {
  throw "POSTGRES_USER and POSTGRES_DB are required in $EnvFile for baseline bootstrap."
}

if (-not (Test-Path $BaselineFile)) {
  throw "Baseline file not found: $BaselineFile"
}

powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1 -ComposeFile $ComposeFile -EnvFile $EnvFile | Out-Host

$maxAttempts = 60
$ready = $false
for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
  try {
    $checkOutput = docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
      psql -U $env:POSTGRES_USER -d $env:POSTGRES_DB -tAc "SELECT 1" 2>&1
  } catch {
    $checkOutput = $null
    $global:LASTEXITCODE = 1
  }

  if ($LASTEXITCODE -eq 0) {
    $ready = $true
    break
  }
  Start-Sleep -Seconds 2
}

if (-not $ready) {
  throw "[dev-bootstrap-baseline] Postgres did not become ready in time."
}

$migrationLedgerQuery = @"
SELECT CASE
  WHEN to_regclass('public.schema_migrations') IS NULL THEN -1
  ELSE (SELECT COUNT(*) FROM public.schema_migrations)
END;
"@

$guardAttempts = 90
for ($attempt = 1; $attempt -le $guardAttempts; $attempt++) {
  try {
    $appliedRaw = docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
      psql -U $env:POSTGRES_USER -d $env:POSTGRES_DB -tAc $migrationLedgerQuery 2>&1
  } catch {
    $appliedRaw = $null
    $global:LASTEXITCODE = 1
  }

  if ($LASTEXITCODE -eq 0) {
    $appliedCount = -1
    [void][int]::TryParse(($appliedRaw | Out-String).Trim(), [ref]$appliedCount)
    if ($appliedCount -gt 0) {
      Write-Host "[dev-bootstrap-baseline] Detected $appliedCount row(s) in public.schema_migrations."
      Write-Host "[dev-bootstrap-baseline] Container bootstrap already initialized schema; skipping baseline apply."
      Write-Host "[dev-bootstrap-baseline] Done."
      return
    }
    if ($appliedCount -eq 0 -or $appliedCount -eq -1) {
      break
    }
  }

  Start-Sleep -Seconds 2
}

Write-Host "[dev-bootstrap-baseline] Applying baseline from $BaselineFile ..."
Get-Content -Raw $BaselineFile | docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
  psql -v ON_ERROR_STOP=1 -U $env:POSTGRES_USER -d $env:POSTGRES_DB | Out-Host

if ($LASTEXITCODE -ne 0) {
  throw "[dev-bootstrap-baseline] Baseline apply failed."
}

Write-Host "[dev-bootstrap-baseline] Baseline applied."
Write-Host "[dev-bootstrap-baseline] Applying tail migrations (if any) remains manual in this rollout."
