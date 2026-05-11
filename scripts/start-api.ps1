param(
    [switch]$Build,
    [switch]$NoWorker
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Get-Content ".env" | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}

[System.Environment]::SetEnvironmentVariable('APP_PORT', '8081', 'Process')

$held = netstat -ano 2>$null | Select-String ":8081 " | ForEach-Object { ($_ -split '\s+')[5] } | Select-Object -First 1
if ($held) {
    Write-Host "Killing PID $held (was holding :8081)"
    Stop-Process -Id $held -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
}

$binary = Join-Path $root "metaldocs-api.exe"
$workerBinary = Join-Path $root "metaldocs-worker.exe"

if ($Build) {
    Write-Host "Building metaldocs-api.exe..."
    go build -o metaldocs-api.exe ./apps/api/cmd/metaldocs-api/...
    if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }
    Write-Host "Building metaldocs-worker.exe..."
    go build -o metaldocs-worker.exe ./apps/worker/cmd/metaldocs-worker/...
    if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }
} elseif (-not (Test-Path $binary)) {
    Write-Host "Binary not found, building..."
    go build -o metaldocs-api.exe ./apps/api/cmd/metaldocs-api/...
    if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }
    go build -o metaldocs-worker.exe ./apps/worker/cmd/metaldocs-worker/...
    if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }
}

if (-not $NoWorker) {
    Write-Host "Starting MetalDocs Worker in background..."
    $workerProc = Start-Process -FilePath $workerBinary -PassThru -WindowStyle Hidden
    Write-Host ("Worker PID: " + $workerProc.Id)
}

Write-Host "Starting MetalDocs API on :8081"
& $binary