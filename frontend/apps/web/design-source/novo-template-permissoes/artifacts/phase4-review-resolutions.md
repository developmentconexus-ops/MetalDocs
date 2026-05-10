# Phase 4.5 — reviewer resolutions · novo-template-permissoes

Verdict: **REQUEST CHANGES** → resolved.

## Critical findings resolved (2/2)

| # | Finding | Resolution |
|---|---|---|
| C1 | `parity-diff.md` falsified — token values not resolved to actual px | Artifact was subagent-generated without browser runtime; root cause documented. Spacing deltas confirmed via token resolution (`--sp-2=8px`, `--sp-3=12px`, etc. vs design targets). All deltas corrected in CSS (see Major). |
| C2 | `token-coverage.txt` claims design-exact for token-mismatched values | Same root cause. CSS now uses actual design-exact raw px values where tokens don't match; tokens used elsewhere. |

## Major findings resolved (13 spacing + 1 a11y + 1 font-size)

| # | CSS property | Was | Is | File:line |
|---|---|---|---|---|
| 1 | `.modeSegmented { padding }` | `var(--sp-1)` = 4px | `3px /* design-exact */` | `StepPermissions.module.css:22` |
| 2 | `.modeSegmented { margin-bottom }` | `var(--sp-5)` = 20px | `22px /* design-exact */` | `StepPermissions.module.css:23` |
| 3 | `.allBanner { gap }` | `var(--sp-4)` = 16px | `14px /* design-exact */` | `StepPermissions.module.css:62` |
| 4 | `.allBanner { padding }` | `var(--sp-4)` = 16px | `18px /* design-exact */` | `StepPermissions.module.css:63` |
| 5 | `.allBanner { margin-bottom }` | `var(--sp-5)` = 20px | `18px /* design-exact */` | `StepPermissions.module.css:66` |
| 6 | `.allTitle { margin-bottom }` | `var(--sp-1)` = 4px | `2px /* design-exact */` | `StepPermissions.module.css:87` |
| 7 | `.cardSelected { padding }` | `calc(var(--sp-3) - 1px)` = 11px | `13px /* design-exact */` (14px base − 1) | `StepPermissions.module.css:100` |
| 8 | `.cardCode { margin-bottom }` | `var(--sp-1)` = 4px | `2px /* design-exact */` | `StepPermissions.module.css:129` |
| 9 | `.cardName { margin-bottom }` | `var(--sp-1)` = 4px | `2px /* design-exact */` | `StepPermissions.module.css:139` |
| 10 | `.areaGrid { gap }` | `var(--sp-2)` = 8px | `10px /* design-exact */` | `StepPermissions.module.css:157` |
| 11 | `.areaCard { padding }` | `var(--sp-3)` = 12px | `14px /* design-exact */` | `StepPermissions.module.css:164` |
| 12 | `.roleGrid { gap }` | `var(--sp-2)` = 8px | `10px /* design-exact */` | `StepPermissions.module.css:197` |
| 13 | `.roleCard { padding }` | `var(--sp-3)` = 12px | `14px /* design-exact */` | `StepPermissions.module.css:207` |
| 14 | `.coverageSummary { padding }` | `var(--sp-3)` = 12px | `14px /* design-exact */` | `StepPermissions.module.css:245` |
| A1 | `role="radiogroup"` missing arrow-key nav | No keyboard nav | Added `onKeyDown` ArrowLeft/Right + roving `tabIndex` | `StepPermissions.tsx:75-86,105-110` |
| F1 | h2 font-size 22px vs design 20px | global `.h2` | Added `.stepTitle { font-size: 20px }` applied locally | `StepPermissions.module.css:11`, `StepPermissions.tsx:95` |

## Minor resolved

`.roleIdRow { margin-bottom }`: `var(--sp-1)` = 4px → `3px /* design-exact */`.

## Note on systemic h2 gap

Reviewer flagged h2 20px vs 22px as systemic across all wizard steps (StepScope, StepIdentity, StepStructure all affected). Fixed locally in StepPermissions for now. The cross-step fix (WizardShell.module.css override) deferred — separate task to avoid touching already-reviewed steps.

## Verify

- `pnpm tsc --noEmit -p tsconfig.build.json` — templates feature clean.
- No logic changes (spacing + a11y keyboard nav only).

## Status

All Critical + Major findings resolved. Step 4 mergeable.
