# F5.2 — Taxonomy Admin Restyle · Close-Out Evidence

> Feature: `f5.2-taxonomy-restyle` · Milestone 5 (Detalhe Signoff + Taxonomy Admin restyle)
> Execution: **verify-only close**, this session, 2026-07-10. Fresh worktree
> `lucid-heisenberg-d6c94f`. Branch: `claude/lucid-heisenberg-d6c94f` (**not pushed** —
> awaiting operator HS-1).

## Headline: F5.2 was already delivered by FE-14 — this feature required ZERO new code

The milestone + ROADMAP row 1.3 were authored **2026-06-23**, describing
`frontend/apps/web/src/features/taxonomy/TaxonomyAdminPage.tsx` with **11 inline `style=`
occurrences** to convert to design tokens.

Commit **`2a371d60` — `refactor(fe): FE-14 modernize taxonomy feature to canonical
structure`** landed **2026-07-02**, nine days after the milestone was authored. FE-14
deleted the flat `features/taxonomy/TaxonomyAdminPage.tsx` (104 LOC, inline-styled) and
recreated it as `features/taxonomy/pages/TaxonomyAdminPage.tsx` (84 LOC) backed by
`pages/TaxonomyAdminPage.module.css` (51 LOC), **100% design-token styling** (commit
message: *"tokenized module.css (zero raw hex, allowlist unchanged); check-css-token-discipline
clean"*). It also added `pages/TaxonomyAdminPage.test.tsx` (3 tests).

**Consequence:** F5.2's stated goal — "convert inline `style=` to redesign design-system
tokens, zero behavior change, zero contract change" — was **already satisfied** on record
before this feature was picked up. Writing a fresh "token swap" diff on an already-tokenized
file would be inventing scope on an already-correct base (anti-slop / global-maximum /
anti-circle violation, HARNESS.md §4). Per CLAUDE.md "runtime truth beats docs," the mismatch
was **surfaced to the operator, who chose the verify-only close path** (record current state
against F5.2's acceptance; do not fabricate a diff).

## Acceptance criteria (milestone.md F5.2 row) — every row proven against current state

| Acceptance criterion | Proof | Outcome |
|----------------------|-------|---------|
| `grep -nE "style=\{\{\|style=\""` over `TaxonomyAdminPage.tsx` = **0** (all 11 inline-style sites removed) | Grep over the whole `features/taxonomy` tree → **0 matches** (page + components + dialogs). Old deleted file had inline styles: `git show 2a371d60 -- features/taxonomy/TaxonomyAdminPage.tsx` shows the removed `style=` lines. | **PASS** |
| Styling uses redesign tokens/primitives only; **no raw hex/px introduced outside tokens** | `TaxonomyAdminPage.module.css` is 100% token-backed — every color/spacing/font resolves to `--sp-*` / `--font-size-*` / `--text-*` / `--border`. Zero hex. Both reviewers hand-verified. Residual raw px (`max-width:1100px`; 2px tab underline — a faithful port of the old `2px solid #333`) are structural dimensions with **no covering token in `tokens.css`** — pre-existing design-system token-coverage debt, **not introduced** by this work. Deferred with trigger (below). | **PASS** |
| Existing taxonomy test(s) pass **unchanged** (no behavior regression) | `pnpm vitest run src/features/taxonomy` → **23/23 pass** (`api/taxonomy.test.ts` 20 + `pages/TaxonomyAdminPage.test.tsx` 3). No test edited this session. | **PASS** |
| Taxonomy API contract (`taxonomy/api/taxonomy.ts`) untouched | `git log` on the file: last touches `f995d2ce` (CON-06 idempotency, backend contract) + `2a371d60` (FE-14) + `fa53d313` (snake_case params) — **no styling-driven change**. Imports generated `components` from `lib/api-types`. | **PASS** |
| `tsc` clean | `pnpm tsc --noEmit` → **EXIT=0**. | **PASS** |
| Both reviewers APPROVE | `frontend-screen-reviewer` **APPROVE WITH NITS**; `frontend-code-reviewer` **APPROVE WITH NITS**. No Critical, no REQUEST CHANGES on either. | **PASS** |

### Rendered verification (HARNESS.md §5 L3 / P5 targeted UI walk)

Dev server `metaldocs-web` (:4173), authenticated as Administrator against the seeded dev
API. `/admin/taxonomy` renders the tokenized page: muted uppercase "TAXONOMIA" kicker, bold
"Tipos Documentais" title, soft description, active-tab underline. Clicked **Áreas** tab →
switched active tab (underline moved) and rendered the real seeded area list
(`producao` / `qualidade` / `rh`). Behavior intact, token styling correct. Screenshots
captured in-session (Famílias default view + Áreas after tab switch). **GREEN.**

## Reviewer verdicts (on record)

- **frontend-code-reviewer — APPROVE WITH NITS.** No Critical. Two Majors, both
  "adopt existing primitive" (`TabBar`, `WorkspaceHeroHeader`). **Deferred, not actioned:**
  adopting `TabBar` would *add* `role="tablist"/"tab"` semantics + arrow-key navigation +
  `:focus-visible` — i.e. a **behavior change**, which F5.2 explicitly forbids
  ("zero behavior change"). Per HARNESS.md §4 anti-circle rule ("reviews verify, never
  generate scope"), these are operator-queue backlog findings, not F5.2 blockers.
  Praised: zero `any`/casts, correct `role="status"`/`role="alert"` states, clean list
  delegation, centralized `QK.taxonomy.*` keys with WHY-comments, 100% token-backed CSS.
- **frontend-screen-reviewer — APPROVE WITH NITS.** No Critical. One Major = **scope/tracking
  discrepancy** (honest): FE-14 changed *more* than styling — it also swapped the data layer
  (manual `useCallback`+`useEffect`+`useState` fetch → three `useQuery` hooks) and error
  surfacing (`resolveErrorMessage(activeQuery.error)`). Disposition: **verified functionally
  equivalent by inspection** — `useTaxonomyMutations` invalidates `QK.taxonomy.all` on every
  mutation (a superset of the old `onRefresh`), and the 23/23 tests cover the same tab-switch +
  empty-state behavior. No code change requested; recommended the tracker be corrected to
  reflect that F5.2's restyle shipped inside the broader FE-14 commit — **done** (this evidence
  + ROADMAP/milestone status update). Minors = the pre-existing raw-px token-coverage debt.

## Honest disclosures (not masked)

1. **F5.2 shipped inside FE-14, not as an isolated styling diff.** FE-14 modernized the
   taxonomy feature wholesale (canonical folder structure + data-layer + tokenized CSS
   together). The *styling* half of that is what F5.2 asked for; it is met. The data-layer
   half was reviewed and tested at FE-14 time and is functionally equivalent (screen-reviewer
   Major disposition above). This close does **not** claim an isolated restyle diff exists — it
   claims F5.2's *acceptance* is satisfied by the current state.
2. **Residual raw-px are design-system token-coverage gaps, not new drift** — no
   content-max-width, overlay-backdrop-color, or dialog-min/max-width token exists anywhere in
   `tokens.css` (same literals appear repo-wide in `approval`/`tokens` dialogs). Trigger to
   close: design-system owner adds the missing layout/overlay tokens, then these sites adopt
   them. Filed as an out-of-scope finding (chip), not chased here (HS-6 scope discipline).
3. **Primitive-adoption follow-up** (`TabBar` / `WorkspaceHeroHeader`) — filed as an
   out-of-scope finding (chip). It is an idiom improvement that alters behavior (ARIA/keyboard),
   so it is deliberately out of F5.2's zero-behavior-change boundary.

## Gate commands (re-runnable)

```
# grep-0 acceptance
grep -rnE "style=\{\{|style=\"" frontend/apps/web/src/features/taxonomy   # 0 matches
cd frontend/apps/web && pnpm.cmd tsc --noEmit                            # EXIT=0
cd frontend/apps/web && pnpm.cmd vitest run src/features/taxonomy        # 23/23
```

## Disposition

All F5.2 acceptance rows **PASS** against current state; both reviewers **APPROVE**; rendered
UI walk **GREEN**; no source diff written (verify-only — the restyle was delivered by FE-14);
tracking discrepancy corrected in the docs; out-of-scope reviewer findings filed as chips with
triggers, not chased. **F5.2 CLOSED. Ready for M5 milestone-validator, then the HS-1 operator
gate.** Not pushed.
