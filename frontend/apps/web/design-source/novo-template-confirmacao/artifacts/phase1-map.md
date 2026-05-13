# Phase 1 — Map · novo-template-confirmacao

## 1.1 Backward primitive scan

| Element | Primitive | Status |
|---|---|---|
| Wizard chrome | `WizardShell` | reuse |
| Footer | `WizardFooter` | reuse |
| Draft status | `StatusPill` (status='draft') | reuse |
| Auth user name | `useAuthStore(s => s.user)` | reuse |
| code-chip, kicker, h2, caption, mono, tiny, pill | globals | reuse |

No new primitives needed.

## 1.2 Forward placement

`features/templates/components/wizard/steps/StepConfirmation.tsx` + `.module.css`

## 1.3 Component tree

```
StepConfirmation (props: name, description, scopeType, selectedProfile,
                          profileCode, startingPoint, selectedDocxName,
                          permissionsMode, selectedRoleIds, selectedAreaIds,
                          onBack, onSubmit)
  .card
    .kicker              "Etapa 5 de 5"
    h2.h2.stepTitle
    p.caption.intro
    div.previewCard      grid(120px 1fr) + gap + padding
      div.thumb          120×152 paper mock (decorative)
        div.thumbHeader
        div.thumbCode    mono tiny
        div.thumbLine × 11
      div.previewBody
        div.headerRow    { span.code-chip + StatusPill + span.pill.mono "v1.0" }
        div.templateName
        div.metaGrid     grid(1fr 1fr)
          div.metaRow × 5  { span.label + span.value }
    div.confirmBlock     brand-pale dashed
      div.kicker.confirmKicker  "Ao confirmar"
      ol.confirmList
        li × 4
    label.checkLabel
      input[type=checkbox]
      span               confirmation text
    WizardFooter (primaryLabel="Criar e abrir editor →", primaryDisabled=!confirmed)
```

## 1.4 Status SSOT

No new status meta. `StatusPill` already handles 'draft'.

## 1.5 State

| Field | Storage | Notes |
|---|---|---|
| `confirmed` | `useState(false)` | local — gates CTA; not in reducer |
| `user.displayName` | `useAuthStore` | already in store from login |
| All wizard data | props from TemplateWizardPage | name, profileCode, etc. |

No TanStack queries — all data from wizard state + auth store.

## 1.6 Backend

| Item | Status | Strategy |
|---|---|---|
| `POST /api/v1/templates` | Exists but wiring deferred | Mock: navigate('/templates') + TODO tag |
| Permissions API | Missing | N/A (visual only) |
| Structure/docx upload | Missing | N/A (visual only) |

Backlog: `confirmacao-backend-submit`.

## 1.7 Tier

**Light.** No @media, ≤100 lines CSS, no new shared primitives. 1 `<input type="checkbox">` + `<label>` → leakage-probe required in Phase 3.

## 1.8 Checkpoint

Visual only. Submit mocked. Checkbox gates CTA.

Props needed from TemplateWizardPage: pass `state` fields individually to StepConfirmation + `selectedProfile`.
