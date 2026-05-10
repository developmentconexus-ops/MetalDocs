# Phase 4.5 — reviewer resolutions · novo-template-estrutura

Verdict: **REQUEST CHANGES** → resolved.

## Major findings (5/5 fixed)

| # | Finding | Resolution | File:line |
|---|---|---|---|
| 1 | `.intro { color: var(--text-soft) }` overrides `.caption`'s `--text-muted` | Removed `color` declaration; `.caption` now governs color | `StepStructure.module.css:3-5` |
| 2 | DOCX badge inline styles (`fontSize:8.5, color:var(--brand)…`) bypass CSS Module | Extracted to `.thumbnailDocxLabel` class | `StepStructure.module.css:67-72`, `StepStructure.tsx:74` |
| 3 | `.startingPointGrid` no `@media` — cramped at 375px | Added `@media (max-width: 640px) { grid-template-columns: 1fr }` | `StepStructure.module.css:18-22` |
| 4 | `aria-pressed` (toggle semantics) on mutex selection | Wrapped in `role="radiogroup"` + `role="radio"`/`aria-checked` per card | `StepStructure.tsx:67-72,93-99` |
| 5 | Missing standalone `parity-diff.md` + `leakage-probe.md` | Created — see `artifacts/parity-diff.md`, `artifacts/leakage-probe.md` | — |

## Minors

Deferred to backlog (existing `wiki/backlog/novo-template-wizard.md` rows cover Step 3 work).

## Verify

- `pnpm tsc --noEmit -p tsconfig.build.json` — templates feature clean (pre-existing errors in documents/auth/shell unrelated, see `phase4-behavior.md`).
- Behavior trace unchanged — no logic touched.
- Visual parity preserved (intro color identical: `.caption` already used `--text-muted`).

## Status

All Major findings resolved. Step 3 mergeable.
