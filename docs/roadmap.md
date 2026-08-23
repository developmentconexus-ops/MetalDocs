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
docs/decisions/governance-case-step-deadline-read.md
docs/decisions/governance-review-layer-seam.md
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

Method: `docs/development/functional-html-wireframe-method.md` v2.2.

## FP1 block roadmap

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED
B03   Document Official / Ficha + Viewer + Discussion
       LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
B04   Document Work / Authoring
       LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
B05   My Work / Work Queues
       LOCKED / OPERATOR-RATIFIED · P8 R2/P9/P10 COMPLETE
B06   Governance Case
       LOCKED / OPERATOR-RATIFIED
       B06-F1 deadline projection                  CLOSED / OPERATOR-RATIFIED
       B06-F2 Governance Review Layer seam         CLOSED / OPERATOR-RATIFIED / FUTURE-SEAM
       P8 R1 / P9 / P10                            COMPLETE
B07   Document History                             NOT OPEN / NEXT ELIGIBLE
B08   Notifications Full Inbox                     NOT OPEN
B09   Audit                                        NOT OPEN
B10   Organization Administration                 NOT OPEN
B11   Access Administration                       NOT OPEN
B12   Document Governance Administration           NOT OPEN
```

## Locked global IA

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

## B06 closure

Canonical B06 records:

```text
docs/work/current/t11-b06-governance-case-r1.md
docs/work/current/t11-b06-governance-case-functional-wireframe.html
docs/work/current/t11-b06-f2-docx-review-layer.md
docs/work/current/t11-b06-screen-contract.md
docs/work/current/t11-b06-pattern-consolidation.md
```

Locked experience:

```text
Content-first Governance Workspace
→ exact immutable governed content dominant
→ exact Step/deadline context
→ GovernanceFeedback separate from Decision
→ deliberate ACCEPT / RETURN_FOR_CHANGES
→ 409/403 authoritative reconciliation
→ disclosure-neutral unavailable case
→ no WorkingContent mutation / frontend Authorization authority
```

B06-F2 preserves only a future provider-neutral review seam. It adds no present inline-review controls, operation 87+, Permission, lifecycle state or B04 contract change.

P10 reuses:

```text
Global App Shell
Notification Quick Inbox
Exact Read-Only Content Viewer Shell
```

No new shared semantic pattern graduates from B06.

## Exact next action

```text
1. B07 Document History is the next eligible FP1 block and remains NOT OPEN.
2. Open B07 only when the operator chooses to continue FP1.
3. Do not open B08+ early.
4. Implementation remains blocked.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no production framework in P8
no dormant inline-review UI
no framework/vendor redefining Product semantics
no automatic remapping of old review anchors onto changed DRAFT
no generic ReviewAnnotation/EventBus/broker/Redis without a real consumer
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
