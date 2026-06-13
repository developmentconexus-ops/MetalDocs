# Post-v1 Backlog (Wave Z anti-circle parking lot)

> **Last verified:** 2026-06-12
> **Scope:** Pre-existing findings discovered during Wave Z that are NOT Wave-Z-caused regressions. Parked here by the anti-circle rule (spec §1). None block backend DONE. Triage post-v1.

| # | Found during | One-line description | Pointer |
|---|---|---|---|
| 1 | Z-6 | `UserAreaRepository` legacy non-Tx methods (`Insert`, `CloseActive`, `GrantAtomic`) only reachable from old test fakes after in-tx migration — removable surface | `internal/modules/iam/infrastructure/postgres/user_area_repository.go` |
| 2 | Z-11 | Worker bootstrap dials its own MinIO client; consolidation candidate if a shared `WorkerDependencies` refactor ever happens | `internal/platform/bootstrap/worker.go` |
| 3 | Z-13 | Approval result types carry status as plain `string` (`MarkObsoleteResult.PriorStatus`, `SupersedeResult.*`, `PublishResult.NewStatus`) — could be typed `DocumentStatus` | `internal/modules/documents/approval/application/{obsolete,supersede,publish}_service.go` |
| 4 | Z-13 | `TemplateVersionCandidate.Status` is `*string` with hardcoded `'published'`/`'obsolete'` comparisons — typed `TemplateVersionStatus` enum candidate | `internal/modules/controlleddocuments/domain/resolution.go:14` |
| 5 | Z-25 | Curated baseline still references dropped `subject_code` lines — remove at next baseline refresh | `db/baseline/0001_current_schema.sql:1354,3237,3240,4057` |
