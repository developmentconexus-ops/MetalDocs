# MetalDocs Wiki

> **Last verified:** 2026-08-14
> **Purpose:** Durable single source of truth for MetalDocs project knowledge.

## Start here

- **Active product architecture reset:** [architecture/cohesive-platform-redesign.md](architecture/cohesive-platform-redesign.md) — **read first for any product/domain work.** Design-only; no product implementation authorized yet.
- **Fresh-session recovery:** [references/current-agent-handoff.md](references/current-agent-handoff.md).
- **Non-trivial engineering decision:** [standards/root-cause-global-maximum-method.md](standards/root-cause-global-maximum-method.md).
- **Architecture index:** [architecture/index.md](architecture/index.md) — distinguishes active target authority from legacy/current-state references.

## Documentation authority

- `wiki/` is durable maintained truth.
- `docs/` is staging only.
- During the active Cohesive Platform Redesign, the canonical target path is:

```text
AGENTS.md
→ wiki/architecture/cohesive-platform-redesign.md
→ docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md
→ wiki/references/current-agent-handoff.md
```

- Historical `docs/superpowers` planning artifacts were removed from the live tree on 2026-08-14. Git history is their archive.
- Wiki pages marked `LEGACY`, `HISTORICAL`, `SUPERSEDED`, or `CURRENT-STATE REFERENCE` may be used to understand the existing implementation but do not define the new target.

## Canonical domains

- [architecture/index.md](architecture/index.md) — active target design + stable cross-cutting architecture references.
- [standards/index.md](standards/index.md) — engineering/documentation standards.
- [decisions/index.md](decisions/index.md) — ADR history; target-affecting ADRs are subject to the active redesign's retained/superseded classification until final replacement ADRs land.
- [vision/index.md](vision/index.md) — product intent and users; old implementation-specific product wording is being reconciled with the redesign.
- [quality/index.md](quality/index.md) — QA and close-out governance.
- [database/index.md](database/index.md) — current schema/migration truth; target schema is not designed yet.

## Current-state / supporting evidence

- [backend/index.md](backend/index.md) — backend atlas/current-state evidence.
- [modules/index.md](modules/index.md) — current module pages; core pages affected by the redesign are explicitly LEGACY/current-state references.
- [workflows/index.md](workflows/index.md) — current/historical workflow evidence; do not treat old approval/document workflow pages as the new target unless carried into the active ledger.
- [tests/index.md](tests/index.md) — repeatable validation and acceptance evidence.
- [reviews/index.md](reviews/index.md) — review/audit evidence.
- [references/index.md](references/index.md) — operator references and handoffs.
- [backlog/index.md](backlog/index.md) — historical/deferred inventory; forward planning is frozen during the cohesive redesign.
- [implementation/index.md](implementation/index.md) — bounded implementation history, not redesign authorization.

## Rule for the reset

Do not add another competing roadmap, architecture page, milestone tree or design authority. New approved decisions are recorded in the active ledger until the integrated design is complete; only then are final durable decisions promoted into the owning wiki/ADR pages.
