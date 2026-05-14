# Phase 4 Behavior Verification — novo-documento

**Date:** 2026-05-14
**Scope:** Plan 12.3 screen-local verification after integration-audit sync

---

## TypeScript

Command:

```powershell
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Result:

- Exit code `0`
- No TypeScript errors in the current workspace state for this command

---

## Tests

Command:

```powershell
cd frontend/apps/web
pnpm test
```

Result:

- Exit code `1`
- `src/features/documents/pages/NewDocumentWizardPage.test.tsx` passed
- Failures were pre-existing and outside `novo-documento` scope

Observed failing areas from this run:

- `src/features/approval/pages/RouteAdminPage.test.tsx`
- `src/lib/api/__tests__/client.test.ts`
- `src/features/documents/__tests__/DocumentsHubView.edit-button.test.tsx`
- `src/features/documents/hooks/v2/useDocumentPdfStatus.test.ts`
- `src/features/documents/pages/DocumentEditorPage.test.tsx`
- `src/features/documents/__tests__/DocumentEditorPage.test.tsx`
- `src/features/templates/__tests__/template-author-page-convergence.test.tsx`

No new wizard-specific failure was introduced by this PR.

---

## Runtime Smoke

Environment:

- API gates passed before implementation:
  - `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents`
  - `scripts/check-module-contract-sync.ps1 -Module documents`
- Frontend dev server run at `http://127.0.0.1:4176`

Smoke trace:

1. Logged in with the local admin account.
2. Opened `/documents-v2/new`.
3. Step 1 rendered live profile cards.
4. Step 2 rendered live area select and live preview code support.
5. Step 3 rendered live template selection filtered by profile.
6. Step 4 rendered the synced summary with real preview code and template label.
7. Final submit attempted `POST /api/v1/controlled-documents`.

Submit result:

- Response status: `500`
- Response body:

```json
{"title":"internal server error","status":500,"code":"INTERNAL_ERROR"}
```

Interpretation:

- Screen-local UI wiring is verified through Step 4.
- End-to-end document creation remains blocked by a shared runtime/backend issue outside this screen-local PR scope.

---

## Screenshots

Saved under:

`frontend/apps/web/design-source/novo-documento/artifacts/screenshots/`

Captured files:

- `step1.png`
- `step2.png`
- `step3.png`
- `step4.png`
- `step4-error.png`
