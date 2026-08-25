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
B11                   NOT OPEN ON INTEGRATED MAIN / CONTINUATION PAUSED FOR REBASELINE
B12                   NOT OPEN
FP2 / P11             NOT OPEN
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

The governance rebaseline changes repository operation only. Product, architecture, accepted B01–B10 frontend LOCKs, the 89-operation census and implementation permission are unchanged.

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

`docs/decisions/api-operation-census.md` owns the detailed census.

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
B11   Access Administration                       NOT OPEN ON INTEGRATED MAIN
B12   Document Governance Administration          NOT OPEN
```

B01–B10 exact frontend Evidence remains recoverable through their durable locators; Evidence identity/details do not belong in this roadmap.

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

After this governance baseline is integrated:

```text
1. Open a separate repository-hygiene acceptance increment.
2. Classify branches/refs as KEEP / DELETE / NEEDS PROOF.
3. Present every proposed deletion to the operator before deleting anything.
4. Evaluate ordinary merged-head auto-delete and CI checkout-history cost with measured proof.
5. Rebaseline B11 from the updated main using the restored operating route.
6. Continue B11 frontend planning from current accepted Product/architecture truth, preserving valid prior Evidence and reopening only material contradictions.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B11/B12 frontend execution before the governance/hygiene rebaseline sequence permits it
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
