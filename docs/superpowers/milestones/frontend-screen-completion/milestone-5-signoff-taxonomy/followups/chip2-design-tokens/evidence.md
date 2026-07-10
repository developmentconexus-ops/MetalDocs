# Chip 2 — Design-system tokens for dialog/overlay/content-width · Evidence

> Post-M5 follow-up (user-authorized chip). Surfaced during F5.2's verify-only close: several CSS
> modules carried raw px/rgba literals with no covering token. Orchestrator directive (2026-07-10):
> **overlay = one canonical wine-tinted scrim `rgba(26,14,14,x)`**; record the visual change with a
> before/after in evidence. Branch `claude/lucid-heisenberg-d6c94f`. **Not pushed.**

## New tokens (`styles/tokens.css`)

| Token | Value | Rationale |
|---|---|---|
| `--content-max` | `1100px` | admin/list content column cap — **2 real sites** (TaxonomyAdminPage, RouteAdmin) |
| `--overlay-scrim` | `rgba(26, 14, 14, 0.4)` | canonical modal scrim = Wine ink `--text` (rgb 26,14,14) @ 0.4 — **5 dialog overlays converge** |
| `--dialog-w` | `480px` | standard fixed dialog panel width — **3 sites** (Cancel, Supersede, TokenEdit) |
| `--dialog-min-w` | `400px` | compact-dialog min-width (TaxonomyDialog) |
| `--dialog-max-w` | `520px` | compact-dialog max-width (TaxonomyDialog) |

**Declined (proportionate):** a `--border-w-2` (2px hairline) token — optional per the chip, `--sp-0-5`
already equals `2px`, and the only remaining 2px sites live in the shared `TabBar` primitive (migrating
it would expand scope beyond this chip). Left as-is.

## Sites migrated (raw literal → token)

| File | Before | After |
|---|---|---|
| `taxonomy/pages/TaxonomyAdminPage.module.css` | `max-width: 1100px` | `var(--content-max)` |
| `approval/pages/route-admin/RouteAdmin.module.css` | `max-width: 1100px` | `var(--content-max)` |
| `taxonomy/components/TaxonomyDialog.module.css` | `rgba(26,14,14,0.4)` · `min-width:400px` · `max-width:520px` | `var(--overlay-scrim)` · `var(--dialog-min-w)` · `var(--dialog-max-w)` |
| `approval/components/CancelInstanceDialog.module.css` | `rgba(0,0,0,0.5)` · `width:480px` | `var(--overlay-scrim)` · `var(--dialog-w)` |
| `approval/components/SupersedePublishDialog.module.css` | `rgba(0,0,0,0.5)` · `width:480px` | `var(--overlay-scrim)` · `var(--dialog-w)` |
| `tokens/components/TokenEditDialog.module.css` | `rgba(0,0,0,0.4)` · `width:480px` | `var(--overlay-scrim)` · `var(--dialog-w)` |
| `documents/components/styles/CheckpointsDialog.module.css` | `::backdrop rgba(0,0,0,0.4)` | `var(--overlay-scrim)` |

**Left raw (out of the token class, noted):** CheckpointsDialog `min-width:480px`/`max-width:640px` — a
distinct *wide* dialog size, not the compact-dialog or `--dialog-w` standard, so not force-converged.
LoginPage `max-width:480px` is a login **card**, not a dialog. `--dialog-w` z-index (50 vs 1000 across
dialogs) is a real pre-existing inconsistency but outside this chip's color/width scope — **flagged for a
separate chip**, not touched.

## Visual change (before → after) — the overlay canonicalization

The visual delta the orchestrator asked to record: the scrim behind every modal is now the **wine ink**
instead of a mix of neutral black + wine.

