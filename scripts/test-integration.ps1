param(
  [string]$EnvFile = ".env",
  [string]$Package = "./...",
  [string]$Run
)

# Canonical entrypoint for running integration tests locally.
# testdb.Open (tests/integration/testdb/db.go) SKIPS every integration test when
# DATABASE_URL / METALDOCS_DATABASE_URL is unset or the DB is unreachable — a
# silent skip reads as green. This script is the single chokepoint that makes
# that impossible: it derives DATABASE_URL from the POSTGRES_* keys already in
# .env (never printing the value), verifies the compose postgres container is
# actually accepting connections, and only then runs the suite.

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if ([string]::IsNullOrWhiteSpace($env:DATABASE_URL) -and [string]::IsNullOrWhiteSpace($env:METALDOCS_DATABASE_URL)) {
  if (-not (Test-Path $EnvFile)) {
    throw "[test-integration] $EnvFile not found and DATABASE_URL not set. Copy .env.example to .env first."
  }

  $vars = @{}
  Get-Content $EnvFile | ForEach-Object {
    if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
    $name, $value = $_ -split '=', 2
    if ($name -in @('POSTGRES_USER', 'POSTGRES_PASSWORD', 'POSTGRES_DB', 'POSTGRES_HOST_PORT')) {
      $vars[$name] = $value
    }
  }

  foreach ($required in 'POSTGRES_USER', 'POSTGRES_PASSWORD', 'POSTGRES_DB') {
    if ([string]::IsNullOrWhiteSpace($vars[$required])) {
      throw "[test-integration] $required missing in $EnvFile — cannot derive DATABASE_URL."
    }
  }

  $port = if ([string]::IsNullOrWhiteSpace($vars['POSTGRES_HOST_PORT'])) { '5433' } else { $vars['POSTGRES_HOST_PORT'] }
  $user = [uri]::EscapeDataString($vars['POSTGRES_USER'])
  $pass = [uri]::EscapeDataString($vars['POSTGRES_PASSWORD'])
  $env:DATABASE_URL = "postgres://${user}:${pass}@127.0.0.1:${port}/$($vars['POSTGRES_DB'])?sslmode=disable"
  Write-Host "[test-integration] DATABASE_URL derived from $EnvFile POSTGRES_* keys (host port $port; value not printed)."
}
else {
  Write-Host "[test-integration] Using DATABASE_URL/METALDOCS_DATABASE_URL from environment."
}

# Fail loud if postgres is not accepting connections — otherwise testdb.Open
# skips every test and the run reads as a false green.
$pgReady = docker exec metaldocs-postgres pg_isready 2>&1
if ($LASTEXITCODE -ne 0) {
  throw "[test-integration] postgres container not accepting connections ($pgReady). Bring the stack up first: docker compose --env-file .env -f deploy/compose/docker-compose.yml up -d postgres"
}
Write-Host "[test-integration] postgres reachable."

$goArgs = @('test', '-tags=integration')
if (-not [string]::IsNullOrWhiteSpace($Run)) {
  $goArgs += @('-run', $Run)
}
$goArgs += $Package

Write-Host "[test-integration] go $($goArgs -join ' ')"
& go @goArgs
if ($LASTEXITCODE -ne 0) {
  throw "[test-integration] integration tests FAILED (exit $LASTEXITCODE)."
}
Write-Host "[test-integration] PASS."
