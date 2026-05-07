# Subagent prompt — Phase 3b Style port

You are a subagent dispatched in a fresh git worktree to perform Phase 3b (Style port) of the MetalDocs screen-implementation workflow. Phase 3a produced the TSX skeleton + empty CSS Module. Your job is to port styles from the reference HTML to the CSS Module using design tokens only.

## Inputs (substitute at dispatch time)

- Worksheet path: `frontend/apps/web/design-source/<SLUG>/IMPLEMENTATION.md`
- CSS Module to fill: `frontend/apps/web/src/features/<DOMAIN>/pages/<PAGENAME>.module.css`
- Tokens file: `frontend/apps/web/src/styles/tokens.css`
- Shared tokens (if any): `@metaldocs/shared-tokens`
- Reference HTML: full inline (substituted at dispatch — see below)
- Reference screenshot: `frontend/apps/web/design-source/<SLUG>/<SLUG>.png`

## Reference HTML (substituted at dispatch)

```html
<!-- DISPATCHER: paste full contents of design-source/<SLUG>/<SLUG>.html here -->
```

## Steps

1. **Extract every CSS rule** from the reference HTML — `<style>` block + any inline `style=`. Build a flat list of (selector, property, value) tuples.

2. **Build the token map** in worksheet §3b.1. For every value (color, spacing, radius, font-size, font-weight, line-height, shadow):
   - Find the matching token in `frontend/apps/web/src/styles/tokens.css`. Record `--existing-token`.
   - If no match within ±5% (color hue, spacing px), it is a **missing token**.

3. **Add missing tokens.** For each missing-token row:
   - Add the token to `frontend/apps/web/src/styles/tokens.css`.
   - Commit: `feat(tokens): add <--token-name> for <SLUG>`. One commit covers all missing tokens for this screen if they land together.
   - Update worksheet to point at the new token.

4. **Port the rules** to the CSS Module. Rules:
   - Use `var(--token)` for every value. NO raw hex. NO raw px for spacing. Raw px allowed only for `1px` borders and `0` values.
   - Keep selectors as CSS Module class references.
   - Preserve the order of rules from the reference for diff readability.

5. **Visual diff — measured, not eyeballed.** Run dev server, open route, capture screenshots at 1440 / 1024 / 375. Then run the **Computed-Style Parity Loop** (see "Pixel Parity Playbook" below). Eyeballing screenshots misses spec violations that show up only as `marginBottom: 0px` vs `8px` in computed style. Always inspect computed numbers, not pixels.

6. **Global CSS leakage scan (HARD).** Before declaring done, for every interactive element on the page (`input`, `select`, `textarea`, `button`, `label span`, `p`, `h2`, `h3`), enumerate its `getMatchedCSSRules` equivalent (iterate `document.styleSheets`) and confirm no global rule from `src/styles.css` is overriding the CSS Module rule. Common offenders documented in Pixel Parity Playbook §2. Any leak → fix in CSS Module via `width: auto; padding: 0; border: none; background: none;` resets, OR scope the global rule narrower in `styles.css` (separate commit). Log finding in worksheet §3b.

7. Update worksheet §3b items with `[x]`. Mark "User approved screenshot diff" as `[ ]` — only the user marks this one, NOT you.

## Pixel Parity Playbook

Use these `mcp__Claude_Preview__preview_eval` patterns. The numbers, not screenshots, are the truth.

### §1 Computed-style snapshot per region

