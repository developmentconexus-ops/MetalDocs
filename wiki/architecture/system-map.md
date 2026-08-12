# System Map — what to read when

> **Last verified:** 2026-08-12  
> **Use this when:** you've finished [`ONBOARDING.md`](../ONBOARDING.md) and need a task-oriented reading path.

## When in doubt, read these four

1. [`docs/engineering/root-cause-global-maximum-method.md`](../../docs/engineering/root-cause-global-maximum-method.md) — required method for non-trivial engineering decisions.
2. [`wiki/diagrams/c4-container-backend.md`](../diagrams/c4-container-backend.md) — the moving parts.
3. [`wiki/architecture/system-overview.md`](system-overview.md) — ports, services, topology.
4. The relevant `wiki/modules/<name>.md` for the boundary you're touching.

## "I'm adding / changing a backend HTTP route"

1. [`backend-api-structure.md`](backend-api-structure.md).
2. [`api-contract.md`](api-contract.md) + [`api-design-system.md`](api-design-system.md).
3. [`concepts/authz-tiers.md`](../concepts/authz-tiers.md).
4. The owning `wiki/modules/<area>.md` route truth table.
5. Follow `AGENTS.md` for the root-cause gate, contract-first rule, verification, and QA routing.

## "I'm building / changing a screen"

1. [`frontend-structure.md`](frontend-structure.md).
2. [`modules/frontend/index.md`](../modules/frontend/index.md).
3. The load-bearing sequence diagram in [`diagrams/`](../diagrams/) for the workflow.
4. [`modules/editor-ui-eigenpal.md`](../modules/editor-ui-eigenpal.md) only when touching the editor boundary.
5. Generated frontend API types for any server-state work.
6. Follow `AGENTS.md` and the relevant QA checklist.

## "I'm wiring a new query / mutation"

1. [`frontend-structure.md` §8](frontend-structure.md).
2. [`frontend/apps/web/src/lib/queryKeys.ts`](../../frontend/apps/web/src/lib/queryKeys.ts).
3. The owning [`modules/frontend/<slice>.md`](../modules/frontend/).
4. Generated API types are the contract authority; do not create a parallel handwritten contract.

## "I'm changing a DB table / writing a migration"

1. [`database/index.md`](../database/index.md) and the relevant database pages.
2. [`database/relationships.md`](../database/relationships.md).
3. [`concepts/authz-tiers.md`](../concepts/authz-tiers.md) when authorization/tripwires are involved.
4. The owning module page and migration/ownership guidance.
5. Apply the canonical engineering method before introducing new schema enforcement or tooling.

## "I'm debugging the approval freeze path"

1. [`diagrams/sequence-signoff-freeze.md`](../diagrams/sequence-signoff-freeze.md).
2. [`modules/approval.md`](../modules/approval.md).
3. [`modules/approval-tech-debt.md`](../modules/approval-tech-debt.md).
4. Relevant approval ADRs.
5. [`modules/frontend/approval.md`](../modules/frontend/approval.md) when the frontend surface is involved.

## "I'm debugging PDF export"

1. [`diagrams/sequence-pdf-export.md`](../diagrams/sequence-pdf-export.md).
2. [`modules/render-fanout.md`](../modules/render-fanout.md).
3. `internal/platform/servicebus/gotenberg_pdf.go` for the Go -> Gotenberg path.

## "I'm debugging local startup"

1. [`references/local-dev-startup.md`](../references/local-dev-startup.md).
2. [`references/local-dev-credentials.md`](../references/local-dev-credentials.md).
3. If runtime truth is unreliable, stop feature work, repair the startup/contract prerequisite, and rerun the failed checkpoint before resuming.

## "I'm starting a new module / feature"

1. Read `.claude/skills/developing-new-work/SKILL.md`.
2. Produce the system-impact analysis and canonical Engineering Decision Record.
3. Proceed to design only on Green/Yellow.

## "I'm reviewing a design / plan / diff"

1. Read `.claude/skills/adversarial-review/SKILL.md`.
2. Verify source anchors first.
3. Require root-cause and local/global-maximum disposition for material findings.
4. Stop the review loop when the target property is settled and findings have converged.

## "I'm onboarding a contributor"

1. [`README.md`](../../README.md), [`AGENTS.md`](../../AGENTS.md), and [`CLAUDE.md`](../../CLAUDE.md).
2. [`ONBOARDING.md`](../ONBOARDING.md).
3. [`diagrams/c4-context.md`](../diagrams/c4-context.md) -> [`c4-container-backend.md`](../diagrams/c4-container-backend.md).
4. [`system-overview.md`](system-overview.md).
5. The role-based deep dives in onboarding.

## "I'm closing out work"

1. [`quality/qa-operating-system.md`](../quality/qa-operating-system.md).
2. The relevant checklist in [`quality/`](../quality/).
3. `AGENTS.md` evidence and root-cause close-out rules.

## See also

- [`wiki/index.md`](../index.md)
- [`wiki/README.md`](../README.md)
- [`wiki/modules/index.md`](../modules/index.md)
- [`wiki/modules/frontend/index.md`](../modules/frontend/index.md)
- [`wiki/decisions/`](../decisions/)
