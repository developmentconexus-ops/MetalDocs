# Frontend modules

> **Last verified:** 2026-06-01
> **Scope:** Per-feature pages for the `frontend/apps/web/src/features/` slices. Mirrors the backend module pages in [`wiki/modules/`](../). Canonical FE structure rules live in [`wiki/architecture/frontend-structure.md`](../../architecture/frontend-structure.md).

## Pages

- [approval.md](approval.md) — inbox, signoff dialog, approval-route admin.
- [auth.md](auth.md) — login, session bootstrap, auth bus.
- [controlled-documents.md](controlled-documents.md) — read-side queries/API for the regulated wrapper (no pages).
- [documents.md](documents.md) — library, editor, published view, new-document wizard.
- [iam.md](iam.md) — admin center, area memberships (`requiresAdmin` gate).
- [templates.md](templates.md) — template list, wizard, eigenpal editor with draft/review/approve.

## How these pair with backend module pages

Each page above links to its backend counterpart under [`wiki/modules/<name>.md`](../). The split is intentional: backend pages own SQL, authz, governance events; frontend pages own routes, query keys, cache invalidation, and the components that surface them.

## Related

- [`wiki/architecture/frontend-structure.md`](../../architecture/frontend-structure.md) — canonical folder layout, hard rules.
- [`wiki/modules/editor-ui-eigenpal.md`](../editor-ui-eigenpal.md) — the eigenpal ACL boundary used by documents + templates editors.
- [`wiki/modules/editor-chrome.md`](../editor-chrome.md) — shared editor overlay surface.
- [`wiki/modules/frontend-primitives.md`](../frontend-primitives.md) — `components/ui/` design-system primitives.
- [`wiki/modules/novo-documento-wizard.md`](../novo-documento-wizard.md) — wizard-specific deep dive.
- Skills: [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md), [`metaldocs-tanstack-query`](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md), [`metaldocs-screen-implementation`](../../../.agents/skills/metaldocs-screen-implementation/SKILL.md).
