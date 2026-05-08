# Phase 3b — Style port

**Screen:** Templates List (`/templates-v2`)
**Date:** 2026-05-07

## Token map

### TemplatesListPage.module.css

| Design value | Property | Token used |
|---|---|---|
| `flex:1, overflow:auto, background:var(--bg)` | `.page` | `var(--bg)` |
| `padding: 24px 28px` (split: 0 horizontal at .page since hero header has its own padding; .content carries 28→32 horizontal + 24 bottom) | `.content` | `var(--sp-7)` (32px ≈ 28px design — see Decorative deviations); `var(--sp-6)` (24px bottom) |
| Tab/grid stack gap (16px between hero, tabs, grid) | `.content gap` | `var(--sp-4)` (16px) |
| Primary button bg | `.newBtn background` | `var(--brand)` |
| Primary button border | `.newBtn border-color` | `var(--brand-deep)` |
| Primary button text | `.newBtn color` | `var(--text-on-brand)` |
| Primary button radius | `.newBtn border-radius` | `var(--r-2)` |
| Primary button gap | `.newBtn gap` | `var(--sp-1)` (4px) |
| Primary button padding-x | `.newBtn padding` | `var(--sp-3)` (12px) |
| Grid gap 14 | `.cardGrid gap` | `var(--sp-4)` (16px ≈ 14 design — closest available; 14 has no token) |

### TemplateCard.module.css

| Design value | Property | Token used |
|---|---|---|
| `background: var(--surface)` | `.card background` | `var(--surface)` |
| `border 1px solid var(--border)` | `.card border` | `var(--border)` |
| `border-radius var(--r-3)` | `.card border-radius` | `var(--r-3)` (8px) |
| Subtle base shadow | `.card box-shadow` | `var(--shadow-1)` |
| Hover elevation | `.card:hover box-shadow` | `var(--shadow-2)` |
| Hover border emphasis | `.card:hover border-color` | `var(--border-strong)` |
| Preview area surface | `.previewArea background` | `var(--surface-2)` |
| Preview area border-bottom | `.previewArea border-bottom` | `var(--border)` |
| Preview area padding 14 | `.previewArea padding` | `var(--sp-4)` (16px ≈ 14) |
| Badges right/top 14 | `.badges right/top` | `var(--sp-4)` |
| Badges col gap 5 | `.badges gap` | `var(--sp-1)` (4px ≈ 5) |
| Body padding 14 | `.body padding` | `var(--sp-4)` |
| Title font-size 14 | `.title font-size` | `14px` (matches design `.h3` mono pattern in design styles where utility uses fixed px; tokens.css has no fs token) |
| Title margin-bottom 6 | `.title margin` | `var(--sp-2)` (8px ≈ 6) |
| Divider | `.divider` | `var(--border)` |
| Divider margin 10 | `.divider margin` | `var(--sp-3) 0` (12px ≈ 10) |
| Footer gap 6 | `.footer gap` | `var(--sp-2)` (8px ≈ 6) |
| Author color | `.author color` | `var(--text-muted)` |
| Time mono color | `.time color` | `var(--text-muted)` |
| Time mono font | `.time font-family` | `var(--font-mono)` |

### MiniDocPreview.module.css

| Design value | Property | Token used |
|---|---|---|
| Paper shadow `0 2px 8px rgba(0,0,0,0.06)` | `.miniDoc box-shadow` | `var(--shadow-paper-2)` (already in tokens.css line 54) |
| Brand bar color | `.miniDocBrand background` | `var(--brand)` |
| Line color (was `--border-strong` in design; we use the dedicated paper-line token introduced for the wizard mini doc) | `.miniDocLine background` | `var(--paper-line)` (tokens.css line 58) |

## Tokens added

**None.** All needs covered by existing tokens. `--shadow-paper-2` and `--paper-line` already exist (introduced for wizard mini-doc; documented in `tokens.css`).

## Decorative raw values accepted

The following raw px values are kept as raw because the design system uses them as fixed dimensions and there is no matching token. Documented per CLAUDE.md "Surgical Changes" — adding new tokens for one-off decorative dims would over-engineer.

