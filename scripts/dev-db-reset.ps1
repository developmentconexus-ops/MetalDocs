param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $EnvFile)) {
  throw "$EnvFile not found. Copy .env.example to .env before resetting the local DB."
}

Write-Host "[dev-db-reset] Stopping app containers..."
docker compose -f $ComposeFile --env-file $EnvFile stop api web gateway worker | Out-Host

Write-Host "[dev-db-reset] Removing Postgres container and volume..."
docker compose -f $ComposeFile --env-file $EnvFile down -v postgres | Out-Host

Write-Host "[dev-db-reset] Starting infra containers again..."
docker compose -f $ComposeFile --env-file $EnvFile up -d postgres redis minio | Out-Host

Write-Host "[dev-db-reset] Local database reset complete."
