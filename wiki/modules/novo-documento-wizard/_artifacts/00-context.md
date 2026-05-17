# novo-documento-wizard - context

Last verified: 2026-05-15 (blank-template create runtime smoke)

- Current doc promoted from standalone feature write-up to governed module page.
- Core feature surface:
  - `NewDocumentWizardPage`
  - `wizard.reducer`
  - step components under `features/documents/components/wizard/steps/`
- Blank-template runtime path:
  - `GET /api/v1/templates/system/blank` supplies the system blank `templateId` + `templateVersionId`.
  - `StepTemplate` keeps the blank card selectable even when profile templates are empty.
  - Submit continues through `POST /api/v1/controlled-documents`, which creates the Registry slot and Documents v2 draft in one transaction.
- Governance companion files:
  - `wiki/modules/novo-documento-wizard-tech-debt.md`
  - `wiki/backlog/novo-documento-wizard-refactor.md`
