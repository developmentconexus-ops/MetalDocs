# Phase 4 — Behavior verify · novo-template-identidade

## tsc

```
pnpm tsc --noEmit -p tsconfig.build.json
```

**Templates feature:** clean — `grep "features/templates"` on tsc output returned **no matches**.

Pre-existing errors exist in `features/documents` (NewDocumentWizardPage, useAreasQuery, DocumentPublishedPage) and `features/shell/Rail.tsx`. **Unrelated to this change** — present on baseline before Step 2 work began. Documented to backlog of doc-wizard rollout, not a regression of this PR.

## Tests

`pnpm test` not run for this Light tier change — no new test files added (no shared primitive, no new query hook, no logic-heavy reducer branch). Reducer extension is pure assignment; trivial.

## Manual smoke trace

Performed during inline Phase 3 implementation:

| # | Action | Expected | Observed |
|---|---|---|---|
| 1 | Nav to `/templates/new?step=1`, pick Generic, advance | step=2, kicker "Etapa 2 de 5 · Template genérico" | ✓ |
| 2 | Step 2 generic recap | "GEN" chip, "Genérico", "Aplicável a qualquer perfil" | ✓ |
| 3 | Code preview generic | `TPL-GEN-XXX`, brand color, dashed brand-soft border | ✓ |
| 4 | Type "ab" in name | Avançar disabled, footer label "Informe o nome para continuar" | ✓ |
| 5 | Type "abc" | Avançar enabled, footer "Pronto para avançar" | ✓ |
| 6 | Click "Trocar" | step=1, scope still "generic" preserved | ✓ |
| 7 | Switch to Profile scope, pick DC, advance | step=2, kicker shows "Perfil DC — …" | ✓ |
| 8 | Profile recap | DC code-chip + name + família line + Trocar | ✓ |
| 9 | Code preview profile (after fix) | `TPL-DC-XXX` uppercase | ✓ |
| 10 | Voltar from step 2 | step=1, profile still selected | ✓ |
| 11 | Direct nav `?step=2` no scope | defensive guard → step=1 | ✓ |
| 12 | URL sync | `?step=N` reflects state.step on every change | ✓ |

## Console

No errors. Stale Vite HMR errors at StepScope L89 from earlier failed-edit cycle were cleared by hard reload (`window.location.reload()`); no recurrence with current source.

## Status

Behavior verified. Proceed to Phase 4.5 review.
