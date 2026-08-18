# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — accepted Launch V1 product authority; `REV000 = initial issuance`, `REV001 = first revision`.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — operator-adjudicated Whole-Product GCR A1–A10.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — operator-approved 4+1 semantic ownership + future-evolution law.
- **[r10-t1-semantic-state-invariants.md](r10-t1-semantic-state-invariants.md)** — operator-ratified T1 authority.
- **[r10-t2-governance-effectivity-transactions.md](r10-t2-governance-effectivity-transactions.md)** — operator-ratified T2 transaction/lifecycle authority.
- **[r10-t3-authorization-audit-enforcement.md](r10-t3-authorization-audit-enforcement.md)** — operator-ratified T3 Authorization/Audit authority.
- **[r10-t4-exact-content-storage-integrity-restore.md](r10-t4-exact-content-storage-integrity-restore.md)** — operator-ratified T4 exact-content/storage/restore authority.
- **[r10-t5-durable-async-search-external-effects.md](r10-t5-durable-async-search-external-effects.md)** — operator-ratified T5 async/Search/external-effects authority.
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — operator-ratified current cross-stage disposition baseline.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — exact current T1→T7 router; **post-T5 Fable checkpoint active / T6 not open**.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point / current gate.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

Current gate:

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore         CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects           CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                    ACTIVE / REVIEW REQUEST STAGED
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN
implementation                                         BLOCKED
```

Active review staging:

- `../../docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md` — **independent cold-review request / review evidence only**.

Completed T5 candidate/subgate/adjudication staging was removed after durable promotion; Git history is the archive.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

Every remaining stage consumes `CURRENT/PRESERVE/REFINED`, designs only its `REOPEN` set, preserves `DEFERRED` seams and rejects `SUPERSEDED` inheritance.

## Active Launch ownership

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Storage/integrity, rendering/viewers, Search, async execution, Historical Migration tooling and backup/restore are mechanisms/projections/cutover/operations, not Launch semantic owners.

## Closed T5 headline

```text
one PostgreSQL-backed durable-job mechanism; River selected/reference mechanism
search_refresh always-required; OfficialRendition render conditional on frozen policy
viewer/preview != OfficialRendition
Search = rebuildable PostgreSQL projection keyed by Document
latest-state refresh converges duplicates/out-of-order jobs
Search lag may omit but never grant stale authority/effectivity
full Search rebuild mandatory
GC periodic reconciliation over GC_PENDING
no mandatory Launch notifications/event bus
no generic ExternalEffectReceipt
bounded-retry terminal-visible/redrivable jobs
```

Renderer/viewer product selection remains deliberately unfrozen pending a representative DOCX fidelity corpus.

## Post-T5 Fable checkpoint

The independent packet asks Fable to cold-review **T1→T5 as one system** before T6 encodes the architecture into public API/frontend journeys.

It attacks cross-stage races, authority uniqueness, Decision Registry drift, overengineering and future seams. Findings are evidence only; T6 remains closed until findings are adjudicated and the checkpoint explicitly closes.

## Prior redesign / evidence

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md) — historical compatibility/evidence entrypoint; former topology/routing is superseded.
- [launch-v1-scope-rebaseline.md](launch-v1-scope-rebaseline.md) — narrow Records-Governance defer overlay.
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — historical consolidated decision inventory only.
- Former B1–B6/R10-C material is usable only as classified by [rebaseline-decision-registry.md](rebaseline-decision-registry.md).

## Stable cross-cutting references

These remain valid only where they do not conflict with Product Contract, GCR, ownership topology, ratified T-stage authorities, Decision Registry or current handoff:

- [backend-api-structure.md](backend-api-structure.md)
- [api-contract.md](api-contract.md)
- [api-design-system.md](api-design-system.md)
- [frontend-structure.md](frontend-structure.md)
- [tenant-context.md](tenant-context.md)
- [trusted-proxy.md](trusted-proxy.md)
- [rate-limiting.md](rate-limiting.md)
- [tech-stack.md](tech-stack.md)
- [deployment.md](deployment.md)

## Legacy/current-state references

These describe prior/current implementation and remain evidence only:

- [backend-target-architecture.md](backend-target-architecture.md)
- [backend-blueprint.md](backend-blueprint.md)
- [system-overview.md](system-overview.md)
- [data-model.md](data-model.md)
- [../backend/index.md](../backend/index.md)

When pages disagree, the Product Contract + GCR + ownership topology + ratified T-stage authorities + Decision Registry + current handoff control the target.