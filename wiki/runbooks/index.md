# Runbooks Index

Operator-facing, step-by-step procedures. Each runbook is written to be run
cold, top-to-bottom, with copy-paste-correct commands and expected output at
every step.

| Runbook | Purpose | Last verified |
|---|---|---|
| [`backup-restore.md`](./backup-restore.md) | PostgreSQL backup, restore, and end-to-end recovery drill (wraps `scripts/backup-postgres.ps1`, `restore-postgres.ps1`, `validate-backup.ps1`, `run-backup-restore-gate.ps1`) | 2026-07-05 |
| [`v1-release-rebaseline.md`](./v1-release-rebaseline.md) | D-4b: convert the working repo into a single-commit, history-free v1.0.0 repo (permanent closure of the F-18 git-history secret residual) | 2026-06-13 |
