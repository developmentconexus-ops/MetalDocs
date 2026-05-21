---
name: metaldocs-frontend
description: "Use for ANY MetalDocs frontend work under `frontend/apps/web/`: screens, components, routes, state, API wiring, feature slices, CSS Modules, generated API types, design implementation, frontend refactors, or UI review. This is the Codex-discoverable bridge to the canonical frontend workflow in `.claude/skills/metaldocs-frontend/SKILL.md`."
---

# MetalDocs Frontend

Read and follow `.claude/skills/metaldocs-frontend/SKILL.md`.

This bridge exists so Codex sessions that discover `.agents/skills` still load the canonical frontend workflow.

Also load:

- `.agents/skills/metaldocs-tanstack-query/SKILL.md` when API calls, query hooks, query keys, cache invalidation, optimistic updates, generated frontend API types, polling, prefetching, or freshness policy are involved.
- `.agents/skills/metaldocs-screen-implementation/SKILL.md` when implementing a designed screen from `frontend/apps/web/design-source/<slug>/`.

Canonical frontend rule to keep in mind while working:

- Server-driven workflow transitions that can change without a local click, such as scheduler cutover or approval progression, must keep freshness policy in the TanStack Query layer with targeted invalidation and selective `refetchInterval`. Do not add page-level timer loops in components for server-state synchronization.

Required sources:

1. `wiki/architecture/frontend-structure.md`
2. `.claude/skills/metaldocs-frontend/SKILL.md`
3. `wiki/concepts/error-ux.md` when API or error UI is touched
4. Affected module or feature wiki docs

Stop if the canonical `.claude/skills/metaldocs-frontend/SKILL.md` file is missing or conflicts with `wiki/architecture/frontend-structure.md`.
