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
       P8 R1 functional HTML                       READY / AWAITING OPERATOR USE
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

Canonical B08 authority/work/evidence:

```text
docs/decisions/notification-inbox-recognition-read.md
docs/work/current/t11-b08-notifications-full-inbox-r1.md
docs/work/current/t11-b08-p8-realization-plan.md
docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html
```

Canonical P8 R1 HTML blob:

```text
bb130535721b2381524763a4885ade5199a15596
```

B08 remains the full persistent Notification triage surface over existing operations 82–86. B01N remains global bell + Quick Inbox; B08 is not `Minha Caixa` and does not rebuild Document Official/Discussion.

### B08-F1 — CLOSED / OPERATOR-RATIFIED

Existing op82 now has bounded human-recognition projection authority:

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

### P8 R1 — READY FOR OPERATOR USE

The functional low-fidelity HTML exercises:

```text
fixed lenses
Nova != Não lida
read/unread
archive/unarchive
mark-all-read
presentation-driven seen batching
loaded-but-off-screen != seen
B03 exact source boundary
access-drift 404 reconciliation
cursor continuation/failure/retry
per-item + mark-all failures
three empty states
SSE invalidation/refetch + disconnect/reconnect
Quick Inbox/full Inbox coherence
responsive/accessibility structure
```

Structural pre-write verification:

```text
HTML parse PASS
33 static ids / 0 duplicates
JavaScript node --check PASS
forbidden Product control hits 0
local blob == repository blob
```

No search/filter/preferences/snooze/priority/delete/bulk-archive platform is admitted.

## Exact next action

```text
1. Operator opens/operates B08 functional P8 R1 from the chat attachment.
2. Exercise lenses, engagement, page-2 seen behavior, source-open/access drift, failures, SSE and Quick Inbox coherence.
3. If friction/finding exists -> revise the same B08 HTML; do not open B09.
4. If operator explicitly LOCKS B08 -> P9 Screen Contract -> P10 pattern consolidation.
5. Do not open B09+ early.
6. Implementation remains blocked.
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
