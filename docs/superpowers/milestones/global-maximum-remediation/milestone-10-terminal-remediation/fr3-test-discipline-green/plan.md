# F-R3 Plan

## Files touched
1. `scripts/check-test-discipline.sh` — R2 allowlist path correction (repository/→infrastructure/) + reconciliation comment.
2. `internal/modules/jobs/stuck_instance_watchdog/job_integration_test.go` — qualify `documents`→`metaldocs.documents` (R4).
3. `internal/modules/controlleddocuments/domain/sequence_test.go` — import testdb; replace two inline set_config sites with `testdb.SetCapsOnTx` (R1).
4. Defer ledger (this feature's `evidence.md`) — REQ-SEARCH-1 + REQ-SEC-3 ratified.

## Order
1. R2 path correction (one line + comment).
2. R4 qualify (one token).
3. R1 migrate both sites to `testdb.SetCapsOnTx`; reword any comment that would itself match the R1 grep.
4. Re-run `check-test-discipline.sh` → clean.
5. `go vet -tags integration` on the 3 packages → compiles (proves no testdb import cycle from the domain_test external package).
6. `req-trace` → uncovered set unchanged.
7. Write defer ledger.

## Test strategy
- The gate script IS the test — clean exit 0 is the acceptance.
- `go vet -tags integration` proves the edits compile (esp. the new testdb import from `domain_test`).
- Behavior preservation: `SetCapsOnTx` asserts the identical caps tx-locally (is_local=true), same as
  the inline set_config it replaces.

## Risk / rollback
- Low. One allowlist path, one SQL qualifier, two helper substitutions.
- Worker-goroutine note: `SetCapsOnTx` uses `t.Fatalf`; on the (effectively unreachable) set_config
  failure path this Goexits the worker instead of the prior rollback+Errorf. The test is still marked
  failed (t.Fail is goroutine-safe) and `wg.Done()` is deferred, so `wg.Wait()` still returns — assertion
  strength preserved. Documented in evidence.
- Rollback = revert the 3 files.
