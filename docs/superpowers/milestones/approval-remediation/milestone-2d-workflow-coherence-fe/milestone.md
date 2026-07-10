# Milestone 2d — Workflow Coherence FE + Viewer Contract

> **Program:** approval-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` (§4, Milestone A)
> **Design brief (visual/UX contract):** `docs/superpowers/specs/2026-07-08-single-screen-design-brief.md` (ratified 2026-07-08)
> **Status:** Approved — operator HS-1 gate passed 2026-07-08; executing in a dedicated session
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
| F2d.1 | `f1-viewer-contract` | Instance view DTO gains server-derived `viewer` block (`is_author`, `eligible_for_active_stage` = snapshot ∪ active delegation − SoD exclusion, `via_delegation_from`, `has_signed_active_stage`), computed in the view-read path (`read_service.go:LoadInstanceByDocumentForView`; delegation via `LoadActiveDelegationsFor` plain SELECT — H-PRE-1 safe; SoD via `CheckSoD` on `instance.SubmittedBy`). OpenAPI edit + `oapi-codegen` regen. ADR: viewer-facts contract. Consumer: `deriveWorkspaceMode` (F2d.3). **SoD accuracy:** enforced SoD is author-exclusion + cross-stage double-sign for signoffs only — there is NO reviewer≠approver rule (`sod.go:38-51`); the `viewer` block reflects exactly that, not more. | Real-DB tests per scenario: author, snapshot actor, delegate, already-signed, observer, and author-who-is-also-an-approver (excluded by author-rule). Regen produces no hand-written DTO consumer. Contract lint/build clean. |
| F2d.2 | `f2-verdicts-display-names` | Instance view DTO gains `verdicts[]` (actor id + display name, verdict, reason, timestamp) and display names on stage actors; display-name joins off-tx (H-PRE-1). Verdict history is ALREADY persisted — table `approval_review_verdicts` (migration 0288) carries id/actor/verdict/comment/verdict_at/actor_display_name_snapshot. **No migration.** Scope = a by-instance read (today `LoadStageVerdicts` is by-stage, `postgres_approval_repository.go:1326`) + DTO projection. Consumer: sidebar timeline (F2d.5). | Real-DB test: verdict history shape after `ready` and `request_changes` verdicts; display names resolve for seeded users; no in-tx display lookups (review gate). |
| F2d.3 | `f3-workspace-mode-machine` | ONE pure FE selector `deriveWorkspaceMode(doc, instance, viewer)` → `author-editing \| author-waiting \| author-changes-requested \| reviewing \| approving \| observing \| lifecycle`. **Factor the stage-mode sub-derivation (reviewing/approving/observing, from `stage_kind` + `viewer.eligible_for_active_stage`) as a SUBJECT-AGNOSTIC helper** (survives M3 kernel extraction / templates reuse); the lifecycle sub-derivation (author-editing/-waiting/-changes-requested/lifecycle, keyed on document status) stays subject-specific. `TRANSITION_POLICY` shrinks to document-lifecycle actions only (publish/cancel); stage actions (signoff/verdict) leave it. Replaces the correct-but-duplicated `resolveEditorMode` (`ApprovalCockpitPage.tsx:41-55`) and the broken `signoffOffered` (`useDocumentApprovalArtifact.ts:205-206`). Consumer: single screen (F2d.5) + `DecisionFooter`. | Unit tests: every mode branch incl. delegation, SoD-excluded author, no-instance draft, no-instance non-author. Grep gate: no remaining FE eligibility derivation (`signoffOffered`-style status-only gates) outside the selector. |
| F2d.4 | `f4-instance-react-query` | Approval instance state moves to `useApprovalInstanceQuery` under `QK.approval`. ETag seeding is preserved trivially: `etagCache` is a module-level Map keyed by documentId (`approvalApi.ts:56-58`, consumed `mutationClient.ts:29`), decoupled from React state — the `queryFn` keeps the `etagCache.set` side effect and the If-Match write path is unchanged. Delete imperative useState/useEffect fetch, 1s staleness `setInterval`, dead `isStale`, `refetchInstanceRef` ordering hack, manual `onRefetchInstance` threading. Consumer: single screen (F2d.5). | Test: `QK.approval` invalidation refetches the instance and the mode re-derives; signoff/publish If-Match still resolves from `etagCache`. Grep gate: `setInterval`/`isStale`/`refetchInstanceRef` gone from the approval/document adapters. |
| F2d.5 | `f5-single-screen-shell` | The mode-adaptive single working screen per the design brief: constant DocumentShell, unified right sidebar in ALL modes (incl. `author-editing` — `ArtifactMetaSidebar` composition retired), header mode chip, contextual panels per mode, `DecisionFooter` variant = `stage_kind` + `viewer.eligible_for_active_stage` (F4 contract honored), frozen-content + delegation disclosure in `approving`, changes-requested banner + F6 panel. **Route restructuring (owned here, per pinned Destination):** (a) move `DocumentDetailLayout` + children (`DocumentDetailRoute` index, `distribution`) from `documents/:documentId` to `documents/:documentId/details` in `routes.tsx`; (b) `documents/:documentId/edit` → `<Navigate>` redirect; (c) new working screen mounts at `documents/:documentId` (leaf, no children); (d) fix `DocumentDistributionPage.tsx:95` breadcrumb href → `/documents/${id}/details`; (e) working screen surfaces a discoverable link to `/details` (record view) in the sidebar meta panel; (f) **`?decision=` preselect survives**: the screen accepts `?decision=approve\|reject` seeding the `approving` decision panel's `defaultOptionKey`, `InboxPage.openDecisionFlow` targets `/documents/${id}?decision=...` directly, and the `/approvals/:documentId` redirect forwards `location.search`; (g) **moved to F2d.5b (2026-07-09 amendment)** — lazy inside F2d.5 is ineffective: `DocumentShell.tsx:2` statically imports `MetalDocsEditor` (TipTap) and the read canvas renders DocumentShell, so read-only modes pull the editor chunk regardless; the assertion would green while saving zero bytes. Consumer: end users (all roles); worklist deep links. | Component tests per mode from the brief's §2 matrix. Explicit test: review-kind active stage renders verdict CTAs and NEVER the signature panel; approval-kind renders signature panel only when `viewer.eligible_for_active_stage`. Brief §6 states covered (loading skeleton, instance error, empty). Route tests: `/documents/:id/details` renders `DocumentDetailLayout`; `/documents/:id/edit` redirects preserving `:id`; `?decision=` preselect regression test (M2c F5 evidence precedent — must not silently disappear). ~~Lazy-load assertion~~ moved to F2d.5b (see (g)). |
| F2d.5b | `f5b-pdf-read-canvas` | **DONE 2026-07-09** (base `9f0af980`; commits `339a022f`/`e06520d8` D1 PdfCanvas, `7cdbfaa0`/`6144ceb9` D1 wiring, `cc2274a1` D2 lazy; evidence `f5b-pdf-read-canvas/evidence.md`; independent whole-feature review clean; tsc 0, documents 261/261, zero-backend gate). **RE-SCOPED 2026-07-09 (operator-ratified; design `docs/superpowers/specs/2026-07-09-f5b-pdf-official-view-design.md`).** Original premise (serve frozen PDF to in-approval viewers; sign over rendition bytes) disproven by RED/AS-2 gate — freeze is terminal-only (`decision_service.go:408`), no PDF exists while `under_review`, and a pre-freeze PDF would be falsely official (computed tokens resolve only at freeze). **New scope, FE-only, zero backend:** PDF = official post-approval artifact; in-approval viewing stays on the in-app source canvas (MetalDocs edits in-app — the Veeva rendition pattern solves an uploaded-binary gap this product doesn't have; Qualio-tier pattern applies). D1: `PdfCanvas` renders the official PDF in the workspace canvas for `approved/scheduled/published` (status-keyed; reuses `useDocumentPdfStatus` + existing `/view`, `viewableStatuses` untouched); docx read-only canvas stays for `draft/under_review` read modes. D2: real editor lazy split — `DocumentShell.tsx:2` static `MetalDocsEditor` import → lazy chunk; docx read modes fetch it lazily on mount, lifecycle (PdfCanvas) never fetches it. Signature subject unchanged (source `content_hash`, Part-11-consistent). ADR 0080 amended (in-approval view = source canvas; reopen trigger = customer demand for print-fidelity in-approval preview → Veeva-pattern continuous rendition, researched + shaped). Consumer: all viewers of published docs (official PDF + zero editor bytes), auditors. | FE per-status test: `approved/scheduled/published` → `PdfCanvas` (pending/failed/ready states covered); `under_review` read modes still docx read canvas. **Real** lazy assertion at chunk level: editor absent from the route's static import graph AND not fetched rendering lifecycle mode. Existing suites green (DocumentEditorPage 30/30, workspace). No backend diff (grep/`git status` gate). |
| F2d.6 | `f6-author-comment-replies` | Surface the author's reply-to / resolve on instance comments in `author-waiting` (never edits content) — brief §9.1. **No authz work:** comment writes (`createComment`/`updateComment`→resolve, `handler.go:1083,1116`) gate only on `CapDocumentView` tenant-grade (`handler.go:1201-1206`); the author already holds it on their own document. This is FE-only surfacing of an existing capability — HS-2 not in play. | Real-DB authz test (confirming existing behavior): author can reply/resolve on own document under review; a non-`CapDocumentView` actor cannot. UI test: reply composer visible in `author-waiting`, content editing not. |
| F2d.7 | `f7-cockpit-retirement` | `/approvals/:documentId` becomes a redirect to `/documents/:id` (forwarding `location.search` — `?decision=` preserved, see F2d.5f); `ApprovalCockpitPage`, the hollowed-shell composition (`screenModel` override), and `useDocumentApprovalArtifact`'s cockpit-only reduced model are deleted; worklist deep links target `/documents/:id`. Mechanical cleanup: remaining `/documents/${id}/edit` literal constructors (`DocumentDetailRoute.tsx:101,145`, `NewDocumentWizardPage.tsx:179`, `documentWorkflow.ts:30`) retargeted to `/documents/${id}` (dead paths, not left to bounce through redirects). ADR: single artifact destination. Consumer: worklist navigation + bookmarks. | Route test: old URL redirects (params + query preserved). Grep gates: no `ApprovalCockpitPage` references; no `/edit` path constructors outside the redirect route itself. Worklist link test targets `/documents/:id`. Build clean after deletion. |
| F2d.8 | `f8-close-live-qa` | Milestone close: **UI-driven** live QA on the real stack (browser preview tools; curl only as corroborating backend evidence) covering BOTH route shapes: review+approval route AND approval-only route, full lifecycle draft→submit→verdict(s)→signoff→publish, plus `changes_requested` round-trip, delegation signoff, observer view. Records closure of the M2c deviation. | `qa/live-qa-log.md` with per-step UI evidence (DOM/a11y snapshots). A review-stage screen showing a signature panel = FAIL. Curl-only walkthrough = FAIL (validator forbidden-list). |

Order is binding: F2d.1/F2d.2 (contract) → F2d.3/F2d.4 (FE substrate) → F2d.5 (screen)
→ F2d.5b (PDF read canvas + real lazy) → F2d.6/F2d.7 (completions) → F2d.8 (close). F2d.8's evidence closes the quality-bar claim.

## Destination (P0 — pin before F2d.5; F2d.1–F2d.4 may start now)

Runtime truth (`frontend/apps/web/src/features/documents/routes.tsx`): a document has THREE
surfaces today, not one — `/documents/:id` = `DocumentDetailLayout` (index `DocumentDetailRoute`
record view + `distribution` child); `/documents/:id/edit` = `DocumentEditorRoutePage` (the
DocumentShell working canvas); `/approvals/:documentId` = `ApprovalCockpitPage`. The single-screen
vision collapses the two **DocumentShell working surfaces** (editor + cockpit). The record view
(revisions/distribution/lineage/metadata) is a different altitude.

**DECISION PINNED — operator, 2026-07-08 (Option 1.5, "canonical artifact URL"):**
- **`/documents/:id` = the mode-adaptive working screen** (canonical URL of the artifact —
  clicking the artifact opens the artifact; evidence: Google Docs one-URL modes, GitHub PR
  review-where-you-read, Figma capability-driven surface, Veeva Doc Info hosting workflow tasks).
- **`/documents/:id/details` = the record surface** — `DocumentDetailRoute`/`DocumentDetailLayout`
  survive unchanged, they *move down* one segment; `distribution` stays a child of `details`.
  Record is a different altitude, not absorbed (YAGNI).
- **Redirects:** `/documents/:id/edit` → `/documents/:id`; `/approvals/:documentId` → `/documents/:id`.
  Params preserved; worklist deep links target `/documents/:id`.
- **Surface vs domain (adversarial ratification, chat 2026-07-08):** ONE surface, domains stay
  modular — canvas/autosave own `features/documents`; timeline/DecisionFooter/signature panel own
  `features/approval` and COMPOSE into the screen per mode. Signature = ceremony WITHIN the screen
  (re-auth modal, Qualio/Veeva pattern), never a separate URL. Reopen trigger: an EXTERNAL signer
  persona (no app account) would justify a DocuSign-style standalone ceremony page — not before.
- **F2d.5 addendum (amended 2026-07-09, re-amended after F2d.5b re-scope):** editor bundle must not
  ship in the route's static graph — relocated to **F2d.5b** D2 (DocumentShell's static TipTap import
  made in-F2d.5 lazy a no-op). Post-re-scope accounting: docx read modes lazy-fetch the chunk on
  mount; lifecycle mode (official PDF canvas) never fetches it.

Governing-spec §1.2-B and design brief §1/§5 updated with this pin (2026-07-08).

**Post-pin adversarial re-validation (2026-07-08, subagent):** APPROVE-WITH-FIXES — topology
consistent across all three docs; findings folded back into the Features table: P0 `?decision=`
preselect migration (→ F2d.5f + F2d.7 redirect query-forwarding), P1 route-restructuring
ownership itemized (→ F2d.5a–c), P1 `DocumentDistributionPage` breadcrumb (→ F2d.5d), P1
`/details` discoverability affordance (→ F2d.5e), P1 lazy-load acceptance (→ F2d.5g), P2
`/edit` literal cleanup (→ F2d.7 grep gate).

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
  - No route/stage schema changes. F2d.2 needs NO migration — verdict history already persists in
    `approval_review_verdicts` (migration 0288); scope is a by-instance read + DTO projection.
  - All server state via react-query under existing `QK` conventions; no new imperative fetch state.
  - Wine tokens only (`src/styles/tokens.css`); slate palette not extended; PT-BR copy; WCAG AA.
- Risks (owner: this milestone):
  - **Destination decision (P0)** open until pinned by operator → blocks F2d.5 only; F2d.1–F2d.4
    proceed. If option (2) is chosen, scope grows by one feature (absorb the record surface) — that
    addition is recorded here, not silent scope creep.
  - `DocumentEditorPage` (661 lines) restructure regression risk → mitigated by F2d.5 per-mode
    component tests + F2d.8 dual-route live QA + M2c regression line.
  - Editor suggest-mode reuse assumption (review canvas) breaks → HS-3 prerequisite repair.
  - (Retired risks from validation 2026-07-08: verdict-history migration — history already persists,
    no migration; F2d.6 authz gap — comment writes already gated on `CapDocumentView` the author
    holds, FE-only; ETag seam — module-level Map, trivially preserved in the queryFn.)

## Applicable hard-stops

| ID | Trip condition in this milestone |
|----|----------------------------------|
| HS-1 | Milestone boundary: validator PASS presented to operator; no M3 start, no merge/push without approval |
| HS-2 | F2d.6 authz redesign beyond a capability addition; any non-additive persistence need in F2d.2; anything requiring kernel write-path change |
| HS-3 | Prerequisite failure: API not runnable, auth/session broken, contract/generated drift, review-canvas suggest mode broken |
| HS-4 | Validator FAIL → named fix feature, re-run lifecycle, re-dispatch validator |
| HS-6 | Scope drift beyond the Features table / rabbit-hole list — stop and replan |

## Amendments

### 2026-07-09 — HS-2 cleared: F2d.2 scope grows to the display-name invariant (operator "A", global maximum)

The milestone body says "F2d.2 needs NO migration" (lines 132-133, 143-144). During F2d.2 speccing the
`actor_display_name_snapshot` name-source question surfaced a **no-fallback** design fork. Operator answer
(verbatim principle): *"A professional System … with no redundancies, does not need a fallback (unless
needs a fail-safe) … analyse … the root cause … we don't care about refactoring."* Presented the A/B
boundary decision (A = global-maximum root-cause fix now; B = read-only, defer); operator chose **A**.

This is a **conscious HS-2 clearance** ("any non-additive persistence need in F2d.2"): the persistence
change below is authorized, overriding the "NO migration" rows for F2d.2 only.

- **In scope now (F2d.2):**
  - **D1 — DB-enforces-invariants migration.** Backfill any NULL/`''` `actor_display_name_snapshot`
    from `iam_users.display_name`, then `SET NOT NULL` + `CHECK (<> '')` on the snapshot column of
    **both** `approval_review_verdicts` and `approval_signoffs`. Insert bindings bind the snapshot
    unconditionally (remove the `if name != ""` omission). No decision-logic change to the write path.
  - **D2 — remove the read fallbacks.** Drop the `coalesce(…, '')` in verdict/signoff loads; collapse
    the signoff mapper's snapshot-else-live branch to **snapshot-only**. Name in history is the value
    cast at the moment of the signed action, immutable (eQMS audit truth).
- **Still deferred:** D3 (`on_behalf_of` delegator display name in verdict history) — needs a new
  snapshot column; bounded to a later feature.
- **Governing ADR:** `wiki/decisions/0079-verdict-history-contract.md` (verdict-history contract +
  immutable actor-name snapshot, no read fallback, DB-enforced).

All other milestone rows stand. The "NO migration" statement remains true for every feature **except
F2d.2**.
