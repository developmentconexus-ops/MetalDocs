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

5. **Visual diff.** Run:
   ```bash
   cd frontend/apps/web
   pnpm dev
   ```
   Open the route in a browser. Compare side-by-side with `<SLUG>.png`. Note any pixel-level differences (alignment, spacing, font size).

6. Update worksheet §3b items with `[x]`. Mark "User approved screenshot diff" as `[ ]` — only the user marks this one, NOT you.

## Output

- Tokens commit (if any): `feat(tokens): add <list> for <SLUG>`.
- Style commit: `feat(<DOMAIN>): style port for <SLUG>`.
- Report: token map summary, missing tokens added, dev URL, side-by-side notes.

## Hard rules

- No raw hex outside `tokens.css`.
- No raw px for spacing in CSS Module.
- If a design value has no clear token match, STOP — log to worksheet Open Questions, ask user whether to add a new token or use the closest existing.
- Do NOT mark "user approved screenshot diff" yourself.
