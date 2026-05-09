# Leakage Probe — Phase 3b: novo-template-escopo

> Target element: first `.profileCard` (SelectableCard button, selected state).
> Method: iterate all document.styleSheets, collect rules matching element.

| element | matched global rules | impact | action |
|---|---|---|---|
| profileCard | `* { box-sizing: border-box }` | Universal reset — harmless | None |
| profileCard | `button, input, select, textarea { font: inherit }` | Base reset — desired for button font inheritance | None |
| profileCard | `button { cursor: pointer }` | Base reset — desired | None |
| profileCard | `._root_v9gta_1` (SelectableCard base) | Own component: flex col, align-items stretch, gap sp-2, padding sp-4, border, border-radius, bg surface, color text, text-align left, cursor pointer, transitions | Correct — component's own styles |
| profileCard | `._selected_v9gta_23` (SelectableCard selected) | Own component state: border 2px brand, bg brand-pale, shadow-1 | Correct — expected for selected card |
| profileCard | `._profileCard_188uq_41` (StepScope override) | `gap: 6px; padding: var(--sp-4)` | Correct — our delta override |

**Result: CLEAN.** No unexpected global rule leakage into profileCard. All 6 matched rules are either universal resets, SelectableCard's own component styles, or our explicit StepScope override.

**No resets needed** for Step 1 profile grid. The phase2-preflight.md global leakage map confirmed no bare-element rules affect the profile grid specifically (no bare inputs/selects in Step 1 grid, `.card h3` not triggered since StepScope uses `.h2` utility class not bare `<h3>` tags).
