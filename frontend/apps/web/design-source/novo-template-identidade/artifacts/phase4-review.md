# Phase 4.5 — Visual review · novo-template-identidade

**Verdict:** APPROVE WITH NITS

## Critical

None.

## Major

- **M1** — `.recapField`/`.field { gap: 6px }` raw px not in `--sp-*` ladder. Fix → `var(--sp-1)` (4px) or `var(--sp-2)` (8px).
- **M2** — `.nameInput { font-size: 14px }` should use `var(--font-size-md)` (token exists).
- **M3** — `.codePreviewValue { font-size: 26px }` no token mapping. Either add `--font-size-hero: 26px` or document with `/* design-exact */` TODO + backlog row.
- **M4** — Missing `NOTES.md` in `design-source/novo-template-identidade/`. Required entry point per `metaldocs-screen-implementation` SKILL.
- **M5** — Name input lacks `required` + `aria-required="true"` despite visual `*`. Add both.
- **M6** — `.fieldHint` has no `id`; input has no `aria-describedby`. Add `id="tpl-name-hint"` + `aria-describedby="tpl-name-hint"`.

## Minor

- N1 — `phase3-combined.md` line about "deferred live diff" is stale (live diff captured in `parity-diff.md`). Update or remove.
- N2 — `kicker` string in StepIdentity.tsx not memoized. Negligible; cosmetic.
- N3 — `screenshots/` directory empty. Add 1440 ref+impl pair.
- N4 — `parity-diff.md` missing recapBox gap row.

## What's Good

- Reducer extension minimal/correct.
- Textarea height-leakage fix applied pre-review (probe → fix → verify).
- Defensive guard for `?step=2` without scope.
- Semantic heading hierarchy intact.
- No nested buttons / no legacy paths / no cross-feature import violations.
- `card` padding inherited from `WizardShell` — primitive contract respected.

## Action

Fix M1, M5, M6 before merge. M2 also — token exists. M3 → add comment + backlog row. M4 → create stub NOTES.md. Minor → optional, address opportunistically.

## Resolution (2026-05-09)

- **M1** ✓ — `gap: 6px` → `var(--sp-2)` in `.recapField` and `.field`.
- **M2** ✓ — `.nameInput { font-size: var(--font-size-md) }`.
- **M3** ✓ — `/* design-exact */` comment added to `.codePreviewValue`; backlog row `font-size-hero` added to `wiki/backlog/novo-template-wizard.md`.
- **M4** ✓ — `NOTES.md` stub created at `design-source/novo-template-identidade/NOTES.md`.
- **M5** ✓ — `required` + `aria-required="true"` added to name input. Verified live: `{required:true, ariaRequired:'true'}`.
- **M6** ✓ — `id="tpl-name-hint"` on hint span + `aria-describedby="tpl-name-hint"` on input. Verified live.

tsc: templates feature clean.
