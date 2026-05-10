# Phase 4 — Behavior · novo-template-permissoes

## tsc

`pnpm tsc --noEmit -p tsconfig.build.json` — templates feature clean. Pre-existing errors in documents/shell/auth unrelated.

## Tests

Not added — Light tier equivalent coverage (no new query hooks, reducer extension is pure toggle, no shared primitive).

## Smoke trace (live)

| # | Action | Expected | Observed |
|---|---|---|---|
| 1 | Step 1 → Genérico → Avançar | step=2 | ✓ |
| 2 | Type name → Avançar | step=3 | ✓ |
| 3 | Em branco → Avançar | step=4 | ✓ |
| 4 | Land on step 4 | h2 "Quem pode usar este template?"; segmented control (3 tabs); "Por funções" active; 6 role cards rendered; Avançar disabled | ✓ |
| 5 | Click any role card | card brand-pale bg + ✓ icon; Avançar enabled; coverage count = 1 | ✓ |
| 6 | Click same card again | card deselected (toggle); Avançar disabled again | ✓ |
| 7 | Select 3 roles | coverage "3 perfis selecionados · cobre ~N usuários ativos" | ✓ |
| 8 | Switch to "Por área" tab | area grid (3-col, 6 cards); coverage reset to areas; Avançar enabled (no gate for areas) | ✓ |
| 9 | Toggle area cards | selected card brand-pale + ✓; coverage text switches to "áreas selecionadas" | ✓ |
| 10 | Switch to "Todos" tab | company-wide banner appears; coverage block hidden; Avançar enabled | ✓ |
| 11 | Voltar → step 3 | step=3; state preserved | ✓ |
| 12 | Avançar back to step 4 | permissionsMode + selectedRoleIds + selectedAreaIds preserved | ✓ |

## Console

No errors.

## Status

Behavior verified. Proceed to Phase 4.5 reviewer.
