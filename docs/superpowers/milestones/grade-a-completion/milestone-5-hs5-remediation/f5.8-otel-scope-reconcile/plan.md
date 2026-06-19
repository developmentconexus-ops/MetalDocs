# Plan — F5.8 OTel scope reconcile

Documentation/reconciliation only. No source, test, or vendor file is changed.

## Tasks
1. **spec.md** — investigation + operator decision (done).
2. **milestone.md amendment:**
   - Amend the "Out of scope" / "Quality goals and constraints" observability-freeze line with a
     ratified, disclosed exception for the two bundled `*_otel_test.go` files (F2.3/ADR 0036 lineage).
   - Add a reclassification note: the semconv v1.24.0/v1.30.0 vendor files in `61389120` are a
     build-prerequisite repair (declared in modules.txt at parent, files missing), not new scope.
   - Add an **F5.8** row to the Features table and an **HS-4** record (validator FAIL → fix feature →
     re-dispatch).
3. **evidence.md** — capture the exact git commands + outputs proving the reclassification; re-run
   `go build ./...` + targeted tests to confirm nothing changed; note the re-dispatch outcome.
4. **Commit** f5.8 docs + milestone amendment.
5. **Re-dispatch milestone-validator**; on PASS, proceed to HS-1 operator gate.

## Test strategy
No code change → existing suite is the regression guard. The "test" of this feature is the
re-dispatched validator returning PASS with the scope finding resolved.
