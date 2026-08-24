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
METHODOLOGY ADOPTION  CANDIDATE / R1 CONVERGED
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

## Methodology adoption

The exact organizational methodology pin and selection route are owned by `AGENTS.md`. This increment adopts that AGENTS-owned pin; selection starts at its pinned `ROUTER.md`. Frontend planning selects `METHOD.md + FRONTEND-METHOD.md`; repository operation selects `REPOSITORY-STANDARD.md`.

The central `FRONTEND-METHOD.md` explicitly consolidates the reusable MetalDocs lineage through local v2.3. Therefore the local reusable `docs/development/functional-html-wireframe-method.md` is superseded and removed; MetalDocs-specific P8 HTML Evidence handling remains in `docs/development/engineering-rules.md`.

Aggregate verification proves the concrete bootstrap properties it encodes: the methodology repository + exact SHA + `ROUTER.md` are present in `AGENTS.md`, a moving-main bootstrap without that pin is rejected, the superseded known frontend-method file stays absent, and active routers do not point to it. It does **not** claim exhaustive semantic detection of every possible renamed methodology fork or simultaneous floating-main reference; those broader duplicate-authority defects remain explicit hard stops and adversarial-review concerns.

Impact sweep:

```text
B01-B09 protected structure / Screen Contracts  UNAFFECTED
Product/backend authority                       UNAFFECTED
89-operation / 11-route census                  UNAFFECTED
exact P8 LOCK Evidence                           UNAFFECTED
FP2 / P11                                        NOT OPEN
B10-B12                                          NOT OPEN
```

No Product/backend authority changed and no current durable Evidence identifies a material OPEN assumption that falsifies an existing LOCK. Stronger central small-delta, assumption, LOCK-impact, and P11-fidelity rules govern future affected work; they do not by themselves reopen valid B01-B09 LOCKs.

Independent methodology-adoption review:

```text
PR #169        CLOSED / UNMERGED
R1             CONVERGED
MATERIAL open  0
IMPORTANT open 0
MINOR          2 / adjudicated by bounded roadmap precision only
```

R1-F1 was refined by narrowing the proof claim to what CI actually detects; no brittle catch-all grep machinery was added. R1-F2 was accepted by removing the duplicate SHA from this roadmap: `AGENTS.md` remains the bootstrap pin owner.

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
1. Run fresh exact-HEAD Draft CI after the R1 minor adjudication.
2. If green, mark PR #168 Ready for Review and require fresh non-Draft `required` CI.
3. Confirm no open review conversation and no Product/frontend/architecture delta.
4. Stop for explicit operator squash-merge authorization.
5. After integration, recover current authority before opening B10+ in a separate increment.
6. T11 remains OPEN; T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 or B10-B12 work in this increment
no B01-B09 reopen without material Evidence
no local reusable methodology fork
no floating methodology main as normative authority
no methodology sync bot/submodule/generated copies/framework
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

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the pinned DevelopmentConexus Engineering Method; methodology-migration ceremony alone is not a reopen trigger.
