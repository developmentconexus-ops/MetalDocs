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
FP0  Frontend Foundation                         ACTIVE / BOUNDED REBASELINE
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

The v2.2 rebaseline changes the planning method, not Product semantics:

```text
P8  = canonical functional low-fidelity HTML/CSS/JS per interactive block
P11 = assembled integration of already-LOCKED block prototypes
```

Static storyboard/mockup HTML may support P7 exploration but cannot receive P8 LOCK for an interactive block.

## FP0 bounded rebaseline

Current Product authority changed during T11 from the prior frontend foundation (`10 routes / 78 operations`) to current `11 / 86` through the independently challenged Discussion/Notifications amendment.

FP0 therefore has one bounded rebaseline obligation:

```text
update frontend flow/coverage/surface/program maps for:
  stable Document Discussion
  semantic @Mention
  in-app Notifications
  /notifications
  document.discuss
  operations 79–86

preserve valid prior frontend LOCKS unless actually falsified
```

This does not reopen Product/architecture authority.

## FP1 block roadmap

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED

B03   Document Official / Ficha + Viewer + Discussion
       CURRENT / CANDIDATE / NOT LOCKED
       leading structure A                         OPERATOR-APPROVED DIRECTION
       prior static HTML P8                        REJECTED — WRONG REPRESENTATION MEDIUM
       canonical functional P8                     PENDING
       B03-F1 allowed_actions                       OPEN / MUST CLOSE BEFORE LOCK

B04   Document Work / Authoring                    NOT OPEN
B05   My Work / Work Queues                        NOT OPEN
B06   Governance Case                              NOT OPEN
B07   Document History                             NOT OPEN
B08   Notifications Full Inbox                     NOT OPEN
B09   Audit                                        NOT OPEN
B10   Organization Administration                  NOT OPEN
B11   Access Administration                        NOT OPEN
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

## B03 current direction

Preserved operator-approved direction:

```text
/documents/:document_id
= stable Document ficha/record first

Ficha
→ deliberate Visualizar documento
→ distinct B03 read-only official-content viewer

Ficha
→ bounded current work context
→ management actions from server-derived hints
→ stable-Document Discussion

Notification DOCUMENT_MENTION
→ same ficha
→ Discussion
→ anchor_message_id target
```

The rejected B03 artifact does not reject this structure. It was rejected because it rendered static storyboard states instead of the v2.2 canonical functional low-fi block experience.

## Exact next action

```text
1. Complete FP0 bounded 86/11 frontend rebaseline only where mappings are stale.
2. Update B03 planning record to v2.2 semantics.
3. Build canonical B03 functional low-fi HTML/CSS/JS:
     ficha normal
     Visualizar documento -> viewer -> voltar
     Notification/mention -> Discussion anchor
     local Discussion interactions material to layout
     explicit boundaries for unopened B04/B07
4. Operator operates/reviews B03 P8.
5. Iterate B03 only until operator LOCK.
6. Resolve B03-F1 before final B03 LOCK if still open.
7. After LOCK, close P9 Screen Contract / bidirectional trace and bounded P10 pattern pass.
8. Open B04 only after B03 progression gate permits it.
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
