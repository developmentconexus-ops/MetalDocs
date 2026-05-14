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
Write-Host "[dev-migrate] Legacy full replay mode: applying the complete migrations/ chain."

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

$initAttempts = 120
for ($attempt = 1; $attempt -le $initAttempts; $attempt++) {
  $topOutput = docker top metaldocs-postgres 2>$null
  if ($LASTEXITCODE -ne 0) {
    Start-Sleep -Seconds 1
    continue
  }

  $initInProgress = $false
  foreach ($line in $topOutput) {
    if ($line -match '/docker-entrypoint-initdb\.d/') {
      $initInProgress = $true
      break
    }
  }

  if (-not $initInProgress) {
    break
  }

  if ($attempt -eq $initAttempts) {
    throw "[dev-migrate] Postgres init scripts are still running after timeout."
  }

  Start-Sleep -Seconds 2
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
      Write-Host "[dev-migrate] Detected $appliedCount row(s) in public.schema_migrations."
      Write-Host "[dev-migrate] Legacy chain is already initialized by container bootstrap; skipping replay."
      Write-Host "[dev-migrate] Done."
      return
    }
    if ($appliedCount -eq 0 -or $appliedCount -eq -1) {
      break
    }
  }

  Start-Sleep -Seconds 2
}

foreach ($migration in $migrations) {
  $containerPath = "/docker-entrypoint-initdb.d/$($migration.Name)"
  Write-Host "[dev-migrate] -> $($migration.Name)"

  docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
    psql -v ON_ERROR_STOP=1 -U $env:POSTGRES_USER -d $env:POSTGRES_DB -f $containerPath

  if ($LASTEXITCODE -ne 0) {
    throw "[dev-migrate] Migration failed: $($migration.Name)"
  }
}

Write-Host "[dev-migrate] Done."
