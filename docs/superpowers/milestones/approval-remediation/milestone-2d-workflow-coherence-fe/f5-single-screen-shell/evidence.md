# Feature F2d.5 — Evidence (accumulating per slice; final close after S3)

## S1 — DecisionFooter three-way variant · CLOSED
- Commit: `66cfb02c` (2026-07-09).
- TDD: extended `DecisionFooter.test.tsx` — review-stage+eligible→verdict CTAs (no signature);
  approval-stage→signature; active-stage+ineligible→neither. Failing first, then variant implemented
  (`stage_kind` + `viewer.eligible_for_active_stage`).
- Gates: `vitest run …/DecisionFooter.test.tsx` PASS; `tsc --noEmit` 0 errors.
- Closes the M2c deviation root at component level (ineligible observer no longer offered CTAs).

## S2a — Workspace shell + sidebar + mode chip + read modes · CLOSED
- Scope: `DocumentWorkspacePage` (thin owner: queries + `deriveWorkspaceMode` + constant shell),
  `WorkspaceSidebar` (embedded ArtifactMetaSidebar panel + ApprovalTimeline + sticky DecisionFooter +
  `/documents/:id/details` record link), `ModeChip` (PT-BR labels + why-tooltip from viewer facts).
  Read modes only (observing / reviewing / author-waiting / lifecycle) on read-only DocumentShell canvas
  (cockpit parity). Editing modes, approving disclosure, `?decision=` seed → S2b (`// S2b:` markers).
- Files (9 new): `pages/DocumentWorkspacePage.{tsx,module.css,test.tsx}`,
  `components/workspace/{WorkspaceSidebar,ModeChip}.{tsx,module.css,test.tsx}`.
- TDD: failing per-mode tests first (§2 read rows, §6 loading skeleton + instance-error-keeps-canvas),
  then implementation to green.
- Gates: `vitest run` 3 files / **22 tests PASS** (post-reboot rerun after clearing stale vite cache);
  `npx tsc --noEmit` **0 errors** (1 error found+fixed: refetch callback typed `Promise<QueryObserverResult>`
  vs `Promise<void>` — async wrapper, `DocumentWorkspacePage.tsx:137`).
- Independent review (cavecrew-reviewer): 1 finding — claimed DecisionFooter import path nonexistent.
  **Rejected as false positive with evidence:** file exists at
  `src/features/approval/components/sidebar/DecisionFooter.tsx`; path resolves; corroborated by tsc 0 +
  suite green. No other findings.
- Footer wiring: `decision={null}` (S2a) → reviewing+eligible shows verdict CTAs; observing/ineligible
  shows none (S1 contract exercised from the screen; fixtures set `viewer.eligible_for_active_stage`
  explicitly per M2c correction).

## Mid-feature amendment (2026-07-09, operator-ratified)
- Lazy editor chunk struck from F2d.5 (spec/plan/milestone amended in place): ineffective while the read
  canvas is DocumentShell — `DocumentShell.tsx:2` statically imports `MetalDocsEditor` (TipTap); the
  assertion would green while saving zero bytes.
- New feature **F2d.5b `f5b-pdf-read-canvas`** inserted (order F2d.5 → F2d.5b → F2d.6): PDF-driven
  viewing (all viewing = PDF; docx/TipTap only edit+review); backend contract for in-approval
  frozen-revision PDF (`view_service.go:43` gap; rendition already materialized at freeze, ADR 0015);
  `PdfCanvas` read modes + REAL lazy split; approver signs over rendition bytes.
- Grounding: adversarial industry review (2026-07-09, two independent opposing research briefs) —
  single canonical doc URL CONFIRMED (Google/Figma/Word/Notion/Veeva); decision act = bounded in-page
  surface (Veeva modal, GitHub review flow); separate ceremony URL only for external signers
  (DocuSign/Adobe) = existing reopen trigger; PDF read = Veeva viewable-rendition pattern.

## S2b — Editing + approving modes · CLOSED
- Scope: `EditorCanvas` extracted from `DocumentEditorPage` body (normal import, NOT lazy — lazy → F2d.5b);
  `DocumentWorkspacePage` canvas now branches on `mode` — author-editing/author-changes-requested →
  writable `EditorCanvas`; approving → read-only frozen DocumentShell + `ApprovingDisclosure`
  (reused `IntegrityDisclosure` + "assinando por delegação de X" badge from `viewer.via_delegation_from`);
  other modes unchanged (S2a read canvas). `author-changes-requested` gets a wine-token warning banner atop
  the canvas + `RequestedChangesPanel` (F6) threaded as `WorkspaceSidebar.contextualPanel` (between timeline
  and footer).
- `?decision=` seed: `useSearchParams().get('decision')` → `buildDocumentSignoffDecision.defaultOptionKey`,
  built ONLY in approving mode; non-approving decision stays null. Signature panel gates on
  `viewer.eligible_for_active_stage` via the S1 DecisionFooter (decision!=null path).
