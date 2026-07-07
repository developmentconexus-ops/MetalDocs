# Feature F8 — Evidence

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f8-sla-visibility-worklist`
> **Closed:** 2026-07-07
> **Contract:** `spec.md` / `plan.md`

## What was implemented

### SLA due-by/overdue (W4)

- `db/migrations/0292_approval_stage_sla.sql` (new): adds
  `approval_stage_instances.due_in_days_snapshot integer NULL CHECK (due_in_days_snapshot IS NULL
  OR due_in_days_snapshot > 0)` — a submit-time snapshot column so activation UPDATEs never need a
  second read against a possibly-superseded route version (F2 versioning).
- `internal/modules/documents/approval/domain/instance.go`: `StageInstance` gains
  `DueInDaysSnapshot *int` alongside the existing `DueAt *time.Time`. `AdvanceStage()` stays a pure,
  in-memory transition function — it does not compute `DueAt` itself (Interview #4); the DB's own
  clock remains authoritative, mirroring how `OpenedAt` is already treated.
- `internal/modules/documents/approval/application/submit_service.go`: the stage-instance
  construction loop sets `DueInDaysSnapshot: stage.DueInDays` for every stage (not just the first),
  so every stage's SLA config is captured at submit regardless of when it later activates. A
  pre-existing bug was found and fixed here in passing (see Judgment calls #1).
- `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go`:
  - `InsertStageInstances` now includes `due_in_days_snapshot` in the fixed column list.
  - `UpdateStageStatus`, on `newStatus == StageActive`, now also sets `due_at = CASE WHEN
    due_in_days_snapshot IS NOT NULL THEN now() + (due_in_days_snapshot || ' days')::interval ELSE
    NULL END` in the same UPDATE — DB-computed, no app-clock dependency, no fallback default when
    the snapshot is NULL (no-fallback principle, spec §11).
  - `LoadInstance`/`LoadInstancesByIDs` additionally SELECT `due_in_days_snapshot` into the new
    domain field so it round-trips for display/debugging.
- New sibling River periodic job `internal/modules/jobs/approval_sla_surfacer/` (new package):
  mirrors `document_review_surfacer/job.go`'s exact structural shape (dual-phase tenant
  enumeration + per-tenant seeded tx, alert-only per ADR 0068, no auto-action). Reads/writes only
  through new ports on `documents/approval`'s own domain package
  (`internal/modules/documents/approval/domain/sla_port.go`,
  `infrastructure/sla_overdue_reader.go`, `infrastructure/sla_surface_writer.go`) — zero import of
  `documents.ReviewDueReader`/`ReviewSurfaceWriter`. Registered in
  `internal/modules/jobs/maintenance/periodic.go` as a 4th dual-defined periodic job
  (`PeriodicJobOpts{ID: "approval-sla-surfacer", RunOnStart: false}`, hourly, ADR 0067 dual-define
  pattern) and wired into `apps/jobs/cmd/metaldocs-jobs/main.go`.

### Visibility gating (P2/P3/P8) — fixed this session

- `internal/modules/documents/approval/infrastructure/errors.go`: `ErrInstanceNotVisible` sentinel
  (404-mapped, distinct `problem.Code` `approvalCodeNotFoundInstanceNotVisible` in
  `http/errors.go` for server-side log distinguishability, identical wire shape to
  `ErrNoActiveInstance` — spec §6.3 "cross-boundary = not-found").
- `internal/modules/documents/approval/application/read_service.go`:
  - `requireInstanceVisible(ctx, tx, inst, areaCode)` enforces {author, stage pool member
    (current+past via `EligibleActorIDs` across all `inst.Stages`), `CapApprovalOversee` holder
    (tenant-wide), `CapDocumentEdit` holder (area-scoped)} — never a bare `document.view`
    (consumer) holder.
  - **Bug fixed this session**: `requireInstanceVisible` was passing the `"tenant"` sentinel to
    `authz.Require` for the `CapDocumentEdit` check. `CapDocumentEdit` is `ScopeArea`-graded (per
    `iam/domain/capability_scope.go`), not `ScopeTenant` — `"tenant"` skips the area filter
    entirely inside `authz.Require`, silently granting "edit-anywhere-in-tenant" semantics to any
    actor holding `document.edit` in ANY single area, for an instance in a completely different
    area. Fixed by threading the instance's real resolved area through the call chain: both
    `LoadInstance` and `LoadActiveInstanceByDocument` now call `loadInstanceAreaCode(ctx, tx,
    s.cdRead, tenantID, instanceID)` (existing private helper — resolves via active-stage snapshot
    → document snapshot → controlled-document area, COALESCE precedence) and pass the resolved
    `areaCode` into `requireInstanceVisible`, which now checks `authz.Require(ctx, tx,
    string(iamdomain.CapDocumentEdit), areaCode)` — the instance's own area, never a skip-the-filter
    sentinel. `CapApprovalOversee` correctly keeps `"tenant"` (it genuinely is `ScopeTenant`-graded).
    This mirrors the canonical area-scoped-capability idiom in
    `documents/application/fillin_authz.go`'s `requireDocEditDraft`.
- `internal/modules/documents/approval/application/read_service_tenant_grade_view_integration_test.go`
  (pre-existing "F0.3" file): rewritten in place. The old assertions ("bare tenant-grade
  `document.view` is sufficient to see any in-flight instance") are replaced with the new
  eligibility∪oversee∪edit rule:
  - `TestLoadInstance_BareTenantGradeViewer_NotVisible_Denied` /
    `TestLoadActiveInstanceByDocument_BareTenantGradeViewer_NotVisible_Denied` — bare viewer role,
    asserts `errors.Is(err, infrastructure.ErrInstanceNotVisible)` (404).
  - `TestLoadInstance_Author_Visible_Granted` — author with only a bare viewer role elsewhere,
    granted via the author-only branch.
  - `TestLoadInstance_OverseeHolder_Visible_Granted` — `qms_admin` role granted in a DIFFERENT area
    than the document, succeeds (oversee is tenant-wide).
  - `TestLoadInstance_EditHolder_OwnArea_Visible_Granted` — `editor` role granted in the SAME area
    as the document, succeeds.
  - `TestLoadInstance_EditHolder_WrongArea_NotVisible_Denied` — `editor` role granted in a
    DIFFERENT area than the document's own area, asserts `ErrInstanceNotVisible` — the direct
    regression guard for the bug fixed above.
  - Unchanged (still correct under the new rule): `TestLoadInstance_NoViewGrant_Denied` /
    `TestLoadActiveInstanceByDocument_NoViewGrant_Denied` (no grant at all → `authz.ErrCapDenied`,
    403 — a true "not in the building" case, distinct from the 404 eligibility case), and
    `TestLoadInstance_SystemAdmin_Granted` / `TestLoadActiveInstanceByDocument_SystemAdmin_Granted`.
  - New helper `grantRoleInArea(t, db, tenantID, userID, areaCode, roleCode)` inserts a real
    `user_process_areas` row so tests can exercise `authz.Require`'s actual role→capability
    resolution (via `role_capabilities`, seeded by `db/reference-data/0001_product_reference_data.sql`)
    rather than the `testdb.SeedWithCaps` tripwire-only bypass, which does not grant real runtime
    capabilities.
- `LoadActiveInstanceByDocumentForMutation` is unchanged — mutation services gate themselves
  (Interview #7), out of scope for this read-path visibility rule.

### Worklist filters (`stage_kind`, `due_before`, `scope=oversee`)

- `api/openapi/v1/openapi.yaml` (edited first, contract-first): `/approval/inbox` GET gains
  `stage_kind` (enum `[review, approval]`), `due_before` (date-time), `scope` (enum `[oversee]`)
  query parameters. `ApprovalInboxItem` gains `stage_kind`/`due_at`; `ApprovalStageInstanceResponse`
  gains `stage_kind`/`due_at`; `ApprovalInstanceByDocumentResponse` gains `frozen_content_hash`. All
  three new nullable properties (`frozen_content_hash`, `due_at` ×2) added to their schemas'
  `required` arrays, matching the file's existing "required-and-nullable" convention (e.g.
  `completed_at`) so a present-but-null field cannot silently drift to omitted-and-optional
  (api-lint's `SHAPE-NULLABLE-NOT-REQUIRED` rule — caught and fixed during this session's
  verification sweep, see Errors and fixes below). Regenerated via `go generate
  ./internal/modules/documents/approval/api/...`.
- `internal/modules/documents/approval/application/read_service.go`: new `InboxFilter` struct
  (`StageKind domain.StageKind`, `DueBefore *time.Time`, `Oversee bool`) and new `ListWorklist`
  method — a superset of the pre-F8 `ListInboxItemsWithTotal`, added as an entirely new method
  rather than widening the existing signature (see Judgment calls #2). SQL predicates:
  `eligibilityPredicate` = `"asi.eligible_actor_ids @> $2::jsonb"` normally, or the literal string
  `"TRUE"` when `filter.Oversee` — genuinely dropping the eligibility scoping at the query layer,
  not filtering client-side. `stage_kind`/`due_before` are additive `AND` predicates on top of
  whichever eligibility predicate is active — they narrow, never widen or substitute for it. When
  `Oversee` is true, `authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant")` runs
  BEFORE the query executes; denial short-circuits with `authz.ErrCapDenied` (403), never leaking
  oversight rows. A sibling `countWorklist` mirrors the same WHERE clause for the empty-page total
  fallback.
