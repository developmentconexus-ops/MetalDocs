# Phase 4 — Behavior verify

## tsc

`pnpm tsc --noEmit -p tsconfig.build.json` — clean for all Templates-touched files (`TemplatesListPage.tsx`, `useTemplatesQuery.ts`, `TemplateCard.tsx`).

Pre-existing errors out of scope (auth tests, useAreasQuery v5 API drift, Rail.tsx, NewDocumentWizardPage, LibrarySidebar). None touch this work.

## tests

`pnpm test --run src/features/templates`

- 8 of 9 files pass · 52 of 56 tests pass
- 4 failures all in `template-author-page-convergence.test.tsx` (TemplateAuthorPage, raw `fetch()` in `catalog.ts` — jsdom URL parsing issue, pre-existing, separate page)
- Zero failures touch TemplatesListPage / TemplateCard / useTemplatesQuery

## Smoke trace (live browser /templates-v2)

| Step | Expected | Observed |
|---|---|---|
| Load `/templates-v2` | Hero + tabs + 4 cards from API | OK — 4 templates rendered |
| Tab counts | Real counts from data | Todos·4 / Publicados·3 / Rascunhos·1 / Arquivados·0 |
| Click "Rascunhos" tab | List filters to 1 draft | OK — 1 card shown, aria-label "Abrir template teste" |
| Click "Arquivados" | Empty state shown | OK — "Nenhum ... encontrado" rendered |
| Card click | Navigate to `/templates-v2/:id/versions/:n` | OK — navigated to `/templates-v2/<uuid>/versions/1` |
| Card focus | role=button + tabIndex=0 + visible focus ring | OK — focusable, aria-label correct |

## Console / network

No errors during smoke trace. Single `GET /api/v2/templates` request resolved 200 with 4 templates.

## Verdict

Phase 4 passes. Pre-existing test/tsc debt unrelated to this work documented above; nothing introduced.
