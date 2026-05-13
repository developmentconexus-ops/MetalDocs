# Module: templates-v2


> **Status:** predecessor (frontend-heavy historical page)

> **Maturity:** L1
> **Scope:** Template authoring, versioning, approval, publishing; Templates List screen (`/templates-v2`); Template creation wizard (`/templates-v2/new`).
> **Out of scope:** Document fill-in (see `modules/documents.md`), eigenpal editor wiring (see `modules/editor-ui-eigenpal.md`), toolbar overlay + eigenpal CSS overrides (see `modules/editor-chrome.md`).
> **Key files:**
> - `internal/modules/templates_v2/` Ã¢â‚¬â€ backend module
> - `frontend/apps/web/src/features/templates/TemplatesListPage.tsx:1` Ã¢â‚¬â€ list page; real query, tab filter, loading/error/empty states, semantic clickable cards
> - `frontend/apps/web/src/features/templates/queries/useTemplatesQuery.ts:1` Ã¢â‚¬â€ TanStack Query hook wrapping `listTemplates()`; key `QK.templates.list()`
> - `frontend/apps/web/src/features/templates/components/TemplateCard.tsx:1` Ã¢â‚¬â€ card primitive (MiniDocPreview + StatusPill + CodeChip + Avatar); `role="button"` + keyboard handler
> - `frontend/apps/web/src/features/templates/components/MiniDocPreview.tsx:1` Ã¢â‚¬â€ decorative A4 thumbnail (8 placeholder lines)
> - `frontend/apps/web/src/components/ui/WorkspaceHeroHeader.tsx:13` Ã¢â‚¬â€ `tone?: "banner" | "flat"` prop; `.headerFlat` strips card chrome (transparent bg, no border-bottom, padding 0)
> - `frontend/apps/web/src/components/ui/TabBar.tsx:17` Ã¢â‚¬â€ roving tabIndex (`tabIndex={isActive ? 0 : -1}`), Arrow/Home/End keyboard nav per WAI-ARIA tablist spec
> - `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx:1` Ã¢â‚¬â€ creation wizard; `useReducer(templateWizardReducer)` + URL sync `?step=N`; `selectedProfile` derived via `useMemo` from profiles query; defensive guard sends `?step=2` without scope back to step 1; `advanceDisabled = selectMaxReachableStep(state) <= state.step` is the single advance gate passed to every step (no per-step disabled vars); Step 5 (`StepConfirmation`) receives no `advanceDisabled` Ã¢â‚¬â€ its gate is the internal checkbox; `handleSubmit` mocked (`navigate('/templates-v2')`); back/forward preserves all state; `export { TemplateWizardPage as Component }` for React Router lazy
> - `frontend/apps/web/src/features/templates/state/templateWizard.reducer.ts:1` Ã¢â‚¬â€ wizard reducer SSOT; public `templateWizardReducer` wraps private `reduceCore` with auto-clamp: after every action, `selectMaxReachableStep(next)` is evaluated and `step` is clamped down if it exceeds the max (prevents URL-injection into a future step); exported `selectMaxReachableStep(state): TemplateWizardStep` encodes all advance-gate logic (step 1 requires `scopeType`, step 2 requires `name Ã¢â€°Â¥ 3`, step 3 requires `startingPoint` + docx when applicable, step 4 requires Ã¢â€°Â¥1 role when `permissionsMode==='roles'`); `TemplateWizardStep = 1|2|3|4|5`; `ScopeType = 'generic' | 'profile'`; `StartingPoint = 'docx' | 'blank'`; `PermissionsMode = 'roles' | 'areas' | 'all'`; `initialState.permissionsMode = 'all'`; actions: `GO_TO_STEP | SET_SCOPE_TYPE | SET_PROFILE | SET_NAME | SET_DESCRIPTION | SET_STARTING_POINT | SET_SELECTED_DOCX | CLEAR_SELECTED_DOCX | SET_PERMISSIONS_MODE | TOGGLE_ROLE_ID | TOGGLE_AREA_ID`
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx:1` Ã¢â‚¬â€ Step 1: profile picker; `DISABLED_PROFILES = new Set(['CHK'])` with TODO for API flag
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepIdentity.tsx:1` Ã¢â‚¬â€ Step 2: name + description fields; mocked code preview; scope recap row with "Trocar" back to Step 1
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepIdentity.module.css:1` Ã¢â‚¬â€ Step 2 CSS; `.descriptionInput` overrides global `.input { height: 32px }` with `height: auto; min-height: 72px` (see `concepts/css-leakage-offenders.md`)
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.tsx:1` Ã¢â‚¬â€ Step 3: two starting-point cards (`docx` | `blank`); hidden `<input type="file">` triggered by card click; mocked file selection (filename + bytes echo, no upload); `Substituir` clears + re-opens picker; switching to `blank` clears docx state; advance gated on `startingPoint !== null && (startingPoint !== 'docx' || selectedDocxName !== null)`
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.module.css:1` Ã¢â‚¬â€ Step 3 CSS; `.startingPointGrid`, `.startingPointCard`, `.selected`, `.fileRow`, `.fileInput` (visually hidden), `formatBytes` helper inline
> - `frontend/apps/web/src/features/taxonomy/queries/useProfilesQuery.ts:1` Ã¢â‚¬â€ shared profiles query (used by both documents and templates wizards)
> - `frontend/apps/web/src/features/shared/components/wizard/WizardShell.tsx:1` Ã¢â‚¬â€ parameterized wizard chrome; `kicker/title/description/steps/currentStep/children`
> - `frontend/apps/web/src/features/shared/components/wizard/WizardFooter.tsx:1` Ã¢â‚¬â€ shared footer; `stepLabel/primaryDisabled/showBack/onAdvance/onBack/onCancel`
> - `frontend/apps/web/src/features/templates/TemplateCreateDialog.tsx` Ã¢â‚¬â€ new template dialog (superseded by wizard, kept for rollback)
> - `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:1` Ã¢â‚¬â€ eigenpal author (renamed from `TemplateAuthorPage` 2026-05-10); inner rail (back / variables / outline) + `PlaceholderCatalogPanel` or `TemplateOutlinePanel` + `EditorChrome` overlay; `submitForReview` + `importDocx` errors funnel through `ApiError` + `resolveErrorMessage`; pt-BR copy throughout (`Submeter para revisÃƒÂ£o`, `Importar .docx`, `SalvandoÃ¢â‚¬Â¦/Salvo/Falha ao salvar`)
> - `frontend/apps/web/src/features/templates/pages/styles/TemplateEditorPage.module.css:1` Ã¢â‚¬â€ page CSS (tokens-only); rail, panel rules, alert variants (`alertError`, `alertSuccess`)
> - `frontend/apps/web/src/features/templates/pages/TemplateEditorRoutePage.tsx:1` Ã¢â‚¬â€ route entry (renamed from `TemplateAuthorRoutePage`); URL `/templates-v2/:templateId/versions/:versionNum`
> - `frontend/apps/web/src/features/templates/TemplateOutlinePanel.tsx:1` Ã¢â‚¬â€ read-only headings panel (`Heading[]` from `readHeadings`); empty-state copy "Aplique estilos de tÃƒÂ­tuloÃ¢â‚¬Â¦"; `data-level` attribute drives indent
> - `frontend/apps/web/src/features/templates/lib/readHeadings.ts:1` Ã¢â‚¬â€ derives `Heading[]` from eigenpal `agent.getAgentContext().outline` (`ParagraphOutline[]` filtered by `isHeading`); levels clamped 1..6; safe against missing `getAgent`/`getAgentContext`
> - `frontend/apps/web/src/features/templates/api/templatesV2.ts:240` Ã¢â‚¬â€ `submitForReview` now uses `apiFetch` (throws `ApiError`)
> - `frontend/apps/web/src/features/templates/VersionActionPanel.tsx` Ã¢â‚¬â€ lifecycle transitions
> - `frontend/apps/web/design-source/templates/artifacts/` Ã¢â‚¬â€ phase 0Ã¢â‚¬â€œ5 implementation artifacts (list screen)
> - `frontend/apps/web/design-source/novo-template-escopo/artifacts/` Ã¢â‚¬â€ phase 0Ã¢â‚¬â€œ5 implementation artifacts (creation wizard Step 1)
> - `frontend/apps/web/design-source/novo-template-identidade/artifacts/` Ã¢â‚¬â€ phase 0Ã¢â‚¬â€œ5 implementation artifacts (creation wizard Step 2)
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.tsx:1` Ã¢â‚¬â€ Step 4: three-mode segmented control (`roles` | `areas` | `all`); ARIA radiogroup with roving tabindex; role/area card grids (`role="checkbox"` + `aria-checked`); coverage summary row (mocked counts); advance gated on `roles` mode requiring Ã¢â€°Â¥1 selection
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.module.css:1` Ã¢â‚¬â€ Step 4 CSS; `.modeSegmented`, `.modeTab`, `.modeTabActive`, `.roleGrid`, `.areaGrid`, `.roleCard`, `.areaCard`, `.cardSelected`, `.coverageSummary`
> - `frontend/apps/web/design-source/novo-template-estrutura/` Ã¢â‚¬â€ design source dir for creation wizard Step 3 (Estrutura)
> - `frontend/apps/web/design-source/novo-template-permissoes/` Ã¢â‚¬â€ design source dir for creation wizard Step 4 (PermissÃƒÂµes)
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepConfirmation.tsx:1` Ã¢â‚¬â€ Step 5: review card (thumb + meta grid) + confirmation checkbox; `HIGHLIGHTED_THUMB_LINES` Set drives alternating decorative line pattern; `confirmed` local state gates "Criar e abrir editor Ã¢â€ â€™"; `handleSubmit` currently calls `onSubmit()` which navigates to list (mocked Ã¢â‚¬â€ see `backlog/novo-template-wizard.md#confirmacao-backend-submit`)
> - `frontend/apps/web/src/features/templates/components/wizard/steps/StepConfirmation.module.css:1` Ã¢â‚¬â€ Step 5 CSS; `.previewCard`, `.thumb`, `.thumbHeader`, `.thumbCode`, `.thumbLine`, `.thumbLineHighlight`, `.previewBody`, `.headerRow`, `.templateName`, `.metaGrid`, `.metaRow`, `.metaLabel`, `.metaValue`, `.confirmBlock`, `.confirmKicker`, `.confirmList`, `.checkLabel`

