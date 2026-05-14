# Runbook: Legacy Migration Replay

Use this runbook only when a full trusted historical replay is required.

Legacy replay streams local `migrations/*.sql` files through `psql`. Docker Postgres no longer auto-runs this directory through `/docker-entrypoint-initdb.d`.

## Commands

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```
