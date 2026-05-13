# Novo Documento (Wizard) — Implementation Worksheet

> **Slug:** novo-documento
> **Owning feature:** features/documents
> **Target route:** /documents-v2/new
> **Reference:** ./NOTES.md + ../novo-perfil/novo-perfil.html + ../novo-area-codigo-visibilidade/novo-area-codigo-visibilidade.html + ../novo-template/novo-template.html + ../novo-confirmacao/novo-confirmacao.html
> **Skill version:** 1.0
> **Started:** 2026-05-06
> **Completed:** —

---

## Open Questions Log

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|
| 1 | 0 | Visibility default when no backend field exists | "Toda empresa" pre-selected in UI; not submitted | 2026-05-06 |
| 2 | 0 | Code preview accuracy ("Próxima: 120" vs placeholder) | `≈ POP-QUA-???` with tooltip "Código final atribuído ao confirmar" | 2026-05-06 |
| 3 | 0 | Implement only backed features or render everything? | Render everything; gray + no-op + TODO for unbacked | 2026-05-06 |
| 4 | 0 | Profile "soon" badge handling | Drop (no model) | 2026-05-06 |
| 5 | 0 | Profile per-card count handling | `—` placeholder + TODO | 2026-05-06 |

---

## Phase 0 — Audit (HARD GATE)

### 0.1 Element-by-element audit

See `./NOTES.md` "Audit decision" table — repeats here for completeness.

| Element | Maps to | Keep / Cut / Defer | Reason |
|---|---|---|---|
| Step 1 profile cards | profile model + `registry.create` cap | Keep | Real backend list |
| Profile card "count" badge | aggregate (no endpoint) | Defer (render `—`) | TODO + backlog |
| Profile "soon" disabled state | none | Cut | No model field |
| Step 2 area `<select>` | `process_areas` table | Keep | Real backend list |
| Step 2 title input | `controlled_documents.title` | Keep | Required field |
| Step 2 code preview | server-assigned post-POST | Keep (placeholder) | Visual only |
| Step 2 visibility radios (4 options) | none in `documents` today | Defer (UI rendered, not submitted) | TODO + backlog |
| Step 2 invitee chips | none | Defer | TODO + backlog |
| Step 2 external sub-controls | none | Defer | TODO + backlog |
| Step 3 template cards | `templates` filtered by profile | Keep | Real backend |
| Step 3 prior version radios | only `latest_version` + `published_version_id` exposed | Defer (only published selectable) | TODO + backlog |
| Step 3 "Em branco" option | no no-template path | Defer (grayed disabled) | TODO + backlog |
| Step 4 summary card | derived | Keep | Read-only |
| Step 4 author + createdAt | `useAuthStore` + `Date()` | Keep | Real |
| Step 4 consent checkbox | frontend gate | Keep | UX guard |
| Step 4 create button | two-call POST | Keep | Real flow |
| Stepper header | new generic primitive | Keep | New `components/ui/Stepper.tsx` |
| Cancel / back / forward buttons | URL nav | Keep | Wizard chrome |

### 0.2 Cut list confirmed

- [x] User reviewed cut list (2026-05-07)
- [x] Cuts recorded in NOTES.md (already in §Audit decision) (2026-05-07)

---

## Phase 1 — Map (HARD GATE)

### 1.1 Reusability scan — backward

| Design element | Existing primitive | Path | Action |
|---|---|---|---|
| Profile/visibility/template radio cards | none | — | Missing → new `SelectableCard` (1.2) |
| Step indicator (1/2/3/4 dots + labels) | none | — | Missing → new `Stepper` (1.2) |
| Doc code chip | `CodeChip` | `components/ui/CodeChip.tsx` | Use |
| Status pill (not used in wizard) | `StatusPill` | `components/ui/StatusPill.tsx` | n/a |
| Form field wrapper (label + input + hint) | `FormFieldBox` | `components/ui/FormFieldBox.tsx` | Use for Title input + invitee fields + external sub-controls |
| Area `<select>` | `SelectMenu` | `components/ui/SelectMenu.tsx` | Use |
| User avatar (Step 4 author) | `Avatar` | `components/ui/Avatar.tsx` | Use |
| Inline icons | `Icon` | `components/ui/Icon.tsx` | Use |
| Cancel/Voltar/Avançar buttons | none (using design tokens directly) | — | Use raw `<button>` + token classes (matches `EditorChrome` pattern) |
| Wizard chrome (header + footer + step content slot) | none | — | Missing → new `WizardShell` (1.2, domain-local) |
| Code preview banner (Step 2) | none | — | Missing → new `CodePreviewBanner` (1.2, domain-local) |
| Toolbar / shell elements | `EditorChrome` etc. | `features/shared/components/editor-chrome/` | n/a — wizard has own chrome |

