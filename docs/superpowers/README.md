# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Current gate:** **POST-T5 FABLE CHECKPOINT CLOSED; T6 CANONICAL API / FRONTEND JOURNEYS ACTIVE; T7 NOT OPEN; IMPLEMENTATION BLOCKED.**

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

- `analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md` — **T6 ACTIVE / NON-AUTHORITATIVE BOOTSTRAP; DESIGN NEXT.**

Completed T5 and post-T5 Fable staging was removed after promotion/checkpoint closure. Git history is the archive.

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
T6 Canonical API / Frontend Journeys                ACTIVE / DESIGN NEXT
T7 Historical Migration & Cutover                   NOT OPEN

→ T6 material decisions
→ T6 platform-facing summary + operator ratification
→ T7
→ Integrated Whole-R10 GCR
→ cold independent final review
→ final operator ratification
→ implementation spec/plan
→ code
```

## T6 official REOPEN set

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
source upload / T4 admission UX
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
exact Search field/ranking UX + materialization proof
EditorSession/UX lease only if a real editor-integration consumer requires it
```

T6 must preserve the ratified T1→T5 authority boundaries. No product implementation or implementation plan is authorized while design gates remain open.
