param(
  [string]$BaselineFile = "db/baseline/0001_current_schema.sql",
  [string]$DictionaryDir = "wiki/database/tables"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $BaselineFile)) {
  throw "Baseline file not found: $BaselineFile"
}

if (-not (Test-Path $DictionaryDir)) {
  throw "Dictionary directory not found: $DictionaryDir"
}

$matches = Select-String -Path $BaselineFile -Pattern 'CREATE TABLE(?: IF NOT EXISTS)?\s+([a-zA-Z0-9_]+\.)?([a-zA-Z0-9_]+)' -AllMatches
$tables = @()
foreach ($match in $matches) {
  foreach ($m in $match.Matches) {
    $tables += $m.Groups[2].Value
  }
}
$tables = $tables | Sort-Object -Unique

$missing = @()
foreach ($table in $tables) {
  $page = Join-Path $DictionaryDir "$table.md"
  if (-not (Test-Path $page)) {
    $missing += $table
  }
}

if ($missing.Count -gt 0) {
  Write-Host "[check-db-dictionary-coverage] Missing dictionary pages:"
  $missing | ForEach-Object { Write-Host "  - $_" }
  exit 1
}

Write-Host "[check-db-dictionary-coverage] Dictionary coverage OK for $($tables.Count) table(s)."
