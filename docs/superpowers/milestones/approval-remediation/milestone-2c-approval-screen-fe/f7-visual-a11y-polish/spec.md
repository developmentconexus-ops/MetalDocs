# F7 — Visual + a11y polish (C4) — Spec

**Feature:** M2c `approval-screen-fe` / F7. Governing: design spec §8 **C4**
(`docs/superpowers/specs/2026-07-07-approval-remediation-design.md`) + master plan §F7. Seeded by the
F7 read-only auditor (agentId a374ee77bf73f6245) — an exact-line violation census across all
M2c-touched FE surfaces.

> C4 (verbatim intent): wine tokens 100% (`--brand #6b1f2a`), legacy slate palette removed from
> approval screens, visible focus on all interactive controls, PT-BR sentence case, and
> loading/empty/error states designed (not blank). WCAG AA.

## Interview record (auditor-driven — findings, not guesses)

| # | Question | Resolution |
|---|----------|------------|
| 1 | Which files carry legacy `--slate-*` tokens in scope? | `CancelInstanceDialog.module.css` (9 refs), `SupersedePublishDialog.module.css` (9 refs). Migrate to wine/semantic tokens (`--border`, `--brand`/`--accent`, `--surface-2`/`--text-muted`, `--danger*`). `StateBadge.module.css` also has slate + 21 raw hex but **has no live consumer in the F7 scope** → adjacent dead-code debt, documented not fixed (avoid scope creep; retokenizing dead CSS is unverifiable). |
| 2 | Raw hex / rgb literals? | `LockBadge.module.css` (4 hex — un-tokenized yellow banner). `SupersedePublishDialog.module.css:122` (`#166534`). `InboxApprovalCard.module.css` (4 rgba: box-shadow + translucent-white-on-dark). `InboxTimeline.module.css:151` (rgba accent glow). Dialog overlay scrims `rgba(0,0,0,0.5)` ×2 — **scrims are a standard pattern, keep but route through a token if one exists (`--overlay`/`--scrim`), else document**. |
| 3 | `var(--token, #hex)` fallbacks? | `SignoffDetailPage.module.css` (18), `InboxPage.module.css` (3), `DocumentShell.module.css:2` (1). Tokens are app-wide guaranteed → strip the hex fallbacks (per no-fallback token hygiene). |
| 4 | Missing visible focus? | 6 concrete gaps: `LockBadge .banner`, `InboxStack .queueItem` (primary keyboard list), `IntegrityDisclosure .summary` + `.copyButton`, `DecisionFooter .actionBtn` (**highest priority — the sign-off control**), `InboxToolbar .viewSwitcherBtn`. All others already have `:focus-visible` (confirmed by auditor). |
| 5 | Missing loading/empty/error states? | 1 real gap: `DocumentShell.tsx` loading renders `null` (blank canvas during signed-URL fetch), not a designed indicator. `RequestedChangesPanel` is prop-driven (no fetch) → empty-only is correct, no gap. All Inbox/Cockpit ladders already CLEAN. |
| 6 | Non-PT-BR / non-sentence-case copy? | **CLEAN** — zero. `.kicker` uppercase is design-system convention, not a violation. No fixes. |
| 7 | `prefers-reduced-motion` guards? | Absent in 4 in-scope files: `InboxApprovalCard`, `InboxTimeline`, `InboxStack`, `InboxToolbar` (+ `RequestedChangesPanel.module.css:103` low-amplitude). 2 infinite loops (`urgentPulse`, `urgentBlink`) are highest priority. Add a `@media (prefers-reduced-motion: reduce)` guard neutralizing animation/transition. |
| 8 | ARIA / label gaps? | **CLEAN** — every control labeled, tabs/tablist wired, `role="button"` rows keyboard-accessible. No fixes. |
| 9 | Carried F5 Minor #1? | Confirmed: `InboxPage.tsx:47,55-60` — `overseeDenied` is a one-way ratchet, never reset. Fix: clear on oversee re-toggle (and/or on next successful oversee fetch). |

## Consumer contract (what F7 must deliver)

**Consumer = a keyboard/AT user and a reduced-motion/contrast-sensitive user on every M2c approval
surface; plus a maintainer reading token-pure CSS.**

1. **Token purity (in-scope files):** zero `--slate-*`, zero raw hex, zero `var(--token, #hex)`
   fallbacks in the fixed files. All colors resolve through wine/semantic tokens
   (`src/styles/tokens.css`). Dead-code `StateBadge.module.css` documented as out-of-scope debt.
2. **Visible focus:** the 6 enumerated interactive controls gain a `:focus-visible` rule consistent
   with the app pattern (`outline: 2px solid var(--accent)` / brand-derived, `outline-offset`).
3. **Designed loading state:** `DocumentShell.tsx` renders a real loading indicator (skeleton/spinner
   with `role="status"`/`aria-live`) while `buffer === undefined`, not `null`.
4. **Reduced motion:** each in-scope animating stylesheet carries a `@media (prefers-reduced-motion:
   reduce)` guard that removes/neutralizes animations + transitions (infinite loops especially).
5. **F5 Minor #1:** `overseeDenied` resets when the user re-enables oversee (no permanent stale
   denial note).
6. **Contrast (7 flagged pairings):** compute the real ratio for each; fix any that fail AA (4.5:1
   normal text, 3:1 large/UI). Record pass/fail per pairing in evidence — a flagged pairing is not
   closed until its ratio is computed.

## Non-goals (explicit)

- **No behavior/logic changes** beyond the F5 `overseeDenied` reset — F7 is visual + a11y only.
- **No edits to `features/shared/controlled-artifact/*`** (shared with template approval; F4
  discipline).
- **No retokenizing `StateBadge.module.css`** — no live consumer in scope; documented as debt.
- **No new components, no contract/backend/type changes, no copy rewrites** (copy is already clean).
- **No `DocumentShell` structural change** beyond swapping the `null` loading branch for an indicator
  (shared shell — minimal, additive, page-agnostic).

## Validation Gate

- **Grep proofs (in-scope files only):**
  - `grep -rn "--slate-" <in-scope css>` → **0**.
  - `grep -rnE "#[0-9a-fA-F]{3,6}" <in-scope css>` → **0** (excluding documented dialog scrim if no
    token exists — flagged inline).
  - `grep -rnE "var\(--[a-z-]+, *#" <in-scope css>` → **0** (fallbacks stripped).
  - each of the 6 focus-gap files → `:focus-visible` present for the named class.
  - each of the 4 motion files → `prefers-reduced-motion` present.
- **Tests:** `npx vitest run src/features/approval src/features/documents src/lib/inbox` → GREEN
  (existing suites unbroken; add/adjust a test for `overseeDenied` reset and for the DocumentShell
  loading indicator presence). `npx tsc --noEmit -p tsconfig.build.json` → clean.
- **Contrast table:** 7 pairings, each with computed ratio + AA verdict, in evidence.md. Any FAIL
  fixed and re-measured.
- **Reviewer:** independent subagent confirms token purity, focus coverage, reduced-motion guards,
  the overseeDenied reset (with test), no shared-file edits, no behavior regressions.

## Deviations to surface at HS-1

- **D1 — `StateBadge.module.css` left un-tokenized** (21 hex + slate) as documented dead-code debt;
  no live consumer in M2c scope. Retokenize if/when wired.
- **D2 — dialog overlay scrims** (`rgba(0,0,0,0.5)`) kept as literals if no `--scrim`/`--overlay`
  token exists (standard pattern; flagged).
