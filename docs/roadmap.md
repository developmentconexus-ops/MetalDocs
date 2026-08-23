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
docs/decisions/notification-inbox-recognition-read.md
docs/decisions/document-official-actions-read.md
docs/decisions/my-work-governance-identification-read.md
docs/decisions/governance-step-deadline.md
docs/decisions/governance-case-step-deadline-read.md
docs/decisions/governance-review-layer-seam.md
docs/decisions/document-history-recognition-read.md
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
       LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
       B06-F1 deadline projection                  CLOSED / OPERATOR-RATIFIED
       B06-F2 Governance Review Layer seam         CLOSED / OPERATOR-RATIFIED / FUTURE-SEAM
B07   Document History
       LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
       B07-F1 human-recognizable History read      CLOSED / OPERATOR-RATIFIED
B08   Notifications Full Inbox
       OPEN / ACTIVE
       entry recovery                              COMPLETE
       P6                                          COMPLETE
       B08-F1 human-recognizable Inbox read        CLOSED / OPERATOR-RATIFIED
       P7 H1 Focused Triage Inbox                  OPERATOR-RATIFIED
       P8 functional HTML                          OPEN / NEXT
       LOCK / P9 / P10                             NOT OPEN
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

## B07 closure

B07 remains LOCKED. Document History uses server-authored Revision recognition, chronological Revision chapters and exact read-only historical content; History remains distinct from Audit. Its P9/P10 closed with no new operation or generic Timeline/Event abstraction.

## B08 current gate

Canonical work record:

```text
docs/work/current/t11-b08-notifications-full-inbox-r1.md
```

Durable B08-F1 authority:

```text
docs/decisions/notification-inbox-recognition-read.md
```

B08 remains the full persistent Notification triage surface over the already-admitted `/notifications` route:

```text
82 listNotifications
83 updateNotificationEngagement
84 markNotificationsSeen
85 markAllNotificationsRead
86 streamNotificationEvents
```

B01N remains global bell + Quick Inbox. B08 is not `Minha Caixa` and does not rebuild Document Official/Discussion.

### B08-F1 — CLOSED / OPERATOR-RATIFIED

Existing op82 is sufficient and now has bounded human-recognition projection authority:

```text
Notification identity + engagement state
+ current-disclosable DocumentReference
+ exact source message_id
+ author UserReference
+ exact official Revision-at-post when one existed
+ bounded server-composed message preview
```

All source presentation is composed after current disclosure. Notifications persistence remains identity + engagement only; no source text/title/profile/ACL copy is promoted.

Fixed views remain:

```text
active    = presentable + non-archived
unread    = presentable + non-archived + unread
archived  = presentable + archived
```

Current census remains 86 operations / 11 routes / 16 PermissionCode values.

### P7 H1 — OPERATOR-RATIFIED

```text
Focused Triage Inbox
→ heading + unseen/unread summaries
→ Mark all read
→ Caixa de entrada / Não lidas / Arquivadas
→ one canonical recency-ordered Notification list
→ per-item read/unread + archive/unarchive
→ exact source open to B03 Discussion anchor
→ cursor continuation/retry
→ actual presentation, not mere fetch, drives seen candidates
→ SSE only invalidates/refetches canonical op82 truth
```

No current search/filter/preferences/snooze/priority/delete/bulk-archive platform is admitted.

## Exact next action

```text
1. Create the temporary B08 P8 realization plan under docs/work/current.
2. Create B08 functional low-fidelity HTML from the ratified B08-F1 + P7 H1.
3. Verify material local interactions/states and responsive/accessibility structure.
4. Deliver the exact HTML to the operator for browser operation.
5. If friction exists, revise the same B08 P8; do not open B09.
6. Only explicit operator LOCK opens B08 P9 then P10.
7. Implementation remains blocked.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no production framework in P8
no Notifications/Minha Caixa semantic merge
no frontend source-disclosure matrix or post-filter
no copied source content/ACL as Notification authority
no generic Inbox/filter/preferences platform without consumer
no source reply/editor/viewer inside Inbox
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
