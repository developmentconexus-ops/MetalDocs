param(
  [switch]$Build,
  [switch]$ForceRestart,
  [int]$ReadyTimeoutSeconds = 180,
  [string]$EnvFile = ".env"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $EnvFile)) {
  throw "$EnvFile not found. Copy .env.example to .env before running local dev."
}

Get-Content $EnvFile | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') {
    return
  }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}

$required = @(
  "METALDOCS_AUTH_SESSION_SECRET",
  "PGHOST",
  "PGPORT",
  "PGDATABASE",
  "PGUSER",
  "PGPASSWORD"
)

foreach ($name in $required) {
  $value = [System.Environment]::GetEnvironmentVariable($name, 'Process')
  if ([string]::IsNullOrWhiteSpace($value)) {
    throw "$name is required in $EnvFile for local API runtime."
  }
}

[System.Environment]::SetEnvironmentVariable('APP_PORT', '8081', 'Process')

$appPort = 8081
$binary = Join-Path $root "metaldocs-api.exe"
$healthURL = "http://127.0.0.1:$appPort/api/v1/health/ready"
$logDir = Join-Path $root "non_git"
$stdoutLog = Join-Path $logDir "dev-api.out.log"
$stderrLog = Join-Path $logDir "dev-api.err.log"

function Test-Ready {
  try {
    $response = Invoke-WebRequest -Uri $healthURL -UseBasicParsing -TimeoutSec 5
    return $response.StatusCode -eq 200
  } catch {
    return $false
  }
}

function Get-ListeningProcessId {
  $listener = Get-NetTCPConnection -LocalPort $appPort -State Listen -ErrorAction SilentlyContinue |
    Select-Object -First 1
  if (-not $listener) {
    return $null
  }
  return [int]$listener.OwningProcess
}

function Get-ProcessPath {
  param([int]$ProcessId)

  $process = Get-Process -Id $ProcessId -ErrorAction SilentlyContinue
  if (-not $process) {
    return $null
  }

  try {
    return $process.Path
  } catch {
    return $null
  }
}

function Get-LogTail {
  param([string]$Path)

  if (-not (Test-Path $Path)) {
    return "<missing log: $Path>"
  }

  return (Get-Content $Path -Tail 40 | Out-String).Trim()
}

Write-Host "[dev-api] Using Docker infra as source of truth:"
Write-Host "  PGHOST=$env:PGHOST"
Write-Host "  PGPORT=$env:PGPORT"
Write-Host "  METALDOCS_STORAGE_PROVIDER=$env:METALDOCS_STORAGE_PROVIDER"

$existingPid = Get-ListeningProcessId
if ($existingPid) {
  $existingPath = Get-ProcessPath -ProcessId $existingPid
  $resolvedBinary = if (Test-Path $binary) { (Resolve-Path $binary).Path } else { $null }
  $isOwnedAPI = $existingPath -and $resolvedBinary -and ((Resolve-Path $existingPath).Path -ieq $resolvedBinary)

  if ($isOwnedAPI -and -not $ForceRestart -and (Test-Ready)) {
    Write-Host "[dev-api] Reusing healthy MetalDocs API on :$appPort (PID $existingPid)"
    exit 0
  }

  if ($isOwnedAPI) {
    Write-Host "[dev-api] Stopping existing MetalDocs API on :$appPort (PID $existingPid)"
    Stop-Process -Id $existingPid -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
  } else {
    throw "APP_PORT $appPort is already in use by another process (PID $existingPid, path=$existingPath)."
  }
}

if ($Build -or -not (Test-Path $binary)) {
  Write-Host "[dev-api] Building metaldocs-api.exe..."
  go build -o $binary ./apps/api/cmd/metaldocs-api/...
  if ($LASTEXITCODE -ne 0) {
    throw "[dev-api] go build failed for metaldocs-api.exe"
  }
}

New-Item -ItemType Directory -Force $logDir | Out-Null
Remove-Item -Force $stdoutLog, $stderrLog -ErrorAction SilentlyContinue

Write-Host "[dev-api] Starting local API on :$appPort ..."
$process = Start-Process `
  -FilePath $binary `
  -WorkingDirectory $root `
  -RedirectStandardOutput $stdoutLog `
  -RedirectStandardError $stderrLog `
  -PassThru `
  -WindowStyle Hidden

$deadline = (Get-Date).AddSeconds($ReadyTimeoutSeconds)
while ((Get-Date) -lt $deadline) {
  if ($process.HasExited) {
    $stdoutTail = Get-LogTail -Path $stdoutLog
    $stderrTail = Get-LogTail -Path $stderrLog
    throw "[dev-api] metaldocs-api exited before readiness (exit=$($process.ExitCode)). stdout tail:`n$stdoutTail`nstderr tail:`n$stderrTail"
  }

  if (Test-Ready) {
    Write-Host "[dev-api] API ready on :$appPort (PID $($process.Id))"
    Write-Host "[dev-api] Logs: $stdoutLog | $stderrLog"
    exit 0
  }

  Start-Sleep -Seconds 1
}

$stdoutTail = Get-LogTail -Path $stdoutLog
$stderrTail = Get-LogTail -Path $stderrLog
throw "[dev-api] timed out waiting for readiness at $healthURL. stdout tail:`n$stdoutTail`nstderr tail:`n$stderrTail"
