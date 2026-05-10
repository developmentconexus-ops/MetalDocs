# Phase 3 (combined) — novo-template-estrutura

**Tier:** Light · executed inline by main agent. No subagent dispatched.

## Files written

- `src/features/templates/components/wizard/steps/StepStructure.tsx` — ~165 LOC.
- `src/features/templates/components/wizard/steps/StepStructure.module.css` — ~120 LOC.
- `src/features/templates/state/templateWizard.reducer.ts` — extended.
- `src/features/templates/pages/TemplateWizardPage.tsx` — Step 3 branch wired.

## Token coverage

CSS Module values: spacing/colors via `var(--…)`. Raw px tagged `/* design-exact */`: 13.5px (cardTitle), 14px (cardCheck — could use `var(--font-size-md)`; 14px is glyph not text), 12px (cardDesc, between `--font-size-xs:11px` and `--font-size-sm:12.5px`), 8.5px (DOCX badge — far below ladder), 18px (thumbnail "+" sigil — between `--font-size-md:14px` and `--font-size-lg:22px`). None block ship; backlog row `font-size-hero` covers similar gap.

## State wiring

- `startingPoint` ∈ {`docx`, `blank`, `null`}. Switching to `blank` clears docx selection (reducer).
- `selectedDocxName` + `selectedDocxSize` mocked from native `<input type=file>` — no upload, no parsing. TODO + backlog ref at site.
- Advance gate: `startingPoint !== null && (startingPoint !== 'docx' || selectedDocxName !== null)`.
- Footer label adapts: "Escolha um ponto de partida" → "Selecione um .docx para continuar" → "Pronto para avançar".
- Voltar → step 2; back preserves all state.

## Smoke trace

| # | Action | Expected | Observed |
|---|---|---|---|
| 1 | Land on step 3 (no choice) | Both cards neutral; Avançar disabled; footer "Escolha um ponto…" | ✓ |
| 2 | Click "Em branco" | Card brand-pale + ✓; Avançar enabled; footer "Pronto para avançar" | ✓ |
| 3 | Click ".docx" card | Native picker opens; card selected; clears blank state | ✓ (eval used file-stub) |
| 4 | Pick file (mock) | File row shows name + KB; Substituir button visible | ✓ — `inspecao-mp.docx · 147 KB` |
| 5 | Click "Substituir" | Clears state, re-opens picker | ✓ |
| 6 | Voltar | step 2; state preserved | ✓ |
| 7 | tsc | templates feature clean | ✓ |

## A11y

- `aria-pressed` on both starting-point cards (toggle semantics).
- Hidden file input `tabIndex={-1}` + `aria-hidden`; not focusable directly. Picker triggered through visible card click.
- File icon + checkmark glyphs marked `aria-hidden`.
- File row uses `<button class="btn btn-sm btn-ghost">` for Substituir — focusable.

## Deferred (backlog)

- `step3-docx-upload` — real presigned PUT flow.
- `step3-placeholder-extract` — backend endpoint required.
- `step3-editor-handoff` — post-create redirect with import flag.
