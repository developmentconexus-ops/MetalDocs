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
T11 checkpoint        B01-B10 ACCEPTED / INTEGRATED
T11 current block     B11 ACCESS ADMINISTRATION / OPEN / ACTIVE / P7 R2 CANDIDATE / OPERATOR REVIEW
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

Adoption impact sweep at integration time:

```text
B01-B09 protected structure / Screen Contracts  UNAFFECTED
Product/backend authority                       UNAFFECTED
89-operation / 11-route census                  UNAFFECTED
exact P8 LOCK Evidence                           UNAFFECTED
FP2 / P11                                        NOT OPEN
B10-B12                                          NOT OPEN at methodology-adoption integration
```

The independent methodology-adoption review converged with `MATERIAL=0` and `IMPORTANT=0`. Methodology adoption did not reopen Product/architecture/frontend LOCKs and did not authorize implementation.

Aggregate verification enforces the concrete bootstrap properties it owns, including the exact AGENTS methodology pin/ROUTER presence, context budget, temporary-work hygiene, known local-method removal, active-router de-reference, and exact Evidence-ref protections. Broader semantic duplicate-authority defects remain hard stops/review concerns rather than falsely claimed as exhaustively grep-detectable.

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
B10   Organization Administration                  LOCKED / P8-P10 COMPLETE / OPERATOR-RATIFIED / INTEGRATED
       locked P8 blob                              1d1cc7d5cb42e034ab9ee71a21c96918cdcf691d
       structural browser verification             32 / 32 PASS
       P9 material regions/controls                 34 / 34 TRACED
       P9 accepted B10 operations                   24 / 24 BOUND (ops 3-26)
       operation 27+ consumed                       0
       B10-A1 paginated-browse sufficiency          VALIDATED FOR CURRENT LAUNCH P8
       unresolved material B10 Findings             0
B11   Access Administration                        OPEN / ACTIVE / P7 R2 CANDIDATE / OPERATOR REVIEW
       P6 reference study                           COMPLETE / OFFICIAL SOURCES
       P7 R1 written spec                           docs/work/current/t11-b11-access-administration-p6-p7.md
       P7 R1 disposition                            SUPERSEDED ONLY WHERE B11-F1 CHANGES IA
       P8 R1 artifact                               docs/work/current/t11-b11-access-administration-p8.html
       P8 R1 Git blob                               c04ff56efa7aae72c59dc0ee9c4d56c9357c4de7
       P8 R1 structural verification                40 / 40 PASS
       P8 R1 operator disposition                   REVISE / UPSTREAM FINDING
       B11-A1 findability / inspectability          FAIL
       B11-A2 membership consequence                FAIL
       B11-A3 access explanation                    REFINED / FULL ENGINE NOT PROVEN
       B11-F1 Access Assignment Read Precision      CLOSED / OPERATOR-RATIFIED / DURABLE
       B11-F1 authority                             docs/decisions/access-assignment-read.md
       op31 listRoleAssignments                     REFINED / FILTERED + HUMAN-RECOGNIZABLE READ
       application-operation delta                  +0 / CENSUS REMAINS 89
       unresolved blocking upstream Findings        0 after B11-F1 reconciliation
       P7 R2 written spec                           docs/work/current/t11-b11-access-administration-p7-r2.md
       P7 R2 leading IA                             POR ÁREA / GRUPOS / FUNÇÕES
       P7 R2 written disposition                    AWAITING OPERATOR REVIEW
       P8 R2                                        NOT STARTED
B12   Document Governance Administration           NOT OPEN
```

Exact locked frontend Evidence remains recoverable through:

```text
B01-B09
docs/decisions/t11-b01-b09-lock-evidence.md
evidence/t11-pr162-b01-b09-locks-20260824
→ adf58e448bc5bd3a20cae5b7228d729c031f94ac

B10
docs/decisions/t11-b10-lock-evidence.md
evidence/t11-pr170-b10-locks-20260824
→ b8c607cbd30d61d6bcf6ec1ea734ed1653d2569e
```

Both Evidence refs are exact-SHA checked and non-authoritative. B10 `docs/work/**` is absent from integrated `main` after exact Evidence preservation.

## Exact next action

```text
1. Operator reviews the exact written B11 P7 R2 at docs/work/current/t11-b11-access-administration-p7-r2.md.
2. Do not begin P8 R2 until the operator explicitly approves that written P7 R2.
3. P7 R2 leading structure is:
     Acessos
     ├── Por Área
     ├── Grupos
     └── Funções
4. The Group lens must expose one Group's canonical RoleAssignments across Company and multiple Areas; never add Group.area_id.
5. The Area lens must show Area-scoped grants and Company-scoped grants in separate truthful regions; do not merge them into fake Area-owned assignments.
6. The Funções lens uses fixed read-only RoleView meaning; no custom Role/Permission editor.
7. P8 R2 may use only op31 precision in docs/decisions/access-assignment-read.md plus accepted supporting Organization reads. Global User/Group/Area search and per-User effective-access explanation remain unproven/not authorized.
8. After written P7 R2 approval, P8 R2 must actively test Group access footprint, Area/Company grant visibility, membership consequence, filtered pagination sufficiency and whether the remaining lack of per-User effective-access explanation is acceptable.
9. If P8 R2 exposes another material insufficiency, STOP only the affected B11 scope and route the smallest UPSTREAM FINDING. Only the operator may LOCK.
10. After explicit B11 LOCK, run P9 bidirectional trace and P10 bounded pattern consolidation.
11. docs/work/** must be absent from the eventual B11 merge candidate/main after exact Evidence preservation.
12. B12 and FP2/P11 remain NOT OPEN. T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B12 work
no FP2/P11 work
no B01-B10 reopen without material Evidence
no application operation 90+ for B11 without a new lawful bounded reopen
no global User/Group/Area search invented for UI convenience
no client-side crawl/post-filter over incomplete pages presented as complete
no frontend effective-permission authority or inferred inherited-access matrix
no Group.area_id or equivalent single-Area ownership inferred from access scope
no custom Role/Permission editor
no assistant/reviewer LOCK
no local reusable methodology fork
no floating methodology main as normative authority
no methodology sync bot/submodule/generated copies/framework
no merge without explicit operator authorization
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
