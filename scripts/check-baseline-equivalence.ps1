param(
  [string]$ReferenceDb = "metaldocs",
  [string]$CandidateDb = "metaldocs"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$evidenceDir = "non_git/db/reference-schema"
if (-not (Test-Path $evidenceDir)) {
  New-Item -ItemType Directory -Force -Path $evidenceDir | Out-Null
}

Write-Host "[check-baseline-equivalence] Comparing runtime-used schema objects..."

docker exec metaldocs-postgres psql -U metaldocs_app -d $ReferenceDb -tAc `
  "SELECT table_schema, table_name, column_name, data_type FROM information_schema.columns ORDER BY 1,2,3;" `
  > "$evidenceDir/reference-columns.txt"

docker exec metaldocs-postgres psql -U metaldocs_app -d $CandidateDb -tAc `
  "SELECT table_schema, table_name, column_name, data_type FROM information_schema.columns ORDER BY 1,2,3;" `
  > "$evidenceDir/candidate-columns.txt"

fc.exe "$evidenceDir\reference-columns.txt" "$evidenceDir\candidate-columns.txt"
