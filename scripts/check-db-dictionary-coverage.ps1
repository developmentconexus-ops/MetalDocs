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
    $schema = $m.Groups[1].Value.TrimEnd(".")
    if ([string]::IsNullOrWhiteSpace($schema)) {
      $schema = "public"
    }
    $tables += [pscustomobject]@{
      Schema = $schema
      Table = $m.Groups[2].Value
      QualifiedName = "$schema.$($m.Groups[2].Value)"
    }
  }
}
$tables = $tables | Sort-Object Schema, Table -Unique

$missing = @()
foreach ($entry in $tables) {
  $page = Join-Path $DictionaryDir "$($entry.Table).md"
  if (-not (Test-Path $page)) {
    $missing += "$($entry.QualifiedName) -> $page"
    continue
  }

  $content = Get-Content -Raw -Path $page
  if ($content -notmatch "(?m)^#{1,2}\s+$([regex]::Escape($entry.QualifiedName))\s*$") {
    $missing += "$($entry.QualifiedName) heading in $page"
  }
}

if ($missing.Count -gt 0) {
  Write-Host "[check-db-dictionary-coverage] Missing dictionary pages:"
  $missing | ForEach-Object { Write-Host "  - $_" }
  exit 1
}

$badContent = Select-String -Path (Join-Path $DictionaryDir "*.md"), "wiki/database/dictionary-index.md" -Pattern '\$\(@\{Schema=|TBD' -List
if ($badContent) {
  Write-Host "[check-db-dictionary-coverage] Dictionary contains generated placeholder artifacts:"
  $badContent | ForEach-Object { Write-Host "  - $($_.Path):$($_.LineNumber)" }
  exit 1
}

Write-Host "[check-db-dictionary-coverage] Dictionary coverage OK for $($tables.Count) table(s)."
