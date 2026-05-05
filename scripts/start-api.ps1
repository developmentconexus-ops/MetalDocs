$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

# Load .env — split on first '=' only so PGPASSWORD=***REDACTED*** is preserved intact
Get-Content ".env" | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}

# APP_PORT must be 8081 — override in case .env is missing it
[System.Environment]::SetEnvironmentVariable('APP_PORT', '8081', 'Process')

# Kill any process already holding :8081
$held = netstat -ano 2>$null | Select-String ":8081 " | ForEach-Object { ($_ -split '\s+')[5] } | Select-Object -First 1
if ($held) {
    Write-Host "Killing PID $held (was holding :8081)"
    Stop-Process -Id $held -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
}

# Always rebuild binaries to avoid stale binary issues
$binary = Join-Path $root "metaldocs-api.exe"
$workerBinary = Join-Path $root "metaldocs-worker.exe"
Write-Host "Building metaldocs-api.exe..."
go build -o metaldocs-api.exe ./apps/api/cmd/metaldocs-api/...
if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }
Write-Host "Building metaldocs-worker.exe..."
go build -o metaldocs-worker.exe ./apps/worker/cmd/metaldocs-worker/...
if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }

# Start worker in background so PDF jobs are consumed alongside the API
if ($args -notcontains "-NoWorker") {
    Write-Host "Starting MetalDocs Worker in background..."
    $workerProc = Start-Process -FilePath $workerBinary -PassThru -WindowStyle Hidden
    Write-Host "Worker PID: $($workerProc.Id)"
}

Write-Host "Starting MetalDocs API on :8081  (admin / AdminMetalDocs123!)"
& $binary
