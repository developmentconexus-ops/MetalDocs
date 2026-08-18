# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; POST-T5 FABLE CHECKPOINT CLOSED; T6 ACTIVE / DESIGN NEXT; T7 NOT OPEN**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. `docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md` — **ACTIVE T6 STAGING**
15. current API/frontend/runtime only as evidence needed to falsify/validate a T6 claim

Completed post-T5 Fable review artifacts were removed from the live tree after explicit operator checkpoint closure; Git history is the archive.

## Current checkpoint

```text
Product Contract                         = REV001 / OPERATOR-APPROVED
Whole-Product GCR A1–A10                 = CLOSED / OPERATOR-APPROVED
Launch ownership topology                = CLOSED / OPERATOR-APPROVED / 4+1
T1 Semantic State & Invariants           = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx             = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit                 = CLOSED / OPERATOR-RATIFIED
T4 Exact Content/Storage/Restore         = CLOSED / OPERATOR-RATIFIED
T5 Durable Async/Search/Effects          = CLOSED / OPERATOR-RATIFIED
Decision Registry                        = CURRENT / RECONCILED / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint      = CLOSED / OPERATOR-APPROVED
T6 Canonical API / Frontend Journeys     = ACTIVE / DESIGN NEXT
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## Post-T5 checkpoint result

```text
DELTA VERDICT = APPROVE
M1–M3 = CLOSED
L1–L5 = CLOSED
NEW MATERIAL FINDINGS = 0
DISAGREEMENT SET = EMPTY
T6 READINESS = MAY OPEN
```

No formal T1→T5 reopen occurred. Ratified amendments remain durable in Product Contract REV001, T1→T5 and the Registry.

## T6 official REOPEN set

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
source upload / T4 admission UX
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
exact Search field/ranking UX + prove whether any derived/expensive fact activates materialized Search seam
EditorSession/UX lease only if a real editor-integration consumer requires it
```

Also carry the non-blocking Fable observation into T6: DRAFT retitle mutation/concurrency must sit explicitly under one existing T2 concurrency law without reopening Revision-owned title semantics.

## T6 law

T6 is **architecture/design only**. Do not jump to endpoint tables, OpenAPI, screen trees or provider/code implementation before T6 material decisions are derived and operator-adjudicated.

Mandatory close:

```text
T6 candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ durable promotion
→ Decision Registry update
→ remove T6 staging
→ only then T7
```

No final SQL/index/package/process topology, Historical Migration execution plan, implementation plan or product code is authorized.
