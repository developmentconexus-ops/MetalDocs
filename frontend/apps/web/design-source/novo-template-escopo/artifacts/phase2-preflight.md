# Phase 2 Pre-flight — novo-template-escopo

> **Status:** COMPLETE
> **Date:** 2026-05-09
> **Branch:** worktree-agent-ae3af20049d1876f3

---

## 1. Codegen

**No.** Step 1 uses the existing `GET /taxonomy/profiles` endpoint via `fetchProfiles()`. No new API endpoints needed for Step 1.

---

## 2. Files moved / promoted

| Old path | New path | Action |
|---|---|---|
| `features/documents/queries/useProfilesQuery.ts` (full impl) | `features/taxonomy/queries/useProfilesQuery.ts` | Moved (canonical home) |
| `features/documents/queries/useProfilesQuery.ts` | kept as re-export shim | Shim for test backward compat |
| `features/documents/components/wizard/WizardFooter.tsx` (full impl) | `features/shared/components/wizard/WizardFooter.tsx` | Promoted (domain-agnostic) |
| `features/documents/components/wizard/WizardFooter.tsx` | kept as re-export shim | Shim so step imports unchanged |
| `features/documents/components/wizard/WizardShell.tsx` (hardcoded) | `features/shared/components/wizard/WizardShell.tsx` | Promoted + parameterized |
| `features/documents/components/wizard/WizardShell.tsx` | kept as thin adapter wrapper | Passes hardcoded doc strings to shared |
| `features/documents/components/wizard/WizardShell.module.css` | `features/shared/components/wizard/WizardShell.module.css` | Copied (no content change, tokens fixed) |

Also created:
- `features/taxonomy/queries/constants.ts` — STALE_FIVE_MINUTES (was in `documents/queries/_constants.ts`, taxonomy needs its own copy)

---

## 3. WizardShell new prop API

```ts
// features/shared/components/wizard/WizardShell.tsx

export type WizardShellStep = { id: string; label: string };

export type WizardShellProps = {
  kicker: string;           // e.g. "Templates / Novo"
  title: string;            // e.g. "Novo template reutilizável"
  description?: ReactNode;  // optional paragraph under h1
  steps: WizardShellStep[]; // replaces hardcoded WIZARD_STEPS
  currentStep: string;      // string id matching steps[n].id
  onStepClick?: (id: string) => void;
  children: ReactNode;
};
```

The documents wrapper (`features/documents/components/wizard/WizardShell.tsx`) retains the old `WizardStepNumber` (1|2|3|4) numeric API and adapts it for `NewDocumentWizardPage` — no change needed there.

---

## 4. Primitive CSS audit findings

File audited: `features/shared/components/wizard/WizardShell.module.css`

| Property | Value | Status |
|---|---|---|
| `.scrollWrapper` padding | `var(--sp-7) var(--sp-7)` | Token — OK |
| `.scrollWrapper` background | `var(--bg)` | Token — OK |
| `.container` max-width | `880px` | Raw — layout constant, no token available, intentional |
| `.header :global(.kicker)` margin-bottom | `6px` | Raw — no `--sp-` token (between sp-1=4px and sp-2=8px); design-specific |
| `.header :global(.display-1)` margin-bottom | `4px` → `var(--sp-1)` | **Fixed** |
| `.description` font-size | `14px` | Raw — no font-size token in design system |
| `.description` line-height | `1.5` | Raw — no line-height token |
| `.description` margin | `0 0 28px` | Raw — 28px has no token (between sp-6=24 and sp-7=32) |
| `.description` color | `var(--text-muted)` | Token — OK |
| `.container > :global(.card)` padding | `28px 32px` → `28px var(--sp-7)` | **Partially fixed** (32px→sp-7; 28px has no token) |
| `.container > :global(.card) > :global(.kicker)` margin-bottom | `var(--sp-2)` | Token — OK |
| `.container > :global(.card) > :global(.h2)` margin-bottom | `6px` | Raw — no token |
| `.container > :global(.card) > :global(.caption)` margin-bottom | `var(--sp-5)` | Token — OK |
| `.footerRow` gap | `10px` | Raw — no token |
| `.footerDivider` margin | `-32px` → `calc(-1 * var(--sp-7))` | **Fixed** |
| `@media` breakpoint | `600px` | Raw — breakpoint, no token system |

**Summary:** 3 raw values replaced with tokens (`4px→sp-1`, `32px→sp-7` ×2). 5 raw values retained — design-specific values with no matching token (`6px`, `10px`, `28px`, `14px`, `880px`). Token gap noted: spacing system jumps sp-1=4 → sp-2=8 → sp-3=12 → sp-4=16 → sp-5=20 → sp-6=24 → sp-7=32. Fine-grained intermediate values (6, 10, 28) are intentional design choices without token coverage.

`WizardFooter.tsx` — no inline styles. References `styles.footerRow` and `styles.footerDivider` from the same CSS module. Audit covered above.

---

## 5. Global CSS leakage map

Sources scanned: `src/styles.css` + `src/styles/base.css`

### bare-element selectors (no class/id parent)

