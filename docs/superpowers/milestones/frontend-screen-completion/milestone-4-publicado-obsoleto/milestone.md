# Milestone 4 — Documento Publicado completion + Documento Obsoleto

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md` (§7 M4, §5 inventory rows 11–12, §8 terminal acceptance)
> **Status:** Validator PASS 2026-06-23 (`qa/milestone-qa.md`) — F4.1 + F4.2 closed; awaiting operator HS-1
> **Authored:** 2026-06-23 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which
> features** it contains, **what each feature implements**, and **what gets validated**. It
> contains **no execution steps** — the "how" of each feature lives in that feature's `plan.md`.
> The end-of-milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

After this milestone, a person opening a **published** controlled document (`/documents/:id`)
sees **no "em breve" placeholder for any item the backend can already serve**, and the same screen
correctly renders its **obsolete** state. Concretely:

- The **Cobertura** KPI and coverage card show the **real obligated-audience denominator** from the
  M2 distribution endpoint (`GET /documents/{id}/distribution`) — an honest count of obligated
  readers, with the read-tracking **numerator labelled as parked (ADR-0042)**, never a fabricated %.
- **Baixar PDF** downloads the real artifact via the existing `POST /documents/{id}/export/pdf`
  endpoint — the button is no longer `aria-disabled`.
- **Páginas** and **Tamanho** render from the `DocumentResponse.current_revision_page_count` /
  `current_revision_file_size_bytes` fields that already ship (honest absent state when null).
- Every remaining gap (next-revision date, classification, related documents, comments) is an
  **explicit defer row with a written trigger** in `wiki/backlog/documento-publicado.md`
  — not a silent "em breve" shell. The screen renders each as an honest, non-misleading empty/absent
  state.
- A document whose status is `obsolete` renders the **Documento Obsoleto** variant (banner + status
  presentation) at its route, **sharing the existing `DocumentPublishedPage` component** (no fork),
  matching `design-source/documento-obsoleto/`.

**Quality bar moved:** mission §8 "zero empty no-API shells / zero silent stubs for in-scope items"
for the Publicado + Obsoleto screens. Re-measured by: `grep -n "em breve"` over
`DocumentPublishedPage.tsx` returns **only** lines that are backed by a defer row in the backlog doc
(no in-scope item left as a bare placeholder), and both reviewer agents APPROVE.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F4.1 | `f4.1-publicado-stubs` | Wire the **backend-available** Publicado placeholders to real data, and convert every genuinely-unbacked placeholder into an explicit defer-with-trigger. **Wired (4):** Cobertura KPI + coverage `<aside>` ← `GET /documents/{id}/distribution` (`DistributionSummaryResponse`, denominator only; numerator parked ADR-0042 → obligated count + parked label, no fabricated %); **Baixar PDF** ← `POST /documents/{id}/export/pdf`; **Páginas** ← `DocumentResponse.current_revision_page_count`; **Tamanho** ← `DocumentResponse.current_revision_file_size_bytes` (both already returned by `useDocumentDetailQuery`; nullable → honest absent state). **Deferred (4, no backend):** Próxima revisão, Classificação, Documentos relacionados, Comentários → defer rows in `wiki/backlog/documento-publicado.md`, each with a named trigger. *(Recon corrected a stale-backlog under-scope: page-count/file-size fields already ship, so they are in-scope, not defers.)* | Cobertura card renders the live obligated count (no `—%` fabricated, no hardcoded literal); Baixar PDF enabled + issues the real export (button no longer `aria-disabled`); Páginas + Tamanho render from the real `DocumentResponse` fields (honest absent state when null); `grep "em breve"` over `DocumentPublishedPage.tsx` leaves **only** rows with a matching backlog defer entry + trigger; generated API types consumed directly (no hand-written mapper); `tsc` clean; vitest for the page green; both `frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE. |
| F4.2 | `f4.2-obsoleto-variant` | Make the **obsolete** state a first-class, design-source-matched variant of the **same** `DocumentPublishedPage` component: status-driven banner/presentation, reachable at its route, parity with `design-source/documento-obsoleto/documento-obsoleto.html` + `NOTES.md`. Audit and consolidate the already-present `status === 'obsolete'` branch — no duplicate/forked obsolete page. **Consumer:** the `/documents/:id` route renders the obsolete variant when `DocumentResponse.status === 'obsolete'`. | A document with `status: 'obsolete'` renders the OBSOLETO banner + obsolete status presentation at `/documents/:id` (driven by real `status`, not a prop hack); the obsolete path **reuses** `DocumentPublishedPage` (one component, asserted by no new page file); visual parity with the design-source HTML confirmed by `frontend-screen-reviewer`; vitest covers the obsolete render branch; `tsc` clean; both reviewers APPROVE. |

For each feature, "what to validate" is **objectively checkable** — a rendered value sourced from a
named endpoint, a button no longer disabled, a grep that returns only backlog-backed rows, a passing
test, a reviewer APPROVE. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For this milestone:

1. **Per-feature acceptance** — F4.1 + F4.2 each meet their "what to validate" above, and each feature's
   `spec.md` consumer contract was honored (the page consumes the generated types for
   `GET /documents/{id}/distribution` and `GET /documents/{id}/export/pdf` exactly as the backend
   serves them; obsolete render is driven by the real `DocumentResponse.status`).
