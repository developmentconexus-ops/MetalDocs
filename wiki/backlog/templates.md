# Templates backlog

> **Last verified:** 2026-05-08
> **Scope:** Deferred items for the Templates feature (List screen + future screens).

## List screen — deferred items

- [ ] `TemplateDTO` missing `updated_at` field — `formatRelative` currently falls back to `created_at`. Add field in backend and update `TemplatesListPage.tsx:43`.
- [ ] Resolve `created_by` user_id → display name. Currently shows raw UUID passed as `author` to `TemplateCard`. Requires a user lookup (batch or eager join in API response).
- [ ] Card grid gap: design specifies 14px; token `var(--sp-6)` resolves to 24px (no 14px token exists). Minor delta accepted; snap to 16px token if one is introduced.
- [ ] Mobile tab clipping at 375px: "Arquivados" tab may render off-screen on narrow viewports. `TabBar` has no horizontal scroll today — consider `overflow-x: auto` + `scrollbar-width: none` on `.bar`.
- [ ] `formatRelative` helper inlined in `TemplatesListPage.tsx:17` — promote to `lib/utils/formatRelative.ts` when a second caller appears.