No primitive needs extension. Two new generic primitives + three domain-local components.

### 1.2 Reusability scan — forward

| Name | Generic? | Used by 2+ screens? | Placement | Rationale |
|---|---|---|---|---|
| `Stepper` | Y | Y (template author submission, future wizards) | `components/ui/Stepper.tsx` | Generic, no domain knowledge |
| `SelectableCard` | Y | Y (3× in wizard alone — profiles/visibility/templates; future taxonomy pickers) | `components/ui/SelectableCard.tsx` | Generic radio-card composition |
| `WizardShell` (header + step indicator + footer with prev/next) | N | N (wizard-only today) | `features/documents/components/wizard/WizardShell.tsx` | Domain-specific wizard chrome |
| Per-step components (`StepProfile`, `StepAreaCodeVisibility`, `StepTemplate`, `StepConfirm`) | N | N | `features/documents/components/wizard/steps/` | Domain panels |
| `CodePreviewBanner` | N | N | `features/documents/components/wizard/CodePreviewBanner.tsx` | One-off Step 2 visual |

### 1.3 Component decomposition

```
NewDocumentWizardPage  (features/documents/pages/NewDocumentWizardPage.tsx)
└─ WizardShell                 (features/documents/components/wizard/WizardShell.tsx)
   ├─ Stepper                  (components/ui/Stepper.tsx)
   ├─ <step content>
   │  ├─ StepProfile           (features/documents/components/wizard/steps/StepProfile.tsx)
   │  │  └─ SelectableCard[]   (components/ui/SelectableCard.tsx)
   │  ├─ StepAreaCodeVisibility
   │  │  ├─ <select> Area
   │  │  ├─ <input> Title
   │  │  ├─ CodePreviewBanner
   │  │  └─ SelectableCard[] (visibility) + sub-control panels
   │  ├─ StepTemplate
   │  │  └─ SelectableCard[] (templates + version radios)
   │  └─ StepConfirm
   │     ├─ SummaryCard (inline)
   │     └─ <input type=checkbox>
   └─ Footer (Cancel / Voltar / Avançar | Criar documento)
```

State + handlers held in `NewDocumentWizardPage` via `useReducer`. Step components are pure (props in, callback out).

### 1.4 Status / enum meta SSOT

Wizard introduces no new status enum. Reuses:
- `DocumentStatus` (`features/documents/lib/documentMeta.ts`) — for the eventual link to editor.
- New file `features/documents/lib/visibilityMeta.ts` — even though backend has no model yet, frontend needs the labels + icons SSOT for Step 2 + future EditorMetaSidebar.

| Key | Label (pt-BR) | Icon | Notes |
|---|---|---|---|
| `area` | Apenas minha área | users | Default |
| `people` | Pessoas específicas | user-plus | No-op submit |
| `company` | Toda empresa | building | No-op submit (visual default for QMS docs) |
| `external` | Compartilhamento externo | external-link | No-op submit |

### 1.5 State design

| Type | Item | Notes |
|---|---|---|
| Server state | `useProfilesQuery()` | New hook in `features/documents/queries/` (proxies `fetchProfiles` from `features/taxonomy/api/taxonomy.ts:21`). Stale time 5min. |
| Server state | `useAreasQuery()` | New hook (proxies `fetchAreas`). Stale time 5min. |
| Server state | `useTemplatesByProfileQuery(profileCode)` | New hook (proxies `listTemplates({ doc_type })`). Enabled only when `profileCode` set. |
| Local state | wizard reducer | step + formState (see NOTES.md state map). Held in page. |
| Persisted | none | Wizard is short-lived — no draft persistence in v1 |
| Cross-cutting | `useAuthStore.user` | Read for Step 4 author display |
| Debounced inputs | none | Title is non-fetching |

### 1.6 Backend contract

