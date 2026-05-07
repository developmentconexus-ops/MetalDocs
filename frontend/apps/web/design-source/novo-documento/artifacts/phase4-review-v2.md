# Screen Review: novo-documento (v2)

**Implementation:** `frontend/apps/web/src/features/documents/`
**Design source:** `frontend/apps/web/design-source/novo-documento/`
**Prior review:** `design-source/novo-documento/artifacts/phase4-review.md` (REQUEST CHANGES)
**Verdict:** REQUEST CHANGES

---

## Critical

- [ ] **`src/features/documents/components/LibrarySidebar.tsx:44-48`** - Cache key collision with wizard areas query.
  LibrarySidebar registers `QK.taxonomy.areas()` via `listTaxonomyAreas` (returns envelope
  `{ items: TaxonomyAreaItem[] }`). `useAreasQuery.ts:9` registers the **same key** via `fetchAreas`
  (returns unwrapped `ProcessArea[]`). Navigate /documents to /documents-v2/new: TanStack Query
  serves cached envelope. `NewDocumentWizardPage.tsx:68` (`const areas = areasQuery.data ?? []`)
  receives `{ items: [...] }` not an array. `TypeError: areas.find is not a function` confirmed
  in browser console.
  *Why:* `wiki/architecture/frontend-structure.md` Server State - one key, one canonical shape.
  *Fix:* Give LibrarySidebar distinct key (e.g. `QK.taxonomy.areasFlat()`) or migrate to `fetchAreas`.

---

## Major

### Architecture

- [ ] **`src/features/documents/queries/useProfilesQuery.ts:9`** - Same cache-collision risk for profiles.
  Verify `QK.taxonomy.profiles()` not registered elsewhere with differing return shape.
  *Why:* Same rule - one key, one canonical shape.
  *Fix:* Audit all callers of `QK.taxonomy.profiles()` for shape parity.

### Visual / Tokens

- [ ] **`src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:92`** -
  `color: var(--text-on-brand, #fff)` uses undeclared token. `--text-on-brand` absent from
  `src/styles/tokens.css` and `@metaldocs/shared-tokens`. Fallback `#fff` works but token is drift.
  *Why:* Phase 3b token-coverage rule - all `var()` references must resolve to declared token.
  *Fix:* Declare `--text-on-brand: #ffffff` in `tokens.css` (brand cluster), remove fallback value.

- [ ] **`src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:56-58`** -
  Stale TODO about selected-state icon-tile color. Override implemented at lines 87-93. TODO not removed.
  *Fix:* Delete stale TODO block at lines 56-58.

- [ ] **`src/features/documents/components/wizard/steps/StepConfirm.module.css:86,101`** -
  Raw `font-size: 16px` (.docTitle) and `font-size: 12px` (.fieldGrid). Acknowledged as
  typography-token defer in comment, but `artifacts/token-coverage.txt` absent - no formal tracking.
  *Why:* `skills/metaldocs-screen-implementation/SKILL.md` Phase 3b hard gate requires `token-coverage.txt`.
  *Fix:* Create `design-source/novo-documento/artifacts/token-coverage.txt` with all raw values + justification.

- [ ] **`src/features/documents/components/wizard/steps/StepTemplate.module.css`** - Same gap:
  `token-coverage.txt` absence means no audit trail for raw values in file.
  *Fix:* Covered by same `token-coverage.txt` artifact.

### Error UX

- [ ] **`src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:97`** -
  Area-fetch error container (`<div role="alert" aria-live="assertive">`) lacks `className="card"`.
  All other error containers in the wizard (StepTemplate, StepConfirm) use the `.card` class.
  *Why:* `wiki/concepts/error-ux.md` - inline error containers use card scaffolding consistently.
  *Fix:* Add `className="card"` to the error div at line 97.

---

## Minor

- [ ] `artifacts/token-coverage.txt` missing - Phase 3b hard gate artifact not present in
  `design-source/novo-documento/artifacts/`. Blocks formal sign-off.
  *Why:* `skills/metaldocs-screen-implementation/SKILL.md` Phase 3b.

- [ ] `artifacts/` screenshots missing - Phase 4 hard gate requires 3-viewport screenshots
  (1440, 1024, 375) per step. Directory contains only review .md files.
  *Why:* Same skill Phase 4.

- [ ] `src/features/documents/components/wizard/steps/StepConfirm.tsx:83` -
  `<span className="pill mono">v1</span>` mixes global utility class `mono` inline with
  module-scoped context. Minor consistency note - not a token violation.

---

## Prior Criticals - Confirmed Fixed

- **Subcontrols not disabled (v1 Critical):** `PeopleSubcontrols` and `ExternalSubcontrols`
  now render with `aria-disabled="true"`, `disabled` on all inputs, and container applies
  `.subcontrolsCard` with `opacity: 0.6; pointer-events: none`. Fixed.

- **`selectProfile` empty-code corruption (v1 Critical):** `wizard.reducer.ts:69` guards
  `if (!action.code.trim()) return state;` before any state mutation. `clearProfile` resets
  step to 1 and nulls profile and template selections. Fixed.

---

## What Is Good

- Mutation flow in `NewDocumentWizardPage.tsx` correctly chains slot-create then doc-create
  via nested `mutateAsync`, with `onError` catching `ApiError` through `resolveErrorMessage`
  into sonner `toast.error` - textbook error UX per `wiki/concepts/error-ux.md`.

- `StepAreaCodeVisibility.tsx` visibility-icon selected state driven by `SelectableCard`
  primitive `data-selected` attribute rather than duplicating selection logic - correct
  primitive composition per `wiki/architecture/frontend-structure.md`.

- `wizard.reducer.ts` `maxReachableStep` + `clampStep` enforces forward-only step guard at
  the state layer, not the UI layer - correct separation of concerns.
