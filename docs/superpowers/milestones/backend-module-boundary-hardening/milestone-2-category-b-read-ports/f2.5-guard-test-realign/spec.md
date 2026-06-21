# Feature F2.5 — Spec — cilint guard-test realign (HS-4 fix feature)

> **Milestone:** M2 (category-b-read-ports) · **Feature:** `f2.5-guard-test-realign`
> **Origin:** HS-4 — opened by the M2 milestone-validator FAIL verdict (`qa/milestone-qa.md`, C4).
> **Approved:** 2026-06-21 (operator standing authorization for validator-named fix features).

## Problem (validator finding C4)

M2/F2.1 correctly drained the **B1** entry (`documents/repository/repository.go` reading
`controlled_documents`) from `hgPendingRemediation` in
`tools/cilint/internal/analyzers/hgcrossmodule.go`. But the cilint **unit** suite's
`TestHGCrossModule_Negative_PendingBaseline` hard-codes the **B1 site** as its
"pending-baseline suppression" fixture. With B1 drained, the analyzer now correctly flags that
synthetic site → the test went from GREEN (pre-M2, `4ac99bed`) to RED. `go test ./tools/cilint/...`
was not re-run during M2, so the regression shipped undisclosed across F2.1–F2.4.

The guard **binary** is sound (`go run ./tools/cilint ./...` exit 0). Only the test's fixture is
stale: it asserts suppression of a site that is no longer on the ledger.

## Consumer contract

The consumer is the cilint unit suite itself. `TestHGCrossModule_Negative_PendingBaseline` must
prove that **a site currently on `hgPendingRemediation` is suppressed** — i.e. it must reference a
**live** ledger entry, not a drained one. The fixture must be a genuinely cross-module read (so the
only reason it is *not* flagged is the ledger), keeping the test a true suppression proof.

## Non-goals

- No production code change. No analyzer logic change. No ledger change.
- Do not re-point at another drained M2 site (B2–B8, N1) — all are drained.
- Do not weaken the assertion to `>= 0` or delete the test (that would drop suppression coverage).

## Decision

Repoint the fixture at the **C1+C2** still-pending site:
`controlleddocuments/infrastructure/repository.go` reading iam-owned `user_process_areas`
(scheduled for M3). It is on the ledger today, is genuinely cross-module
(controlleddocuments→iam), and so is suppressed only by `hgPendingRemediation` — exactly what the
test must prove. When M3 ports C1+C2, that milestone will realign this fixture again (the same
discipline this feature enforces).

## Validation Gate

- **Acceptance:** `go test ./tools/cilint/...` GREEN (all cases, incl. the realigned baseline);
  guard binary `go run ./tools/cilint ./...` still exit 0; the fixture references a table+file pair
  present in `hgPendingRemediation`.
- **Named test:** `TestHGCrossModule_Negative_PendingBaseline` (realigned).
- **Proof commands:** `go test ./tools/cilint/...`; `go run ./tools/cilint ./...`.
