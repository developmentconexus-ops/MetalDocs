# Feature F4.1 — Publicado stubs (coverage + PDF wiring, defer the rest)

> **Milestone:** 4 — Documento Publicado completion + Documento Obsoleto  ·  **Folder:** `f4.1-publicado-stubs`
> **Status:** Planning

This is the feature's **execution plan** — the "how" `milestone.md` left out. Contract is `./spec.md`.

## Source

- Milestone spec row (F4.1): wire Cobertura KPI + coverage aside to `GET /documents/{id}/distribution`
  (denominator only; numerator parked ADR-0042); enable Baixar PDF via `exportPDF`; convert the 6
  unbacked fields to defer-with-trigger backlog rows. Acceptance: live count rendered (no fabricated %),
  PDF button enabled + issues real export, every surviving "em breve" backed by a backlog defer row,
  `tsc` clean, both reviewers APPROVE.
- Governing-spec reference: mission §5 inventory row 11; mission §7 M4/F4.1; mission §8 (no silent stub).

## Plan

**Files touched (frontend only — no Go change):**

1. `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx` — the only component edit.
2. `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.test.tsx` — new vitest cases (TDD first).
3. `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.module.css` — only if the
   coverage card markup change needs a class tweak (drop the `%`/bar). Prefer reusing existing classes.
4. `wiki/backlog/documento-publicado.md` — promote the 6 unbacked fields to explicit defer-with-trigger rows.

**Reuse (no new client/hook — built in M2 / existing):**
- `useDistributionSummaryQuery(documentId)` (`queries/useDistributionSummaryQuery.ts`) → `{ total_targets }`.
- `exportPDF(documentID, { paper_size:'A4' })` (`api/exports.ts`) → `{ signed_url, cached, size_bytes, … }`.
- M2 precedent for error semantics: `total_targets ?? EM_DASH`, EM_DASH on `isError` (see `DocumentDistributionPage.tsx:75`).
- PDF flow precedent: `ExportMenu.tsx` (POST → `window.open(signed_url,'_blank','noopener')`; 429 → rate-limited state).

**Task list (TDD order):**

1. **TDD — coverage render.** Add failing vitest cases to `DocumentPublishedPage.test.tsx`: (a) summary
   query mocked → `total_targets: 12` ⇒ Cobertura cell/aside shows "12" obligated count, no `%` text,
   no `progressbar` role element; (b) summary `isError` ⇒ EM_DASH, no `%`. Run → red.
2. **Implement coverage.** In `DocumentPublishedPage.tsx`: call `useDistributionSummaryQuery(documentId)`;
   replace the Cobertura KPI placeholder (`kpiValuePlaceholder` "em breve", ~516) with the count
   (EM_DASH on error/missing) + parked-numerator hint; rewrite the coverage `<aside>` (~611-630) to show
   the obligated count + "leitura em acompanhamento (ADR-0042)" label and **remove** the fabricated
   `—%` value + `role="progressbar"` bar. Keep "abrir fanout →" navigation. Run → green.
3. **TDD — PDF.** Add failing cases: (c) "Baixar PDF" button not `aria-disabled`; click ⇒ `exportPDF`
   spy called `{paper_size:'A4'}` + `window.open` called with `signed_url`; (d) `exportPDF` rejects
   `{status:429}` ⇒ graceful rate-limited state, no unhandled rejection, button usable again. Run → red.
4. **Implement PDF.** Replace the `aria-disabled` button (380-383) with a handler that mirrors
   `ExportMenu.handlePDF`: local `useState` status (idle/pending/rate_limited), `exportPDF` → `window.open`,
   429 handling, pending disables the button + shows progress affordance. Run → green.
5. **TDD + implement — Páginas + Tamanho** (runtime-truth correction, see spec Interview #3). Failing
   vitest cases: page_count `1` ⇒ "1" rendered (not "em breve"); page_count `null` ⇒ honest absent/"—";
   file_size `1024` ⇒ formatted (e.g. "1 KB"); `null` ⇒ absent. Implement: read
   `detail.current_revision_page_count` + `current_revision_file_size_bytes` (already on
   `useDocumentDetailQuery` data); replace the Páginas KPI placeholder (~526) and the Tamanho fact (~593)
   with the formatted value / absent state. Add a small `formatFileSize(bytes)` helper (bytes→KB/MB,
   pt-BR) — prefer an existing util if one exists in `lib/`. Run → green.
6. **Defer rows (the genuinely-unbacked 4).** In `wiki/backlog/documento-publicado.md`: (a) **correct the
   stale claims** that page-count/file-size have no backend — they ship on `DocumentResponse`, now wired;
   (b) add/confirm an explicit defer-with-trigger row for each of the 4: Próxima revisão (trigger:
   review-due field on document/CD), Classificação (confidentiality field on `DocumentResponse`),
   Documentos relacionados (relationship model + endpoint), Comentários (display-comments architecture
   decision). Leave each on-screen state as an honest empty/absent state.
7. **Cross-check gate.** `grep -n "em breve" DocumentPublishedPage.tsx` → every surviving hit maps to a
   backlog defer row (Cobertura, PDF, Páginas, Tamanho hits are gone). Run
   `npx tsc --noEmit -p tsconfig.build.json` and
   `npx vitest run src/features/documents/pages/DocumentPublishedPage` → both green.
7. **Review.** Dispatch `frontend-screen-reviewer` then `frontend-code-reviewer`; resolve nits at root.

**Test strategy:** vitest at fixture level (mocked queries/spies) proves consumer wiring; the live
endpoints were already proven by M2 (distribution) and the existing export flow. The grep+backlog
cross-check is the real proof for the no-silent-stub bar.

**Ordering rationale:** coverage first (pure read, lowest risk), PDF second (adds local state +
side-effect), defers third (doc-only), gate+review last.

## Execution notes

_(filled during execution — model choices, deviations with rationale)_
