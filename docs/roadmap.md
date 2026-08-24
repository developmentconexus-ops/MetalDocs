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
T11 checkpoint        B01-B09 ACCEPTED / INTEGRATED
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`.

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

`docs/decisions/api-operation-census.md` is the sole current application-operation / Idempotency / ETag / exact-byte census authority.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / 89 operations / 11 routes REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

Current local frontend-planning method: `docs/development/functional-html-wireframe-method.md` v2.3, operator-ratified on 2026-08-23.

## FP1 block status

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED
B03   Document Official                            LOCKED / P8-P10 COMPLETE
B04   Document Work                                LOCKED / P8-P10 COMPLETE
B05   My Work                                      LOCKED / P8 R2-P10 COMPLETE
B06   Governance Case                              LOCKED / P8-P10 COMPLETE
       B06-F1 deadline projection                  CLOSED / OPERATOR-RATIFIED
       B06-F2 Governance Review Layer seam         CLOSED / FUTURE-SEAM
B07   Document History                             LOCKED / P8-P10 COMPLETE
       B07-F1 recognition read                     CLOSED / OPERATOR-RATIFIED
B08   Notifications Full Inbox                     LOCKED / P8-P10 COMPLETE
       B08-F1 recognition read                     CLOSED / OPERATOR-RATIFIED
B09   Audit                                        LOCKED / P8-P10 COMPLETE
       B09-F1 Audit query/evidence capability      CLOSED / OPERATOR-RATIFIED
       op78 + op87-op89 package                    OPERATOR-RATIFIED / DURABLE
       unresolved BLOCKING / IMPORTANT             0 / 0
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

## B09 closure

Durable authority:

```text
docs/decisions/audit-investigation-read.md
docs/decisions/api-operation-census.md
```

Ratified Audit read family:

```text
78  listAuditEvents                REFINED
87  listAuditQueryAreas            SAFE_READ
88  searchAuditQueryActors         SAFE_READ
89  searchAuditQueryResources      SAFE_READ
```

Binding laws remain:

```text
AuditEvent = immutable semantic action evidence, not current state
Audit != Document History
historical visibility is snapshotted at action time
current authorization + historical visibility filter before pagination
structured query filters server-side complete evidence
mutable recognition labels are optional current/non-historical context
filter identity remains stable IDs/enums
Audit Query Assist is purpose-built and Audit-visible-only
owner links are secondary and destination rechecks current AuthZ/disclosure
```

Closure proof:

```text
P7 upstream findings       0 unresolved
Fable unresolved           0 BLOCKING / 0 IMPORTANT
P8 artifact blob           7daa6054851e617aeacb95a28d907d0d6d4bd3d6
P8 static proof            PASS
P8 internal behavior       79 / 79 PASS
P8 operator disposition    LOCK / APPROVED
P9 material controls       33 / 33 traced / 0 unbound
P9 invented operations     0 / operation 90+ absent
P10 new shared patterns    0
P10 false abstractions     0
```

Exact locked B01-B09 P8 identities and the pre-cleanup temporary planning tree are preserved by:

```text
docs/decisions/t11-b01-b09-lock-evidence.md

evidence/t11-pr162-b01-b09-locks-20260824
→ adf58e448bc5bd3a20cae5b7228d729c031f94ac
```

The Evidence ref is exact-SHA pinned by the repository aggregate verification and is non-authoritative. `docs/work/**` remains absent from accepted main.

## B01-B09 checkpoint result

The accepted checkpoint contains only durable Product/architecture decisions, repository governance/CI, the current local frontend-planning method, routing/status authority, and the exact LOCK Evidence locator.

This checkpoint does **not** close T11. It establishes a clean accepted base from which later T11 increments continue.

Independent whole-checkpoint adversarial review converged with:

```text
MATERIAL open   0
IMPORTANT open  0
```

## Exact next action

```text
1. Revalidate accepted main after the B01-B09 checkpoint integration.
2. Adopt the accepted central DevelopmentConexus methodology in a separate repository-governance acceptance increment:
   developmentconexus-ops/conexus-methodology
   @ 9c7210d1504bef01c0d134a6c3ae8627deebb535
3. Do not mix that methodology adoption with Product/frontend/architecture progression.
4. After the methodology adoption is independently accepted/integrated, recover current authority again before opening B10+.
5. T11 remains OPEN; B10-B12 remain NOT OPEN until explicitly opened by later authority.
6. T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B10/B11/B12 work until a later explicit acceptance increment opens it
no B01-B09 reopen without material Evidence
no Audit-to-current-state reconstruction
no Audit/Document History semantic merge
no screen-shaped API by convenience
no backend-shaped UX suppression by current-plan inertia
no generic search/export/reference/deep-link platform without a proven consumer
no frontend Authorization matrix
```

## Implementation gate

Implementation remains blocked until:

```text
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the DevelopmentConexus Engineering Method. Preference, sunk cost, framework availability or hypothetical scale alone are not reopen triggers.