| Endpoint | Path | Status | Shape | Backlog |
|---|---|---|---|---|
| List profiles | `GET /api/v1/profiles` | Existing | `ProfileDTO[]` | — |
| List areas | `GET /api/v1/process-areas` | Existing | `ProcessAreaDTO[]` | — |
| List templates by profile | `GET /api/v1/templates?doc_type=X` | Existing | `TemplateDTO[]` | — |
| Create slot | `POST /api/v1/controlled-documents` | Existing | `{profileCode, processAreaCode, title, ownerUserId, overrideTemplateVersionId?}` → `ControlledDocumentDTO` | — |
| Create document | `POST /api/v1/documents` | Existing | `{controlled_document_id, template_version_id?, name, form_data}` → `DocumentResponse` | — |
| Visibility model | — | Needed | enum + storage on `controlled_documents` | `wiki/backlog/novo-documento.md#visibility` |
| Share / invitee model | — | Needed | new tables + endpoints | `wiki/backlog/novo-documento.md#sharing` |
| Sequence preview | `GET /api/v1/controlled-documents/next-code?profile=X&area=Y` | Needed (optional polish) | `{nextCode: string}` | `wiki/backlog/novo-documento.md#sequence-preview` |
| Template versions list | `GET /api/v1/templates/:id/versions` | Needed | `[{ID, VersionNum, IsPublished, CreatedAt, Label}]` | `wiki/backlog/novo-documento.md#template-versions` |
| No-template doc create | `POST /api/v1/documents` with `template_version_id:null` | Needed | (existing endpoint relaxed) | `wiki/backlog/novo-documento.md#blank-template` |
| Profile doc count | — | Needed (low priority) | aggregate | `wiki/backlog/novo-documento.md#profile-counts` |
| Slot rollback | — | Needed (UX) | compensating delete on `controlled-documents` if doc create fails | `wiki/backlog/novo-documento.md#slot-rollback` |

Mock fallback strategy:
- Visibility / share — UI rendered, choices captured in form state, **NOT included in submit payload**. TODO comment block above the visibility section + matching backlog row.
- Sequence preview — `≈ POP-QUA-???` placeholder + tooltip. No mock data.
- Template versions — only published version rendered as selectable; older versions rendered as disabled radios with `aria-disabled` + `title="Em breve"`.
- "Em branco" — disabled card with `aria-disabled` + `title="Em breve"`.
- Profile counts — `—` placeholder.
- Slot rollback — show inline error on Step 4 with retry button; orphan slot remains. Document this in backlog.

### 1.7 User review checkpoint

- [ ] Reusability classifications reviewed
- [ ] Backend contract reviewed
- [ ] No open Phase-1 questions

---

## Phase 2 — Pre-flight (advisory)

- [x] OpenAPI codegen run (no backend changes — codegen unchanged)
- [x] `Stepper` primitive committed (`components/ui/Stepper.tsx` + `.module.css` + token additions)
- [x] `SelectableCard` primitive committed (`components/ui/SelectableCard.tsx` + `.module.css`)
- [x] `visibilityMeta.ts` committed (`features/documents/lib/visibilityMeta.ts`)
- [x] TanStack Query hooks committed (`useProfilesQuery`, `useAreasQuery`, `useTemplatesByProfileQuery` in `features/documents/queries/`)
- [x] Route stub registered: `documents-v2/new` → `NewDocumentWizardPage` (replaces existing `DocumentCreatePage`)
- [x] LibrarySidebar broken-link fix `/documents/new` → `/documents-v2/new` (`features/documents/components/LibrarySidebar.tsx:59`)

---

## Phase 3a — Structure mirror (HARD GATE)

Subagent receives all 4 step HTMLs + selected-wizard.jsx.

- [x] DOM tree of `WizardShell` mirrors design header/footer
- [x] Each step component mirrors its respective HTML
- [x] CSS Module class names = direct rename of design HTML class names
- [x] No logic yet — TSX skeletons only
- [ ] Main agent diffed structure vs design HTML — match confirmed

### 3a.1 Class-name mapping (design → CSS Module)

