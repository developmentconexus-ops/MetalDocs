# Phase 1 Map — novo-template-escopo

> **Status:** GATE PASSED — awaiting user §1.7 checkpoint sign-off
> **Date:** 2026-05-09

---

## 1.1 Reusability scan — backward

| Design element | Existing primitive | Path | Action |
|---|---|---|---|
| 5-step stepper | `Stepper` | `components/ui/Stepper.tsx` | Use as-is |
| Profile selection cards | `SelectableCard` | `components/ui/SelectableCard.tsx` | Use as-is |
| Scope tab switcher (profile / doc) | `TabBar` | `components/ui/TabBar.tsx` | Use as-is (doc tab disabled) |
| Profile code badge | `CodeChip` | `components/ui/CodeChip.tsx` | Use as-is |
| Scrollable layout + page header + stepper wrapper | `WizardShell` | `features/documents/components/wizard/WizardShell.tsx` | **Promote** to `features/shared/components/wizard/` + parameterize `kicker`, `title`, `description`, `steps` props (remove hardcoded doc-wizard text) |
| Back / Next footer | `WizardFooter` | `features/documents/components/wizard/WizardFooter.tsx` | **Promote** to `features/shared/components/wizard/` — already generic, no API change |
| CSS for both above | `WizardShell.module.css` | `features/documents/components/wizard/WizardShell.module.css` | Move with `WizardShell` to shared; rename if needed |
| Profiles data | `useProfilesQuery` | `features/documents/queries/useProfilesQuery.ts` | **Move** to `features/taxonomy/queries/useProfilesQuery.ts` — wraps taxonomy API, misplaced in documents; both features need it |

---

## 1.2 Reusability scan — forward

| Name | Generic? | Used by 2+ features? | Placement | Rationale |
|---|---|---|---|---|
| `WizardShell` (parameterized) | Yes — layout pattern | Yes: documents + templates | `features/shared/components/wizard/WizardShell.tsx` | Promote from documents; remove hardcoded strings |
| `WizardFooter` | Yes — footer pattern | Yes: documents + templates | `features/shared/components/wizard/WizardFooter.tsx` | Promote from documents; already generic |
| `TemplateWizardPage` | No — domain entry point | No | `features/templates/pages/TemplateWizardPage.tsx` | Route entry, orchestrates all 5 steps + wizard reducer |
| `StepScope` | No — template domain | No | `features/templates/components/wizard/steps/StepScope.tsx` | Step 1 UI: tab switcher + profile grid |
| `templateWizardReducer` | No — domain state | No | `features/templates/state/templateWizard.reducer.ts` | 5-step wizard state, URL sync pattern (same as doc wizard) |

---

## 1.3 Component decomposition

```
TemplateWizardPage  (features/templates/pages/TemplateWizardPage.tsx)
  state: useReducer(templateWizardReducer)
  data:  useProfilesQuery()  [moved to features/taxonomy/queries/]
  url:   ?step=N sync via useEffect + setSearchParams

  └── WizardShell [features/shared]
        props: kicker="Templates / Novo"
               title="Novo template reutilizável"
               description="Templates publicados ficam disponíveis…"
               steps={TPL_STEPS}   // ['Perfil','Identidade','Estrutura','Permissões','Confirmação']
               current={state.step}
               onStepClick={onStepClick}

        └── StepScope  (features/templates/components/wizard/steps/StepScope.tsx)
              props: profiles, isLoading, isError, error, onRetry
                     selectedCode, onSelect
                     onAdvance, onCancel, advanceDisabled

              ├── TabBar [ui]
              │     tabs: [
              │       { id: 'profile', label: 'Para um perfil', sub: 'genérico (POP, IT, POL…)' },
              │       { id: 'document', label: 'A partir de um documento',
              │         disabled: true, title: 'Em breve' }
              │     ]
              │     active: 'profile' (always — doc tab disabled)
              │
              ├── [loading] skeleton cards ×2
              ├── [error]   role="alert" + retry button
              ├── [empty]   empty state + taxonomy link
              └── [success] profile grid
                    └── SelectableCard [ui] × profile
                          ├── CodeChip [ui]  (profile.code)
                          ├── profile.name
                          ├── profile.description (caption)
                          ├── "Família: {profile.familyCode}"
                          ├── "— templates publicados"  ← TODO deferred count
                          └── [disabled] "em breve" pill on disabled profiles

        └── WizardFooter [features/shared]
              showBack=false (step 1)
              onCancel → navigate('/templates')
              onAdvance → dispatch goToStep(2)
              primaryDisabled={advanceDisabled}
              stepLabel="Etapa 1 de 5 · Selecione um perfil para continuar"
                      / "Etapa 1 de 5 · Pronto para avançar"
```

---

## 1.4 Status / enum meta SSOT

No new status enums for Step 1. Profile codes are dynamic strings from taxonomy API (`DocumentProfile.code`). No frontend enum needed.

Template wizard step type:
```ts
// features/templates/state/templateWizard.reducer.ts
export type TemplateWizardStep = 1 | 2 | 3 | 4 | 5;
```

---

## 1.5 State design

| Type | Item | Notes |
|---|---|---|
| Server state | `useProfilesQuery()` | Moved to `features/taxonomy/queries/`. `QK.taxonomy.profiles()` already defined. |
| Wizard state | `useReducer(templateWizardReducer)` | `step`, `profileCode`. Steps 2–5 fields added as later steps are implemented. |
| URL sync | `?step=N` | Same `useEffect + setSearchParams` pattern as `NewDocumentWizardPage`. |
| Local (step) | None beyond reducer | All step state in reducer for URL-sync ability. |
| Persisted | None for Step 1 | Draft created on Step 2 confirm (not step 1). |

---

## 1.6 Backend contract

| Endpoint | Path | Status | Notes |
|---|---|---|---|
| List profiles | `GET /taxonomy/profiles` (via `fetchProfiles`) | Existing ✅ | Used by doc wizard already |
| Create template | `POST /templates` `{ profileCode, name }` | Existing ✅ | Called at Step 5 confirm, not Step 1 |
| Template count per profile | Not available | **Needed** (deferred) | Show `—` with TODO block |

Mock fallback: none needed for Step 1 — real profiles API used. Template count shows `—` with TODO comment.

Backlog file: `wiki/backlog/novo-template-wizard.md`

---

## 1.7 Phase 2 work summary

Phase 2 subagent will:
1. Move `useProfilesQuery` → `features/taxonomy/queries/useProfilesQuery.ts`; update `features/documents/` import
2. Promote `WizardShell` + `WizardFooter` + `WizardShell.module.css` → `features/shared/components/wizard/`; parameterize `WizardShell`; update `NewDocumentWizardPage` import
3. Register route `/templates/new` in `features/templates/routes.tsx`
4. Create `features/templates/state/templateWizard.reducer.ts` (step 1 + step fields only)
5. Stub `TemplateWizardPage.tsx` + `StepScope.tsx` (empty renders, no logic)

---

## Open questions resolved

| # | Question | Answer |
|---|---|---|
| 1 | "A partir de um documento" tab | **Cut entirely** — wrong concept. Tab removed. Replaced by `Escopo de aplicação` in Step 2 (deferred). |
