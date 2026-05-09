# Phase 2 — Pre-flight

> **Slug:** distribuicao
> **Completed:** 2026-05-08

## Codegen
Skipped — no backend API exists for distribution.

## Primitives audited

| Primitive | Issues found | Fix applied |
|---|---|---|
| `components/ui/SearchBar.module.css` | Raw `gap: 0.5rem`, `border-radius: 10px`, `padding: 0.2rem 0.7rem` | Replaced with `var(--sp-2)`, `var(--r-5)`, `var(--sp-1) var(--sp-3)` — commit `96b3e70c` |
| `components/ui/Avatar.module.css` | Uses `var(--sp-5)` for xs size, `var(--brand)` for bg, `var(--text-on-brand)` for color — all tokens ✓ | No fix needed |
| `components/ui/TabBar.module.css` | (Verify pass — uses CSS Module variables) | No fix needed |

## Global leakage map (base.css)

| Selector | Effect | Impact on this screen |
|---|---|---|
| `button, input, select, textarea { font: inherit }` | All form elements inherit body font | Fine — desired behavior |
| `button { cursor: pointer }` | All buttons get pointer cursor | Fine |
| `button:disabled { opacity: 0.5; cursor: not-allowed }` | Disabled buttons dim | Not triggered — CTAs use `aria-disabled`, not `disabled` attribute |
| `input:not([type=checkbox])...` → `width: 100%; border-radius; border; background; padding` | Text inputs get full-width + styled borders | SearchBar `.input` already overrides with `border: 0; background: transparent; width: 100%` — border/bg neutralized, width matches intent. Safe. |

No new leakage risks beyond what SearchBar already handles.

## Status-meta
`frontend/apps/web/src/features/documents/lib/distributionMeta.ts` — commit `fdf4e050`
Contains: `RecipientStatus`, `RECIPIENT_STATUS_TONE`, `RECIPIENT_TABS`, mock data interfaces, `MOCK_DISTRIBUTION`.

## Route stub
`documents/:documentId/distribution` → `DocumentDistributionPage` (stub)
Commit: `57d54d6b`

## tsc
No errors from new files. Pre-existing unrelated errors in `Rail.tsx` / taxonomy query remain (not introduced by this phase).
