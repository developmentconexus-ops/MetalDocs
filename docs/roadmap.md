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

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`. Discussion / `@mention` / Notifications is current under `docs/decisions/discussion-notifications-launch.md`.

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

`docs/decisions/api-operation-census.md` is the sole current numeric census.

## Frontend Product Experience Program

Method:

```text
docs/development/functional-html-wireframe-method.md
Frontend Product Experience Planning Method v2.2
```

Program mapping:

```text
FP0  Frontend Foundation                         CLOSED / R1 86/11 REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

Method meaning:

```text
P8  = canonical functional low-fidelity HTML/CSS/JS per interactive block
P11 = assembled integration of already-LOCKED block prototypes
```

Static storyboard/mockup HTML may support P7 exploration but cannot receive P8 LOCK for an interactive block.

## FP0-R1 closure

Current bounded frontend foundation:

```text
docs/work/current/t11-frontend-foundation-r1.md
```

R1 reconciles the pre-reopen frontend map to current authority without reopening existing LOCKs:

```text
20 accepted human goals mapped
11 stable Product routes represented
operations mapped           86 / 86
operations 79–81            B03 Discussion/Mention
operations 82–86            B01N + B08 Notifications
unassigned operations       0
invented operations         0
frontend Authorization      absent
```

Older `t11-frontend-blueprint.md` and `t11-wireframes.md` remain pre-reopen/pre-v2.2 planning evidence; their 78/10 snapshots do not override R1/current authority.

## FP1 block roadmap

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED

B03   Document Official / Ficha + Viewer + Discussion
       CURRENT / CANDIDATE / NOT LOCKED
       leading structure A                         OPERATOR-APPROVED DIRECTION
       prior static HTML                           REJECTED — WRONG REPRESENTATION MEDIUM
       current R1 planning record                  READY
       canonical functional P8                     RENDERED / OPERATOR OPERATION+REVIEW
       B03-F1 allowed_actions                       OPEN / MUST CLOSE BEFORE FINAL LOCK

B04   Document Work / Authoring                    NOT OPEN
B05   My Work / Work Queues                        NOT OPEN
B06   Governance Case                              NOT OPEN
B07   Document History                             NOT OPEN
B08   Notifications Full Inbox                     NOT OPEN
B09   Audit                                        NOT OPEN
B10   Organization Administration                 NOT OPEN
B11   Access Administration                       NOT OPEN
B12   Document Governance Administration          NOT OPEN
```

## Locked global IA preserved

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

Notifications remains transversal utility chrome, not `Minha Caixa` authority:

```text
utility-header bell + unseen badge
desktop Quick Inbox
narrow/mobile accessible transformation
stable /notifications full Inbox route
```

## B03 current candidate

Current R1 record:

```text
docs/work/current/t11-b03-document-official-r1.md
```

Canonical Method-v2.2 P8:

```text
docs/work/current/t11-b03-document-official-functional-wireframe.html
```

Preserved operator-approved structure:

```text
/documents/:document_id
= stable Document ficha/record first

Ficha
→ deliberate Visualizar documento
→ distinct B03 read-only official-content viewer
→ Voltar para ficha

Ficha
→ bounded current work context
→ management actions from server-derived hints
→ stable-Document Discussion

Notification DOCUMENT_MENTION
→ Quick Inbox
→ same B03 Document
→ Discussion
→ anchor_message_id target
```

Functional P8 must be operated before any LOCK. Local fixture behavior is Evidence only; it does not become Product/server authority.

## B03-F1 — open finding before final LOCK

Candidate precision remains:

```text
DocumentOfficialView.allowed_actions
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request
```

It is a server-derived UX hint only; commands always recheck current truth. B03 cannot receive final LOCK until the precision is durably reconciled or proven unnecessary.

## Exact next action

```text
1. Operator operates the canonical B03 functional P8 in browser.
2. Review only B03 hierarchy, proportions, discoverability and local interaction:
     ficha-first reading order
     current-work separation
     viewer open/back
     management placement
     Discussion placement/density
     reply / @mention composer
     Notification -> exact message anchor
     responsive behavior
3. Iterate only B03 until visual/interaction findings close.
4. Resolve B03-F1 before final operator LOCK.
5. Operator-only B03 LOCK.
6. After LOCK, close P9 Screen Contract / bidirectional trace and bounded P10 pattern pass.
7. Open B04 only after the B03 progression gate permits it.
```

## Hard stops

```text
no Product code/schema/OpenAPI implementation/runtime/deploy work
no T12 work
no production frontend framework in P8
no P8 static storyboard accepted as interactive-block lock evidence
no framework/library allowed to redefine Product semantics
no generic EventBus/broker/Redis without a named material trigger
no frontend Authorization matrix
no legacy implementation restoration by sunk cost
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

Accepted Product/R10/frontend LOCK decisions reopen only on material evidence under the DevelopmentConexus Engineering Method. Preference, sunk cost, framework availability or hypothetical scale are not reopen triggers. A methodology correction may invalidate an evidence artifact without invalidating the underlying operator-approved Product/UX direction.
