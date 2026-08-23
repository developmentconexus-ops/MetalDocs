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

Current bounded T11 authorities:

```text
docs/decisions/discussion-notifications-launch.md
docs/decisions/document-official-actions-read.md
docs/decisions/my-work-governance-identification-read.md
docs/decisions/governance-step-deadline.md
docs/decisions/governance-case-step-deadline-read.md
docs/decisions/governance-review-layer-seam.md
```

Current system census remains:

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
       OPEN / ACTIVE
       B06-F1 deadline projection                  CLOSED / OPERATOR-RATIFIED
       B06-F2 DOCX Review Layer seam               CLOSED / OPERATOR-RATIFIED / FUTURE-SEAM
       P6                                          COMPLETE
       P7 H1 Content-first                         OPERATOR-APPROVED
       P8 R1 functional HTML                       OPERATOR-APPROVED
       P8 R2 inline review                         NOT REQUIRED FOR CURRENT LAUNCH
       LOCK                                        READY / AWAITING EXPLICIT OPERATOR LOCK
       P9 / P10                                    NOT OPEN
B07   Document History                             NOT OPEN
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

## B06 current gate

Current work/evidence:

```text
docs/work/current/t11-b06-governance-case-r1.md
docs/work/current/t11-b06-governance-case-functional-wireframe.html
docs/work/current/t11-b06-f2-docx-review-layer.md
```

Current B06 Product boundary:

```text
/work/governance/:attempt_id
→ exact immutable governed Submission / obsolescence subject
→ ordered Steps + exact persisted deadline context
→ bounded GovernanceFeedback
→ ACCEPT / RETURN_FOR_CHANGES
→ server-derived allowed_actions; no frontend Authorization authority
```

P8 R1 was operated and approved by the operator on 2026-08-23. The selected shape remains **Content-first Governance Workspace**.

B06-F2 is now durable future-seam authority only:

```text
future selected-range review binds to exact immutable reviewed snapshot
stable Document Discussion != inline governance review
DRAFT EditorialComment remains separately deferred
tracked changes/suggestions require separate semantic promotion
vendor/editor ids never become MetalDocs semantic authority
RETURN leaves old review context with old immutable Submission
old anchors never blindly remap onto changed DRAFT
future B04 remediation needs explicit server-authored review-context identity
```

Current no-dormant-capability consequence:

```text
GovernanceFeedback wire      unchanged
allowed_actions              accept | return_for_changes | add_feedback
operations/routes/permissions 86 / 11 / 16
B04 contract                 unchanged
P8 R1 controls               unchanged
P8 R2 inline-review controls absent
```

## Exact next action

```text
1. Operator explicitly LOCKS B06 R1 if satisfied with the already-operated/approved experience.
2. After explicit LOCK, execute B06 P9 Screen Contract.
3. Then execute B06 P10 pattern consolidation.
4. Only after B06 P10 closure may B07 open.
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
