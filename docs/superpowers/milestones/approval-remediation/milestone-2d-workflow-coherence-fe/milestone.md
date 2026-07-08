# Milestone 2d — Workflow Coherence FE + Viewer Contract

> **Program:** approval-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` (§4, Milestone A)
> **Design brief (visual/UX contract):** `docs/superpowers/specs/2026-07-08-single-screen-design-brief.md` (ratified 2026-07-08)
> **Status:** Spec (drafting)
> **Authored:** 2026-07-08 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

After this milestone, any actor opening a document sees ONE screen (`/documents/:id`)
that adapts to what they may actually do — and what they may do is a **server-derived
fact**, never a client guess:

- A reviewer on a review stage sees verdict CTAs and can suggest; the signature panel
  never appears on a review stage — the M2c live-QA defect class
  (`412 precondition.content_hash_mismatch` from offering signoff on an unfrozen stage)
  is structurally impossible, not patched.
- An approver on an approval stage sees the frozen content and the signature panel,
  including when eligible only via delegation.
- The author edits in draft, replies to review comments while waiting, and gets the
  request-changes panel on `changes_requested`.
- Everyone else observes: read-only + timeline, oversight actions only per capability.

**Quality bar moved:** closes the M2c recorded deviation (DecisionFooter violated the F4
`stage_kind` contract) at the root: eligibility truth moves server-side (`viewer` block),
the FE collapses 4 parallel action-derivations into one pure selector, and the duplicate
destination (`ApprovalCockpitPage`) is deleted. Re-measured by the C5 gate below.

**Appetite:** FE + contract + thin backend *read-path* only. No kernel write-path
changes, no route/stage schema changes. 8 features cap.

**Rabbit holes (do not chase):**
- Approval kernel extraction / templates unification — that is M3 (governing spec §5).
- Actor selectors / route-builder changes — that is M4 (§6).
- Worklist (`/approvals`) redesign — shipped C3/C5; only deep-link targets change here.
- Editor/eigenpal internals (autosave, buffer, docx render) — reuse as-is.
- Parallel stages, dynamic assignment rules — refused non-goals (§7 of governing spec).

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F2d.1 | `f1-viewer-contract` | Instance view DTO gains server-derived `viewer` block (`is_author`, `eligible_for_active_stage` = snapshot ∪ active delegation − SoD exclusion, `via_delegation_from`, `has_signed_active_stage`), computed in the view-read path. OpenAPI edit + `oapi-codegen` regen. ADR: viewer-facts contract. Consumer: `deriveWorkspaceMode` (F2d.3). | Real-DB tests per scenario: author, snapshot actor, delegate, already-signed, observer, SoD-excluded author-who-is-also-approver. Regen produces no hand-written DTO consumer. Contract lint/build clean. |
| F2d.2 | `f2-verdicts-display-names` | Instance view DTO gains `verdicts[]` (actor id + display name, verdict, reason, timestamp) and display names on stage actors; display-name joins off-tx (H-PRE-1). Verify at spec time whether verdict history is already persisted; if not, an **additive** migration is in-scope for this feature. Consumer: sidebar timeline (F2d.5). | Real-DB test: verdict history shape after `ready` and `request_changes` verdicts; display names resolve for seeded users; no in-tx display lookups (review gate). |
| F2d.3 | `f3-workspace-mode-machine` | ONE pure FE selector `deriveWorkspaceMode(doc, instance, viewer)` → `author-editing \| author-waiting \| author-changes-requested \| reviewing \| approving \| observing \| lifecycle`. `TRANSITION_POLICY` shrinks to document-lifecycle actions only (publish/cancel); stage actions (signoff/verdict) leave it. Consumer: single screen (F2d.5) + `DecisionFooter`. | Unit tests: every mode branch incl. delegation, SoD-excluded author, no-instance draft, no-instance non-author. Grep gate: no remaining FE eligibility derivation (`signoffOffered`-style status-only gates) outside the selector. |
| F2d.4 | `f4-instance-react-query` | Approval instance state moves to `useApprovalInstanceQuery` under `QK.approval` (ETag seeding preserved in the fetcher). Delete imperative useState/useEffect fetch, 1s staleness `setInterval`, dead `isStale`, `refetchInstanceRef` ordering hack, manual `onRefetchInstance` threading. Consumer: single screen (F2d.5). | Test: `QK.approval` invalidation refetches the instance and the mode re-derives. Grep gate: `setInterval`/`isStale`/`refetchInstanceRef` gone from the approval/document adapters. |
| F2d.5 | `f5-single-screen-shell` | `DocumentEditorPage` becomes the mode-adaptive single screen per the design brief: constant DocumentShell, unified right sidebar in ALL modes (incl. `author-editing` — `ArtifactMetaSidebar` composition retired), header mode chip, contextual panels per mode, `DecisionFooter` variant = `stage_kind` + `viewer.eligible_for_active_stage` (F4 contract honored), frozen-content + delegation disclosure in `approving`, changes-requested banner + F6 panel. Consumer: end users (all roles); worklist deep links. | Component tests per mode from the brief's §2 matrix. Explicit test: review-kind active stage renders verdict CTAs and NEVER the signature panel; approval-kind renders signature panel only when `viewer.eligible_for_active_stage`. Brief §6 states covered (loading skeleton, instance error, empty). |
| F2d.6 | `f6-author-comment-replies` | Author replies to / resolves instance comments in `author-waiting` (never edits content) — brief §9.1. Verify the authz surface for author comment writes against the existing instance-comments capability model; if a capability gap exists it is a contract item in this feature (tier-1 + tier-2 per ADR 0022), not a client workaround. | Real-DB authz test: author can reply/resolve on own document under review; non-participant cannot. UI test: reply composer visible in `author-waiting`, content editing not. |
| F2d.7 | `f7-cockpit-retirement` | `/approvals/:documentId` becomes a redirect to `/documents/:id`; `ApprovalCockpitPage`, the hollowed-shell composition (`screenModel` override), and `useDocumentApprovalArtifact`'s cockpit-only reduced model are deleted; worklist deep links target `/documents/:id`. ADR: single artifact destination. Consumer: worklist navigation + bookmarks. | Route test: old URL redirects (params preserved). Grep gate: no `ApprovalCockpitPage` references. Worklist link test targets `/documents/:id`. Build clean after deletion. |
| F2d.8 | `f8-close-live-qa` | Milestone close: **UI-driven** live QA on the real stack (browser preview tools; curl only as corroborating backend evidence) covering BOTH route shapes: review+approval route AND approval-only route, full lifecycle draft→submit→verdict(s)→signoff→publish, plus `changes_requested` round-trip, delegation signoff, observer view. Records closure of the M2c deviation. | `qa/live-qa-log.md` with per-step UI evidence (DOM/a11y snapshots). A review-stage screen showing a signature panel = FAIL. Curl-only walkthrough = FAIL (validator forbidden-list). |

Order is binding: F2d.1/F2d.2 (contract) → F2d.3/F2d.4 (FE substrate) → F2d.5 (screen)
→ F2d.6/F2d.7 (completions) → F2d.8 (close). F2d.8's evidence closes the quality-bar claim.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate"; each
   feature's consumer contract (`spec.md`) honored (producer matches consumer).
2. **Workflow-class QA checklist** — `wiki/quality/qa-operating-system.md` + the screen/frontend
   checklist under `wiki/quality/` applicable to FE screen work; contract checklist for F2d.1/F2d.2.
3. **Regression** — M2b kernel gates still pass (`go build ./...`, approval real-DB suites); M2c
   C3/C5 worklist + author-panel behavior still pass (deep links updated, not broken); frontend
   suite green (`make test` scope for `frontend/apps/web`).
4. **Quality-bar / root-cause check** — the M2c deviation is re-measured CLOSED at the root:
   (a) F2d.3 grep gate shows no FE eligibility derivation outside the selector; (b) F2d.5 test
   proves review-kind stage cannot render the signature panel; (c) F2d.8 live QA shows the
   review→approval lifecycle UI-driven without a 412. Symptom-patch (e.g. re-gating the old
   `signoffOffered` with one more boolean) = FAIL.
5. **No unplanned scope** — anything beyond the Features table is recorded with rationale;
   rabbit-hole list above is the drift baseline.

## Dependencies & constraints

- Depends on: M2b (kernel, passed), M2c (worklist/sidebar components, validator PASS with
  recorded deviation — this milestone closes it). Governing spec + design brief ratified.
- **Quality goals (ranked):** 1. Eligibility-truth correctness (server-derived, FE renders facts)
  > 2. Contract fidelity (OpenAPI-first, generated DTOs only, F4 stage_kind contract honored)
  > 3. UX coherence per the design brief (single screen, constant shell).
- Architectural constraints (validator can fail on each):
  - Contract-first: routes/DTOs change ONLY via `api/openapi` + `oapi-codegen` regen (no
    hand-edited generated code, no hand-written `body.data.X` consumers — ADR 0035 discipline).
  - H-PRE-1: display-name/authz-recording reads never inside a lock-holding tx; viewer/verdict
    display joins off-tx or in the plain read tx of the view path.
  - No kernel write-path changes (freeze/signoff/verdict/submit services untouched), EXCEPT a
    capability addition for F2d.6 if the authz verification finds a gap (ADR 0022 two-tier).
  - No route/stage schema changes; F2d.2 may add an **additive** verdict-history persistence
    migration only if history is not already persisted (verified at spec time).
  - All server state via react-query under existing `QK` conventions; no new imperative fetch state.
  - Wine tokens only (`src/styles/tokens.css`); slate palette not extended; PT-BR copy; WCAG AA.
- Risks (owner: this milestone):
  - Verdict-history persistence unknown → verified at F2d.2 spec time; additive migration
    in-scope; anything non-additive → HS-2.
  - `DocumentEditorPage` (661 lines) restructure regression risk → mitigated by F2d.5 per-mode
    component tests + F2d.8 dual-route live QA + M2c regression line.
  - F2d.6 authz gap could cross the module boundary → HS-2 stop, surface, prerequisite plan.
  - Editor suggest-mode reuse assumption (review canvas) breaks → HS-3 prerequisite repair.

## Applicable hard-stops

| ID | Trip condition in this milestone |
|----|----------------------------------|
| HS-1 | Milestone boundary: validator PASS presented to operator; no M3 start, no merge/push without approval |
| HS-2 | F2d.6 authz redesign beyond a capability addition; any non-additive persistence need in F2d.2; anything requiring kernel write-path change |
| HS-3 | Prerequisite failure: API not runnable, auth/session broken, contract/generated drift, review-canvas suggest mode broken |
| HS-4 | Validator FAIL → named fix feature, re-run lifecycle, re-dispatch validator |
| HS-6 | Scope drift beyond the Features table / rabbit-hole list — stop and replan |
