# Feature F4c.4 — Spec

> **Milestone:** 4c — Unified test-fixture framework  ·  **Folder:** `f4c.4-ci-grep-guards`
> **Status:** Approved (pre-code) — interview answers locked.
> **Approved before code:** 2026-06-15 — operator leandrotca.work ("Approved", post-recommendation).

> The feature's contract, written **before** any code. The milestone-validator judges F4c.4 against
> this file (C1). The "how" lives in `plan.md`; close-out proof lives in `evidence.md`.

## Interview record (fail-closed gate)

Five open items surfaced before drafting; operator delegated and approved the recommendations below.

| # | Question | Recommendation (approved) |
|---|----------|---------------------------|
| Q1 | Guard implementation language — PowerShell+bash parallel (matches `check-governance.{ps1,sh}`), single bash, or pure Go test? | **Single bash script** `scripts/check-test-discipline.sh`. CI is ubuntu-only — PS1 sibling is dead duplication. Go-test option tangles guard with `testdb` package + needs build tags for a 4-rule grep — overkill. Local devs invoke via git-bash on Windows. |
| Q2 | New `.github/workflows/test-discipline.yml` or extend `module-boundaries.yml`? | **Extend `module-boundaries.yml`** — same conformance class (architectural discipline grep-guards), same trigger (PR→main), same runner. Add a second step `Run test-discipline guard`. Independent fail-fast. |
| Q3 | Exception zones — permit `set_config('metaldocs.asserted_caps'` inside `tests/integration/testdb/**` only, or also where wrapped via `testdb.SeedWithCaps(...)` in test files? | **Hard-forbid the literal string in `*_test.go` outside `tests/integration/testdb/**`.** Tests should never type the literal — they call the Go helper. The literal only appears in `factory.go`. One-line allowlist, zero ambiguity. |
| Q4 | `pgtest` retirement — delete entirely, keep + document scope, or audit-first? | **Audit-first → delete if zero real callers.** Confirm `grep -rn "pgtest\." --include="*.go"` returns only `internal/testsupport/pgtest/pgtest.go` itself. If zero callers → delete the directory (CLAUDE.md §5.3 orphans rule). Any surviving caller → keep + add `// no-write only` header + scope the guard. |
| Q5 | Clean-baseline `go test` capture — local via operator DSN or CI dry-run? | **Local capture, commit log.** CLAUDE.md §1: scripts are canonical truth. CI dry-run masks the operator-DSN distinction the milestone is measuring. `go test -tags integration -count=1 ./...` against operator DSN, tee to evidence. Matches F4c.2/F4c.3 evidence pattern. |

## Consumer contract (FIRST — before any producer)

The **producer** is `scripts/check-test-discipline.sh` + the new workflow step in
`.github/workflows/module-boundaries.yml`. The **consumers** are:

- **CI** (`module-boundaries` workflow, PR→main) — invokes `bash scripts/check-test-discipline.sh`
  as a second step. Must exit **non-zero on any rule violation** (each violation prints `path:line:
  <rule-id>: <offending text>`); exit zero on clean tree. CI fails the PR check on non-zero.
- **Local developer** — invokes `bash scripts/check-test-discipline.sh` from repo root. Same exit
  semantics. Must be runnable from PowerShell via git-bash on Windows (no PS-specific deps).
- **The migrated tree (F4c.3 output)** — the guard must report exit 0 on the current HEAD. Any
  non-zero on HEAD means F4c.4 itself is broken or F4c.3 closed dirty. Empty / sanctioned
  `pgtest` directory must not trip the guard.
- **A future migrating developer** — adding a violating line in a new `*_test.go` file must trip
  the guard locally before CI catches it. The error message must name the rule + the offending
  text, no diagnostic spelunking required.

**Source of truth for the rule set:** this `spec.md` (Validation Gate, below). The shipped script
must implement *exactly* these rules — no more, no less.

## What this feature implements

### A. Discipline rules (4)

