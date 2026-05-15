# Runbook: Baseline Local Bootstrap

Use this runbook for fresh local environments after the baseline file has been generated and validated.

This path uses the curated baseline plus optional local dev seed. It does not replay historical migrations.

## Commands

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -StartApi -TargetRoute /api/v1/controlled-documents
```