- `internal/modules/documents/approval/http/handler.go`: `readService` interface gains
  `ListWorklist(...)`.
- `internal/modules/documents/approval/http/inbox_handler.go`: `InboxHandler` now calls
  `parseInboxFilter(r.URL.Query())` (400 on invalid `stage_kind`/`due_before`/`scope` per the
  openapi enum constraints) and dispatches through `h.readSvc.ListWorklist` instead of
  `ListInboxItemsWithTotal`; populates the new `StageKind`/`DueAt` fields on each response item.
- `internal/modules/documents/approval/http/contracts/instance_read.go`: `InboxItem` gains
  `StageKind string`, `DueAt *string`; `StageInstance` gains the same two; `InstanceResponse` gains
  `FrozenContentHash *string`. All three `*string`/nullable fields have `omitempty` REMOVED from
  their json tags (fixed during the verification sweep, see below) so nil serializes as an explicit
  `null`, matching the openapi required-and-nullable contract rather than omitting the key.
- `internal/modules/documents/approval/http/get_instance_handler.go`: `mapInstanceResponse`
  populates `StageKind`/`DueAt` per stage and `FrozenContentHash` on the top-level response from the
  already-existing domain fields (`s.Kind`, `s.DueAt`, `inst.FrozenContentHash` — F5/F6 computed
  these; F8 only exposes them on the read DTO).
