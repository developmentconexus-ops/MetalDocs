# Templates backlog

> **Last verified:** 2026-07-06 (F9.4 doc-truth pass: fixed `templates/repository/postgres.go`→`templates/infrastructure/postgres.go` in FE-08 closure narrative, F9.5 rename) | **Prior:** 2026-07-02 (FE-08 closure; see below)
> **Scope:** Deferred items for the Templates feature (List screen + future screens).

## List screen — deferred items

- [x] `TemplateDTO` missing `updated_at` field — **CLOSED 2026-07-02 (FE-08).** Runtime truth check found `templates_template` had NO `updated_at` column at all (not just an unsurfaced field) — added via `db/migrations/0267_templates_template_updated_at.sql` (timestamptz NOT NULL DEFAULT now(), backfilled to created_at; app explicitly `SET updated_at = now()` on every UPDATE in `internal/modules/templates/infrastructure/postgres.go` since no DB-level auto-touch trigger convention exists in this schema). Spec: additive optional `updated_at` on `TemplateDTO` (api/openapi/v1/openapi.yaml). `TemplatesListPage.tsx` now prefers `updated_at` over `created_at` for the relative-time label, falling back when absent.
- [x] Resolve `created_by` user_id → display name. **CLOSED 2026-07-02 (FE-08).** Resolved via the existing iam-owned `UserDisplayNameReader` port (`internal/modules/iam/domain/user_display_name_port.go`, M4/F4.1 — the same port already used by documents/security/distribution/approval), NOT a new templates→iam access path. Spec: additive optional `created_by_display_name` on `TemplateDTO`. Backend batch-resolves in `listTemplates` (one `DisplayNames` call for the page) and single-resolves in `getTemplate`/`CreateTemplate`/`archiveTemplate`. `TemplatesListPage.tsx` (`author`) and `useTemplateArtifact.ts` (`ownerName`, template detail screen) now prefer `created_by_display_name`, falling back to the raw `created_by` id when the port has no resolvable name.
- [x] Card grid gap aligned to tokenized 16px (`var(--sp-4)`) for list cards (pre-existing in code; re-verified in Plan 12.1 on 2026-05-13).
- [x] Mobile tab clipping at 375px fixed via horizontal scroll support on `TabBar` (`overflow-x: auto` + hidden scrollbar) (2026-05-13).
- [ ] `formatRelative` helper inlined in `TemplatesListPage.tsx:17` — promote to `lib/utils/formatRelative.ts` when a second caller appears.

