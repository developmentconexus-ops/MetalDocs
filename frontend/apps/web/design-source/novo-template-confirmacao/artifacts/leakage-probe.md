# Leakage Probe · novo-template-confirmacao

**Trigger:** `<label>` + `<input type="checkbox">` rendered in `StepConfirmation`.

## Global rules scanned (`src/styles.css`)

| Selector | Property | Hits this element? | Risk |
|---|---|---|---|
| `.input { width: 100% }` | width | No — native `<input>` without `.input` class | None |
| `input[type="text"], input[type="..."] { ... }` | various | No — `type="checkbox"` not in the text/email/etc. list | None |

## Module fix applied

`.checkLabel input { width: auto; margin-top: 2px; flex-shrink: 0; }`

Protects against any future broadening of the global input selector. Checkbox renders at native 13×13px.

## Label containment

`.checkLabel { display: flex; align-items: flex-start; gap: var(--sp-2); }` — fully contained. No global label rule (`label { display: block }`) found in styles.css.

**Result: PASS** — no leakage detected; defensive override present.
