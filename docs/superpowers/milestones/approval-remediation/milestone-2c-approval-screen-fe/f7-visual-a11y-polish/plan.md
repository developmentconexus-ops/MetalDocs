# F7 — Plan

Seeded from master plan §F7 + design §8 C4 + the F7 auditor census (agentId a374ee77bf73f6245),
reconciled against `src/styles/tokens.css` (FE-18 slate note). TDD-light: F7 is visual/CSS + one
behavior fix (overseeDenied). Fresh implementer subagent (sonnet) + independent reviewer subagent.
**Implementer uses its OWN Read/Edit/Bash tools; does NOT spawn sub-agents.**

## Ground truth (do not re-derive)

- **Wine tokens** (`tokens.css`): `--brand #6b1f2a`, `--brand-deep #3e1018`, `--accent #c8364a`,
  `--surface #fff`, `--surface-2 #faf6f6`, `--surface-3 #f0e9e9`, `--border #e6dcdc`,
  `--border-strong #d4c2c2`, `--text-soft #4a3434`, `--text-muted #8a7575`, `--text-on-brand #fff`,
  `--success #1a6b35`/`--success-bg`, `--warning #b07016`/`--warning-bg #fbf2dc`,
  `--danger #c8364a`/`--danger-bg #fae8eb`, `--info #1a3a7a`/`--info-bg`. Shadows: `--shadow-1`,
  `--shadow-2`, `--shadow-doc-card`, `--shadow-danger`. **No `--scrim`/`--overlay-dark` token exists.**
- **`--slate-*` ARE defined** (tokens.css:59-73, FE-18) — intentional one-source-of-truth for the
  literals, **explicitly "not to endorse a second design system"**. C4 and FE-18 agree the endpoint
  is wine → the slate→wine remap is aligned, not a contradiction.
- **App focus pattern** (`src/styles.css:59`): `.btn:focus-visible { … }`. Match its outline style.
- **overseeDenied** one-way ratchet: `InboxPage.tsx:47` init false; `:55-60` effect sets true + flips
  `filters.oversee→false`; **no reset path anywhere**.
- **DocumentShell.tsx:116-117**: `buffer === undefined` → `body = null` (blank during signed-URL
  fetch). Error branch (`:110-115`) already exists.

## Token remap tables (apply exactly — no improvisation)

