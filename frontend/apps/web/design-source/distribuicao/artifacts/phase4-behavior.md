# Phase 4 — Behavior Verify

> **Slug:** distribuicao
> **Completed:** 2026-05-09

## tsc

Pre-existing errors only (not introduced by this screen):
- `useAreasQuery.ts` — QueryFunction overload mismatch (pre-existing)
- `Rail.tsx:68` — `string | undefined` not assignable to `string` (pre-existing)

Zero new type errors from distribuicao files.

## Tests

```
Test Files  8 failed | 46 passed (54)
Tests       17 failed | 238 passed (255)
```

All 17 failures are pre-existing in unrelated features:
- `useAuthSession.returnTo.test.tsx` — auth session
- `RouteAdminPage.test.tsx` — approval admin
- `useDocumentPdfStatus.test.ts` — PDF status hook
- `DocumentsHubView.edit-button.test.tsx` — hub edit button
- `DocumentEditorPage.test.tsx` — editor page
- `useDocumentComments.load.test.tsx` — comments
- `template-author-page-convergence.test.tsx` — templates

No distribuicao test files exist (all data is mock, no query hooks to test).

## Manual smoke trace

| Step | Expected | Observed |
|---|---|---|
| Navigate to `documents/:id/distribution` | DocumentDetailLayout renders with centered tab strip, Distribuição tab active | ✓ |
| Tab strip | "Documento" and "Distribuição" tabs centered, active tab has brand underline | ✓ |
| Hero | DocumentHero renders: breadcrumb, DocRefCard, badges (PR-EHS-014·v3.2, VENCE EM 8 DIAS, §03.05·FANOUT), title, subtitle, 4 disabled CTAs | ✓ |
| KPI strip | Alvos totais: 248, Reconheceram: 156, Apenas leram: 26, Pendentes: 66, Em atraso: 12 | ✓ |
| DonutCard | 63% derived from MOCK_DISTRIBUTION (156/248), legend shows 156/26/66 | ✓ |
| RecipientsCard — pending tab default | Shows 2 rows (non-overdue pending), pagination "Mostrando 2 de 11" | ✓ |
| Checkboxes | 12 total (1 header + 11 rows), readOnly removed | ✓ |
| 4 hero CTAs | aria-disabled="true", title="Em breve", pointer-events none | ✓ |
| Switch to Documento tab | DocumentPublishedPage renders with same DocumentHero shell | ✓ |
| Switch back to Distribuição | State preserved, distribution page returns | ✓ |

## Console errors

None. Only pre-existing `No HydrateFallback` warnings (React Router, unrelated).