## Template Creation Wizard

**Route:** `/templates-v2/new`
**Page:** `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx:1`

5-step wizard using the same `WizardShell` + `WizardFooter` shared primitives as the document creation wizard.

### State management

`useReducer(templateWizardReducer, initialState, urlInitializer)` Ã¢â‚¬â€ URL-sync pattern: `useEffect` on `state.step` writes `?step=N` back; lazy initializer reads it on mount. Same pattern as doc wizard.

**Reducer SSOT (commit bf2f6571):** `templateWizardReducer` is a thin wrapper around private `reduceCore`. After every action, it calls `selectMaxReachableStep(next)` and auto-clamps `step` down if it exceeds the max. This eliminates per-step disabled state in the page Ã¢â‚¬â€ the page computes `advanceDisabled = selectMaxReachableStep(state) <= state.step` once and passes it to every step component. `selectMaxReachableStep` is the single source of truth for all advance-gate logic.

### Step 1 Ã¢â‚¬â€ Escopo (profile picker)

Profile cards from `useProfilesQuery` (taxonomy). CHK hardcoded as disabled until Checklist feature ships Ã¢â‚¬â€ see `wiki/backlog/novo-template-wizard.md#chk-disabled`.

### Step 2 Ã¢â‚¬â€ Identidade (name + description)

