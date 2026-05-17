# Editor Design Notes

> Last updated: 2026-05-17
> Route: `/documents/:documentID/edit`
> Owning feature: `frontend/apps/web/src/features/documents`

## Plan 12 Review Comments Lifecycle

The editor treats comments as active review feedback. Reviewers add comments during review. If the document is rejected, the document returns to draft with comments still visible so the author can revise the content. Approved/released output must remain clean and must not render active editor comments in PDF/released views.

Documents are the baseline for Template Editor parity. Template Editor may only gain comments after template-owned backend/database/API capability exists.

## Audit Status (2026-05-17)

Task 1 is currently blocked by a prerequisite gate failure:

- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents` -> PASS
- `scripts/check-module-contract-sync.ps1 -Module documents` -> FAIL (`shared contract prerequisite`, missing expected wrapper path `frontend/apps/web/src/features/documents/api/documentsV2.ts`)

See `wiki/backlog/editor.md` under `Integration Audit (2026-05-17)` for full evidence and next steps.

Until the gate is repaired and rerun as PASS:

- do not implement new editor behavior
- do not claim lifecycle hardening is complete
- keep template comments deferred and never fake comment UI
