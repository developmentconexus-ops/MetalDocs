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
T11 current block     B11 ACCESS ADMINISTRATION / OPEN / ACTIVE / P8 R5 CANDIDATE / CONTENT WALKTHROUGH
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
B11   Access Administration                        OPEN / ACTIVE / P8 R5 CANDIDATE / CONTENT WALKTHROUGH
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
       P7 R2 written disposition                    OPERATOR-APPROVED
       P8 R2 plan                                   docs/work/current/t11-b11-access-administration-p8-r2-plan.md
       P8 R2 artifact                               docs/work/current/t11-b11-access-administration-p8-r2.html
       P8 R2 Git blob                               8940238a5234c81ba0651693a0fe36c1da7f10a5
       P8 R2 operator disposition                   REVISE / VISUAL TOPOLOGY REGRESSION
       P8 R2 topology continuity                    4 / 13 PASS
       P8 R2 revise evidence                        docs/work/current/t11-b11-p8-r2-operator-revise.md
       P8 R3 artifact                               docs/work/current/t11-b11-access-administration-p8-r3.html
       P8 R3 Git blob                               273b9d16f23c37663024a5a0ec638d08e52002f0
       P8 R3 topology continuity                    13 / 13 PASS against wrong B11/B10-derived baseline
       P8 R3 functional browser verification        42 / 42 PASS
       P8 R3 operator disposition                   REVISE / LOCKED-WIREFRAME FIDELITY FAILURE
       P8 R3 revise evidence                        docs/work/current/t11-b11-p8-r3-operator-revise.md
       canonical wireframe fidelity baseline        B01 / B01N / B07 / B08 exact LOCK Evidence
       P8 R4 artifact                               docs/work/current/t11-b11-access-administration-p8-r4.html
       P8 R4 Git blob                               a9c4ef8f558771848079ce74e221be40bc5c2ad4
       P8 R4 LOCKED-wireframe fidelity               30 / 30 PASS
       P8 R4 functional/browser verification        34 / 34 PASS
       P8 R4 combined verification                  64 / 64 PASS
       P8 R4 frame fidelity                         OPERATOR-APPROVED
       P8 R4 content disposition                    REVISE / LOCAL P8 CONTENT FINDINGS
       P8 R4 content findings                       docs/work/current/t11-b11-p8-r4-content-findings.md
       P8 R5 artifact                               docs/work/current/t11-b11-access-administration-p8-r5.html
       P8 R5 Git blob                               96094773435a88c357e308779639415d9853b327
       P8 R5 frame                                  INHERITED FROM OPERATOR-APPROVED R4 / NO REDESIGN
       P8 R5 corrections                            USER PICKER + DYNAMIC MEMBERSHIP REVIEW + REMOVE CONFIRM + CONTEXTUAL GRANT
       P8 R5 operator disposition                   AWAITING CONTENT WALKTHROUGH
       B11-R2-A Group footprint comprehension       OPEN / FALSIFIABLE DURING CONTENT WALKTHROUGH
       B11-R2-B Area configuration comprehension    OPEN / FALSIFIABLE DURING CONTENT WALKTHROUGH
       B11-R2-C membership consequence              OPEN / FALSIFIABLE DURING CONTENT WALKTHROUGH
       B11-R2-D remaining effective-access gap      OPEN / FALSIFIABLE DURING CONTENT WALKTHROUGH
       B11-R2-E filtered pagination sufficiency     OPEN / FALSIFIABLE DURING CONTENT WALKTHROUGH
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
1. Operator operates the exact B11 P8 R5 artifact at docs/work/current/t11-b11-access-administration-p8-r5.html. The R4 frame/wireframe fidelity is already OPERATOR-APPROVED and is inherited; do not redesign it.
2. Test B11-R2-A in Grupos using `Aprovadores Financeiro`: confirm different Roles across Financeiro, Comercial, Company and Jurídico are immediately understandable without implying one Group-owned Area.
3. Test B11-R2-B in Por Área using `Comercial`: confirm Area-specific grants and Company-wide grants that also apply remain visibly separate and truthful.
4. Test B11-R2-C using the R5 membership flow: choose an existing User through the paginated picker, review exact User + Group consequence, add membership, and separately confirm a removal; judge whether visible Group footprint + bounded copy is sufficient.
5. Test contextual grant entry from both Area and Group lenses: Area preselects its real scope; Group preselects its exact subject; Role/remaining dimensions remain deliberate and the final Subject × Role × Scope review remains mandatory.
6. Test B11-R2-D while direct and Group grants coexist: judge whether Launch access administration remains safe without a per-User effective-access troubleshooter.
7. Test B11-R2-E by paging Area-specific, Company, Group, Role and membership-User slices; judge whether filtered pagination is operationally sufficient without global matrix/search capability.
8. Exercise Funções read-only meaning, role scope restrictions, exact revoke, ambiguous retry with same logical command/Idempotency-Key, membership conflict and 403/404/409/failure fixtures.
9. If disposition REVISE, revise only the smallest affected P8 content scope unless operator Evidence exposes a new semantic insufficiency; then route the smallest UPSTREAM FINDING.
10. Only the operator may LOCK B11. Only after explicit B11 LOCK run P9 bidirectional trace and P10 bounded pattern consolidation.
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
no Group single-Area ownership inferred from access scope
no custom Role/Permission editor
no P8 R6 before explicit operator disposition of P8 R5 content
no unapproved shell/sidebar/header/local-nav topology redesign inside a block P8
no production-like visual design in P8
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