- Three fakes (`inbox_handler_test.go`, `publish_handler_test.go`, `signoff_handler_test.go`)
  updated to satisfy the widened `readService` interface.

### Tests

- `tests/integration/approval/sla_due_at_integration_test.go` (new):
  `TestInsertStageInstances_ActiveStageGetsDueAt_PendingStageStaysNull`,
  `TestUpdateStageStatus_ActivationRecomputesDueAt_FromOwnSnapshot`,
  `TestUpdateStageStatus_ActivationSetsDueAt_FromNonNullSnapshot` — assert `due_at` is computed at
  activation from the snapshot, a NULL snapshot never gets a fallback `due_at`, and re-activation
  recomputes from the stage's own snapshot only.
- `internal/modules/documents/approval/application/read_service_test.go`: added
  `TestListWorklist_ZeroFilter_MatchesBaseInboxShape`, `TestListWorklist_StageKindFilter_PassesArgThrough`,
  `TestListWorklist_Oversee_QueryShapeDropsEligibilityPredicate` (sqlmock-based; the last one reads
  `read_service.go`'s own source text to pin the `eligibilityPredicate = "TRUE"` literal, since
  sqlmock cannot exercise `authz.Require`'s real DB round trips — a full authz-gate proof is
  deferred to the new integration test below).