**Shipped 2026-05-09.** Component: `StepScope.tsx` Ã¢â€ â€™ `StepIdentity.tsx`.

**Inputs:**
- `nome` (required, Ã¢â€°Â¥ 3 chars trimmed) Ã¢â‚¬â€ max 120 chars, `aria-required`
- `descriÃƒÂ§ÃƒÂ£o` (optional) Ã¢â‚¬â€ `<textarea rows={3}`, max 500 chars

**Advance guard:** `state.name.trim().length < 3` disables the primary button. Label changes to "Informe o nome para continuar" when blocked.

**Scope recap row:** shows the selected profile (`code-chip` + name + family) or "GenÃƒÂ©rico" for generic scope. "Trocar" and "Voltar" both dispatch `GO_TO_STEP(1)` Ã¢â‚¬â€ scope + profile state is preserved.

**Code preview (mocked):** `TPL-{PROFILE_UPPER|GEN}-XXX`. No real `next-code` endpoint yet Ã¢â‚¬â€ see `wiki/backlog/novo-template-wizard.md#next-code-preview`. When endpoint ships, replace with `useNextTemplateCodeQuery(profileCode)`.

**Version label:** static `v1.0` Ã¢â‚¬â€ first version is always v1.0; increments on publish.

**CSS note:** `StepIdentity.module.css:.descriptionInput` must override the global `.input { height: 32px }` rule (set `height: auto; min-height: 72px`). Any future step with a `<textarea>` must apply the same override. See `wiki/concepts/css-leakage-offenders.md`.

