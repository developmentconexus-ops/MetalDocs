# Novo Template — Etapa 1 · Escopo — Implementation Worksheet

> **Slug:** novo-template-escopo
> **Owning feature:** features/templates
> **Target route:** /templates/new (step 1)
> **Reference:** ./novo-template-escopo.html + ./NOTES.md
> **Skill version:** 1.2
> **Started:** 2026-05-09
> **Completed:** —

---

## Open Questions Log

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|
| 1 | 0 | "A partir de um documento" tab: render disabled ("em breve") or cut entirely? | Cut entirely — no clone-from-doc endpoint; scope step is profile-only picker | ✅ |

---

## Phase 0 — Audit (HARD GATE)

### 0.1 Element-by-element audit

| Element (HTML region) | Maps to | Keep / Cut / Defer | Reason |
|---|---|---|---|
| Rail (56px nav) | AppShell — always present | Keep | Standard shell, not in page scope |
| Toolbar breadcrumb "Templates / Novo" | AppToolbar breadcrumb via route handle | Keep | Real navigation context |
| Page kicker "Templates / Novo" | WizardShell `kicker` prop | Keep | Real nav context |
| Page h1 "Novo template reutilizável" | WizardShell `title` prop | Keep | Real title |
| Page description + `{{campo}}` mono | WizardShell `description` prop | Keep | Real UX copy |
| 5-step Stepper | `Stepper` component, steps from wizard state | Keep | Real wizard nav |
| Card "Etapa 1 de 5" kicker | Step counter rendered by step component | Keep | Real |
| "Escopo do template" h2 | Step card heading | Keep | Real |
| Card intro description | Static copy in step component | Keep | Real |
| Tab "Para um perfil" | scope=`profile` state via `TabBar` | Keep | API exists (`POST /templates` takes `profileCode`) |
| Tab "A partir de um documento" | scope=`document` state | **Defer** | No clone-from-doc endpoint; render tab disabled |
| Profile cards grid 2×3 | `SelectableCard` + `DocumentProfile[]` from taxonomy API | Keep | API exists (`GET /taxonomy/profiles` or similar) |
| Profile code badge (mono "POP") | `CodeChip` or inline `mono` span | Keep | Real |
| Profile name + description in card | `DocumentProfile.name` + `.description` | Keep | Real |
| Profile family label ("Família: Qualidade") | `DocumentProfile.familyCode` | Keep | Real |
| "X templates publicados" count in card | No summary endpoint | **Defer** | Show `—`; same TODO pattern as doc wizard StepProfile |
| "em breve" badge on CHK card | `disabled` + `aria-disabled` + `title="Em breve"` | Keep | Real state (CHK not yet implemented) |
| Document search input (Tab B content) | debounced published docs search | **Defer** | Tied to deferred document-clone tab |
| Document list cards (Tab B) | Published docs from API | **Defer** | Tied to deferred document-clone tab |
| Info banner "O conteúdo será clonado…" | Static info copy | **Defer** | Tied to deferred document-clone tab |
| Footer "Cancelar" (step 1, no back) | `WizardFooter` `showBack=false` + `onCancel` | Keep | Real |
| Footer step label caption | `WizardFooter` `stepLabel` prop | Keep | Real |
| Footer "Avançar →" primary button | `WizardFooter` `primaryDisabled` when nothing selected | Keep | Real |

### 0.2 Cut list confirmed

- [ ] User reviewed cut list (above)
- [ ] Cuts recorded in NOTES.md

---

## Phase 1 — Map (HARD GATE)

### 1.1 Reusability scan — backward

| Design element | Existing primitive | Path | Action |
|---|---|---|---|
| 5-step stepper | `Stepper` | `components/ui/Stepper.tsx` | Use |
| Profile selection cards | `SelectableCard` | `components/ui/SelectableCard.tsx` | Use |
| Scope tab switcher | `TabBar` | `components/ui/TabBar.tsx` | Use |
| Document search (Tab B) | `SearchBar` | `components/ui/SearchBar.tsx` | Use (deferred) |
| Scrollable layout + header | `WizardShell` | `features/documents/components/wizard/WizardShell.tsx` | **Promote + extend** (add props: kicker, title, description, stepCount) |
| Back/Next footer | `WizardFooter` | `features/documents/components/wizard/WizardFooter.tsx` | **Promote** (no API change needed) |
| Code badge in profile card | `CodeChip` | `components/ui/CodeChip.tsx` | Use |

### 1.2 Reusability scan — forward

| Name | Generic? | Used by 2+ screens? | Placement | Rationale |
|---|---|---|---|---|
| `WizardShell` (parameterized) | Yes — layout + header | Yes (doc wizard + template wizard) | `features/shared/components/wizard/` | Promote from documents; remove hardcoded doc text |
| `WizardFooter` | Yes — footer pattern | Yes (doc wizard + template wizard) | `features/shared/components/wizard/` | Promote from documents; API already generic |
| `TemplateWizardPage` (entry) | No — domain specific | No (templates only) | `features/templates/pages/` | Wizard entry page + step router |
| `StepScope` | No — domain specific | No (templates only) | `features/templates/components/wizard/steps/` | Step 1 component |

