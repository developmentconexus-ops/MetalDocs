param(
  [switch]$WithDevSeed,
  [switch]$VerifyForwardMigration
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$argsList = @()
if ($WithDevSeed) {
  $argsList += "-WithDevSeed"
}

Write-Host "[check-db-bootstrap] Bootstrapping curated database..."
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 @argsList | Out-Host

Write-Host "[check-db-bootstrap] Checking dictionary coverage..."
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1 | Out-Host

Write-Host "[check-db-bootstrap] Checking migration ledger..."
$ledger = docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT version FROM public.schema_migrations WHERE version = 'baseline-2026-05-14';"
if (($ledger | Out-String).Trim() -ne "baseline-2026-05-14") {
  throw "[check-db-bootstrap] baseline ledger marker missing"
}

if ($VerifyForwardMigration) {
  Write-Host "[check-db-bootstrap] Checking post-baseline forward migration execution..."
  $fixtureVersion = "9999"
  $fixtureFile = Join-Path $root "db/migrations/9999_check_forward_migration_fixture.sql"
  $fixtureSql = @"
BEGIN;
CREATE TABLE IF NOT EXISTS metaldocs.forward_migration_fixture (
  id text PRIMARY KEY
);
INSERT INTO public.schema_migrations(version, description)
VALUES ('$fixtureVersion', 'check forward migration fixture')
ON CONFLICT (version) DO NOTHING;
COMMIT;
"@
  $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($fixtureFile, $fixtureSql, $utf8NoBom)

  try {
    Get-Process -Name metaldocs-api -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    $fixtureApiOutLog = Join-Path $root "non_git/check-db-bootstrap-fixture-api.out.log"
    $fixtureApiErrLog = Join-Path $root "non_git/check-db-bootstrap-fixture-api.err.log"
    Remove-Item -Force $fixtureApiOutLog,$fixtureApiErrLog -ErrorAction SilentlyContinue

    $apiBinary = Join-Path $root "metaldocs-api.exe"
    go build -o $apiBinary ./apps/api/cmd/metaldocs-api/...
    if ($LASTEXITCODE -ne 0) {
      throw "[check-db-bootstrap] failed to build metaldocs-api.exe for forward migration verification"
    }

    Get-Content ".env" | ForEach-Object {
      if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
      $name, $value = $_ -split '=', 2
      [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
    }
    [System.Environment]::SetEnvironmentVariable('APP_PORT', '18081', 'Process')

    $apiProcess = Start-Process `
      -FilePath $apiBinary `
      -WorkingDirectory $root `
      -PassThru `
      -RedirectStandardOutput $fixtureApiOutLog `
      -RedirectStandardError $fixtureApiErrLog `
      -WindowStyle Hidden

    $deadline = (Get-Date).AddSeconds(420)
    $ready = $false
    while ((Get-Date) -lt $deadline) {
      if ($apiProcess.HasExited) {
        $errTail = if (Test-Path $fixtureApiErrLog) { (Get-Content $fixtureApiErrLog -Tail 80 | Out-String) } else { "<no stderr log>" }
        $outTail = if (Test-Path $fixtureApiOutLog) { (Get-Content $fixtureApiOutLog -Tail 80 | Out-String) } else { "<no stdout log>" }
        throw "[check-db-bootstrap] metaldocs-api exited before readiness on APP_PORT=18081 (exit=$($apiProcess.ExitCode)). stdout tail:`n$outTail`nstderr tail:`n$errTail"
      }
      try {
        $response = Invoke-WebRequest -Uri "http://127.0.0.1:18081/api/v1/health/ready" -Method Get -TimeoutSec 10 -UseBasicParsing
        if ($response.StatusCode -eq 200) {
          $ready = $true
          break
        }
      } catch {
      }
      Start-Sleep -Seconds 1
    }

    if (-not $ready) {
      $errTail = if (Test-Path $fixtureApiErrLog) { (Get-Content $fixtureApiErrLog -Tail 80 | Out-String) } else { "<no stderr log>" }
      $outTail = if (Test-Path $fixtureApiOutLog) { (Get-Content $fixtureApiOutLog -Tail 80 | Out-String) } else { "<no stdout log>" }
      throw "[check-db-bootstrap] timed out waiting for API readiness on APP_PORT=18081 during forward migration verification. stdout tail:`n$outTail`nstderr tail:`n$errTail"
    }
  } finally {
    Remove-Item -Force $fixtureFile -ErrorAction SilentlyContinue
    Get-Process -Name metaldocs-api -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
  }

  $fixtureLedger = docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT version FROM public.schema_migrations WHERE version = '$fixtureVersion';"
  if (($fixtureLedger | Out-String).Trim() -ne $fixtureVersion) {
    throw "[check-db-bootstrap] forward migration fixture ledger marker missing"
  }
  docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "DROP TABLE IF EXISTS metaldocs.forward_migration_fixture; DELETE FROM public.schema_migrations WHERE version = '$fixtureVersion';" | Out-Null
}

Write-Host "[check-db-bootstrap] Checking critical tables..."
$criticalTables = @(
  "public.schema_migrations",
  "public.documents",
  "public.controlled_documents",
  "public.templates_template",
  "public.templates_template_version",
  "public.approval_routes"
)
$criticalSqlList = ($criticalTables | ForEach-Object { "'$_'" }) -join ", "
$missingTables = docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc @"
WITH expected(table_name) AS (
  SELECT unnest(ARRAY[$criticalSqlList])
)
SELECT expected.table_name
FROM expected
LEFT JOIN (
  SELECT quote_ident(table_schema) || '.' || quote_ident(table_name) AS table_name
  FROM information_schema.tables
  WHERE table_type = 'BASE TABLE'
) actual ON actual.table_name = expected.table_name
WHERE actual.table_name IS NULL
ORDER BY expected.table_name;
"@
$missingTableList = (($missingTables | Out-String).Trim() -split "`r?`n" | Where-Object { $_ -ne "" })
if ($missingTableList.Count -gt 0) {
  throw ("[check-db-bootstrap] critical tables missing: {0}" -f ($missingTableList -join ", "))
}

Write-Host "[check-db-bootstrap] Curated DB bootstrap checks passed."
