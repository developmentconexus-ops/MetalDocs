# Phase 2 — Pre-flight

**Screen:** Templates List (`/templates-v2`)
**Date:** 2026-05-07

## Summary

| Step | Status | Notes |
|---|---|---|
| 0. Read primitives + styles + tokens | Done | StatusPill, Avatar, WorkspaceHeroHeader, Icon, CodeChip, styles.css scanned |
| 1. Codegen | Skipped | `/api/v2/templates` already wired in `features/templates/api/templatesV2.ts`. No new endpoints introduced this phase |
| 2. Primitive extension (WorkspaceHeroHeader) | Done | Added optional `kicker` + `action`; made `searchQuery`/`onSearchQueryChange` optional |
| 2a. Primitive CSS audit | Done | Avatar drift recorded but not fixed (out of scope — would touch all callers) |
| 2b. Global leakage map | Done | Bare-element selectors limited to `table/th/td/tr` + 2 eigenpal portal overrides; no risk for templates page |
| 3. Status-meta SSOT | Skipped | Reused `StatusPill` (`DocumentStatus` union already covers `published` / `draft` / `archived`); no per-domain meta required |
| 4. New shared atom — TabBar | Done | Generic ARIA tablist primitive in `components/ui/` |
| 5. Route stub | N/A | `/templates-v2` route already registered → `pages/TemplatesListRoutePage.tsx` → `TemplatesListPage`. Phase 3 will rewrite the page body in place. No stub committed (would break the working page) |
| 6. Migrate listTemplates to apiFetch | Done | Single function migrated; other endpoints in `templatesV2.ts` deferred (out of phase 2 scope) |

## Primitive CSS audit

| Primitive | File | Drift Found | Fix Commit | Notes |
|---|---|---|---|---|
| StatusPill | `src/components/ui/StatusPill.module.css` | Raw px on dot/padding (`6px`, `8px`, `3px`, `5px`); `border-radius: 999px` (acceptable — pill convention); `font-size: 0.72rem` (acceptable rem) | — (deferred) | All colors come from tokens. Raw px on dimensions only; not a token-system violation per se but ideally would use `--sp-1`. **Not fixed this phase** — would require global verification. Logged as backlog. |
| Avatar | `src/components/ui/Avatar.module.css` | `color: #fff` (raw hex); raw px sizes (`28px`, `22px`, `36px`); `font-size: 0.68rem` etc | — (deferred) | `--text-on-brand` token exists (`#ffffff`). Should use that. Sizes (`28/22/36px`) have no matching token. **Not fixed this phase** — Avatar is widely used; risk of visual regression outweighs benefit before Phase 3 visual baseline. Logged as backlog. |
| WorkspaceHeroHeader | `src/components/ui/WorkspaceHeroHeader.module.css` | None in pre-existing rules (all colors via `var(--*)`, padding via rem). New `.kicker` + `.action` rules use only tokens (rem font-size, `var(--font-mono)`, `var(--text-muted)`) | `744c9704` | Clean. |
| Icon | `src/components/ui/Icon.tsx` | No module.css (inline SVG component). `stroke-width: "1.5"` is a stroke param, not a colour/dimension | n/a | Clean. |
| CodeChip | `src/components/ui/CodeChip.tsx` | No module.css; relies on global `.code-chip` + `.mono` classes from `styles.css` | n/a | No drift in primitive itself. Global `.code-chip` selector definition is in `styles.css` and uses tokens. |
| TabBar (new) | `src/components/ui/TabBar.module.css` | None — tokens-only (`var(--sp-*)`, `var(--border)`, `var(--brand)`, `var(--text*)`, `var(--font-mono)`); 1px border + 2px underline (acceptable per spec) | `38e59239` | Clean. |

**Note:** Computed-style probe skipped — subagent has no dev server access. Phase 3 page-assembly subagent will probe rendered styles and reconcile.

## Global Leakage Map (`src/styles.css`)

| Selector | File:Line | Effect | Risk for Templates page |
|---|---|---|---|
| `table` | `styles.css:339` | width:100%, border-collapse:collapse | None (templates list has no `<table>`) |
| `th, td` | `styles.css:344-345` | padding, border-bottom, text-align, vertical-align | None |
| `tr` | `styles.css:352` | cursor:pointer | None |
| `tr:hover, .row-active` | `styles.css:356` | background: var(--brand-pale) | None |
| `td small` | `styles.css:361` | display:block; color/margin | None |
| `body > .ep-root button.text-left` | `styles.css:3660` | eigenpal font-size dropdown — padding/font-size overrides | None (scoped to `.ep-root` portal) |
| `body > .ep-root div:has(> button.text-left)` | `styles.css:3666` | eigenpal portal sizing | None |
| `button[aria-label="Font size"]` | `styles.css:3679` | eigenpal toolbar font-size button | None (attribute selector specific) |

**Conclusion:** No global selectors leak into the templates list page. Bare `<button>` (header CTA, tab buttons) and `<h3>` (card title) elements used in the design will not be styled by global rules. The codebase relies on class-based styling. Templates list page can render bare elements safely.

## Tokens added

None. All needs covered by existing tokens (`--brand`, `--text*`, `--border`, `--surface*`, `--sp-*`, `--font-mono`, status `--success/--warning/--info` etc).

## Commits

| Hash | Subject |
|---|---|
| `744c9704` | `feat(workspace-hero-header): add kicker + action for templates` |
| `38e59239` | `feat(tab-bar): add generic tab filter primitive` |
| `a6f622fb` | `refactor(templates): migrate listTemplates to apiFetch` |

## Skipped / deferred (with rationale)

- **Codegen (Step 1)** — Endpoint already wired in pre-existing `templatesV2.ts`; no OpenAPI regen needed.
- **Status-meta SSOT (Step 3)** — Templates use `StatusPill` directly with derived status (`archived_at` → archived, `published_version_id` → published, else draft). No per-domain meta map needed.
- **Route stub (Step 5)** — `/templates-v2` already routes to `TemplatesListRoutePage` → `TemplatesListPage`. Phase 3 will rewrite `TemplatesListPage` body in place per scope; replacing with a `<div>Loading…</div>` stub now would regress the working page.
- **Computed-style probe (Step 2a §"Pixel Parity Playbook §1")** — No dev server in subagent context. Phase 3b page-assembly will probe.
- **StatusPill / Avatar drift fixes** — Out of phase 2 scope. Logged in artifact for future cleanup. Not blocking templates list assembly because:
  - StatusPill: raw px values are dimensions (consistent across uses), not unauthorised colours.
  - Avatar: `#fff` should be `--text-on-brand`, but visual rendering is identical.
- **Other `templatesV2.ts` raw `fetch()` calls** — Only `listTemplates` migrated per spec. Remaining endpoints (`createTemplate`, `getTemplate`, etc.) untouched; out of scope.
