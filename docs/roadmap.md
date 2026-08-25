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
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                   OPEN / ACTIVE
T11 checkpoint        B01-B10 ACCEPTED / INTEGRATED
CURRENT INCREMENT     REPOSITORY GOVERNANCE REBASELINE / ACTIVE CANDIDATE
B11                   NOT OPEN ON INTEGRATED MAIN / CONTINUATION PAUSED
B12                   NOT OPEN
FP2 / P11             NOT OPEN
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

This branch changes **repository operation only**. Product, architecture, accepted B01–B10 frontend LOCKs, 89-operation census and implementation permission are not reopened by this increment.

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

## Local method baseline

Target operating model for this acceptance increment:

```text
engineering-method.md
  DevelopmentConexus Engineering Method v1.0.0

repository-method.md
  DevelopmentConexus Repository Operating Method v1.0.0 / CANDIDATE

frontend-product-experience-planning-method.md
  Frontend Product Experience Planning Method v2.3
```

Engineering and Frontend method bytes remain unchanged. The Repository Operating Method restores local fresh-session recovery, selective context, documentation/Git/Evidence and acceptance-increment continuity without an external methodology router/pin.

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

B01–B10 exact frontend Evidence remains recoverable through their durable locators. Evidence identity/details live in those locators rather than in this roadmap.

## Active governance increment

Goal:

```text
restore fast repository-local operating model
→ AGENTS bootstrap/router
→ three local methods
→ selective task/intention index
→ compact roadmap snapshot
→ compact decision register
→ accurate provenance/CI descriptions
```

Acceptance conditions:

```text
Engineering Method unchanged
Frontend Method unchanged
repository-method present and routed
no active external methodology router/pin claim
no duplicate documentation-governance authority
roadmap remains snapshot, not journal
decision register routes rather than retells
repository-reset describes current controls truthfully
required stays objective and single-entry
docs/work/** absent before Ready
no Product/B11 semantic delta
```

## Exact next action

```text
1. Complete and verify this repository-governance rebaseline on its isolated branch/PR.
2. Remove temporary docs/work/** and make the governance PR Ready only after the complete consistency sweep passes.
3. Obtain explicit operator merge authorization; squash merge if authorized.
4. After governance integration, run a separate repository-hygiene increment:
   - classify branches/refs KEEP / DELETE / NEEDS PROOF;
   - present the deletion list to the operator before deleting anything;
   - evaluate head-branch auto-delete and CI checkout-history cost.
5. Only after governance/hygiene alignment, rebaseline B11 from updated main and continue frontend work under the restored operating model.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B11/B12 frontend execution in this governance increment
no FP2/P11 work
no B01-B10 reopen without material Evidence
no rewrite of Engineering Method or Frontend Method
no branch/ref deletion in this increment
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

Frontend discovery may still falsify earlier Product/backend planning when real Evidence proves it insufficient; the Engineering + Frontend methods then reopen only the smallest affected authority.
