# QA Report — Document Editor (qa/documents-editor)

- Route(s): `/documents/:documentId/edit`
- Page: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Owning module: `documents` (wiki/modules/documents.md)
- QA class: heavy (editor + autosave + ACL + persistence)
- Acting role(s): admin (dev seed `admin / AdminMetalDocs123!`)
- Date: 2026-05-29
- CI: billing-blocked → local-evidence gate per `wiki/quality/qa-operating-system.md`

## Gate results

- **Gate 0 (scope/startup/auth/route truth):** pass. API up via `.\scripts\start-api.ps1 -Build` on :8081, frontend on :4173, login OK, target route reachable on a real draft document.
- **Gate 1 (impl truth):** mixed.
  - `corepack pnpm tsc --noEmit -p tsconfig.build.json` → green.
  - `corepack pnpm exec vitest run src/features/documents/pages/DocumentEditorPage.test.tsx` → red on import resolution (`@metaldocs/editor-ui`, `@metaldocs/shared-tokens`). Classified as workflow-tooling gap, not screen-local. See findings.
  - `go test ./internal/modules/documents/application/... ./internal/modules/documents/delivery/...` → green.
  - `go test ./internal/modules/documents/repository/...` → 2 pre-existing failures unrelated to this fix (see findings).
- **Gate 3 (product QA — Preview-driven):** pass after fix.
  - Initial load honest, eigenpal 7-token literal preserved (`{doc_code}`, `{doc_title}`, `{revision_number}` typed and stayed literal in ProseMirror).
  - Edit + save: BEFORE fix — autosave commit returned HTTP 500, UI stuck on `Erro ao salvar`. AFTER fix — `Salvando → Salvo`, new persisted revision.
  - Optimistic→persisted settle correct after fix.
  - Loading/disabled/saving states behaved.
  - Refresh / return-nav: reopened document shows persisted content (revision 4).
- **Gate 5 (regression):** targeted application + delivery layer green; repository layer has unrelated pre-existing scan-mismatch failures already present on `main` (filed as defer below).

## Findings (severity-ordered)

| # | Severity | Family | Finding | Disposition |
|---|---|---|---|---|
| 1 | critical | module-local implementation | `Repository.GetPendingForCommit` SELECT used unqualified `content_hash, storage_key, expires_at, consumed_at, session_id, document_id, base_revision_id` in a JOIN against `documents` (which has its own `content_hash bytea` governance column). Postgres returned `column reference "content_hash" is ambiguous`; every autosave commit 500-ed. Real SQL never hit in unit tests because they use in-memory `fakeRepo`. | **FIXED** — qualified columns with `p.` prefix in `internal/modules/documents/repository/repository.go` `GetPendingForCommit`. Verified end-to-end via Preview + DB inspection. |
| 2 | high | module-local implementation (pre-existing) | `TestListDocumentsPaginated_StatusFilter` and `TestListDocumentsPaginated_PageOffset` fail with `sql: expected 14 destination arguments in Scan, not 18`. Column-list/scan-target drift in `ListDocumentsPaginated`. | **DEFER** — separate defect, present on `main`, not in scope of qa/documents-editor. Track under `wiki/modules/documents-tech-debt.md`. |
| 3 | medium | workflow/tooling gap | `frontend/apps/web/vitest.config.ts` missing the workspace aliases that `vite.config.ts` defines (`@metaldocs/editor-ui`, `@metaldocs/shared-tokens`, `@eigenpal/*`). DocumentEditorPage vitest fails import resolution at collect time. | **DEFER** — tooling gap, file under FE tech debt; does not affect runtime. Add aliases to `vitest.config.ts` in a separate change. |
| 4 | low | wiki-memory drift | `wiki/modules/documents.md` `Last verified` stale. | **FIXED** — bumped to 2026-05-29 with bug + fix evidence. |

## Evidence

### Commands run
```
.\scripts\start-api.ps1 -Build                              # API :8081
corepack pnpm tsc --noEmit -p tsconfig.build.json           # green
corepack pnpm exec vitest run src/features/documents/pages/DocumentEditorPage.test.tsx  # red (tooling gap #3)
go test ./internal/modules/documents/application/... ./internal/modules/documents/delivery/...  # green
go test ./internal/modules/documents/repository/...        # 2 pre-existing failures (finding #2)
```

### Runtime proof (Preview)
- `preview_snapshot` initial load: route renders, ep-root mounted, document title + content visible.
- `preview_fill` ProseMirror: typed `{doc_code} {doc_title} {revision_number}` — tokens stayed literal in DOM (no client-side substitution). Eigenpal ACL OK.
- Autosave status indicator transitioned `Salvando → Salvo` after fix (was stuck on `Erro ao salvar` before).
- `preview_console_logs` + `preview_network`: no client errors after fix; `POST /api/v1/documents/:id/autosave/commit` returned 200.

### Persisted/API proof
Direct DB inspection on `metaldocs-postgres`:
- `document_revisions`: new row, `revision_num=4`, `content_hash=685d0823…`, `file_size_bytes=1332`.
- `autosave_pending_uploads`: latest row `consumed_at` populated (was NULL with 500). Two earlier dangling rows from the broken-commit window remain with NULL `consumed_at`; `idx_pending_expired` sweep reaps them.
- pg log line that exposed the bug: `column reference "content_hash" is ambiguous`.

### Diff summary
- `internal/modules/documents/repository/repository.go` — qualified `p.session_id, p.document_id, p.base_revision_id, p.content_hash, p.storage_key, p.expires_at, p.consumed_at` in `GetPendingForCommit`. Surgical, no other changes.
- `wiki/modules/documents.md` — `Last verified` bumped with qa/documents-editor evidence.
- `frontend/apps/web/QA-REPORT-documents-editor.md` — this report.

## Hard-stops / Bounded defers

- **No hard-stops triggered.** Root cause was module-local (one SELECT) and fix was local at the owning boundary.
- **Defer #2** — `ListDocumentsPaginated` 14-vs-18 scan-arg mismatch: pre-existing, unrelated to editor flow. File under `wiki/modules/documents-tech-debt.md`. Minimum-fix: realign the column list with the Scan target struct in `ListDocumentsPaginated` and add a real-DB integration test so column drift can't ship again.
- **Defer #3** — vitest workspace alias gap: add `@metaldocs/editor-ui`, `@metaldocs/shared-tokens`, `@eigenpal/*` aliases to `frontend/apps/web/vitest.config.ts` to mirror `vite.config.ts`. File under FE tooling debt.

## Self-check

- [x] Screen driven through Preview as user, not inferred from code.
- [x] All applicable checklist paths exercised for admin role.
- [x] Every finding classified (family + severity) with disposition.
- [x] Root-cause fix at owning boundary (repository SELECT), no symptom patching at handler or UI.
- [x] No redesign-grade work performed; no hard-stop boundary violated.
- [x] Evidence recorded (commands + Preview runtime + persisted DB proof).
- [x] `wiki/modules/documents.md` `Last verified` bumped.
