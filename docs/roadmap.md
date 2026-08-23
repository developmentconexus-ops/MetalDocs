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
       LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
       B08-F1 human-recognizable Inbox read        CLOSED / OPERATOR-RATIFIED
B09   Audit                                        NOT OPEN / NEXT ELIGIBLE
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

## B08 closure

Canonical authority/work/evidence:

```text
docs/decisions/discussion-notifications-launch.md
docs/decisions/notification-inbox-recognition-read.md
docs/work/current/t11-b08-notifications-full-inbox-r1.md
docs/work/current/t11-b08-p8-realization-plan.md
docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html
docs/work/current/t11-b08-screen-contract.md
docs/work/current/t11-b08-pattern-consolidation.md
```

Canonical P8 R1 HTML blob:

```text
bb130535721b2381524763a4885ade5199a15596
```

Locked B08 experience:

```text
Focused Triage Inbox
→ unseen + unread summaries
→ Caixa de entrada / Não lidas / Arquivadas
→ one canonical recency-ordered Notification list
→ human-recognizable current-disclosable source context
→ Nova distinct from Não lida
→ per-item read/unread + archive/unarchive
→ Mark all read
→ presentation-driven seen batching; fetch/cache != seen
→ exact source handoff to B03 Discussion anchor
→ cursor continuation/retry
→ access-drift neutral reconciliation
→ SSE invalidation/refetch only
→ Quick Inbox + Full Inbox share one Notifications authority
```

B08-F1 keeps op82 as the single Inbox list authority. Current census remains 86 operations / 11 routes / 16 PermissionCode values.

P9 closure:

```text
material B08 regions/controls traced         22 / 22
unbound material controls                    0
invented operations                          0
operation 87+                                absent
screen-shaped APIs                           0
frontend presentability/AuthZ evaluator      0
second Notification state store              0
source-workspace duplication                 0
SSE business-payload authority               0
material findings                            0
```

P10 closure:

```text
existing shared patterns reused              2
new shared semantic patterns graduated        0
B08-local patterns retained                   8
false abstractions                            0
Notifications/Minha Caixa semantic merges    0
source-workspace duplications                 0
```

Shared patterns reused are Global App Shell and Notification Quick Inbox. B08 does not graduate a generic Inbox, NotificationRow, Activity/Event feed, filter engine, deep-link resolver or realtime entity-store abstraction.

Not current B08 scope:

```text
search / free-form / saved filters
filter by author/Document
bulk selection/archive
snooze / priority / reminders
preferences / email / push
Notification delete
Notification-kind selector
source reply/editor/viewer inside Inbox
```

## Exact next action

```text
1. B09 Audit is the next eligible FP1 block and remains NOT OPEN.
2. Open B09 only when the operator chooses to continue FP1.
3. Do not open B10+ early.
4. Implementation remains blocked.
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
