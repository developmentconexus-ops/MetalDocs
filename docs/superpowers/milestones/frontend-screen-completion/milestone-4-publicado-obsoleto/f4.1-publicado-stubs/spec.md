# Feature F4.1 — Spec

> **Milestone:** 4 — Documento Publicado completion + Documento Obsoleto  ·  **Folder:** `f4.1-publicado-stubs`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-23 / leandrotca — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Driven via `superpowers:brainstorming` as the feature interview engine, seeded with the F4.1 row of
`../milestone.md`. Recon established that both target endpoints already ship and that M2 already built
reusable consumers (`useDistributionSummaryQuery`, `api/exports.ts`), so most of the contract is fixed
by precedent; the two genuinely-open consumer-contract decisions were put to the operator.

| # | Question | Answer |
|---|----------|--------|
| 1 | Coverage card: endpoint serves only the obligated denominator (`total_targets`); read numerator is parked (ADR-0042). How to render? | **Count + parked label.** Show `total_targets` as the obligated-audience count, drop the fabricated `%`/progress bar, add an explicit "leitura em acompanhamento (ADR-0042)" parked note. EM_DASH on error/missing — never a fabricated number (mirrors M2's `DocumentDistributionPage` rule). |
| 2 | PDF download: reuse existing `exportPDF` client (POST → `window.open(signed_url)`, 429 handling). What UI? | **Wire the existing single "Baixar PDF" button** (loading state, 429 handling, `window.open(signed_url,'_blank','noopener')`). Do NOT swap in the two-button `ExportMenu` component — keep the Publicado design's one-button layout. |
| 3 | The 6 unbacked fields (próxima revisão, páginas, tamanho, classificação, related docs, comments) — build now or defer? | **Runtime-truth correction (recon, 2026-06-23):** the backlog doc (stale, 2026-05-29) claimed page-count/file-size have no backend, but `DocumentResponse` **already carries** `current_revision_page_count` + `current_revision_file_size_bytes` (openapi.yaml:4832/4837; generated FE types index.d.ts:2540-2541; both nullable). By the mission's own rule (no silent stub for a *backend-available* item), **Páginas + Tamanho are in-scope and MUST be wired** (nullable → honest absent state when null). **Defer the remaining 4** — Próxima revisão, Classificação, Documentos relacionados, Comentários — which genuinely have no backend field/model. So F4.1 = **4 wired (coverage, PDF, páginas, tamanho) + 4 defer**. (Note: `size_bytes` from the PDF export is the *PDF* size, not the source-doc size — Tamanho uses `current_revision_file_size_bytes`, not the export response.) |

## Consumer contract (FIRST — before any producer)

This feature has **no new producer** — both producers already exist and are contract-frozen. F4.1 is a
*consumer* of two shipped endpoints via two already-built FE clients. The "contract" is therefore the
existing, generated shape the screen must consume unchanged.

- **Consumer(s):** `DocumentPublishedPage` (`features/documents/pages/DocumentPublishedPage.tsx`) — the
  Cobertura KPI cell + coverage `<aside>`, and the "Baixar PDF" button.
- **Contract — coverage:** `GET /documents/{id}/distribution` → `DistributionSummaryResponse`
  `{ total_targets: integer ≥ 0 }` (denominator only; numerator out of scope per ADR-0042).
  Consumed via the existing `useDistributionSummaryQuery(documentId)` hook
  (`features/documents/queries/useDistributionSummaryQuery.ts`) → `getDistributionSummary`
  (`features/documents/api/distribution.ts`). Error/loading semantics follow M2 precedent: EM_DASH on
  `isError`, `total_targets ?? EM_DASH`.
- **Contract — PDF:** `POST /documents/{id}/export/pdf` body `{ paper_size?: 'A4'|'Letter',
  landscape?: boolean }` → `{ storage_key, signed_url, composite_hash, size_bytes, cached, revision_id }`,
  rate-limited 20/min (429 → `retry_after_seconds`). Consumed via the existing `exportPDF(documentID,
  { paper_size: 'A4' })` client (`features/documents/api/exports.ts`).
- **Contract — páginas / tamanho:** `DocumentResponse.current_revision_page_count` (`integer ≥ 1 |
  null`) and `current_revision_file_size_bytes` (`int64 ≥ 0 | null`) — already returned by
  `GET /documents/{id}` and consumed via the existing `useDocumentDetailQuery`. Nullable: render the
  value when present, an honest absent/"—" state when null. Source of truth: openapi.yaml:4832/4837;
  generated FE types `lib/api-types/index.d.ts:2540-2541`.
- **Source of truth for the contract:** `api/openapi/v1/openapi.yaml` ops `getDocumentDistribution`
  (DistributionSummaryResponse, openapi.yaml:4154) + `exportDocumentPDF` (openapi.yaml:3026); generated
  FE types `lib/api-types/index.d.ts`; existing callers `DocumentDistributionPage.tsx` (summary) and
  `ExportMenu.tsx` (PDF flow precedent).

## What this feature implements

After F4.1, on the **published** document screen (`/documents/:id`):

1. **Cobertura KPI + coverage aside** render the live obligated-audience count from
   `useDistributionSummaryQuery` — `total_targets` as "N destinatários obrigados", EM_DASH on
   error/missing — with the read numerator shown as an explicit parked label (ADR-0042). The fabricated
   `—%` value and the percentage progress bar are removed (no fake numerator). The existing
   "abrir fanout →" navigation to `/distribution` is preserved.
2. **Baixar PDF** is enabled (no longer `aria-disabled`) and, on click, calls the existing `exportPDF`
   client (POST, `paper_size: 'A4'`), shows a pending state, opens the returned `signed_url` in a new
   tab on success, and handles 429 (rate-limited) gracefully.
3. **Páginas + Tamanho** render from `DocumentResponse.current_revision_page_count` and
   `current_revision_file_size_bytes` (already returned by `useDocumentDetailQuery`) — formatted value
   when present, honest absent/"—" state when null (no "em breve").
4. **The 4 genuinely-unbacked fields** (Próxima revisão, Classificação, Documentos relacionados,
   Comentários) each become a **defer-with-trigger** row in `wiki/backlog/documento-publicado.md` naming
   the exact backend field/model that unblocks them, and the screen renders each as an honest empty/
   absent state. After F4.1, every surviving "em breve" in `DocumentPublishedPage.tsx` corresponds to a
   backlog defer row (the coverage, PDF, páginas, and tamanho "em breve" placeholders are gone).

## Non-goals (mandatory)

- **No backend / Go source change.** Both endpoints already exist; F4.1 changes only frontend. (HS-2 if a
  placeholder turns out to need backend.)
- **No read-tracking numerator / % coverage.** Parked (ADR-0042). The card shows the denominator only.
- **No backend for the 4 genuinely-deferred fields** (review-due date, confidentiality classification,
  related-docs relationship model, display-comments architecture) — defers, not code. (Páginas + Tamanho
  are NOT deferred — their fields already exist; see Interview #3.)
- **No `ExportMenu` swap-in / docx button** on Publicado — wire the existing single PDF button only.
- **No Publicado restyle / layout redesign** — it is already on redesign tokens; F4.1 closes gaps only.
- **No obsolete-variant work** — that is F4.2.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Cobertura KPI + aside render `total_targets` from `useDistributionSummaryQuery` (mocked to a count, e.g. 12) — text shows the count, not "em breve", not a `%` | new case in `DocumentPublishedPage.test.tsx` (vitest, MSW/mocked query) asserting the count renders and no `%`/progress-bar element is present | fixture (vitest) |
| Coverage card shows EM_DASH (never a fabricated 0/%) when the summary query errors | vitest case forcing `isError` → asserts EM_DASH, no `%` | fixture (vitest) |
| "Baixar PDF" button is enabled (not `aria-disabled`) and calls `exportPDF`; success opens `signed_url` | vitest case: click → `exportPDF` spy called with `{paper_size:'A4'}`; `window.open` spy called with the signed_url | fixture (vitest) |
| 429 on PDF export is handled (no unhandled rejection; user-visible rate-limited state) | vitest case: `exportPDF` rejects `{status:429}` → asserts graceful state, button re-enabled | fixture (vitest) |
| Páginas renders `current_revision_page_count` when present (fixture `1`) and an honest absent state when null — never "em breve" | vitest cases: page_count `1` ⇒ "1" shown; `null` ⇒ absent/"—", no "em breve" | fixture (vitest) |
| Tamanho renders a formatted `current_revision_file_size_bytes` when present (fixture `1024` ⇒ e.g. "1 KB") and an honest absent state when null — never "em breve" | vitest cases: size `1024` ⇒ formatted size; `null` ⇒ absent/"—", no "em breve" | fixture (vitest) |
| No in-scope silent stub remains: every surviving "em breve" in `DocumentPublishedPage.tsx` maps to a backlog defer row | `grep -n "em breve" DocumentPublishedPage.tsx` cross-checked against `wiki/backlog/documento-publicado.md` defer rows (each surviving placeholder has a trigger) | real (grep + doc) |
| Generated types consumed directly; no hand-written snake→camel mapper added | `npx tsc --noEmit -p tsconfig.build.json` → exit 0 | real |
| Both reviewer agents APPROVE | `frontend-screen-reviewer` + `frontend-code-reviewer` reports on record | real (review) |

> TDD: write the failing vitest case first, then wire to green. Vitest cases are fixture-level (mocked
> queries) — they prove the consumer wiring, not the live endpoint; the live endpoints were already
> proven by M2 (distribution) and the existing export flow. The grep+backlog cross-check is real.

## ADR needed?

- [x] No durable decision — skip. (F4.1 consumes existing contracts; the parked-numerator decision is
  already governed by ADR-0042, and the defer rows live in `wiki/backlog/documento-publicado.md`.)
