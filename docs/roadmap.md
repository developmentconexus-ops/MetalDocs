---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE       CLEAN-SLATE / ARCHITECTURE-FIRST
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                   OPEN / ACTIVE
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
Draft PR               #162
branch                 arch/t11-implementation-program
opening main           cae6ba48df5d611959c0390e0f2b9b8194d62a9d
```

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`.

Current bounded T11 authorities include:

```text
docs/decisions/discussion-notifications-launch.md
docs/decisions/document-official-actions-read.md
docs/decisions/my-work-governance-identification-read.md
docs/decisions/governance-step-deadline.md
```

Current system census:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           86
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

`docs/decisions/api-operation-census.md` is the sole numeric census.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / R1 86/11 REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

Method: `docs/development/functional-html-wireframe-method.md` — Frontend Product Experience Planning Method v2.2.

## FP1 block roadmap

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED
B03   Document Official / Ficha + Viewer + Discussion
       LOCKED / OPERATOR-RATIFIED
       P8 / P9 / P10                               COMPLETE

B04   Document Work / Authoring
       LOCKED / OPERATOR-RATIFIED
       B04-F1 hybrid persistence                   CLOSED / OPERATOR-RATIFIED
       P8 / P9 / P10                               COMPLETE

B05   My Work / Work Queues
       CURRENT / CANDIDATE / NOT LOCKED
       B05-F1 governance row recognition           CLOSED / OPERATOR-RATIFIED
       B05-F2 neutral governance ordering          REOPENED BY F3
       B05-F3 governance Step deadline             CLOSED / OPERATOR-RATIFIED
       P7 focused queue A                          APPROVED
       P8 R1 base structure                        OPERATOR-APPROVED
       B05-F4 due-aware filter/order behavior      OPEN / NEXT DECISION
       P8 R2                                       NOT OPEN

B06   Governance Case                              NOT OPEN
B07   Document History                             NOT OPEN
B08   Notifications Full Inbox                     NOT OPEN
B09   Audit                                        NOT OPEN
B10   Organization Administration                 NOT OPEN
B11   Access Administration                       NOT OPEN
B12   Document Governance Administration           NOT OPEN
```

## Locked global IA preserved

```text
Início       = current operational situation
Minha Caixa  = assigned work
  Para aprovação
  Em edição
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

Notifications remains transversal utility chrome, not `Minha Caixa` authority.

## B05 current bounded truth

Planning record:

```text
docs/work/current/t11-b05-my-work-r1.md
```

Governance work now has two distinct current precisions:

```text
recognition
  exact governed RevisionReference

temporal attention
  optional active-Step due_at
```

Deadline law:

```text
GovernanceRouteStep.due_in_days?
→ immutable attempt snapshot
→ activation freezes activated_at + due_at
→ 1 day = 24 elapsed hours
→ no hidden default
→ overdue has no automatic lifecycle effect
```

The prior neutral code-first governance order was valid before deadline truth existed; its explicit reopen trigger fired. B05-F4 now owns final server ordering and bounded deadline filters.

P8 R1 validated the focused-queue composition, lane switching, density, keyboard traversal, cursor load-more, stale/error/empty recovery and owner-lens handoffs. B05 remains unlocked because the operator identified the material deadline-triage need after operating R1.

## Exact next action

```text
1. Adjudicate B05-F4 — due-aware governance queue behavior.
2. Decide default server order, bounded deadline filters, null/no-deadline placement and cursor binding.
3. Preserve no manual priority state and no frontend page re-sort.
4. Generate B05 P8 R2 with the approved F4 behavior.
5. Operator operates R2.
6. Operator-only B05 LOCK.
7. Then P9 exact Screen Contract + P10 bounded pattern consolidation.
8. B06+ remain NOT OPEN.
```

## Hard stops

```text
no Product code/schema/OpenAPI implementation/runtime/deploy work
no T12 work
no production frontend framework in P8
no static storyboard accepted as P8 lock evidence
no framework/library redefines Product semantics
no generic EventBus/broker/Redis without a named material trigger
no frontend Authorization matrix
no unopened downstream block design
no legacy restoration by sunk cost
no merge authorization implied
```

## T11 / implementation gate

Implementation remains blocked until:

```text
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Accepted Product/R10/frontend LOCK decisions reopen only on material evidence under the DevelopmentConexus Engineering Method. Preference, sunk cost, framework availability or hypothetical scale are not reopen triggers.
