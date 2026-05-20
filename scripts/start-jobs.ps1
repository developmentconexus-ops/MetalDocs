param(
    [switch]$Build
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Get-Content ".env" | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}

$binary = Join-Path $root "metaldocs-jobs.exe"
if (-not (Test-Path $binary) -or $Build) {
    Write-Host "Building metaldocs-jobs.exe..."
    go build -o metaldocs-jobs.exe ./apps/jobs/cmd/metaldocs-jobs/...
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed"
        exit 1
    }
}

Write-Host "Starting MetalDocs Jobs (queues=temporal)"
try {
    & $binary
} catch {
    $message = $_.Exception.Message
    if ($message -notmatch 'Acesso negado' -and $message -notmatch 'Access is denied') {
        throw
    }

    Write-Host "Launching metaldocs-jobs.exe was denied by the local Windows environment; falling back to go run"
    $env:GOFLAGS = "-mod=mod"
    go run ./apps/jobs/cmd/metaldocs-jobs/...
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
