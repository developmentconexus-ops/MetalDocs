# Milestone 0 — Docs Progression De-Staling

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** Spec approved
> **Authored:** 2026-06-14 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which
> features** it contains, **what each feature implements**, and **what gets validated**. It
> contains **no execution steps** — the "how" of each feature lives in that feature's `plan.md`.
> The end-of-milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Produce **one unambiguous progression surface** so the architecture milestones (M1–M4) run against
clean docs and stale material stops polluting agent context. This milestone advances no code grade —
it is sequenced first (decision D5, "docs first") so that ADRs, roadmap, and backlog reflect reality
before the architecture work begins.

**Bar:** ambiguity / staleness eliminated across `wiki/decisions/`, roadmap, backlog, and the
post-deletion `docs/` reference graph. **Criterion that proves it moved:** the doc-QA gate below
passes — every ADR carries an accurate status, zero wiki links point at deleted `docs/` trees, and
exactly one forward roadmap exists.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F0.1 | `f0.1-adr-audit-ledger` | Run the `decisions/README.md` status-gate over every `wiki/decisions/` ADR; verify each ADR's decision still matches code; flag drift; refresh the `decisions/index.md` ledger | Every `wiki/decisions/` ADR has a `> **Status:**` line **and** that status matches code; drift is marked Historical/Superseded/amended; `decisions/index.md` ledger refreshed to match |
| F0.2 | `f0.2-stale-ref-repair` | Fix the 12 wiki refs that point at deleted `docs/` trees (`documentation-governance.md:73-96`, `decisions/index.md:34`, `quality/*`, `architecture/data-model.md`, `modules/frontend/iam.md`) | `grep` proves **0** wiki links to deleted `docs/` paths; broken-link sweep clean |
| F0.3 | `f0.3-roadmap-consolidation` | Mark `wiki/backlog/roadmap.md` (May) and `wiki/backend/roadmap.md` (June) **historical**; create **one** forward roadmap carrying this program + post-v1 progression | Exactly **1** forward roadmap exists; the 2 old roadmaps are clearly labeled historical |
| F0.4 | `f0.4-backlog-hygiene` | Close/archive completed items in `wiki/backlog/*`; keep only active deferred work | `wiki/backlog/*` contains active deferred work only; closed items archived, not deleted-without-trace |
| F0.5 | `f0.5-archive-convention` | Establish `wiki/_archive/`; move superseded-historical docs there; update domain `index.md`s + the governance migration map to post-deletion reality | `wiki/_archive/` tree exists; domain indexes + governance migration map accurately reflect moved docs; no dangling index entries |

For each feature, "what to validate" is **objectively checkable** (a grep count, a status line present
and matching, a single named artifact existing) — not "works" / "looks right".

## Milestone validation definition

What `qa/milestone-qa.md` will check at the end (the contract this gate enforces):

1. **Per-feature acceptance** — every feature above meets its declared "what to validate".
2. **Workflow-class QA** — **doc-QA** (no canonical code checklist; docs-only milestone):
   - ADR status-gate script passes (every ADR has an accurate `> **Status:**`).
   - `0`-stale-ref grep: zero wiki links to deleted `docs/` paths.
   - Broken-link sweep across `wiki/` is clean.
3. **Regression** — none (M0 is the first milestone; nothing prior to regress).
4. **Quality-bar / root-cause check** — root cause = ambiguity/stale **eliminated**, not papered over:
   one forward roadmap (not three competing surfaces), ADR ledger matches code, archive convention in
   place. **No code-audit slice** (docs-only milestone — spec §6 M0).
5. **No unplanned scope** — anything beyond these five features is recorded with rationale.

## Dependencies & constraints

- Depends on: nothing (first milestone). Builds on the post-deletion repo state (the `docs/` trees
  removed in the current branch — see `git status` deletes).
- Constraints respected: **no code changes** in this milestone (docs-only); superseded docs are
  **archived, not destroyed** (F0.4/F0.5); the governance migration map must stay the single source
  of truth for where moved docs went.

## Applicable hard-stops

Default catalog HS-1..HS-6 (spec §7) in force. For this docs-only milestone:

- **HS-1** (every milestone boundary): operator review gate at close; **no merge without approval**.
- **HS-4** (QA finds symptom-patch / unmet criterion): if the doc-QA gate finds a roadmap still
  competing, an ADR status not matching code, or any surviving stale ref → stop, replan the offending
  feature, re-run its close-out.
- **HS-6** (scope drift mid-milestone): if de-staling uncovers an ADR whose *decision* is wrong (not
  just its status stale) — i.e. a real architecture decision needs revisiting — that is off-plan for a
  docs milestone → stop, surface it, replan before continuing (do not silently rewrite the decision).
- HS-2/HS-3/HS-5 are not expected to trip in a docs-only milestone; if a doc edit somehow requires a
  code/contract change to stay truthful, that is HS-2 (boundary) — stop and report.