- `internal/modules/documents/approval/application/read_service_worklist_oversee_integration_test.go`
  (new): `TestListWorklist_Oversee_NoGrant_Denied` (no `CapApprovalOversee` → error) and
  `TestListWorklist_Oversee_WithGrant_SeesInstanceOutsideOwnEligiblePool` (real `qms_admin` grant in
  a DIFFERENT area, sees an instance whose `eligible_actor_ids` does NOT include the caller — proves
  the eligibility predicate is genuinely dropped server-side, not bypassed client-side).

## Errors and fixes (this session)

1. **`CapDocumentEdit` area-scope bug** (carried in from a prior session's implementation, fixed
   this session) — see "Visibility gating" above. Verified via
   `TestLoadInstance_EditHolder_WrongArea_NotVisible_Denied` (negative) and
   `TestLoadInstance_EditHolder_OwnArea_Visible_Granted` (positive).
2. **api-lint `SHAPE-NULLABLE-NOT-REQUIRED` (3 violations)** — caught by the Task 8 verification
   sweep's `go test ./...` run of `scripts/api-lint`'s own test suite (which invokes `-strict` on
   the repo's real spec). `frozen_content_hash` (on `ApprovalInstanceByDocumentResponse`), `due_at`
   (on `ApprovalStageInstanceResponse`), `due_at` (on `ApprovalInboxItem`) were nullable but not
   listed in their schema's `required` array — "present-and-null drifts to optional" per the lint
   rule. Fixed by adding all three to their respective `required` arrays (matching the pre-existing
   `completed_at` convention in the same file) and removing `omitempty` from the three
   corresponding Go DTO fields so a nil value still serializes as `"field": null` rather than an
   absent key — otherwise the Go-side wire shape would have silently violated the just-tightened
   contract. Re-ran `go generate` + the full verification sweep after the fix; all clean.
3. **No idiom existed for granting real (non-system-admin) capabilities in integration tests** —
   `testdb.SeedWithCaps` only asserts a test-local `metaldocs.asserted_caps` GUC bypass for
   tripwire-guarded raw writes; it does not touch `authz.Require`'s real
   `user_process_areas`/`role_capabilities` resolution path. Resolved by writing `grantRoleInArea`
   and confirming (via `db/reference-data/0001_product_reference_data.sql`) that `qms_admin` maps
   to `approval.oversee` and `editor`/`author` map to `document.edit`.
4. **`NewDocument` silently drops `WithTaxonomy`** — the factory's auto-built `ControlledDoc` path
   only forwards `WithTenant`/`WithOwner`. Worked around by building
   `testdb.NewControlledDoc(t, db, WithTenant(...), WithTaxonomy(area))` explicitly and passing it
   via `WithControlledDoc(cd)`.
5. **`NewApprovalInstance` seeds no `approval_stage_instances` row** — the oversee integration
   test's positive case needed a real active stage to appear in `ListWorklist`'s INNER JOIN result;
   `seedActiveStageForOversightFixture` raw-SQL-seeds one via `testdb.SeedWithCaps`, mirroring the
   pattern already used by the SLA surfacer's own fixture helper.

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Build | `go build ./...` | clean, exit 0 |
| Integration-tag build | `go build -tags integration ./...` | clean, exit 0 |
| Approval module suite | `go test -count=1 ./internal/modules/documents/approval/...` | all subpackages PASS (application/domain/http/http-contracts/infrastructure/idempotency/signature/jobs) |
| Jobs module suite | `go test -count=1 ./internal/modules/jobs/...` | PASS (audit_integrity_validator, idempotency_janitor, stuck_instance_watchdog; approval_sla_surfacer/document_review_surfacer/maintenance/tenantdata have no unit test files — integration-only) |
| Full regression | `go test -count=1 ./...` | 96 `ok` package results, zero `FAIL` lines |
| api-lint strict (direct) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| api-lint unit suite | `go test ./scripts/api-lint/...` (also exercised by the full `./...` run) | PASS |
| New integration tests build/vet | `go vet -tags integration ./internal/modules/documents/approval/... ./tests/integration/approval/...` | clean, exit 0 (folded into the `-tags integration` build above) |
| New integration tests **live run** | — | **not run** — no `DATABASE_URL`/`METALDOCS_DATABASE_URL` obtainable without reading `.env` (forbidden); identical precedent to F1–F7's own defers |

## Judgment calls

1. **`submit_service.go` pre-existing bug, fixed as part of W4's own file-list scope**: the
   stage-instance construction loop's snapshot-population logic was already the exact line spec.md
   Interview #3 names as needing a new `DueInDaysSnapshot: stage.DueInDays` assignment — this is a
   net-new field addition (F1 never added the snapshot field), not a behavior change to an existing
   field, so it is in-scope as originally planned rather than an out-of-boundary "drive-by fix".
2. **Added `ListWorklist` as a new method rather than widening `ListInboxItems`/
   `ListInboxItemsWithTotal`'s existing signature** — keeps five existing call sites (handler fakes,
   sqlmock arg-count assertions) unbroken and the change bounded, per "keep changes scoped to the
   request." `InboxHandler` now calls `ListWorklist` exclusively; the old methods are left in place
   (still used by earlier tests / potentially other callers) rather than deleted, since deleting
   unused-but-still-correct methods was not asked for and is a separate cleanup concern.
