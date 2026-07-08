# F7 — Evidence

## Commands + real output (main-session self-verified, not just implementer-asserted)

- **Token purity** (13 enumerated CSS files): `grep -rn "--slate-"` → **0** (exit 1);
  `grep -rnE "#[0-9a-fA-F]{3,6}"` → **0** (exit 1); `grep -rnE "var\(--[a-z-]+, *#"` → **0** (exit 1).
  Only remaining `rgba()` in scope = the 2 documented dialog scrims
  (`CancelInstanceDialog.module.css:4`, `SupersedePublishDialog.module.css:4` `rgba(0,0,0,0.5)`, D2).
  `StateBadge.module.css` still carries slate/hex → **out of scope, untouched** (no live consumer, D1);
  no global `--slate-` sweep occurred.
- **Focus**: all 6 classes present with `outline: 2px solid var(--brand); outline-offset: 2px;`
  (matches `.btn:focus-visible` `src/styles.css:59`) — `LockBadge .banner` (`:20`),
  `InboxStack .queueItem` (`:66`), `IntegrityDisclosure .summary` (`:20`) + `.copyButton` (`:71`),
  `DecisionFooter .actionBtn` (`:93`), `InboxToolbar .viewSwitcherBtn` (`:66`).
- **Reduced motion**: `@media (prefers-reduced-motion: reduce)` present in InboxApprovalCard,
  InboxTimeline, InboxStack, InboxToolbar, RequestedChangesPanel. Reviewer verified each guard's
  selectors match the ACTUAL animating rules — infinite loops covered (`urgentPulse` via
  `.railDotUrgent`, `urgentBlink` via `.urgentDot`). Minor #1 applied: base `.railDot` (non-looping
  0.2s transition) added to the InboxTimeline guard for full coverage.
- **TSX changes**: `DocumentShell.tsx:116-119` — `buffer === undefined` → `<div role="status"
  aria-live="polite">Carregando documento…</div>` (was `null`); reviewer confirmed this is the ONLY
  DocumentShell hunk (shared shell). `InboxPage.tsx:59-63` — `else if (filters.oversee && !isError)
  setOverseeDenied(false)` clears the stale denial on the next successful oversee fetch; no `setFilters`
  in that branch → no re-trigger loop.
- `npx tsc --noEmit -p tsconfig.build.json` → **clean, exit 0** (self-verified + reviewer).
- `npx vitest run src/features/approval src/features/documents src/lib/inbox` → **Test Files 52 passed
  (52) · Tests 357 passed (357)** (baseline 355; +2 new F7 tests, 0 regressions). Self-verified +
  reviewer re-ran from clean state, same result.

## TDD proof (2 behavior changes; CSS-only changes need no test)

- **overseeDenied reset** — `InboxPage.test.tsx` new case: toggle oversee → 403 → note shown + checkbox
  reverted → re-toggle (now succeeding) → note cleared. Reviewer verified genuine RED→GREEN by
  `git stash` reverting `InboxPage.tsx` — test failed on the final `queryByText` assertion (timeout);
  restored → passes. Non-tautological.
- **DocumentShell loading indicator** — `DocumentShell.test.tsx` new case asserts `getByRole('status')`,
  `aria-live="polite"`, PT text while buffer undefined. Reviewer verified RED→GREEN via stash-revert —
  failed `Unable to find an accessible element with the role "status"`; restored → passes.

## Runtime proof (observable change) + fixture-vs-real

