# Phase 2 Pre-flight Audit — novo-documento wizard

**Date:** 2026-05-07
**Branch:** fix/novo-documento-criticals
**Skill version:** 1.0 (grandfathered — see IMPLEMENTATION.md §Artifact Grandfathering)

---

## Primitives Audited

### New generic primitives (created from scratch)

| Primitive | Path | Status | Notes |
|---|---|---|---|
| `Stepper` | `components/ui/Stepper.tsx` + `.module.css` | Created | Step indicator dots + labels; no domain knowledge; reusable across future wizards |
| `SelectableCard` | `components/ui/SelectableCard.tsx` + `.module.css` | Created | Radio-card composition; aria-checked; used 3× in wizard (profiles, visibility, templates) |

### New domain-local components (created from scratch)

| Component | Path | Status | Notes |
|---|---|---|---|
| `WizardShell` | `features/documents/components/wizard/WizardShell.tsx` | Created | Step routing, header/footer chrome, `?step=N` URL sync |
| `StepProfileSelect` (Step 1) | `features/documents/components/wizard/steps/StepProfile.tsx` | Created | Profile card grid; uses `SelectableCard` |
| `StepAreaCodeVisibility` (Step 2) | `features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx` | Created | Area select, title input, code preview banner, visibility grid |
| `StepTemplateSelect` (Step 3) | `features/documents/components/wizard/steps/StepTemplate.tsx` | Created | Template list filtered by profile; published-version gate |
| `StepConfirm` (Step 4) | `features/documents/components/wizard/steps/StepConfirm.tsx` | Created | Summary card, consent checkbox, two-call POST submit |

### Modified existing primitives

None. No existing primitives required extension.

---

## Summary Table

| Item | Result |
|---|---|
| OpenAPI codegen run | No backend changes — codegen unchanged |
| New generic primitives | 2 (`Stepper`, `SelectableCard`) |
| New domain components | 5 (`WizardShell` + 4 steps) |
| Existing primitives modified | 0 |
| Route stub registered | `documents-v2/new` → `NewDocumentWizardPage` |
| LibrarySidebar broken-link fixed | `/documents/new` → `/documents-v2/new` (`LibrarySidebar.tsx:59`) |
| TanStack Query hooks committed | `useProfilesQuery`, `useAreasQuery`, `useTemplatesByProfileQuery` |
| Global CSS leakage found | 2 offenders (see below) |

---

## Global CSS Leakage Map

Full probe details in `leakage-probe.md`. Two offenders found and fixed:

### Offender 1 — `input, select, textarea { width: 100% }` in `base.css`

- **Element hit:** `<input type="checkbox">` on Step 4 consent row
- **Effect:** Checkbox width clobbered to 814px, covering the label; span collapsed to 0 width
- **Resolution:** Scoped rule with `:not([type="checkbox"]):not([type="radio"]):not([type="file"]):not([type="range"]):not([type="color"]):not([type="image"])`
- **Status:** Fixed

### Offender 2 — `label span { text-transform: uppercase }` (legacy rule)

- **Element hit:** Consent label `<span>` on Step 4
- **Effect:** Consent text forced to ALL-CAPS with reduced font-size — misleading UX for legal text
- **Resolution:** Removed `label span` from legacy selector group in prior commit
- **Status:** Fixed (prior commit)

---

## Checklist

- [x] All design-source references listed (`selected-wizard.jsx`, `selected-wizard-v2.jsx`, 4 step HTMLs)
- [x] New primitives committed with CSS modules and token additions
- [x] `visibilityMeta.ts` committed (`features/documents/lib/visibilityMeta.ts`)
- [x] Route stub registered
- [x] LibrarySidebar link corrected
- [x] Global CSS leakage probe complete — 2 offenders found and fixed
- [x] Leakage findings documented in `leakage-probe.md`