| Dialog | Scrim before | Scrim after | Visual delta |
|---|---|---|---|
| TaxonomyDialog | `rgba(26,14,14,0.4)` | `rgba(26,14,14,0.4)` | **none** (already canonical — value-preserving) |
| CancelInstanceDialog | `rgba(0,0,0,0.5)` | `rgba(26,14,14,0.4)` | black→wine, **alpha unified to 0.4** alongside hue (matches the other 4 dialogs) |
| SupersedePublishDialog | `rgba(0,0,0,0.5)` | `rgba(26,14,14,0.4)` | black→wine, **alpha unified to 0.4** alongside hue (matches the other 4 dialogs) |
| TokenEditDialog | `rgba(0,0,0,0.4)` | `rgba(26,14,14,0.4)` | black→wine (same alpha) |
| CheckpointsDialog | `rgba(0,0,0,0.4)` | `rgba(26,14,14,0.4)` | black→wine (same alpha) |

**Rendered proof (live, logged in):** opened the TaxonomyDialog ("+ Nova Família") at `/admin/taxonomy`.
`getComputedStyle` confirmed the tokens resolve at runtime:
- overlay `background-color` = **`rgba(26, 14, 14, 0.4)`** (from `--overlay-scrim`)
- panel `min-width` = **400px**, `max-width` = **520px** (from `--dialog-min-w`/`--dialog-max-w`)
- root tokens present: `--content-max:1100px`, `--dialog-w:480px`, `--dialog-min-w:400px`, `--dialog-max-w:520px`, `--overlay-scrim:rgba(26,14,14,0.4)`.

The screenshot shows the warm wine-tinted backdrop (not neutral black) behind the compact panel — this is
the canonical appearance **all five dialogs now share**. The TaxonomyDialog render is value-preserving (it
was already the canonical scrim/size); the visible delta lands on the four sibling dialogs, whose overlays
move from black to this exact wine scrim (proven by the code diff + computed token + their green suites).

## Gates

- **L0 tsc:** `pnpm.cmd tsc --noEmit` → **EXIT=0**.
- **L1 vitest:** taxonomy suite **23/23** (api 20 + page 3); dialog suites **17/17** (CancelInstance 4, SupersedePublish 9, TokenEdit 4). CSS-token migration is value-preserving where noted; no test changes needed.
- **L2 rendered:** tokens resolve live (computed values above); TaxonomyDialog renders with the canonical wine scrim + compact widths. No stray literals remain in migrated files (grep clean).

## Reviewer verdicts

- **frontend-screen-reviewer: APPROVE** — no Criticals/Majors. Confirmed the scrim is genuinely the Wine
  ink (`--text` #1a0e0e = rgb(26,14,14) at 0.4, not an arbitrary hue), token values match the replaced
  literals (no layout shift), deferrals reasonable, and none of the 5 tokens are speculative
  (`--dialog-min-w`/`--dialog-max-w` replace two already-present compact-dialog literals the chip explicitly
  named). Noted the 0.5→0.4 alpha unification on Cancel/Supersede *improves* scrim parity across all five
  dialogs (no legibility risk). One Minor: make the alpha-unification explicit in the before/after table —
  **applied** (table above now states "alpha unified to 0.4 alongside hue").
- **frontend-code-reviewer: APPROVE** — no Criticals/Majors. Line-by-line value fidelity verified
  (1100/480/400/520 exact; `--overlay-scrim` provenance from `--text` confirmed by hex math); completeness
  grep clean (only the documented CheckpointsDialog wide-size + LoginPage card left raw, both intentional);
  box-shadow `rgba(0,0,0,…)` correctly out of the overlay-scrim class (shadow ≠ scrim); token
  placement/naming consistent with tokens.css conventions; CSS-only (`git diff --name-only` = 8 .css files,
  no .ts/.tsx). Soft forward-looking note: if no second compact dialog appears this program,
  `--dialog-min-w`/`--dialog-max-w` are candidates to inline at the next token audit — not a defect.

## Disposition

Five new tokens name previously-raw dialog/overlay/content-width literals; seven sites migrated; the modal
scrim is unified to one canonical wine ink per the orchestrator directive. tsc clean, 40 tests green,
rendered live-GREEN with computed-token proof. Two items honestly deferred to a separate chip (dialog
z-index 50↔1000 inconsistency; CheckpointsDialog wide-size literals). Committed; not pushed.
