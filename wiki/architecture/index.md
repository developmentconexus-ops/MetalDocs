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
- **[r10-t5-durable-async-search-external-effects.md](r10-t5-durable-async-search-external-effects.md)** — operator-ratified T5 authority.
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — current operator-ratified cross-stage disposition baseline.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — current router; **T6 PLATFORM-FACING SUMMARY RATIFICATION NEXT**.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point.

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
T6 material decisions                               OPERATOR-APPROVED
T6 platform-facing summary                          STAGED / RATIFICATION NEXT
T6 durable authority                                NOT YET
T7 Historical Migration & Cutover                   NOT OPEN
implementation                                       BLOCKED
```

## Active T6 decision/gate artifacts

- `../../docs/superpowers/analysis/2026-08-18-r10-t6-operator-material-adjudication.md` — material decisions approved.
- `../../docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary.md` — **current operator ratification target**.

Candidate/evidence artifacts remain staging provenance until T6 promotion and are then removed from the live tree; Git history is the archive.

## T6 platform direction

```text
rebuild pre-launch /api/v1; no compatibility layer
OpenAPI contract-first + generated Go/TS boundaries
Keycloak Authorization Code → ApplicationSession + session-bound CSRF
semantic-lens frontend with stable truth-specific routes
one DRAFT generation via strong ETag/If-Match
T4-bound upload/admission; client never owns exact-content descriptor
immutable Submission governance; reviewer never mutates WorkingContent by case access
singleton User eligibility and singleton Step Decision transport
semantic byte URLs; no provider identity in product contract
fidelity-gated single DOCX provider; no EditorSession correctness baseline
closed TYPE | TYPE_AREA numbering
canonical Search only; materialized Search OFF
Domain history != Audit
RFC9457 + canonical semantic problem codes
natural HTTP idempotency before replay machinery
opaque cursor lists
blank/template/revise seed exact source, never OfficialRendition
```

Current API/frontend/runtime remain evidence only and receive no target entitlement from sunk cost.

## Exact next gate

```text
operator ratifies platform-facing T6 summary
→ durable T6 promotion
→ Decision Registry reconciliation
→ staging cleanup
→ only then T7
```

No implementation plan/product code is authorized.