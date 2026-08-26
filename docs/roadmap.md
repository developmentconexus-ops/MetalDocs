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
T11 checkpoint        B01-B11 ACCEPTED / INTEGRATED
B11-F1                ACCESS ASSIGNMENT READ PRECISION / OPERATOR-RATIFIED / INTEGRATED
B11                   LOCKED / P8-P10 COMPLETE / INTEGRATED
B12                   LOCKED / P8-P10 COMPLETE / INTEGRATED
FP2 / P11             NOT OPEN
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

Repository governance and ordinary historical branch cleanup are **not Product blockers**. Current Product/architecture work starts from `main`; old branches/PRs are Evidence/history unless current authority explicitly names them.

`docs/decisions/repository-readiness.md` records the repository closeout. `docs/decisions/access-assignment-read.md` owns integrated B11-F1. `docs/decisions/api-operation-census.md` is the sole current numeric application-operation census authority.

## Local methods

Current work uses the accepted local Engineering Method, Repository Operating Method and Frontend Product Experience Planning Method. `AGENTS.md` owns fresh-session bootstrap/method selection; `docs/index.md` routes tasks to semantic owners.

Independent review uses the bounded ClaudeCode/FABLE posture and Git dialogue protocol in `docs/development/engineering-rules.md` only when a real material gate requires it. It is not an inner-loop requirement for normal frontend iteration.

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
B11   Access Administration                       LOCKED / P8-P10 COMPLETE / INTEGRATED
B12   Document Governance Administration          LOCKED / P8-P10 COMPLETE / INTEGRATED
```

B01–B11 exact frontend Evidence remains recoverable through their durable locators; Evidence identity/details do not belong in this roadmap.

B11-F1 remains the upstream op31 authority consumed by integrated B11. The old B11 PR remains superseded Evidence only.

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

Exact Evidence is routed by `docs/decisions/t11-b11-lock-evidence.md`. Temporary `docs/work/**` is absent from `main`.

## B12 acceptance result

```text
operator P8 LOCK                                      R4 / 2026-08-26
P9 material regions / controls                        22 / 22 TRACED
P9 B12 primary operations                             12 / 12 BOUND (ops 34–43, 50–51)
supporting reads                                      ops 6,16,22
B12-F1 (code/scope immutability read)                 REJECTED by operator — try-and-fail
B12-F2 (op43 read precision)                          RATIFIED → decisions/template-configuration-read.md
unresolved material findings                          0
application operations                                89 unchanged
integration                                           OPERATOR-AUTHORIZED 2026-08-26
```

Exact Evidence is routed by `docs/decisions/t11-b12-lock-evidence.md`. Temporary `docs/work/**` is absent from `main`.

Open B11 pattern-graduation obligation:

```text
the accepted B11 P10 Evidence proves the reusable pattern
→ graduate collection / idempotency / fixture-truthfulness realization law
→ absorb the durable class-level rule into docs/architecture/frontend.md
→ complete this graduation before FP2/P11 opens
```

The durable rule must be derived from the proved B11 P10 Evidence (`docs/decisions/t11-b11-lock-evidence.md`) rather than guessed in advance. No new checker/framework is implied.

## Exact next action

```text
1. Close the open B11 pattern-graduation obligation as its own small acceptance increment:
   derive the collection / idempotency / fixture-truthfulness realization law from the proved
   B11 + B12 P10 Evidence (t11-b11-lock-evidence.md + t11-b12-lock-evidence.md) and absorb the
   durable class-level rule into docs/architecture/frontend.md. No new checker/framework.
2. Only after that graduation closes may later roadmap authority open FP2/P11.
3. FP2/P11, T12 and Product implementation remain BLOCKED until later roadmap authority opens them.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no FP2/P11 work
no B01-B11 reopen without material Evidence
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
