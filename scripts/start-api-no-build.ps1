$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

# Load .env — split on first '=' only so PGPASSWORD=Lepa12<>! is preserved intact
Get-Content ".env" | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}

# APP_PORT must be 8081
[System.Environment]::SetEnvironmentVariable('APP_PORT', '8081', 'Process')

$binary = Join-Path $root "metaldocs-api.exe"
if (-not (Test-Path $binary)) {
    Write-Error "metaldocs-api.exe not found. Run .\scripts\start-api.ps1 -Build first."
    exit 1
}

Write-Host "Starting MetalDocs API on :8081 (no rebuild)"
& $binary
