# Chip 1 — TaxonomyAdminPage primitive adoption · Evidence

> Post-M5 follow-up (user-authorized chip). Surfaced by `frontend-code-reviewer` during F5.2's
> verify-only close: the page's hand-rolled tab strip + header block duplicated the shared
> `TabBar` and `WorkspaceHeroHeader` primitives. **Behavior-additive** (adds ARIA roles +
> keyboard nav) — which is why it was deferred *out* of M5 F5.2 (F5.2 mandated zero behavior change).
> Branch `claude/lucid-heisenberg-d6c94f`. **Not pushed.**

## What changed (entire diff — 3 files)

| File | Change |
|---|---|
| `features/taxonomy/pages/TaxonomyAdminPage.tsx` | header `<div>` (kicker/title/description) → `<WorkspaceHeroHeader tone="flat" kicker title subtitle>`; local `<button>` tab strip → `<TabBar tabs activeKey onTabChange ariaLabel>` |
| `features/taxonomy/pages/TaxonomyAdminPage.module.css` | removed now-dead `.headKicker/.headTitle/.headDescription/.tabs/.tab/.tabActive`; kept `.root`; added `.tabsRow { margin: var(--sp-5) 0 }` |
| `features/taxonomy/pages/TaxonomyAdminPage.test.tsx` | tab-switch queries `getByRole('button',…)` → `getByRole('tab',…)` to track the new semantic role |

Data flow, query hooks, loading/error branches, toggle state, and all three list components' props
are byte-for-byte unchanged. The a11y gain (roving tabIndex, `aria-selected`, Arrow/Home/End nav) comes
entirely from the shared `TabBar` — no bespoke a11y code added here. Convention matches the established
dual-primitive caller `features/templates/TemplatesListPage.tsx`.

## Gates

- **L0 tsc:** `pnpm.cmd tsc --noEmit` → **EXIT=0**.
- **L1 vitest:** `pnpm.cmd vitest run TaxonomyAdminPage.test.tsx` → **3/3** (default families render, switch→Perfis empty state, switch→Áreas empty state).
- **L2 rendered (live, logged in, seeded dev DB):** `/admin/taxonomy` —
  - `WorkspaceHeroHeader` renders: kicker "TAXONOMIA" (mono uppercase), title "Tipos Documentais", subtitle.
  - `TabBar` renders as `role="tablist"` with three `role="tab"` (Famílias active w/ brand underline, Perfis, Áreas).
  - Click switches tab (Famílias→…); **ArrowRight keyboard nav** moved selection Perfis→Áreas with a visible focus ring and the content panel followed (Áreas list populated: producao/qualidade/rh).
  - Live seeded data populated (no fabricated values).

## Reviewer verdicts

- **frontend-screen-reviewer: APPROVE** — no Criticals/Majors. Confirmed prop usage matches both primitive
  contracts and the `onTabChange(key)=>setTab(key as Tab)` cast pattern from `TemplatesListPage`, CSS cleanup
  complete (grep 0 refs to removed classes), test change is presentational-only. Minors were awareness-only
  (the `.tabsRow` margin wrapper is the correct token-backed adaptation for a padded-block container vs
  Templates' flex column; the larger hero title is the primitive's canonical contract — intended).
- **frontend-code-reviewer: APPROVE** — no Criticals/Majors/Minors. Confirmed correct import depth, type-safe
  `TabBarItem[]`, deleted classes have zero remaining references, `.tabsRow` token-backed, `.root`
  `max-width:1100px` literal correctly left untouched (owned by the separate token chip = chip 2), the
  `'button'→'tab'` assertion is genuine (not weakened) since `TabBar` emits `role="tab"` inside `role="tablist"`,
  and no query/state/prop wiring changed.

## Disposition

Pure presentational + additive-a11y swap to two already-shipped, already-tested primitives. tsc clean,
3/3 unit, rendered live-GREEN incl. keyboard nav, both reviewers APPROVE. Committed; not pushed.
The `.root max-width:1100px` literal remains for **chip 2** (design-system token additions).
