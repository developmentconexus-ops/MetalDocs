# R10-T6 — Canonical API / Frontend Journeys — Stage Bootstrap

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T6 CANDIDATE READY / OPERATOR MATERIAL ADJUDICATION NEXT**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product Contract:** `wiki/architecture/launch-v1-product-contract.md` — REV001  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This file routes T6. Current implementation is evidence only and creates no compatibility obligation. The active material candidate is:

`docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md`

The candidate is **not authority** until material decisions are operator-adjudicated, summarized functionally, explicitly ratified and promoted.

## 1. Authority order

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. this bootstrap
15. active T6 material candidate
16. current API/frontend/runtime only as evidence needed to falsify or validate a concrete T6 claim

Historical/current implementation shape is not target authority.

## 2. Binding laws

```text
smallest sustainable solution
one authority per meaning
mechanism != authority
proof before implementation
revalidation does not mean reinvention
prepare the seam, not the dormant capability
legacy/current implementation = evidence only, never compatibility requirement
```

T6 may delete/rewrite every current route/screen/DTO/module if the greenfield target requires it.

## 3. T1→T5 baseline T6 may not casually reopen

```text
Document != Revision != WorkingContent != Submission
REV000 initial issuance / REV001 first revision
human-readable title = Revision-governed metadata
WorkingContent OCC/CAS = DRAFT concurrency authority
Submission immutable exact attempt
one sequential Governance Step model
Release = sole normal effectivity authority
bounded withdrawal of active human-governed obsolescence request
current Authorization = live grants + scope + domain predicates
Audit = action evidence, not current state
ExactContentDescriptor = SHA-256 + exact size + ContentFormat
managed_content_id = retrieval mechanism only
OPEN→READY/admission/malware laws remain T4 authority
viewer/preview != OfficialRendition
Search journey required; baseline = canonical PostgreSQL query/view
materialized Search/search_refresh only if a real consumer proves it
no mandatory Launch notifications/event bus
no generic integration/event platform
```

## 4. Official T6 REOPEN set

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
source upload / T4 admission UX
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
exact Search field/ranking UX + materialization proof question
EditorSession/UX lease only if a real editor-integration consumer requires it
```

The post-Fable non-blocking retitle question is also in T6: DRAFT title mutation must use one existing T2 concurrency law without reopening Revision-owned title semantics.

## 5. Current candidate direction

The material candidate recommends a greenfield semantic-lens surface:

```text
Library
My Work
Document official lens
Document work lens
exact Governance case lens
Document history
Audit
Administration
```

It explicitly rejects preserving current public `Approvals`, `Templates`, `Controlled Documents`, Distribution, writer-session or legacy capability surfaces by inertia.

Candidate material decisions are T6-A→T6-R in the active candidate file.

## 6. Stage protocol

```text
candidate T6-A→T6-R
→ operator material-decision adjudication NEXT
→ platform-facing T6 summary
→ explicit operator summary ratification
→ promote durable T6 authority
→ update Decision Registry
→ remove completed T6 staging
→ only then open T7
```

## 7. Current gate

```text
Product Contract REV001        OPERATOR-APPROVED
T1→T5                          CLOSED / OPERATOR-RATIFIED
Post-T5 Fable checkpoint       CLOSED / OPERATOR-APPROVED
Decision Registry              CURRENT / OPERATOR-RATIFIED
T6 evidence/inversion pass     COMPLETE ENOUGH FOR CANDIDATE
T6 material candidate          READY / NON-AUTHORITATIVE
operator adjudication          NEXT
T7                             NOT OPEN
implementation                 BLOCKED
```

Do not implement from the candidate.