**slate → wine** (Cancel + Supersede dialogs only; StateBadge is out-of-scope dead code):
| slate | → wine |
|---|---|
| `--slate-border` / `--slate-border-2` / `--slate-border-3` | `--border` |
| `--slate-accent` (#3b82f6) | `--accent` (interactive/primary btn bg) |
| `--slate-surface-muted` | `--surface-2` |
| `--slate-text-muted` / `--slate-text-muted-2` | `--text-muted` |
| `--slate-danger-text` | `--danger` |
| `--slate-danger-border` | `--danger` |
| `--slate-danger-bg` / `--slate-danger-bg-2` | `--danger-bg` |
| `--slate-success-border` | `--success` |

**raw hex → token:**
| file:hex | → |
|---|---|
| `LockBadge.module.css` `#fef9c3`/`#fef08a` (bg) | `--warning-bg` |
| `LockBadge.module.css` `#fde047` (border) | `--warning` |
| `LockBadge.module.css` `#713f12` (text) | `--warning` (verify AA; darken to `--text-soft` if fail) |
| `SupersedePublishDialog.module.css:122` `#166534` | `--success` |

**rgba → token / color-mix:**
| file:literal | → |
|---|---|
| `InboxApprovalCard.module.css:10` box-shadow | `--shadow-2` |
| `InboxApprovalCard.module.css:44` `rgba(255,255,255,0.15)` | `color-mix(in srgb, var(--text-on-brand) 15%, transparent)` |
| `InboxApprovalCard.module.css:61` `rgba(255,255,255,0.7)` codeVersion | `color-mix(in srgb, var(--text-on-brand) X%, transparent)` — X ≥ passing AA (start 0.85) |
| `InboxApprovalCard.module.css:83` `rgba(255,255,255,0.65)` deadlineLabel | same, X ≥ passing AA |
| `InboxTimeline.module.css:151` `rgba(200,74,42,0.15)` glow | `color-mix(in srgb, var(--accent) 15%, transparent)` |
| dialog overlay scrims `rgba(0,0,0,0.5)` ×2 | **KEEP literal** — no scrim token; document D2 |

**`var(--token, #hex)` fallbacks → strip the fallback** (tokens app-wide guaranteed):
`SignoffDetailPage.module.css` (18), `InboxPage.module.css:13,15,16`, `DocumentShell.module.css:2`.
(Do NOT touch any `var(--color-X, #hex)` — the `--color-*` namespace is deliberately undefined,
resolves to fallback by design; auditor found none in scope anyway.)

## Ordered tasks

1. **Token migration** (CSS only, no DOM/logic):
   - `CancelInstanceDialog.module.css`, `SupersedePublishDialog.module.css` — apply slate→wine table;
     Supersede `#166534`→`--success`; keep scrim literal.
   - `LockBadge.module.css` — apply hex→warning table.
   - `InboxApprovalCard.module.css`, `InboxTimeline.module.css:151` — apply rgba table.
   - Strip `var(--token,#hex)` fallbacks in SignoffDetailPage / InboxPage / DocumentShell module CSS.
2. **Visible focus** — add `:focus-visible` (match `.btn:focus-visible` from `styles.css:59`) to:
   `LockBadge .banner`, `InboxStack .queueItem`, `IntegrityDisclosure .summary` + `.copyButton`,
   `DecisionFooter .actionBtn` (all variants inherit the base class rule), `InboxToolbar
   .viewSwitcherBtn`.
3. **Reduced motion** — append `@media (prefers-reduced-motion: reduce)` guard neutralizing
   `animation`/`transition` in: `InboxApprovalCard.module.css`, `InboxTimeline.module.css`,
   `InboxStack.module.css`, `InboxToolbar.module.css` (also `RequestedChangesPanel.module.css:103`).
   Priority: the infinite loops `urgentPulse` (InboxTimeline) + `urgentBlink` (InboxStack).
4. **DocumentShell loading state** — replace the `buffer === undefined` `null` branch with a designed
   indicator: `<div role="status" aria-live="polite">Carregando documento…</div>` (or a skeleton),
   styled token-pure. **Minimal, additive** — no other DocumentShell change (shared shell). Add/adjust
   a test asserting the loading indicator renders while buffer is undefined.
5. **F5 Minor #1 — overseeDenied reset** (`InboxPage.tsx`): clear `overseeDenied` when the user
   re-enables oversee (in the filter-change handler / `setFilters` path where `oversee` flips true),
   and on a subsequent successful oversee fetch (effect: `!isError && filters.oversee` →
   `setOverseeDenied(false)`). No loop. **Add a test**: 403 → note shown → re-toggle oversee → note
   cleared.
6. **Contrast pass** — compute the real ratio for all 7 flagged pairings (codeVersion/deadlineLabel
   translucent white on dark; btnReturn `--warning` on surface; StageContextHeader `.chipOverdue`;
   ApprovalTimeline 4 stage color/bg; SignoffDetailPage `--text-muted` on `--surface`; InboxTimeline
   `.bucketCountEmpty`). Fix any **in-scope** fail (bump color-mix opacity, swap to `--text-soft`,
   etc.). **Token-level fails** (e.g. `--text-muted` on white ≈ 3.5:1 — used app-wide) are NOT fixed
   here (changing the token is cross-cutting = out of scope); record the ratio + defer as a deviation.
7. **Verify** (implementer, own tools):
   - `grep -rn "--slate-"` across the fixed CSS → **0**.
   - `grep -rnE "#[0-9a-fA-F]{3,6}"` across fixed CSS → **0** except the documented scrim.
   - `grep -rnE "var\(--[a-z-]+, *#"` across fixed CSS → **0**.
   - 6 focus classes present; 4 motion guards present.
   - `npx vitest run src/features/approval src/features/documents src/lib/inbox` GREEN.
   - `npx tsc --noEmit -p tsconfig.build.json` clean.
8. **Review pass** — independent reviewer subagent (sonnet): token purity (0 slate/hex/fallback in
   scope), 6 focus rules match app pattern, 4 reduced-motion guards, DocumentShell loading indicator +
   test (and NO other shell change), overseeDenied reset + non-tautological test, contrast table
   complete with AA verdicts, NO edits to `features/shared/controlled-artifact/*`, NO StateBadge
   retokenize, NO behavior change beyond overseeDenied, no scrim-token global add. Apply findings.

## Files touched

- CSS: `CancelInstanceDialog.module.css`, `SupersedePublishDialog.module.css`, `LockBadge.module.css`,
  `InboxApprovalCard.module.css`, `InboxTimeline.module.css`, `InboxStack.module.css`,
  `InboxToolbar.module.css`, `sidebar/IntegrityDisclosure.module.css`, `sidebar/DecisionFooter.module.css`,
  `RequestedChangesPanel.module.css`, `pages/SignoffDetailPage.module.css`, `pages/InboxPage.module.css`,
  `documents/components/DocumentShell.module.css`.
- TSX: `documents/components/DocumentShell.tsx` (loading branch only), `approval/pages/InboxPage.tsx`
  (overseeDenied reset).
- Tests: `InboxPage.test.tsx` (overseeDenied reset), a DocumentShell loading test (new or extended).

## Non-goals / guardrails

- **No `StateBadge.module.css` retokenize** — no live consumer in scope (D1 dead-code debt).
- **No `features/shared/controlled-artifact/*` edits** (F4 discipline).
- **No `--scrim` global token add** — scrims kept literal + documented (D2).
- **No `--text-muted` token change** — cross-cutting; token-level contrast deferred (D4).
- **No copy/PT-BR changes** (already clean), no new components, no contract/backend/type changes.
- **DocumentShell change is loading-branch-only** — reviewer diffs to confirm.
- junction drift → vitest broken ⇒ full `pnpm install` in `frontend/apps/web`; no config hack.

## Risks

- **Over-broad grep-fix** — the implementer must edit ONLY the enumerated files; a global find/replace
  of `--slate-` would hit StateBadge/template-approval/out-of-scope files. Scope to the list.
- **color-mix support** — modern evergreen; the app already targets it elsewhere. If a target env
  lacks it, fall back to a defined token (not a raw rgba).
- **Contrast overreach** — do not "fix" a token-level ratio by diverging one usage from the design
  system; document and defer instead.
- **DocumentShell shared** — loading indicator must be page-agnostic and minimal; any broader shell
  edit is out of scope.
