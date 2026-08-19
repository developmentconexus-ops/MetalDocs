# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — Launch V1 product authority, **REV001**.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — operator-adjudicated Whole-Product GCR A1–A10.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — operator-approved 4+1 semantic ownership.
- **[r10-t1-semantic-state-invariants.md](r10-t1-semantic-state-invariants.md)** — operator-ratified T1 authority.
- **[r10-t2-governance-effectivity-transactions.md](r10-t2-governance-effectivity-transactions.md)** — operator-ratified T2 authority.
- **[r10-t3-authorization-audit-enforcement.md](r10-t3-authorization-audit-enforcement.md)** — operator-ratified T3 authority.
- **[r10-t4-exact-content-storage-integrity-restore.md](r10-t4-exact-content-storage-integrity-restore.md)** — operator-ratified T4 authority.
- **[r10-t5-durable-async-search-external-effects.md](r10-t5-durable-async-search-external-effects.md)** — operator-ratified T5 authority with post-Fable bounded amendments.
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — current operator-ratified prior-decision disposition baseline.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — current stage router; **T6 ACTIVE / DESIGN NEXT**.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — fresh-session recovery point.

## Current gate

```text
Product Contract                                      REV001 / OPERATOR-APPROVED
T1 Semantic State & Invariants                       CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore       CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects         CLOSED / OPERATOR-RATIFIED
Decision Registry                                    CURRENT / RECONCILED / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                  CLOSED / OPERATOR-APPROVED
T6 Canonical API / Frontend Journeys                ACTIVE / DESIGN NEXT
T7 Historical Migration & Cutover                   NOT OPEN
implementation                                       BLOCKED
```

Active T6 staging:

- `../../docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md` — non-authoritative stage bootstrap.

Completed post-T5 Fable staging was removed from the live tree after the operator closed the checkpoint; Git history is the archive.

## T6 current problem surface

T6 translates ratified product semantics into canonical user/API journeys without creating duplicate authority. Its official REOPEN set is:

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
source upload / T4 admission UX
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
exact Search field/ranking UX + prove materialization need
EditorSession/UX lease only if a real editor-integration consumer requires it
```

The post-Fable retitle observation is carried as a T6 proof question: DRAFT retitle mutation must use an existing T2 concurrency law without reopening Revision-owned title semantics.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

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

## Prior redesign / evidence

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md) — historical compatibility/evidence entrypoint; former topology/routing is superseded.
- [launch-v1-scope-rebaseline.md](launch-v1-scope-rebaseline.md) — narrow Records-Governance defer overlay.
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — historical consolidated decision inventory only.

Current implementation/schema/OpenAPI/frontend are evidence only where current authority permits them to answer a concrete T6 question.