- **Token migration** (observable in browser under the wine theme): Cancel/Supersede dialogs, LockBadge
  banner, InboxApprovalCard, InboxTimeline glow now render via wine/semantic tokens — no slate blue
  (#3b82f6) or raw yellow (#fef9c3) on approval screens. **Fixture/static** proof (grep + token map);
  live visual confirmed in F8 QA walkthrough.
- **Focus rings** — keyboard Tab now shows a visible 2px wine outline on the 6 previously-unfocusable
  controls (incl. the sign-off `DecisionFooter .actionBtn` — the most important control). Static-CSS
  proof; live keyboard pass deferred to F8.
- **Reduced motion** — with OS "reduce motion" on, the infinite pulse/blink loops and card/timeline
  animations are neutralized. Static-CSS proof (guard selectors verified against animating rules).
- **DocumentShell loading** — a designed status indicator renders during the signed-URL fetch instead
  of a blank canvas. Test-proven (role=status).
- **overseeDenied** — the "Você não tem permissão de supervisão." note no longer persists permanently
  after a transient 403; clears on the next successful oversee fetch. Test-proven.

## Contrast table (7 flagged pairings — real WCAG ratios)

| # | Pairing | Ratio | AA | Action |
|---|---------|-------|----|--------|
| 1 | `codeVersion` — `color-mix(text-on-brand 85%, transparent)` on `--brand` header | **8.56:1** | PASS | Fixed (was untestable raw `rgba(255,255,255,0.7)`) |
| 2 | `deadlineLabel` — same color-mix on brand header | **8.56:1** | PASS | Fixed (was `rgba(255,255,255,0.65)`) |
| 3 | `bucketCountEmpty` — was `--text-muted` on `--surface-3` | 3.59 → **9.57:1** | PASS | Fixed (swapped to `--text-soft`, in-scope) |
| 4 | `LockBadge` banner text — `--text-soft` on `--warning-bg` | **10.28:1** | PASS | Fixed (chose `--text-soft`; `--warning` alone was 3.64:1) |
| 5 | `btnReturn` — `--warning` (#b07016) on `--surface` | **4.06:1** | FAIL normal / PASS UI 3:1 | **Deferred D5** (see below) |
| 6 | `StageContextHeader .chipOverdue` — `--danger` on `--danger-bg` | **4.36:1** | marginal FAIL | Recorded only — file NOT in F7 scope (D6) |
| 7 | `ApprovalTimeline` stages — active 9.33 / passed 5.83 PASS; failed 4.36 / cancelled 3.64 FAIL | mixed | 2/4 FAIL | Recorded only — file NOT in F7 scope (D6) |
| + | `SignoffDetailPage` `--text-muted` on `--surface` | **4.30:1** | marginal FAIL | Token-level, cross-cutting — deferred (D4) |

In-scope fails fixed (3,4 + 1,2 hardened). D5/D6/D4 are token-level or out-of-scope-file — deferred
with recorded ratios, not silently skipped.

## Review / QA disposition

- Independent reviewer subagent (separate from implementer, own tools, no edits): **APPROVE**,
  **0 Critical, 0 Major, 2 Minor**. All 8 validation-gate checks confirmed with command output;
  slate→wine remap checked line-by-line for semantic correctness (no cross-category token mismatch);
  reduced-motion selectors verified against actual animating rules; both tests verified genuine
  RED→GREEN by stash-revert; scope discipline confirmed (no shared-file/backend/contract/type/copy
  drift; StateBadge untouched; no global slate sweep).
- **Minor findings + disposition:**
  1. `.railDot` (non-looping 0.2s transition) omitted from the InboxTimeline reduced-motion guard.
     **APPLIED** (main-session) — added to the guard selector list; full motion coverage.
  2. `outline-offset` variance note (pre-existing `RequestedChangesPanel.actionBtn:focus-visible` uses
     1px, out of F7 scope). **No action** — not touched by F7.
- **D5 btnReturn (4.06:1) defer judged acceptable by the reviewer:** UI text at 0.8125rem on a 44px
  button clears the 3:1 UI-component threshold; the plan pre-committed the "don't diverge one usage
  from the design system" policy; `--warning` is a shared semantic token — the global-maximum fix is a
  token-level darkening of `--warning` (all consumers), legitimately cross-cutting and outside F7's
  CSS-file-scoped mandate. Deferred as a documented deviation, not a swept AA failure.

## Deviations to surface at HS-1

- **D1** — `StateBadge.module.css` (21 hex + slate) left un-tokenized; no live consumer in M2c scope.
- **D2** — dialog overlay scrims (`rgba(0,0,0,0.5)` ×2) kept literal; no `--scrim` token exists.
- **D3 (FE-18 ↔ C4)** — slate palette IS tokenized on purpose (FE-18, "not to endorse a second design
  system"); C4 and FE-18 agree the endpoint is wine, so the slate→wine remap on the M2c dialogs is
  aligned, not a contradiction. Informational.
- **D4** — `--text-muted` on `--surface` = 4.30:1 (marginal AA fail). Token-level, app-wide; not fixed
  here (changing `--text-muted` is cross-cutting).
- **D5** — `btnReturn` `--warning` on `--surface` = 4.06:1. Token-level (shared `--warning`); the
  correct fix is a global `--warning` darken, deferred outside F7.
- **D6** — `StageContextHeader .chipOverdue` (4.36:1), `ApprovalTimeline .stage_failed` (4.36:1) /
  `.stage_cancelled` (3.64:1) fail AA but their files are NOT in F7's "Files touched" list — ratios
  recorded, edits deferred to avoid scope creep.

## Bounded defers (with triggers)

- **Token-level contrast** (D4/D5/D6) → a follow-up `--warning`/`--text-muted` token-darkening review
  affecting all consumers. **Trigger:** operator wants AA-normal-text compliance on the marginal
  pairings, or a broader a11y sweep beyond M2c approval screens.
- **StateBadge retokenize** (D1) → **Trigger:** StateBadge gets wired into a live M2c-scope surface.
