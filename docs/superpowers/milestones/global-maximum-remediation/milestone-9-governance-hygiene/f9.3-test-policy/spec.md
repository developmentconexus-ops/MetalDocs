# F9.3 — test-policy (feature spec)

> **Milestone:** M9 governance-hygiene · **Contract:** `../validation-contract.md` §3 (binding)
> **Approved:** 2026-07-06 — approved against mission.md M9 row + validation-contract §3 (operator-locked
> sources; autonomous session per mission D2). Code may start.

## Consumer contract (first)

**Consumer 1 — maintainer with a broken legacy test:** opens one page,
`wiki/quality/legacy-test-policy.md`, and gets a mechanical decision: *does this test guard a REQ ID,
a tripwire arm, a wire-contract shape, or a DB invariant?* → repair on the canonical framework;
*else (one-off task scaffolding)* → delete with a one-line commit rationale. ≥2 worked examples from
this repo. Page linked from `wiki/quality/index.md` and `wiki/quality/test-discipline.md`.

**Consumer 2 — CI wall-clock budget:** integration packages that can safely parallelize do; the gain
is measured (same command, same box, before/after) on a **named representative package set**, honest
numbers recorded even if ~0.

**Consumer 3 — test-framework hard gate (existing):** the policy cites and does not weaken ADR 0034 /
the testdb-factory gate; repair-class work lands on the canonical framework.

## Interview record (B1.5 — resolved from normative sources)

| Q | A | Source |
|---|---|--------|
| Policy home + name? | New `wiki/quality/legacy-test-policy.md`; index + test-discipline links. | Contract §3.1 (doc under wiki/quality, linked both places) |
| Taxonomy? | Repair-class = guards REQ-ID / tripwire arm / contract shape (OpenAPI-generated DTO or route) / DB invariant (trigger/constraint/RLS). Delete-class = one-off scaffolding (drive-a-fix tests, superseded harnesses). Procedure: classify → repair on canonical framework OR delete citing the class. | mission M9 row; memory rule being codified |
| Baseline for t.Parallel? | 12 of 386 `_test.go` files under `internal/` call `t.Parallel()` (2026-07-06). Expansion targets package-level: subtests + top-level tests in integration files using the testdb factory (template-DB-per-test = isolated by construction, ADR 0034). | Sweep + ADR 0034 |
| Representative set for measurement? | Chosen by implementer from the heaviest DB-integration packages (must name ≥3 packages, e.g. documents/repository-class, templates, approval suites), fixed BEFORE expansion, measured with `go test <pkgs> -count=1` timing before/after on this box. Full suite explicitly excluded (mission §10). | Contract §0.5, §3.2 |
| Parallel-unsafe files? | Left serial, one-line reason recorded in plan/evidence table (shared fixture mutation, global state, t.Setenv, port binding, ordered assertions). No reason-comments sprayed into code. | Contract §3.2 |
| Known trap? | Never assert goroutine lifecycle via NumGoroutine in parallel suites (flaky-ratelimit lesson, TST-04 695bd8e0) — policy records this as an anti-pattern example. | Memory `flaky-ratelimit-sweeper-test` |

## Non-goals (mandatory)

- No mass migration of legacy tests to the canonical framework (policy governs future triage).
- No test deletions in this feature (F9.3 writes the rulebook; deletions happen when tests break).
- No CI workflow changes; no -p/GOMAXPROCS tuning; no testdb pool-size changes.
- No semantic changes to any test (parallelization must not alter assertions or fixtures).
- No flake-hunting beyond the touched set staying green.

## Validation Gate

1. Policy doc exists with taxonomy + procedure + ≥2 repo-real worked examples; linked from
   `wiki/quality/index.md` AND `wiki/quality/test-discipline.md`; cites ADR 0034 + hard gate.
2. Before/after wall-clock outputs for the named package set captured (same command, `-count=1`,
   both runs in evidence; label: real DB vs fixture per suite).
3. `go test <named set> -count=1` green after expansion.
4. Files-left-serial table with reasons in evidence.
5. Diff scope: only `*_test.go` files (t.Parallel additions), wiki/quality docs, feature folder.
