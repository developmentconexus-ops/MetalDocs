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

  # Fail loud: capture stderr separately so a non-zero psql exit (bad SQL,
  # connection refused, missing DB, ...) aborts the gate instead of writing
  # an empty/partial file that then compares as "equal" to another empty
  # file (empty==empty false-green is the exact hazard being closed here).
  $stderrFile = "$OutFile.stderr"
  docker exec metaldocs-postgres psql -v ON_ERROR_STOP=1 -U metaldocs_app -d $DbName -tAc $Sql 1> $OutFile 2> $stderrFile
  $exitCode = $LASTEXITCODE
  $stderrText = ""
  if (Test-Path $stderrFile) {
    $stderrText = (Get-Content -Raw $stderrFile -ErrorAction SilentlyContinue)
    Remove-Item $stderrFile -Force -ErrorAction SilentlyContinue
  }

  if ($exitCode -ne 0) {
    throw "[check-baseline-equivalence] psql failed (exit $exitCode) against '$DbName' for query -> $OutFile. stderr: $stderrText"
  }
  if (-not [string]::IsNullOrWhiteSpace($stderrText)) {
    throw "[check-baseline-equivalence] psql wrote to stderr against '$DbName' for query -> $OutFile (treated as failure). stderr: $stderrText"
  }
}

Write-Host "[check-baseline-equivalence] Comparing runtime-used schema objects..."

$queries = @(
  # ordinal_position is the ORDER BY key (so a real column *reordering* is still
  # caught) but is deliberately NOT in the compared projection: a freshly-created
  # baseline table has contiguous attnums while an incrementally-migrated table
  # carries attnum gaps from past ALTER ... DROP COLUMN (e.g. folded 0236 dropping
  # templates_template.visibility). That gap is invisible at runtime — every client
  # binds columns by name via generated/contract-first SQL — so comparing the
  # absolute number would reject a correct squash on a non-runtime difference. The
  # column set, types, nullability, defaults, and relative order are what equivalence
  # means, and those are all still compared.
  @{ Name = "columns"; Sql = "SELECT table_schema, table_name, column_name, data_type, udt_name, is_nullable, column_default FROM information_schema.columns WHERE table_schema IN ('public','metaldocs') ORDER BY table_schema, table_name, ordinal_position;" },
  @{ Name = "constraints"; Sql = "SELECT n.nspname AS table_schema, c.relname AS table_name, con.conname, con.contype, pg_get_constraintdef(con.oid) FROM pg_constraint con JOIN pg_class c ON c.oid = con.conrelid JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname IN ('public','metaldocs') ORDER BY 1,2,3;" },
  @{ Name = "indexes"; Sql = "SELECT schemaname, tablename, indexname, indexdef FROM pg_indexes WHERE schemaname IN ('public','metaldocs') ORDER BY 1,2,3;" },
  @{ Name = "triggers"; Sql = "SELECT trigger_schema, event_object_table, trigger_name, action_timing, event_manipulation, action_statement FROM information_schema.triggers WHERE trigger_schema IN ('public','metaldocs') ORDER BY 1,2,3,4,5;" },
  @{ Name = "functions"; Sql = "SELECT n.nspname, p.proname, pg_get_function_identity_arguments(p.oid), pg_get_functiondef(p.oid) FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname IN ('public','metaldocs') ORDER BY 1,2,3;" },
  @{ Name = "extensions"; Sql = "SELECT extname, extversion FROM pg_extension ORDER BY 1;" },
  @{ Name = "tables"; Sql = "SELECT n.nspname, c.relname, c.relkind FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname IN ('public','metaldocs') AND c.relkind IN ('r','p') ORDER BY 1,2;" },
  @{ Name = "rls-flags"; Sql = "SELECT n.nspname, c.relname, c.relrowsecurity, c.relforcerowsecurity FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname IN ('public','metaldocs') AND c.relkind IN ('r','p') ORDER BY 1,2;" },
  @{ Name = "policies"; Sql = "SELECT schemaname, tablename, policyname, cmd, roles, qual, with_check FROM pg_policies WHERE schemaname IN ('public','metaldocs') ORDER BY 1,2,3;" },
  @{ Name = "views"; Sql = "SELECT schemaname, viewname, definition FROM pg_views WHERE schemaname IN ('public','metaldocs') ORDER BY 1,2;" },
  @{ Name = "sequences"; Sql = "SELECT sequence_schema, sequence_name, data_type, start_value, increment FROM information_schema.sequences WHERE sequence_schema IN ('public','metaldocs') ORDER BY 1,2;" }
  # Ownership/ACLs deliberately NOT compared: both scratch DBs in this gate are
  # created and populated by the same connecting role (metaldocs_app), so
  # relowner/relacl are constant across reference and candidate by construction
  # -- comparing them would add noise, not signal.
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