| Source (JSX) | Module class | File |
|---|---|---|
| WizardShell page-scroll wrapper | `WizardShell.scrollWrapper` | `wizard/WizardShell.module.css` |
| WizardShell max-width container | `WizardShell.container` | `wizard/WizardShell.module.css` |
| WizardShell footer row | `WizardShell.footerRow` | `wizard/WizardShell.module.css` |
| Step1 profiles 2-col grid | `StepProfile.profileGrid` | `wizard/steps/StepProfile.module.css` |
| Step1 profile card header row (code+name+check) | `StepProfile.profileHeader` | `wizard/steps/StepProfile.module.css` |
| Step1 profile name | `StepProfile.profileName` | `wizard/steps/StepProfile.module.css` |
| Step1 profile meta row (família · count) | `StepProfile.profileMeta` | `wizard/steps/StepProfile.module.css` |
| Step2 profile-chip + area select 2-col grid | `StepAreaCodeVisibility.profileAreaRow` | `wizard/steps/StepAreaCodeVisibility.module.css` |
| Step2 surface-2 selected-profile chip row | `StepAreaCodeVisibility.profileChip` | `wizard/steps/StepAreaCodeVisibility.module.css` |
| Step2 title row | `StepAreaCodeVisibility.titleRow` | `wizard/steps/StepAreaCodeVisibility.module.css` |
| Step2 visibility 2x2 grid | `StepAreaCodeVisibility.visibilityGrid` | `wizard/steps/StepAreaCodeVisibility.module.css` |
| Step2 visibility card body (icon-tile + text + check) | `StepAreaCodeVisibility.visibilityCardBody` | `wizard/steps/StepAreaCodeVisibility.module.css` |
| Step2 visibility icon tile | `StepAreaCodeVisibility.visibilityIconTile` | `wizard/steps/StepAreaCodeVisibility.module.css` |
| Step2 visibility text col | `StepAreaCodeVisibility.visibilityText` | `wizard/steps/StepAreaCodeVisibility.module.css` |
| Step2 dashed brand-pale callout | `CodePreviewBanner.banner` | `wizard/CodePreviewBanner.module.css` |
| Step2 generated-code mono text | `CodePreviewBanner.code` | `wizard/CodePreviewBanner.module.css` |
| Step3 templates col stack | `StepTemplate.templateStack` | `wizard/steps/StepTemplate.module.css` |
| Step3 template horizontal row (preview+main+check) | `StepTemplate.templateRow` | `wizard/steps/StepTemplate.module.css` |
| Step3 mini doc preview | `StepTemplate.templatePreview` | `wizard/steps/StepTemplate.module.css` |
| Step3 preview title bar | `StepTemplate.previewTitleBar` | `wizard/steps/StepTemplate.module.css` |
| Step3 preview line | `StepTemplate.previewLine` | `wizard/steps/StepTemplate.module.css` |
| Step3 template main col | `StepTemplate.templateMain` | `wizard/steps/StepTemplate.module.css` |
| Step3 template label+pills row | `StepTemplate.templateLabelRow` | `wizard/steps/StepTemplate.module.css` |
| Step3 tiny by/updated meta | `StepTemplate.templateMeta` | `wizard/steps/StepTemplate.module.css` |
| Step4 doc preview-card 2-col | `StepConfirm.previewCard` | `wizard/steps/StepConfirm.module.css` |
| Step4 mini doc thumbnail 120x152 | `StepConfirm.docThumbnail` | `wizard/steps/StepConfirm.module.css` |
| Step4 thumbnail title bar | `StepConfirm.thumbnailTitleBar` | `wizard/steps/StepConfirm.module.css` |
| Step4 thumbnail line | `StepConfirm.thumbnailLine` | `wizard/steps/StepConfirm.module.css` |
| Step4 summary header (chip+status+ver) | `StepConfirm.summaryHeaderRow` | `wizard/steps/StepConfirm.module.css` |
| Step4 doc title | `StepConfirm.docTitle` | `wizard/steps/StepConfirm.module.css` |
| Step4 2-col field grid | `StepConfirm.fieldGrid` | `wizard/steps/StepConfirm.module.css` |
| Step4 field row | `StepConfirm.fieldRow` | `wizard/steps/StepConfirm.module.css` |
| Step4 field label | `StepConfirm.fieldLabel` | `wizard/steps/StepConfirm.module.css` |
| Step4 field value | `StepConfirm.fieldValue` | `wizard/steps/StepConfirm.module.css` |
| Step4 dashed brand-pale "what happens" callout | `StepConfirm.nextStepsCallout` | `wizard/steps/StepConfirm.module.css` |
| Step4 consent checkbox row | `StepConfirm.consentRow` | `wizard/steps/StepConfirm.module.css` |

