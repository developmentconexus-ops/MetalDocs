# Tech Debt Register - render-fanout

> Companion to `wiki/modules/render-fanout.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-07-02 (APP-07 — T-002 closed: retry/terminal contract for the staging outbox repos documented + unit-tested; see `wiki/backend/flows/async-job-pipeline.md` §7)

## Items

### T-001 · Reconstruction/fanout behavior is spread across fanout and resolver packages without a consolidated module doc
- **Severity:** major
- **Surface:** `internal/modules/render/fanout/reconstruction.go:55`
- **Observation:** implementation is strong but documentation is still pipeline-stub level.
- **Evidence:** module page is currently a high-level stub with no flow/error matrix.
- **Linked backlog row:** `R-001`
- **Linked ADR:** missing-ADR

### T-002 · Outbox retry/terminal semantics are not documented as an explicit contract — CLOSED 2026-07-02 (APP-07)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/render/fanout/staging_outbox.go` — the generic `StagingOutboxRepository` (the code that was `pdf_outbox_repository.go:80` at the time this item was opened has since been consolidated here; `NewPDFOutboxRepository`/`NewMaterializeOutboxRepository` bind the same repo to `metaldocs.pdf_dispatch_outbox` / `metaldocs.materialize_dispatch_outbox`). `ClaimPending` (`:73-102`), `MarkDispatched` (`:104-121`), `MarkFailed` (`:123-160`), `ResetStaleClaims` (`:197-212`), `CountDeadLettered` (`:165-175`).
- **Observation (original):** retry progression, finalize behavior, and stale-claim reset exist in code but are not captured in module-level contract docs.
- **Resolution:** `wiki/backend/flows/async-job-pipeline.md` §7 "Staging outbox tables — retry/terminal contract (APP-07)" now documents the full state machine (`pending → processing → dispatched | failed`), the claim→process→mark choreography (worker owns the `finalize` decision, repo is a dumb CAS layer), the backoff formula and its config source (`config.StagingOutboxWorkerConfig.MaxAttempts`/`StaleAfterSeconds`, env-overridable), `ResetStaleClaims` semantics (only `status='processing'` rows older than the stale threshold; who runs it — the same worker goroutine, every tick, before claiming), dead-letter visibility (`CountDeadLettered`), idempotency expectations (`Enqueue` ON CONFLICT DO NOTHING + `outbox_events` idempotency-key dedupe), and ADR 0054 tenancy rules (cross-tenant claim is sanctioned; compensating rules cited). The stale "Open flags" row claiming an unresolved `TODO(render)` tenant-predicate gap was also corrected — that TODO was already resolved by ADR 0054/commit `b4302dbf` (SEC-13) but the flags table had not been updated.
- **Test evidence:** `internal/modules/render/fanout/pdf_outbox_repository_test.go` — added `TestPDFOutboxRepository_MarkFailed_RetryPath_ResetsClaimAndBumpsAttempts`, `TestPDFOutboxRepository_MarkFailed_FinalizePath_SetsDeadLetteredAndFailedStatus`, `TestPDFOutboxRepository_MarkFailed_RowNotFound` (both branches), `TestPDFOutboxRepository_ResetStaleClaims_OnlyTouchesProcessingOlderThanCutoff`, `TestPDFOutboxRepository_ClaimPending_RespectsAttemptsAndRetryGate`, `TestPDFOutboxRepository_ClaimPending_NoEligibleRowsReturnsEmpty` (sqlmock, pins exact SQL predicates/columns per branch). Pre-existing tests in the same file (`TestPDFOutboxRepository_MarkFailed_AppliesBackoff`, `TestPDFOutboxRepository_ResetStaleClaims`) only asserted exec success, not the SQL shape — the new tests close that gap. `go build/vet/test -count=1 ./internal/modules/render/...` all green; `gofmt -l` clean on touched files.
- **Linked backlog row:** `R-002` (closed)
- **Linked ADR:** `wiki/decisions/0054-cross-tenant-outbox-claim.md` (tenancy sub-facet only; retry/finalize semantics remain convention-level, no dedicated ADR needed — documented as contract, not a decision requiring one)

### T-003 · Resolver registry compatibility/version policy lacks explicit ADR
- **Severity:** minor
- **Surface:** `internal/modules/render/resolvers/registry.go:13`
- **Observation:** resolver key/version behavior exists, but compatibility policy is convention-only.
- **Evidence:** registry and resolver contract tests.
- **Linked backlog row:** `R-003`
- **Linked ADR:** missing-ADR

## Coverage stats

- Public symbols undocumented: n/a (not fully audited)
- Operations missing C4 placement: n/a (stub-level doc)
- Cross-deps missing in section map: n/a (stub-level doc)
- State transitions missing: n/a (pipeline module) — staging outbox state machine now documented (T-002, closed)
- Decisions without ADR link: 2 (T-001, T-003 — T-002 closed, tenancy sub-facet covered by ADR 0054)
