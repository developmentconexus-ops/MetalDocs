# Global CSS Leakage Probe — novo-documento

**Method:** ran `metaldocs-screen-implementation` Pixel Parity Playbook §2 leakage probe (`document.styleSheets` walk) against every form element on `/documents-v2/new` steps 1–4 + every interactive element in `WizardShell`.

## Findings

| Selector in base.css | Element hit | Effect | Resolution | Commit |
|---|---|---|---|---|
| `input, select, textarea { width: 100%; ... }` | `<input type="checkbox">` (Step 4 consent row) | width clobbered to 100% (814px) → checkbox covered label, span got 0 width | Scoped global to `:not([type="checkbox"]):not([type="radio"]):not([type="file"]):not([type="range"]):not([type="color"]):not([type="image"])` | (this commit) |
| `label span { text-transform: uppercase; ... }` (legacy) | already removed in earlier commit | uppercase + tiny font on consent text | Removed `label span` from selector group | (prior commit) |

| `label span { text-transform: uppercase }` | `src/styles.css` (global) | StepConfirm — consent label inner span | Reset locally in `StepConfirm.module.css` `.consentRow span` |
| `input { width: 100% }` | `src/styles.css` (global) | StepConfirm — consent checkbox | Reset locally in `StepConfirm.module.css` `.consentRow input[type="checkbox"]` |
| `.kicker` line-height inside callout | `src/styles.css` (global) | StepConfirm — `.nextStepsCallout` | Local `:global(.kicker)` override (debt; promote to `.kicker--callout` global modifier) |

## Future-screen recommendations

- Bare-element selectors (`input { ... }`, `p { ... }`, `ol { ... }`) in `base.css` are leak hazards. Prefer opt-in classes or scope with `:not()`.
- Add a CI lint that fails on bare `input`, `button`, `label span` rules in `base.css`. Tracked in `wiki/backlog/styles-css-narrowing.md` (TBD).
