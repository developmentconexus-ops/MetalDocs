# Phase 1 — Map · novo-template-estrutura

**Tier:** Light (no new primitive · ≤120 LOC new CSS · stacked layout · 0 form inputs after cuts; only file picker + 2 toggle cards · no @media).

## 1.1 Reusability — backward (existing primitives reused)

| Element | Primitive | Status |
|---|---|---|
| Page shell | `features/shared/components/wizard/WizardShell` | reused |
| Footer | `features/shared/components/wizard/WizardFooter` | reused (`showBack`, `onBack`) |
| Card containers | `card` global class | reused |
| Kicker labels | `kicker` global | reused |
| `caption`, `mono`, `tiny` | global utilities | reused |
| `btn btn-sm btn-ghost` | global btn primitives | reused |

## 1.2 Reusability — forward (NEW)

`StepStructure` — feature-local at `features/templates/components/wizard/steps/StepStructure.tsx`. Single-screen use. No shared placement — wizard-step-specific.

The 2 starting-point cards are inline `<button>` blocks (mirrors design DOM). NOT extracted to a shared `StartingPointCard` primitive — single-use, would inflate Light tier into Heavy without payoff.

## 1.3 Decomposition

```
TemplateWizardPage
└─ WizardShell
   └─ StepStructure (new)
      ├─ Two starting-point cards (.docx | blank), grid 1×2
      ├─ Selected-docx row (when startingPoint === 'docx' && selectedDocxName)
      └─ WizardFooter
```

## 1.4 Status / enum SSOT

`startingPoint: 'docx' | 'blank'` — only 2 values; inline literal type, no SSOT file needed.

## 1.5 State design

Extend `templateWizard.reducer.ts`:
- `startingPoint: 'docx' | 'blank' | null` (default `null`)
- `selectedDocxName: string | null` (default `null`) — **mocked**, real upload deferred
- `selectedDocxSize: number | null` (default `null`) — bytes, formatted on render

New actions:
- `SET_STARTING_POINT { value }`
- `SET_SELECTED_DOCX { name, size }` — null pair clears
- `CLEAR_SELECTED_DOCX`

When `startingPoint` switches to `'blank'`, clear docx fields automatically (reducer handles).

No persisted state. No debounced inputs.

## 1.6 Backend contract

| Field | Endpoint | Status |
|---|---|---|
| starting-point choice | client-only flag → drives editor handoff post-create | **deferred** → backlog `step3-editor-handoff` |
| docx upload | `POST /api/v2/templates/{id}/{n}/docx-upload-url` | exists — but requires template id (created at Step 5). **Deferred** → backlog `step3-docx-upload`. Step 3 mocks file pick. |
| placeholder extraction | none — `POST /publish` only | **deferred** → backlog `step3-placeholder-extract` |

## 1.7 Step 3 → 2 / 4 advance/back rules

- Advance enabled: `startingPoint !== null` AND (if `'docx'`: `selectedDocxName !== null`).
- Back: `dispatch GO_TO_STEP step:2` (preserves all state).
- Step click on stepper: same as before.

## 1.8 Tier classification rationale

Light triggers met:
- Reuses primitives (`card`, `kicker`, `btn-ghost`, `WizardShell`/`WizardFooter`).
- ~100 LOC new CSS (2 cards + selected-file row).
- No new shared component.
- No `@media` rule — stacked single-column at any width; cards grid `1fr 1fr` collapses cleanly via `min-width: 0`.
- 1 native input (file picker, not in leakage probe scope — hidden + label-clicked).

## Execution plan

Inline (Light tier, single screen, ≤300 LOC total). Same approach as Step 2: main agent does Phase 2+3 inline; Phase 4 verify; Phase 4.5 reviewer.
