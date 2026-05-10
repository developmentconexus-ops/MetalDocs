# Leakage probe · novo-template-identidade

Probe via `document.styleSheets` rule walk over each form element.

## Textarea (`#tpl-description.input.descriptionInput`)

| Selector | Source | Property | Value | Status |
|---|---|---|---|---|
| `input:not([type=...]), select, textarea` | `src/styles.css` (global) | padding | `0.5rem 0.75rem` | overridden by `.descriptionInput { padding: 10px }` |
| `.input` | `src/styles.css` (global) | height | `32px` | **CLOBBER** — overridden in fix by `.descriptionInput { height: auto; min-height: 72px }` |
| `.input` | `src/styles.css` (global) | padding | `0 10px` | overridden |
| `.descriptionInput` | CSS Module | font-size, padding, resize, height, min-height | local | ✓ |

## Name input (`#tpl-name.input.nameInput`)

| Selector | Source | Property | Value | Status |
|---|---|---|---|---|
| `input:not(...)` | `src/styles.css` global | padding | `0.5rem 0.75rem` (= 8px 12px) | applied — wins over `.input` due to source order |
| `.input` | `src/styles.css` global | height | `32px` | overridden by `.nameInput { height: 38px }` |
| `.input` | `src/styles.css` global | border | `0.8px solid var(--border)` | applied |
| `.nameInput` | CSS Module | height, font-size | local | ✓ |

No further leakage. The 8px 12px padding from the global `input:not(...)` rule is design-intentional and matches MetalDocs' canonical input look.

## Action

- Added `height: auto; min-height: 72px` to `.descriptionInput` to neutralize global `.input { height: 32px }` clobber on multi-line textareas.

## Known offenders confirmed

`.input { height: 32px }` in `src/styles.css` is a known offender for `<textarea>` usage. Future screens with textareas must override `height` and set `min-height` explicitly.
