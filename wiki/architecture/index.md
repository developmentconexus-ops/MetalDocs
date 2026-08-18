# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — accepted Launch V1 product authority, now **REV001**; business revision convention remains `REV000 = initial issuance`, `REV001 = first revision`.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — operator-adjudicated Whole-Product GCR A1–A10.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — operator-approved 4+1 semantic ownership + future-evolution law.
- **[r10-t1-semantic-state-invariants.md](r10-t1-semantic-state-invariants.md)** — operator-ratified T1 authority + bounded post-T5 title amendment.
- **[r10-t2-governance-effectivity-transactions.md](r10-t2-governance-effectivity-transactions.md)** — operator-ratified T2 transaction/lifecycle authority + bounded obsolescence-withdraw/late-rendition amendment.
- **[r10-t3-authorization-audit-enforcement.md](r10-t3-authorization-audit-enforcement.md)** — operator-ratified T3 Authorization/Audit authority + bounded provider/obsolescence amendments.
- **[r10-t4-exact-content-storage-integrity-restore.md](r10-t4-exact-content-storage-integrity-restore.md)** — operator-ratified T4 exact-content/storage/restore authority + restore-security/GC-liveness amendments.
- **[r10-t5-durable-async-search-external-effects.md](r10-t5-durable-async-search-external-effects.md)** — operator-ratified T5 async/Search/external-effects authority + canonical-Search/conditional-materialization amendments.
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — operator-ratified current cross-stage disposition baseline, reconciled after Fable Round 1.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — exact current T1→T7 router; **Fable delta review pending / T6 not open**.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point / current gate.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

Current gate:

```text
Product Contract                                       REV001 / OPERATOR-APPROVED
T1 Semantic State & Invariants                        CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions  CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                 CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore        CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects          CLOSED / OPERATOR-RATIFIED
Decision Registry                                     CURRENT / RECONCILED / OPERATOR-RATIFIED
Post-T5 Fable Round-1 amendments                     OPERATOR-RATIFIED / PROMOTED
Fable delta review                                   PENDING
T6 Canonical API / Frontend Journeys                 NOT OPEN
T7 Historical Migration & Cutover                    NOT OPEN
implementation                                        BLOCKED
```

Active review staging:

- `../../docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md` — original cold-review request / evidence.
- `../../docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` — independent review / evidence only.
- `../../docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md` — operator-ratified Round-1 disposition/promotion record.
- `../../docs/superpowers/analysis/2026-08-18-t1-t5-fable-delta-review-request.md` — **active delta-review request**.

Expected delta response:

- `../../docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-delta-review.md`

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

## Current T5/Search headline

```text
one PostgreSQL-backed durable-job mechanism; River selected/reference mechanism
policy-required OfficialRendition render = current conditional durable job
viewer/preview != OfficialRendition
Search journey required
Search baseline = canonical PostgreSQL query/view over current canonical facts
materialized Search + search_refresh + rebuild only on proven derived/expensive/measured consumer
if materialized Search exists, per-Document projection-write serialization spans canonical read→write
Search never grants access/effectivity
GC periodic reconciliation over GC_PENDING with semantic/live-claim/backup recheck
no mandatory Launch notifications/event bus
no generic ExternalEffectReceipt
bounded-retry terminal-visible/redrivable jobs only for activated effects
```

## Current restore headline

```text
all restored ApplicationSessions invalid before ordinary serving
post-snapshot lawful erasure + required known security teardown reconciled/proven before ordinary serving
T7 chooses smallest recovery proof/choreography
no generic per-grant security journal frozen by T4
```

## Post-T5 Fable checkpoint

The original independent review returned `APPROVE T1→T5 WITH MATERIAL FIXES`, with no formal T-stage reopen. The operator ratified the bounded Round-1 amendments and they are now durable authority.

The remaining checkpoint is a **delta review only**. Fable must verify the promoted corrections and return an exact disagreement set; it must not restart the whole T1→T5 critique.

Close condition:

```text
DELTA VERDICT = APPROVE
DISAGREEMENT SET = EMPTY
T6 READINESS = MAY OPEN
```

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