```js
(() => {
  const el = document.querySelector('<selector>');
  const cs = getComputedStyle(el);
  const r = el.getBoundingClientRect();
  return {
    box: {w: r.width, h: r.height, x: r.x, y: r.y},
    spacing: {mt: cs.marginTop, mb: cs.marginBottom, mr: cs.marginRight, ml: cs.marginLeft,
              pt: cs.paddingTop, pb: cs.paddingBottom, pr: cs.paddingRight, pl: cs.paddingLeft},
    type: {fs: cs.fontSize, fw: cs.fontWeight, lh: cs.lineHeight, ff: cs.fontFamily, tt: cs.textTransform, ls: cs.letterSpacing},
    color: {c: cs.color, bg: cs.backgroundColor, b: cs.border, br: cs.borderRadius},
    layout: {display: cs.display, flex: cs.flex, flexDir: cs.flexDirection, gap: cs.gap, ai: cs.alignItems, jc: cs.justifyContent, of: cs.overflow}
  };
})()
```

Compare impl-side output against design-source served on the design preview server. Any mismatch in spacing/type/layout → fix.

### §2 Global-rule leakage probe

```js
(() => {
  const el = document.querySelector('<selector>');
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

Known offenders in `src/styles.css` to actively probe for on every form-bearing screen:

| Selector | Effect | Reset in CSS Module |
|---|---|---|
| `input, select, textarea` | `width: 100%; border; background; padding` — clobbers checkboxes, radios, segmented controls | `width: auto; padding: 0; border: none; background: none; border-radius: 0;` |
| `button, input, select, textarea` | `font: inherit` — fine, but be aware |  — |
| `label span` (legacy) | uppercase + tiny | `text-transform: none; letter-spacing: normal; font-size: inherit; color: inherit;` on the local span |
| browser default `p` | `margin: 1em 0` — adds visual height inside flex | `.<scope> p { margin: 0; }` |
| browser default `ol, ul` | `padding-inline-start: 40px` | reset to design value |

If you find a NEW global-rule leak, add a row to this table in the skill (separate commit) so the next run catches it.

### §3 Parent → child inheritance traps

Primitives like `SelectableCard`, `Button`, `Modal` apply `gap`, `padding`, `font`, `align-items` that combine with the child's own rules. The child often gets *double* spacing.

For every primitive used by the page, snapshot:
- Primitive's `gap`, `padding`, `flex-direction`.
- Inside child, what spacing primitives add. Then decide which side owns spacing — usually primitive owns container padding/gap, child owns nothing on its outermost element.

If child has its own `margin`/`padding` on root, that often means double spacing. Move spacing to ONE side.

### §4 Specificity loop (when a fix appears not to apply)

If you set `.foo { line-height: 1 }` and computed style still shows `normal`, the rule did not apply. Causes:
- Class is global (`.kicker` from styles.css) but selector wraps it as a CSS Module class. Use `:global(.kicker)`.
- Higher-specificity rule wins. Add specificity (`.parent .foo`, or `:global(...)`).
- Source-order tie. Add `!important` only as last resort and document why.

Verify by re-running §1 snapshot. Numbers must change.

### §5 Reference parity check

For each region in the design HTML:

1. Render `<slug>.html` in the design preview server, run §1 snapshot.
2. Render the impl page, run §1 snapshot.
3. Diff fields. Any spacing/type field that differs is a defect — fix BEFORE declaring Phase 3b done.

The §5 diff is what proves visual parity. A screenshot eye-test does not.

## Output

- Tokens commit (if any): `feat(tokens): add <list> for <SLUG>`.
- Style commit: `feat(<DOMAIN>): style port for <SLUG>`.
- Report: token map summary, missing tokens added, dev URL, **§5 reference parity diff (numerical, per region)**, **§2 leakage probe results**.

## Output

- Tokens commit (if any): `feat(tokens): add <list> for <SLUG>`.
- Style commit: `feat(<DOMAIN>): style port for <SLUG>`.
- Report: token map summary, missing tokens added, dev URL, side-by-side notes.

## Hard rules

- No raw hex outside `tokens.css`.
- No raw px for spacing in CSS Module.
- If a design value has no clear token match, STOP — log to worksheet Open Questions, ask user whether to add a new token or use the closest existing.
- Do NOT mark "user approved screenshot diff" yourself.
