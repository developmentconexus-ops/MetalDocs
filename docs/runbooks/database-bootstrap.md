# Runbook: Database Bootstrap

## Fresh local setup

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
```

## Fresh product schema without dev seed

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1
```

## Legacy replay for recovery/debugging

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
```

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v2/controlled-documents
```
