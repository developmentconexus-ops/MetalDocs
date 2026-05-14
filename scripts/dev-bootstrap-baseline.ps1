param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env",
  [switch]$WithDevSeed
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

function Invoke-DbSqlFile {
  param([string]$Path)

  if (-not (Test-Path $Path)) {
    throw "SQL file not found: $Path"
  }

  Write-Host "[dev-bootstrap-baseline] Applying $Path"
  Get-Content -Raw $Path | docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
    psql -v ON_ERROR_STOP=1 -U $env:POSTGRES_USER -d $env:POSTGRES_DB | Out-Host

  if ($LASTEXITCODE -ne 0) {
    throw "[dev-bootstrap-baseline] SQL apply failed: $Path"
  }
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

Invoke-DbSqlFile -Path "db/prerequisites/0001_extensions.sql"
Invoke-DbSqlFile -Path "db/baseline/0001_current_schema.sql"
Invoke-DbSqlFile -Path "db/reference-data/0001_product_reference_data.sql"

if ($WithDevSeed) {
  Invoke-DbSqlFile -Path "db/dev-seeds/0001_local_dev_seed.sql"
}

$tailMigrations = Get-ChildItem -Path "db/migrations" -Filter *.sql | Sort-Object Name
foreach ($migration in $tailMigrations) {
  Invoke-DbSqlFile -Path $migration.FullName
}

Write-Host "[dev-bootstrap-baseline] Curated baseline bootstrap completed."
