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
METHODOLOGY ADOPTION  ACTIVE / BOUNDED REPOSITORY-GOVERNANCE INCREMENT
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

## Organizational methodology adoption

Target organizational methodology authority:

```text
developmentconexus-ops/conexus-methodology
@ 9c7210d1504bef01c0d134a6c3ae8627deebb535
```

Selection begins at the pinned `ROUTER.md`. Relevant current organizational methods include:

```text
METHOD.md                 DevelopmentConexus Engineering Method v1.1.0
REPOSITORY-STANDARD.md    DevelopmentConexus Repository Standard v1.1.0
FRONTEND-METHOD.md        DevelopmentConexus Frontend Product Experience Method v1.0.0
```

The organizational `FRONTEND-METHOD.md` explicitly consolidates the reusable MetalDocs frontend-method lineage through the operator-ratified local v2.3 generation. Therefore `docs/development/functional-html-wireframe-method.md` is superseded reusable authority and is removed in this increment rather than maintained as a parallel fork.

MetalDocs-specific frontend Evidence handling remains in `docs/development/engineering-rules.md`.

## Methodology-adoption impact boundary

This increment changes methodology ownership/routing, not Product/backend/frontend structure.

```text
B01-B09 protected Product/UX structure      UNAFFECTED
B01-B09 Screen Contracts / LOCK meaning      UNAFFECTED
Product/architecture authorities             UNAFFECTED
89-operation / 11-route census               UNAFFECTED
Evidence ref / exact P8 LOCK blobs           UNAFFECTED
FP2 / P11                                    NOT OPEN
B10-B12                                      NOT OPEN
```

Basis:

- the central frontend method declares the MetalDocs v2.3 lineage as an input to its consolidation;
- no Product/backend authority changes in this adoption;
- no current durable repository authority identifies a material OPEN frontend assumption that falsifies an existing LOCK;
- strengthened central rules for future small-delta handling, lock-impact sweeps, lock-time assumptions, and P11 fidelity apply prospectively to the next affected work/reopen/assembly and do not by themselves falsify prior LOCK structure.

If future Evidence identifies a specific prior LOCK whose protected structure depends on a still-OPEN material assumption, apply the pinned frontend method's smallest-scope `REVALIDATE`/`REOPEN` law then; do not preemptively restart B01-B09.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / 89 operations / 11 routes REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

Frontend planning now selects the organizational `METHOD.md + FRONTEND-METHOD.md` from the exact methodology pin in `AGENTS.md`; there is no repository-local reusable frontend-method authority.

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

The Evidence ref is exact-SHA pinned by the repository aggregate verification and is non-authoritative. `docs/work/**` remains absent from accepted main/merge candidates except temporary lawful Draft/review transport.

## Exact next action

```text
1. Complete the bounded methodology-adoption candidate only; do not open B10+.
2. Prove AGENTS exposes the exact accepted methodology SHA and routes through pinned ROUTER.md.
3. Prove repository-local reusable frontend methodology has been removed without losing MetalDocs-specific controls.
4. Run the repository aggregate verification and prove its methodology-pin/duplicate-authority negative guards can fire.
5. Perform a fresh independent adversarial review using the pinned ADVERSARIAL-REVIEW-METHOD.md because this increment changes governing methodology consumption.
6. Resolve any MATERIAL/IMPORTANT finding against the smallest owning authority.
7. Require fresh Ready/non-Draft `required` CI and explicit operator squash-merge authorization.
8. Only after this adoption is integrated, recover current authority again before explicitly opening B10+ in a separate acceptance increment.
9. T11 remains OPEN; T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B10/B11/B12 work in this methodology-adoption increment
no B01-B09 reopen without material Evidence
no local reusable methodology fork
no floating methodology main as normative authority
no methodology sync bot/submodule/generated copies/framework
no Audit-to-current-state reconstruction
no Audit/Document History semantic merge
no screen-shaped API by convenience
no backend-shaped UX suppression by current-plan inertia
no generic search/export/reference/deep-link platform without a proven consumer
no frontend Authorization matrix
no merge authorization implied
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

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the pinned DevelopmentConexus Engineering Method. Preference, sunk cost, framework availability, methodology migration ceremony, or hypothetical futures alone are not reopen triggers.
