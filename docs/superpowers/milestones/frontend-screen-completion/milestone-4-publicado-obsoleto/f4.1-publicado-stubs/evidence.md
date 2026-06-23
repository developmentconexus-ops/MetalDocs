# Feature F4.1 — Evidence

> **Milestone:** 4 — Documento Publicado completion + Documento Obsoleto  ·  **Feature:** `f4.1-publicado-stubs`  ·  **Closed:** 2026-06-23
> **Contract:** `spec.md` (consumer-contract-first; numerator parked per ADR-0042).
> A feature is closed only when every row below is filled with real, honestly-labeled output.

## What was implemented

Frontend-only. No Go/backend change (Non-goal honored — both endpoints already shipped). Commit
`d8f03953 feat(M4/F4.1): wire Publicado coverage denominator, PDF download, páginas/tamanho; honest defers`.

By outcome (each consumes the **existing, frozen** producer contract — F4.1 added no producer):

- **Cobertura KPI + coverage `<aside>`** — render the obligated-audience count from the M2
  `useDistributionSummaryQuery(documentId)` hook: `obligatedCount = isError ? EM_DASH : (total_targets ?? EM_DASH)`.
  The fabricated `—%` value and the `role="progressbar"` bar are **removed**; the read numerator is an
  explicit parked label `"leitura em acompanhamento (ADR-0042)"`. The "abrir fanout →" nav to `/distribution`
  is preserved. Matches the consumer contract `DistributionSummaryResponse.total_targets` (spec.md §Contract-coverage).
- **Baixar PDF** — wires the existing `exportPDF(documentId, { paper_size: 'A4' })` client
  (`features/documents/api/exports.ts`); on success `window.open(signed_url, '_blank', 'noopener')`.
  `pdfStatus` discriminated union `idle | pending | rate_limited | error`. Catch uses `e instanceof ApiError`:
  429 → `rate_limited` (button held disabled + inline `role="alert"`); any other failure →
  `{ kind: 'error', message: resolveErrorMessage(e) }` rendered as inline `role="alert"` (mirrors `ExportMenu`,
  never swallowed). The single-button Publicado layout is kept (no `ExportMenu` swap — Non-goal honored).