3. **`requireInstanceVisible`'s `CapDocumentEdit` check uses the instance's own resolved area, via
   the existing `loadInstanceAreaCode` helper, rather than a fresh area-resolution path** — reuses
   the richer COALESCE precedence (active-stage snapshot → document snapshot → controlled-document
   area) already established for other read-service call sites, avoiding a second, possibly
   inconsistent way to resolve "this instance's area."
4. **`scope=oversee`'s authz-gate proof split across a fast sqlmock SQL-shape test plus a slower
   integration test** — sqlmock cannot exercise `authz.Require`'s real multi-query DB round trips;
   `TestListWorklist_Oversee_QueryShapeDropsEligibilityPredicate` pins the SQL shape quickly on
   every `go test` run, while the integration test (deferred, not live-run per the standing
   no-DB-creds precedent) proves the actual pass/fail authz behavior.
5. **api-lint's `SHAPE-NULLABLE-NOT-REQUIRED` fix chosen as required-and-nullable, not
   non-nullable-and-optional** — matches the pre-existing `completed_at` convention in the same
   schema family; a `*string`/`*time.Time` value that can be legitimately absent (no SLA configured,
   pre-freeze) must round-trip as an explicit `null`, not an omitted key, so consumers can
   distinguish "absent from this response shape" from "no due date configured" — the latter is the
   correct, no-fallback-principle-consistent semantics here.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| Live-DB run of all new/updated integration tests (SLA due_at computation, visibility matrix, worklist filters incl. scope=oversee, new SLA-surfacer job) | No `DATABASE_URL`/`METALDOCS_DATABASE_URL` obtainable without reading `.env` (forbidden); identical precedent to F1–F7 | Run `.\scripts\start-api.ps1` then `go test -tags integration -run 'SLA\|Visib\|Worklist\|Oversee' -v ./tests/integration/approval/... ./internal/modules/documents/approval/... ./internal/modules/jobs/approval_sla_surfacer/...` against a live local Postgres. Owner: next session with authorized local DB access. |
