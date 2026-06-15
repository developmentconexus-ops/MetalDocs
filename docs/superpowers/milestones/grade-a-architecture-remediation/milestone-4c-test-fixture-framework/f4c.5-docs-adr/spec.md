# Feature F4c.5 — Spec

> **Milestone:** 4c — Unified test-fixture framework  ·  **Folder:** `f4c.5-docs-adr`
> **Status:** Approved (pre-code) — interview locked, no open ambiguities.
> **Approved before code:** 2026-06-15 — operator leandrotca.work ("New page (b)").

> The feature's contract, written **before** any code. Documentation-only — no production code,
> no schema change, no test file change. The milestone-validator judges F4c.5 against this file.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| Q1 | Wiki page placement: expand `test-discipline.md` in place (a) or new `integration-test-harness.md` (b)? | **New page (b)** — `wiki/quality/integration-test-harness.md`. `test-discipline.md` stays as CI-rules-only reference (developer hitting red CI consults it). New page = architectural explanation + factory API + how-to (developer writing a new test from scratch). Different readers, different contexts. |
| Q2 | ADR number? | **ADR 0034** (next after 0033 `fix-placeholder-value-tenant-trigger`). |

## Consumer contract (FIRST — before any producer)

Consumers of F4c.5's deliverables:

- **A developer writing a new integration test** — opens `wiki/quality/integration-test-harness.md`
  and gets: (1) why template-DB-per-test was chosen, (2) how to open a DB in a test, (3) which
  factory builders exist and what they seed, (4) how to handle guarded writes, (5) a minimal
  working example. No need to read `factory.go` source to write a correct test.
- **A developer hitting a CI red check** — opens `wiki/quality/test-discipline.md` (F4c.4, not
  modified here) and gets the per-rule explanation. `integration-test-harness.md` is cross-linked
  for context but is not the remediation reference.
- **The milestone-validator** — judges that both `wiki/quality/integration-test-harness.md` and
  `wiki/decisions/0034-integration-test-fixture-framework.md` exist, are linked from indexes,
  accurately describe the shipped API (no drift from `factory.go`), and that `wiki-curator` was
  dispatched.
- **Future ADR readers** (`wiki/decisions/index.md`) — ADR 0034 records the durable architecture
  decision: why IntegreSQL template-DB-per-test was chosen over shared-DB, what the alternatives
  were, what the consequences are, and which implementation artifacts it governs.

## What this feature implements

### A. `wiki/quality/integration-test-harness.md` (new)

Content sections (minimum, not a scratchpad — must be accurate at close):

1. **Harness choice** — IntegreSQL template-DB-per-test pattern: one template DB cloned per test
   run, curated baseline applied once, every test gets its own isolated clone. Why: prevents
   cross-test state leakage, enables parallel runs, matches M4b/M4c root-cause diagnosis (bare
   `search_path` + shared-pool leakage).
2. **Opening a DB in a test** — canonical snippet using `testdb.Open(t)` + `db.SetMaxOpenConns(1)`
   when SUT takes `*sql.DB` directly.
3. **Factory builders** — table of every exported builder/composite in `factory.go` as-shipped
   (F4c.1 canonical): `NewTenant`, `NewUser` + `WithUserID`/`WithRole`/`WithTenant`, `NewTaxonomy`,
   `NewControlledDoc`, `NewDocument` + `WithStatus`/`WithRevisionVersion`/`WithScheduleGen`,
   `Scenario.PublishedDocument`, `Scenario.ScheduledRevision(gen)`. Each with a one-line description
   of what it seeds and the IDs it returns.
4. **Guarded writes** — when the SUT performs a tripwire-guarded write (INSERT/UPDATE on a capped
   table), the test must assert caps. Three sanctioned patterns (from F4c.4 `test-discipline.md`):
   `SeedWithCaps`, `SetCapsOnTx`, `SetCapsOnDB` — when to use each.
5. **Qualified table names** — `testdb.Qualified(schema, "tablename")` for any raw SQL query
   against a module-owned table. Why: schema clone isolation.
6. **Minimal working example** — a 15–25 line end-to-end snippet: `Open(t)` → `NewTenant` →
   `NewDocument` → assert a repository method's output. Not a copy of existing tests; purpose-built
   for the how-to audience.
7. **Cross-links** — `test-discipline.md` (CI rules), ADR 0034 (decision record), `factory.go` source.

The page must be **accurate at close** — every builder name, every helper signature must match what
`factory.go` exports. No invented API, no forwarded description from a plan that may not have landed.

### B. `wiki/decisions/0034-integration-test-fixture-framework.md` (new ADR)

Follows the canonical `wiki/decisions/` format (Status/Last verified/Scope/Key files/Context/
Decision/Consequences/References, matching ADRs 0029–0033).

Content:
- **Status:** Accepted 2026-06-15
- **Scope:** the test-fixture framework decision — IntegreSQL template-DB-per-test, `testdb`
  factory package, four discipline rules (R1–R4), CI guard placement.
