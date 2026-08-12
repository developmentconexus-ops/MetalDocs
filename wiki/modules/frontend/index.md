# Frontend modules

> **Last verified:** 2026-08-12
> **Scope:** Per-feature pages for the `frontend/apps/web/src/features/` slices. Mirrors the backend module pages in [`wiki/modules/`](../). Canonical FE structure rules live in [`wiki/architecture/frontend-structure.md`](../../architecture/frontend-structure.md).

## Pages

- [approval.md](approval.md) — inbox, signoff dialog, approval-route admin.
- [auth.md](auth.md) — login, session bootstrap, auth bus.
- [controlled-documents.md](controlled-documents.md) — read-side queries/API for the regulated wrapper (no pages).
- [documents.md](documents.md) — library, editor, published view, new-document wizard.
- [iam.md](iam.md) — admin center, area memberships (`requiresAdmin` gate).
- [templates.md](templates.md) — template list, wizard, eigenpal editor with draft/review/approve.

## How these pair with backend module pages

Each page above links to its backend counterpart under [`wiki/modules/<name>.md`](../). The split is intentional: backend pages own SQL, authorization, governance events; frontend pages own routes, query keys, cache invalidation, and the components that surface them.

## Current workflow routing

For non-trivial changes, start with [`AGENTS.md`](../../../AGENTS.md) and the canonical engineering method at [`docs/engineering/root-cause-global-maximum-method.md`](../../../docs/engineering/root-cause-global-maximum-method.md). Frontend architecture and server-state rules are defined in the wiki and generated API types; the retired `.agents/skills/metaldocs-*` paths are not live workflow entrypoints.

Use `.claude/skills/developing-new-work/SKILL.md` only for new feature/module pre-design and `.claude/skills/adversarial-review/SKILL.md` for adversarial review.

## Related

- [`wiki/architecture/frontend-structure.md`](../../architecture/frontend-structure.md) — canonical folder layout, hard rules.
- [`wiki/modules/editor-ui-eigenpal.md`](../editor-ui-eigenpal.md) — the eigenpal ACL boundary used by documents + templates editors.
- [`wiki/modules/editor-chrome.md`](../editor-chrome.md) — shared editor overlay surface.
- [`wiki/modules/frontend-primitives.md`](../frontend-primitives.md) — `components/ui/` design-system primitives.
- [`wiki/modules/novo-documento-wizard.md`](../novo-documento-wizard.md) — wizard-specific deep dive.
