# Tech Debt Register - novo-documento-wizard

> Companion to `wiki/modules/novo-documento-wizard.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-05-13

## Items

### T-001 · Wizard page still depends on multiple deferred controls with disabled UX branches
- **Severity:** major
- **Surface:** `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/PeopleSubcontrols.tsx:12`
- **Observation:** people/external sharing controls are present but intentionally disabled pending backend support.
- **Evidence:** deferred sections and disabled sub-controls.
- **Linked backlog row:** `R-001`
- **Linked ADR:** missing-ADR

### T-002 · Step contract and reducer invariants are documented, but API contract linkage is partial
- **Severity:** major
- **Surface:** `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:111`
- **Observation:** atomic create path is documented, but full route truth linkage remains split across module docs.
- **Evidence:** dependencies on registry/documents APIs.
- **Linked backlog row:** `R-002`
- **Linked ADR:** missing-ADR

### T-003 · Wizard wrapper abstraction and shared footer extraction lack standalone ADR
- **Severity:** minor
- **Surface:** `frontend/apps/web/src/features/documents/components/wizard/WizardShell.tsx:1`
- **Observation:** wrapper architecture is implicit convention.
- **Evidence:** shared wizard shell plus local adapter layer.
- **Linked backlog row:** `R-003`
- **Linked ADR:** missing-ADR

## Coverage stats

- Public symbols undocumented: n/a (not fully audited)
- Operations missing C4 placement: n/a (frontend flow module)
- Cross-deps missing in section map: n/a (partial module doc)
- State transitions missing: n/a
- Decisions without ADR link: 3
