---
name: metaldocs-screen-implementation
description: "Use when implementing a MetalDocs designed screen from `frontend/apps/web/design-source/` into the feature-sliced frontend. This Codex-discoverable bridge points to the canonical screen workflow in `.claude/skills/metaldocs-screen-implementation/SKILL.md` and must be used before writing TSX or CSS for design-source screens."
---

# MetalDocs Screen Implementation

Read and follow `.claude/skills/metaldocs-screen-implementation/SKILL.md`.

This bridge exists so Codex sessions that discover `.agents/skills` still load the canonical screen implementation workflow.

Always load these together:

1. `.agents/skills/metaldocs-frontend/SKILL.md`
2. `.claude/skills/metaldocs-screen-implementation/SKILL.md`
3. `.agents/skills/metaldocs-tanstack-query/SKILL.md` if the screen wires server state, queries, mutations, generated API types, polling, prefetching, or invalidation

Honor the Iron Law from the canonical screen skill: no phase progression without evidence artifacts, no self-graded visual parity, and the user is the only visual approver.

Stop if the canonical `.claude/skills/metaldocs-screen-implementation/SKILL.md` file is missing.
