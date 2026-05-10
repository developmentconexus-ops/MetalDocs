# Phase 4 — Behavior · novo-template-estrutura

## tsc

`pnpm tsc --noEmit -p tsconfig.build.json` — templates feature clean. Pre-existing errors in documents/shell unrelated.

## Tests

Not added — Light tier, reducer extension is pure assignment, no new query hooks, no shared primitive.

## Smoke trace (live)

| # | Action | Expected | Observed |
|---|---|---|---|
| 1 | Step 1 → Genérico → Avançar | step=2 | ✓ |
| 2 | Type "Teste de template" → Avançar | step=3 | ✓ |
| 3 | Land on step 3 | "Estrutura do template" h2; 2 cards rendered; footer "Escolha um ponto de partida"; Avançar disabled | ✓ |
| 4 | Click "Em branco" | aria-pressed=true; brand-pale bg + ✓; Avançar enabled; footer "Pronto para avançar" | ✓ |
| 5 | Click ".docx" card | aria-pressed=true on docx; aria-pressed=false on blank; clears blank state | ✓ |
| 6 | Inject mock file (eval) | File row appears: DOCX badge + filename + "147 KB" + Substituir | ✓ |
| 7 | Voltar → step 2 | name/desc state preserved; back to identity | ✓ |
| 8 | Forward to step 3 again | startingPoint + selectedDocx state preserved | ✓ |

## Console

No errors.

## Status

Behavior verified. Proceed to Phase 4.5.
