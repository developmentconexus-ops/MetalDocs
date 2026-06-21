# Templates backlog

> **Last verified:** 2026-06-21 (verify-and-archive sweep; see _cleanup-2026-06-21.md)
> **Scope:** Deferred items for the Templates feature (List screen + future screens).

## List screen — deferred items

- [ ] `TemplateDTO` missing `updated_at` field — `formatRelative` currently falls back to `created_at`. Add field in backend and update `TemplatesListPage.tsx:43`.
- [ ] Resolve `created_by` user_id → display name. Currently shows raw UUID passed as `author` to `TemplateCard`. Requires a user lookup (batch or eager join in API response).
- [x] Card grid gap aligned to tokenized 16px (`var(--sp-4)`) for list cards (pre-existing in code; re-verified in Plan 12.1 on 2026-05-13).
- [x] Mobile tab clipping at 375px fixed via horizontal scroll support on `TabBar` (`overflow-x: auto` + hidden scrollbar) (2026-05-13).
- [ ] `formatRelative` helper inlined in `TemplatesListPage.tsx:17` — promote to `lib/utils/formatRelative.ts` when a second caller appears.

