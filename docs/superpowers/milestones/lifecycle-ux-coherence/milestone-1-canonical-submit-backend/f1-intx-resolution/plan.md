# Feature F1 — Plan

> **Milestone:** 1 — canonical-submit-backend · **Folder:** `f1-intx-resolution`
> **Status:** Done

## Source
- Milestone row: *In-tx server resolution of route_id/profile/content_hash inside the submit tx;
  REV0 zero-governance submit; REV≥1 title/reason gates from derived governed rev number.*
- Governing-spec reference: §1.3 (submit owns prereq resolution); ADR 0073 §2.

## Plan

Executed as an approved inline spike (operator recovery choice 2026-07-06), then validated. Task order:

1. **Narrow port** — add `SubmitDefaultsResolver` interface (`submit_defaults.go`):
   `LoadControlledDocumentID(tx, tenant, docID)`, `LoadActiveRouteIDByProfile(tx, tenant, profile)`,
   `LoadHeadContentHash(tx, tenant, docID)`. ISP: service depends on the port, not the fat repo.
2. **Postgres impl** on the existing approval repo (`postgres_approval_repository.go`) — non-recording
   SELECTs; route query mirrors the wrapper's `GetFinalizePrereqs` (`repository.go:1801-1817`),
   head-hash query mirrors `repository.go:1819-1828` (no-rows tolerant → `""`).
3. **Wire** in `services.go` — `resolver, _ := repo.(SubmitDefaultsResolver)`; pass into `SubmitService`
   alongside the already-wired `cdRead` (`CDFieldReader`).
4. **Service logic** (`submit_service.go`) — `resolveActiveRoute` (empty route_id → controlled-doc-id →
   `cdRead.ProfileCode` → `LoadActiveRouteIDByProfile`; nil resolver → `ErrApprovalRouteMissing`
   fail-closed) and `resolveContentFormData` (present-but-empty `_content_hash` → `LoadHeadContentHash`;
   caller map never mutated). Both run **inside** the submit tx before the existing CAS.
5. **Governed rev gates** — `normalizeGovernedRevisionTitle` (REV0 empty → default "Criacao do
   documento"; REV≥1 empty → `ErrRevisionTitleRequired`), `normalizeReasonForChange` (REV≥1 gates).

### Files touched
- `internal/modules/documents/approval/application/submit_defaults.go` (new port)
- `internal/modules/documents/approval/application/submit_service.go` (resolution + gates)
- `internal/modules/documents/approval/application/services.go` (wiring)
- `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go` (impl)

### Test strategy
- **Integration (real testdb, pg16)** — `submit_service_defaults_integration_test.go`, 6 cases
  (resolve-route, bind-head-hash, REV0 zero-gov, REV1 reason gate, explicit route+hash, replay).
  This is the primary gate — resolution is DB-shaped, unit fakes would not prove it.
- **Live QA (docker curl)** — decisive REV0 zero-governance path.
- Existing approval unit tests remain green (resolver-nil + explicit path unaffected).

### Ordering
Port → impl → wire → service logic → gates → integration test → live QA.

## Execution notes
Built inline (spike), then retro-formalized under the operator's "Formalize + validate current code"
choice. Integration test authored by a fresh subagent on postgres:16 (pg17 wedged during initdb on
this Windows/Docker host; testdb bootstrap needs a connected superuser named `metaldocs_app`). Live
QA required an api rebuild (stale 4h binary predated the parseSubmitIfMatch v0 relaxation).
