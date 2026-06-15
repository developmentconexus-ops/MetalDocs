# Feature F4c.5 — Evidence

> **Milestone:** 4c — Unified test-fixture framework  ·  **Feature:** `f4c.5-docs-adr`  ·  **Closed:** 2026-06-15
> **Contract:** [`spec.md`](spec.md) (harness wiki page + ADR 0034 + index updates + wiki-curator).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output.
> **Baseline SHA entering F4c.5:** `a554bb07` (F4c.4 close)

---

## AC1 — `wiki/quality/integration-test-harness.md` exists + ≥ 80 lines

```
$ wc -l wiki/quality/integration-test-harness.md
225 wiki/quality/integration-test-harness.md
```

225 ≥ 80. **PASS**

---

## AC2 — Every factory builder in `factory.go` named in harness doc (no invented builders)

```
$ grep -E "^func (New|.*Scenario)" tests/integration/testdb/factory.go
func NewTenant(t *testing.T, db *sql.DB, opts ...Opt) Tenant {
func NewUser(t *testing.T, db *sql.DB, opts ...Opt) User {
func NewTaxonomy(t *testing.T, db *sql.DB, opts ...Opt) Taxonomy {
func NewControlledDoc(t *testing.T, db *sql.DB, opts ...Opt) ControlledDoc {
func NewDocument(t *testing.T, db *sql.DB, opts ...Opt) Document {
func NewApprovalRoute(t *testing.T, db *sql.DB, opts ...Opt) ApprovalRoute {
func NewApprovalInstance(t *testing.T, db *sql.DB, opts ...Opt) ApprovalInstance {
func (Scenario) PublishedDocument(t *testing.T, db *sql.DB, opts ...Opt) Document {
func (Scenario) ScheduledRevision(t *testing.T, db *sql.DB, gen int64, opts ...Opt) Document {
```

Harness doc check per name:

| Builder | Present in harness doc |
|---------|----------------------|
| `NewTenant` | FOUND |
| `NewUser` | FOUND |
| `NewTaxonomy` | FOUND |
| `NewControlledDoc` | FOUND |
| `NewDocument` | FOUND |
| `NewApprovalRoute` | FOUND |
| `NewApprovalInstance` | FOUND |
| `PublishedDocument` (Scenario) | FOUND |
| `ScheduledRevision` (Scenario) | FOUND |

All 9 builders present. No invented builders. **PASS**

---

## AC3 — ADR 0034 exists with canonical headers

```
$ grep -E "^## (Context|Decision|Consequences|References)" wiki/decisions/0034-integration-test-fixture-framework.md
## Context
## Decision
## Consequences
## References
```

4 matches — all required canonical headers present. **PASS**

---

## AC4 — ADR 0034 registered in `wiki/decisions/index.md`

```
$ grep "0034" wiki/decisions/index.md
| [0034](0034-integration-test-fixture-framework.md) | Integration Test Fixture Framework (IntegreSQL template-DB-per-test) | Accepted 2026-06-15 | — | `tests/integration/testdb/` canonical harness; template-DB-per-test isolation; factory builders; R1–R4 discipline rules; CI guard `scripts/check-test-discipline.sh` in `module-boundaries` workflow (M4c F4c.1 + F4c.4) |
```

Match. **PASS**

---

## AC5 — `integration-test-harness.md` registered in `wiki/quality/index.md`

```
$ grep "integration-test-harness" wiki/quality/index.md
- [integration-test-harness.md](integration-test-harness.md) - how-to guide for writing integration tests: harness choice, factory builders, guarded writes, qualified table names, working examples
```

Match. **PASS**

---

## AC6 — Cross-links correct

```
$ grep -n "test-discipline\|0034\|integration-test-harness" wiki/quality/integration-test-harness.md wiki/decisions/0034-integration-test-fixture-framework.md
wiki/quality/integration-test-harness.md:4:> See also: [test-discipline.md](test-discipline.md) (CI guard rules R1–R4), [ADR 0034](../decisions/0034-integration-test-fixture-framework.md).
wiki/quality/integration-test-harness.md:29:See [ADR 0034](../decisions/0034-integration-test-fixture-framework.md) for the full decision record.
wiki/quality/integration-test-harness.md:126:**Never use raw `set_config` SQL in test files** — that is R1/R2 in `test-discipline.md`.
wiki/quality/integration-test-harness.md:225:See [test-discipline.md](test-discipline.md) for the full per-rule CI enforcement reference.
wiki/decisions/0034-integration-test-fixture-framework.md:121:- [wiki/quality/integration-test-harness.md](../quality/integration-test-harness.md) — how-to guide
wiki/decisions/0034-integration-test-fixture-framework.md:122:- [wiki/quality/test-discipline.md](../quality/test-discipline.md) — R1–R4 rules, allowlist, CI usage reference
```

Four required cross-links:
- harness doc → `test-discipline.md`: FOUND (lines 4, 126, 225)
- harness doc → ADR 0034: FOUND (lines 4, 29)
- ADR 0034 → harness doc: FOUND (line 121)
- ADR 0034 → `test-discipline.md`: FOUND (lines 13, 81, 89, 122)

All four cross-links present. **PASS**

---

## AC7 — `wiki-curator` dispatched and returned

`wiki-curator` dispatched after all artifacts written. Curator report (verbatim summary):

- **Reachability chain verified intact:** `wiki/README.md` → `wiki/index.md` → `wiki/quality/index.md` → `integration-test-harness.md`; `wiki/index.md` → `wiki/decisions/index.md` → ADR 0034. No gaps.
- **Key file anchors verified:** all 5 key files named in ADR 0034 exist on disk (`factory.go`, `db.go`, `fixtures.go`, `scripts/check-test-discipline.sh`, `.github/workflows/module-boundaries.yml`). No broken anchors.
- **One live doc updated:** `wiki/backend/repo-topology.md` — stamp bumped to 2026-06-15, `internal/testsupport/pgtest/` entry replaced with deletion notice + cross-link to harness doc, `tests/integration/testdb/` row description updated to accurate canonical harness description with cross-links to new docs.
- **Historical artifacts left untouched** as per curator policy.
- **No drift found** in the two new docs or the index files.

Curator returned with a "targeted changes, no drift" report. **PASS**

---

## AC8 — No production-source change

```
$ git diff --name-only a554bb07 HEAD -- internal/ db/ scripts/ tests/
(empty — no changes to those paths)
```

F4c.5 changed only:
- `wiki/quality/integration-test-harness.md` (new)
- `wiki/decisions/0034-integration-test-fixture-framework.md` (new)
- `wiki/decisions/index.md` (two rows added)
- `wiki/quality/index.md` (one row added)
- `wiki/backend/repo-topology.md` (curator stamp + two rows updated)
- `docs/superpowers/milestones/.../f4c.5-docs-adr/evidence.md` (this file)

Zero production-source changes. **PASS**

---

## Bounded defers

None — F4c.5 is documentation-only. All deliverables complete. No code paths touched.