### 3a.2 Notes / non-trivial substitutions

- The reference JSX renders raw `<button>` for each card; substituted with the `SelectableCard` primitive (radio role, aria-checked). All cards have a single click target (no nested buttons), preserving semantics.
- The reference Step 2 renders a raw `<select>`; substituted with `SelectMenu` primitive (props: `id`, `value`, `options`, `onSelect`).
- Visibility-card icons use `taxonomy / users / home / link` (the JSX reference values). `visibilityMeta.ts` uses different names (`user-plus`, `building`, `external-link`) which are not yet present in `Icon.tsx`. Reconciliation deferred to Phase 3b/3c — TODO comment in `StepAreaCodeVisibility.tsx`.
- `WizardShell` exports both the chrome and a co-located `WizardFooter` so each step renders its own footer (matching the reference where each step calls `<Footer ... />`).
- Step 4 consent checkbox uses `defaultChecked readOnly` for now; Phase 3c wires controlled state.
- Step 1 / Step 2 / Step 3 / Step 4 page headers (kicker / display-1 / body / Stepper) are owned by `WizardShell`, not duplicated per step (reference JSX did the same via `WizardShell` higher-order component).
- "Em branco" template card rendered with `pill cuidado` warning chip; will be `disabled` (aria-disabled + tooltip "Em breve") in Phase 3c per worksheet §1.6.
- Page CSS module (`NewDocumentWizardPage.module.css`) intentionally empty — page is a thin dispatch wrapper around `WizardShell`.

---

## Phase 3b — Style port (HARD GATE)

> **Note (grandfathered):** This screen was implemented prior to
> `metaldocs-screen-implementation` SKILL v1.2, which introduced the
> `parity-diff.md` HARD artifact. Per the v1.2 changelog and
> `frontend-code-reviewer` agent calibration, parity-diff is not
> backfilled here; remaining v1.2 artifacts (token-coverage,
> screenshots, leakage-probe, phase4-behavior) backfilled on
> 2026-05-07. Future screens enforce v1.2 in full from Phase 0.

### 3b.1 Token map

Sources: `design-source/selected-wizard.jsx` (Stepper, WizardShell, Footer, Step1, Step3, Step4) and `design-source/selected-wizard-v2.jsx` (Step2). All values listed are inline `style={{ ... }}` props on the reference JSX.

#### Colors

| Design value | Existing token | New token | Notes |
|---|---|---|---|
| `var(--bg)` (page bg) | `--bg` | — | exact |
| `var(--surface)` (stepper bg, mini-thumb paper) | `--surface` (= `#ffffff`) | — | also covers literal `'white'` for mini-thumbnails |
| `var(--surface-2)` (profile chip, preview card) | `--surface-2` | — | exact |
| `var(--surface-3)` (icon-tile idle) | `--surface-3` | — | exact |
| `var(--brand)` (selected border, kicker accents) | `--brand` | — | exact |
| `var(--brand-soft)` (dashed callout border) | `--brand-soft` | — | exact |
| `var(--brand-pale)` (selected card bg, callouts) | `--brand-pale` | — | exact |
| `var(--border)` (cards, dashed field rows) | `--border` | — | exact |
| `var(--border-strong)` (template preview bars) | `--border-strong` | — | exact |
| `var(--text)` | `--text` | — | exact |
| `var(--text-soft)` (callout body, consent) | `--text-soft` | — | exact |
| `var(--text-muted)` (meta, label) | `--text-muted` | — | exact |
| `'#888'` (Step 4 thumbnail mono code) | `--text-muted` (`#8a7575`) | — | hex→token (visually equivalent) |
| `'#e0d0d0'` (Step 4 thumbnail line bars) | — | `--paper-line` | new token (paper aesthetic) |
| `'white'` (mini-thumbnail bg) | `--surface` | — | `--surface` is `#ffffff` |
| `var(--warning)` (cuidado pill text) | `--warning` | — | already used via `pill` utility |

#### Border-radius

| Design value | Existing token | New token | Notes |
|---|---|---|---|
| `var(--r-2)` (chip, callout, icon-tile) | `--r-2` | — | exact |
| `var(--r-3)` (cards, banner, preview-card) | `--r-3` | — | exact |
| `2` (mini-thumbnails) | — | — | raw `2px` (fixed paper-thumbnail aesthetic per dispatcher allow-list) |
| `6` (visibility icon-tile) | `--r-2` (6px) | — | exact match |

