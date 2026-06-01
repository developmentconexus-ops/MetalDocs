# MetalDocs Wiki

> **Last verified:** 2026-06-01
> **Purpose:** Durable single source of truth for MetalDocs project knowledge.

## Start here

- **New to the codebase?** [ONBOARDING.md](ONBOARDING.md) — day-1 engineer tour, reading order, role-based paths.
- **Need the visual map?** [diagrams/](diagrams/) — C4 Context + Container + the 4 load-bearing sequence diagrams.
- **"What should I read for X?"** [architecture/system-map.md](architecture/system-map.md) — task-oriented reading paths (backend route, screen, migration, freeze debug, onboarding).

## Ownership boundary

- `wiki/` is the durable source of truth for maintained project knowledge.
- `docs/` is the staging workspace for specs, plans, imported research, prompts, and other non-canonical material until promotion.
- Governance and migration rules live in [standards/documentation-governance.md](standards/documentation-governance.md).

## Canonical domains

- [architecture/index.md](architecture/index.md) - system architecture, API contract rules, platform boundaries
- [modules/index.md](modules/index.md) - per-module living docs, tech-debt registers, maturity state
- [database/index.md](database/index.md) - schema ownership, dictionary, migration policy, reference data
- [standards/index.md](standards/index.md) - engineering and documentation standards
- [workflows/index.md](workflows/index.md) - end-to-end product and operator flows
- [tests/index.md](tests/index.md) - repeatable validation and acceptance flows
- [quality/index.md](quality/index.md) - QA operating-system home, release-quality governance, deep-QA placement
- [decisions/index.md](decisions/index.md) - ADRs and durable technical decisions
- [vision/index.md](vision/index.md) - product intent, audience, and longer-term direction

## Supporting memory

- [backlog/index.md](backlog/index.md) - governed deferred work and refactor queues
- [implementation/index.md](implementation/index.md) - bounded implementation trackers and execution history
- [bugs/index.md](bugs/index.md) - historical grouped defect packets
- [reviews/index.md](reviews/index.md) - review packets and audit evidence
- [references/index.md](references/index.md) - operator references, handoffs, external-package notes, and archived support material
- [glossary.md](glossary.md) - stable glossary entrypoint

## How to use this wiki

- Start with the domain index, not with a global file dump.
- Prefer canonical domain docs over references when both exist.
- Keep path stability unless the link impact is bounded and verified.
- Promote from `docs/` into a domain folder only when the material is durable enough to become maintained truth.