2. **Workflow-class QA checklist** — frontend: `wiki/quality/screen-definition-of-done.md` (D2) +
   `wiki/quality/screen-qa-checklist.md`. No backend feature in this milestone (both endpoints already
   exist) → no backend-api checklist run, but **0 backend regressions** must hold (see §3).
3. **Regression** — M0 (routing/truth: 17 routes, no dead stub, tracker accurate), M1 (dashboard live
   data untouched), M2 (distribution endpoint + sacred CD/taxonomy views unchanged — Publicado only
   *reads* the published distribution endpoint), M3 (notifications untouched) all still pass their
   gates. `go build ./...` + `go test ./...` green (no backend code changed); FE `tsc` + the
   notifications/dashboard/distribution test subsets still green.
4. **Quality-bar / root-cause check** — the mission §8 "no silent stub for an in-scope item" bar is
   re-measured by the grep in **Objective**: every surviving "em breve" maps to a backlog defer row
   with a trigger; the two backend-available placeholders are wired to real endpoints (root cause =
   "endpoint existed but unwired", fixed by wiring — **not** masked by a nicer placeholder).
5. **No unplanned scope** — anything implemented beyond F4.1 + F4.2 is recorded with rationale.
   Building any deferred-field backend (next-revision/pages/size/classification/related/comments) or a
   read-tracking numerator is **out of scope** (see Rabbit holes) and would be scope drift.

## Dependencies & constraints

**Appetite:** small — frontend-only, 2 features, no new backend feature, no migrations, no new
endpoints (both consumed endpoints already shipped: `GET /documents/{id}/export/pdf`,
`GET /documents/{id}/distribution`). If a feature appears to need backend work, **stop (HS-2/HS-3)** —
it is either a defer or a missed prerequisite, not in-milestone code.

**Rabbit holes (do not chase):**
- **Read-tracking numerator / % coverage** — parked per ADR-0042; the coverage card shows the
  denominator only. Building who-has-read is a separate parked mission, no consumer here.
- **Backends for the 4 genuinely-unbacked fields** (next-revision date, confidentiality classification,
  related-documents relationship model, display-comments architecture) — no endpoint/field/model exists;
  each is a defer-with-trigger, **not** in-milestone work. *(Page count + file size are NOT here — their
  `DocumentResponse` fields already ship, so F4.1 wires them; deferring them would be a silent stub.)*
- **Redesigning the Publicado layout** — it already uses redesign tokens (finding 12); this milestone
  closes gaps and adds the obsolete variant, it does not restyle.
- **A separate Obsoleto page/route component** — the obsolete state must reuse `DocumentPublishedPage`;
  forking a second page is the anti-goal.

**Quality goals (ranked):** (1) **honesty / truth** — no silent stub, no fabricated number, parked
things labelled parked; (2) **design-source parity** — Publicado gaps closed and Obsoleto matches its
HTML reference; (3) **simplicity / reuse** — one shared component for published + obsolete, generated
types consumed directly, no hand-written mappers.

**Architectural constraints respected:**
- **Frontend-only** — no Go source change; reads stay live; **0 backend regressions**.
- **Contract-first consumption** — consume generated FE types for the two endpoints; no snake→camel
  hand-mapper; `tsc` clean.
- **Defer-with-trigger, not silent stub** — every unbacked gap is an explicit backlog row naming the
  backend field/model that unblocks it.
- **One component for published + obsolete** — obsolete is a status-driven branch, not a fork.

**Risks (named):**
- **R1 — coverage honesty.** The distribution endpoint serves the denominator only; naively rendering
  a "%" would fabricate data. *Mitigation:* coverage card renders obligated count + an explicit
  parked-numerator label (ADR-0042); F4.1 acceptance forbids a fabricated `%`.
- **R2 — half-built obsolete branch.** `DocumentPublishedPage` already has a `status === 'obsolete'`
  branch + banner; risk of duplicate or inconsistent logic. *Mitigation:* F4.2 audits and consolidates
  the existing branch against the design source rather than adding a parallel path; reviewer confirms
  no second page file.
- **R3 — defer creep.** Temptation to half-wire an unbacked field to "look done". *Mitigation:* each
  unbacked field is a backlog defer row with a trigger; the validator greps that every surviving
  "em breve" is backlog-backed.

## Applicable hard-stops

- **HS-1 — milestone boundary.** On validator PASS, the main session flips status and presents the
  operator gate; no M5 and no merge without explicit approval.
- **HS-2 — redesign outside boundary.** If closing a placeholder turns out to need a backend endpoint/
  field/model (shared API/contract change), **stop** — it is a defer-with-trigger or a separate
  full-stack feature, not in-milestone frontend code. Do not symptom-patch by faking data.
- **HS-3 — prerequisite boundary fails.** If `GET /documents/{id}/export/pdf` or
  `GET /documents/{id}/distribution` does not actually serve the contracted shape at runtime (build/
  auth-session/route/contract truth), repair the prerequisite (or convert to a defer) and re-run the
  checkpoint before claiming the feature.
- **HS-4 — validator FAIL.** Open the named fix feature, re-run its lifecycle, re-dispatch the validator.
- **HS-6 — scope drift.** Building a deferred-field backend, a coverage numerator, or a forked Obsoleto
  page is off-plan; surface and replan before continuing.