#### Spacing (margin / padding / gap)

| Design value | Existing token | Notes |
|---|---|---|
| `4` | `--sp-1` (4) | exact |
| `8` | `--sp-2` (8) | exact |
| `12` | `--sp-3` (12) | exact |
| `16` | `--sp-4` (16) | exact |
| `20` | `--sp-5` (20) | exact |
| `24` | `--sp-6` (24) | exact |
| `32` | `--sp-7` (32) | exact |
| `3` (visibility label mb, thumbnail mono mb) | `--sp-1` (4) | rounded up — visual drift ≤ 1px |
| `6` (profile/header mb, thumbnail title-bar mb) | `--sp-2` (8) | rounded up — visual drift ≤ 2px |
| `10` (footer gap, card gaps, profile-chip pad-y) | `--sp-2` (8) | rounded down — visual drift ≤ 2px |
| `14` (callout pad, doc-title mb) | `--sp-4` (16) | rounded up — visual drift ≤ 2px |
| `18` (preview-card grid+pad+mb, profileArea mb, ol pad-left) | `--sp-4` (16) | rounded down — visual drift ≤ 2px |
| `22` (banner mb, callout mb) | `--sp-5` (20) | rounded down — visual drift ≤ 2px |
| `28` (page horizontal pad) | `--sp-7` (32) | rounded up — symmetric with vertical 32 |

#### Fixed pixel-perfect (allowed per dispatcher)

| Design value | Use | Notes |
|---|---|---|
| `120 × 152` | Step 4 doc thumbnail | dispatcher allow-list |
| `60 × 76` | Step 3 template preview | dispatcher allow-list |
| `32 × 32` | visibility icon tile | dispatcher allow-list |
| `22 × 22` | stepper dot (in `Stepper` primitive) | dispatcher allow-list |
| `1px` | borders | dispatcher allow-list |
| `0` | margin/padding zero | dispatcher allow-list |
| `2px` | mini-thumbnail border-radius | paper-thumbnail aesthetic |
| `6px 5px` | template preview pad | inside fixed-thumb layout |
| `10px 9px` | doc thumbnail pad | inside fixed-thumb layout |
| `3px` | thumbnail-line content mb (Step 4) | inside fixed-thumb layout |
| `2.5px` | preview-line content mb (Step 3) | inside fixed-thumb layout |
| `1.5px` | preview/thumbnail line bar height | inside fixed-thumb layout |
| `3px / 4px` | preview/thumbnail title-bar height | inside fixed-thumb layout |

#### Typography

| Design value | Existing token | Notes |
|---|---|---|
| `'JetBrains Mono', ...` (mono code) | `--font-mono` | via `.mono` utility |
| `'Inter Tight', ...` (body) | `--font-sans` | via global rule |
| `lineHeight: 1.4 / 1.7` | — | unitless ratios — kept raw per dispatcher rules |
| `letterSpacing: '0.02em'` | — | kept raw per dispatcher rules |
| `fontSize: 7 / 9 / 10 / 11 / 12 / 12.5 / 13 / 14 / 16 / 26 px` | — | **No font-size tokens** in current scale. One-off raw px with TODO comment in CodePreviewBanner per dispatcher rules; used per-element where needed. |

#### Shadow

| Design value | Existing token | New token | Notes |
|---|---|---|---|
| `0 2px 6px rgba(0,0,0,0.06)` (template preview) | — | `--shadow-paper-1` | new (neutral, not brand-tinted) |
| `0 2px 8px rgba(0,0,0,0.06)` (Step 4 thumbnail) | — | `--shadow-paper-2` | new (neutral, not brand-tinted) |
| `--shadow-1` (selected card) | `--shadow-1` | — | already in `SelectableCard` |

### 3b.2 Open questions

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|
| 6 | 3b | **Visibility icon mismatch.** `features/documents/lib/visibilityMeta.ts` declares icons `users` / `user-plus` / `building` / `external-link`; the JSX reference renders `taxonomy` / `users` / `home` / `link`; only the JSX-reference set exists in `components/ui/Icon.tsx`. Phase 3a TSX renders the JSX-reference icons (working). Choose: (A) extend `Icon.tsx` with `user-plus` / `building` / `external-link` and update visibilityMeta to match, OR (B) remap `visibilityMeta.ts` to existing icons. | **B** — remapped `visibilityMeta.ts` to `taxonomy` / `users` / `home` / `link` (matches JSX reference). `IconName` type promoted to `export` from `components/ui/Icon.tsx`; meta now imports it for type safety. | 2026-05-06 |