**Defensive guard (page level):** if URL is `?step=2` but `scopeType === null` (e.g. deep-link or refresh), `TemplateWizardPage` dispatches `GO_TO_STEP(1)` immediately via `useEffect`.

### Step 3 Ã¢â‚¬â€ Estrutura (starting point)

**Shipped 2026-05-09.** Component: `StepStructure.tsx`. DOCX upload is mocked Ã¢â‚¬â€ real upload deferred to post-create editor handoff.

**Two starting points (card selection):**
- `docx` Ã¢â‚¬â€ opens native file picker (hidden `<input type="file" accept=".docx">`triggered via card click). On selection: `SET_SELECTED_DOCX(name, size)` stores filename + bytes in reducer. File is NOT uploaded at this step.
- `blank` Ã¢â‚¬â€ selects immediately; dispatches `SET_STARTING_POINT('blank')` and clears any docx state.

**Advance gate:** driven by `selectMaxReachableStep` (reducer SSOT) Ã¢â‚¬â€ page passes `advanceDisabled` computed from the selector. Step 3 blocks when `startingPoint === null` or `startingPoint === 'docx' && selectedDocxName === null`. WizardFooter label is context-aware: "Selecione um .docx para continuar" vs "Escolha um ponto de partida" vs "Pronto para avanÃƒÂ§ar".

**Substituir flow:** button dispatches `CLEAR_SELECTED_DOCX`, then immediately re-triggers the hidden file input so picker re-opens without an intermediate cleared state.

**State preservation:** back to Step 2 preserves `startingPoint` + `selectedDocxName` + `selectedDocxSize`. Forward (advance) navigates to Step 4 stub.

**Phase 0 cuts (confirmed):** placeholder chips block, auto-fill flag, file metadata processing time (`147 KB Ã‚Â· processado em 1.2s`), info banner. All cut Ã¢â‚¬â€ no backend support. See `wiki/backlog/novo-template-wizard.md#step3-placeholder-extract`.

**Deferred backlog items:** `step3-docx-upload` (presigned upload post-create), `step3-placeholder-extract` (token extraction endpoint), `step3-editor-handoff` (redirect after Step 5 create). See `wiki/backlog/novo-template-wizard.md`.

