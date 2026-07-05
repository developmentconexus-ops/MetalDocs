# Feature F6.4 — Evidence

> **Milestone:** 6  ·  **Feature:** `f6.4-surfacer-contract-and-consumer` (HS-4 fix-feature)  ·  **Closed:** 2026-07-05
> **Contract:** `spec.md` (D1/D2/D3 + Validation Gate 1–8) + `../validation-contract.md` §4/§6(d) (binding, HS-7).
> **Resolves:** `../qa/milestone-qa.md` FAIL (findings 1 blocking, 2 blocking, 3 non-blocking) + C7 fix-req #4 (execute the authored-not-executed integration suite).
> **Real-DB proofs executed by the main session** on the local Postgres (docker `metaldocs-postgres`, `metaldocs_app`), DSN built from `.env` via PowerShell (`[uri]::EscapeDataString` on `PGPASSWORD`; **secret never printed** — only `pw_len`/host/port/db echoed). HEAD `d91d3bb8`.

## What was implemented

Made the shipped F6.2 River surfacer **conform** to the binding §4 consumer contract and gave its
`review_surfaced_at` side effect a real consumer — no contract re-open, no HS-7 erratum (conform was
feasible). Four commits:

- **Task A — surfacer conforms to §4.2/§4.3 (per-tenant seed).** `7d398d92`. `document_review_surfacer/job.go`
  `run()` now: (1) one cross-tenant **system read** under `authz.BypassSystem`, GUC-unset
  (`ListTenantsWithDueReviews`, mirrors `stuck_instance_watchdog.listStuckInstances`); (2) **per tenant**
  a fresh tx → `authz.BypassSystem` → `authz.SeedTxTenant(tenantID)` → `ListDueForReview` (count/log) +
  `MarkSurfaced` → commit, aggregating with `errors.Join`. No unseeded tenant-scoped write survives.
  New read-port `ReviewDueReader.ListTenantsWithDueReviews`. False "mirrors watchdog sweep" comment
  deleted; the code now genuinely mirrors the watchdog's per-item `SeedTxTenant`.
- **Task B — filter reads the marker (worklist) + DTO exposure.** `400714ab`. `repository.go`
  `buildDocumentFilter` `ReviewDue` branch is now the **surfaced-worklist** predicate
  (`… AND review_surfaced_at IS NOT NULL AND review_surfaced_at >= review_due_at`); `mark-reviewed`
  advancing `review_due_at` auto-expels. `review_surfaced_at` added to `Document` +
  `DocumentSummary`/`DocumentDetailResponse` DTOs (contract-first openapi + BE/FE regen);
  `GetDocument`/`ListDocumentsPaginated` SELECT/scan it; handler nil-safe map.
- **Task A2 — explicit `tenant_id` predicate (correct-by-construction isolation).** `bf9eadaf`. The two
  tenant-scoped review-due queries relied **solely** on FORCE RLS for scoping — the only such outliers in
  the documents module. Under the dev/CI role (`metaldocs_app`, **superuser + BYPASSRLS** → FORCE RLS
  inert) the §4.3 isolation tests leaked tenant B. Added `tenant_id = $2::uuid` to `ListDueForReview` +
  `MarkSurfaced` (RLS stays the backstop, per M3/ADR 0027 — matches `active_instance_reader`/`fillin`/
  `export`); surfacer threads the loop `tenantID`. **Global-maximum rationale** in `spec.md` D-decisions +
  the commit body. Bundled real-DB fixes surfaced by the suite: `ReviewDueView.Code *string` +
  `sql.NullString` scan (NULLABLE `documents.code`); F6.3 audit query drops `::uuid` on TEXT
  `resource_id`; `testdb.WithSubmitReadySnapshots()` opt; surfacer isolation test `WithBackgroundBypass`.
- **Task C / gate#7 — schedule-publish proof identity fix.** `d91d3bb8`. See gate #7 row.

## Verification (real, honestly-labeled)

### Deterministic gates (no DB) — `scratchpad/m6_gates.log`

