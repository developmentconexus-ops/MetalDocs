# Modules

> **Last verified:** 2026-08-14
> **Status:** Current-runtime evidence index during Cohesive Platform Redesign

The current `internal/modules/*` directories are **not assumed to be the target bounded contexts**. For target architecture, read [../architecture/cohesive-platform-redesign.md](../architecture/cohesive-platform-redesign.md).

## LEGACY / boundary-under-redesign

These module pages describe runtime/history only. Their old detailed living docs and tech-debt narratives have been collapsed because their nouns/boundaries are being replaced rather than incrementally perfected.

- [approval.md](approval.md) — **LEGACY current-state**; target is Approval V1.
- [auth.md](auth.md) — **CURRENT-STATE / boundary under redesign**; V1 AuthN retained, seams simplified.
- [iam.md](iam.md) — **LEGACY boundary**; target concepts are Organization + Authorization.
- [controlled-documents.md](controlled-documents.md) — **LEGACY target concept**; separate context/object is being retired.
- [documents.md](documents.md) — **LEGACY current-state module**; responsibilities are being re-homed into Controlled Information.
- [taxonomy.md](taxonomy.md) — **LEGACY boundary**; Area moves to Organization, Profile→DocumentType direction, remaining classification re-evaluated.
- [templates.md](templates.md) — **LEGACY parallel lifecycle**; template becomes a role/designation of an exact DocumentRevision.
- [jobs.md](jobs.md) — **LEGACY module classification**; periodic jobs are orchestration/composition, not a business bounded context.

Associated `*-tech-debt.md` documents for these boundaries are historical implementation evidence only and must not be treated as target work queues.

## Supporting modules currently retained/re-evaluated

These responsibilities remain real and several already exhibit healthy seams. Do not rewrite them merely because the core is changing; revalidate them against the final domain/event model.

- [audit.md](audit.md), [audit-tech-debt.md](audit-tech-debt.md) — regulatory evidence authority.
- [distribution.md](distribution.md) — released-document distribution/coverage read surface; future read/ack semantics still open.
- [notifications.md](notifications.md) — delivery/inbox consumer; must not own workflow semantics.
- [render-fanout.md](render-fanout.md), [render-fanout-tech-debt.md](render-fanout-tech-debt.md) — rendering/rendition infrastructure; must bind the canonical Revision truth.
- [search.md](search.md), [search-tech-debt.md](search-tech-debt.md) — read-model/projection consumer.
- [security.md](security.md), [security-tech-debt.md](security-tech-debt.md), [security-signals.md](security-signals.md) — security signals/tenant-key concerns; final boundary follows Organization/AuthN/Tenant-lifecycle design.
- [tokens.md](tokens.md), [tokens-tech-debt.md](tokens-tech-debt.md) — tenant dictionary/value provider; snapshot timing/provenance will be revalidated.

## Frontend/current UI evidence

Frontend pages remain useful evidence about user journeys but do not own product semantics:

- [frontend/index.md](frontend/index.md)
- [editor-chrome.md](editor-chrome.md)
- [editor-ui-eigenpal.md](editor-ui-eigenpal.md)
- [frontend-primitives.md](frontend-primitives.md)
- [novo-documento-wizard.md](novo-documento-wizard.md)

## Rule

When a current module page conflicts with the active redesign on a noun, lifecycle, role, permission, workflow or owner, the redesign wins for the target. Use Git history when detailed legacy implementation archaeology is needed.