- [x] All design values mapped
- [x] Missing tokens added in separate commit
- [x] CSS Module uses ONLY tokens
- [ ] User approved screenshot diff vs each step HTML

---

## Phase 3c — State wiring (advisory)

- [x] `useProfilesQuery` / `useAreasQuery` / `useTemplatesByProfileQuery` wired
- [x] Reducer dispatches: `selectProfile`, `setArea`, `setTitle`, `setVisibility`, `addInvitee` / `removeInvitee` / `setExternal`, `selectTemplate`, `setConsent`, `goToStep`, `submitStart` / `submitError` / `submitSuccess`
- [x] URL `?step=N` sync — `useSearchParams` driving reducer
- [x] Pre-fill from URL params (`?profile=POP`) on mount
- [x] Submit flow: `createControlledDocument` → on success `createDocument` → `navigate(/documents-v2/:id)` editor
- [x] Error UX: `ApiError` + `resolveErrorMessage(code, msg)` + `role="alert"` rendering. Slot-already-created surfaces inline retry on doc POST.
- [x] Disabled CTAs: `disabled aria-disabled="true" title="Em breve"` for "Em branco" + templates without published version
- [x] Loading states: skeletons on profiles / areas / templates lists
- [x] Empty states: "Nenhum perfil cadastrado" CTA → `/taxonomy/profiles`; "Nenhuma área" CTA → `/taxonomy/areas`; "Nenhum template publicado" caption
- [x] Semantic HTML: `<button>` only for true buttons; no nested buttons; cards use `<button type="button">` (radio cards via `SelectableCard`)
- [x] TODO comment blocks above visibility submit + template-versions render + Em-branco card + invitee + external + slot-rollback + profile-counts, all referencing `wiki/backlog/novo-documento.md#<anchor>`

> Note (deferred to Phase 4 user review): Older-version radios are not rendered as separate radios — only the published version is selectable per worksheet §1.6 (template-versions endpoint not yet shipped). Templates whose `published_version_id` is `null` are rendered disabled with `title="Em breve"`. The "Em breve" disabled treatment for older versions is satisfied at the per-template level rather than per-version.

---

## Phase 4 — Verify (HARD GATE)

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

- [x] tsc green (2026-05-07)
- [x] vitest green (2026-05-07)
- [ ] Manual smoke
- [ ] Screenshot diff vs each step HTML

Smoke steps:

1. Click "Novo documento" CTA from Library sidebar → lands on `/documents-v2/new?step=1`
2. Step 1: pick profile (e.g. POP) → "Avançar" enables → click → `?step=2`
3. Step 2: pick area, type title, see code preview update, change visibility, fill people invite, → "Avançar" enables only with area+title → click → `?step=3`
4. Step 3: see template list filtered by profile, only published version selectable on each card, "Em branco" grayed, → "Avançar" enables on selection → click → `?step=4`
5. Step 4: summary card shows all selections, check consent → "Criar documento" enables → click → 2-call POST → editor opens
6. Refresh on `?step=2` retains step but resets state (no persistence in v1) — verify graceful behavior (redirect to step=1)
7. Back-button from `?step=3` → `?step=2` retains state via reducer (browser back nav)
8. Pre-fill: navigate to `/documents-v2/new?profile=POP` → wizard opens at step=1 with POP pre-selected (forward-pass)

---

## Phase 5 — Document (advisory)

- [ ] `wiki/modules/documents.md` — bump `Last verified`, add wizard route + components
- [ ] `wiki/modules/registry.md` — add wizard as new caller of `createControlledDocument`
- [ ] `wiki/backlog/novo-documento.md` — created with all deferred items from §1.6
- [ ] `wiki/modules/iam-rbac.md:89` — fix stale visibility migration claim
- [ ] `wiki/implementation/screen-redesign-tracker.md` — bump Wizard status to "Done (UI)"
- [ ] `wiki-curator` agent dispatched
- [ ] PR description references this worksheet path
