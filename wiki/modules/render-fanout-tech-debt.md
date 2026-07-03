# Tech Debt Register - render-fanout

> Companion to `wiki/modules/render-fanout.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-07-02 (DOC-07a — T-001 closed: module page promoted from stub to full living-doc shape; APP-07 — T-002 closed: retry/terminal contract for the staging outbox repos documented + unit-tested; see `wiki/backend/flows/async-job-pipeline.md` §7)

## Items

### T-001 · Reconstruction/fanout behavior is spread across fanout and resolver packages without a consolidated module doc — CLOSED 2026-07-02 (DOC-07a)
- **Severity:** major (closed)
- **Surface:** `internal/modules/render/fanout/reconstruction.go:55`
- **Observation (original):** implementation is strong but documentation was still pipeline-stub level (no flow/error matrix).
- **Resolution:** `wiki/modules/render-fanout.md` now carries the full living-doc shape: Key files list (`render/domain/computed_catalog.go`, `fanout/client.go`, `fanout/pdf_dispatcher.go`, `fanout/pdf_dispatch_adapter.go`, `fanout/pdf_outbox_repository.go`, `fanout/pdf_outbox_worker.go`, `platform/worker/pdf_job_runner.go`, docx-renderer routes), a Computed-token catalog section (ADR 0050), a numbered Pipeline section (freeze → outbox enqueue → dispatch → Gotenberg → MinIO store), and a Failure modes table (5 rows: Gotenberg/LibreOffice down, docx-renderer substitution error, resolver empty value, nil tx, outbox replay, MinIO upload failure) with Symptom/Detection/Response columns — this is the flow/error matrix T-001 asked for.
- **Evidence:** `wiki/modules/render-fanout.md:1-103` (Key files §, Pipeline §68-74, Failure modes §76-85); page's own Last-verified stamp confirms currency (ADR 0050 + APP-01 outbox-only dispatch, 2026-06-29/2026-07-01).
- **Linked backlog row:** `R-001` (closed)
- **Linked ADR:** missing-ADR (doc-completeness item, not a design decision — no ADR needed)

### T-002 · Outbox retry/terminal semantics are not documented as an explicit contract — CLOSED 2026-07-02 (APP-07)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/render/fanout/staging_outbox.go` — the generic `StagingOutboxRepository` (the code that was `pdf_outbox_repository.go:80` at the time this item was opened has since been consolidated here; `NewPDFOutboxRepository`/`NewMaterializeOutboxRepository` bind the same repo to `metaldocs.pdf_dispatch_outbox` / `metaldocs.materialize_dispatch_outbox`). `ClaimPending` (`:73-102`), `MarkDispatched` (`:104-121`), `MarkFailed` (`:123-160`), `ResetStaleClaims` (`:197-212`), `CountDeadLettered` (`:165-175`).
- **Observation (original):** retry progression, finalize behavior, and stale-claim reset exist in code but are not captured in module-level contract docs.
- **Resolution:** `wiki/backend/flows/async-job-pipeline.md` §7 "Staging outbox tables — retry/terminal contract (APP-07)" now documents the full state machine (`pending → processing → dispatched | failed`), the claim→process→mark choreography (worker owns the `finalize` decision, repo is a dumb CAS layer), the backoff formula and its config source (`config.StagingOutboxWorkerConfig.MaxAttempts`/`StaleAfterSeconds`, env-overridable), `ResetStaleClaims` semantics (only `status='processing'` rows older than the stale threshold; who runs it — the same worker goroutine, every tick, before claiming), dead-letter visibility (`CountDeadLettered`), idempotency expectations (`Enqueue` ON CONFLICT DO NOTHING + `outbox_events` idempotency-key dedupe), and ADR 0054 tenancy rules (cross-tenant claim is sanctioned; compensating rules cited). The stale "Open flags" row claiming an unresolved `TODO(render)` tenant-predicate gap was also corrected — that TODO was already resolved by ADR 0054/commit `b4302dbf` (SEC-13) but the flags table had not been updated.
- **Test evidence:** `internal/modules/render/fanout/pdf_outbox_repository_test.go` — added `TestPDFOutboxRepository_MarkFailed_RetryPath_ResetsClaimAndBumpsAttempts`, `TestPDFOutboxRepository_MarkFailed_FinalizePath_SetsDeadLetteredAndFailedStatus`, `TestPDFOutboxRepository_MarkFailed_RowNotFound` (both branches), `TestPDFOutboxRepository_ResetStaleClaims_OnlyTouchesProcessingOlderThanCutoff`, `TestPDFOutboxRepository_ClaimPending_RespectsAttemptsAndRetryGate`, `TestPDFOutboxRepository_ClaimPending_NoEligibleRowsReturnsEmpty` (sqlmock, pins exact SQL predicates/columns per branch). Pre-existing tests in the same file (`TestPDFOutboxRepository_MarkFailed_AppliesBackoff`, `TestPDFOutboxRepository_ResetStaleClaims`) only asserted exec success, not the SQL shape — the new tests close that gap. `go build/vet/test -count=1 ./internal/modules/render/...` all green; `gofmt -l` clean on touched files.
- **Linked backlog row:** `R-002` (closed)
- **Linked ADR:** `wiki/decisions/0054-cross-tenant-outbox-claim.md` (tenancy sub-facet only; retry/finalize semantics remain convention-level, no dedicated ADR needed — documented as contract, not a decision requiring one)

### T-003 · Resolver registry compatibility/version policy lacks explicit ADR — CLOSED 2026-07-02 (ADR 0062)
- **Severity:** minor (closed)
- **Surface:** `internal/modules/render/resolvers/registry.go:13`
- **Observation (original):** resolver key/version behavior exists, but compatibility policy is convention-only.
- **Resolution:** ADR 0062 records the mechanism as-is: `Version()` is a per-resolver, self-reported integer stamped onto each `ResolvedValue` (`resolver.go:32`) for provenance/audit purposes only — it is never consulted to gate compatibility. `Registry.Register` allows overwrite-by-key (one live implementation per key); no multi-version coexistence exists or is currently needed. A real multi-version need would require a successor ADR, not a silent extension.
- **Evidence:** registry and resolver contract tests; `wiki/decisions/0062-render-resolver-registry-version-policy.md`.
- **Linked backlog row:** `R-003` (can be closed)
- **Linked ADR:** `wiki/decisions/0062-render-resolver-registry-version-policy.md`

## Coverage stats

- Public symbols undocumented: n/a (not fully audited)
- Operations missing C4 placement: n/a (module page now full living-doc shape, T-001 closed)
- Cross-deps missing in section map: n/a (module page now full living-doc shape, T-001 closed)
- State transitions missing: n/a (pipeline module) — staging outbox state machine now documented (T-002, closed)
- Decisions without ADR link: 1 (T-003 — T-001/T-002 closed, no ADR required for either)
