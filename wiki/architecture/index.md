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
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — current router; **T6 CORRECTED GLOBAL-MAXIMUM ADJUDICATION READY**.
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
T6 Canonical API / Frontend Journeys                ACTIVE / ADJUDICATION READY
operator material adjudication                      NEXT
T7 Historical Migration & Cutover                   NOT OPEN
implementation                                       BLOCKED
```

## Active T6 staging

- `../../docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md` — stage scope / hard boundaries.
- `../../docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md` — material architecture candidate + alternatives / Structural Inversion.
- `../../docs/superpowers/analysis/2026-08-18-r10-t6-external-evidence-docket.md` — primary/current evidence + claim boundaries.
- `../../docs/superpowers/analysis/2026-08-18-r10-t6-global-maximum-adjudication-packet.md` — **corrected material disposition / operator decision target**.

When the corrected packet is more specific than the base candidate, it is the proposed T6 disposition. Neither staging document is durable authority before operator ratification.

## T6 Global-Maximum direction

```text
rebuild pre-launch /api/v1 from current semantics; no legacy compatibility layer
Keycloak Authorization Code → MetalDocs ApplicationSession; no local credential API
semantic-lens frontend rather than legacy module navigation
one DRAFT ETag/If-Match OCC token for title + WorkingContent
T4-bound upload OPEN→READY→OCC attach; client never owns descriptor
review exact immutable Submission; no reviewer WorkingContent mutation
semantic byte resources hide provider/storage identity
fidelity-gated single DOCX adapter; no EditorSession baseline
closed TYPE | TYPE_AREA numbering; no generic formatting language
Search materialization/search_refresh OFF for Launch
domain history and Audit remain separate
RFC9457 error contract + closed semantic problem families
natural HTTP idempotency first; targeted Idempotency-Key replay
opaque cursor pagination for unbounded lists
blank/template/revise seed exact source semantics; never OfficialRendition as editable source
```

Current API/frontend/runtime remain evidence only. No existing route, screen, DTO, module or provider receives target-architecture entitlement from sunk cost.

## T6 hard boundary

Everything outside T6's current REOPEN set stays frozen unless a material counterexample explicitly reopens it. T6 does not own Historical Migration execution.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed.**

## Exact next gate

```text
operator adjudicates corrected T6 material slate
→ revise only rejected/refined items
→ platform-facing T6 summary
→ explicit operator summary ratification
→ durable T6 promotion + Decision Registry reconciliation + staging cleanup
→ only then T7
```

No implementation plan/product code is authorized.
