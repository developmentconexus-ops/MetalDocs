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
B11                   CLEAN REBASELINE AUTHORIZED / NEXT
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
B11   Access Administration                       CLEAN REBASELINE AUTHORIZED / NEXT
B12   Document Governance Administration          NOT OPEN
```

B01–B10 exact frontend Evidence remains recoverable through their durable locators; Evidence identity/details do not belong in this roadmap.

B11-F1 is current upstream authority. The old B11 PR is superseded Evidence only and must not be resumed as the current workspace.

## B11 clean-rebaseline inputs

Start one new B11 acceptance increment from current `main`.

Preserve:

```text
integrated B11-F1 op31 precision
+ accepted Authorization / Role / Scope semantics
+ prior B11 structure only where current Evidence still supports it
```

The first clean candidate must address these already-known failure classes:

1. paginated member/selector reads use visible server-page traversal; no hidden all-page crawl;
2. op6 User selection preserves raw server-page boundaries — DISABLED Users remain visible but unavailable, never pre-filtered before pagination;
3. add-member UX does not assume complete Group membership knowledge; idempotent op28 `PUT` reconciles first-add vs already-member truth;
4. repeated grant confirmation / same Idempotency-Key causes zero second semantic mutation.

These findings do **not** authorize operation 90, `Group.area_id`, custom Role editing or a browser effective-access engine.

## Exact next action

```text
1. Create a new B11 branch from current main.
2. Recover only the B11 authority pack through AGENTS → roadmap → Engineering + Frontend methods → docs/index.
3. Use PR #173/R5-R7 only as Evidence, not as current authority or a workspace to continue.
4. Produce one clean current P8 candidate that already closes the four known failure classes.
5. Operator operates/iterates P8 and sets LOCK only when satisfied.
6. Close exact P9/P10 against the locked candidate.
7. Run one strong independent adversarial challenge when the material B11 candidate reaches its real final gate; do not run FABLE as an inner-loop tax.
8. Integrate the coherent B11 increment, then open B12.
```

Ordinary old branch/ref deletion is administrative cleanup and may occur when suitable tooling is available; it does not precede step 1 unless a specific ref is proven to threaten provenance or current authority.

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