### Step 4 Ã¢â‚¬â€ PermissÃƒÂµes (visibility scope)

**Shipped 2026-05-09.** Component: `StepPermissions.tsx`.

**Three modes (segmented control, ARIA radiogroup):**
- `roles` Ã¢â‚¬â€ show role cards (QUA-INSP, PROD-OP, etc.); at least one must be selected to advance. Mocked via `MOCK_ROLES` Ã¢â‚¬â€ see `wiki/backlog/novo-template-wizard.md#permissions-roles-api`.
- `areas` Ã¢â‚¬â€ show area cards (QUA, PROD, MANÃ¢â‚¬Â¦); empty selection is valid per design. Mocked user counts Ã¢â‚¬â€ see `wiki/backlog/novo-template-wizard.md#permissions-area-counts`.
- `all` Ã¢â‚¬â€ company-wide; no selection required. Mocked total count (`COMPANY_USER_COUNT = 340`) Ã¢â‚¬â€ see `wiki/backlog/novo-template-wizard.md#permissions-user-count`.

**Advance gate:** driven by `selectMaxReachableStep` (reducer SSOT) Ã¢â‚¬â€ page passes `advanceDisabled`. Step 4 blocks when `permissionsMode === 'roles' && selectedRoleIds.length === 0`. Footer label is `'Selecione ao menos um perfil'` when blocked.

**Coverage summary:** Derived client-side from mocked counts. Roles mode: sum of `count` for each selected role. Areas mode: sum of `count` for each selected area. All mode: `COMPANY_USER_COUNT`.

**Keyboard nav:** mode segmented control uses roving tabindex (ArrowLeft/ArrowRight) matching WAI-ARIA APG radiogroup pattern. Same approach as `TabBar` used on the list screen.

**State preservation:** back to Step 3 preserves `permissionsMode` + `selectedRoleIds` + `selectedAreaIds`.

**Phase 0 cuts:** no area-admin preselection, no "inherited from profile" badge, no user-search within roles. All deferred Ã¢â‚¬â€ no backend support.

### Step 5 Ã¢â‚¬â€ ConfirmaÃƒÂ§ÃƒÂ£o

**Shipped 2026-05-10.** Component: `StepConfirmation.tsx`. Submit is mocked Ã¢â‚¬â€ real API wiring deferred. See `wiki/backlog/novo-template-wizard.md#confirmacao-backend-submit`.

**Review card layout:** two-column (`thumb` | `previewBody`). Left thumb: decorative A4-shaped mini preview with header stripe, code label (`TPL-{PROFILE}-001` or `TPL-GEN-001` mocked Ã¢â‚¬â€ same backlog as Step 2 `next-code-preview`), and 11 alternating placeholder lines (`HIGHLIGHTED_THUMB_LINES = Set([1,4,7,10])`). Right body: `code-chip` + `StatusPill status="draft"` + version pill; template name; meta grid (Perfil, Familia, Origem, Permissoes, Autor).

**Confirmation checkbox:** local `confirmed` state. `WizardFooter primaryDisabled={!confirmed}` Ã¢â‚¬â€ CTA "Criar e abrir editor Ã¢â€ â€™" only enabled when checked. Footer label changes between `'Etapa 5 de 5 Ã‚Â· Confirme para continuar'` and `'Etapa 5 de 5 Ã‚Â· Tudo pronto para criar'`.

**Submit handler:** `handleSubmit` in `StepConfirmation.tsx` calls `onSubmit()` (prop from `TemplateWizardPage`) Ã¢â‚¬â€ currently `navigate('/templates-v2')` with no API call. See `wiki/backlog/novo-template-wizard.md#confirmacao-backend-submit`.

