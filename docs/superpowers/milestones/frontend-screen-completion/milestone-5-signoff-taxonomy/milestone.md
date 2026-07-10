# Milestone 5 — Detalhe Signoff + Taxonomy Admin restyle (net-new / polish — last)

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md` (§7 M5, §5 inventory rows 13–14, §8 terminal acceptance)
> **Status:** FEATURES CLOSED — awaiting milestone-validator + HS-1 (2026-07-10). Spec authored 2026-06-23. **F5.1** closed with both-reviewer APPROVE (`f5.1-signoff-detail/evidence.md`). **F5.2** closed **verify-only** (`f5.2-taxonomy-restyle/evidence.md`): the restyle was already delivered by `2a371d60 (FE-14, 2026-07-02)` — this milestone's "11 inline `style=`" premise predated FE-14; current page is fully tokenized (grep=0), tsc clean, vitest 23/23, rendered GREEN, both reviewers APPROVE. No new F5.2 diff written (fabricating one would invent scope on an already-correct base).
> **Sequencing:** the **last** milestone (D5). Net-new screen + styling polish run last so they cannot regress M0–M4.

> This file is a **spec**, authored up front. It says **what** this milestone is, **which
> features** it contains, **what each feature implements**, and **what gets validated**. It
> contains **no execution steps** — the "how" of each feature lives in that feature's `plan.md`.
> The end-of-milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

After this milestone, the **two remaining unfinished screens** in the routed app are
production-complete to the mission §8 bar, and **no in-scope screen is left unbuilt or
off-design-system**:

- A reviewer opening an approval sign-off from the inbox reaches a real **Detalhe Signoff**
  screen — the A4 document/diff view, the approval-flow timeline panel, and the decision
  (approve / reject sign-off) form — rendering **live data** from the existing approval APIs
  (`GET /documents/{id}/approval-instance` or `GET /approval/instances/{instance_id}`) and
  recording a decision through the existing sign-off endpoint
  (`POST /documents/{id}/signoff` / `POST /approval/instances/{instance_id}/stages/{stage_id}/signoffs`).
  No mock data, no dead-end. Matches `design-source/detalhe-signoff/detalhe-signoff.html`
  (the `SignoffDetail` reference in `priority-screens.jsx`).
- `TaxonomyAdminPage` renders identically in behavior but is styled **entirely through the
  redesign design-system tokens** — the 11 inline `style=` occurrences are gone, replaced by
  token-backed classes/components. No behavior change, no contract change.

**Quality bar moved:** mission §8 "every in-scope screen production-complete: real API data,
redesign tokens, **both** reviewer agents APPROVE on record, tests green — zero net-new screen
left unbuilt, zero off-design-system screen remaining." Re-measured by:
`grep -nE "style=\{\{|style=\"" frontend/apps/web/src/features/taxonomy/TaxonomyAdminPage.tsx`
returns **0**; a `detalhe-signoff` route renders the screen against the real approval API; both
`frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE on both screens.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F5.1 | `f5.1-signoff-detail` | **Assemble** the **Detalhe Signoff** screen at a new route from **existing tested parts** (recon-corrected — see README HS-6 2026-06-23; the decision surface already exists, it is not built from scratch), wired to the **already-shipped** approval/sign-off APIs (no new backend). **Approach α:** a new route `/approvals/:documentId` reached from the approval inbox (`InboxStack` approve/reject **navigates here, replacing the in-inbox `SignoffDialog` modal**). Two columns per `design-source/detalhe-signoff/detalhe-signoff.html`: (1) **left — A4 document body**: embed the rendered PDF from `GET /documents/{id}/view` (`pdf_url`/`signed_url`; honest loading when `pdf_status` pending) + tab strip (Documento; **Comentários live** ← `GET /documents/{id}/comments`; **Mudanças vs vX diff = deferred-with-trigger** — no diff backend, honest absent state + backlog row, no faked diff); (2) **right — decision surface**: **mount the orphan `ControlledDocumentDetailPanel`** (already exported + unit-tested, mounted nowhere today) which itself reuses `ApprovalTimelinePanel` + `SignoffDialog` and carries the per-state policy / integrity / Assinar-Cancelar-Publicar actions. **Data in:** `getActiveDocumentContext(cd_id)` (→ documentId, instanceId, content_hash, revision) + `getInstance(documentId)` → `ApprovalInstance`. **Decision out:** `signoff(documentId, …)` with `If-Match` via the reused panel. Generated FE types consumed directly — no hand-written mapper. | The `/approvals/:documentId` route renders the A4 PDF embed + Comentários (live) + the mounted `ControlledDocumentDetailPanel` populated from a **real** approval-instance query (no `MOCK_`/illustrative literal); inbox approve/reject **navigates to the route** (modal path removed for that flow); recording a decision issues the real `signoff` mutation and reflects the outcome; **reuses `ControlledDocumentDetailPanel` (hence `ApprovalTimelinePanel` + `SignoffDialog`)** — asserted: no forked timeline/decision/sign-off component, no second decision form; the diff tab has no faked data (deferred-with-trigger backlog row); visual parity with `detalhe-signoff.html` confirmed by `frontend-screen-reviewer`; vitest covers the load + decision-submit branches; `tsc` clean; both `frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE. |
| F5.2 | `f5.2-taxonomy-restyle` | Convert `frontend/apps/web/src/features/taxonomy/TaxonomyAdminPage.tsx` from inline `style=` to **redesign design-system tokens** (token-backed classes / existing primitives), with **zero behavior change** and **zero contract change**. **Consumer:** the existing taxonomy admin route (`TaxonomyAdminRoutePage`) renders the restyled page identically. | `grep -nE "style=\{\{\|style=\""` over `TaxonomyAdminPage.tsx` = **0** (all 11 current inline-style sites removed); styling uses redesign tokens/primitives only (no raw hex/px literals introduced outside tokens); the existing taxonomy test(s) still pass unchanged (no behavior regression); the taxonomy API contract (`taxonomy/api/taxonomy.ts`) is untouched; `tsc` clean; both reviewers APPROVE. |

For each feature, "what to validate" is **objectively checkable** — a rendered value sourced from a
named endpoint, a mutation fired, a `grep` that returns 0, a reuse assertion, a passing test, a
reviewer APPROVE. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For this milestone:

1. **Per-feature acceptance** — F5.1 + F5.2 each meet their "what to validate" above, and each
   feature's `spec.md` consumer contract was honored: F5.1 consumes the generated types for the
   approval-instance + sign-off endpoints exactly as the backend serves them and reuses the existing
   `ApprovalTimelinePanel`/`SignoffDialog` rather than forking; F5.2 changes only presentation, not
   the taxonomy API contract or behavior.
2. **Workflow-class QA checklist** — frontend: `wiki/quality/screen-definition-of-done.md` (D2) +
   `wiki/quality/screen-qa-checklist.md`. **No backend feature** in this milestone (all approval/
   sign-off endpoints already ship) → no backend-api checklist run, but **0 backend regressions**
   must hold (§ Dependencies).
3. **Regression** — M0 (routing/truth: single index route, no dead stub, tracker accurate),
   M1 (dashboard live data), M2 (distribution endpoint + sacred CD/taxonomy views unchanged),
   M3 (notifications), M4 (Publicado + Obsoleto) all still pass their gates. `go build ./...` +
   `go test ./...` green (no backend code changed); FE `tsc` + the approval/inbox + taxonomy test
   subsets still green; `npm run build` clean.
4. **Quality-bar / root-cause check** — the mission §8 bar is re-measured by the greps in
   **Objective**: the Detalhe Signoff route renders against the real approval API (root cause =
   "screen never built", fixed by building it on the existing endpoints — **not** masked by mock
   data); `TaxonomyAdminPage` inline-style count = 0 (root cause = "off-design-system inline styles",
   fixed by adopting tokens — **not** masked by leaving styles and renaming).
5. **No unplanned scope** — anything implemented beyond F5.1 + F5.2 is recorded with rationale.
   Building any **new backend** approval/sign-off endpoint, redesigning the approval *flow* /
   state machine, or restyling any screen other than Taxonomy Admin is **out of scope** (see Rabbit
   holes) and would be scope drift.

## Dependencies & constraints

**Appetite:** small — frontend-only, 2 features, **no new backend**, no migrations, no new endpoints.
F5.1 wires to endpoints that already ship (`/documents/{id}/approval-instance`,
`/approval/instances/{instance_id}`, `/documents/{id}/signoff`, stage-signoffs, `/approval/inbox`);
F5.2 is pure presentation. If a feature appears to need backend work, **stop (HS-2/HS-3)** — it is a
missed prerequisite or a defer, not in-milestone code.

**Rabbit holes (do not chase):**
- **New approval/sign-off backend or flow redesign** — the approval state machine, route-admin, and
  sign-off semantics are settled and shipped (Grade-A backend); F5.1 *consumes* them, it does not
  touch them. Changing approval semantics trips HS-2.
- **Re-reviewing or restyling already-DONE screens** — Inbox (`InboxPage`), route-admin, and the
  approval components (`ApprovalTimelinePanel`, `SignoffDialog`) are shipped; F5.1 **reuses** them.
  Only `TaxonomyAdminPage` is in scope for restyling (mission §5 row 14).
- **Redesigning the design system / tokens** — tokens + primitives are settled (mission Non-Goals);
  F5.2 *consumes* tokens, it does not add or change them. A needed token change trips HS-2.
- **Authoring a NOTES.md / re-deriving design decisions for detalhe-signoff** — the design source
  ships only the HTML (no NOTES.md, unlike documento-obsoleto). Parity is judged against the HTML +
  the `SignoffDetail` reference in `priority-screens.jsx`, not a decisions doc; do not invent scope to
  fill the gap (see R1).

**Quality goals (ranked):** (1) **honesty / real-API truth** — Detalhe Signoff renders and submits
against the real approval API, no mock; (2) **reuse / simplicity** — F5.1 reuses the existing
approval components and generated types (no forked timeline/decision UI, no hand mapper); F5.2 adopts
existing tokens (no new styling system); (3) **design-source parity** — Detalhe Signoff matches its
HTML reference and Taxonomy matches the redesign token look.

**Architectural constraints respected:**
- **Frontend-only** — no Go source change; reads/decisions go through existing endpoints; **0 backend
  regressions**.
- **Contract-first consumption** — consume generated FE types for the approval/sign-off endpoints; no
  snake→camel hand-mapper; `tsc` clean.
- **Reuse over fork** — F5.1 mounts the existing **orphan** `ControlledDocumentDetailPanel` (which
  reuses `ApprovalTimelinePanel` + `SignoffDialog`); no second timeline, decision, or sign-off
  component, no re-implemented decision form. F5.2 reuses tokens/primitives; no new styling layer.
- **No behavior/contract change in F5.2** — restyle is presentation-only; the taxonomy API and the
  page's behavior are byte-stable except styling.
- **Both-reviewer gate (D2)** — each screen closes only on `frontend-screen-reviewer` **and**
  `frontend-code-reviewer` APPROVE, tests green.

**Risks (named):**
- **R1 — no NOTES.md for detalhe-signoff.** Unlike documento-obsoleto, the design source is HTML-only.
  Parity intent (exact spacing, copy, states) is less pinned. *Mitigation:* F5.1's `spec.md` interview
  (Phase 3) pins the route, the exact endpoint binding, and the parity bar against the HTML +
  `SignoffDetail`; `frontend-screen-reviewer` judges visual parity against the HTML reference. If a
  genuine design decision is missing, it is interviewed with the operator, not invented.
- **R2 — approval endpoint binding ambiguity.** Two read paths exist (`/documents/{id}/approval-instance`
  vs `/approval/instances/{instance_id}`) and two sign-off paths (document-level vs stage-level). Naive
  choice could mis-wire the decision form. *Mitigation:* F5.1's `spec.md` consumer contract names the
  exact endpoint(s) the inbox→detail flow uses (verified against `InboxPage`/`approval/queries`)
  before code; HS-3 if the chosen endpoint doesn't serve the contracted shape at runtime.
- **R3 — restyle behavior regression.** Converting 11 inline-style sites risks altering layout/behavior.
  *Mitigation:* F5.2 acceptance requires the existing taxonomy test(s) to pass **unchanged** and the
  API contract untouched; reviewer confirms presentation-only diff.
- **R4 — orphan panel visual/UX fit (F5.1).** `ControlledDocumentDetailPanel` was built for a generic
  controlled-document detail context, not styled to the `detalhe-signoff.html` cockpit layout; mounting
  it as-is in the right column may not match the design-source sidebar look, and its `window.prompt`
  cancel-reason + `navigator.clipboard` integrity affordances may be heavier than the mock implies.
  *Mitigation:* F5.1's `spec.md` pins which panel sections are in-scope for the cockpit and whether any
  token-level styling adaptation (no behavior change) is needed for parity; `frontend-screen-reviewer`
  judges the assembled result against `detalhe-signoff.html`. If parity demands forking the panel,
  that is HS-2 (stop + replan), not a silent rewrite.

## Applicable hard-stops

- **HS-1 — milestone boundary.** On validator PASS, the main session flips status and presents the
  operator gate. M5 is the **last** milestone — on HS-1 approval the program proceeds to **terminal
  acceptance** (mission §8: per-screen re-audit fan-out + `mission-validator`), not to another
  milestone. No merge without explicit approval.
- **HS-2 — redesign outside boundary.** If building Detalhe Signoff or restyling Taxonomy turns out to
  need a backend endpoint/field, an approval-flow change, or a shared-primitive/token change, **stop** —
  it is a separate full-stack feature or a defer-with-trigger, not in-milestone frontend code. Do not
  symptom-patch by faking data or forking a primitive.
- **HS-3 — prerequisite boundary fails.** If the chosen approval-instance / sign-off endpoint does not
  serve the contracted shape at runtime (build/auth-session/route/contract truth), repair the
  prerequisite (or re-bind to the correct endpoint) and re-run the checkpoint before claiming F5.1.
- **HS-4 — validator FAIL.** Open the named fix feature, re-run its lifecycle, re-dispatch the validator.
- **HS-6 — scope drift.** Building new approval backend, redesigning the approval flow, restyling any
  screen other than Taxonomy Admin, or changing design tokens is off-plan; surface and replan before
  continuing.
