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
REPOSITORY READINESS  READY TO RESUME PRODUCT PLANNING
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                   OPEN / ACTIVE
T11 checkpoint        B01-B10 ACCEPTED / INTEGRATED
B11-F1                ACCESS ASSIGNMENT READ PRECISION / OPERATOR-RATIFIED / INTEGRATED
B11                   LOCKED / P8-P10 COMPLETE / ACCEPTANCE CANDIDATE
B12                   NOT OPEN
FP2 / P11             NOT OPEN
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

Repository governance and ordinary historical branch cleanup are **not Product blockers**. Current Product/architecture work starts from `main`; old branches/PRs are Evidence/history unless current authority explicitly names them.

`docs/decisions/repository-readiness.md` records the repository closeout and review posture. `docs/decisions/access-assignment-read.md` owns integrated B11-F1.

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

Independent review uses the bounded ClaudeCode/FABLE-style posture in `docs/development/engineering-rules.md` when a real material gate requires it. It is not an inner-loop requirement for every frontend iteration.

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
B11   Access Administration                       LOCKED / P8-P10 COMPLETE / ACCEPTANCE CANDIDATE
B12   Document Governance Administration          NOT OPEN
```

B01–B11 exact frontend Evidence remains recoverable through their durable locators; Evidence identity/details do not belong in this roadmap.

B11-F1 remains the upstream op31 authority consumed by the B11 acceptance candidate. The old B11 PR remains superseded Evidence only.

## B11 acceptance result

```text
clean P8 R3 operator LOCK                              PASS
P9 material regions / controls                        38 / 38 TRACED
P9 B11 primary operations                             7 / 7 BOUND (ops 27–33)
supporting reads                                      ops 6,16,22
four clean-rebaseline failure classes                 CLOSED
R1 / R2 challenge findings                            CLOSED
fresh independent final challenge                     CONVERGED WITH NON-BLOCKING NOTES
unresolved material findings                          0
application operations                                89 / operation 90 absent
```

Exact Evidence is routed by `docs/decisions/t11-b11-lock-evidence.md`. Temporary `docs/work/**` is absent from the acceptance candidate.

## Exact next action

```text
1. Verify the cleaned B11 acceptance candidate with the repository aggregate gate; `docs/work/**` must remain absent.
2. Inspect current PR review conversations/findings and adjudicate only material current-authority defects.
3. If required CI/review is clean, mark the complete candidate Ready.
4. Stop for explicit operator merge authorization on the exact Ready HEAD.
5. After B11 is accepted/integrated, update the roadmap to open B12 as the next FP1 block.
6. FP2/P11, T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B12 before B11 accepted integration
no FP2/P11 work
no B01-B10 reopen without material Evidence
no rewrite of Engineering Method or Frontend Method
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

Repository cleanup preference alone is not an implementation or Product-planning stop condition.
