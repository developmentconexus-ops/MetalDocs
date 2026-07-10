# F2d.5b — PDF Official View (re-scoped) — Design

**Date:** 2026-07-09
**Status:** Ratified by operator 2026-07-09 (this doc = the ratification record)
**Supersedes:** the original F2d.5b premise ("serve frozen-revision PDF to in-approval viewers; approver signs over rendition bytes") — disproven by the RED/AS-2 system-impact gate
(`docs/superpowers/analysis/2026-07-09-f5b-pdf-read-canvas-system-impact.md`).
**Program:** approval-remediation · M2d `workflow-coherence-fe` · feature F2d.5b `f5b-pdf-read-canvas`.

---

## 1. Decision

**PDF is the official post-approval artifact; in-approval viewing stays on the in-app source canvas.**

| Document status | Workspace canvas (read modes) |
|---|---|
| `draft`, `under_review` — all read modes (reviewing / approving / observing / author-waiting) | docx read-only (`DocumentShell`, built in F2d.5) |
| `approved`, `scheduled`, `published` — lifecycle mode | **official PDF** (`PdfCanvas` via existing `GET /documents/{id}/view`) |

Canvas branch is keyed on **document status** (the artifact truth), not on workspace mode; mode
continues to drive sidebar/footer/chip as in F2d.5.

**Backend: ZERO change.** `viewableStatuses` (`view_service.go:43`) stays `{approved, scheduled,
published}`. No new pipeline, no contract change, no migration. Signature subject unchanged:
approver signs the source `content_hash` (existing If-Match contract).

## 2. Rationale (evidence-backed, adversarial research 2026-07-09)

1. **Veeva renders early renditions because its source is an uploaded binary the browser cannot
   display; MetalDocs has no such gap.** MetalDocs edits in-app (TipTap) — the source is natively
   viewable read-only, with comments and tracked changes. The modern in-app-editing SaaS tier
   (Qualio) reviews/approves on the in-app document view, PDF only as export/official artifact.
   Copying Veeva's continuous-rendition pipeline would solve a problem this product does not have.
2. **A mid-review PDF would be falsely official.** Computed placeholders (doc_code, approvers,
   approval_date, effective_date) resolve only at freeze, which fires on the *final* signoff
   (`decision_service.go:408`). A pre-freeze PDF would render unresolved tokens /
   "[aguardando aprovação]" while *looking* final — a misleading third artifact that is neither the
   source of truth nor the official rendition. The only truly official PDF is the post-approval one,
   which already exists and is already served.
3. **21 CFR Part 11 requires nothing more.** The signature binds the electronic record
   (`content_hash`); the regulation only requires the human-readable form to carry the signature
   manifestation. Signing "rendition bytes" was never a requirement (verified against eCFR Part 11 +
   Veeva's own model: signature page overlaid on rendition, record is what is signed).

## 3. Scope — F2d.5b is now FE-only

**D1 — `PdfCanvas` (lifecycle mode).** New component under
`frontend/apps/web/src/features/documents/components/workspace/`. Renders the official PDF for
`approved/scheduled/published` docs inside `DocumentWorkspacePage`'s canvas slot. Reuses the
existing `useDocumentPdfStatus` hook (poll `/view`, states `pending | ready | failed`) — no new
fetching logic. `pending` → generating-state UI; `failed` → error state with retry; `ready` →
embedded PDF (presigned URL, `<iframe>`/`<object>`).

**D2 — Real editor lazy split.** `DocumentShell.tsx:2` statically imports `MetalDocsEditor`
(`@metaldocs/editor-ui`), so the editor chunk (TipTap/ProseMirror + docx deps) ships in the route
bundle for every visitor. Split it into a lazy chunk. Honest accounting: docx read modes
(under_review) render content *via* read-only TipTap, so they still fetch the chunk — lazily, on
canvas mount. The wins are (a) the workspace route's initial bundle drops the editor entirely, and
(b) **lifecycle mode (PdfCanvas) never fetches it** — the most-viewed state (published docs) pays
zero editor bytes. Assert at chunk level (editor absent from the route's static import graph AND not
fetched when rendering lifecycle mode) — the F2d.5 lesson: an ineffective lazy assertion greens
while saving zero bytes.

**Non-goals (held):** no backend change of any kind; no continuous/preview rendition; no change to
signature subject or If-Match contract; no comment replies (F2d.6); no cockpit deletion (F2d.7).

## 4. Reopen trigger (recorded, not built)

If customers demand print-fidelity preview **during** approval → the documented upgrade is the
Veeva-pattern continuous rendition: eager render at submit/resubmit via transactional outbox, object
keyed by `content_hash`, lazy fallback in `/view`. Researched and shaped 2026-07-09; slots into the
existing pipeline architecture. YAGNI until the demand is real.

## 5. Decision record

Amend **ADR 0080** (single-artifact destination) with a short section: in-approval viewing = source
canvas; PDF = official post-approval artifact; signature binds `content_hash`; reopen trigger =
customer demand for print-fidelity in-approval preview. No new ADR (no MUST-deviation — this
*conforms* to ADR 0015 freeze semantics instead of fighting them).

## 6. Test plan

- **D1:** component tests (vitest) — lifecycle doc → PdfCanvas rendered (not DocumentShell);
  pending/failed/ready states; under_review read modes still DocumentShell. TDD, failing first.
- **D2:** split assertion at build level (chunk manifest / import graph), plus existing suites stay
  green (DocumentEditorPage 30/30, workspace suites).
- **Gates:** `vitest run` targeted + full FE suite, `tsc --noEmit` 0, independent cavecrew-reviewer,
  evidence.md, UI-driven live QA folds into F2d.8.

## 7. Resolution of the RED gate

The AS-2 hard-stop is resolved by *removing the false premise*, not by building around it: no
in-approval PDF is needed, so no foundation change is needed. The developing-new-work gate does not
need a re-run: F2d.5b now touches zero backend modules, zero invariants (FE-only rendering of an
already-served contract), and reuses existing FE frameworks (`useDocumentPdfStatus`, workspace
canvas slot).
