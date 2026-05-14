# Runbook: Baseline Local Bootstrap

Use this runbook for fresh local environments after the baseline file has been generated and validated.

## Commands

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```
