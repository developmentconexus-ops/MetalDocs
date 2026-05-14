# Caixa de Aprovação

> **Route:** `/approvals`
> **Owning feature:** `frontend/apps/web/src/features/approval/`
> **Last updated:** 2026-05-14

## Reality-first summary

- `Foco` and `Linha do tempo` remain available in the real screen.
- Runtime inbox source is `GET /api/v1/approval/inbox` only.
- Timeline now represents a submitted-time review stream (not synthetic deadline urgency).
- Heatmap slot is preserved as an honest unavailable-data state.
- `Abrir documento` and timeline review actions navigate to `/registry-v2/{controlled_document_id}`.
- `Aprovar e assinar` and `Devolver` now enter the real signoff flow via active-document context.

## Classification summary

- `implemented and aligned`
  - protected `/approvals` route and toolbar view switcher (`stack`/`timeline`)
  - inbox rendering from real `InboxItem` fields
  - open-document navigation to `/registry-v2/{controlled_document_id}`
  - approve/reject entry through active-document + signoff flow
  - timeline review actions wired to real document navigation
- `implemented but legacy-wired`
  - `features/approval/api/approvalApi.ts` keeps legacy manual wrappers as pre-existing debt
- `missing backend capability`
  - decision-history API for heatmap
  - deadline/urgency metadata for timeline urgency mode
  - richer inbox fields (`kind`, `code`, `summary`, `changes`, `version`)
- `defer`
  - filter panel interaction contract

## Keep out of this PR

- no fake inbox fallback data
- no hardcoded heatmap/urgency behavior
- no backend contract expansion
- no cross-screen refactor outside `caixa-aprovacao`
