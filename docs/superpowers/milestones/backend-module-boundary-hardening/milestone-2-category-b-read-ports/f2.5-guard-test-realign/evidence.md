# Feature F2.5 — Evidence — cilint guard-test realign (HS-4 fix)

> **Milestone:** M2 (category-b-read-ports) · **Feature:** `f2.5-guard-test-realign` · **Closed:** 2026-06-21
> **Origin:** HS-4 — M2 milestone-validator FAIL (C4). **Contract:** `spec.md`.

## What was implemented

- **Root cause (validator C4):** `TestHGCrossModule_Negative_PendingBaseline` asserted suppression
  of the **B1** site, which F2.1 ported and drained from `hgPendingRemediation`. With B1 gone the
  analyzer correctly flags the synthetic fixture → the test went RED. `go test ./tools/cilint/...`
  was not re-run during M2, so the break shipped undisclosed.
- **Fix (test only, root-cause not symptom):** realigned the fixture in
  `tools/cilint/internal/analyzers/hgcrossmodule_test.go` to a **still-pending** ledger entry —
  the C1+C2 site `controlleddocuments/infrastructure/repository.go` reading iam-owned
  `user_process_areas` (scheduled for M3). The assertion semantics are unchanged
  (`len(findings) != 0 → fail`), so the test still proves the suppression path — now against a live
  `hgPendingRemediation` row rather than a drained one. The doc comment records the realign and the
  F2.1 drain that necessitated it.
- **No production / analyzer / ledger change.** The guard binary was already exit 0; this feature
  does not touch it.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| RED reproduced (validator finding) | `go test ./tools/cilint/...` (pre-fix) | `--- FAIL: TestHGCrossModule_Negative_PendingBaseline` — `pending-remediation baseline site must be suppressed, got 1: [...controlled_documents...]` | real |
| GREEN after realign | `go test ./tools/cilint/...` | `ok metaldocs/tools/cilint/internal/analyzers` (all cases) | real |
| Guard binary unaffected | `go run ./tools/cilint ./...` | `CILINTEXIT=0` | real |
| Fixture maps to a LIVE ledger row | review: `{controlleddocuments/infrastructure/repository.go, user_process_areas}` present in `hgPendingRemediation` (C1+C2) | confirmed | real |
| Suppression is genuine (not compliant for another reason) | the fixture is a cross-module read (controlleddocuments→iam `user_process_areas`); removing the C1+C2 ledger row would flag it — only the ledger suppresses it | confirmed by ledger presence + analyzer owner-map | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `go test ./tools/cilint/...` GREEN (incl. realigned baseline) | yes | suite `ok` |
| Guard binary still exit 0 | yes | `CILINTEXIT=0` |
| Fixture references a table+file pair on `hgPendingRemediation` | yes | C1+C2 row present |
| No production/analyzer/ledger change | yes | only `hgcrossmodule_test.go` touched |

## Review disposition

- **Spec-compliance review:** PASS. Test-only realign to a live ledger entry; assertion preserved;
  no weakening, no deletion, no production change. Matches the validator-named fix exactly.
- **Code-quality review:** PASS. The new fixture is a true cross-module read suppressed only by the
  ledger, keeping the test a real suppression proof. The forward-discipline (M3 will realign again
  when it drains C1+C2) is documented in the comment.

## Bounded defers

None.
