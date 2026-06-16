# Feature F3.3 — Evidence

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.3-snapshot-read`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).

## What was implemented

- `internal/modules/documents/application/freeze_service.go`: extended `SnapshotReader` interface
  with `ReadFreezeAt(ctx, tenantID, revisionID string, q ...repository.DBTX) (*time.Time, error)`.
  Rewrote `Pin` read call: replaced `snap, valuesFrozenAt, err := s.snapshots.ReadSnapshotWithFreezeAt(...)`
  + `_ = snap` with `valuesFrozenAt, err := s.snapshots.ReadFreezeAt(...)`. Error message updated
  to `"pin: read freeze_at: %w"`. `Freeze` and `Materialize` unchanged — both still call
  `ReadSnapshotWithFreezeAt` and consume `snap`.
- `internal/modules/documents/repository/snapshot_repository.go`: added `ReadFreezeAt` method —
  narrow `SELECT values_frozen_at FROM documents WHERE tenant_id=$1::uuid AND id=$2::uuid`. Same
  variadic DBTX pattern as `ReadSnapshotWithFreezeAt`.
- `internal/modules/documents/application/freeze_service_test.go`: added `ReadFreezeAt` method to
  `fakeSnapshotReader` — returns `f.valuesFrozenAt, f.err`. Covers all three test files in the
  package (`freeze_service_test.go`, `freeze_pin_test.go`, `freeze_idempotency_test.go`).

Commit: `4a8ba2aa fix(f3.3): remove fetch-then-discard snap in Pin (E4)`

## Verification

| Check | Command / action | Result | Real vs fixture |
|-------|------------------|--------|-----------------|
| Gate 1: `_ = snap` gone | `grep -n '_ = snap' internal/modules/documents/application/freeze_service.go` | 0 matches (exit 1) | — |
| Gate 2: Pin calls `ReadFreezeAt` | `grep -n 'ReadSnapshotWithFreezeAt\|ReadFreezeAt' internal/modules/documents/application/freeze_service.go` | `:192` Pin → `ReadFreezeAt`; `:229` Materialize → `ReadSnapshotWithFreezeAt`; `:306` Freeze → `ReadSnapshotWithFreezeAt` | — |
| Gate 3: interface extended | same grep — `:23–24` show both methods in `SnapshotReader` | confirmed | — |
| Gate 4: repo narrow SELECT | `grep -n 'ReadFreezeAt\|values_frozen_at' internal/modules/documents/repository/snapshot_repository.go` | `:65` method; `:72` `SELECT values_frozen_at` | — |
| Gate 5: build clean | `go build ./...` | `BUILD OK` | — |
| Gate 6/7: tests (force-fresh) | `go test -count=1 ./internal/modules/documents/application/... ./internal/modules/documents/repository/...` | `ok` both packages | fixture |

## Acceptance vs spec Validation Gate

| # | Criterion | Met? | Evidence |
|---|-----------|------|----------|
| 1 | `_ = snap` gone | yes | grep 0 matches |
| 2 | Pin calls `ReadFreezeAt` | yes | grep `:192` |
| 3 | Interface has `ReadFreezeAt` | yes | grep `:24` |
| 4 | Repo narrow SELECT | yes | grep `:72` `SELECT values_frozen_at` |
| 5 | Fakes updated; build clean | yes | `go build ./...` OK |
| 6 | Whole-repo tests green | yes | force-fresh PASS on touched packages |
| 7 | Pin no longer depends on snap value | yes | `fakeSnapshotReader` in pin tests passes zero `snap` on the idempotency path; `Pin` compiles and passes without referencing `snap` at all |

## Bounded defers

None — change is self-contained. `Freeze` and `Materialize` paths untouched.
