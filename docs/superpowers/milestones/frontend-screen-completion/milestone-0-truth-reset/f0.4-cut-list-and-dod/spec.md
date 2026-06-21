# Feature F0.4 — Spec

> **Milestone:** 0 — Truth reset & structural cleanup  ·  **Folder:** `f0.4-cut-list-and-dod`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca — *the DoD content is fixed by mission D2; the CUT set by D3; both operator-locked. No open decision.*

> This is the feature's **contract**, written and approved **before any code**. The milestone-validator
> judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | The per-screen Definition-of-Done is exactly the **D2 gate** ("both `frontend-screen-reviewer` and `frontend-code-reviewer` return APPROVE on record, and tests green") quoted verbatim from `mission.md` §3. The CUT set is **D3** (`alternativas-inicio-caixa` + `catalogo-slots`). Both are operator-locked; F0.4 records them, it decides nothing new. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** every later milestone's screen feature (M1–M5) reads the DoD doc to know when a
  screen "counts done"; the `milestone-validator` for M1+ cites it; the operator reads the CUT registry
  to confirm the two cut slugs are never built; the router must not mount the CUT slugs.
- **Contract:** a single durable wiki doc exists that enumerates the **D2 per-screen gate** (visual
  reviewer APPROVE + code reviewer APPROVE, both on record, + tests green), composed with the existing
  runtime [`screen-qa-checklist.md`](../../../../wiki/quality/screen-qa-checklist.md) (it references that
  checklist, does not duplicate it). The two CUT slugs are recorded **as CUT with rationale** and are
  **absent from the router** (already true after F0.2/F0.3; F0.4 documents + asserts it).
- **Source of truth for the contract:** `mission.md` D2 (gate), D3 (cut list), §7 program-architecture
  note "FE features also: screen-reviewer + code-reviewer APPROVE (D2)", and the existing
  `wiki/quality/screen-qa-checklist.md` (the functional dimension the DoD composes).

## What this feature implements

1. **Author `wiki/quality/screen-definition-of-done.md`** — the per-screen DoD every M1–M5 screen
   feature cites. It enumerates, as objective close criteria:
   - real API data (no `MOCK_` / illustrative / "em breve" literal),
   - redesign design-system tokens (no ad-hoc inline styles),
   - **`frontend-screen-reviewer` APPROVE on record** (visual/parity),
   - **`frontend-code-reviewer` APPROVE on record** (architecture/maintainability),
   - tests green (`make test` / `npm run build` clean),
   - the runtime functional pass — by reference to `screen-qa-checklist.md` (not duplicated),
   - a **CUT registry** sub-section listing `alternativas-inicio-caixa` + `catalogo-slots` with the D3
     rationale (no route, no NOTES, no product intent) and `biblioteca` clarified as already shipped
     (`LibraryPage`, not a gap).
2. **Cross-link:** add a pointer from the screen tracker
   (`wiki/implementation/screen-redesign-tracker.md`) to the new DoD doc so the resume doc and the gate
   doc reference each other.

## Non-goals (mandatory)

- **Not** building or designing the CUT slugs (D3).
- **Not** re-writing `screen-qa-checklist.md` — the DoD *composes* it by reference.
- **Not** changing any reviewer-agent definition — F0.4 documents the gate, it does not re-spec the agents.
- **Not** re-stamping the tracker's redesign-spec governance lineage (M0 rabbit-hole).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| DoD doc exists under `wiki/` | `ls wiki/quality/screen-definition-of-done.md` → present | real |
| DoD enumerates the D2 two-reviewer + tests gate | `grep -E "frontend-screen-reviewer\|frontend-code-reviewer" screen-definition-of-done.md` → both present; tests-green criterion present | real |
| Both CUT slugs documented as CUT | `grep -E "alternativas-inicio-caixa\|catalogo-slots" screen-definition-of-done.md` → both present with CUT rationale | real |
| CUT slugs absent from router | `grep -rEn "alternativas-inicio-caixa\|catalogo-slots" frontend/apps/web/src` → **0** | real |
| Tracker ↔ DoD cross-link | `grep "screen-definition-of-done" wiki/implementation/screen-redesign-tracker.md` → ≥1 | real |

> No automated unit test (governance docs). Every criterion is a deterministic `ls`/`grep`.

## ADR needed?

- [x] No durable decision — skip. The DoD restates mission **D2**; the cut list restates mission **D3**.
  Both decisions are already operator-locked in `mission.md`; F0.4 records them in the wiki, it makes no
  new decision.
