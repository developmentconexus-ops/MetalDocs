# F-R3 Evidence — Test-discipline green + defer ratification (Dim 8 → CONFIRMED)

Closes the Dim-8 DEBT: `check-test-discipline.sh` is now GREEN at HEAD, all 4 violations resolved at
root; the two absent-feature MUSTs are ratified as bounded defers.

## Changes

| File | Change | Rule |
|------|--------|------|
| `scripts/check-test-discipline.sh` | R2 allowlist path `templates/repository/…` → `templates/infrastructure/…` (F9.5-rename reconciliation) + comment. | R2 |
| `internal/modules/jobs/stuck_instance_watchdog/job_integration_test.go:186` | `FROM documents` → `FROM metaldocs.documents`. | R4 |
| `internal/modules/controlleddocuments/domain/sequence_test.go` | Import `tests/integration/testdb`; two inline `asserted_caps` set_config sites → `testdb.SetCapsOnTx(t, tx, …)`. | R1 |

## Before → after (root-cause, not suppression)

- **R2 (mission-introduced):** the entry was already allowlisted as a legitimate single-pinned-conn
  RLS probe; M9 F9.5 renamed the directory and left the path stale. This is an allowlist **path
  correction**, not a widening — the allowlist set did not grow.
- **R4:** bare `documents` → schema-qualified `metaldocs.documents` (the M4b legacy-schema hazard the
  rule guards).
- **R1:** the two inline `SELECT set_config('metaldocs.asserted_caps', …, true)` writes are replaced
  by the sanctioned `testdb.SetCapsOnTx(t, tx, …)`, which asserts the identical caps transaction-
  locally. Same tripwire assertion, sanctioned primitive.

## Gate output (GREEN)

```
$ bash scripts/check-test-discipline.sh
test-discipline: clean (124 integration test files checked)
GATE_EXIT=0
```

First re-run after the R1 migration flagged one more R1 hit — my *explanatory comment* contained the
literal `set_config('metaldocs.asserted_caps'`, which the grep matches. Reworded the comment (the
grep guards intent, and a comment quoting the banned literal is correctly caught). Clean thereafter.

## Compile proof (no testdb import cycle)

```
$ go vet -tags integration ./internal/modules/controlleddocuments/domain/ \
                           ./internal/modules/jobs/stuck_instance_watchdog/ \
                           ./internal/modules/templates/infrastructure/
VET_EXIT=0
```

The `domain_test` external test package importing `testdb` (which imports `controlleddocuments/domain`)
does **not** cycle — external `_test` packages exist precisely to allow this.

## No-regression proof

```
$ go build ./...                                             → BUILD_EXIT=0
$ go run ./scripts/req-trace                                 → 4 MUST uncovered; stale=false; exit 1
    UNCOVERED: REQ-AUTHN-1, REQ-AUTHN-3, REQ-SEARCH-1, REQ-SEC-3   (unchanged)
$ go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .  → 0 violations
$ scripts/check-module-boundaries.ps1                        → [module-boundaries] OK
```

## Worker-goroutine behavior note (repair faithfulness)

`SetCapsOnTx` uses `t.Fatalf`. At the per-worker site (a goroutine), on the effectively-unreachable
set_config-failure path this Goexits the worker rather than the prior `rollback + t.Errorf + return`.
The test is still marked failed (`t.Fail` is goroutine-safe) and `wg.Done()` is deferred, so
`wg.Wait()` still returns — assertion strength is preserved. The happy path (the path this test
actually exercises) is byte-for-byte equivalent.

## Defer ledger — ratified bounded defers (§8 uncovered MUSTs)

These two are **absent product features**, not hygiene; ratified as bounded defers with named
triggers (they stay in the `req-trace` uncovered set by design). REQ-AUTHN-1 / REQ-AUTHN-3 are the
separate doc-vs-runtime drift defers already ratified in M9 (erratum E1) and are **not** re-opened here.

| REQ | Finding (why uncovered) | Trigger to close | Owner |
|-----|------------------------|------------------|-------|
| **REQ-SEARCH-1** | No operationalized search **reindex procedure** exists — there is no runbook/tool to rebuild the search index from source-of-truth, and no test asserting one. Search today is functional but has no disaster-recovery/reindex path. | First of: a change to searchable-field schema; a search-index corruption/loss incident; adoption of a new search backend; or a pre-production DR sign-off requiring reindex. | `search` module |
| **REQ-SEC-3** | OWASP **ASVS** was never operationalized — the target spec names an ASVS expectation but it was never mapped to a tracked checklist or verification, so no coverage exists. | First of: pre-production security sign-off; an external penetration-test engagement; or a compliance audit requiring an ASVS-level mapping. | `security` module |

Both triggers are external events, not silent time-based decay — consistent with the mission's
"bounded defer carries a named trigger" rule.
