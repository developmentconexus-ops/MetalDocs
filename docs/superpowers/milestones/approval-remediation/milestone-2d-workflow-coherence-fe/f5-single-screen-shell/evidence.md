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

## S2b — pending
## S3 — pending (route flip + deep-links + breadcrumb + ADR)
