# Editor Design Notes

> Last updated: 2026-05-18
> Route: `/documents/:documentID/edit`
> Owning feature: `frontend/apps/web/src/features/documents`

## Plan 12 Review Comments Lifecycle

The editor treats comments as active review feedback. Reviewers add comments during review. If the document is rejected, the document returns to draft with comments still visible so the author can revise the content. Approved/released output must remain clean and must not render active editor comments in PDF/released views.

Documents are the baseline for Template Editor parity. Template Editor may only gain comments after template-owned backend/database/API capability exists.

## Audit Status (2026-05-17)

Task 1 gates are currently passing:

- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents` -> PASS
- `scripts/check-module-contract-sync.ps1 -Module documents` -> PASS

The full integration classification is recorded in `wiki/backlog/editor.md` under `Integration Audit (2026-05-17)`.

Current implementation boundary:

- proceed with document comments query normalization and permission split
- keep unresolved-comment approval blocking as backend/API prerequisite
- keep template comments deferred until real template comments capability exists
- never fake template comments UI

## Audit Status (2026-05-18)

The refreshed integration classification is recorded in `wiki/backlog/editor.md` under `Integration Audit (2026-05-18)`.

Current editor boundary:

- keep session lease and autosave as intentional editor-local hooks
- treat document detail loading and comments wrappers as legacy wiring to normalize later
- finalize route contract prerequisite closed on 2026-05-18: frontend wrapper now sends `Idempotency-Key`, and OpenAPI/generated types match the runtime `201 { instanceId }` response
- keep sidebar and template-comment items deferred until backend capability exists
