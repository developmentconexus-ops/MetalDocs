# Feature F0.4 — Evidence

> **Milestone:** 0  ·  **Feature:** `f0.4-cut-list-and-dod`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer = every M1–M5 screen feature + the milestone-validator + operator).

## What was implemented

- **Authored `wiki/quality/screen-definition-of-done.md`** — the binding per-screen Definition-of-Done
  the whole `frontend-screen-completion` program cites. It enumerates the **D2 gate** as 6 objective
  close criteria: real API data (no `MOCK_`/illustrative/"em breve"), redesign tokens,
  **`frontend-screen-reviewer` APPROVE on record**, **`frontend-code-reviewer` APPROVE on record**,
  tests green (`make test` + `npm run build`), and the runtime functional pass *by reference* to
  `screen-qa-checklist.md` (composed, not duplicated). Includes the D1 backend-blocked rule (endpoint
  built to Grade-A bar before criterion 1 can be met).
- **CUT registry (D3):** `alternativas-inicio-caixa` + `catalogo-slots` recorded as **CUT** with the
  "no route, no NOTES, no product intent" rationale; `biblioteca` clarified as already shipped
  (`LibraryPage`), not a gap. Carries the re-measure grep that must stay 0.
- **Cross-link:** added a "Per-screen Definition of Done (D2 gate)" pointer in the screen tracker header
  (`wiki/implementation/screen-redesign-tracker.md`) → the DoD doc, so resume-doc and gate-doc reference
  each other.
- **Producer matches consumer contract:** the doc's criteria are the D2 text quoted from `mission.md`
  §3; later milestones read exactly the gate the mission locked.
- Not yet committed (M0 commits at milestone close / operator discretion).

## Verification

| Check | Command | Result | Real vs fixture |
|-------|---------|--------|-----------------|
| DoD doc exists under `wiki/` | `ls wiki/quality/screen-definition-of-done.md` | **present** | real |
| DoD enumerates both reviewers | `grep -cE "frontend-screen-reviewer\|frontend-code-reviewer" …` | **2** lines | real |
| DoD has tests-green criterion | `grep -c "Tests green" …` | **1** | real |
| Both CUT slugs documented | `grep -cE "alternativas-inicio-caixa\|catalogo-slots" …` | **3** matches (both slugs + re-measure line) | real |
| CUT slugs absent from router | `grep -rEn "alternativas-inicio-caixa\|catalogo-slots" frontend/apps/web/src` | **exit 1, 0 matches** | real |
| Tracker ↔ DoD cross-link | `grep -c "screen-definition-of-done" …/screen-redesign-tracker.md` | **1** | real |

> No automated unit test — governance docs. Every criterion is a deterministic `ls`/`grep` against the
> working tree this session.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| DoD doc exists under `wiki/` | yes | `ls` present |
| DoD enumerates D2 two-reviewer + tests gate | yes | both reviewers (2) + tests-green (1) |
| Both CUT slugs documented as CUT | yes | grep = 3 (both slugs + rationale) |
| CUT slugs absent from router | yes | grep exit 1 / 0 |
| Tracker ↔ DoD cross-link | yes | grep = 1 |

All 5 criteria **met**.

## Review disposition

- Spec-compliance review: self-review against `spec.md` — PASS. DoD content is the D2 text from
  `mission.md`; CUT set is D3; `screen-qa-checklist.md` composed by reference (non-goal: not rewritten);
  no reviewer-agent re-spec; tracker governance lineage untouched.
- Code-quality review: n/a (docs). Independent judgement deferred to the M0 `milestone-validator`.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| DoD doc is unexercised until M1 | It is a forward-looking gate; M0 ships no screen to run it against | Trigger: M1/F1.1 (Dashboard) is the first screen feature to cite it. Owner: M1 |
