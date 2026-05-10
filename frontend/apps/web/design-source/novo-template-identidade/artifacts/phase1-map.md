# Phase 1 — Map · novo-template-identidade

**Tier:** Light (no new primitive · ≤100 LOC new CSS · stacked recap row drops media query · only 2 form inputs after tag-cut).

## 1.1 Reusability — backward (existing primitives reused)

| Element | Primitive | Status |
|---|---|---|
| Page shell | `features/shared/components/wizard/WizardShell` | reused |
| Footer | `features/shared/components/wizard/WizardFooter` | reused (`showBack=true`, `onBack`) |
| Profile chip | `code-chip` global class | reused |
| Name/description fields | bare `<input className="input">`, `<textarea className="input">` | reused (matches design exactly) |
| Card containers | `card` global class | reused |
| Kicker labels | `kicker` global class | reused |

## 1.2 Reusability — forward (NEW)

`StepIdentity` — feature-local at `features/templates/components/wizard/steps/StepIdentity.tsx`. Single-screen use; no shared placement.

## 1.3 Decomposition

```
TemplateWizardPage
└─ WizardShell
   └─ StepIdentity (new)
      ├─ ScopeRecap row (profile card + version card, stacked)
      ├─ Name input
      ├─ Description textarea
      ├─ Code preview card (mocked next-code)
      └─ WizardFooter
```

## 1.4 Status/enum SSOT — N/A (no enums introduced).

## 1.5 State design

Extend `templateWizard.reducer.ts` with:
- `name: string` (default `''`)
- `description: string` (default `''`)

New actions:
- `SET_NAME` { value }
- `SET_DESCRIPTION` { value }

No persisted state. No debounced inputs (form fields, not search).

## 1.6 Backend contract

| Field | Endpoint/source | Status |
|---|---|---|
| name | `POST /api/v2/templates` body `name` | exists |
| description | `POST /api/v2/templates` body `description` | exists |
| version v1.0 | static literal | N/A |
| code preview | mocked `TPL-{PROFILE}-XXX` (`GEN` for generic) | **deferred** → backlog row `next-code-preview` |
| key | auto-slug at submit (Step 5) | **deferred** → backlog row `key-generation` |

## 1.7 Step 2 → 1 advance/back rules

- Advance enabled: `name.trim().length >= 3`.
- Back: `dispatch GO_TO_STEP step:1` (preserves scope/profile state).
- "Trocar" button (inside profile recap): same as Back.

## 1.8 Tier classification rationale

Light triggers: reuses primitives, ~80 LOC new CSS, no new shared component, no media query (stacked recap), 2 form inputs (below 3-input Heavy threshold).

## Execution plan

Phase 2 inline (reducer extension only — no codegen, no new atoms, no route changes). Phase 3 inline (Light tier, single screen, ≤300 LOC total — subagent boot overhead exceeds savings per skill v1.3 intent). Phase 4 verify. Phase 4.5 reviewer.
