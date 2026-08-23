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
docs/decisions/document-history-recognition-read.md
```

B07-F1 authority: [Document History human-recognition read precision](decisions/document-history-recognition-read.md).

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
       LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
       B06-F1 deadline projection                  CLOSED / OPERATOR-RATIFIED
       B06-F2 Governance Review Layer seam         CLOSED / OPERATOR-RATIFIED / FUTURE-SEAM
B07   Document History
       OPEN / ACTIVE
       entry recovery                              COMPLETE
       B07-F1 human-recognizable History read      CLOSED / OPERATOR-RATIFIED
       P6                                          COMPLETE
       P7 H1 Revision Chapters                     OPERATOR-APPROVED
       P8 R1 functional HTML                       READY / AWAITING OPERATOR USE
       LOCK / P9 / P10                             NOT OPEN
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

B06 remains LOCKED. Its exact content, Step/deadline context, GovernanceFeedback/Decision separation, 403/409 reconciliation and future provider-neutral review seam are closed. P10 reused only:

```text
Global App Shell
Notification Quick Inbox
Exact Read-Only Content Viewer Shell
```

No B06 semantic rail/workspace abstraction was generalized.

## B07 current gate

Canonical current evidence:

```text
docs/work/current/t11-b07-document-history-r1.md
docs/work/current/t11-b07-document-history-functional-wireframe.html
```

Current History boundary:

```text
/documents/:document_id/history
→ op47 getDocument for current orientation only
→ op53 getDocumentHistory for Controlled Documents history
→ document.read_history
→ read-only; no History mutation
→ History != Audit
```

Current op53 remains cursor-paginated in server order:

```text
occurred_at ASC,
kind,
semantic id
```

### B07-F1 — CLOSED / OPERATOR-RATIFIED

Operation 53 remains the sole History list read and the 86-operation census is unchanged.

Every `DocumentHistoryItem` now projects exact human-recognition context:

```text
all variants
  revision: RevisionIdentity

governance_decision
  subject_kind
  frozen step_label

feedback_added
  subject_kind

release predecessor
  predecessor_revision?: RevisionIdentity
```

The exact law is owned by `docs/decisions/document-history-recognition-read.md`.

No Audit join, browser historical graph authority, fabricated title-at-revision-creation snapshot, new operation, route, Permission, owner or lifecycle state is introduced.

### P7 — OPERATOR-APPROVED

```text
H1 Revision Chapters + chronological event spine
```

The chronological stream is authoritative. A later event may target an older Revision after a newer Revision cycle has occurred.

P8 R1 exercises this explicitly:

```text
REV001 original cycle / Release
REV002 later cycle / cancellation
REV001 marker appears again later
  obsolescence request / withdrawal
  new request / governance / completion
```

The repeated Revision marker is a P8 UX hypothesis for preserving server chronology without moving later events backward in time.

Content-bearing Submission/Release events open the existing Exact Read-Only Content Viewer Shell.

P8 R1 also exercises cursor continuation/failure, exact-content failure, disclosure-neutral 404, responsive reflow and inherited B01N Quick Inbox.

Verified canonical HTML blob:

```text
20ec64d34085fbc9075b136a61e69c48c0cad981
```

## Exact next action

```text
1. Operator opens/operates B07 functional P8 R1 from the chat attachment.
2. Exercise cursor continuation until REV001 reappears after REV002 cancellation.
3. Exercise Submission/Release historical viewer, pagination failure, content failure and 404 state.
4. If friction/finding exists -> revise the same B07 HTML; do not open B08.
5. If operator explicitly LOCKS B07 -> P9 Screen Contract -> P10 pattern consolidation.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no production framework in P8
no frontend History graph as business authority
no History/Audit semantic merge
no compare/restore/delete without current Product authority
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
