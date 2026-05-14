param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $EnvFile)) {
  throw "$EnvFile not found. Copy .env.example to .env before running migrations."
}

Get-Content $EnvFile | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') {
    return
  }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name, $value, 'Process')
}

if ([string]::IsNullOrWhiteSpace($env:POSTGRES_USER) -or [string]::IsNullOrWhiteSpace($env:POSTGRES_DB)) {
  throw "POSTGRES_USER and POSTGRES_DB are required in $EnvFile to apply migrations."
}

if (-not (Test-Path $ComposeFile)) {
  throw "Compose file not found: $ComposeFile"
}

$migrationsPath = Join-Path $root "migrations"
if (-not (Test-Path $migrationsPath)) {
  throw "migrations folder not found: $migrationsPath"
}

$migrations = Get-ChildItem -Path $migrationsPath -Filter "*.sql" | Sort-Object -Property Name
if ($migrations.Count -eq 0) {
  throw "No migrations found in $migrationsPath"
}

Write-Host "[dev-migrate] Applying $($migrations.Count) migration(s) to Postgres container..."
Write-Host "  user=$env:POSTGRES_USER db=$env:POSTGRES_DB"
Write-Host "[dev-migrate] LEGACY REPLAY MODE."
Write-Host "[dev-migrate] This script applies the historical migrations/ chain for recovery/debugging."
Write-Host "[dev-migrate] Normal fresh local setup uses scripts/dev-bootstrap-baseline.ps1."

$maxAttempts = 60
$ready = $false
$previousNativeErrorPreference = $null
if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -ErrorAction SilentlyContinue) {
  $previousNativeErrorPreference = $PSNativeCommandUseErrorActionPreference
  $PSNativeCommandUseErrorActionPreference = $false
}
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
if ($null -ne $previousNativeErrorPreference) {
  $PSNativeCommandUseErrorActionPreference = $previousNativeErrorPreference
}

if (-not $ready) {
  throw "[dev-migrate] Postgres did not become ready in time."
}

foreach ($migration in $migrations) {
  Write-Host "[dev-migrate] -> $($migration.Name)"

  Get-Content -Raw $migration.FullName | docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
    psql -v ON_ERROR_STOP=1 -U $env:POSTGRES_USER -d $env:POSTGRES_DB | Out-Host

  if ($LASTEXITCODE -ne 0) {
    throw "[dev-migrate] Migration failed: $($migration.Name)"
  }
}

Write-Host "[dev-migrate] Done."
