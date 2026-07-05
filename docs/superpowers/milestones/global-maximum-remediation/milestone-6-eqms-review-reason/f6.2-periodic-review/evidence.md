# Feature F6.2 — Evidence

> **Milestone:** 6  ·  **Feature:** `f6.2-periodic-review`  ·  **Closed:** 2026-07-05
> **Contract:** `spec.md` (consumer contract + Validation Gate) + `../validation-contract.md` §2–§4 (binding, HS-7).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output.

## What was implemented

Review/expiry model on `public.documents` + capability-gated review workflow + River surfacer,
contract-first. Producer matches the `spec.md` consumer contract (FE DTO fields + review-due filter;
`ReviewDueReader` read-port for the jobs surfacer; mark-reviewed application service under
`CapDocumentReview`).

- **Migration (T1)** `db/migrations/0274_document_review_and_reason.sql` — adds `review_due_at`,
  `last_reviewed_at` (+ F6.3's reason columns) and CHECKs `ck_documents_effective_window`,
  `ck_documents_review_due_sane`, `ck_documents_reason_category`. `effective_from`/`effective_to`
  reused from baseline (no duplicate column family). `0276_document_review_surfaced_marker.sql`
  adds `review_surfaced_at` (idempotency marker). — `b8a32144`, `3118027e`.
- **Capability (T2)** `document.review` (ScopeTenant), all 10 IAM touchpoints; registry 34→35;
  M2-**generated** tripwire arm (documents/UPDATE caps now include `document.review`, migration
  `0275`); reference-data grant to area_admin + qms_admin; tier-1 route→capability entry; ADR 0069.
  — `2d5b47b3`.
- **Read-port (T3)** `ReviewDueReader.ListDueForReview(ctx, tx, now, limit)` +
  `repository/review_due_reader.go` — published, currently-effective, `review_due_at <= now`,
  RLS-scoped. — `1e8c66e5`.
- **Surfacer (T4)** `internal/modules/jobs/document_review_surfacer/job.go` — River periodic job,
  queue `maintenance`, leader-elected, `RunOnStart:false`, hourly; `authz.BypassSystem` scheduler
  bypass; reads via `ReviewDueReader`, writes via `ReviewSurfaceWriter.MarkSurfaced` (idempotent).
  Zero raw `documents` SQL in the jobs module. — `3118027e`.
- **Mark-reviewed (T5)** `approval/application/mark_reviewed_service.go` + handler —
  `POST /documents/{id}/review`, `authz.Require(document.review,"tenant")`, published precondition
  (not a status transition; no 11th state), friendly first-line mirrors of the 0274 CHECKs, OCC CAS
  bumping `revision_version`, emits `EventTypeDocumentReviewed`. SchedulePublish also wires
  `effective_to` + `review_due_at`. — `4f0cdf57`.
- **Read-side (T6)** `domain.Document` +4 nullable review/expiry fields; `GetDocument` +
  `ListDocumentsPaginated` SELECT/scan them; `DocumentSummary`/`DocumentDetailResponse` DTOs carry
  them; `GET /documents?review_due=true` filter (mirrors `ListDueForReview` predicate); FE
  `useDocumentArtifact` wires `effectiveFrom`/`nextReviewAt` from the real fields. — `78dc1400`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | per-task subagent TDD (mark_reviewed, surfacer, read-port, read-side) | red→green each task; final suite green below | real |
| Static build | `go build ./...` | `BUILD_DONE=0` (clean) | — |
| Registry bump | `go test -run TestCapabilityRegistrySize\|TestEveryCapabilityClassified ./internal/modules/iam/domain/...` | `ok metaldocs/internal/modules/iam/domain 1.163s` (35 classified) | real |
| M2 tripwire golden + arm parity | `go test -run TestRenderMigration_MatchesCommittedFile\|TestArms\|TestTripwire ./internal/platform/tripwire/...` | `ok metaldocs/internal/platform/tripwire 1.307s` (0275 byte-faithful, arm carries `document.review`) | real |
| Read-side DTO + filter wiring | `go test ./internal/modules/documents/delivery/... ./internal/modules/documents/repository/...` | `ok` both packages | real |
| Wire-contract pin | `go test -run TestDocumentSummaryAndDetail_ReviewFieldsWireContract ./internal/modules/documents/delivery/http/...` | `ok ... 3.564s` | real |
| Contract lint (post-M6) | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | 2 violations — **both pre-existing at M6 base 93cd6114** (verified via worktree); zero M6-introduced | real |
| DB CHECKs enforced | `go test -run TestDocumentReviewCheckConstraints ./internal/modules/documents/repository/... -tags integration` | **executed on real Postgres** (F6.4 gate#7, 2026-07-05) — `TestDocumentReviewCheckConstraints` **PASS 76.80s** | real (testdb) |
| mark-reviewed authz + dates + OCC + tenant isolation | `TestMarkReviewed_{RequiresDocumentReviewCapability,SetsReviewDates,OCCConflict,RejectsNonPublished,TenantIsolation}` (integration) | **executed real DB** (F6.4 gate#7) — all 5 **PASS** (62.02/3.18/6.16/3.85/8.85s). Also proven end-to-end by the live HTTP drive below | real (testdb + live) |
| tripwire arm negative (no-cap UPDATE → P0001) | `TestTripwire_DocumentsUpdate_{DocumentReviewArm,NoCapAssertedIsRejected}` (integration) | **executed real DB** (F6.4 gate#7) — both **PASS** (131.47/2.55s) | real (testdb) |
| surfacer flags due, idempotent, tenant-isolated | `TestIntegration_Surfacer_{FullTick_IteratesAllTenants,Writer_TenantSeed_DoesNotSurfaceOtherTenant,Idempotent_SecondRunNoOp,ReSurfacesAfterReviewDueAdvances}` (integration; **renamed + isolation rewritten by F6.4** to conform to §4.3) | **executed real DB** (F6.4 gate#7) — all 4 **PASS** (120.45/31.06/7.05/10.62s) | real (testdb) |
| read-port used by surfacer; no documents SQL in jobs | grep census + `TestListDueForReview_*` (integration) | census clean (jobs → ports only); `TestListDueForReview_{FiltersPublishedAndDue,TenantIsolation,LimitAndOrder}` **executed real DB** (F6.4 gate#7) — **PASS** | real |
| **Live drive: capability-gated mark-reviewed on a published doc** | `.\scripts\start-api.ps1 -Build` → login → mark-reviewed (see capture below) | **GREEN** — `MARK_REVIEWED_STATUS=200`; `ETag "v3"→"v4"`; `AFTER {last_reviewed:2026-07-05…, review_due:2027-07-05…, rev:4}`; `review_due=true` filter excludes the just-reviewed doc (next-due is future); unauth POST → `401` (tier-1). One real T6 regression caught here (`500 invalid UUID ""` from a missed `documentId`→`id` rename) and root-fixed (`7b3f0f82`). | real (live) |

> **Real-provider proof split (honest labeling).** The live HTTP drive against the real Postgres
> brought up by `start-api.ps1` is the real-provider proof for the **mark-reviewed flow + read-side
> DTO/filter** (headline consumer contract). The deeper DB-invariant proofs (CHECK rejection, OCC
> conflict, cross-tenant isolation) and the **River surfacer** are testdb-factory integration tests
> authored during each task's TDD. At F6.2 close they were **authored-not-executed** (no `DATABASE_URL`
> in the validator shell). **The M6 milestone-validator FAIL fix-feature F6.4 executed them on real
> Postgres on 2026-07-05** — all `--- PASS`; full run inventory + timings in `../f6.4-surfacer-contract-and-consumer/evidence.md`.
> (F6.4 also **conformed** the surfacer to the binding §4.2/§4.3 seed/isolation contract — see that
> feature; the surfacer rows above reflect the conformed, renamed tests.)

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Migration applies; CHECK rejects bad windows/review dates | **yes** — migration applied live (mark-reviewed wrote real dates); CHECK rejection **executed real DB** (F6.4 gate#7, PASS) | `TestDocumentReviewCheckConstraints` row |
| `document.review` reachable only with capability; no-cap UPDATE trips tripwire | **yes** (live: capability path 200, unauth 401); tripwire-arm negative **executed real DB** (F6.4 gate#7, P0001 PASS) | mark-reviewed authz (live) + tripwire negative rows |
| Registry 34→35; M2 arm drift green; arm includes `document.review` | **yes** | registry + tripwire golden rows (real) |
| Surfacer flags due doc; idempotent; tenant-isolated | **yes — executed real DB** (F6.4 gate#7; surfacer conformed to §4.2/§4.3 + isolation PASS) | surfacer integration row |
| Read-port used by surfacer; no documents SQL in jobs | **yes** (census) / **executed real DB** (F6.4 gate#7, `TestListDueForReview_*` PASS) | read-port row |
| mark-reviewed sets dates, published precondition, OCC CAS | **yes (live)** — 200, dates set, ETag v3→v4 (OCC bump) | live drive row + mark-reviewed integration row |
| Contract: response fields + filter + op in openapi; FE type present | **yes** | wire pin + api-lint + FE `useDocumentArtifact` |
| Live drive: capability-gated mark-reviewed end-to-end | **yes** | live drive row (GREEN) |

## Review disposition

- Spec-compliance review: per-task subagent review (sonnet) + main-session diff review before each
  commit. RevisionNumber-unwired defect (F6.3-adjacent) caught in review and root-fixed as T8b.
- Code-quality review: main-session review of each diff by root-cause family; 656-line approval
  regen churn verified as genuine oapi-codegen path-sort reorder (not semantic), not a defect.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Notification/escalation on overdue review | Contract §8 explicitly defers; M6 surfaces + gates only | Trigger: eQMS escalation milestone; owner: distribution/notifications |
| api-lint `SEED-CHOKEPOINT-ALLOWLIST-STALE cancel_service.go:76` | Pre-existing at M6 base (93cd6114, verified); M3 F3.1 concern, out of M6 boundary | Trigger: M3 allowlist hygiene pass; owner: mission tracker |
| api-lint `ASYNC-TENANT-SEED fanout_worker.go:98` | Pre-existing at M6 base (verified); M3 F3.2 concern, out of boundary | Trigger: M3 tenancy pass; owner: mission tracker |
