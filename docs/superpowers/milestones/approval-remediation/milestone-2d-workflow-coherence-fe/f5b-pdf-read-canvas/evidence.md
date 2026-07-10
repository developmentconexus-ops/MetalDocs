# Feature F2d.5b — Evidence (PDF official-view canvas + real lazy editor split)

**Re-scope note (2026-07-09):** the ORIGINAL F2d.5b premise ("serve the already-materialized
frozen-revision PDF to *in-approval* viewers; approver signs over rendition bytes") was disproven
at source by the `developing-new-work` system-impact gate → **RED / AS-2**. Freeze is
**terminal-only**: Pin fires only at `instance.Status == InstanceApproved`
(`internal/modules/documents/approval/application/decision_service.go:408`), same tx flips the doc
`under_review → approved`; the PDF (`final_pdf_s3_key`) is materialized async *after*. So an
in-approval doc has **no PDF**, and its content stays mutable through review. Operator ratified the
refined **Option C** redesign (analysis `docs/superpowers/analysis/2026-07-09-f5b-pdf-read-canvas-system-impact.md`;
design `docs/superpowers/specs/2026-07-09-f5b-pdf-official-view-design.md`): **PDF is the official
POST-approval artifact; in-approval viewing stays on the in-app source canvas.** Feature re-scoped
**FE-only, zero backend** — the `/documents/:id/view` contract is consumed as-is, never changed.

Base SHA: `9f0af980`. Design: `docs/superpowers/specs/2026-07-09-f5b-pdf-official-view-design.md`.
ADR amendment: `wiki/decisions/0080-single-artifact-destination.md` (F2d.5b re-scope note, lines 80-94).

---

## D1 — `PdfCanvas` (status-keyed official-PDF surface) · CLOSED
- Commits: `339a022f` (impl) + `e06520d8` (fail-closed fix).
- Component `components/workspace/PdfCanvas.tsx` (+ `.module.css`, `.test.tsx`). Reuses the existing
  `useDocumentPdfStatus(documentId, true)` hook (polls `GET /documents/{id}/view`). Render contract:
  - `ready && url` → `<iframe title="Documento oficial (PDF)" src={url}>`.
  - `failed` **OR** `ready && !url` → `role="alert"` + "Tentar novamente" (calls `pdf.retry`).
  - else → `role="status" aria-live="polite"` "Gerando o PDF oficial…".
- **Fail-closed integrity (no-fallback invariant):** a `ready` status with missing/empty `url` is a
  backend-contract violation and is surfaced as an alert — NOT masked as an infinite spinner.
  Independent-review finding on the first cut (🟡 ready-without-url fell through to pending); fixed in
  `e06520d8` with `const missingReadyUrl = pdf.status === 'ready' && !pdf.url;` OR'd into the alert
  branch, plus a dedicated failing-first test.
- TDD: RED shown, then GREEN. Tests: ready→iframe, pending→status, failed→alert+retry,
  ready-without-url→alert (fail-closed). `vitest run …/PdfCanvas.test.tsx` **4/4 PASS**; `tsc` 0.
- CSS uses only verified Wine tokens present in `DocumentWorkspacePage.module.css`
  (`--sp-2/-3`, `--r-2`, `--text`, `--text-muted`, `--border`, `--surface`).

## D1 — wiring into `DocumentWorkspacePage` (status-keyed canvas branch) · CLOSED
- Commits: `7cdbfaa0` (impl) + `6144ceb9` (robustness assertion).
- `DocumentWorkspacePage.tsx`: module constant `OFFICIAL_PDF_STATUSES = new Set(['approved',
  'scheduled', 'published'])` — the exact set the backend `/documents/:id/view` serves
  (`view_service.go` `viewableStatuses`). New canvas branch renders `<PdfCanvas documentId={documentId} />`
  **AFTER** approving/author-editing/author-changes-requested, **BEFORE** the generic docx read
  canvas (`DocumentShell` readonly) fallback.
- **Keyed on STATUS, not mode** — deliberate: the `lifecycle` workspace mode also covers
  `superseded`/`obsolete`, which `/view` does **not** serve; those must keep the docx read canvas.
  Gating on `mode === 'lifecycle'` would wrongly route them to a PDF that never materializes.
- TDD: 4 new tests (published→PdfCanvas; approved-author→PdfCanvas; superseded→docx not PdfCanvas;
  under_review-observing→docx not PdfCanvas), each with positive + negative assertions. One
  pre-existing lifecycle test's fixture was moved `approved → superseded` (was incidentally colliding
  with the new official-PDF set) — its original intent (lifecycle-mode chrome: "Visualizando" chip,
  no verdict CTAs, timeline) is preserved; `superseded` still derives `lifecycle`. Reviewer confirmed
  no coverage lost.
