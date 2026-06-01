# System Map — what to read when

> **Last verified:** 2026-06-01
> **Use this when:** you've finished [`ONBOARDING.md`](../ONBOARDING.md) and need a task-oriented reading path. This page is *not* a domain catalog (that's [`wiki/index.md`](../index.md)) — it's a "what do I read for X" lookup.

## When in doubt, read these three

1. [`wiki/diagrams/c4-container-backend.md`](../diagrams/c4-container-backend.md) — the moving parts.
2. [`wiki/architecture/system-overview.md`](system-overview.md) — ports, services, topology.
3. The relevant `wiki/modules/<name>.md` for the boundary you're touching.

---

## "I'm adding / changing a backend HTTP route"

1. [`architecture/backend-api-structure.md`](backend-api-structure.md) — module layout.
2. [`architecture/api-contract.md`](api-contract.md) + [`api-design-system.md`](api-design-system.md) — OpenAPI is source of truth; `oapi-codegen` generates handlers.
3. [`concepts/authz-tiers.md`](../concepts/authz-tiers.md) — two-tier authz + Postgres tripwire (do not skip).
4. The owning `wiki/modules/<area>.md` route truth table.
5. Skill: [`metaldocs-backend-api`](../../.agents/skills/metaldocs-backend-api/SKILL.md) (route truth-table workflow).

## "I'm building / changing a screen"

1. [`architecture/frontend-structure.md`](frontend-structure.md) — feature-sliced layout, hard rules.
2. [`modules/frontend/index.md`](../modules/frontend/index.md) — pick the feature slice.
3. [`diagrams/`](../diagrams/) — the load-bearing sequence that drives the screen (autosave, signoff, PDF export, create).
4. [`modules/editor-ui-eigenpal.md`](../modules/editor-ui-eigenpal.md) — only if you touch the editor.
5. Skill: [`metaldocs-frontend`](../../.agents/skills/metaldocs-frontend/SKILL.md) + [`metaldocs-tanstack-query`](../../.agents/skills/metaldocs-tanstack-query/SKILL.md).
6. If the screen is sourced from `design-source/<slug>/`: [`metaldocs-screen-implementation`](../../.agents/skills/metaldocs-screen-implementation/SKILL.md).

## "I'm wiring a new query / mutation"

1. [`architecture/frontend-structure.md` §8](frontend-structure.md) — TanStack Query rules.
2. [`frontend/apps/web/src/lib/queryKeys.ts`](../../frontend/apps/web/src/lib/queryKeys.ts) — add the `QK.*` entry there first.
3. The owning [`modules/frontend/<slice>.md`](../modules/frontend/) — confirm invalidation rules.
4. Skill: [`metaldocs-tanstack-query`](../../.agents/skills/metaldocs-tanstack-query/SKILL.md).

## "I'm changing a DB table / writing a migration"

1. [`database/overview.md`](../database/overview.md), [`migration-policy.md`](../database/migration-policy.md), [`reference-data.md`](../database/reference-data.md).
2. [`database/relationships.md`](../database/relationships.md) — relational graph.
3. [`concepts/authz-tiers.md`](../concepts/authz-tiers.md) — tripwires apply to mutating SQL.
4. The owning `wiki/modules/<name>.md` §7 (Deployment / migrations) and §11 (tech-debt).
5. Skill: [`metaldocs-database`](../../.agents/skills/metaldocs-database/SKILL.md).

## "I'm debugging the approval freeze path"

1. [`diagrams/sequence-signoff-freeze.md`](../diagrams/sequence-signoff-freeze.md) — current + planned async design.
2. [`modules/approval.md`](../modules/approval.md) §6 (RecordSignoff) + §8.3 (idempotency).
3. [`modules/approval-tech-debt.md`](../modules/approval-tech-debt.md) — T-004 (PDF outbox), T-005 (inbox drift).
4. [`decisions/0009-pdf-dispatch-outbox.md`](../decisions/0009-pdf-dispatch-outbox.md).
5. [`modules/frontend/approval.md`](../modules/frontend/approval.md) — inbox cache invalidation surface.

## "I'm debugging PDF export"

1. [`diagrams/sequence-pdf-export.md`](../diagrams/sequence-pdf-export.md).
2. [`modules/render-fanout.md`](../modules/render-fanout.md).
3. `internal/platform/servicebus/gotenberg_pdf.go` — Go → Gotenberg direct (docx-renderer is *not* in this path).

## "I'm debugging local startup"

1. [`references/local-dev-startup.md`](../references/local-dev-startup.md).
2. [`references/local-dev-credentials.md`](../references/local-dev-credentials.md).
3. If the script lies: fix the script. Skill: [`runtime-contract-prereq`](../../.agents/skills/runtime-contract-prereq/SKILL.md).

## "I'm onboarding a junior dev"

1. [`README.md`](../../README.md) + [`CLAUDE.md`](../../CLAUDE.md) — culture + operating rules.
2. [`ONBOARDING.md`](../ONBOARDING.md) — day-1 tour.
3. [`diagrams/c4-context.md`](../diagrams/c4-context.md) → [`c4-container-backend.md`](../diagrams/c4-container-backend.md) → the four sequence diagrams in [`diagrams/`](../diagrams/).
4. [`architecture/system-overview.md`](system-overview.md).
5. The role-based deep dives in [`ONBOARDING.md` §3](../ONBOARDING.md).

## "I'm closing out work (review + QA)"

1. [`quality/qa-operating-system.md`](../quality/qa-operating-system.md) — gates + evidence rule.
2. The relevant checklist in [`quality/`](../quality/) (screen / backend-api / workflow-async / release-closeout).
3. [`CLAUDE.md` §4](../../CLAUDE.md) — mandatory gates + hard-stop rule.

---

## See also

- [`wiki/index.md`](../index.md) — domain catalog (different shape from this page).
- [`wiki/README.md`](../README.md) — top-level entry.
- [`wiki/modules/index.md`](../modules/index.md) + [`wiki/modules/frontend/index.md`](../modules/frontend/index.md) — per-module pages.
- [`wiki/decisions/`](../decisions/) — ADRs.
