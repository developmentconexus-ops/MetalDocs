# Milestone 1 — Dashboard real data (frontend-only)

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Spec approved
> **Authored:** 2026-06-21 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates the
> milestone against *this* document.

## Objective

After M1, the Início (Dashboard) home screen renders **100% live data — zero mocks**. The
operator opening `/` sees real numbers and a real activity feed sourced from existing backend
endpoints, never illustrative placeholders. Concretely: the two remaining mock constants in
`frontend/apps/web/src/features/dashboard/pages/DashboardPage.tsx` — `MOCK_STATS` and
`MOCK_ACTIVITY` — are deleted, and the surfaces they fed render from the live API or a truthful
empty/error state.

**Quality bar moved:** the program's "no `MOCK_`/illustrative data on a routed screen" bar
(per `wiki/quality/screen-definition-of-done.md` D2 criterion 1) goes from **violated** on the
Dashboard to **met**. Re-measured by: `grep -rEn "MOCK_" frontend/apps/web/src/features/dashboard`
= 0, plus runtime proof the cards/feed render live values.

### Operator-locked contract decision (2026-06-21)

The mission named `/documents/stats` + `/iam/kpi` as F1.1's producers. Runtime truth at authoring
time: `/iam/kpi` carries **security** KPIs (locked accounts, MFA, failed logins) — wrong domain for
an approval pulse; and **no endpoint** serves the three cards as originally designed (*Aprovados esta
semana* / *Devolvidos aguardando* / *Tempo médio por decisão*) — those need a time window and
decision-timing data that nothing produces. The operator was shown the gap and **chose to re-scope
the cards to what `/documents/stats` truthfully provides** (counts by status). The "build an
approval-throughput endpoint" path was **refused** for this milestone (see rabbit holes). This
re-scope is the approved F1.1 consumer contract; F1.1's `spec.md` distills it.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F1.1 | `f1.1-dashboard-stats-wire` | Delete `MOCK_STATS`. Wire the hero "§ SEU PULSO" stat cards to live `GET /api/v1/documents/stats` (`by_status` counts), re-scoped to honest status labels (e.g. Aprovados total / Em revisão+devolvidos / Publicados — exact status→label map fixed in `spec.md` after reading the real status enum). The existing live "Aguardando você" pill (approval inbox total) is untouched. | `grep -nE "MOCK_STATS" …/dashboard` = **0**; the stat cards render values pulled from `documents/stats` (consumer shape matches generated `DocumentStatsResponse`); a query-hook test passes against a fixtured stats response; **`frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE on record** (D2). |
| F1.2 | `f1.2-dashboard-activity-wire` | Delete `MOCK_ACTIVITY`. Wire the "§ MURMÚRIOS" activity list to live `GET /api/v1/audit/events` (recent events → who/what/code/relative-time), with a truthful empty state and a graceful 403/error state (audit is `CapAuditRead`-guarded). | `grep -nE "MOCK_ACTIVITY" …/dashboard` = **0**; the activity list renders live audit events (consumer shape matches the generated audit events response); empty + error states render without crashing; query-hook test green; **both reviewer agents APPROVE on record** (D2). |

Order: **F1.1 then F1.2** — independent, but stats is the simpler shape and clears the larger mock
first. "What to validate" for each is objectively checkable (a grep = 0, a contracted response shape,
a passing test, a reviewer APPROVE on record), never "works".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — F1.1 and F1.2 each meet their declared "what to validate", and each
   feature's **consumer contract** (`spec.md`) was honored: the FE consumer shape matches the real
   generated producer types (`DocumentStatsResponse` for F1.1; the audit events response for F1.2).
2. **Workflow-class QA** — `wiki/quality/screen-definition-of-done.md` (D2 gate: both reviewer
   APPROVEs on record) + the runtime functional pass by reference to `wiki/quality/screen-qa-checklist.md`.
3. **Regression** — M0 still passes its gate: single root index route, 0 dead-stub routes, M0 RM
   greps hold; the FE suite holds at the operator-accepted baseline; **backend is untouched**
   (`go build ./...` unchanged — any backend diff is scope drift, see HS-6).
4. **Quality-bar re-measure (root cause, not symptom)** — `grep -rEn "MOCK_" frontend/apps/web/src/features/dashboard`
   = **0**, and the cards/feed render **live** values at runtime (mock deleted at root, not hidden
   behind a flag or a dash placeholder).
5. **No unplanned scope** — only F1.1 + F1.2 are implemented; nothing beyond the two data wires
   (no layout redesign, no new endpoint, no extra screen) without a recorded rationale.

## Dependencies & constraints

- **Depends on:** M0 (truth-reset) passed — honest router + tracker + per-screen DoD in place.
- **Appetite:** small — a wire-only quick win. 2 features, frontend-only, **no backend changes**.
- **Quality goals (ranked):** 1) **truthfulness** (show only data a real producer serves; never
  fabricate a window/average that doesn't exist) > 2) **contract-correctness** (consumer shape ==
  generated producer type) > 3) **simplicity** (smallest frontend change; reuse existing query/client
  patterns).
- **Architectural constraints (validator can fail on these):** frontend-only — **zero** backend/Go
  changes; reads stay live (no caching workarounds); **no new mock/illustrative/"em breve" data**;
  redesign tokens respected (D2); consume the **generated** API types (contract-first consumer, never
  hand-rolled shapes); reviewer separation-of-powers at the gate.
- **Rabbit holes (do NOT chase):**
  - *Building an approval-throughput stats endpoint* (approved-this-week / avg-decision-time) — operator
    refused for M1; it is backend work that D5 sequences after the wire-only milestones. No consumer
    forces it once cards are re-scoped.
  - *Re-introducing time-window / average-time semantics* — no producer exists; would require fabricated
    data → violates quality goal 1.
  - *Redesigning the dashboard layout* — M0 already made it honest; layout is in scope only insofar as
    the two data wires require it.
  - *Re-wiring the already-live "Aguardando você" pill or the pending-approvals list* — already real.

- **Risks:**
  - **R1 — status-enum mapping.** `documents/stats.by_status` keys may not map 1:1 onto the re-scoped
    card labels. *Mitigation:* F1.1 `spec.md` reads the real document status enum first and pins the
    status→label map; if a needed status is absent, fail closed and surface it (do not invent a label).
  - **R2 — audit authz.** `/audit/events` is `CapAuditRead`-guarded; a non-privileged user may get 403.
    *Mitigation:* F1.2 renders a truthful empty/disabled state on 403/empty — never a crash, never a mock.
  - **R3 — stale mission producers.** The mission doc still names `/iam/kpi` (wrong domain) and frames
    F1.1 as a pure wire. *Mitigation (accepted, doc-only):* record the correction in F1.1 evidence and
    leave a `wiki-curator` follow-up to align `mission.md` §M1; not blocking.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | Milestone close: operator review gate before M2 and before any merge — mandatory. |
| HS-2 | A "fix" turns out to require a backend redesign (e.g. F1.1 genuinely needs a new endpoint). Operator pre-resolved via the F1.1 re-scope; if a *new* such need appears, **stop** and report the boundary — do not symptom-patch. |
| HS-3 | A prerequisite fails at runtime: app won't start, no auth session, or `/documents/stats` // `/audit/events` route or contract is broken. Repair the prerequisite, rerun the checkpoint, then resume the feature. |
| HS-4 | `milestone-validator` returns FAIL → open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Any backend/Go change, a third feature, or layout redesign beyond the two wires appears → scope drift; stop and replan. |
