# Regenerates the curated schema baseline from a live database.
#
# The default target IS the canonical baseline (db/baseline/0001_current_schema.sql) --
# the previous default, migrations_baseline/0001_baseline_2026_05.sql, wrote to an
# orphaned path with zero consumers and was deleted on 2026-07-29.
#
# MANDATORY GATE: the raw dump this writes is NOT a drop-in baseline. After running
# it you must (a) curate it back to the file's conventions -- BEGIN/COMMIT wrap,
# `SET check_function_bodies = false;`, session SETs and \restrict/\unrestrict lines
# stripped, extensions left to db/prerequisites/0001_extensions.sql, CREATE SCHEMA
# IF NOT EXISTS, updated fold header -- and (b) prove equivalence with
# scripts/check-baseline-equivalence.ps1 against a reference database built from the
# full migration chain. A baseline that has not passed that gate must not be committed.
#
# Note --no-privileges below: ACLs and roles are NOT part of the baseline. They live
# in db/grants/0001_role_grants.sql; keep that file in sync by hand when a migration
# changes GRANT/REVOKE/role posture.
param(
  [string]$OutputFile = "db/baseline/0001_current_schema.sql"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$outDir = Split-Path -Parent $OutputFile
if (-not (Test-Path $outDir)) {
  New-Item -ItemType Directory -Force -Path $outDir | Out-Null
}

docker exec metaldocs-postgres pg_dump `
  -U metaldocs_app `
  -d metaldocs `
  --schema-only `
  --no-owner `
  --no-privileges `
  > $OutputFile

Write-Host "[export-schema-baseline] Wrote baseline schema to $OutputFile"
