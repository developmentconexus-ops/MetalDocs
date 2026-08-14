# System Map — what to read when

> **Last verified:** 2026-08-14
> **Active program:** Cohesive Platform Redesign

## When in doubt

Read these first:

1. [`../standards/root-cause-global-maximum-method.md`](../standards/root-cause-global-maximum-method.md)
2. **[`cohesive-platform-redesign.md`](cohesive-platform-redesign.md)**
3. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
4. [`../references/current-agent-handoff.md`](../references/current-agent-handoff.md)

If the task changes product/domain semantics, stop at the design gate. No product implementation is authorized yet.

## "I'm working on Organization / AuthZ / Approval / Documents / Templates / Taxonomy / Release"

Use the active redesign stack above. Current module docs are LEGACY/current-state evidence only.

Do not follow the old module topology, old role/capability vocabulary, approval-route model or template lifecycle by inertia.

## "I'm inspecting current runtime/API behavior"

1. [`backend-api-structure.md`](backend-api-structure.md)
2. [`api-contract.md`](api-contract.md) + [`api-design-system.md`](api-design-system.md)
3. current OpenAPI/generated types/code
4. relevant current-state module page from [`../modules/index.md`](../modules/index.md)

Current contract/code answers what runs today, not what must survive the redesign.

## "I'm inspecting the current database"

1. [`../database/index.md`](../database/index.md)
2. current baseline/migrations/relationships
3. [`data-model.md`](data-model.md) only as a LEGACY current-state reference

Do not design migrations until the target domain/data model is closed.

## "I'm inspecting frontend journeys"

1. [`frontend-structure.md`](frontend-structure.md)
2. [`../modules/frontend/index.md`](../modules/frontend/index.md)
3. generated API types
4. relevant QA evidence

Frontend behavior is evidence about user needs; it does not own lifecycle or authorization semantics.

## "I'm investigating the old approval/freeze/PDF defect"

Start from the active redesign's content-truth invariant. Historical sequence diagrams/module docs may be inspected only as evidence.

The known architecture failure was: editor-authored content was reviewed, while freeze rendered a different blank template snapshot. The target must make it impossible for Approval, freeze and Rendition to bind different content identities.

Useful current-support evidence:

- [`../modules/render-fanout.md`](../modules/render-fanout.md)
- current `documents`/`approval`/renderer code
- Git history for historical QA reports if a specific fact must be recovered

## "I'm working on supporting audit/render/search/notifications/distribution/tokens/security"

First read the active redesign's whole-product checklist. Then inspect the relevant supporting module page.

Do not rewrite a supporting component just because the core is changing. Preserve healthy seams and revalidate only the contracts that depend on newly-settled core truths.

## "I'm starting a new module / feature"

If the request falls inside a boundary already owned by the Cohesive Platform Redesign, continue the **single active redesign ledger** instead of creating another design authority.

For genuinely independent future work, use `.claude/skills/developing-new-work/SKILL.md` after confirming it does not overlap the active reset.

## "I'm using the harness / dispatch hub"

Read `.claude/skills/harness-hub/SKILL.md`.

During the redesign gate, the harness is design-only: read-only research/census/review may be parallelized; product implementation may not.

## "I'm reviewing a design"

Use `.claude/skills/adversarial-review/SKILL.md`, verify claims against source, and require root-cause/local-vs-global-max disposition. Do not reopen settled decisions without a material new finding or changed constraint.

## "I'm onboarding"

1. [`../../AGENTS.md`](../../AGENTS.md)
2. [`../../CLAUDE.md`](../../CLAUDE.md) when applicable
3. [`../ONBOARDING.md`](../ONBOARDING.md)
4. active redesign authority + ledger + handoff

## "I'm closing out work"

Current redesign decisions close by being recorded in the active ledger after operator approval. Product implementation close-out/QA only resumes after the implementation gate opens.

## See also

- [`../index.md`](../index.md)
- [`index.md`](index.md)
- [`../modules/index.md`](../modules/index.md)
- [`../decisions/index.md`](../decisions/index.md)
