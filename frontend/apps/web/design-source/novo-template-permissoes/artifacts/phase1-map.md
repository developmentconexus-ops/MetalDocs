# Phase 1 — Map · novo-template-permissoes

## 1.1 Backward primitive scan

| Design element | Primitive | Status |
|---|---|---|
| Wizard chrome | `WizardShell` | reuse |
| Footer (Voltar / Avançar) | `WizardFooter` | reuse |
| Icons (home, taxonomy, users, check) | `Icon` | reuse |
| Segmented control (pill style) | — | inline (TabBar is underline-style; wrong visual) |
| Role / Area toggle cards | — | inline (same pattern as StepStructure starting-point cards) |
| Coverage summary | — | inline (design-only block) |
| card / kicker / h2 / caption / mono / tiny / btn | globals | reuse |

No new shared primitive needed. Segmented control is a 3-button pill — small enough to inline.

## 1.2 Forward placement

No new component in `components/ui/` or `features/shared/`. All new code goes to:
- `features/templates/components/wizard/steps/StepPermissions.tsx`
- `features/templates/components/wizard/steps/StepPermissions.module.css`

## 1.3 Component tree

```
StepPermissions (props: permissionsMode, selectedRoleIds, selectedAreaIds,
                         onSetMode, onToggleRole, onToggleArea,
                         onAdvance, onBack, advanceDisabled)
  .card
    .kicker                          "Etapa 4 de 5"
    h2.h2                            "Quem pode usar este template?"
    p.caption.intro
    div.modeSegmented [role=radiogroup]
      button × 3  [role=radio, aria-checked]  Por funções / Por área / Todos
    [mode === 'all']
      div.allBanner
        span.allIcon   <Icon name="home" size={20}/>
        div
          div.allTitle
          p.caption
    [mode === 'areas']
      div.areaGrid  (3-col, @media 640px → 2-col, @media 480px → 1-col)
        button.areaCard × 6  [role=checkbox, aria-checked]
          span.cardIcon   <Icon name="taxonomy" size={16}/>
          div.cardBody
            div.cardCode   mono
            div.cardName
            div.tiny       count
          <Icon name="check" size={16}/>  (when checked)
    [mode === 'roles']
      div.roleGrid  (2-col, @media 640px → 1-col)
        button.roleCard × 6  [role=checkbox, aria-checked]
          span.cardIcon   <Icon name="users" size={16}/>
          div.cardBody
            div.cardIdRow  { mono code · tiny area }
            div.cardName
            div.tiny       count
          <Icon name="check" size={16}/>  (when checked)
    [mode !== 'all']
      div.coverageSummary
        div.kicker   "Cobertura estimada"
        div.coverageRow
          span.mono.coverageCount   N
          span.caption   "perfis/áreas selecionadas · cobre ~X usuários"
    WizardFooter
```

## 1.4 Status/enum SSOT

No status meta. `permissionsMode` union type defined in reducer.

## 1.5 State

| Field | Storage | Default |
|---|---|---|
| `permissionsMode` | reducer | `'roles'` |
| `selectedRoleIds` | reducer | `[]` |
| `selectedAreaIds` | reducer | `[]` |

No local `useState` needed — all persisted via reducer for back-navigation.
No TanStack queries — all mock data.

## 1.6 Backend contract

| Item | Status | Strategy |
|---|---|---|
| Personnel roles list | Missing | Hardcode `MOCK_ROLES` + TODO tag |
| Area user counts | Missing | Hardcode `MOCK_AREAS` with counts + TODO tag |
| Company user count | Missing | `COMPANY_USER_COUNT = 340` constant + TODO tag |

Backlog rows: `permissions-roles-api`, `permissions-area-counts`, `permissions-user-count`.

## 1.7 Tier classification

**Heavy.** Triggers:
- `@media` rules: 3-col area grid + 2-col role grid must collapse at mobile.

Heavy workflow: Phase 3a inline (DOM skeleton) → Phase 3b subagent (style + parity) → Phase 3c subagent (state wiring).

Note: no `<input>`/`<select>` rendered → leakage-probe optional in Phase 3b.

## 1.8 Checkpoint

Tile is Heavy. All mock data tagged with TODO + backlog. Phase 3a follows inline.
