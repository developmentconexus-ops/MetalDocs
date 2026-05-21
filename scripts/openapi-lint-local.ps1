param(
    [string]$SpecPath = "api/openapi/v1/openapi.yaml"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $SpecPath)) {
    throw "Spec file not found: $SpecPath"
}

# Local Windows guard: avoids a post-lint Node/libuv assertion observed with npm exec.
$env:CI = "1"

npx.cmd @redocly/cli lint $SpecPath