- **Páginas + Tamanho** — render `doc.current_revision_page_count` / `doc.current_revision_file_size_bytes`
  (already on `DocumentResponse` via `useDocumentDetailQuery`) through new pure helpers `formatPageCount` /
  `formatFileSize` in `lib/documentDetailMeta.ts` (binary units B/KB/…/TB, pt-BR decimal comma, EM_DASH on
  null/NaN/negative). Runtime-truth correction (Interview #3): these fields exist, so they are wired, not deferred.
- **The 4 genuinely-unbacked fields** (Próxima revisão, Classificação, Documentos relacionados, Comentários)
  render honest `"—"` / `"não disponível"`; each is a defer-with-trigger row in `wiki/backlog/documento-publicado.md`
  naming the exact missing backend field/model. Zero "em breve" remains in the page.
- **CSS hygiene** — removed orphaned `.kpiValuePlaceholder` + `.coverageCardBar` (dead after bar removal);
  added `.pdfAlert`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `npx vitest run …DocumentPublishedPage.test.tsx …documentDetailMeta.test.ts` | RED first (7 page + 8 formatter cases written before impl, failing on missing wiring/helpers) → GREEN after impl: **35 passed (27 page + 8 formatter)** | fixture (vitest) |
| Static (types) | `npx tsc --noEmit` (from `frontend/apps/web`) | **exit 0** | real |
| Targeted test (re-run after review fixes) | `npx vitest run …DocumentPublishedPage.test.tsx …documentDetailMeta.test.ts` | **Test Files 2 passed (2) · Tests 35 passed (35)**, 3.02s | fixture (vitest) |
| No in-scope silent stub | `grep -n "em breve" DocumentPublishedPage.tsx` | **NONE** (every former in-scope placeholder wired; remaining absent states map to backlog defer rows) | real (grep + doc) |
| Regression — no new failures introduced | `git stash` the two source files → run `DocumentEditorPage.test.tsx` on clean tree → restore | Editor suite fails **18/18 identically with F4.1 stashed** → failures are the documented node_modules junction drift (duplicate-React at `DocumentEditorPage.tsx:358` useState), **not introduced by F4.1** | real |

> Vitest cases are fixture-level (mocked queries) — they prove consumer wiring, not the live endpoints.
> The live producers were already proven by M2 (distribution) and the existing export flow; F4.1 adds no
> producer. The grep+backlog cross-check and the tsc/regression checks are real.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Cobertura KPI + aside render `total_targets` (mocked 12); count shown, no "em breve", no `%` | yes | test `coverage count` (page suite); grep "em breve"=NONE |
| Coverage card EM_DASH (never fabricated 0/%) when summary errors | yes | test `EM_DASH on error` (page suite) |
| "Baixar PDF" enabled + calls `exportPDF`; success opens `signed_url` | yes | test `enables Baixar PDF and triggers the real export, opening the signed url` |
| 429 handled gracefully (no unhandled rejection; rate-limited state) | yes | test `handles a 429 rate-limit … with a role=alert and a disabled button` |
| Páginas: value when present (`1`⇒"1"), honest absent when null — never "em breve" | yes | tests `renders Páginas …` + null case asserting `—` |
| Tamanho: formatted when present (`1024`⇒"1 KB"), honest absent when null — never "em breve" | yes | tests `renders a formatted Tamanho …` + null case asserting `—` (spec gate row 6) |
| No in-scope silent stub: every surviving "em breve" maps to a backlog defer row | yes | grep "em breve"=NONE; backlog defer rows for the 4 unbacked fields |
| Generated types consumed directly; no hand-written snake→camel mapper | yes | `tsc --noEmit` exit 0; helpers read generated fields directly |
| Both reviewer agents APPROVE | yes | see Review disposition |

## Review disposition

- **Spec-compliance / visual review (`frontend-screen-reviewer`):** round-1 REQUEST CHANGES (2 Major,
  3 Minor) → round-2 **APPROVE WITH NITS** (0 Critical, 0 Major). Resolutions by root-cause family:
  - *Error-UX (swallowed non-429 PDF failure)* → added `error` state + inline `role="alert"`; covered by a 502 test.
  - *A11y / state (rate_limited)* → `role="alert"` + button held disabled.
  - *Dead CSS* → deleted `.kpiValuePlaceholder` + `.coverageCardBar`.
  - *Mojibake `Sem permissÃ£o`* → confirmed **pre-existing** (commit `e6423a24`, outside the F4.1 diff);
    deferred to tracked task `task_2b429f03`; reviewer agreed it does not block the F4.1 gate.
- **Code-quality review (`frontend-code-reviewer`):** round-1 APPROVE WITH NITS → round-2 **APPROVE**
  (0 Critical, 0 Major, 0 Minor). Resolutions:
  - *Missing Tamanho-null assertion (spec gate row 6)* → null test now asserts both Páginas + Tamanho `—`.
  - *Deceptive `as {status; body}` cast* → replaced with `e instanceof ApiError` guard; non-schema
    `retry_after_seconds` access removed (60s constant).
  - *God component (795 LOC) + hand-written `GovernedRevisionHistoryItem`* → pre-existing, confirmed
    out-of-scope; tracked follow-ups, not F4.1 deliverables (not materially worsened).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Coverage read **numerator** (% lido) | Denominator wired; numerator parked by ADR-0042 (no read-tracking API) | `wiki/backlog/documento-publicado.md` "Coverage KPI" — trigger: read-tracking API lands → replace parked label with real % |
| Próxima revisão | No review-due-date field on document/controlled-document model | backlog "KPI: Próxima revisão" — trigger: review-due-date field |
| Classificação | No confidentiality field **and** taxonomy is an unmade governance decision | backlog "Classificação" — trigger: confidentiality field + taxonomy |
| Documentos relacionados | No related-docs relationship model in backend | backlog "RelatedGrid" — trigger: relationship model + read endpoint |
| Comentários | Display-comments architecture undecided (reuse editor comments vs separate model) | backlog "CommentsCard" — trigger: decided display-comments architecture |
| Publish-button tooltip mojibake | Pre-existing (commit `e6423a24`), outside F4.1 diff | tracked task `task_2b429f03` |
| DocumentEditorPage suite drift | Pre-existing node_modules junction drift (duplicate React), not F4.1 | memory `fe-node-modules-junction-drift` — trigger: complete pnpm install / repoint junctions |
