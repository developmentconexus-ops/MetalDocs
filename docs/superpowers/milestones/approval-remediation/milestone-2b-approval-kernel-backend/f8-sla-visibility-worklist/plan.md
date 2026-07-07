# Feature F8 — `sla-visibility-worklist` — plan.md

Ref: `spec.md` (this folder), plan.md `docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md` F8 section, design spec.md §4/§6.3 (W4/P2/P3/P8).

## Task 1 — Migration: SLA snapshot column

- [ ] `db/migrations/0292_approval_stage_sla.sql`: `ALTER TABLE approval_stage_instances ADD COLUMN due_in_days_snapshot integer NULL CHECK (due_in_days_snapshot IS NULL OR due_in_days_snapshot > 0);` + ledger insert (`ON CONFLICT DO NOTHING`).
- [ ] Grep-confirm no existing column of this name (`grep -n "due_in_days_snapshot" db/`).

## Task 2 — Domain: `DueInDaysSnapshot` field

- [ ] Add `DueInDaysSnapshot *int` to `domain.StageInstance` (`internal/modules/documents/approval/domain/instance.go`).
- [ ] Failing test: extend `domain/instance_test.go` — construct a `StageInstance{DueInDaysSnapshot: ptr(3)}`, assert the field round-trips (pure struct test, TDD formality since there's no behavior yet).

## Task 3 — Repository: persist + compute `due_at` on activation

- [ ] Failing integration test first (`tests/integration/approval/sla_due_at_integration_test.go`, testdb factory): seed a route with stage 1 `due_in_days=3`, stage 2 `due_in_days=NULL`; submit → assert stage 1's `due_at` ≈ `now()+3d`; record enough signoffs to advance to stage 2 → assert stage 2's `due_at` is NULL.
- [ ] `InsertStageInstances`: add `due_in_days_snapshot` to the fixed column list + args (mirrors the existing `stage_kind` NOT-NULL-default handling pattern at line ~92-97, but this column is nullable so no default substitution needed).
- [ ] `UpdateStageStatus`: extend the SET clause so activation (`newStatus='active'`) also computes `due_at = CASE WHEN due_in_days_snapshot IS NOT NULL THEN now() + (due_in_days_snapshot || ' days')::interval ELSE NULL END` — single UPDATE, DB-computed, no app clock read.
- [ ] `LoadInstance`/`LoadInstancesByIDs` SELECT lists: add `due_in_days_snapshot` alongside the existing `stage_kind, due_at` (lines ~643/704, ~884/941) into `StageInstance.DueInDaysSnapshot`.
- [ ] `submit_service.go`: stage-instance construction loop sets `DueInDaysSnapshot: stage.DueInDays` for every stage (not just stage 0).
- [ ] Run the Task 3 test — PASS. `go build ./...` clean.

## Task 4 — SLA surfacer: new sibling job

- [ ] New published ports on `internal/modules/documents/approval/domain` (or a small new file, e.g. `sla_ports.go`): `SLAOverdueReader` (`ListTenantsWithOverdueApprovalStages(ctx, tx, now) ([]string, error)`, `ListOverdueApprovalStages(ctx, tx, tenantID, now, limit) ([]StageInstance, error)`), `SLAOverdueWriter` (`MarkSLASurfaced(ctx, tx, tenantID, now) ([]string, error)` — idempotent, mirrors `ReviewSurfaceWriter.MarkSurfaced`'s shape; requires a `sla_surfaced_at` column on `approval_stage_instances`, add it in migration 0292 alongside `due_in_days_snapshot`).
- [ ] Failing unit test for the job's `run()` orchestration (mirror `document_review_surfacer/job_integration_test.go`'s structure) using fake reader/writer — asserts phase-1 tenant enumeration, phase-2 per-tenant seeded tx, idempotent surfacing.
- [ ] Implement `internal/modules/jobs/approval_sla_surfacer/job.go` — copy `document_review_surfacer/job.go`'s exact structural shape (package doc, `Work`, `run`, `listTenantsWith...`, `surfaceTenant`), swapping in the new ports. No raw SQL in this package.
- [ ] Notification fanout: newly-surfaced overdue stages enqueue an outbox row (existing outbox repo, existing notification event conventions) inside the same per-tenant tx — alert-only, no state mutation on `approval_instances`/`approval_stage_instances` beyond the idempotent `sla_surfaced_at` stamp.
- [ ] Register in `internal/modules/jobs/maintenance/periodic.go`: 4th `river.NewPeriodicJob`, hourly, `PeriodicJobOpts{ID: "approval-sla-surfacer", RunOnStart: false}`.
- [ ] Run the job's tests — PASS. `go build ./...` clean.

## Task 5 — Visibility: new sentinel + predicate

- [ ] Failing integration test first: extend `read_service_tenant_grade_view_integration_test.go` — add `TestLoadInstance_BareViewer_NotEligible_NotFound` (bare `document.view`, not in any stage's `EligibleActorIDs`, no oversee/edit) asserting `errors.Is(err, infrastructure.ErrInstanceNotVisible)`; add `TestLoadInstance_StagePoolMember_Granted`, `TestLoadInstance_ApprovalOversee_Granted`, `TestLoadInstance_DocumentEdit_Granted`. Update the existing "TenantGradeViewer...Granted" tests to seed the viewer INTO the eligible pool (since bare tenant-grade view alone no longer suffices) — document this rewrite in evidence.md, do not silently delete the old assertions' intent (system_admin-bypass and no-grant-denied cases stay as-is).
- [ ] `infrastructure/approval_repository.go`: add `var ErrInstanceNotVisible = errors.New(...)`.
- [ ] `application/read_service.go`: add `instanceVisible(...)` predicate; wire into `LoadInstance` and `LoadActiveInstanceByDocument` (both call it after the existing `authz.Require` gate, before returning the loaded instance).
- [ ] `http/errors.go`: add the `ErrInstanceNotVisible` → 404 case + a new `approvalCodeNotFoundInstanceNotVisible` constant (find the existing `approvalCode*` constant block and match its naming convention).
- [ ] Run the full visibility matrix — PASS. `go build ./...` clean.

## Task 6 — Worklist filters + `scope=oversee`

- [ ] Failing integration test first: extend inbox-related tests — actor eligible on instance A (stage_kind=review) and instance B (stage_kind=approval, due tomorrow) → `stage_kind=review` filter returns only A; `due_before=<yesterday>` returns neither; oversee-holder with `scope=oversee` sees instances they're not personally eligible on; non-oversee actor with `scope=oversee` → `ErrCapDenied`.
- [ ] `application/read_service.go`: extend `listInboxItems`'s SQL + signature with `stageKind, dueBefore *time.Time, scope string` params; oversee branch checks `authz.Require(CapApprovalOversee, "tenant")` inside the tx before swapping the predicate.
- [ ] `http/inbox_handler.go`: parse `stage_kind`, `due_before` (RFC3339), `scope` query params; pass through.
- [ ] `http/contracts/instance_read.go`: `InboxItem` gains `StageKind string`, `DueAt *string`; `StageInstance` gains `StageKind`, `DueAt *string`; `InstanceResponse` gains `FrozenContentHash *string`.
- [ ] `http/get_instance_handler.go`: `mapInstanceResponse` populates the three new fields.
- [ ] Run — PASS. `go build ./...` clean.

## Task 7 — Contract (openapi) + regen

- [ ] Edit `api/openapi/v1/openapi.yaml`: inbox GET route gains `stage_kind`, `due_before`, `scope` query params; nearest instance/inbox response schemas gain `stage_kind`, `due_at`, `frozen_content_hash` properties.
- [ ] Run the module's oapi-codegen regen target; `go build ./...` clean; confirm no unexpected diff beyond the new fields/params.

## Task 8 — Full verification sweep

- [ ] `go build ./...`
- [ ] `go build -tags integration ./...`
- [ ] `go test -count=1 ./internal/modules/documents/approval/...`
- [ ] `go test -count=1 ./internal/modules/jobs/...`
- [ ] `go test -count=1 ./...` (grep zero FAIL)
- [ ] `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 violations)
- [ ] `go test ./scripts/api-lint/...`

## Task 9 — Self-review pass

- [ ] Confirm visibility gating is enforced at the query/repository/service layer (predicate runs inside the same tx as the load, before returning data) — NOT a client-side or post-hoc filter over an already-fetched unscoped result.
- [ ] Confirm no column/field reused across two unrelated semantic purposes (SLA surfacer's `sla_surfaced_at` is a NEW column, not reusing `document_review_surfacer`'s `review_surfaced_at`).
- [ ] Confirm no silently swallowed error path in the new job, predicate, or SQL.
- [ ] Confirm the SLA surfacer is a genuine sibling (own port, own table columns), not a fork sharing `document_review_surfacer`'s table/port.

## Task 10 — Evidence + commit

- [ ] Write `evidence.md` (implementation summary, verification table, judgment calls, bounded defers, self-review confirmations).
- [ ] `git add` explicit touched paths only (never `-A`).
- [ ] Commit: `feat(approval): F8 SLA due dates + surfacer + visibility 404 model + worklist filters (W4/P2/P3/P8)`.
