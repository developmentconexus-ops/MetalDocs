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

## S3 — pending (route flip + deep-links + breadcrumb + ADR)
