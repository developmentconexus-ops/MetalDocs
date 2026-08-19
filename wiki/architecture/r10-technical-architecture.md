# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T6 CLOSED / OPERATOR-RATIFIED; T7 ACTIVE / EVIDENCE CENSUS NEXT; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file owns current technical-stage status and exact next action.

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
10. `wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md`
11. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
12. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
13. `wiki/architecture/r10-t6-canonical-api-frontend-journeys.md`
14. `wiki/architecture/rebaseline-decision-registry.md`
15. `wiki/architecture/rebaseline-decision-registry-d4-amendment.md`
16. `wiki/architecture/rebaseline-decision-registry-t6-amendment.md`
17. this router
18. `docs/superpowers/analysis/2026-08-18-r10-t7-historical-migration-cutover-bootstrap.md`
19. actual source-system evidence only where needed for T7

Current/legacy implementation is evidence only and has no compatibility entitlement.

## 2. Binding Method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
Structural Inversion
revalidation does not mean reinvention
prepare the seam, not dormant future capability
```

## 3. Technical descent

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR A1→A10                 CLOSED / OPERATOR-APPROVED
Launch ownership topology                CLOSED / OPERATOR-APPROVED / 4+1
T1 — Semantic State & Invariants         CLOSED / OPERATOR-RATIFIED
T2 — Governance/Effectivity/Tx           CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit               CLOSED / OPERATOR-RATIFIED + D4 amendment
T4 — Exact Content/Storage/Restore       CLOSED / OPERATOR-RATIFIED
T5 — Durable Async/Search/Effects        CLOSED / OPERATOR-RATIFIED
T6 — Canonical API/Frontend Journeys     CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                        CURRENT + D4 + T6 amendments
T7 — Historical Migration & Cutover      ACTIVE / EVIDENCE CENSUS NEXT
implementation                           BLOCKED
```

## 4. T6 closure

Durable T6 authority:

`wiki/architecture/r10-t6-canonical-api-frontend-journeys.md`

Registry reconciliation:

`wiki/architecture/rebaseline-decision-registry-t6-amendment.md`

Closure evidence:

```text
Platform Summary REV2                    OPERATOR-RATIFIED
C1→C8                                    CLOSED
L1→L5                                    CLOSED
D1→D4                                    CLOSED
final exact delta                        APPROVE
new material findings                    0
disagreement set                         EMPTY
completed T6 staging in live tree        REMOVED
Git history                              archive/provenance
```

T6 may reopen only through one of its durable authority reopen triggers or a later material cross-stage contradiction.

## 5. T7 — ACTIVE

Bootstrap:

`docs/superpowers/analysis/2026-08-18-r10-t7-historical-migration-cutover-bootstrap.md`

T7 owns only the remaining Registry REOPEN set:

```text
actual source evidence census
CURRENT_STATE / FULL_HISTORY or smaller real migration-mode set
imported target-owned fact shapes
ordinal/content/governance provenance
plan / dry-run / idempotency / reconciliation
semantic-unit atomicity
cutover / readiness / rollback / deletion map
concrete restore/erasure and post-snapshot security-teardown reconciliation choreography where cutover/recovery requires it
```

### T7 inherited hard boundaries

```text
imported history never becomes fake native history
unknown stays unknown
Historical Migration != generic Interchange domain
migration writes through owning semantic seams
T4 exact-content admission/integrity remains binding
T3 current security/offboarding truth remains binding
historical side effects are never replayed as new current effects
migration convenience cannot redefine Document/Revision/Submission/Release
```

### Exact next action

```text
actual source evidence census
→ classify PROVEN / INFERABLE / UNKNOWN
→ only then derive 2–3 T7 approaches
→ T7 candidate/design
→ material adjudication
→ platform-facing T7 summary
→ explicit operator summary ratification
→ durable promotion + Registry reconciliation + staging cleanup
```

No substantive T7 design is accepted yet.

## 6. Final R10 gate after T7

```text
Integrated Whole-R10 Global Coherence Review
→ cold independent final review
→ operator final ratification
→ implementation spec/plan
→ code
```

Implementation remains **BLOCKED**.