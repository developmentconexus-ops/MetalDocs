# Screen Review: novo-documento

**Implementation:** `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx` + `components/wizard/**`
**Design source:** `frontend/apps/web/design-source/novo-documento/` (NOTES.md + IMPLEMENTATION.md; reference HTMLs in sibling `novo-perfil/`, `novo-area-codigo-visibilidade/`, `novo-template/`, `novo-confirmacao/` and `selected-wizard.jsx` / `selected-wizard-v2.jsx`)
**Verdict:** REQUEST CHANGES

Two Critical issues block approval: (1) a Cut item from NOTES.md ("Pessoas específicas" + "Compartilhamento externo" subcontrols are flagged as Defer/no-op but were rendered as fully interactive forms — the "Em branco" template is similarly active in the disabled card path is fine but the subcontrols issue presents UI that implies unsupported behavior); (2) the wizard reducer's `selectProfile` is dispatched with an empty string in a "safety net" that the reducer accepts as a valid code (`NewDocumentWizardPage.tsx:88-91`), which silently corrupts state. The remainder of the implementation is solid: tokens used pervasively, primitives respected, reducer + URL sync clean, error UX wired with `ApiError + resolveErrorMessage + role="alert"`.

---

## Critical

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:203-289` — `PeopleSubcontrols` and `ExternalSubcontrols` render fully interactive forms (add/remove invitees, password/watermark/expiry checkboxes + number input) but the values are **not submitted** (TODO comments at :204 and :250). NOTES.md `Audit decision` table (lines 36-37) marks invitee chips and external sub-controls as Defer with "No-op" submit; rendering them fully interactive misleads operators into believing share controls work. *Why:* `wiki/concepts/design-workflow-audit.md` — UI must not imply unsupported behavior; Defer items half-implemented in a misleading way → Critical (per agent rule §4.C). *Fix:* either gate behind `disabled aria-disabled="true" title="Em breve"` (matching the "Em branco" treatment in StepTemplate.tsx:128-148) or remove until backend exists. Don't render the inputs as live.

- [ ] `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:85-92` — On profile-list refetch, if `state.profileCode` is not in the loaded list, the page dispatches `selectProfile` with an empty string. The reducer (`wizard.reducer.ts:65-74`) accepts any string and stores `profileCode: ''`. `maxReachableStep` then treats `''` as truthy and allows step ≥ 2, while `canAdvance` step-1 check (`profileCode !== null`) also passes for `''`. State is corrupted: no profile selected, but UI permits advance. The author's own comment at :90-91 acknowledges this ("invalid — instead, dispatch a goToStep + reset"). *Why:* `wiki/architecture/frontend-structure.md` — reducer must enforce invariants; goal-driven execution rule (CLAUDE.md §4) — no half-fixed safety nets. *Fix:* introduce a `clearProfile` action that resets `profileCode: null` + `templateID/templateVersionID: null` + clamps `step` to 1, and use it here.

## Major

### Architecture

- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:111-156` — `handleCreate` is an `async` function on the page rather than a `useMutation`. Other state-changing flows in this codebase use TanStack Query mutations for retry, status, and `onError` toasts. *Why:* `wiki/architecture/frontend-structure.md § State` + `wiki/concepts/error-ux.md` — mutations should run through `useMutation` + `onError` → `resolveErrorMessage` → sonner toast. *Fix:* wrap the two-call sequence in `useMutation`; surface `mutation.error` via toast in addition to the inline `role="alert"` on Step 4.

- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:148-154` — Errors are surfaced inline via `role="alert"` (good) but **no sonner `toast.error(...)` is fired**. *Why:* `wiki/concepts/error-ux.md` — user-facing errors should also trigger sonner toast for visibility when the user has scrolled away from the alert region. *Fix:* add `toast.error(resolveErrorMessage(...))` alongside the dispatch.

### Visual / tokens

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:79-83` — `.visibilityLabel` declares `font-size: 13px` and `font-weight: 500` — raw `13px` is not in the spacing/font-size token set. The IMPLEMENTATION.md §3b.1 Typography note (line 329) acknowledges "No font-size tokens in current scale. One-off raw px with TODO comment". The TSX/CSS does not include the TODO comment for these one-offs. *Why:* `frontend/apps/web/src/styles/tokens.css` is the source of truth; `metaldocs-frontend` SKILL — flag raw `px` for spacing/typography unless TODO + token-deferred. *Fix:* add a `TODO(novo-documento:typography-tokens)` comment block above each raw font-size declaration in this file (and in StepProfile.module.css:24-26, StepTemplate.module.css:84, StepConfirm.module.css:84,97 etc.) referencing a backlog row, OR introduce a `--fs-*` scale in `tokens.css` and migrate.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.module.css:115-118` — `.fieldValue` is an empty rule. Either delete it or document why it exists. *Why:* `metaldocs-frontend` SKILL — no dead CSS. *Fix:* remove the empty selector; the comment "kept empty for symmetry and for Phase 3c hooks" should not survive into Phase 4.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:51-68` — TODO at :55-58 says the visibility icon-tile "selected state changes background to var(--brand) + color white; wired in Phase 3c when SelectableCard exposes selected context to children". Phase 3c is checked off in IMPLEMENTATION.md (line 354), but this TODO is unaddressed: selected visibility cards still show idle icon-tile colors. *Why:* `wiki/concepts/design-workflow-audit.md` — Keep items must be implemented; the design's selected-state visual indicator is a Keep. *Fix:* either pass `selected` down via a prop on the icon span, or expose context from `SelectableCard` and react in CSS via a `[data-selected]` attribute selector.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:208-210, 256-256` — Inline `style={{ marginTop: 'var(--sp-3)' }}` + `style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--sp-2)', ... }}` and identical patterns scattered through `PeopleSubcontrols` / `ExternalSubcontrols`. *Why:* `metaldocs-frontend` SKILL — no inline `style` for layout; CSS Modules only. *Fix:* hoist into `StepAreaCodeVisibility.module.css` as `.subcontrolsCard`, `.inviteeChips`, `.externalForm`. (Less critical because the path is Defer-deprecated, but if subcontrols stay in the tree they need real styles.)

### NOTES.md compliance

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:73-76` — TODO acknowledges that older versions are not selectable, but renders no per-version radio UI at all. NOTES.md §Audit (line 39) says "older versions grayed + tooltip 'Em breve'" — implementation collapses this to per-template treatment (only published-version-bearing templates selectable). IMPLEMENTATION.md §3c "Note" (line 366) records this as a Phase 4 review item — review verdict: this is a Major drift from NOTES.md contract. *Why:* `wiki/concepts/design-workflow-audit.md` — Defer items half-implemented in a misleading way → Major. *Fix:* either render disabled per-version radios with `aria-disabled="true" title="Em breve"`, or update NOTES.md to record the simplification (per-template only).

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:104-112` — "publicada" pill is shown for selectable templates with `<span className="dot" />` — this depends on `.dot` being a global utility. Verify it exists in tokens/styles; otherwise the dot won't render. *Why:* `metaldocs-frontend` SKILL — utility classes must come from token CSS. *Fix:* confirm `.dot` exists in `styles/tokens.css` or `styles/components.css`; if not, replace with a tokenized inline span or scoped module class.

### Error UX

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepProfile.tsx:50-61`, `StepAreaCodeVisibility.tsx:96-105`, `StepTemplate.tsx:55-66` — Error containers use `<div role="alert" className="card">` with the retry button inside `<div>`. The retry `<button>` lives inside an `<div>` inside the alert — fine — but the alert lacks an `aria-live` region for re-announcement on retry-failure. *Why:* `wiki/concepts/error-ux.md` — alerts should announce; using `role="alert"` is OK but if the same node persists across retries the change won't re-announce. *Fix:* add `aria-live="assertive"` or remount the alert when error identity changes (key by error message hash).

### A11y

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:217-225` — Invitee remove button is `<button class="btn btn-ghost btn-sm">` nested inside a `<span class="pill">`. The pill is not a button, so this is OK semantically, but the `×` glyph has no `aria-hidden="true"` and the surrounding `aria-label` is on the button — fine. However the pill itself has no role — screen readers will read "Convidado (placeholder) Remover Convidado (placeholder) ×". *Why:* `wiki/concepts/error-ux.md` is silent; general a11y. *Fix:* wrap the `×` in `aria-hidden="true"` or use `<span aria-hidden="true">×</span>` as the button content.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:116-126` — Consent label wraps both `<input type="checkbox">` and a `.mono` span containing the code. Implicit labeling works, but the consent text length (~25 words) inside `<label>` makes screen-reader announcement verbose. *Why:* general a11y. *Fix:* keep `<label htmlFor="wizard-consent">`, move long descriptive text to a sibling `<p id="wizard-consent-desc">` linked via `aria-describedby="wizard-consent-desc"`.

