# Milestone 0 — Truth reset & structural cleanup

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Validated — `milestone-validator` **PASS** 2026-06-21 (after FAIL → fix F0.5 → re-judge PASS). Awaiting **HS-1** operator sign-off before M1.
> **Authored:** 2026-06-21 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

After this milestone, the routed MetalDocs web app is **honest**: a user landing on `/`
reaches exactly one intended page (no shadowed duplicate index), every mounted route
renders a real screen (no empty no-API shell), the screen tracker reflects verified
2026-06-21 reality instead of two-week-stale claims, and the per-screen quality bar (the
D2 reviewer gate) is written down where every later milestone can cite it.

**Quality bar moved:** the "routed app contains dead/dishonest surface" defect class is
closed — re-measured by three greps returning 0 (single root index route, zero
`OperationsCenter`-shell routes, zero tracker rows contradicting the implemented page set)
plus a clean frontend build + test run. This milestone ships **no user-facing feature**;
it clears the ground so M1..M5 build on solid, truthful structure (D5 "clear-the-ground
first").

## Appetite + rabbit holes

**Appetite:** small — 4 features, structural/governance only. No new screens, no API
wiring, no redesign of any primitive or token. If a feature here grows past
"delete / fix route / rewrite a doc," it has escaped M0.

**Rabbit holes (do not chase):**
- **Re-styling or re-reviewing any already-DONE screen** — out of scope; M0 only removes
  dead surface and records truth. Touching a shipped screen trips HS-2.
- **Wiring Dashboard mock data to real APIs** — that is M1 (F1.1/F1.2), not M0. M0 stops
  at recording that Dashboard ships mocks; it does not fix them.
- **Building or designing the CUT slugs** (`alternativas-inicio-caixa`, `catalogo-slots`)
  — they are recorded as CUT, never built (D3).
- **Migrating IAM Admin Center metrics/audit/sessions** — IAM already owns them (D7); M0
  only deletes the *redundant* Operations/Audit stubs, it does not rebuild their function
  anywhere.
- **Repointing the tracker's `Branch`/`Spec` header to a new redesign program** — M0
  rewrites row *truth*, not the tracker's governance lineage.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F0.1 | `f0.1-tracker-rewrite` | Rewrite `wiki/implementation/screen-redesign-tracker.md` so every screen row states its verified 2026-06-21 status (done / partial / stub / not-started / cut), sourced from this mission's inventory + discovery findings. **Consumer:** the operator + every later milestone reads this as the durable resume doc. | Every tracker row matches a grep of the implemented pages under `frontend/apps/web/src/features/**/pages` and the mission §5 inventory; **0** rows contradict reality (the M0 validator re-greps a sample). The previously-wrong rows (finding 1: Editor/Publicado claimed "not started" though both ship) now read correctly. |
| F0.2 | `f0.2-index-route-fix` | Resolve the duplicate root `index: true` (`dashboard/routes.tsx:5` vs `operations/routes.tsx:5`, finding 2) so `/` resolves to exactly one intended page (Dashboard). **Consumer:** the app router / a user navigating to `/`. | Exactly **one** root-level `index: true` route remains after F0.3 removes the operations route; a router/render test asserts `/` mounts the intended Dashboard component; `npm run build` is clean. |
| F0.3 | `f0.3-dead-stub-disposition` | **Delete** `OperationsPage` (`features/operations/pages/OperationsPage.tsx`), `AuditPage` (`features/audit/pages/AuditPage.tsx`), the orphaned `OperationsCenter` component (`src/components/OperationsCenter.tsx`), and their route registrations — IAM Admin Center already owns metrics/audit/sessions (D7). **Consumer:** the router (no longer mounts an empty shell) + future maintainers (no dead code). | `OperationsPage`/`AuditPage`/`OperationsCenter` source files are gone; `grep -rn "OperationsCenter" frontend/apps/web/src` = **0**; no route renders an empty no-API shell; their route entries are removed from the router config; frontend build + `make test` green. |
| F0.4 | `f0.4-cut-list-and-dod` | Record `alternativas-inicio-caixa` + `catalogo-slots` as **CUT** (D3) with rationale, and author the per-screen **Definition-of-Done** doc enumerating the D2 gate (both `frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE on record + tests green) that every later milestone's screen features cite. **Consumer:** every later milestone's screen feature reads the DoD; the router must not mount the CUT slugs. | A DoD doc exists under `wiki/` enumerating the D2 two-reviewer + tests-green gate; the two CUT slugs are documented as CUT **and** absent from the router (`grep` for either slug in router config = 0). |
| F0.5 | `f0.5-tracker-post-deletion-reconcile` | **Fix feature** opened per the M0 milestone-validator FAIL (HS-4): F0.1 rewrote the tracker *before* F0.3 deleted Operations/Audit, leaving rows 22–23 presenting deleted screens as live `stub`s citing deleted files (split-brain, RM3). Reconcile those two rows to post-F0.3 reality (`cut`/deleted) + correct the evidence-base count. **Consumer:** the tracker resume doc + validator RM3. | RM3 returns **0 MISSING**: no live-presented row cites a non-existent file; Operations + Audit rows read `cut`; evidence-base count matches the post-deletion page set. |

For each feature, "what to validate" above is objectively checkable — a grep returning a
number, a route test, a clean build. No row closes on "works" or "looks right."

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it
judges and writes `qa/milestone-qa.md`; the main session flips status only on its PASS),
following the binding C1–C7 checklist in
`.claude/skills/milestone/references/milestone-end-validation.md`. What that gate enforces
for M0:

1. **Per-feature acceptance** — each of F0.1–F0.4 meets its "What to validate" cell above,
   and each feature's `spec.md` consumer contract was honored.
2. **Workflow-class QA checklist** — frontend build + test discipline:
   `wiki/quality/test-discipline.md` and the frontend test suite (`make test` /
   `npm run build` clean). M0 is structural/FE-only — no backend-api checklist applies.
3. **Regression** — none prior (M0 is the first milestone). The validator confirms the
   broader frontend build/test suite is green after M0's deletions (no orphan imports, no
   broken router).
4. **Quality-bar / root-cause check** — the "dead/dishonest routed surface" class is
   confirmed **removed at the root** (files deleted, routes unregistered, tracker truthful),
   not hidden behind a redirect or a commented-out route. Re-measure commands:
   - `grep -rn "index: true" frontend/apps/web/src/features/**/routes.tsx` → exactly one
     **root** index (Dashboard) remains (nested `documents`/`iam` index entries are their
     own parents, not root).
   - `grep -rn "OperationsCenter" frontend/apps/web/src` → 0.
   - tracker row-vs-implemented-page sample re-grep → 0 contradictions.
5. **No unplanned scope** — anything implemented beyond F0.1–F0.4 (or touching a rabbit-hole
   item above) is recorded with rationale or is a FAIL.

## Dependencies & constraints

- **Depends on:** nothing prior (first milestone). Consumes the operator-approved
  `mission.md` (D3, D5, D7) + `discovery-brief.md` findings 1–5, 16.
- **Quality goals (ranked):** 1) **truth/correctness** (the app and its tracker stop
  lying) > 2) **simplicity** (delete dead code rather than wire it) > 3) **zero
  regression** (deletions must not break the build or a shipped screen).
- **Architectural constraints (hard rules):**
  - Design system is **consumed, not redesigned** — M0 changes no token, no shared
    primitive (changing one trips HS-2).
  - **Root-cause over symptom-patch** — dead stubs are *deleted*, not hidden behind a
    redirect or feature flag (binds validation §4).
  - **No merge / no push by the agent;** commits allowed after verified work (CLAUDE.md §5.0).
  - FE screen work routes through the documented FE workflow; M0 is structural so no
    reviewer-agent gate is required on F0.1–F0.4 (no screen is built/restyled here) — the
    D2 reviewer gate that F0.4 *documents* binds M1+ screens, not M0 itself.
- **Risks:**
  - *Deleting Operations/Audit breaks an import elsewhere* → mitigate: F0.3 greps for all
    `OperationsCenter`/`OperationsPage`/`AuditPage` references before deletion and the
    validator re-runs build + tests (named in validation §3).
  - *Index-route fix changes the landing page a user sees* → low code blast, **high
    visibility** (discovery brief, blast-radius note) → mitigate: F0.2 ships a router test
    asserting `/` → Dashboard and a clean build.
  - *Tracker rewrite drifts into editorializing instead of recording verified state* →
    mitigate: every row must be grep-backed against an implemented page; the validator
    samples rows.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | **Milestone boundary (always).** After the validator PASSes M0, the operator reviews before M1 starts. No M1, no merge, without explicit approval. |
| HS-2 | A "structural cleanup" turns out to require changing a shared primitive or design token (e.g. removing `OperationsCenter` forces a token change). Stop; report the boundary + minimum prerequisite; no symptom-patch. |
| HS-3 | A prerequisite fails — frontend build won't run, the router won't compile, or a route/contract truth is broken by a deletion. Repair the prerequisite, rerun the failed checkpoint, then resume. |
| HS-4 | The `milestone-validator` returns FAIL (a feature missed acceptance, a forbidden-list hit, a symptom-patch). Open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Mid-milestone discovery that a "dead" stub is actually load-bearing, or the tracker rewrite surfaces an unmapped screen. Stop; surface the deviation; replan before continuing. |
