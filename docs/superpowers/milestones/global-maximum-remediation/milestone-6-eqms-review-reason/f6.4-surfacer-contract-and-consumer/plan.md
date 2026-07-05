# Feature F6.4 — Plan

> Engine: `superpowers:subagent-driven-development` — one fresh implementer subagent per task
> (sonnet implement+review), main session orchestrates + reviews diff + commits. TDD: failing
> integration/pin test first, then green.

## Task A — Surfacer conforms to §4.2/§4.3 (per-tenant seed + writable isolation proof)

**Files:**
- `internal/modules/documents/domain/review_due_port.go` (or wherever `ReviewDueReader` lives) — add
  published read-port method `ListTenantsWithDueReviews(ctx, tx, now) ([]string, error)` (distinct
  tenant_ids with ≥1 due doc; system-level cross-tenant read).
- `internal/modules/documents/repository/review_due_reader.go` — impl `ListTenantsWithDueReviews`
  using the **due-core** predicate (extract it to one const `dueCorePredicate` here, D3); keep
  `ListDueForReview` referencing the same const.
- `internal/modules/documents/repository/review_surface_writer.go` — `MarkSurfaced` keeps its
  no-tenant-predicate SQL (RLS scopes it via the seeded GUC); **delete** the false "mirrors
  watchdog sweep / all tenants" comment block (lines ~41-52, 72-81 analog), replace with the
  per-tenant-seeded contract note.
- `internal/modules/jobs/document_review_surfacer/job.go` — `run()` becomes: (1) bypass tx →
  `ListTenantsWithDueReviews` → tenantIDs (cross-tenant system read, like
  `watchdog.listStuckInstances`); (2) **per tenant**: new tx → `authz.SeedTxTenant(tenantID)` →
  `ListDueForReview` (count/log) + `MarkSurfaced` → commit. No unseeded write survives. Fix the
  header comment (no more "sweeps every tenant in one UPDATE").
- `internal/modules/jobs/document_review_surfacer/job_integration_test.go` — **rewrite**
  `TestIntegration_Surfacer_CrossTenant_*`: seed A-due + B-due + A-not-due; run the surfacer;
  assert A-due surfaced, A-not-due untouched, **B-due surfaced too** (full tick covers all tenants
  by iterating) — BUT add a NEW direct-isolation unit at the writer level: seed **only tenant A**
  (`SeedTxTenant(A)`), call `MarkSurfaced`, assert **B-due untouched** (RLS blocked it) → the real
  §4.3 "not surfaced under tenant-A identity" proof. Idempotency test stays.
- Composition root wiring for the new port method if the interface is injected.

**Test strategy:** integration (testdb) for the isolation + idempotency; the writable isolation
assertion is the one that proves §4.3. Failing test first (assert B untouched under A-seed → currently
the all-tenant sweep surfaces B → red), then implement per-tenant → green.

## Task B — Filter reads the marker (worklist) + DTO exposure + predicate single-source

**Files:**
- `internal/modules/documents/repository/repository.go` — `buildDocumentFilter` `ReviewDue` branch
  (~485): change predicate to the **surfaced-worklist** form: `status='published' AND effective
  window AND review_surfaced_at IS NOT NULL AND review_surfaced_at >= review_due_at`. Reference the
  shared effective/published fragment.
- `api/openapi/v1/openapi.yaml` — add `review_surfaced_at` (nullable date-time) to `DocumentSummary`
  + `DocumentDetailResponse`; regen BE (`documents/api/api.gen.go`) + FE (`api-types/index.d.ts`).
- `internal/modules/documents/domain/model.go` — `Document.ReviewSurfacedAt *time.Time`.
- `internal/modules/documents/repository/repository.go` — SELECT/scan `review_surfaced_at` in
  `GetDocument` + `ListDocumentsPaginated` (like the T6 four fields).
- `internal/modules/documents/delivery/http/handler.go` — nil-safe map into both DTOs.
- Pin test: extend `TestDocumentSummaryAndDetail_ReviewFieldsWireContract` for the new field.
- Integration: `TestIntegration_ReviewDueFilter_ReadsSurfaced` — surface a doc → `review_due=true`
  includes it → `mark-reviewed` (advances `review_due_at`) → filter **excludes** it (auto-expel).

**Test strategy:** pin test (wire contract) + integration (worklist round-trip). Failing first
(filter currently recomputes, so a surfaced-but-not-recompute-due edge differs) → green.

## Task C — Execute the integration suite on real Postgres (main session, not a subagent)

PowerShell loader: parse `.env` → `$env:METALDOCS_DATABASE_URL` (never printed) → run TARGETED:
`go test -tags integration -run 'TestIntegration_Surfacer|TestIntegration_ReviewDueFilter|TestMarkReviewed_|TestDocumentReviewCheckConstraints|TestSubmit.*_RealDB|TestListDueForReview' ./internal/modules/documents/... ./internal/modules/jobs/document_review_surfacer/...`
Attach real output to `evidence.md`. Do NOT run the full 20-min suite.

## Ordering
A (surfacer conform + read-port + due-core const) → commit → B (filter/DTO/predicate) → commit →
C (integration run) → update evidence + milestone status → re-dispatch milestone-validator.

## Bounded defers (recorded)
- **ASYNC-TENANT-SEED lint port blind spot** — the lint can't see writes behind a cross-module port
  (documented "function/handler-local" limit). Real follow-up: extend the lint to flag jobs/worker
  call sites that invoke a known tenant-scoped write-port without a seed in scope. Trigger: a lint-
  hardening pass (M-tools / next tenancy sweep). Out of F6.4 boundary.
