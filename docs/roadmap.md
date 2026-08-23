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

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`. Discussion / `@mention` / Notifications is current under `docs/decisions/discussion-notifications-launch.md`. Document Official management action-hint precision is current under `docs/decisions/document-official-actions-read.md`.

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
       LOCKED / OPERATOR-RATIFIED
       P7 historical C Two-column dossier          APPROVED
       functional low-fi P8 R2                     APPROVED / COMPLETE
       B03-F1 allowed_actions                       CLOSED / OPERATOR-RATIFIED
       P9 Screen Contract                          COMPLETE
       P10 bounded pattern pass                    COMPLETE

B04   Document Work / Authoring
       CURRENT / CANDIDATE / NOT LOCKED
       content-first workspace P7                  OPERATOR-APPROVED
       DOCX → Eigenpal editable boundary           OPERATOR-APPROVED
       PDF / SUBMITTED read-only boundary          OPERATOR-APPROVED
       right operational rail                      OPERATOR-APPROVED
       B04-F1 hybrid persistence UX                CLOSED / OPERATOR-RATIFIED
       functional low-fi P8 R1                     RENDERED / CI #1320 SUCCESS / OPERATOR OPERATION+REVIEW

B05   My Work / Work Queues                        NOT OPEN
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

## B03 locked baseline

Planning/closure records:

```text
docs/work/current/t11-b03-document-official-r1.md
docs/work/current/t11-b03-document-official-functional-wireframe.html
docs/work/current/t11-b03-screen-contract.md
docs/work/current/t11-b03-pattern-consolidation.md
```

Durable B03 read precision:

```text
docs/decisions/document-official-actions-read.md
```

Locked composition:

```text
/documents/:document_id
= stable Document ficha/record first

Document hero
↓
Two-column dossier
  left
    current-work context
    ficha / classification / responsibility
    server-derived management actions

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

Management action law:

```text
DocumentOfficialView.allowed_actions
-> UX guidance only
-> same canonical predicates as commands
-> commands always recheck current truth
-> no frontend Authorization evaluator
```

P9 closure:

```text
material B03 regions/controls traced        15 / 15
unbound material controls                   0
invented operations                         0
screen-shaped APIs                          0
frontend Authorization evaluator            0
navigation identities unsourced             0
material B03 Screen Contract findings       0
```

P10 closure:

```text
existing locked shared patterns reused      2
new shared abstractions created             0
B03-local semantic patterns retained        7
false abstractions introduced               0
```

## B04 current P8 candidate

Planning record:

```text
docs/work/current/t11-b04-document-work-r1.md
```

Canonical Method-v2.2 functional evidence:

```text
docs/work/current/t11-b04-document-work-functional-wireframe.html
```

Operator-approved P7 structure:

```text
/documents/:document_id/work
= exact current open Revision Work lens

MetalDocs minimal Work header
↓
CONTENT-FIRST WORKSPACE
  main canvas
    DOCX DRAFT     → Eigenpal toolbar/chrome + editable DOCX canvas
    PDF DRAFT      → read-only viewer + source replacement path
    SUBMITTED      → read-only exact submitted-content view

  right rail
    Trabalho atual
    Fonte
    Ações
    Contexto do documento — collapsed by default
```

Legacy comparison disposition:

```text
preserve useful context density and visible save state
reject legacy Work + Approval + History mode-adaptive collapse
full revision history     -> B07
approval timeline/actions -> B06
full ficha                -> B03
```

B04-F1 hybrid persistence law:

```text
local Eigenpal changes become DIRTY immediately
background autosave coalesces after an implementation-appropriate quiet period
at most one save pipeline in flight
Salvar agora / Ctrl+S force the same pipeline
save accepts new DocumentWorkView + strong DRAFT ETag
save failure preserves local buffer
412 stops stale autosave and requires explicit human reconciliation
no automatic merge / silent LWW
submit waits for in-flight save and force-flushes remaining local changes
submit proceeds only against the exact accepted DRAFT ETag
no IndexedDB/localStorage/offline durable client DRAFT baseline
```

Work truth laws remain:

```text
DocumentWorkView + DRAFT ETag = server truth
local editor buffer            = FORM DRAFT only
provider PUT success != READY
READY != WorkingContent
WorkingContent != Submission
```

P8 R1 exercises locally:

```text
DOCX edit → dirty → autosave/save-now → new ETag
coalesced edits while save is in flight
save failure + retry path
412 conflict + preserved local input + explicit reconciliation
PDF read-only Work
source replacement allocation → PUT → READY → attach
expired upload → same local bytes + new allocation
submit force-flush → SUBMITTED read-only
withdraw Submission → same Revision DRAFT
cancel Revision reason → no-current-work state
return to B03; no History fallback
responsive rail reflow
```

Review-only controls that force failure/conflict/expiry are Evidence only, not Product UI.

## Exact next action

```text
1. Operator operates B04 functional P8 R1 in browser.
2. Review only B04 layout / discoverability / persistence / recovery behavior:
     Eigenpal canvas vs rail balance
     save-state visibility
     autosave + Salvar agora feel
     submit force-flush understanding
     source replacement progression
     expired-upload recovery
     412 reconciliation clarity
     DRAFT vs SUBMITTED separation
     withdraw/cancel behavior
     narrow/mobile reflow
3. Iterate only B04 material findings.
4. Operator-only B04 LOCK.
5. Then P9 exact Screen Contract + P10 bounded pattern pass.
6. B05+ remain NOT OPEN until normal progression permits them.
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
