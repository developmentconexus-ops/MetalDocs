# Screen Review: novo-template-escopo (final re-review)

**Implementation:** `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx` + `features/templates/components/wizard/steps/StepScope.tsx`
**Design source:** `frontend/apps/web/design-source/novo-template-escopo/`
**Visual comparison:** Partial — computed-style probes from the prior review stand; no new visual regressions introduced by the Critical fix (import path change only, no layout or style changes).
**Verdict:** APPROVE WITH NITS

---

## Critical

None.

---

## Major

None.

---

## Minor

- [ ] `frontend/apps/web/src/components/ui/SelectableCard.module.css` — `transition: border-color 120ms ease, background 120ms ease, box-shadow 120ms ease` has no `@media (prefers-reduced-motion: reduce) { transition: none }` block. This screen renders `SelectableCard` for every profile card. *Why:* `wiki/architecture/frontend-structure.md § 5E A11y` — every animated element requires a reduced-motion fallback. Pre-existing primitive gap, not introduced by this screen. *Fix:* Add `@media (prefers-reduced-motion: reduce) { .root { transition: none; } }` to `SelectableCard.module.css`. Acceptable for backlog.

- [ ] `frontend/apps/web/design-source/novo-template-escopo/artifacts/parity-diff.md:19-22` — `profileCode` and `profileName` rows remain "inferred from class; not directly snapshotted." *Why:* `.claude/skills/metaldocs-screen-implementation/SKILL.md` Phase 3b — measured values required, not inferred. *Fix:* Run `getComputedStyle` on `.profileCode` and `.profileName` and fill in ref/impl/delta columns with actual numbers. Hygiene deficit only; does not affect runtime behavior.

---

## Critical fix verification

### C1 (prior review): `StepAreaCodeVisibility` broken import after WizardFooter shim deletion

**RESOLVED.** `StepAreaCodeVisibility/index.tsx:9` now reads:

```
import { WizardFooter } from '../../../../../shared/components/wizard/WizardFooter';
```

Resolves to `features/shared/components/wizard/WizardFooter`. Correct canonical path. The shim at `features/documents/components/wizard/WizardFooter.tsx` does not exist (Glob returns no match). No other callers of the deleted path found. Crash loop is cleared.

---

## All previously reported issues — final status

| ID | Finding | Status |
|----|---------|--------|
| C1 (prior) | StepAreaCodeVisibility broken import after shim deletion | RESOLVED |
| C2 | Re-export shim `documents/components/wizard/WizardFooter.tsx` | RESOLVED |
| C3 | Duplicate route + path rename (`templates/novo` → `templates/new`) | RESOLVED |
| M1 | WizardFooter divider hardcoded `-32px` | RESOLVED |
| M3 | `.disabledBadge` undefined token with raw hex fallback | RESOLVED |
| M4 | Disabled SelectableCard missing `title="Em breve"` | RESOLVED |
| M5 | `wiki/backlog/novo-template-wizard.md` did not exist | RESOLVED |
| M6 | `wiki/modules/templates.md` stale | RESOLVED |
| m1 | IMPLEMENTATION.md Open Questions Log Q1 unresolved | RESOLVED |
| minor-1 | SelectableCard missing prefers-reduced-motion | CARRIED — pre-existing primitive gap |
| minor-2 | parity-diff inferred rows not measured | CARRIED — hygiene only |

---

## What's good

- Every Critical and Major from the prior two review passes is now closed. The fix was surgical — one import path, no collateral changes.
- `features/documents/queries/useProfilesQuery.ts` confirmed deleted. No re-export shims remain.
- `features/templates/routes.tsx` is clean: one lazy entry for `templates/new`, no duplicates, no `HashRouter`, no string-pattern dispatchers.
- `StepScope.module.css` `.disabledBadge` correctly uses `var(--surface-3)` with no raw hex fallback.
- `StepScope.tsx` disabled SelectableCard has `title={isDisabled ? 'Em breve' : undefined}` — correct a11y pattern.
- `WizardFooter.module.css` divider uses `calc(-1 * var(--sp-7))` with responsive mobile override to `calc(-1 * var(--sp-4))`. Token-correct and responsive.
- All four WizardFooter callers (StepProfile, StepTemplate, StepConfirm, StepAreaCodeVisibility) now import from the canonical `features/shared/components/wizard/WizardFooter`. Migration complete.

---

## Iron-Law cross-check

- Phase 0 audit signed: `phase0-audit.md` gate passed; Q1 resolved. YES
- Phase 1 worksheet complete: `phase1-map.md` complete. YES
- Phase 2 primitive audit verified: `phase2-preflight.md` complete. Shim deletions now fully executed including previously missed caller. YES
- Phase 3a DOM diff approved: `phase3a-structure.md` complete. YES
- Phase 3b: parity-diff covers primary regions YES; `profileCode` / `profileName` rows still inferred not measured (minor hygiene); leakage-probe covers form elements YES; token-coverage empty YES; measured fields pass spot-checks YES
- Phase 4 behavior trace present: `phase4-behavior.md` complete. YES
- Open Questions Log resolved: YES