| Gate | Command | Result |
|------|---------|--------|
| Build | `go build ./...` | exit 0, no output |
| Vet | `go vet ./...` | exit 0 |
| DTO wire-contract pin (gate #5) | `go test -run TestDocumentSummaryAndDetail_ReviewFieldsWireContract ./internal/modules/documents/delivery/http/...` | `ok … 3.704s` |
| Registry size 35 + classify | `go test -run 'TestCapabilityRegistrySize\|TestEveryCapabilityClassified' ./internal/modules/iam/domain/...` | `ok … 1.171s` |
| Tripwire golden + arm parity | `go test ./internal/platform/tripwire/...` | `ok … 0.305s` |
| api-lint strict (gate #8) | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | 2 violations — `SEED-CHOKEPOINT-ALLOWLIST-STALE cancel_service.go:76`, `ASYNC-TENANT-SEED fanout_worker.go:98`; **both pre-existing at M6 base `93cd6114`** (validator C2-verified), **zero F6.4-introduced** |

### Integration suite executed on real Postgres (gate #7) — `scratchpad/m6_itest_v.log`, `m6_f62_v.log`, `m6_sched_fix.log`, `m6_tripwire.log`

All `-tags integration`, `-count=1`, real testdb clones. **`--- PASS` for every test** (was
`--- SKIP`/authored-not-executed at validator time — no DB):

| Package | Test | Result |
|---------|------|--------|
| documents/repository | `TestListDueForReview_TenantIsolation` (§4.3 read isolation, was RED under BYPASSRLS) | PASS 6.91s |
| documents/repository | `TestListDueForReview_FiltersPublishedAndDue` / `_LimitAndOrder` | PASS 7.83s / 8.25s |
| documents/repository | `TestIntegration_ReviewDueFilter_ReadsSurfaced` (gate #4: surface→filter includes→mark-reviewed→excludes) | PASS 12.91s |
| documents/repository | `TestListDocumentsPaginated_ReviewDueFilter` | PASS 157.20s |
| documents/repository | `TestDocumentReviewCheckConstraints` (0274 CHECK rejections + `ck_documents_reason_category`) | PASS 76.80s |
| documents/repository | `TestGetDocument_ReturnsReviewAndExpiryFields` | PASS 0.50s |
| documents/approval/application | `TestSubmitPersistsReason_RealDB` (F6.3 reason-persist) | PASS 122.30s |
| documents/approval/application | `TestSubmitReasonOnAuditTrail_RealDB` (F6.3 reason on `governance_events`) | PASS 20.48s |
| documents/approval/application | `TestMarkReviewed_RequiresDocumentReviewCapability` | PASS 62.02s |
| documents/approval/application | `TestMarkReviewed_SetsReviewDates` / `_OCCConflict` / `_RejectsNonPublished` / `_TenantIsolation` | PASS 3.18s / 6.16s / 3.85s / 8.85s |
| documents/approval/application | `TestSchedulePublish_PersistsEffectiveToAndReviewDueAt` / `_EffectiveToAndReviewDueAtOptional` | PASS 146.81s / 2.66s |
| jobs/document_review_surfacer | `TestIntegration_Surfacer_Writer_TenantSeed_DoesNotSurfaceOtherTenant` (§4.3 write isolation, was RED) | PASS 31.06s |
| jobs/document_review_surfacer | `TestIntegration_Surfacer_FullTick_IteratesAllTenants` | PASS 120.45s |
| jobs/document_review_surfacer | `TestIntegration_Surfacer_Idempotent_SecondRunNoOp` (gate #3) | PASS 7.05s |
| jobs/document_review_surfacer | `TestIntegration_Surfacer_ReSurfacesAfterReviewDueAdvances` | PASS 10.62s |
| tests/integration/documents | `TestTripwire_DocumentsUpdate_DocumentReviewArm` (P0001 arm) | PASS 131.47s |
| tests/integration/documents | `TestTripwire_DocumentsUpdate_NoCapAssertedIsRejected` (P0001 negative) | PASS 2.55s |

> **Runtime-truth finding (recorded).** The dev/CI DB role `metaldocs_app` is `rolsuper=t rolbypassrls=t`
> (verified via `pg_roles`), so `FORCE ROW LEVEL SECURITY` is **inert** for it. The two §4.3 isolation
> tests were therefore false-negatives under sole-RLS scoping — this is precisely why Task A2's explicit
> `tenant_id` predicate is the global maximum (isolation provable independent of role privilege). Two
> F6.2 schedule-publish proofs (authored-not-executed at close) also failed first run on
> `authz: metaldocs.actor_id GUC not set` — a **test-setup** gap (bare ctx, no identity for TxRunner to
> seed), fixed in `d91d3bb8` (production authn middleware seeds identity; the test now mirrors it). No
> production code defect surfaced by the suite.

## Acceptance vs spec Validation Gate

| # | Acceptance (spec.md) | Met? | Evidence |
|---|----------------------|------|----------|
| 1 | Surfacer seeds per-tenant; no unseeded tenant-scoped write | **yes** | `job.go` per-tenant `SeedTxTenant`+explicit predicate; `7d398d92`+`bf9eadaf`; census: surfacer holds no GUC-unset write path |
| 2 | Cross-tenant **isolation** (matches §4.3, not its inverse) | **yes (real DB)** | `TestIntegration_Surfacer_Writer_TenantSeed_DoesNotSurfaceOtherTenant` + `TestListDueForReview_TenantIsolation` PASS |
| 3 | Idempotent (twice → once) under per-tenant model | **yes (real DB)** | `TestIntegration_Surfacer_Idempotent_SecondRunNoOp` PASS |
| 4 | `review_due=true` returns **surfaced** set; mark-reviewed expels | **yes (real DB)** | `TestIntegration_ReviewDueFilter_ReadsSurfaced` PASS |
| 5 | `review_surfaced_at` on DTO; pin | **yes** | `TestDocumentSummaryAndDetail_ReviewFieldsWireContract` PASS; openapi + BE/FE regen; api-lint clean of new violations |
| 6 | Due-core predicate single-sourced | **yes** | `const dueCorePredicate` (review_due_reader.go); referenced by `ListDueForReview`, `ListTenantsWithDueReviews`, `MarkSurfaced` (grep census) |
| 7 | Integration suite executed on real Postgres | **yes** | full table above — every M6 DB proof `--- PASS` |
| 8 | Build/vet/registry/tripwire/pins green | **yes** | deterministic table above |

## Review disposition

- Task A + Task B: fresh implementer subagent (sonnet) TDD; main-session diff review before each commit.
- Task A2: fresh implementer subagent (sonnet), main-session diff review (placeholder numbering,
  `dueCorePredicate` single-source pinned, correct-by-construction isolation reasoning) before commit.
- All real-DB failures diagnosed to root cause (superuser/BYPASSRLS; NULLABLE code; TEXT resource_id;
  snapshot-on-submit; bare-ctx identity) and fixed by family — no symptom patches; no production code
  touched to make a test pass.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| **Sole-RLS tenant-scoping false-negative in dev/CI** (class defect beyond the 2 ports F6.4 fixed) | **Prod is safe** — dev/CI `metaldocs_app` is superuser+BYPASSRLS (FORCE RLS inert), but prod is NOSUPERUSER+NOBYPASSRLS per ADR 0022 Phase-5 §7 / ADR 0027 (documented in migrations `0234`/`0237`), so RLS is active in prod. The residual is a **test false-negative**: any other tenant-scoped query relying solely on RLS gives a false-green isolation test in dev. Sweeping the whole system is out of the M6 boundary and is a milestone-class fix (flip CI role → suite-as-census → explicit predicate + lint; **trap:** FORCE doesn't apply to the table *owner*, so CI role must be non-owner). Not a live vuln → defer, don't reopen the signed M3 gate or the PASSED M6 gate. | Trigger: **M7 F7.4 rls-truth-sweep** (mission.md §7 M7); owner: mission tracker |
| ASYNC-TENANT-SEED lint **port blind spot** | Lint can't see writes behind a cross-module port (documented function-local scan). The §4.2 human backstop that caught this surfacer is now satisfied; hardening the lint is a real follow-up out of F6.4 boundary | Trigger: lint-hardening / **M7 F7.4** tenancy sweep (pairs with the RLS-truth fix); owner: mission tracker |
| Notification/escalation on overdue review | Contract §8 defers to M8; F6.4 surfaces + gates only | Trigger: eQMS escalation milestone; owner: distribution/notifications |
| api-lint `cancel_service.go:76` / `fanout_worker.go:98` | Pre-existing at M6 base (validator-verified); M3 F3.1/F3.2 concern | Trigger: M3 allowlist/tenancy pass; owner: mission tracker |
