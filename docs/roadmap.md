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
METHODOLOGY ADOPTION  ACCEPTED / INTEGRATED
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`.

Current system census:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           89
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

`docs/decisions/api-operation-census.md` owns the current application-operation / Idempotency / ETag / exact-byte census.

## Organizational methodology baseline

The exact accepted organizational methodology pin and method-selection route are owned by `AGENTS.md`. Method selection starts at that pin's `ROUTER.md`.

The reusable local Frontend Product Experience Planning Method v2.3 was superseded by the central `FRONTEND-METHOD.md`, whose lineage explicitly consolidates the MetalDocs v2.3 generation. MetalDocs-specific P8 rendered-Evidence handling remains local in `docs/development/engineering-rules.md`.

Adoption impact sweep:

```text
B01-B09 protected structure / Screen Contracts  UNAFFECTED
Product/backend authority                       UNAFFECTED
89-operation / 11-route census                  UNAFFECTED
exact P8 LOCK Evidence                           UNAFFECTED
FP2 / P11                                        NOT OPEN
B10-B12                                          NOT OPEN
```

The independent methodology-adoption review converged with `MATERIAL=0` and `IMPORTANT=0`. Methodology adoption did not reopen Product/architecture/frontend LOCKs and did not authorize implementation.

Aggregate verification enforces the concrete bootstrap properties it owns, including the exact AGENTS methodology pin/ROUTER presence, context budget, temporary-work hygiene, known local-method removal, active-router de-reference, and existing exact Evidence-ref protections. Broader semantic duplicate-authority defects remain hard stops/review concerns rather than falsely claimed as exhaustively grep-detectable.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / 89 operations / 11 routes REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

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

Exact locked B01-B09 P8 identities remain recoverable through:

```text
docs/decisions/t11-b01-b09-lock-evidence.md

evidence/t11-pr162-b01-b09-locks-20260824
→ adf58e448bc5bd3a20cae5b7228d729c031f94ac
```

The Evidence ref remains exact-SHA checked and non-authoritative.

## Exact next action

```text
1. Revalidate accepted main and recover authority through AGENTS → docs/index → docs/roadmap → pinned ROUTER.
2. Select METHOD + FRONTEND-METHOD for the next frontend-planning increment.
3. Revalidate the smallest authority pack and forward obligations before deciding whether B10 is ready to open.
4. Do not treat B10 as opened merely because methodology adoption is integrated.
5. T11 remains OPEN; B10-B12 remain NOT OPEN until explicitly opened by later authority.
6. FP2/P11 remains NOT OPEN until the block program reaches its accepted assembly boundary.
7. T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B10-B12 work until a later explicit acceptance increment opens it
no B01-B09 reopen without material Evidence
no local reusable methodology fork
no floating methodology main as normative authority
no methodology sync bot/submodule/generated copies/framework
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

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the pinned DevelopmentConexus Engineering Method; methodology-migration ceremony alone is not a reopen trigger.
