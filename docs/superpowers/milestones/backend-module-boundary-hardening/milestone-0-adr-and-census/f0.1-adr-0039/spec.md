# Feature F0.1 — Spec — ADR-0039: lock the H-G definition + exemption list

> **Milestone:** 0 — ADR-0039: lock the definition + binding census + CI guard  ·  **Folder:** `f0.1-adr-0039`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-20 / leandrotca — *contract locked in `../mission.md` §3 (D1–D6) + §7 F0.1; no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

The consumer contract for this feature was **already discovered and locked by the operator** during the
`mission` skill's Phase-2 strategy interview, recorded as decisions D1–D6 in `../mission.md` §3 and the F0.1
acceptance row in §7. Re-interviewing would re-litigate settled decisions. The rows below record the locked
answers and cite their source, satisfying the fail-closed gate (the contract was *read from the locked
decisions*, not guessed).

| # | Question | Answer |
|---|----------|--------|
| 1 | What is the H-G violation/compliant boundary? | **Locked (D1, D4):** raw `SELECT`/`JOIN`/`EXISTS` against another module's **base table** = violation; JOIN/read of another module's **published, versioned view**, or a call through an owner-**published read-port**, or reading one's **own** tables = compliant. "Owned table" is refined to "owned **base** table" so "H-G=0 under both readings" is coherent. |
| 2 | Which Category-C mechanism does the ADR sanction? | **Locked (D1):** C-α — iam publishes a versioned active-membership view; CD publishes a visibility projection for search. The ADR declares the published view/read-model compliant. (M0 only *declares* this; M2/M3/M4 *build* it.) |
| 3 | How does the active-now membership view encode "active"? | **Locked (D4, §10):** exactly `effective_to IS NULL` per ADR 0037 — no interval reinterpretation. |
| 4 | What must the ADR say about H-PRE-1? | **Locked (§10, HS-PRE-1):** record the rule — no authz-recording read inside a lock-holding atomic tx; the C-α view is SELECT-only and tx-structure-neutral, satisfying it for free. |
| 5 | What does the ADR supersede / relate to? | **Locked (§10):** relates to / extends ADR 0022 (authz root cause), 0037 (membership temporal model), 0038 (owner-published port precedent, sibling of 0031). Supersedes nothing. New number = **0039** (verified next free). |
| 6 | Anything genuinely open needing a fresh operator question? | **none needed** — the operator ratifies the *authored* ADR at the M0 HS-1 gate (acceptance row: "operator-ratified in the M0 HS-1 gate"). The authoring contract itself is fully specified by D1–D6 + §7. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  1. **F0.2 binding census** — classifies every cross-module read as compliant/violation. It needs a rule it
     can apply *mechanically* (no judgement) and the owned-base-table set the rule ranges over.
  2. **F0.3 cilint H-G guard** — encodes the ADR's allowlist (published views, owner ports, own-table reads)
     as its exemption set. It needs the exemption categories named precisely enough to implement.
  3. **The operator** — ratifies the definition at the M0 HS-1 gate.
  4. **M1–M4 + the terminal mission-validator** — cite ADR-0039 as the done-bar for "H-G = 0 under both
     readings".
- **Contract (the exact shape the consumers rely on):**
  - A **one-sentence decision rule** classifying any cross-module read as violation vs compliant, keyed on
    *base table* vs *published view / owner read-port / own table*.
  - A **named exemption list** (D4): (a) JOIN/read of another module's published versioned view; (b) call
    through an owner-published read-port; (c) reading one's own tables. Each stated so F0.3 can mechanize it.
  - The **active-now view contract**: `effective_to IS NULL` (ADR 0037 reference, verbatim semantics).
  - The **H-PRE-1 note**.
  - A **worked classification table** mapping each mission §5 site (rows 3–15) to compliant/violation under
    the rule — this is the proof the definition is *unambiguous* (acceptance), and the seed F0.2 reconciles
    against and F0.3's expected-red set is checked against.
  - Status **Accepted**; standard ADR front-matter (date, deciders, related ADRs, key code anchors).
- **Source of truth for the contract:** `../mission.md` §3 (D1–D6), §7 (F0.1 acceptance), §10 (constraints);
  `../discovery-brief.md` (the site inventory the worked table classifies); ADR 0037/0038 (antecedents).

## What this feature implements

A new ADR `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (slug TBD at authoring) that locks
the H-G definition + exemption list + active-now view contract + H-PRE-1 note, with a worked classification
table over the mission §5 sites, status **Accepted**, properly linked/supersession-stamped per docs
governance. Plus the `Last verified` stamp and code anchors per the ADR house style.

## Non-goals (mandatory)

- **No code.** F0.1 writes one wiki markdown file. No Go, no SQL, no migration.
- **No view/port creation.** The ADR *declares* the published view + read-ports compliant-by-design; building
  them is M2/M3/M4. Anything that creates `metaldocs.v_active_user_areas` here is scope drift.
- **No census.** Reproducing/verifying the full site inventory is F0.2. F0.1's worked table uses the
  discovery-brief sites as-given to *demonstrate the rule is unambiguous*, not to re-census.
- **No reinterpretation of ADR 0037** — the view predicate is `effective_to IS NULL`, cited, unchanged.
- **No guard.** Implementing the analyzer is F0.3.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| ADR file exists at `wiki/decisions/0039-*.md` | `ls wiki/decisions/0039-*.md` | real |
| Status is **Accepted** | `grep -i "Status.*Accepted" wiki/decisions/0039-*.md` | real |
| Definition is unambiguous — every mission §5 site (rows 3–15) appears in a worked classification table with a compliant/violation verdict and the rule-clause that decides it | manual read of the ADR's classification table vs `../mission.md` §5 — **0 sites without a verdict** | real |
| Exemption list names all three D4 categories (published view / owner read-port / own tables) | `grep` the ADR for each exemption category | real |
| Active-now contract states `effective_to IS NULL` with an ADR-0037 citation | `grep "effective_to IS NULL" wiki/decisions/0039-*.md` | real |
| H-PRE-1 note present | `grep -i "H-PRE-1" wiki/decisions/0039-*.md` | real |
| Links to ADR 0022/0037/0038 present | `grep -E "0022|0037|0038" wiki/decisions/0039-*.md` | real |
| Operator ratification | recorded at the M0 HS-1 gate (post-validator) | real |

> No TDD applies (documentation artifact). "Proof" = the grep/read assertions above, run and recorded in
> `evidence.md`. The worked classification table is the substantive proof of the unambiguity acceptance.

## ADR needed?

- [x] **Durable decision made → this feature *is* the ADR.** ADR-0039 under `wiki/decisions/`. Linked here on
  authoring: `wiki/decisions/0039-cross-module-base-table-read-boundary.md`.
