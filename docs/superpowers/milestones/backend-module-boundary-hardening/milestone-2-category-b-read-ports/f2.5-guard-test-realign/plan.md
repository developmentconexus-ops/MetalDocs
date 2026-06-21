# Feature F2.5 — Plan — cilint guard-test realign

## Plan

1. **Realign the fixture** in `tools/cilint/internal/analyzers/hgcrossmodule_test.go` →
   `TestHGCrossModule_Negative_PendingBaseline`: replace the drained B1 fixture
   (`documents/repository/repository.go` reading `controlled_documents`) with the still-pending
   C1+C2 site (`controlleddocuments/infrastructure/repository.go` reading `user_process_areas`).
   Update the doc comment to record the realign + the M2/F2.1 drain that caused it.
2. **Re-green the suite:** `go test ./tools/cilint/...` → all `ok` (RED→GREEN on the baseline case).
3. **Confirm the binary is unaffected:** `go run ./tools/cilint ./...` → exit 0 (no production or
   ledger change, so it must stay 0).
4. **Confirm the fixture maps to a live ledger row:** the (file-suffix, table) pair
   `{controlleddocuments/infrastructure/repository.go, user_process_areas}` is present in
   `hgPendingRemediation` (C1+C2).
5. Evidence + commit (local only).

## Files touched

- `tools/cilint/internal/analyzers/hgcrossmodule_test.go` (test only) — 1 fixture.

## Test strategy

The change IS a test fix; "failing test first" is the validator-observed RED. Drive RED→GREEN by the
fixture realign; assertion semantics unchanged (still `len(findings) != 0 → fail`), so coverage of
the suppression path is preserved against a live ledger entry.

## Ordering

Edit → `go test ./tools/cilint/...` → `go run ./tools/cilint ./...` → evidence → commit.
