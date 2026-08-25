---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, blocker, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE       CLEAN-SLATE / ARCHITECTURE-FIRST
REPOSITORY GOVERNANCE REBASELINED / OPERATOR-APPROVED
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                   OPEN / ACTIVE
T11 checkpoint        B01-B10 ACCEPTED / INTEGRATED
B11-F1                ACCESS ASSIGNMENT READ PRECISION / OPERATOR-RATIFIED / INTEGRATED
B11                   NOT OPEN / CONTINUATION PAUSED FOR CLEAN REBASELINE
B12                   NOT OPEN
FP2 / P11             NOT OPEN
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

The repository-governance rebaseline changed repository operation only. B11-F1 separately preserves the already-ratified op31 Access Assignment read precision needed when B11 resumes. Neither increment opens B11 implementation/frontend execution by itself.

## Current system census

```text
semantic owners              4 business + 2 supporting
stable SPA routes            11
PermissionCode values        16
application operations       89
Idempotency-Key creations    11
ETag read / mutation domains 13 / 13
exact-byte resources         4
```

`docs/decisions/api-operation-census.md` owns the detailed census. `docs/decisions/access-assignment-read.md` owns the bounded op31 Access read precision.

## Local methods

```text
engineering-method.md
  DevelopmentConexus Engineering Method v1.0.0 / ACCEPTED

repository-method.md
  DevelopmentConexus Repository Operating Method v1.0.0 / OPERATOR-APPROVED

frontend-product-experience-planning-method.md
  Frontend Product Experience Planning Method v2.3 / ACCEPTED
```

Operating route:

```text
revalidate repository / branch / main / PR / required
→ AGENTS.md
→ roadmap
→ applicable local method(s)
→ docs/index.md
→ 1–2 task owners by default
→ relevant section / operation first
→ expand only when material Evidence can change the conclusion
```

The roadmap is a snapshot, not an execution journal. Detailed iteration/review history belongs to temporary work, Git/PR history, owning decisions or exact Evidence locators when required.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / 89 operations / 11 routes REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

### FP1 blocks

```text
B01   App Shell + Global IA + Home                LOCKED / OPERATOR-RATIFIED
B01N  Notifications global chrome + Quick Inbox  LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                         LOCKED / OPERATOR-RATIFIED
B03   Document Official                           LOCKED / P8-P10 COMPLETE
B04   Document Work                               LOCKED / P8-P10 COMPLETE
B05   My Work                                     LOCKED / P8 R2-P10 COMPLETE
B06   Governance Case                             LOCKED / P8-P10 COMPLETE
B07   Document History                            LOCKED / P8-P10 COMPLETE
B08   Notifications Full Inbox                    LOCKED / P8-P10 COMPLETE
B09   Audit                                       LOCKED / P8-P10 COMPLETE
B10   Organization Administration                 LOCKED / P8-P10 COMPLETE / INTEGRATED
B11   Access Administration                       NOT OPEN / CLEAN REBASELINE NEXT
B12   Document Governance Administration          NOT OPEN
```

B01–B10 exact frontend Evidence remains recoverable through their durable locators; Evidence identity/details do not belong in this roadmap.

B11-F1 is integrated upstream authority only. Prior B11 candidate PR/history remains Evidence of exploration and known findings, not a current complete P8/P9 baseline.

## Repository operating baseline

```text
AGENTS.md                  compact bootstrap + method router
docs/index.md              task/intention → Start / Add when / Do not read by default
docs/decisions/index.md    compact current disposition router
docs/roadmap.md            sole mutable current-status snapshot
repository-method.md       selective context + documentation + Git/Evidence + PR continuity
engineering-rules.md       MetalDocs-specific execution/CI/provenance controls only
Git                        normal archive
docs/work/**               temporary Draft-only work, absent before Ready/main
```

Frontend discovery remains allowed to falsify earlier Product/backend planning when real Evidence proves it insufficient. Engineering + Frontend methods then reopen only the smallest owning authority and boundedly rebaseline affected frontend work.

## Exact next action

```text
1. Finish current PR/branch cleanup so no obsolete active PR remains.
2. Run repository hygiene: classify ordinary branches/refs KEEP / DELETE / NEEDS PROOF before any deletion.
3. Rebaseline B11 from current main under the restored operating model.
4. Preserve B11-F1 as integrated upstream authority.
5. Use prior B11 P8/P9 work only as Evidence of validated structure/findings; rebuild one clean current candidate rather than continuing the 80+ commit workspace.
6. Continue block-by-block frontend planning; reopen upstream authority only when material frontend Evidence proves it necessary.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B11/B12 frontend execution before the cleanup/hygiene rebaseline permits it
no FP2/P11 work
no B01-B10 reopen without material Evidence
no rewrite of Engineering Method or Frontend Method
no branch/ref deletion without explicit operator review of the deletion set
no force-push/shared-history rewrite
no assistant/reviewer LOCK
no merge without explicit operator authorization
```

## Implementation gate

Implementation remains blocked until:

```text
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```