- **Context:** M4b root cause (bare `search_path` + shared-DB cross-test leakage); the three
  alternatives considered (shared-DB with truncate, per-test migration, IntegreSQL clone); why the
  third was chosen.
- **Decision:** `tests/integration/testdb/` is the canonical home for the factory; template-DB-per-
  test via IntegreSQL is the harness; four rules (R1–R4) are enforced by CI guard
  `scripts/check-test-discipline.sh`; no test file may use raw `set_config`/bare table names/
  hardcoded tenant literals outside testdb/.
- **Consequences:** positive (isolation, parallel safety, root-cause closed); negative/constraints
  (IntegreSQL dependency, `INTEGRESQL_URL` must be set in CI/dev); deferred (allowlist 5 files).
- **Key files:** `tests/integration/testdb/factory.go`, `tests/integration/testdb/db.go`,
  `tests/integration/testdb/fixtures.go`, `scripts/check-test-discipline.sh`,
  `.github/workflows/module-boundaries.yml`.
- **References:** ADR 0032 (M4b teardown decision), `wiki/quality/integration-test-harness.md`,
  `wiki/quality/test-discipline.md`.

### C. Index updates

- `wiki/decisions/index.md` — add ADR 0034 entry.
- `wiki/quality/index.md` — add `integration-test-harness.md` entry (alongside existing
  `test-discipline.md` entry landed in F4c.4).
- `wiki/README.md` / `wiki/index.md` — no change needed (both already index `quality/index.md`).

### D. `wiki-curator` dispatch

After all artifacts are written and committed, dispatch the `wiki-curator` agent to:
- Refresh `Last verified:` stamps on docs that reference `testdb` / `factory.go`.
- Fix any broken `file:line` anchors in affected wiki docs.
- Verify the new docs are reachable from `wiki/README.md` read-path.

## Non-goals (mandatory)

- **No code change.** `factory.go`, `db.go`, `fixtures.go`, `scripts/check-test-discipline.sh`,
  any test file — all read-only in F4c.5.
- **No new factory builder.** The how-to must describe the API that shipped in F4c.1, not a
  hypothetical extension.
- **No ADR for F4c.4 specifically** — the guard is an enforcement mechanism, not a durable
  architecture decision beyond what ADR 0034 covers. ADR 0034 covers the framework holistically;
  `test-discipline.md` (F4c.4) covers the rules.
- **No wiki rebuild / full maturity promotion** — F4c.5 ships two files + index updates. Full
  module wiki rebuild is a separate `metaldocs-module-doc` skill invocation.

## Validation Gate (concrete — approved before code)

| AC | Acceptance criterion | Named proof command | Real vs fixture |
|----|----------------------|---------------------|-----------------|
| **AC1** | `wiki/quality/integration-test-harness.md` exists + ≥ 80 lines. | `wc -l wiki/quality/integration-test-harness.md` ≥ 80 | real |
| **AC2** | Every factory builder in `factory.go` is named in the harness doc (no invented builders). | `grep -E "^func (New\|.*Scenario)" tests/integration/testdb/factory.go` — each exported name appears in the harness doc. | real |
| **AC3** | `wiki/decisions/0034-integration-test-fixture-framework.md` exists with canonical Status/Scope/Key files/Context/Decision/Consequences/References headers. | `grep -E "^## (Context\|Decision\|Consequences\|References)" wiki/decisions/0034-integration-test-fixture-framework.md` → 4 matches | real |
| **AC4** | ADR 0034 registered in `wiki/decisions/index.md`. | `grep "0034" wiki/decisions/index.md` → match | real |
| **AC5** | `integration-test-harness.md` registered in `wiki/quality/index.md`. | `grep "integration-test-harness" wiki/quality/index.md` → match | real |
| **AC6** | Cross-links correct: harness doc links to `test-discipline.md` and ADR 0034; ADR 0034 links to harness doc and `test-discipline.md`. | `grep -n "test-discipline\|0034\|integration-test-harness" wiki/quality/integration-test-harness.md wiki/decisions/0034-integration-test-fixture-framework.md` → all four cross-links present | real |
| **AC7** | `wiki-curator` dispatched; curator returns with refreshed stamps or a "no drift" report. | curator dispatch + return recorded in evidence | real |
| **AC8** | No production-source change. | `git diff --name-only HEAD~2..HEAD -- internal/ db/ scripts/ tests/` → empty (only wiki/ docs/ changed) | real |

> F4c.5 is documentation-only — TDD does not apply. Evidence rows substitute structural checks
> (file exists, line counts, grep matches) for test GREEN/RED. The "RED" equivalent is: the spec
> says these files must exist; they do not exist yet; implementing = creating them.

## ADR needed?

- [x] **Yes — ADR 0034** is this feature's primary deliverable. No additional ADR needed.