- `frontend/apps/web/src/features/documents/components/wizard/WizardShell.tsx:30-34` — `Stepper.onStepClick` is wired, allowing keyboard arrow nav (good — `Stepper.tsx:19-32`). However the `<button class="btn">` Cancel/Voltar/Avançar in `WizardFooter` (`:67-84`) lacks any `:focus-visible` proof — relies on global `.btn` styles. *Why:* `metaldocs-frontend` SKILL — verify focus visible. *Fix:* spot-check `styles/components.css` (or wherever `.btn` lives) for `:focus-visible` styling; add if missing.

### Responsive

- `frontend/apps/web/src/features/documents/components/wizard/WizardShell.module.css:4-9` — `.scrollWrapper { padding: var(--sp-7) var(--sp-7); }` — 32px horizontal padding at 375px viewport leaves ~311px for content. The `.previewCard` in StepConfirm uses `grid-template-columns: 120px 1fr` (`StepConfirm.module.css:9`). At 311px, the 1fr column gets ~175px — tight but acceptable. *Why:* `metaldocs-frontend` SKILL — flag horizontal overflow at 375. *Fix:* add a `@media (max-width: 600px)` rule to drop the previewCard to one column and reduce `.scrollWrapper` horizontal padding to `var(--sp-4)`. Verify at 375 with preview tools before merge.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:4-9` — `.profileAreaRow` 2-col grid does not collapse on narrow viewports. Same for `.visibilityGrid` (:35-40). *Why:* same. *Fix:* `@media (max-width: 600px) { grid-template-columns: 1fr; }` for both.

## Minor

- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:23` — `import { StepConfirm }` mixes default + named imports across step files (StepConfirm exports both). Pick one style. *Fix:* use named imports throughout.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:80, 167-178` — Helper named `selected_` (trailing underscore) is unusual. *Fix:* rename `isTemplateSelected`.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:74` — `<div className={`${styles.thumbnailCode} mono`}>` mixes module class and global utility. Works, but consider applying `font-family: var(--font-mono)` directly in `.thumbnailCode` for self-containment. *Fix:* optional.

- `frontend/apps/web/src/features/documents/components/wizard/WizardShell.tsx:79` — `className={primaryVariant === 'submit' ? 'btn btn-primary' : 'btn btn-primary'}` — both branches identical. *Fix:* drop the ternary; just use `'btn btn-primary'`. Or differentiate (e.g. `btn-success` for submit).

- `frontend/apps/web/src/features/documents/state/wizard.reducer.ts:65-74` — `selectProfile` reducer accepts any string including `''`. The TS type allows it. *Fix:* tighten the action type to require a non-empty branded string, or guard with `if (!action.code) return state;`.

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:145-145` — "cuidado" pill text inside the disabled "Em branco" card is literal Portuguese for "warning" — likely should be the visible label "em breve" matching the title. *Fix:* unify copy.

- `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.tsx:14-16` — Caption duplicates the `≈ POP-QUA-???` already shown via `.code`. Tooltip + caption + code is redundant. *Fix:* drop one (keep tooltip + code; remove caption duplication).

- `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.module.css:17-18` — TODO comment proposes typography token scale; create `wiki/backlog/novo-documento.md#typography-tokens` row (or whichever existing backlog file) so it's not orphaned. *Fix:* file the backlog row.

## What's good

- Reducer-driven state (`wizard.reducer.ts`) with clean action types, derived `maxReachableStep` / `clampStep` / `canAdvance` helpers — exactly the pattern called for in `wiki/architecture/frontend-structure.md`.
- URL `?step=N` sync via `useSearchParams` (page :52-58) — refresh-safe and back-button-correct.
- Token discipline in CSS Modules is exemplary: every spacing/border/color value cites the design source line and rounds to a token with a comment explaining the rounding (`StepProfile.module.css`, `StepConfirm.module.css`, `WizardShell.module.css`).
- `SelectableCard` primitive (`components/ui/SelectableCard.tsx`) cleanly composes radio semantics (`role="radio"`, `aria-checked`, `aria-disabled`) and is reused across all three card grids without per-page CSS overrides — textbook primitive composition.
- Visibility option B (icon remap) is consistently applied (`visibilityMeta.ts:24-45`) with comments explaining the choice.
- Error UX scaffolding present in every step (`StepProfile.tsx:50-61`, `StepAreaCodeVisibility.tsx:96-105`, `StepTemplate.tsx:55-66`) with `ApiError` + `resolveErrorMessage` + retry button.
- TODO trail is disciplined: every Defer item carries an inline comment referencing `wiki/backlog/novo-documento.md#<anchor>` (visibility, sharing, slot-rollback, blank-template, template-versions, profile-counts).
- Stepper primitive (`Stepper.tsx`) is correctly generic — keyboard arrow nav, `aria-current="step"`, focus management — and can be reused for future wizards.