- Submit: reused `useSignoffMutation` verbatim (same If-Match/content_hash contract), mirroring
  ApprovalCockpitPage's `decisionSubmit` (signOff → refetch). `contentHash`/`revisionVersion` from
  `useControlledDocumentActiveDocumentQuery`, gated `enabled` unless approving. No new mutation, no
  If-Match/DTO change (F2d.4 owns instance state).
- Files (new): `components/workspace/EditorCanvas.{tsx,test.tsx}`,
  `components/workspace/ApprovingDisclosure.{tsx,module.css,test.tsx}`. (changed) `DocumentEditorPage.tsx`
  (body → EditorCanvas), `DocumentWorkspacePage.{tsx,module.css,test.tsx}`, `WorkspaceSidebar.tsx`
  (decision?/contextualPanel? props). Hooks hoisted above early returns in the owner (reviewer-verified
  stable order, no conditional hooks).
- TDD: failing per-mode tests first (author-editing, author-changes-requested banner+F6, approving eligible
  → signature+disclosure, approving+`?decision=approve` preselect, approving NOT eligible → no panel with
  explicit `viewer.eligible_for_active_stage=false` fixture), then implementation to green.
- Gates (self-verified, not subagent-claimed): `vitest run` 6 files / **63 tests PASS** incl.
  `DocumentEditorPage.test.tsx` 30/30 (extraction fidelity); `tsc --noEmit` **0 errors**.
- Independent review (cavecrew-reviewer): **no findings** — hooks-order correct, mode source sole,
  decision seed approving-only, submit reuse intact, EditorCanvas handlers identical, approving canvas
  read-only (PDF → F2d.5b), Wine tokens only, no `any`.

## S3 — Route flip + deep-links + breadcrumb + ADR · CLOSED
- `documents/routes.tsx`: `/documents/:id` → `DocumentWorkspacePage` (the new screen goes LIVE);
  `DocumentDetailLayout` subtree (index `DocumentDetailRoute` + `distribution` child) moved to
  `/documents/:id/details`; `/documents/:id/edit` → `RedirectEditToWorkspace` (reads `useParams`,
  preserves `:id`). `documents/new` + `RedirectToLibrary` statics still win over `:documentId`.
- `approval/routes.tsx`: `/approvals/:documentId` → `RedirectApprovalToWorkspace` (reads `useParams` +
  `useLocation`, forwards `location.search` so `?decision=` survives) → `/documents/:id`. Does NOT import
  or delete `ApprovalCockpitPage` (F2d.7 owns deletion — file confirmed still present).
- Deep-links: `InboxPage.tsx` both navigations retargeted `/approvals/` → `/documents/` (grep gate:
  no `/approvals/` left in InboxPage); `DocumentDistributionPage.tsx:95` breadcrumb href →
  `/documents/${id}/details`.
- ADR: `wiki/decisions/0080-single-artifact-destination.md` (+ index.md row) — canonical `/documents/:id`,
  record at `/details`, `/edit` + `/approvals/:id` redirect (query preserved), reopen trigger =
  external-signer persona, notes the F2d.5b PDF-read-canvas amendment.
- TDD: `routes.test.tsx` rewritten — 4 assertions (RED first): `/documents/:id/details`→layout,
  `/documents/:id`→workspace, `/documents/:id/edit`→redirect (pathname+param), `/approvals/:id?decision=approve`
  →redirect (pathname AND search preserved). `InboxPage.test.tsx` assertions updated to the intentional
  retarget.
- Gates (self-verified): routes 4/4 + Inbox 20/20 + Distribution 7/7 = **31 PASS**; `tsc --noEmit` **0**.
  Full suite 866 pass / 1 fail = `ApprovalCockpitPage ?decision=reject preselect` — **provably not a
  regression**: S3 modifies no cockpit file/component (git status confirms), so the failure is independent
  of this slice (owned by cockpit which F2d.7 retires anyway).
- Independent review (cavecrew-reviewer): **no findings** — route priority correct, redirects preserve
  param+query, cockpit/editor files intact, deep-links retargeted, tests real, ADR format matches 0079.
- Deferred (flagged, out of S3 scope → F2d.7 mechanical cleanup per milestone.md): 4 residual
  `/documents/${id}/edit` navigation constructors (`documentWorkflow.ts:30`, `NewDocumentWizardPage.tsx:179`,
  `DocumentDetailRoute.tsx:101,145`) — functional today (bounce one hop through the `/edit` redirect).

---
## F2d.5 — FEATURE COMPLETE
All slices closed: S1 (66cfb02c) · S2a (d2de8e97) · S2b (9e516725) · S3 (this commit). The mode-adaptive
single working screen is live at `/documents/:id`. Lazy editor split relocated to **F2d.5b** (PDF read
canvas — operator-ratified 2026-07-09). Non-goals held: no comment replies (F2d.6), no cockpit deletion
(F2d.7), no DTO/mutation change.