### 1.3 Component decomposition

```
TemplateWizardPage (features/templates/pages/TemplateWizardPage.tsx)
└── WizardShell [shared] (kicker="Templates / Novo", title="Novo template reutilizável", steps=TPL_STEPS, current=1)
    ├── StepScope (features/templates/components/wizard/steps/StepScope.tsx)
    │   ├── TabBar [ui] (tabs: ["Para um perfil", "A partir de um documento"*disabled])
    │   ├── [scope=profile] Profile grid
    │   │   └── SelectableCard [ui] × N profiles (from taxonomy API)
    │   │       └── CodeChip [ui] + name + desc + family + count(—)
    │   └── [scope=document — disabled/em breve]
    └── WizardFooter [shared] (showBack=false, onCancel, primaryDisabled when !selected)
```

### 1.4 Status / enum meta SSOT

No new status enums for this step. Profile codes are `DocumentProfile.code` strings from taxonomy API (dynamic, not enumerated in frontend).

### 1.5 State design

| Type | Item | Notes |
|---|---|---|
| Server state | `useDocumentProfilesQuery()` | Existing in taxonomy feature or documents feature — check before creating new |
| Local state | `scope: 'profile' \| 'document'` | `useState<'profile' \| 'document'>('profile')` in `TemplateWizardPage` |
| Local state | `selectedProfileCode: string \| null` | `useState<string \| null>(null)` in `TemplateWizardPage` |
| Local state | `currentStep: 1\|2\|3\|4\|5` | `useState<TemplateWizardStep>(1)` in `TemplateWizardPage` |
| Persisted | None for Step 1 | Draft created on Step 2 advance (not before) |

### 1.6 Backend contract

| Endpoint | Path | Status | Shape | Backlog |
|---|---|---|---|---|
| List document profiles | `GET /taxonomy/profiles` or `GET /profiles` | Existing (used by doc wizard) | `DocumentProfile[]` | — |
| Create template | `POST /templates` | Existing | `{ profileCode, name }` → `TemplateDraftDTO` | — |
| Clone from document | `POST /templates/clone-from-document` | **Needed** | `{ documentId }` → `TemplateDraftDTO` | wiki/backlog/novo-template-wizard.md |
| Template count per profile | `GET /templates/summary` | **Needed** | `{ profileCode → count }` | wiki/backlog/novo-template-wizard.md |

Mock fallback: document-clone tab rendered disabled with `aria-disabled title="Em breve"`. No mock data needed for Step 1 since profile list comes from real API.

### 1.7 User review checkpoint

- [ ] Reusability classifications reviewed
- [ ] Backend contract reviewed
- [ ] No open Phase-1 questions

---

## Phase 2 — Pre-flight

- [ ] OpenAPI codegen run (no new endpoints for Step 1)
- [ ] `WizardShell` + `WizardFooter` promoted to `features/shared/components/wizard/` + doc wizard updated to import from new location
- [ ] Route stub `/templates/new` registered in `features/templates/routes.tsx`

---

## Phase 3a — Structure mirror (HARD GATE)

- [ ] DOM tree mirrors design HTML — same tag, same nesting depth, same order
- [ ] CSS Module class names = direct rename of design HTML class names
- [ ] No logic yet — TSX skeleton only
- [ ] Main agent diffed structure vs design HTML — match confirmed

---

## Phase 3b — Style port (HARD GATE)

### 3b.1 Token map

| Design value | Existing token | New token |
|---|---|---|
| (filled by subagent) | | |

- [ ] All design values mapped
- [ ] Missing tokens added to `styles/tokens.css` in separate commit
- [ ] CSS Module uses ONLY tokens
- [ ] User approved screenshot triple-diff

---

## Phase 3c — State wiring

- [ ] `useDocumentProfilesQuery` wired, loading/error/empty/success states
- [ ] `TabBar` controlled — document tab disabled + `title="Em breve"`
- [ ] Profile cards selection wired → `selectedProfileCode` state
- [ ] "em breve" badge on CHK card: `disabled aria-disabled="true" title="Em breve"`
- [ ] Template counts show `—` with TODO comment block
- [ ] `WizardFooter` advance disabled when `selectedProfileCode === null`
- [ ] On advance: navigate to Step 2 (pass selectedProfileCode via wizard state)

---

## Phase 4 — Verify (HARD GATE)

Smoke steps:
1. Navigate to `/templates/new` — Step 1 renders, Stepper shows step 1 active
2. Profile grid loads from API — 6 cards visible (CHK disabled)
3. Select POP card — card highlights, Avançar enabled
4. Click disabled CHK card — no selection change, tooltip "Em breve"
5. "A partir de um documento" tab — click → no state change, tooltip "Em breve"
6. Click Cancelar → navigate back to `/templates`
7. Select profile → Avançar → navigate to Step 2

---

## Phase 5 — Document

- [ ] `wiki/modules/templates.md` updated
- [ ] `wiki/backlog/novo-template-wizard.md` created
- [ ] `wiki-curator` dispatched
