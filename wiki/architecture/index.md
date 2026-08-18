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
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — operator-ratified disposition baseline for prior decisions.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — exact current T1→T7 router; **T5 active on rendition/viewer subgate**.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point / current gate.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

Current gate:

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore         CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects           ACTIVE / RENDITION-VIEWER SUBGATE
T5-A→T5-P whole adjudication                           PAUSED UNTIL RV-1→RV-6
T6→T7                                                  NOT OPEN
implementation                                         BLOCKED
```

Active T5 staging:

- `../../docs/superpowers/analysis/2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md` — parent corrected T5 candidate.
- `../../docs/superpowers/analysis/2026-08-18-t5-rendition-viewer-strategy-evaluation.md` — active material subgate; RV-1→RV-6 next.

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

## Current T5 rendition/viewer direction

The current recommended split is:

```text
PDF source
  → direct PDF viewer by default

DOCX + SourceOnly
  → native/read-only DOCX viewer
  → no persistent governed PDF merely for viewing

DOCX + RequireOfficialRendition(PDF)
  → conditional durable PDF render from exact Submission
  → immutable OfficialRendition
  → Release gate
```

A preview/viewing PDF and a policy-required `OfficialRendition` are different meanings. A preview/cache may be rebuildable mechanism; only `OfficialRendition` is immutable semantic state and a Release gate.

Renderer product is not frozen; a representative DOCX fidelity corpus must prove the mechanism.

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