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
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — current router; **T6 FINAL GLOBAL-MAXIMUM ADJUDICATION READY**.
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
T6 Canonical API / Frontend Journeys                ACTIVE / FINAL ADJUDICATION READY
operator material adjudication                      NEXT
T7 Historical Migration & Cutover                   NOT OPEN
implementation                                       BLOCKED
```

## Active T6 staging / precedence

```text
bootstrap
→ docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md

base candidate
→ docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md

evidence docket
→ docs/superpowers/analysis/2026-08-18-r10-t6-external-evidence-docket.md

corrected adjudication
→ docs/superpowers/analysis/2026-08-18-r10-t6-global-maximum-adjudication-packet.md

final FR-1..FR-4 precedence
→ docs/superpowers/analysis/2026-08-18-r10-t6-final-adjudication-refinements.md
```

Operator decision precedence is base candidate → corrected packet → final refinements where named. None is durable authority yet.

## T6 Global-Maximum direction

```text
rebuild pre-launch /api/v1; no legacy compatibility layer
server-side OIDC Authorization Code + MetalDocs ApplicationSession + CSRF
semantic-lens frontend, stable route meanings
one DRAFT generation expressed by ETag/If-Match; PATCH title/source under same OCC
T4-bound upload/admission; client never owns exact-content descriptor
immutable Submission governance; singleton Step Decision resource
singleton User eligibility resource executes T3 offboarding/reenable semantics
semantic exact-byte URLs; no provider identity in product contract
fidelity-gated single DOCX adapter; no EditorSession baseline
closed TYPE | TYPE_AREA numbering; no custom grammar
canonical Search only; materialized Search OFF
Domain history != Audit
RFC9457 errors + semantic problem codes
natural HTTP idempotency before replay machinery
cursor lists; no generic filter DSL
template/revision editing seeds exact released source, never OfficialRendition
```

Current API/frontend/runtime remain evidence only and receive no target entitlement from sunk cost.

## Exact next gate

```text
operator adjudicates final T6 material slate
→ revise only rejected/refined items
→ platform-facing T6 summary
→ explicit operator summary ratification
→ durable T6 promotion + Registry reconciliation + staging cleanup
→ only then T7
```

No implementation plan/product code is authorized.
