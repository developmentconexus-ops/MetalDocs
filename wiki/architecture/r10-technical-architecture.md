# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; POST-T5 FABLE CHECKPOINT CLOSED; T6 ACTIVE / DESIGN NEXT; T7 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation gate:** **CLOSED — design/documentation only**

This file owns current technical-stage status, reading order and exact next action.

## 1. Binding authority chain

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. this router
14. active T6 staging
15. current API/frontend/runtime only as evidence for a concrete T6 claim

Historical/current implementation is never target authority by existence.

## 2. Binding method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

## 3. Technical descent

```text
Product Contract                                        REV001 / OPERATOR-APPROVED
T1 — Semantic State & Invariants                       CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                CLOSED / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore       CLOSED / OPERATOR-RATIFIED
T5 — Durable Async, Search & External Effects         CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Registry                      CURRENT / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                   CLOSED / OPERATOR-APPROVED
T6 — Canonical API / Frontend Journeys                ACTIVE / DESIGN NEXT
T7 — Historical Migration & Cutover                   NOT OPEN
implementation                                         BLOCKED
```

## 4. Post-T5 checkpoint — CLOSED

The independent review and delta review ended with:

```text
DELTA VERDICT = APPROVE
M1–M3 = CLOSED
L1–L5 = CLOSED
NEW MATERIAL FINDINGS = 0
DISAGREEMENT SET = EMPTY
T6 READINESS = MAY OPEN
```

The operator explicitly closed the checkpoint on 2026-08-18. Completed Fable staging was removed from the live tree; Git history remains the archive. No formal T1→T5 reopen occurred.

Promoted bounded amendments remain in Product Contract REV001, T1→T5 and the Decision Registry.

## 5. T6 — Canonical API / Frontend Journeys — ACTIVE

Active staging:

`docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md`

T6 consumes only the current Registry REOPEN set:

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

The post-Fable non-blocking retitle observation is also a T6 proof question: place DRAFT retitle mutation under one existing T2 concurrency law without reopening Revision-owned title semantics.

### T6 hard boundaries

T6 may not casually reopen:

```text
Document/Revision/WorkingContent/Submission meaning
Release/effectivity
T3 Authorization/Audit authority
T4 exact-content/admission authority
viewer/preview vs OfficialRendition distinction
Search authority boundary
canonical Search baseline / conditional materialization law
no-notification/event-platform baseline
```

T6 does not own Historical Migration execution.

## 6. Mandatory T-stage closure protocol

```text
read Decision Registry
→ consume CURRENT / PRESERVE / REFINED
→ design only T6 REOPEN set
→ candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promote durable T6 conclusions
→ update Decision Registry
→ remove completed T6 staging
→ only then open T7
```

A technical-decision approval alone does not open T7.

## 7. T7 — NOT OPEN

T7 remains Historical Migration & Cutover and consumes only its Registry REOPEN set, including restore/erasure/post-snapshot security-teardown reconciliation choreography.

## 8. Final gate

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent final review
→ operator final ratification
→ implementation spec/plan
→ code
```

**Implementation remains BLOCKED.**
