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
docs/decisions/audit-investigation-read.md
```

Current authoritative system census:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           89
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

`docs/decisions/api-operation-census.md` is the sole numeric census.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / R2 89/11 REBASELINED
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
       B09-F1 Audit query/evidence capability      CLOSED / OPERATOR-RATIFIED
       Structured Audit Query                      OPERATOR-RATIFIED
       Human recognition / historical labels      OPERATOR-RATIFIED
       Audit Query Assist                          OPERATOR-RATIFIED
       Owner-lens cross-link policy                OPERATOR-RATIFIED
       Exact op78 + op87-op89 package              OPERATOR-RATIFIED / DURABLE
       P7 H1                                       OPERATOR-APPROVED IN CHAT
       P7 written candidate                        REVIEW REQUIRED
       P8                                          BLOCKED pending written P7 ratification
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

## B09 bounded authority and proof

```text
durable authority  docs/decisions/audit-investigation-read.md
numeric authority  docs/decisions/api-operation-census.md
finding ledger     docs/work/current/t11-b09-audit-upstream-replan.md
R2 evidence        docs/work/current/t11-b09-audit-capability-decision-candidate-r2.md
rebaseline proof   docs/work/current/t11-b09-f1-rebaseline-proof.md
P7 candidate       docs/work/current/t11-b09-audit-r1.md
```

Ratified Audit read surface:

```text
78  listAuditEvents                REFINED
87  listAuditQueryAreas            SAFE_READ
88  searchAuditQueryActors         SAFE_READ
89  searchAuditQueryResources      SAFE_READ
```

Binding B09 laws:

```text
AuditEvent = immutable semantic action evidence, not current state
Audit != Document History
historical visibility is snapshotted at action time
current authorization + historical visibility filter before pagination
structured query filters server-side complete evidence
mutable recognition labels are optional current/non-historical context
filter identity remains stable IDs/enums
Audit Query Assist is purpose-built and Audit-visible-only
Audit-native same actor/resource/action investigation is universal
owner links are secondary and destination rechecks current AuthZ/disclosure
```

Explicit YAGNI:

```text
free-text generic Audit search   DEFERRED
query DSL                        REJECTED
saved searches                   DEFERRED
custom sort                      REJECTED
analytics/dashboard              REJECTED as B09 responsibility
export                           DEFERRED
admin-directory selector dependency REJECTED
generic entity/reference-data/deep-link resolver REJECTED
```

B01-B08 remain preserved; no bounded rebaseline contradiction was found.

P7 candidate result:

```text
leading hypothesis    Audit Investigation Ledger
upstream findings     0 unresolved
P8 eligibility        pending operator ratification of written candidate
```

## Exact next action

```text
1. Operator reviews docs/work/current/t11-b09-audit-r1.md as the written B09 P7 exit candidate.
2. If changes are requested, revise P7 before any P8 work.
3. If the written candidate is operator-ratified, record clean P7 exit and only then begin P8 functional low-fidelity HTML.
4. Any new material Product/backend insufficiency becomes a new upstream FINDING under Method v2.3.
5. Do not open B10+ early.
6. Implementation remains blocked.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B09 P8 until written P7 ratification and clean exit
no browser-side filtering of incomplete Audit pages as complete truth
no Audit-to-current-state reconstruction
no Audit/Document History semantic merge
no screen-shaped API by convenience
no backend-shaped UX suppression by current-plan inertia
no generic search/export platform without a proven consumer
no admin-only directory as required Audit filter infrastructure
no generic entity/reference-data/deep-link resolver
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
