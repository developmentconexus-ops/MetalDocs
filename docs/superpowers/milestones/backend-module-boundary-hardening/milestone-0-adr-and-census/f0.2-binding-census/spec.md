# Feature F0.2 — Spec — binding re-census against ADR-0039

> **Milestone:** 0 — ADR-0039: lock the definition + binding census + CI guard  ·  **Folder:** `f0.2-binding-census`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-20 / leandrotca — *contract locked in `../mission.md` §7 F0.2 + §4 (discovery summary's deferred-to-M0 assumptions); no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. The milestone-validator judges
> the feature against *this* file (C1).

## Interview record (fail-closed gate)

The census scope and acceptance were locked by the mission (§7 F0.2, §4, §2 Non-Goals). The one genuinely
open question the discovery brief deferred — *how wide does "widen beyond the named tokens" go* — is answered
by the mission's own non-goal language. Recorded below; no fresh operator interview needed.

| # | Question | Answer |
|---|----------|--------|
| 1 | What must the census reproduce? | **Locked (§7 F0.2):** the ~20 discovery-brief sites — none dropped. The brief (`../discovery-brief.md`) is the starting inventory; this census is the skeptic re-run against ADR-0039. |
| 2 | How wide is the "re-scope owned-table set beyond the named tokens"? | **Locked (§2 Non-Goals + §4):** widen from the named H-G tokens (`user_process_areas`, `controlled_document*`, `document_process_areas`, `documents`, `approval_instances`, `document_profiles`) to the **full owned-base-table set** across `internal/modules/**` non-test — i.e. enumerate every module's owned base tables and grep for non-owner reads of each. The brief explicitly deferred this widening to M0. |
| 3 | What is the classification rule? | **Locked (ADR-0039 D1/D3):** raw cross-module **base-table** read = violation; published view / owner read-port / own table = compliant. The census applies *that* rule — it does not invent a new one. |
| 4 | What happens if a NEW in-scope site is found? | **Locked (§7 F0.2 + HS-6):** add it to the relevant milestone if it fits the existing shape; **if it changes a milestone's shape → HS-6 stop**, surface to the operator before continuing. Genuinely out-of-scope reads are recorded with reason. |
| 5 | Is Docker / runtime verification required? | **Locked (§8.5, brief coverage statement):** no — the census is **static read** (grep + source inspection). Docker Postgres (:5433) may be down; M0 has no integration step. The residual (dynamic/aliased SQL invisible to a token grep) is **recorded as an assumption**, not runtime-reproduced. |
| 6 | Anything else open? | none needed — scope fully specified by the locked decisions. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  1. **M1–M4** — consume the census's **authoritative in-scope site list** as their work inventory (which
     sites each milestone ports). A missed site = a missed remediation = a third undercount.
  2. **F0.3 cilint guard** — consumes the census's **table→owner map** (which base tables are owned by which
     module) to know what to flag, and the **expected-red site set** to test the guard against.
  3. **The terminal mission-validator** (§8) — spot-checks the census's "0 remaining" claim by re-grepping a
     sample of sites.
- **Contract (the exact shape consumers rely on):** a census document (`census.md` in this folder) containing:
  - A **table→owner map** for every owned base table referenced cross-module (table, owning module, evidence
    = where its writes live).
  - The **authoritative in-scope site list**: each site = `file:line`, SQL excerpt, owner read, reader
    module, ADR-0039 verdict (violation), class (A/B/C/C4), assigned milestone. Reproduces the ~20 brief sites
    + any new ones from the widen.
  - A **delta vs the discovery brief**: sites added / moved / confirmed / (none) dropped, each with reason.
  - A **coverage statement**: what was swept (tokens, paths), what is assumed/not-swept (dynamic/aliased SQL,
    `_test.go`), and the explicit `0 sites unclassified` count.
- **Source of truth for the contract:** `../discovery-brief.md` (starting inventory + method),
  `wiki/decisions/0039-*.md` (the classification rule + worked table), the live tree under `internal/modules/`.

## What this feature implements

A binding census document `f0.2-binding-census/census.md` that re-runs the cross-module owned-base-table read
sweep against ADR-0039, widened to the full owned-table set, producing the authoritative in-scope inventory +
table→owner map + brief-delta + coverage statement, with **0 sites unclassified**.

## Non-goals (mandatory)

- **No porting.** The census *measures*; it does not *fix*. Zero production SQL edited.
- **No view/port/guard creation.** Those are F0.3 (guard) and M2–M4 (views/ports).
- **No new classification rule.** The census applies ADR-0039 D1/D3 as-is. If a site can't be classified by
  that rule, the rule is wrong — stop (HS-4 back to F0.1), don't invent a verdict.
- **No runtime/Docker verification** — static read only; the residual is recorded, not reproduced.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Census reproduces every discovery-brief site (A1–A3, B1–B8, C1–C4 ⇒ ~20) — none dropped | `census.md` site list cross-checked vs `../discovery-brief.md`; re-grep each named token confirms the cited `file:line` still reads the foreign table | real (static grep) |
| Owned-table set widened beyond the named tokens — table→owner map enumerates every cross-module-read owned base table, not just the 6 named ones | `census.md` table→owner map vs a fresh full-tree grep of `FROM`/`JOIN`/`EXISTS` targets in `internal/modules/**` | real |
| **0 sites unclassified** — every cross-module SQL read has an ADR-0039 verdict | `census.md` explicit count line `unclassified: 0`; every site row carries a verdict | real |
| New in-scope sites (if any) routed to a milestone or recorded out-of-scope with reason; shape-changing ones flagged HS-6 | `census.md` delta section | real |
| Coverage statement present — swept set, assumed/not-swept residual, no silent caps | `census.md` coverage section | real |

> Static-analysis feature — "real" = the actual grep output over the live tree, recorded. No fixture. The
> census is the skeptic pass over the brief; its rigor is the per-site re-grep, not a re-assertion of the brief.

## ADR needed?

- [x] **No durable decision** — the census *applies* ADR-0039; it makes no new decision. (If the widen
  surfaces a site the rule can't classify, that is an ADR-0039 defect → HS-4 to F0.1, not a new ADR here.)
