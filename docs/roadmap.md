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

Method: `docs/development/functional-html-wireframe-method.md` v2.3 — OPERATOR-RATIFIED on 2026-08-23.

Binding v2.3 law:

```text
NO screen-shaped backend
NO backend-shaped UX
current pre-implementation Product/backend plan = falsifiable baseline
material user need + insufficient authority = blocking UPSTREAM FINDING before P8
```

## FP1 block roadmap

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED
B03   Document Official                            LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
B04   Document Work                                LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
B05   My Work                                      LOCKED / OPERATOR-RATIFIED · P8 R2/P9/P10 COMPLETE
B06   Governance Case                              LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
       B06-F1 deadline projection                  CLOSED / OPERATOR-RATIFIED
       B06-F2 Governance Review Layer seam         CLOSED / OPERATOR-RATIFIED / FUTURE-SEAM
B07   Document History                             LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
       B07-F1 recognition read                     CLOSED / OPERATOR-RATIFIED
B08   Notifications Full Inbox                     LOCKED / OPERATOR-RATIFIED · P8/P9/P10 COMPLETE
       B08-F1 recognition read                     CLOSED / OPERATOR-RATIFIED
B09   Audit                                        OPEN / ACTIVE
       B09-F1 Audit query/evidence capability      OPEN / BLOCKING UPSTREAM FINDING
       Auditor point investigation                 LAUNCH CORE / OPERATOR-RATIFIED
       Auditor period + scope review               LAUNCH CORE / OPERATOR-RATIFIED
       Audit export                                 DEFERRED / OPERATOR-RATIFIED
       Structured Audit Query                      OPERATOR-RATIFIED
       Human recognition / historical labels      OPERATOR-RATIFIED
       Audit Query Assist                          OPERATOR-RATIFIED
       Owner-lens cross-link policy                NEXT DECISION
       P7                                          PAUSED pending B09-F1
       P8                                          BLOCKED pending B09-F1
B10   Organization Administration                  NOT OPEN
B11   Access Administration                        NOT OPEN
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

## Locked-block evidence routing

Detailed closed-block evidence remains in `docs/work/current/**` while this roadmap owns only mutable stage/gate state.

```text
B07 → docs/work/current/t11-b07-document-history-r1.md
B08 → docs/work/current/t11-b08-notifications-full-inbox-r1.md
B09 → docs/work/current/t11-b09-audit-upstream-replan.md
```

B07 History remains distinct from Audit. B08 keeps op82 as its single Inbox list authority. Neither closure changes the current system census.

## B09 current gate — Audit upstream replan

Current baseline:

```text
GET /api/v1/audit/events
listAuditEvents
AuditEventPage
occurred_at DESC, event_id DESC
cursor + limit only
audit.read
```

Ratified B09-F1 direction:

```text
Launch jobs
  point investigation / exact evidence question
  period + authorized historical-scope review
  export DEFERRED

Structured Audit Query
  occurred_at interval
  exact USER actor identity or SYSTEM
  one-or-more AuditOperationCode values
  resource_kind
  exact resource_id when known
  optional historical Area narrowing within audit.read authority
  fixed occurred_at DESC,event_id DESC

Human recognition
  immutable IDs/facts remain Audit evidence authority
  mutable labels are optional current/non-historical enrichment
  immutable human identifiers may be stable recognition
  filter identity remains IDs/enums, never mutable names

Audit Query Assist
  purpose-built discovery; no admin-directory dependency
  action = closed AuditOperationCode vocabulary
  Area = audit.read/historical-visibility bounded candidates
  actor = Audit-visible stable identities + optional current label
  resource = resource_kind-first Audit-visible candidate discovery
  no global entity/reference-data platform
```

Binding invariants:

```text
AuditEvent = semantic action evidence, not current state
Audit != Document History
historical visibility is snapshotted at action time
current authorization + historical visibility filter before pagination
Audit remains PII-minimized
browser never post-filters incomplete pages as complete evidence truth
```

## Exact next action

```text
1. Adjudicate B09-F1 owner-lens cross-link policy.
2. Ratify the smallest exact op78 + Audit Query Assist Product/API/wire refinement required.
3. Perform bounded FP0/B09 rebaseline if authority changes.
4. Only then resume B09 P7 and P8 functional HTML.
5. Do not open B10+ early.
6. Implementation remains blocked.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B09 P8 while B09-F1 is unresolved
no browser-side filtering of incomplete Audit pages as complete truth
no Audit-to-current-state reconstruction
no Audit/Document History semantic merge
no screen-shaped API by convenience
no backend-shaped UX suppression by current-plan inertia
no generic search/export platform without a proven consumer
no admin-only directory as required Audit filter infrastructure
no generic entity/reference-data platform for Audit selectors
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

Accepted Product/R10/frontend LOCK decisions reopen only on material evidence under the DevelopmentConexus Engineering Method. Frontend Method v2.3 treats stronger pre-implementation user/operator/reference evidence as a legitimate trigger to test and, when material, reopen the smallest owning Product/backend authority. Preference, sunk cost, framework availability or hypothetical scale alone are not reopen triggers.
