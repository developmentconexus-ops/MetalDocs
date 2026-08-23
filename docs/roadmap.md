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
       record/ficha-first semantics                 OPERATOR-APPROVED
       historical C two-column dossier P7          OPERATOR-APPROVED
       canonical functional P8 R2                  OPERATOR-APPROVED / P8 COMPLETE
       prior static storyboard                     REJECTED — WRONG REPRESENTATION MEDIUM
       B03-F1 allowed_actions                       CANDIDATE READY / OPERATOR ADJUDICATION

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

## B03 operator-approved P8

Current planning record:

```text
docs/work/current/t11-b03-document-official-r1.md
```

Canonical Method-v2.2 P8 R2:

```text
docs/work/current/t11-b03-document-official-functional-wireframe.html
```

Operator-approved structure:

```text
/documents/:document_id
= stable Document ficha/record first

Document hero
↓
Two-column dossier
  left
    current-work context
    ficha / classification / responsibility
    server-hinted management actions

  right
    official-content preview
    exact current official Revision label
    deliberate Visualizar completo
↓
Revisions context — full width
↓
Stable-Document Discussion — full width
```

Preview law:

```text
contextual recognition only
never exact-content authority
never DRAFT substitution
click -> same separate B03 read-only official viewer
```

Current interaction flow:

```text
Ficha / preview
→ deliberate Visualizar documento
→ distinct B03 read-only official-content viewer
→ Voltar para ficha

Notification DOCUMENT_MENTION
→ Quick Inbox
→ same B03 Document
→ Discussion
→ anchor_message_id target
```

The operator approved the P8 R2 content, interaction behavior, layout and proportions. Local fixture behavior remains Evidence only; it does not become Product/server authority.

## B03-F1 — current final pre-LOCK gate

Candidate analysis:

```text
docs/work/current/t11-b03-f1-document-official-actions.md
```

Recommended precision:

```text
DocumentOfficialAction =
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request

DocumentOfficialView.allowed_actions: unique DocumentOfficialAction[]
```

Candidate law:

```text
server-derived UX hints only
same canonical Authorization + Controlled Documents predicates used by commands
required array on a disclosed DocumentOfficialView; may be []
canonical order fixed
commands always recheck current truth
no denial reasons / no frontend permission matrix
no new operation / route / Permission / semantic owner / persistence / ETag domain
```

B03-F1 is not yet durable authority. Operator adjudication is required before promotion and before final B03 LOCK.

## Exact next action

```text
1. Operator adjudicates B03-F1 allowed_actions candidate.
2. If approved:
     promote bounded precision into current T6/T8-E/T8-F authority
     reconcile executable DocumentOfficialView schema
     mark B03-F1 CLOSED / OPERATOR-RATIFIED
3. Operator-only final B03 LOCK.
4. Close B03 P9 Screen Contract / bidirectional trace.
5. Run bounded P10 pattern consolidation.
6. Open B04 only after the B03 progression gate permits it.
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
no unopened downstream block design
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