| Selector | Rule | Impact on TemplateWizardPage |
|---|---|---|
| `body` | `background: var(--bg); color: var(--text); font-family: var(--font-sans)` | Page inherits — desired |
| `button, input, select, textarea` | `font: inherit` | Buttons/inputs in WizardFooter inherit page font — desired |
| `button` | `cursor: pointer` | All buttons get pointer — desired |
| `button:disabled` | `opacity: 0.5; cursor: not-allowed` | Conflicts with `.btn:disabled { opacity: 0.45 }` — `.btn` rule overrides via specificity |
| `input:not([type=...]), select, textarea` | `width: 100%; border-radius: var(--r-2); border; background; color; padding` | **Leaks.** Any bare `<input>` / `<select>` in StepScope will get full-width + global styles. StepScope has no inputs in Step 1 profile grid — no reset needed for Step 1. |
| `textarea` | `resize: vertical` | No textarea in Step 1 — no impact |
| `table` | `width: 100%; border-collapse: collapse` | No table in Step 1 — no impact |

### class-scoped selectors (impact only if class used)

| Selector | Rule | Impact |
|---|---|---|
| `.hero h1` | `margin; font-size: clamp(2rem, 4vw, 4rem); line-height: 0.95` | TemplateWizardPage does not use `.hero` — no impact |
| `.card h3, .panel-heading .kicker` | `text-transform; letter-spacing; font-size: 0.72rem; color: var(--muted)` | WizardShell uses `.card` class → any `<h3>` inside step cards gets uppercased. StepScope uses `.h2`/`.h3` utility classes, not bare `<h3>` tags — verify in page assembly |
| `.approvals-list li p` | `margin: 0.2rem; color: var(--muted)` | Not used in wizard — no impact |
| `.panel-heading h2` | `margin: 0.2rem 0 0` | Not used in wizard — no impact |

**Reset needed in page CSS Module:** none for Step 1 (no bare inputs/selects in profile grid). Monitor for `.card h3` if StepScope renders any `<h3>` tags inside a `.card` wrapper.

---

## 6. New files created

| Path | Description |
|---|---|
| `features/taxonomy/queries/useProfilesQuery.ts` | Canonical home for profiles query |
| `features/taxonomy/queries/constants.ts` | STALE_FIVE_MINUTES for taxonomy queries |
| `features/shared/components/wizard/WizardShell.tsx` | Parameterized shared shell |
| `features/shared/components/wizard/WizardShell.module.css` | CSS for shared shell (token-fixed copy) |
| `features/shared/components/wizard/WizardFooter.tsx` | Shared footer (verbatim copy) |
| `features/templates/pages/TemplateWizardPage.tsx` | Stub page (returns `<div>Loading…</div>`) |
| `features/templates/components/wizard/steps/StepScope.tsx` | Stub step (returns `<div />`) |
| `features/templates/state/templateWizard.reducer.ts` | Step 1 reducer + action types |

---

## 7. Route stub commit hash

`c44e07cd` — `feat(templates): register /templates/new route with TemplateWizardPage stub`

Full commit log for this phase:
- `cad11410` — `refactor(taxonomy): move useProfilesQuery from documents to taxonomy queries`
- `815638bc` — `refactor(shared): promote WizardFooter to features/shared/components/wizard`
- `4ad5a4ed` — `refactor(shared): promote WizardShell to features/shared, parameterize kicker/title/steps`
- `262e504f` — `fix(shared): replace raw values with design tokens in WizardShell`
- `c44e07cd` — `feat(templates): register /templates/new route with TemplateWizardPage stub`
- `ee270e9a` — `feat(templates): add templateWizard reducer (step 1 + profile fields)`

---

## 8. tsc result

The worktree does not have `node_modules` (pnpm symlinks not set up for worktree). Running `tsc` via the main repo's binary against both trees shows:

- **Main repo (with node_modules):** 30 lines of pre-existing errors in `useAreasQuery.ts`, `LibrarySidebar.tsx`, `Rail.tsx`, `NewDocumentWizardPage.tsx`, `useAuthSession.test.tsx` — zero errors in any file touched by this phase.
- **Worktree (no node_modules):** ~196 lines of `TS2307: Cannot find module` errors for `react-router-dom`, `@tanstack/react-query`, `zustand`, etc. — all environment errors, not type errors from our code.

**Conclusion:** No new type errors introduced. Pre-existing baseline errors unchanged.

---

## 9. Anything skipped + why

- **`NewDocumentWizardPage.tsx` import not changed to shared WizardShell directly** — The documents `WizardShell` wrapper retains the numeric `WizardStepNumber` API that `NewDocumentWizardPage` depends on. Updating `NewDocumentWizardPage` to use the shared component directly would require changing the call site's type from `WizardStepNumber` to `string` and passing `steps` + `kicker` + `title` as props. This is a valid next step but risks breaking the existing test mock (`vi.mock('../components/wizard/WizardShell')`). The shim approach keeps `NewDocumentWizardPage` unchanged as instructed.

- **`features/documents/components/wizard/WizardShell.module.css` not updated** — The task says only fix the shared copy. The documents wrapper now imports `WizardShell.module.css` via the shared WizardShell, so the local copy is no longer used by the documents wrapper (which delegates rendering to the shared component). No action needed.

---

## Open questions

None. All ambiguities from Phase 1 were resolved.
