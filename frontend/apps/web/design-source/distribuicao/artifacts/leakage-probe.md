# Leakage Probe — distribuicao

**Date:** 2026-05-08
**Route:** `/documents/PR-EHS-014/distribution`
**Viewport:** 1440×900

## Probe selector

```js
(() => {
  const el = document.querySelector('input[type=search], .searchInput, input[placeholder*="Filtrar"]');
  if (!el) return 'not found';
  const matched = [];
  for (const sheet of document.styleSheets) {
    let rules; try { rules = sheet.cssRules; } catch(e) { continue; }
    for (const r of rules) {
      if (!r.selectorText) continue;
      try { if (el.matches(r.selectorText)) matched.push({sel: r.selectorText, css: r.style.cssText}); } catch(e) {}
    }
  }
  return matched;
})()
```

## Element found

`input[type=search]` with `placeholder="Filtrar por nome, área, cargo…"` — the SearchBar primitive rendered inside RecipientsCard's filter bar.

Class applied: `._input_1y743_19` (SearchBar.module.css `.input`)

## Matched rules (in cascade order)

| # | Selector | CSS applied |
|---|----------|-------------|
| 1 | `*` | `box-sizing: border-box` |
| 2 | `button, input, select, textarea` | `font: inherit` |
| 3 | `input:not([type="checkbox"]):not([type="radio"]):not([type="file"]):not([type="range"]):not([type="color"]):not([type="image"]), select, textarea` | `width: 100%; border-radius: var(--r-2); border: 1px solid var(--border); background: var(--surface); color: var(--text-soft); padding: 0.5rem 0.75rem;` |
| 4 | `._input_1y743_19` | `border: 0px; background: transparent; outline: none; color: var(--text); font-size: 0.85rem; width: 100%;` |

## Computed styles on input

```
borderWidth:  0.8px  (≈1px — global rule winning over SearchBar reset)
borderStyle:  solid
borderColor:  rgb(230, 220, 220)
```

## Assessment

**Pre-existing issue in SearchBar.module.css**, not introduced by distribuicao styles.

The global selector (rule #3) uses multiple `:not()` pseudo-classes which increase specificity above a single class selector (rule #4). As a result `border: 1px solid var(--border)` from the global leaks into the SearchBar input despite `border: 0` in SearchBar.module.css.

**Visual impact:** Minor — the input has a thin 0.8px border. The parent `.root` label wrapper already has a styled border (`1px solid var(--border)`) so the double-border adds a faint artefact visible only on close inspection.

**RecipientsCard-specific note:** Phase 3a used the `SearchBar` primitive (not a raw `<input>`), so no additional reset was needed in `RecipientsCard.module.css`. The leakage originates in SearchBar itself.

## Fix recommendation (deferred)

In `SearchBar.module.css`, increase specificity of `.input` reset:

```css
label.root > .input {
  border: 0;
  background: transparent;
  padding: 0;
}
```

Or add `!important` to the border reset as a targeted workaround. Tracked in backlog.