**Meta values derived from wizard state (all client-side):**
- `perfilValue` Ã¢â‚¬â€ `{profileCode}Ã¢â‚¬â€{profile.name}` or `'Generico'`
- `origemValue` Ã¢â‚¬â€ `selectedDocxName` or `'Em branco'`
- `permissoesValue` Ã¢â‚¬â€ `'Toda a empresa'` / `'{n} area(s) selecionada(s)'` / `'{n} perfil(s) selecionado(s)'`
- `autorValue` Ã¢â‚¬â€ `useAuthStore((s) => s.user?.displayName)`

**No `advanceDisabled` prop (YAGNI strip, commit a91ce046).** `StepConfirmation` does not accept `advanceDisabled` Ã¢â‚¬â€ it never needs an external gate. The sole gate is the internal `confirmed` checkbox wired directly to `WizardFooter primaryDisabled`. Keeping the prop would have been dead code from day 1.

---

## Templates List screen

**Route:** `/templates-v2`
**Page:** `frontend/apps/web/src/features/templates/TemplatesListPage.tsx:29`

### Status derivation rule

```ts
status = dto.archived_at
  ? "archived"
  : dto.published_version_id
  ? "published"
  : "draft"
```

Source of truth: `TemplatesListPage.tsx:37Ã¢â‚¬â€œ41`. No separate status field on `TemplateDTO` Ã¢â‚¬â€ derived client-side.

### WorkspaceHeroHeader tone

The list page uses `tone="flat"` on `WorkspaceHeroHeader` Ã¢â‚¬â€ hero typography without card chrome (transparent background, no `border-bottom`, `padding: 0`). Use `tone="banner"` (default) for pages that need the card chrome (Documents Library, future Operations Center).

`tone` prop added at `WorkspaceHeroHeader.tsx:13`. CSS variant `.headerFlat` in `WorkspaceHeroHeader.module.css`.

### TabBar a11y

`TabBar` (`TabBar.tsx:17`) implements WAI-ARIA `tablist` pattern:
- `tabIndex={isActive ? 0 : -1}` (roving tabIndex)
- `ArrowLeft` / `ArrowRight` / `Home` / `End` keyboard navigation
- `aria-selected` on each `role="tab"` button

### Card grid layout

`TemplatesListPage.module.css` Ã¢â‚¬â€ `padding: 24px 28px`, gap `var(--sp-6)`, responsive 3-col grid (3 columns Ã¢â€°Â¥1440px, 2 at Ã¢â€°Â¥1024px, 1 at 375px).

### formatRelative helper

`formatRelative` (inlined at `TemplatesListPage.tsx:17`) formats ISO timestamps to Portuguese relative strings (hoje / ontem / N dias atrÃƒÂ¡s / N meses atrÃƒÂ¡s / N anos atrÃƒÂ¡s). Currently uses `dto.created_at` because `updated_at` is absent from `TemplateDTO` Ã¢â‚¬â€ see backlog. Promote to `lib/utils/` when a second caller appears.

---

## Lifecycle

```
draft Ã¢â€ â€™ in_review Ã¢â€ â€™ approved Ã¢â€ â€™ published
```

- **draft**: editable by author. Can iterate freely.
- **in_review**: locked content. Awaiting approver.
- **approved**: signed off but not yet usable downstream.
- **published**: selectable when creating documents. Immutable.

Re-authoring Ã¢â€ â€™ create a new version (e.g. `v2 draft`); previous published version remains until replaced.

## Domain rules

- One template can be the default for multiple profiles (taxonomy binding, see `modules/taxonomy.md`).
- Only **published** versions show up in the document creation wizard.
- ISO segregation: author of a version cannot be its approver.

## API surface

TBD Ã¢â‚¬â€ extract from `internal/modules/templates_v2/transport/` (or equivalent).

## Template Editor screen (rebuilt 2026-05-10)

**Route:** `/templates-v2/:templateId/versions/:versionNum`
**Page:** `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:1`

Mirrors `DocumentEditorPage` layout (narrower scope Ã¢â‚¬â€ templates have placeholders + outline, no document-instance metadata):

