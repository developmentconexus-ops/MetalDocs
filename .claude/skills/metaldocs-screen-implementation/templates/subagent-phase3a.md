# Subagent prompt — Phase 3a Structure mirror

You are a subagent dispatched in a fresh git worktree to perform Phase 3a (Structure mirror) of the MetalDocs screen-implementation workflow. Phase 2 produced the scaffolding. Your job is to write a TSX skeleton whose DOM mirrors the reference HTML exactly. NO LOGIC — structure only.

## Inputs (substitute at dispatch time)

- Worksheet path: `frontend/apps/web/design-source/<SLUG>/IMPLEMENTATION.md`
- Owning feature: `features/<DOMAIN>`
- Target page file: `frontend/apps/web/src/features/<DOMAIN>/pages/<PAGENAME>.tsx`
- Target CSS Module: `frontend/apps/web/src/features/<DOMAIN>/pages/<PAGENAME>.module.css`
- Reference HTML: full inline (substituted at dispatch — see below)
- Reference screenshot: `frontend/apps/web/design-source/<SLUG>/<SLUG>.png`

## Reference HTML (substituted at dispatch)

```html
<!-- DISPATCHER: paste full contents of design-source/<SLUG>/<SLUG>.html here -->
```

## Hard rules

1. **Mirror the DOM.** Same tag, same nesting depth, same order as the reference HTML. No restructuring. No "I think this is cleaner". If the reference has a wrapper `<div>` you do not personally need, keep it.
2. **Class names.** CSS Module class names = direct rename of design HTML class names (camelCase if HTML used kebab-case). No invented names.
3. **No state, no handlers, no data.** Render literal placeholder text where dynamic data will go (e.g. `Documento exemplo`, `MS`, `Aprovado`). Phase 3c wires the real data.
4. **Use primitives from worksheet §1.1 + §1.2.** Where the design HTML has a status pill, the TSX uses the existing `StatusPill` (or whatever the worksheet specifies). Don't re-implement primitives.
5. **CSS Module skeleton.** Create the file with empty rules per class — Phase 3b ports the styles.

## Steps

1. Read the worksheet §0–§2 fully. Confirm Phase 2 commits landed.
2. Replace the route stub in `frontend/apps/web/src/features/<DOMAIN>/routes.tsx` with a lazy import of `<PAGENAME>`.
3. Create `<PAGENAME>.tsx` mirroring the reference HTML with primitives substituted.
4. Create `<PAGENAME>.module.css` with empty rules — one per CSS Module class used in the TSX.
5. Run `pnpm.cmd tsc --noEmit -p tsconfig.build.json` — must be green.
6. Update worksheet §3a with `[x]` items.

## Output

Single commit: `feat(<DOMAIN>): structure mirror for <SLUG>`. Plus a 4-line report:
- Page file path
- CSS Module path
- Class-name mapping table (design class → module class)
- Anything that did not map cleanly + why

## What "anything that did not map cleanly" means — STOP and ask main agent

- Reference HTML uses `<button>` inside `<button>` → can't mirror, must restructure with `<div role="button">`.
- Reference HTML uses non-semantic tags where semantics matter (table data inside `<div>`s).
- A primitive listed in worksheet §1.1 doesn't actually fit the design markup.

In any of those cases — pause, write the conflict to worksheet `Open Questions Log`, end your run, return control. Do NOT decide.
