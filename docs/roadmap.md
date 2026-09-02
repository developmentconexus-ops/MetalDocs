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
FP2 / P11             COMPLETE / OPERATOR-ACCEPTED / INTEGRATED
FP2-F1                CREATION COVERAGE GAP / OPERATOR-DISPOSED → B13 OPEN
FP2-F2                CONTENT-FORMAT SCOPE / OPERATOR-RATIFIED → decisions/content-format-vocabulary.md
FP2-F3                DOCUMENT CONFIDENTIALITY / LAUNCH SCOPE
                      RATIFICATION CANDIDATE — authority deltas written; census 89 → 97
                      → decisions/document-confidentiality-launch.md
B02 B03 B06 B10       BOUNDED REBASELINE PENDING (FP2-F3 delta only)
B11 B12               BOUNDED REBASELINE PENDING (FP2-F3 delta only)
B13                   LOCKED R4 / P8-P10 COMPLETE / INTEGRATED
B13-Q1                RESOLVED — hypothesis B (A blocked by B13-F2)
B13-F1                OPEN UPSTREAM FINDING (op44 projection) / NON-BLOCKING
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
FP1  Block-by-block Product Experience           COMPLETE + BOUNDED REOPEN — B13 LOCKED; FP2-F3 rebaseline pending
FP2  Integrated Low-Fidelity Product / P11       COMPLETE / OPERATOR-ACCEPTED
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
B13   Document Creation                            LOCKED / P8 R4-P10 COMPLETE / INTEGRATED
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

B11 pattern-graduation obligation: **CLOSED 2026-08-26.** The collection / idempotency / fixture-truthfulness realization laws are absorbed as durable class-level rules in `docs/architecture/frontend.md` §19, derived from the proved B11 + B12 P10 Evidence (`docs/decisions/t11-b11-lock-evidence.md` + `docs/decisions/t11-b12-lock-evidence.md`). No new checker/framework was introduced.

## B13 acceptance result

```text
operator P8 LOCK                                      R4 / 2026-08-27
P9 material regions                                   9 / 9 TRACED
P9 B13 primary operations                             2 / 2 BOUND (op44, op46)
supporting operations                                 ops 59 → 60 → 58 (source attachment)
P10 headless proof                                    33 / 33 PASS
B13-Q1                                                RESOLVED — hypothesis B (A blocked by B13-F2)
B13-F1                                                OPEN UPSTREAM FINDING / NON-BLOCKING
FP2-F3 confidentiality                                ABSORBED BEFORE LOCK
unresolved material findings                          0
application operations                                89 unchanged by B13
integration                                           OPERATOR-AUTHORIZED 2026-09-02 (PR #183)
```

Exact Evidence is routed by `docs/decisions/t11-b13-lock-evidence.md`. Temporary `docs/work/**` is absent from `main`.

## Exact next action

```text
1. FP2-F3 is OPEN and authorized: docs/decisions/document-confidentiality-launch.md is the
   bounded Product/T6 reopen promoting in-Area document confidentiality into Launch V1.
   The semantic model stays owned by document-confidentiality-seam.md and is NOT reopened;
   only its future-only disposition is. The non-hierarchical class law is decided.
2. FP2-F3 Q1-Q5 are OPERATOR-RATIFIED (2026-08-27): author classifies at creation within
   own clearance; Company-wide vocabulary; materialized default class; total
   non-disclosure in collection reads; governance routing does NOT confer read.
3. B13 is LOCKED at R4 and INTEGRATED (PR #183, operator-authorized 2026-09-02).
4. FP2-F3 authority deltas are WRITTEN: census 97 / 13 idempotent creations / 14 ETag
   domains, T3 clearance term, contract §4 concept + §6 Launch Core, domain-model,
   wire ops 90-97, persistence, frontend §19.4, validation V3/GF1. Ratification is
   operator-only and happens by authorizing the merge.
5. Bounded rebaseline of B02/B03/B06/B10/B11/B12 for the FP2-F3 delta only; proved regions
   untouched. Then re-integrate FP2/P11.
6. FP3/P12 opens only after B13 and the FP2-F3 rebaseline close.
7. FP4/P13-P14, T12 and Product implementation remain BLOCKED.

The census authority now reads 97. Ratification is the operator's merge authorization; `main` carries 89 until then.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no FP3/P12, FP4/P13-P14 work
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
