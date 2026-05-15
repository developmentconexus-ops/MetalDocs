# Runbook: Database Bootstrap

The supported fresh path is the curated baseline. Historical migration replay is not a normal local bootstrap path.

## Fresh local setup

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
```

`-WithDevSeed` creates the canonical local login used by smoke tests: `admin` / `AdminMetalDocs123!`.

## Fresh product schema without dev seed

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1
```

## Historical migration evidence/recovery only

Use this only to investigate old database upgrade history or recovery evidence. Do not use it to make a fresh local database runnable.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
```

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```
