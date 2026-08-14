# Architecture

> **Last verified:** 2026-08-14
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[cohesive-platform-redesign.md](cohesive-platform-redesign.md)** — **ACTIVE architecture authority.** MetalDocs is in a design-only whole-platform reset covering Organization/AuthZ, Approval, Controlled Information and every supporting concern. No product implementation until its integrated design gate opens.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding Root-Cause / Global-Maximum method.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point.

## Stable cross-cutting references

These remain valid where they do not conflict with the active domain redesign:

- [backend-api-structure.md](backend-api-structure.md) — backend route/OpenAPI structure.
- [api-contract.md](api-contract.md) — contract-first policy and generated-surface rules.
- [api-design-system.md](api-design-system.md) — API behavior standards.
- [frontend-structure.md](frontend-structure.md) — frontend structure and boundary rules.
- [tenant-context.md](tenant-context.md) — runtime tenant trust boundary.
- [trusted-proxy.md](trusted-proxy.md) — network trust boundary.
- [rate-limiting.md](rate-limiting.md) — request throttling architecture.
- [tech-stack.md](tech-stack.md), [deployment.md](deployment.md) — supporting runtime/operations references.

## Legacy/current-state architecture references

The following describe prior target programs or the current implementation and **must not be used as the target domain authority while the cohesive redesign is active**:

- [backend-target-architecture.md](backend-target-architecture.md) — prior normative backend target; current infrastructure requirements may still be useful, but its module/domain topology is under redesign.
- [backend-blueprint.md](backend-blueprint.md) — current/prior-state composition evidence.
- [system-overview.md](system-overview.md) — current-state overview.
- [data-model.md](data-model.md) — current/legacy data-model reference; target data model will be written only after domain closure.
- [../backend/index.md](../backend/index.md) — forensic/current-state backend atlas; evidence, not target design.

When these pages disagree with the active cohesive redesign on domain nouns, ownership or module boundaries, the cohesive redesign wins for the target.
