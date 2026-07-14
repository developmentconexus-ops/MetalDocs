# Backlog: Caixa de Aprova��o screen (`/approvals`)

> Last updated: 2026-06-21 (verify-and-archive sweep; see _cleanup-2026-06-21.md)
> Implementation artifacts: `frontend/apps/web/design-source/caixa-aprovacao/artifacts/`

## Integration Audit � 2026-05-14

Preflight evidence:

- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/approvals` � PASS (normalized to runtime `/api/v1/approval/inbox`).
- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module approval` � PASS.
- Runtime sample: `GET /api/v1/approval/inbox?limit=3` returned empty payload shape `{ items: [], total: 0 }`.

## Classification (post-implementation)

- `implemented but legacy-wired`
  - `features/approval/api/approvalApi.ts` remains on legacy manual wrappers (kept as bounded debt)
- `missing backend capability`
  - decision-history endpoint for heatmap widget
  - deadline/urgency fields for urgency timeline mode
  - richer inbox metadata (`kind`, `code`, `summary`, `changes`, `version`)
- `defer`
  - filter panel behavior until audited interaction contract is defined

## Explicit non-implementation rationale

- Heatmap remains placeholder text because no runtime/API surface currently returns actor decision history.
- Deadline urgency timeline was removed because inbox payload has no deadline semantics.
- Rich mock-era decorations were removed because they are not present in runtime contract truth.

## Verification evidence

- `pnpm --filter web test -- src/features/approval/pages/InboxPage.test.tsx` � PASS
- `pnpm --filter web test -- src/features/approval/components/SignoffDialog.test.tsx` � PASS
- `pnpm.cmd --filter web exec tsc --noEmit -p tsconfig.build.json` � PASS
- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/approvals` � PASS
