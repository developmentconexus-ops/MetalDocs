# ADR 0054 — Cross-tenant claim is sanctioned for background outbox consumers

- **Status:** Accepted
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Sanctions the tenant-unscoped `FOR UPDATE SKIP LOCKED` claim step used by background outbox consumers (`metaldocs.outbox_events` platform consumer and the render staging outboxes `metaldocs.pdf_dispatch_outbox` / `metaldocs.materialize_dispatch_outbox`), and defines the compensating rules that keep the pooled multi-tenant invariant intact. Closes grade-A finding SEC-13.
- **Depends on:** the transactional-outbox invariant (async = outbox, consumers idempotent) and the pooled multi-tenant invariant (every tenant table carries `tenant_id`; tx-local GUCs only).

---

## Context

SEC-13 flagged `StagingOutboxRepository.ClaimPending` (`internal/modules/render/fanout/staging_outbox.go`): the claim query selects `status = 'pending'` rows across **all tenants** with no `tenant_id` predicate, under a self-documented `TODO(render): thread tenant scope through the worker claim path`. Read in isolation, that looks like a pooled-tenancy violation — every request-path invariant in MetalDocs scopes reads by tenant.

Investigation shows this is not an outlier but the platform's own canonical shape: the platform outbox consumer (`internal/platform/messaging/outbox/postgres/consumer.go:36-63`) claims from `metaldocs.outbox_events` with the same tenant-unscoped `FOR UPDATE SKIP LOCKED` pattern. Both consumers run inside trusted binaries (`metaldocs-worker`, and the render fanout loop), never in a request path, and never act on behalf of a caller.

The alternative — per-tenant claim loops — would require enumerating tenants on every poll, multiply queries by tenant count, introduce fairness/starvation machinery (a busy tenant's backlog must not starve others, so a scheduler would be needed), and diverge the staging outboxes from the platform consumer they deliberately mirror. That is redesign cost with no security payoff: the worker holds a single privileged DB connection either way; a tenant predicate in the claim CTE does not reduce its authority.

## Decision

**The claim step of a background outbox consumer MAY select work across all tenants.** The pooled-tenancy invariant is enforced one row later, at processing time, not at claim time. The following compensating rules are binding:

1. **Row-carried tenancy.** Every outbox table carries `tenant_id` per row, and the claim query MUST return it (`ClaimPending` returns `OutboxRow.TenantID`; the platform consumer carries tenant in the event payload/aggregate).
2. **Tenant-scoped processing.** Everything the consumer does *with* a claimed row — business reads/writes, blob access, GUC assumption — MUST be scoped to that row's `tenant_id` (tx-local GUCs per item; tenant-namespaced blob keys). A consumer must never mix rows from different tenants inside one business transaction.
3. **Worker-internal only.** Tenant-unscoped claim queries are permitted only in consumer/janitor code paths of the background binaries. No API-request code path may reuse them. (The staging repo's construction-time table allowlist stays as the injection guard.)
4. **Idempotent consumers.** Unchanged from the outbox invariant; cross-tenant claim does not weaken it.

The `TODO(render)` at `staging_outbox.go:68-70` is resolved by this ADR: it is replaced by a comment citing ADR 0054 instead of promising a tenant predicate that would contradict the platform pattern.

## Consequences

- Reviewers stop flagging tenant-unscoped claims in worker code as tenancy violations, and start checking rules 1–3 instead (the real invariant surface).
- Any future per-tenant fairness need (one tenant flooding an outbox) is a scheduling feature to be designed on top of this claim shape (e.g. per-tenant LIMIT quotas inside the claim CTE), not a reason to fork per-tenant loops.
- The DB tripwire / RLS posture is unaffected: outbox tables are worker-owned queues, not user-facing data surfaces.
