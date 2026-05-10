# Subagent prompt — Phase 3b Style port (Heavy tier)

Fresh git worktree subagent. Phase 3a (main agent inline) produced the TSX skeleton + empty CSS Module. Port styles using tokens only.

## Inputs

- Worksheet: `frontend/apps/web/design-source/<SLUG>/IMPLEMENTATION.md`
- CSS Module: `frontend/apps/web/src/features/<DOMAIN>/pages/<PAGENAME>.module.css`
- Tokens: `frontend/apps/web/src/styles/tokens.css`
- Reference HTML: `frontend/apps/web/design-source/<SLUG>/<SLUG>.html` (read yourself, do NOT inline)
- Reference screenshot: `<SLUG>.png`

## Steps

1. **Token map** in worksheet §3b.1. Find matching token for every value (color/spacing/radius/font/shadow). No match within ±5% = missing.

2. **Add missing tokens.** One commit: `feat(tokens): add <list> for <SLUG>`.

3. **Port rules** to CSS Module. `var(--token)` only. Raw px allowed only for `1px` borders / `0`.

4. **Token coverage check:**
   ```bash
   cd frontend/apps/web
   grep -REn '#[0-9a-fA-F]{3,8}|rgb\(|[0-9]+px' src/features/<DOMAIN>/pages/<PAGENAME>.module.css | grep -v 'var(--' | grep -vE '\b(0|1)px\b' > design-source/<SLUG>/artifacts/token-coverage.txt
   ```
   Empty = pass.

5. **Parity-diff (HARD).** Run dev server. Pixel Parity Playbook §1 snapshot on impl AND design HTML. Per region → `parity-diff.md`:
   ```
   region | field | ref | impl | delta
   ```
   Any non-zero delta in spacing/typography/layout → fix and re-snapshot. Empty deltas = pass.

6. **Multi-viewport screenshots (1440 / 1024 / 375)** → `artifacts/screenshots/{viewport}-{ref,impl}.png`. Heavy tier always captures all three (media queries present by definition).

7. **Leakage probe (conditional).** If page renders any `<input>`/`<select>`/`<textarea>`/`<label>` → Playbook §2 probe → `leakage-probe.md`. Otherwise note "no form inputs, skipped".

8. **NO self-approval.** Worksheet "User approved" stays `[ ]`. User marks.

## Pixel Parity Playbook

### §1 Computed-style snapshot

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
    layout: {display: cs.display, flex: cs.flex, flexDir: cs.flexDirection, gap: cs.gap, ai: cs.alignItems, jc: cs.justifyContent}
  };
})()
```

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

Known offenders in `src/styles.css`:

| Selector | Effect | Reset |
|---|---|---|
| `input, select, textarea` | width/border/bg/padding clobber | `width: auto; padding: 0; border: none; background: none; border-radius: 0;` |
| `label span` (legacy) | uppercase | `text-transform: none; letter-spacing: normal; font-size: inherit;` |
| `p` browser default | `margin: 1em 0` | `.<scope> p { margin: 0; }` |
| `ol, ul` browser default | `padding-inline-start: 40px` | reset to design value |

New leak found → add row in separate commit.

### §3 Specificity loop (when fix doesn't apply)

If `.foo { line-height: 1 }` set but computed shows `normal`:
- Class is global → wrap `:global(.kicker)`.
- Higher-specificity rule wins → add specificity (`.parent .foo`).
- Last resort: `!important` with comment.

Re-run §1 snapshot. Numbers must change.

## Output (`phase3b-style.md`, ≤30 lines)

- Tokens commit hash
- Style commit hash
- Token coverage: empty/non-empty
- Parity-diff: zero deltas (or list residuals + reason)
- Leakage probe: ran/skipped + findings
- Viewports captured
- Dev URL

## Hard rules

- No raw hex / spacing px in CSS Module.
- Stop on token mismatch (ask main agent).
- No tsc.
- No self-approval of parity.