Grep-based, applied to files matching `*_test.go` with first line `//go:build integration` (the
F4c.3 HS-6 amendment's scope filter — sqlmock unit tests are out of scope at runtime). Exception
zone: `tests/integration/testdb/**` (the framework's own helpers).

| Rule id | Rule | Regex (POSIX ERE) | Why forbidden |
|---------|------|-------------------|---------------|
| R1 | No inline `metaldocs.asserted_caps` set_config in test files | `set_config\('metaldocs\.asserted_caps'` | Sanctioned path is `testdb.SeedWithCaps(...)`; the literal only lives in `factory.go` (Q3). |
| R2 | No `is_local=false` on a `set_config` | `set_config\([^)]*,\s*false\s*\)` | F4c.1 invariant: tripwire writes must be tx-local (`is_local=true`). `false` leaks across the pool. |
| R3 | No hardcoded tenant-UUID literal in test files | `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}` matching a known dev tenant constant list (`testdb.DevTenantID` literal + any other operator-supplied constant). Refined regex: bare UUIDv4 *outside* a comment/string-only context. To keep this guard small + low-false-positive, the implemented form is: forbid the literal `testdb.DevTenantID`'s string value when it appears as a bare quoted string in `*_test.go` files outside `testdb/`. | Tests must obtain tenant identity via `factory.NewTenant(...)` (random UUID per test) or `testdb.DevTenant` (one named const). Hardcoded literals are the F4c.3 census's primary anti-pattern. |
| R4 | No bare unqualified `documents` table reference in test SQL | `(FROM\|JOIN\|INTO\|UPDATE)\s+documents([^_a-zA-Z]\|$)` | M4b legacy-schema teardown: bare `documents` resolves to dead legacy table under wrong `search_path`. Tests must use `testdb.Qualified("documents")` or factory-returned identifiers. |

> **Allow-list:** files under `tests/integration/testdb/**` are exempt from R1, R2, R3, R4 (they
> are the framework's owners of those primitives). Files matching `*_test.go` without the
> `//go:build integration` first line are exempt entirely (sqlmock-mock-string-as-payload class —
> F4c.3 HS-6 filter).

### B. `pgtest` retirement

Audit phase:
1. `grep -rn "pgtest\." --include="*.go"` — list every reference.
2. `grep -rn "testsupport/pgtest" --include="*.go"` — list every import.

Resolution branches:
- **If zero callers** (only `internal/testsupport/pgtest/pgtest.go` itself appears): **delete the
  entire `internal/testsupport/pgtest/` directory.** Update any orphan references. Record the
  before/after grep in evidence.
- **If callers exist**: classify each as **no-write** (read-only against the test DB — keep,
  document) or **stateful-write** (must migrate). Migration is out-of-scope for F4c.4 (would be a
  return to F4c.3); surface as HS-6 scope drift if any stateful-write caller is found post-F4c.3
  close. Add a `// no-write only` header comment to `pgtest.go` documenting the scope decision and
  the audited caller list.

### C. New workflow step

Append to `.github/workflows/module-boundaries.yml`:

```yaml
      - name: Run test-discipline guard
        shell: bash
        run: bash scripts/check-test-discipline.sh
```

After the existing `Run module boundaries conformance` step. No reorder of existing steps. No
change to triggers or runner.

### D. Discipline rules document

A short reference under `wiki/quality/test-discipline.md` (or extend an existing wiki page —
operator-decided in plan) describing each rule, the allow-list, the sanctioned pattern
(`testdb.SeedWithCaps`, `testdb.Qualified`, `factory.NewTenant`), and the rationale per rule.
Linked from `wiki/README.md`. This document is the **why**; the script is the **what**.

> **Scope note:** the full *framework* doc + ADR is F4c.5. F4c.4 ships only the *rules reference*
> needed to make the guard self-explanatory to a developer hitting a CI failure.

## Non-goals (mandatory)

- **No factory API change.** F4c.1 owns the API.
- **No edit to `tests/integration/testdb/db.go` or `factory.go`.** F4c.4 is purely additive (script
  + workflow step + optional doc page + `pgtest/` deletion if Q4 → zero callers).
- **No migration of any test file.** F4c.3 closed that scope. Any test still violating the rules
  after F4c.3 close is an F4c.3 close defect → HS-4 fix-feature, not F4c.4 in-line work.
- **No production-source change** (`internal/...` excluding `internal/testsupport/pgtest/`
  retirement under Q4, `db/migrations/...`, `db/baseline/...`).
- **No PS1 sibling** to `check-test-discipline.sh`. Bash only (Q1).
- **No new top-level workflow file.** Extend `module-boundaries.yml` (Q2).
- **No tripwire weakening.** Guard reads the tree; never writes.
- **No framework ADR / wiki rebuild.** That is F4c.5 — F4c.4 ships only the per-rule reference doc.

## Validation Gate (concrete — approved before code)

| AC | Acceptance criterion | Named test / proof command | Real vs fixture |
|----|----------------------|----------------------------|-----------------|
| **AC1** | Guard passes on HEAD clean tree. | `bash scripts/check-test-discipline.sh; echo $?` → `0` (captured in evidence). | real |
| **AC2** | Guard fails (exit non-zero) on a planted violation **of each rule R1–R4**, then passes again after revert. | Per-rule: plant the violation in a throwaway `_test.go` file (under e.g. `internal/scratch/`), run guard → non-zero + named-rule output; revert; rerun → zero. Captured per-rule in evidence. | real |
| **AC3** | The script's rule set is **exactly** R1–R4 (no more, no less); allow-list is **exactly** `tests/integration/testdb/**` + non-`//go:build integration` files. | `diff` of `scripts/check-test-discipline.sh` against this spec's rule table, or a self-documenting `--rules` flag whose output is asserted. | real |
| **AC4** | CI workflow runs the new step on PR; failure red-fails the check. | `cat .github/workflows/module-boundaries.yml` shows the added step; PR-class smoke (push planted violation to a throwaway branch, observe red, revert) — or, if no throwaway branch is acceptable, dry-run via `act` locally + log. Captured in evidence. | real |
| **AC5** | `pgtest` audited + resolved per Q4: either `internal/testsupport/pgtest/` deleted (zero callers branch) or scoped-with-header-comment (callers-exist branch). Resolution recorded. | `grep -rn "pgtest" --include="*.go"` before/after; if delete branch → `git diff --stat internal/testsupport/pgtest/`. | real |
| **AC6** | Full integration suite green from a clean baseline under the operator DSN (M4b post-teardown debt in `tests/integration/scenarios/` continues to be a bounded out-of-scope defer carried from F4c.3 — same baseline, not a new failure). | `go test -tags integration -count=1 ./...` against operator DSN; tee to `evidence/clean-baseline.log`; compare red-set against F4c.3 close baseline (zero new failures). | real |
| **AC7** | Regression: F4c.3 + F4c.2 + F4.1a + M4 gates still pass under the unchanged operator DSN. | `go test -tags integration -count=1 -run 'TestCreateDocumentTx_PopulatesAllSnapshotColumns\|TestScheduledPublishWorker_\|TestValidateScheduledSupersedeTarget_RealRows\|TestLoadCurrentPublishedHeadForDocument_RealRows\|TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion\|TestLoadInstance_LoadsDocumentRevisionVersion\|TestScheduleGenerationIncrementsOnScheduledWritePath' ./...` GREEN. | real |
| **AC8** | No production-source change beyond Q4 `pgtest/` retirement (if zero callers). | `git diff --name-only origin/main...HEAD -- internal/ db/` returns at most `internal/testsupport/pgtest/...`. | real |
| **AC9** | Discipline reference doc exists (`wiki/quality/test-discipline.md` or named-equivalent decided in plan) + linked from `wiki/README.md` + names each rule's allow-list + sanctioned pattern. | `wc -l wiki/quality/test-discipline.md` > 30; `grep -n test-discipline wiki/README.md` → match. | real |

> **TDD discipline:** AC2 is the RED test — plant each rule's violation first, capture guard
> non-zero exit + named-rule output, **then** revert. The script implementation is GREEN only when
> AC1 + AC2 both hold. Per-rule planted-violation evidence is mandatory (no "trust me, the regex
> matches" — show the matched line).

## ADR needed?

- [x] **No** — F4c.4 is enforcement of F4c.1's framework decisions, not a new durable architecture
  decision. The framework ADR is F4c.5 (next feature). The `pgtest` retirement (if Q4 → zero
  callers) is a mechanical deletion of dead code, not a decision; the decision was already made in
  F4c.1 ("testdb-factory is the unified harness").

## Inherited constraints (from CLAUDE.md + M4c milestone.md)

- **§5.3 surgical changes:** touch only `scripts/check-test-discipline.sh` (new),
  `.github/workflows/module-boundaries.yml` (one step appended), `internal/testsupport/pgtest/`
  (delete iff zero callers), `wiki/quality/test-discipline.md` (new), `wiki/README.md` (one index
  line). No other files.
- **HS-2:** if implementing the guard reveals that a guard rule cannot be expressed without a
  factory-API change or production-source change → **stop**, report boundary, do not symptom-patch.
- **HS-6:** if the audit (Q4) finds a stateful `pgtest` caller F4c.3 missed → stop, surface to
  operator, do not silently migrate (that would be F4c.3 close-defect class → HS-4).
- **CI failure messaging contract:** the script must print `path:line: <rule-id>: <quoted offending
  text>` per violation. No silent failure. No diagnostic spelunking required from a developer
  hitting a red CI check.
