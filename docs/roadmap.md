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

Method:

```text
docs/development/functional-html-wireframe-method.md
Frontend Product Experience Planning Method v2.2
```

```text
FP0  Frontend Foundation                         CLOSED / R1 86/11 REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

P8 means browser-operable functional low-fidelity evidence. P11 assembles already-LOCKED blocks; it is not the first functionalization.

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
       authority + legacy ergonomics recovery      COMPLETE
       B05-F1 governance row RevisionReference     CLOSED / OPERATOR-RATIFIED
       B05-F2 governance queue ordering            CLOSED / OPERATOR-RATIFIED
       P7 focused queue A                          APPROVED
       functional low-fi P8 R1                     RENDERED / OPERATOR OPERATION+REVIEW

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

## B03 / B04 locked references

```text
B03
  docs/work/current/t11-b03-document-official-r1.md
  docs/work/current/t11-b03-document-official-functional-wireframe.html
  docs/work/current/t11-b03-screen-contract.md
  docs/work/current/t11-b03-pattern-consolidation.md

B04
  docs/work/current/t11-b04-document-work-r1.md
  docs/work/current/t11-b04-document-work-functional-wireframe.html
  docs/work/current/t11-b04-screen-contract.md
  docs/work/current/t11-b04-pattern-consolidation.md
```

B04 remains the exact current open-Revision Work lens; B03 remains the stable Document official/management lens.

## B05 current candidate

Planning record:

```text
docs/work/current/t11-b05-my-work-r1.md
```

Current governance-work read precision:

```text
WorkGovernanceItem {
  governance_attempt_id
  subject_kind
  document
  revision: RevisionReference
  created_at
}

listGovernanceWork fixed order
  document.code ASC,
  governance_attempt_id ASC
```

The projection remains read-only recognition/navigation truth. B06 remains authority for Governance Case state, Steps, feedback, governed content, allowed actions and decisions.

Operator-approved P7 structure:

```text
Minha Caixa
→ intent switch
    Para aprovação | Em edição
→ one focused full-width queue for the selected intent
→ dense human-recognizable rows
→ server cursor order preserved
→ owner-lens continuation
→ load-more cursor continuation
```

P8 R1 exercises:

```text
intent switching
row selection + keyboard navigation
cursor/load-more behavior
stale destination + refresh
load failure + retry
empty lane
B04 handoff boundary
B06 unopened boundary
B01N Quick Inbox reuse
responsive reflow
```

P8 R1 deliberately renders `SUBMITTED` under the currently LOCKED `Em edição` label. This is a test, not a B01 reopen. Reopen B01 terminology only if operator use proves material confusion.

## Exact next action

```text
1. Operator operates B05 functional P8 R1.
2. Review only B05 scanability / density / lane switching / pagination / stale recovery / responsive behavior.
3. Explicitly judge whether SUBMITTED under "Em edição" is understandable.
4. Iterate only material B05 findings.
5. Operator-only B05 LOCK.
6. Then P9 exact Screen Contract + P10 bounded pattern consolidation.
7. B06+ remain NOT OPEN until normal progression permits them.
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
