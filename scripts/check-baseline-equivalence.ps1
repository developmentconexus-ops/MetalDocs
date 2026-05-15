param(
  [string]$ReferenceDb = "metaldocs_reference",
  [string]$CandidateDb = "metaldocs",
  [switch]$AllowSameDatabase
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if ($ReferenceDb -eq $CandidateDb -and -not $AllowSameDatabase) {
  throw "[check-baseline-equivalence] ReferenceDb and CandidateDb are both '$ReferenceDb'. Pass distinct DB names or -AllowSameDatabase for a smoke check."
}

$evidenceDir = "non_git/db/reference-schema"
if (-not (Test-Path $evidenceDir)) {
  New-Item -ItemType Directory -Force -Path $evidenceDir | Out-Null
}

function Export-Query {
  param(
    [string]$DbName,
    [string]$Sql,
    [string]$OutFile
  )

  docker exec metaldocs-postgres psql -U metaldocs_app -d $DbName -tAc $Sql > $OutFile
}

Write-Host "[check-baseline-equivalence] Comparing runtime-used schema objects..."

$queries = @(
  @{ Name = "columns"; Sql = "SELECT table_schema, table_name, column_name, ordinal_position, data_type, udt_name, is_nullable, column_default FROM information_schema.columns WHERE table_schema IN ('public','metaldocs') ORDER BY 1,2,4;" },
  @{ Name = "constraints"; Sql = "SELECT n.nspname AS table_schema, c.relname AS table_name, con.conname, con.contype, pg_get_constraintdef(con.oid) FROM pg_constraint con JOIN pg_class c ON c.oid = con.conrelid JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname IN ('public','metaldocs') ORDER BY 1,2,3;" },
  @{ Name = "indexes"; Sql = "SELECT schemaname, tablename, indexname, indexdef FROM pg_indexes WHERE schemaname IN ('public','metaldocs') ORDER BY 1,2,3;" },
  @{ Name = "triggers"; Sql = "SELECT trigger_schema, event_object_table, trigger_name, action_timing, event_manipulation, action_statement FROM information_schema.triggers WHERE trigger_schema IN ('public','metaldocs') ORDER BY 1,2,3,4,5;" },
  @{ Name = "functions"; Sql = "SELECT n.nspname, p.proname, pg_get_function_identity_arguments(p.oid), pg_get_functiondef(p.oid) FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname IN ('public','metaldocs') ORDER BY 1,2,3;" },
  @{ Name = "extensions"; Sql = "SELECT extname, extversion FROM pg_extension ORDER BY 1;" }
)

foreach ($query in $queries) {
  $refFile = Join-Path $evidenceDir ("reference-{0}.txt" -f $query.Name)
  $candFile = Join-Path $evidenceDir ("candidate-{0}.txt" -f $query.Name)

  Export-Query -DbName $ReferenceDb -Sql $query.Sql -OutFile $refFile
  Export-Query -DbName $CandidateDb -Sql $query.Sql -OutFile $candFile

  fc.exe $refFile $candFile | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "[check-baseline-equivalence] Mismatch in $($query.Name). Compare $refFile and $candFile."
  }
}

Write-Host "[check-baseline-equivalence] Reference and candidate schema objects match."
