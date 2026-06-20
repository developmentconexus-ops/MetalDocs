# Feature F0.1 — Evidence — ADR-0039

> **Milestone:** 0  ·  **Feature:** `f0.1-adr-0039`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).

## What was implemented

- **New ADR** `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (status **Accepted 2026-06-20**):
  - **D1** — one-sentence boundary rule: raw `SELECT`/`JOIN`/subquery/`EXISTS` against another module's owned
    **base table** = H-G violation; published view / owner read-port / own tables = compliant. Owner = module
    holding the writes.
  - **D2** — "owned table" refined to "owned **base** table" so H-G = 0 holds under both readings (strict +
    canonical greps agree, because the only sanctioned cross-module SQL is against published views).
  - **D3** — exemption list: (a) published versioned view/read-model, (b) owner-published read-port, (c) own
    tables.
  - **D4** — active-now membership view encodes exactly `effective_to IS NULL` (ADR 0037, no reinterpretation).
  - **D5** — H-PRE-1 preserved (SELECT-only views are tx-structure-neutral; no authz-recording read inside a
    lock-holding tx).
  - **D6** — CI-enforced by the cilint `hgcrossmodule` guard (built in F0.3).
  - **Worked classification table** — all 13 mission §5 SQL/coupling sites (rows 3–15) classified
    violation/compliant with the deciding clause; **0 unclassified**.
- **`wiki/decisions/index.md`** — added the 0039 row; **backfilled the missing 0038 row** (pre-existing
  governance gap encountered while editing the index — 0038 was Accepted but absent from the maintained list);
  bumped `Last verified` to 2026-06-20.
- Producer-matches-consumer: the ADR's definition (D1–D3) is the exact mechanical rule F0.2's census and
  F0.3's guard consume; the worked table is the seed F0.2 reconciles against and F0.3's expected-red set
  derives from. Consumer contract honored.
- Commits: local only (not yet committed; staged for the M0 close commit — see milestone-level evidence).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| ADR file exists | `ls wiki/decisions/0039-*.md` | `0039-cross-module-base-table-read-boundary.md` | real |
| Status Accepted | `grep -i "Status.*Accepted" …/0039-*.md` | `> **Status:** Accepted 2026-06-20` | real |
| Active-now contract present | `grep -c "effective_to IS NULL" …/0039-*.md` | `3` matches | real |
| H-PRE-1 note present | `grep -c "H-PRE-1" …/0039-*.md` | `5` matches | real |
| All 3 D4 exemptions named | `grep -cE "Published versioned view\|Owner-published read-port\|Own tables"` | `3` | real |
| Links to ADR 0022/0037/0038 | `grep -oE "003[78]\|0022" … \| sort -u` | `0022`, `0037`, `0038` | real |
| Worked classification covers §5 rows 3–15 | `grep -cE '^\| [0-9]+ \|' …/0039-*.md` | `13` rows (rows 3–15), 0 unclassified | real |
| Docs-governance status gate | `grep -L '^> \*\*Status:\*\*' …/0039-*.md` | (no output — gate clean) | real |

> Documentation artifact — no TDD/runtime provider. The "unambiguous definition" acceptance is proven by the
> worked classification table: a reviewer applying D1/D3 to each mission §5 site reaches the tabulated verdict
> with no judgement. This is real (the table is in the committed ADR), not fixture.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| ADR file exists at `wiki/decisions/0039-*.md` | yes | row 1 above |
| Status is Accepted | yes | row 2 |
| Definition unambiguous — every §5 site (3–15) has a verdict + deciding clause, 0 unclassified | yes | row 7 (13 rows) + the ADR "Worked classification" section |
| Exemption list names all three D4 categories | yes | row 5 |
| Active-now contract states `effective_to IS NULL` + ADR-0037 citation | yes | row 3 + ADR D4 |
| H-PRE-1 note present | yes | row 4 |
| Links to ADR 0022/0037/0038 present | yes | row 6 |
| Operator ratification | **deferred** → M0 HS-1 gate | bounded defer below |

## Review disposition

- Spec-compliance review: self-reviewed against `spec.md` consumer contract — all contract elements present
  (rule, exemption list, view contract, H-PRE-1, worked table, status, links). The independent
  `milestone-validator` (Phase 4) is the separation-of-powers reviewer of record for M0.
- Code-quality review: N/A (no code). Docs-governance conformance checked: status header from the canonical
  vocabulary, four-digit sequential number (next free = 0039), index synced, `Last verified` stamped.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Operator ratification of the ADR definition | The acceptance row defines ratification as occurring at the M0 HS-1 gate, *after* the milestone-validator PASS — by design, not skipped. | **Trigger:** M0 HS-1 operator review gate (post-validator). **Owner:** operator (leandrotca). |

## Addendum 2026-06-20 — ADR-0039 amended (HS-4 ← F0.2 census, operator-ruled)

The F0.2 binding census, widening beyond the §5 token set, surfaced sites the original worked table did not
cover (N1 document-domain; X1–X8 auth/audit/platform), creating a Non-Goal/terminal-bar contradiction → HS-6.
The operator ruled (`../f0.2-binding-census/hs-6-scope-decision.md`): fold N1 into M2, and add **principled
exemptions** for X1–X8. F0.1's ADR was therefore amended (this is the HS-4 "back to the named feature" loop):

- **`wiki/decisions/0039-*.md`** gained **D3(d)** platform append-sink (`audit_events`), **D3(e)**
  parent-grade-a-dispositioned auth (`auth_identities`/`auth_sessions`, ADR 0029/0031), **D3(f)** worker-layer
  (`jobs`); a second classification table (N1 + X1–X8); and an **honest scope note** redefining the terminal
  "H-G=0 under both readings" as **0 violations outside the recorded F0.3 allowlist** (the carve-outs are
  enumerated and justified, not pretended absent). Header marked *amended 2026-06-20*.
- The original worked table (mission §5 rows 3–15) is unchanged and still **0 unclassified**; the amendment
  *adds* the widen's sites, it does not revise any prior verdict.
- This keeps F0.1's "definition is mechanical / 0 unclassified" acceptance intact under the wider census, and
  is the dependency F0.3's guard allowlist consumes.