- Gates: `vitest run …/DocumentWorkspacePage.test.tsx` **16/16 PASS**; `tsc` 0.

## D2 — real lazy `MetalDocsEditor` chunk · CLOSED
- Commit: `cc2274a1`.
- `DocumentShell.tsx`: `MetalDocsEditor` converted from static value-import to
  `const MetalDocsEditor = lazy(() => import('@metaldocs/editor-ui').then((m) => ({ default:
  m.MetalDocsEditor })));`. Line-2 import is now `import type { MetalDocsEditorRef, EditorComment,
  TrackedChange }` (erased at compile). Mount wrapped in `<Suspense>` **inside** DocumentShell
  (fallback `role="status" aria-live="polite"` "Carregando editor…") so no consumer needs its own
  boundary. Every editor prop unchanged; `ref={editorRef}` still forwards through `lazy`.
- **Effect:** the heavy TipTap/ProseMirror bundle is fetched ONLY when a docx read/edit canvas
  mounts; the PDF/lifecycle canvas (`PdfCanvas`) and the loading/error states never pull it.
- TDD proof — `components/editorChunk.lazy.test.tsx` runtime evaluation-counter (`vi.hoisted`):
  importing `DocumentShell` ⇒ editor-ui eval count unchanged (static graph does NOT pull it);
  mounting the ready docx path ⇒ eval count > 0 (lazy chunk fetched on mount). RED shown before the
  conversion (`expected 1 to be +0`), GREEN after. Chosen the runtime counter over a source-string
  assertion because it empirically discriminates static-vs-lazy.
- No existing test needed `getBy → findBy` conversion: the docx suites already awaited the editor via
  `waitFor(() => getByTestId('editor'))`, so the added Suspense microtask didn't break them.
- Gates: `tsc` 0; `vitest run src/features/documents` **261/261 PASS**.

---

## Feature gates (Task 4, self-verified from clean state)
- **tsc:** `npx tsc --noEmit -p .` → **0 errors**.
- **Static-graph gate:** the ONLY remaining `MetalDocsEditor` **value**-imports are the two templates
  whitelist files (`features/templates/pages/TemplateEditorPage.tsx`,
  `features/templates/components/TemplateReviewCanvas.tsx`, both out of scope). `DocumentShell.tsx:2`
  is now `import type`; its sole runtime edge to editor-ui is the dynamic `import()` inside `lazy()`.
  Every other editor-ui import across `src` is type-only.
- **Zero-backend gate:** `git diff --name-only 9f0af980 HEAD` = 7 files, ALL under
  `frontend/apps/web/src/features/documents`. No backend, no contract, no OpenAPI, no non-FE file.
- **documents suite:** `vitest run src/features/documents` → **261/261 PASS**.
- **approval suite:** `vitest run src/features/approval` → **163/164** — the sole failure is
  `ApprovalCockpitPage.test.tsx:342` (`?decision=reject` inline preselect), the documented
  pre-existing fail in the cockpit that F2d.5b never touches (git diff confirms no cockpit file in
  the feature). Cockpit retires in F2d.7. **Not a regression.**
- **Bundle-split build evidence:** bounded defer — `npm run build` is unreliable under the known
  pnpm-junction drift (memory `fe-node-modules-junction-drift`). The `editorChunk.lazy.test.tsx`
  runtime counter is the primary, deterministic chunk-boundary proof.

## Independent reviews (cavecrew-reviewer)
- D1 PdfCanvas: 1 🟡 (ready-without-url infinite pending) → fixed `e06520d8`, re-green.
- D1 wiring: 0 bugs, 1 minor nit (test-2 negative assertion) → folded in `6144ceb9`.
- D2 lazy split: **no issues**.
- **Whole-feature final review (`9f0af980..HEAD`): no issues — feature ships.** Confirmed contract
  fidelity (status set == backend viewableStatuses), fail-closed integrity, lazy-split correctness
  (single dynamic edge), no assertion hollowing, no dead code, a11y (iframe title, alert/status roles).

---
## F2d.5b — FEATURE COMPLETE
D1 (`339a022f`+`e06520d8`+`7cdbfaa0`+`6144ceb9`) · D2 (`cc2274a1`). PDF is the official
post-approval artifact (approved/scheduled/published → `PdfCanvas` via `/view`); in-approval viewing
stays on the in-app source canvas; the TipTap bundle is a real lazy chunk fetched only by docx
canvases. Zero backend. Signature subject unchanged (source `content_hash`, existing If-Match).
NOT pushed. Next: F2d.6 (author comment replies) → F2d.7 (cockpit deletion) → F2d.8 (UI-driven live QA).
