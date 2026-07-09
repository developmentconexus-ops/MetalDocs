# Feature F2d.5 — Plan (decomposed into ordered TDD slices)

> Input: `spec.md` (approved pre-code). Large feature → 3 ordered slices, each: failing test → green →
> independent review → evidence-fragment. All fold into ONE `evidence.md` at feature close.

## Slice order & rationale
S1 (footer variant) is isolated + unblocks the screen's footer. S2 (the screen) is the bulk. S3 (route flip
+ deep-links) is last so the new screen exists before routes point at it. Each slice is independently green.

---

### S1 — `DecisionFooter` three-way variant (isolated component)
**Problem:** `DecisionFooter.tsx:224` is binary (`decision != null ? ApprovalModeFooter : ReviewModeFooter`)
→ an ineligible observer on an active stage wrongly gets verdict CTAs.
**Target:** variant driven by `stage_kind` + `viewer.eligible_for_active_stage` (or the WorkspaceMode) —
`approving`→signature; `reviewing`+eligible→verdict CTAs; else→render nothing.
- **TDD:** extend `DecisionFooter.test.tsx` — add: review-stage+eligible→verdict CTAs (no signature);
  approval-stage+eligible→signature (no CTAs); active-stage+ineligible→neither; then implement the variant.
- Wire the footer's variant input from the caller (pass `mode` or `{stageKind, eligible}`); keep
  `onRefetchInstance` prop (dies in F2d.7).
- Files: `DecisionFooter.tsx`, `DecisionFooter.test.tsx` (+ `ApprovalSidebar.tsx` if it must thread the new prop).
- Gate: `vitest run …/sidebar/__tests__/DecisionFooter.test.tsx`; `tsc --noEmit`.

### S2 — `DocumentWorkspacePage` (the mode-adaptive screen) — the bulk
**Target:** new `features/documents/pages/DocumentWorkspacePage.tsx` (thin owner) + lazy `EditorCanvas`
extraction from `DocumentEditorPage`'s body. Composes: constant `DocumentShell`, unified right sidebar
(meta + route-preview panels [ArtifactMetaSidebar composition retired] + accountability timeline +
`DecisionFooter`), header mode chip, per-mode canvas/footer/panel per §2, `changes_requested` banner + F6
panel, frozen-content + delegation disclosure in `approving`, `?decision=` seed, `React.lazy`+`Suspense`
for the editor chunk in editing modes only.
- Derivation: `useDocumentDetailQuery` + `useApprovalInstanceQuery` (F2d.4) + `deriveWorkspaceMode` (F2d.3).
  Reuse `buildDocumentSignoffDecision` for the approving decision (F2d.4 adapter already models it).
- **TDD:** `DocumentWorkspacePage.test.tsx` — one test per §2 mode (correct canvas + footer variant +
  contextual panel), §6 states (loading skeleton, instance error keeps canvas, empty/edge), `changes_requested`
  banner, `?decision=` preselect regression, and a lazy-load assertion (editor chunk dynamically imported;
  absent in `observing`/`approving` initial render). Write failing tests first per branch, implement to green.
- **Sub-structure (keep units small):** `EditorCanvas` (lazy, editing modes), `WorkspaceSidebar` (panel
  stack: meta / route-preview / timeline / footer), `ModeChip` (header), `ApprovingDisclosure` (frozen +
  delegation). Each independently testable.
- Files (new): `DocumentWorkspacePage.tsx`, `EditorCanvas.tsx` (extracted), `WorkspaceSidebar.tsx`,
  `ModeChip.tsx`, `ApprovingDisclosure.tsx`, + their tests. (edit) `DocumentEditorPage.tsx` (body →
  `EditorCanvas`), `ArtifactMetaSidebar` consumers.
- Gate: `vitest run …/DocumentWorkspacePage.test.tsx` + sub-component suites; `tsc --noEmit`.
- **Note:** this slice is large — may sub-decompose S2 into S2a (shell + sidebar + mode chip + read modes)
  and S2b (editing modes via lazy EditorCanvas + changes_requested/F6 + approving disclosure/decision) if the
  single-slice diff exceeds reviewable size. Decide at execution.

### S3 — Route flip + deep-links + breadcrumb
**Target:** point routing at the new screen; retire the cockpit ROUTE (not the file — F2d.7).
- `features/documents/routes.tsx`: move `DocumentDetailLayout` subtree `documents/:documentId` →
  `documents/:documentId/details`; new `documents/:documentId` leaf → `DocumentWorkspacePage`;
  `documents/:documentId/edit` → `<Navigate to="/documents/:id" replace>` (preserve `:id`).
- `features/approval/routes.tsx`: `approvals/:documentId` → redirect component forwarding `location.search`
  to `/documents/:id` (a tiny `<Navigate>` wrapper reading `useParams`+`useLocation`).
- `InboxPage.tsx:110`: `navigate('/documents/${id}?decision=${decision}')`.
- `DocumentDistributionPage.tsx:95`: breadcrumb href → `/documents/${documentId}/details`.
- Sidebar meta panel: discoverable link to `/documents/:id/details` (from S2's `WorkspaceSidebar`).
- **TDD:** `routes` tests — `/documents/:id/details` renders `DocumentDetailLayout`; `/documents/:id` renders
  the workspace; `/documents/:id/edit` redirects preserving `:id`; `/approvals/:documentId` redirects
  preserving `location.search`. Grep gates: no bare `/documents/${id}/edit` targets except the redirect;
  `InboxPage` targets `/documents/`.
- Gate: `vitest run …/routes*.test.tsx`; `tsc --noEmit`; full `vitest run` (no regression).

---

## ADR
Write `wiki/decisions/00xx-single-artifact-destination.md` (governing §8.2) during S3; link in evidence.

## Files touched (aggregate)
S1: DecisionFooter(+test), ApprovalSidebar? · S2: DocumentWorkspacePage/EditorCanvas/WorkspaceSidebar/
ModeChip/ApprovingDisclosure(+tests), DocumentEditorPage, ArtifactMetaSidebar consumers · S3: routes.tsx ×2,
InboxPage, DocumentDistributionPage, ADR.

## Non-goals carried from spec
Comment replies (F2d.6); cockpit file/adapter deletion + worklist-target cleanup (F2d.7); no DTO/mutation change.

## Execution notes
_<filled per slice on close>_
