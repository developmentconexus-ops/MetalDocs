$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $root ".env.docx-renderer"

if (-not (Test-Path $envFile)) {
    Write-Error "[dev-docx-renderer] Missing $envFile - copy from .env.docx-renderer.example and fill in values."
    exit 1
}

# Load env vars from .env.docx-renderer
Get-Content $envFile | ForEach-Object {
    if ($_ -match '^\s*([^#=]+?)\s*=\s*(.*)\s*$') {
        [System.Environment]::SetEnvironmentVariable($Matches[1], $Matches[2], 'Process')
    }
}

Write-Host "[dev-docx-renderer] Starting docx-renderer on port $env:DOCX_RENDERER_PORT ..."
Set-Location (Join-Path $root "apps/docx-renderer")
npm.cmd run dev