```
[ AppShell Rail 56px ] [ Inner rail 48px ] [ optional 280px panel ] [ canvas + EditorChrome overlay ]
```

### Inner rail icons

| Icon | Action | Notes |
|---|---|---|
| Ã¢â€ Â Voltar | `onBack()` Ã¢â€ â€™ `/templates-v2` | Brand-bg back button (mirrors `DocumentEditorPage`) |
| VariÃƒÂ¡veis | toggles `PlaceholderCatalogPanel` | Default-active (`leftActive='variables'` initial); detected vs available token UI |
| Estrutura | toggles `TemplateOutlinePanel` | Reads `agent.getAgentContext().outline` via `readHeadings`; empty state when no headings |

Only one panel is open at a time (`leftActive: 'variables' | 'outline' | null`). Clicking the active icon closes the panel.

### EditorChrome wiring

- `center` Ã¢â‚¬â€ template name + dot Ã‚Â· "Template" Ã‚Â· `VersionBadge` (`vN`) Ã‚Â· `StatusPill` (mapped from `VersionStatus`)
- `right` Ã¢â‚¬â€ `AutosaveStatus` (pt-BR labels: `''/SalvandoÃ¢â‚¬Â¦/Salvo/Falha ao salvar`) + (when `isDraft`) `Importar .docx` + `Submeter para revisÃƒÂ£o`
- `alert` Ã¢â‚¬â€ success (`role="status"`, `.alertSuccess`) or error (`role="alert"`, `.alertError`); CSS Module classes Ã¢â‚¬â€ no inline styles

### Error UX

`submitForReview` and `autosave.importDocx` errors run through a local `resolveError(err, fallback)` helper that funnels `ApiError Ã¢â€ â€™ resolveErrorMessage(code, message)`, then `Error.message`, then a pt-BR fallback. Same triad documented in `wiki/concepts/error-ux.md`.

### Outline derivation

`readHeadings(editorRef)` walks `agent.getAgentContext().outline` (eigenpal `ParagraphOutline[]`) and emits `Heading[] = { id, level: 1..6, text }`. Headings refresh on editor change (debounced 600ms) and on initial editor-content-ready effect. No new backend endpoint Ã¢â‚¬â€ pure eigenpal-native surface. Future enhancements (click-to-scroll, drag-reorder, persisted panel state) tracked in `wiki/backlog/template-editor.md`.

### Cuts (intentional, see `wiki/backlog/template-editor.md`)

- No right sidebar (templates have no `EditorMetaSidebar`-equivalent metadata).
- No design-source toolbar Ã¢â‚¬â€ eigenpal's built-in toolbar wins (Decision A, mirrors `DocumentEditorPage`).
- No `layout`/`media`/`search` rail icons Ã¢â‚¬â€ never shipped, dropped from legacy `TemplateAuthorPage`.
- No version-history / comments panels Ã¢â‚¬â€ backend gaps.

Eigenpal CSS overrides (wine formatting bar, compact title bar, gradient scrollbar) live in `EditorChrome.module.css`, not in `TemplateEditorPage.module.css`. The page CSS module covers only rail, panel border, alert variants, and canvas layout (~150 lines, tokens-only).

### Implementation artifacts

`frontend/apps/web/design-source/template-editor/artifacts/` Ã¢â‚¬â€ `phase0-audit.md`, `phase1-map.md`, `phase2-preflight.md` (Heavy tier).

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) Ã¢â‚¬â€ Steps 2Ã¢â‚¬â€œ4
- [workflows/template-authoring.md](../workflows/template-authoring.md) (TBD)
- [concepts/placeholders.md](../concepts/placeholders.md)
- [modules/editor-ui-eigenpal.md](editor-ui-eigenpal.md)
- [modules/editor-chrome.md](editor-chrome.md) Ã¢â‚¬â€ toolbar overlay primitive
