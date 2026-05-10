# Global CSS leakage probe — novo-template-estrutura

**Method:** Pixel Parity Playbook §2 — for each interactive/form element, walk `document.styleSheets` and report rules whose selectors match the element. Flag any non-CSS-Module hits (i.e. `src/styles.css` globals).

## Elements probed

### Hidden file input — `<input type="file" className={fileInput}>`

| Source | Selector | Matched? | Rules applied | Action |
|---|---|---|---|---|
| `src/styles.css` | `.input` | ✗ (no `.input` class) | none | OK |
| `src/styles.css` | `input` (bare element selector) | ✗ (none in styles.css) | — | OK |
| Module | `.fileInput` | ✓ | visually-hidden (1×1, clip, `position: absolute`) | scoped, intentional |

No leakage — input is visually-hidden, no global form rules clobber width/height.

### Starting-point card — `<button role="radio">`

| Source | Selector | Matched? | Rules applied | Action |
|---|---|---|---|---|
| `src/styles.css` | `button` (bare) | ✗ (none) | — | OK |
| `src/styles.css` | `.btn`, `.btn-sm`, `.btn-ghost` | ✗ (cards don't use these classes) | — | OK |
| Module | `.startingPointCard` | ✓ | padding, border, gap, hover, focus-visible, selected | scoped |

UA defaults (font-family, color) overridden in module via `font-family: inherit; color: inherit;`. No leakage.

### Substituir button — `<button className="btn btn-sm btn-ghost">`

| Source | Selector | Matched? | Rules applied | Action |
|---|---|---|---|---|
| `src/styles.css` | `.btn`, `.btn-sm`, `.btn-ghost` | ✓ | intentional shared button styles | OK |
| Module | — | — | — | — |

Intentional reuse of global btn primitives.

### File row — `<div className={fileRow}>`

| Source | Selector | Matched? | Rules applied | Action |
|---|---|---|---|---|
| `src/styles.css` | div globals | none | — | OK |
| Module | `.fileRow`, `.fileIcon`, `.fileMeta`, `.fileName`, `.fileSize` | ✓ | scoped | scoped |

No leakage.

### Intro paragraph — `<p className="caption {intro}">`

| Source | Selector | Matched? | Rules applied | Action |
|---|---|---|---|---|
| `src/styles.css` | `.caption` | ✓ | font-size, line-height, color (`--text-muted`), margin: 0 | intentional |
| `src/styles.css` | `p` (browser default) | ✓ (UA) | margin top/bottom — overridden by `.caption { margin: 0 }` | reset by `.caption` |
| Module | `.intro` | ✓ | margin-bottom: var(--sp-5) | scoped |

Phase 4.5 finding (intro color override) resolved: `.intro` no longer sets `color`, defers to `.caption`'s `--text-muted`.

## Summary

No unscoped global rules clobber Step 3 elements. Known offenders from skill table not encountered here:
- `input { width: 100% }` — N/A (file input visually-hidden, not styled by global).
- `label span { uppercase }` — N/A (no `<label>` used).
- `p` UA margin — reset by `.caption`.

Status: clean.
