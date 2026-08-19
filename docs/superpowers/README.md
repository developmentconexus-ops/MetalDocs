# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Current gate:** **T6 MATERIAL CANDIDATE READY / OPERATOR ADJUDICATION NEXT; T7 NOT OPEN; IMPLEMENTATION BLOCKED.**

Durable accepted truth belongs in `wiki/`. Active, not-yet-promoted design analysis belongs here. Completed/superseded staging is removed from the live tree and remains recoverable from Git history.

## Current durable authority

```text
wiki/architecture/launch-v1-product-contract.md          REV001
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-t1-semantic-state-invariants.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
→ wiki/architecture/r10-t3-authorization-audit-enforcement.md
→ wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
→ wiki/architecture/r10-t5-durable-async-search-external-effects.md
→ wiki/architecture/rebaseline-decision-registry.md
→ wiki/architecture/r10-technical-architecture.md
```

## Current active staging

- `analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md` — T6 router/bootstrap.
- `analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md` — **T6-A→T6-R MATERIAL CANDIDATE / NON-AUTHORITATIVE / OPERATOR ADJUDICATION NEXT.**

Completed T5 and post-T5 Fable staging was removed after promotion/checkpoint closure. Git history is the archive.

## Greenfield law for T6

Current routes/modules/screens/DTOs/capabilities are legacy/current-state evidence only. T6 has no compatibility obligation to preserve them.

```text
Product Contract + T1→T5
→ Structural Inversion
→ smallest sustainable API/UX
```

Do not preserve a legacy surface because migration would be easier.

## Active technical path

```text
Product Contract                                      REV001 / OPERATOR-APPROVED
T1 Semantic State & Invariants                       CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore       CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects         CLOSED / OPERATOR-RATIFIED
Decision Registry                                    CURRENT / RECONCILED
Post-T5 Fable checkpoint                             CLOSED / OPERATOR-APPROVED
T6 material candidate                               READY / ADJUDICATION NEXT
T7 Historical Migration & Cutover                   NOT OPEN

→ operator adjudicates T6-A→T6-R
→ T6 platform-facing summary + operator ratification
→ durable T6 promotion / registry update / staging cleanup
→ T7
→ Integrated Whole-R10 GCR
→ cold independent final review
→ final operator ratification
→ implementation spec/plan
→ code
```

No product implementation or implementation plan is authorized while design gates remain open.