| File | Line | Value | Reason |
|---|---|---|---|
| TemplatesListPage.module.css | 23, 69 | `height: 32px`, `font-size: 13px` | Standard btn dims from design `.btn` rule (height 32, font 13). No `--btn-height` / `--fs-13` tokens; matches existing `.btn` global definition in `src/styles.css:39, 42` which also uses raw px. |
| TemplatesListPage.module.css | 26 | `font-size: 13px` | Same as above. |
| TemplatesListPage.module.css | 51, 57 | `1024px`, `640px` (media queries) | Standard responsive breakpoints; no breakpoint tokens in tokens.css. |
| TemplatesListPage.module.css | 69 | `font-size: 13px` (loading/error/empty) | Matches body-sm pattern. |
| TemplateCard.module.css | 25 | `height: 110px` | Design `previewArea` fixed height (line 201 of screens-2.jsx); decorative card dim, no token. |
| TemplateCard.module.css | 48 | `font-size: 14px` | Design title size override on `.h3` (line 212 inline `fontSize:14`); title-on-card-specific. |
| TemplateCard.module.css | 71, 78 | `font-size: 12px`, `font-size: 11px` | Design `.caption` (12) and `.tiny` (11) global classes use these literal sizes — `src/styles.css:27-28` matches. |
| MiniDocPreview.module.css | 8-15, 19-28 | `18px`, `14px`, `80px`, `100px`, `8px 7px`, `2px` radius, `4px` brand bar, `1.5px` line, `3px`/`5px` margins | Decorative micro-thumbnail dimensions. The whole `.miniDoc` is a stylized 80×100 paper rendered absolutely at fixed coords; values are unique to this thumbnail and reused only inside this module. Fixed in design as raw px (lines 202-204 of screens-2.jsx). Adding tokens would have one consumer each — documented exception. |
| MiniDocPreview.module.css | 12 | `background: #ffffff` | Mini doc paper is intentionally pure white regardless of palette (it depicts a printed page). The design source uses `background: 'white'` literally. `--surface` token is `#ffffff` in wine palette but semantically wrong (not a UI surface). The literal `#ffffff` here is decorative paper, not a UI color. Documented exception. |

## Hover / focus styles

- `.card:hover` — border shifts from `--border` to `--border-strong` and shadow lifts from `--shadow-1` to `--shadow-2` (subtle elevation).
- `.card:focus-visible` — 2px brand outline with 2px offset (matches `.btn:focus-visible` convention from `src/styles.css:52`).
- `.newBtn:hover` — bg darkens to `--brand-deep` (matches design `.btn-primary:hover`).
- `.newBtn:focus-visible` — same outline pattern as `.btn`.

## Per-viewport observations

**Note:** Live `mcp__Claude_Preview__preview_*` visual probe was not run in this subagent — the worktree assigned (`thirsty-raman-989031`) is a docs-only worktree and the dev server is hosted by the parent repo. Pixel parity against the design `templates.html` reference is **deferred to main agent / user verification**. CSS structurally mirrors the inline-style rules from `screens-2.jsx` lines 171-229 byte-for-byte except for the documented token rounding (14 → 16, 6 → 8, etc.) per project token discipline.

| Viewport | Expected behaviour |
|---|---|
| 1440 | 3-column grid (`repeat(3, 1fr)`); hero, tabs, cards stacked at 16px gaps; preview area 110px high; mini doc thumbnail visible top-left of preview, badges top-right. |
| 1024 | 2-column grid (`@media max-width: 1024px`); rest unchanged. |
| 375 | 1-column grid (`@media max-width: 640px`); cards span full width; hero header collapses to its own primitive's compact layout. |

## Global leakage probe

| Element rendered by templates page | Global selector(s) potentially matching | Resolution |
|---|---|---|
| `.page > .content > .cardGrid > TemplateCard > div.card` (CSS-Modules-scoped) | Global `.card` rule in `src/styles.css:174-178` and `src/styles.css:405-408` | **No collision.** CSS Modules renames `.card` to a hashed local class; the global `.card` selector targets the literal string `.card`, not the module-scoped name. Verified by inspection: `<div className={styles.card}>` compiles to `<div class="TemplateCard_card__abc12">` which does NOT match the global `.card` selector. |
| `.title` div inside card | Global `.card h3` rule at lines 120, 298 | **No risk.** Title is a `<div>`, not `<h3>`. |
| `.newBtn` button rendered in hero header `.action` slot | Global bare `button` rule in `src/styles.css` | None — no bare `button` selector in `src/styles.css`. Verified via `grep`. |
| `.divider` div inside card | Global `.divider` rule at `src/styles.css:32` | **No collision** (CSS-Modules-scoped). |
| TabBar buttons | Global `button` rule | None. TabBar uses its own module class `.tab` and explicitly resets `border:none; background:none`. |
| `.miniDoc` and children | None (CSS-Modules-scoped, all `<div>` elements) | Clean. |

## Artifacts

- `token-coverage.txt` — only acceptable raw px (font-sizes, breakpoints, decorative dims, btn heights). All values explained in "Decorative raw values accepted" table above.
- `parity-diff.md` — **not produced** (Preview probe deferred — see "Per-viewport observations").
- `leakage-probe.md` — superseded by the inline "Global leakage probe" table above.

## Typecheck

`pnpm tsc --noEmit -p tsconfig.build.json` — no new errors in `src/features/templates/`. Pre-existing errors only in unrelated areas (auth, documents, shell — same set as Phase 3a).

## User approval

User approved screenshot diff: [ ]