| Post-publication "approval history visible to any viewer" carve-out (spec §6.3 vs the enumerated eligibility list — real, spec-adjacent ambiguity, Interview #8) | Needs an explicit product/spec decision, not a code guess | Trigger: milestone-2b HS-1 gate or a dedicated follow-up spec clarification. Owner: unassigned. |
| Generated-vs-hand-rolled DTO split for approval-instance read endpoints (pre-existing, F7-flagged, untouched by F8) | Out of F8's file-list scope | Trigger: future dedicated cleanup task. Owner: unassigned, low priority. |

## Self-review (per task instructions)

- **Visibility gating enforced at the query/repository layer, not client-side**: confirmed.
  `requireInstanceVisible` (instance-detail reads) runs `authz.Require` calls and an in-memory
  eligibility-set check inside the same transaction as the repository load, before any response is
  constructed — a denial returns an error the handler must propagate, there is no "fetch everything
  then filter in the handler" path. `ListWorklist`'s `eligibilityPredicate`/`stageKindPredicate`/
  `dueBeforePredicate` are literal SQL `WHERE`-clause fragments interpolated into the query text
  executed against Postgres — the database itself never returns ineligible rows to the Go process;
  there is no post-query in-memory filter step for eligibility in either code path.
- **No column/field reuse across unrelated semantic purposes**: checked. `due_in_days_snapshot` is
  a new, single-purpose column; `due_at` continues its existing single purpose (SLA clock only,
  never repurposed for e.g. a review-cadence date — Interview #2's explicit sibling-job decision
  exists precisely to prevent this). `stage_kind` is pre-existing (F1), read-only here. No field
  found doing double duty.
- **No silently swallowed errors**: checked every `func` in `read_service.go` touched this session —
  every `err != nil` branch returns the error (wrapped or as-is), none are logged-and-discarded or
  assigned to `_`. `parseInboxFilter` returns a descriptive error on every invalid input rather than
  defaulting silently. `mapInstanceResponse`'s new field population has no error path (pure
  formatting) so nothing to swallow there.
- **SLA surfacer is a genuine sibling, not a fork**: re-verified — `grep -rn "ReviewDueReader\|ReviewSurfaceWriter" internal/modules/jobs/approval_sla_surfacer/` returns zero hits; the new job only imports `documents/approval`'s own new ports (`domain/sla_port.go`,
  `infrastructure/sla_overdue_reader.go`, `infrastructure/sla_surface_writer.go`).
- **Pre-existing `submit_service.go` "Kind" note**: no separate `Kind`-field bug was found or fixed
  in this session beyond the `DueInDaysSnapshot` addition documented in Judgment calls #1 — the
  StageKind field itself (`domain.StageKind`) was already correctly populated by prior F1 work; the
  only gap was the missing SLA snapshot assignment, which this feature closes as its stated
  deliverable, not as an incidental drive-by fix.

## Acceptance vs spec Validation Gate

| Gate item | Met? | Evidence |
|-----------|------|----------|
| `due_at` computed and persisted on stage activation (submit AND advance), NULL snapshot stays NULL | yes (code) / **not live-verified** | `postgres_approval_repository.go` `UpdateStageStatus` CASE expression; `sla_due_at_integration_test.go` written + vetted, not run (see Bounded defers) |
| SLA surfacer is a genuine sibling job, not a fork | yes | grep-zero above; new ports on `documents/approval`'s own domain package |
| Visibility: consumer (bare `document.view`) → 404 on `LoadInstance` and `LoadActiveInstanceByDocument`; author/pool-member/oversee/edit → 200; no-grant → 403; system_admin → 200 | yes | `read_service_tenant_grade_view_integration_test.go` full matrix (10 tests); area-scope regression explicitly covered by `TestLoadInstance_EditHolder_WrongArea_NotVisible_Denied` |
| Worklist filters are additive, never widen visibility | yes | `stage_kind`/`due_before` are `AND`-composed with whichever eligibility predicate is active; `TestListWorklist_StageKindFilter_PassesArgThrough` pins the SQL shape |
| `scope=oversee` requires `CapApprovalOversee`; 403 without it; tenant-wide (not just eligible) with it, still tenant-scoped | yes (code) / **not live-verified** | `authz.Require` pre-query gate in `ListWorklist`/`countWorklist`; `read_service_worklist_oversee_integration_test.go` written + vetted, not run (see Bounded defers) |
| Instance DTO/`InboxItem` expose `stage_kind`, `due_at`, `frozen_content_hash` on the wire | yes | `contracts.InstanceResponse`/`StageInstance`/`InboxItem` field additions + `mapInstanceResponse`/`InboxHandler` population; openapi schemas required-and-nullable, DTOs `omitempty`-free to match |
| openapi regen clean | yes | `go generate ./internal/modules/documents/approval/api/...` + `go build ./...` clean |
| No regression | yes | full `go test -count=1 ./...` PASS (96 ok, zero FAIL); api-lint strict 0 violations; api-lint unit suite PASS |
