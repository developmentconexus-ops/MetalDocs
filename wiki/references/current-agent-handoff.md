# Current Agent Handoff

> **Last verified:** 2026-08-20  
> **Status:** ACTIVE — **T1→T8-D CLOSED / OPERATOR-RATIFIED; T8-E ACTIVE / EXECUTABLE WIRE CONTRACT; T8-F→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/r10-technical-architecture.md` — sole status/next-action router
5. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
6. T1→T8-D durable authorities
7. Decision Registry + amendments through T8-D
8. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
9. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
10. `docs/superpowers/analysis/2026-08-20-r10-t8e-executable-wire-contract-bootstrap.md`
11. current OpenAPI/generated/runtime/frontend code only when a concrete T8-E evidence/reuse claim needs it

Do not route target design through superseded/historical architecture or T8-D staging tombstones.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-D                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + amendments through T8-D
Post-T6 Stage-Decomposition GCR          OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-E                                     ACTIVE / EXECUTABLE WIRE CONTRACT
T8-F                                     NOT OPEN
T8-G                                     NOT OPEN
T8-H                                     NOT OPEN
T9                                       NOT OPEN
T10                                      NOT OPEN
T11                                      NOT OPEN
T12                                      NOT OPEN
implementation                           BLOCKED
```

## Binding execution law

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

## Reference-backed decision law

```text
MetalDocs semantic/product authority
→ DevelopmentConexus Engineering Method
→ current repository evidence
→ primary/current standards and official tool/library documentation
→ relevant reference patterns as falsification evidence
→ credible alternatives + Global Maximum
→ proof/adversarial review
→ operator ratification
```

External practice never creates Product requirements.

## T7 — CLOSED

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current business DB/content/history is DEV/test/throwaway and creates no historical-business compatibility entitlement.

## T8-A — CLOSED

Durable authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Ratified law:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

## T8-B — CLOSED

Durable authority:

`wiki/architecture/r10-t8b-backend-module-package-topology.md`

Ratified law:

```text
ONE GO MODULE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
```

Semantic homes remain exactly Authentication / Organization / Authorization / Controlled Documents / Audit.

## T8-C — CLOSED

Durable authority:

`wiki/architecture/r10-t8c-internal-communication-contracts.md`

Ratified law:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

Key carry-forward laws:

```text
database/sql-family Scope
application owns local transaction lifecycle
owner-private SQL only
same-Scope Audit
Authorization sole final ALLOW/default-DENY
protected Organization eligibility semantics
owner VersionToken
ManagedContent/AdmissionClaims/River named seams
self-contained PII-free ReplaySnapshot
no generic EventBus/UnitOfWork/ServiceLocator/policy language
```

## T8-D — CLOSED / PROMOTED

Durable authority:

`wiki/architecture/r10-t8d-persistence-realization.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8d-amendment.md`

Ratified Global Maximum:

```text
OWNER-NAMESPACED POSTGRESQL RELATIONAL CORE
+ DECLARATIVE CORRECTNESS
+ PRIVILEGE-ENFORCED IMMUTABLE HISTORY
+ READ COMMITTED NARROW SERIALIZATION
+ EXPLICIT CAS
+ IDENTITY-ONLY CROSS-OWNER REFERENTIAL INTEGRITY
+ TRANSACTIONAL KEY↔REPLAY COMPLETION
+ THIRD-PARTY RIVER SCHEMA ISOLATION
+ PROOF-BACKED SELECTIVE LEGACY PROPERTY REUSE
- LEGACY PHYSICAL SHAPE INHERITANCE
- GENERIC PERSISTENCE FRAMEWORKS
- DUPLICATE CURRENT TRUTH
```

Key binding persistence laws consumed by T8-E:

```text
schemas authn/org/authz/controlled_docs/audit/platform/river
PostgreSQL READ COMMITTED + narrow row locks
explicit BIGINT VersionToken / WorkingContent generation
Revision.state + immutable Release
no Document current-status/current-revision pointer
closed relational governance + frozen candidate FK
immutable semantic history through grants
semantic descriptor truth vs immutable ManagedContent descriptor proof
AdmissionClaim spans allocation through consume/release/expiry
semantic content attach FOR SHARE; GC FOR UPDATE
Idempotency Key↔Replay completed-commit invariant
HMAC semantic fingerprint remains internal
River under river.* / self-REINDEX OFF on PG16 / runtime non-owner
four DB trust classes with runtime/verifier grant parity
materialized Search OFF
```

Independent-review convergence:

```text
Round 1: BLOCKER 2 / MAJOR 11 / LOW 10; Global Maximum CONFIRMED
Round 2: BLOCKER 0 / MAJOR 7 / LOW 6; both blockers CLOSED
Final Lead: every Round-2 finding closed; surviving contradiction 0
third full review NOT REQUIRED
operator ratified 2026-08-20
```

All T8-D staging/reviewer/adjudication files are historical tombstones. Use Git history only for provenance.

## T8-E — ACTIVE

Active bootstrap:

`docs/superpowers/analysis/2026-08-20-r10-t8e-executable-wire-contract-bootstrap.md`

T8-E exact question:

> **What is the smallest exact OpenAPI 3.0.3 wire contract that realizes the already-ratified T6 semantic journeys and T8-D persistence/concurrency laws without inventing new state, leaking internal realization, leaving Writers to choose missing fields/enums/errors, or creating a second lifecycle/AuthZ authority?**

### T8-E owns

```text
/api/v1 paths + operationIds
request/response schemas
fields/enums/nullability/requiredness
success status/body/header matrix
RFC 9457 Problem Details + problem-code vocabulary
strong ETag/If-Match wire representation
Idempotency-Key matrix
cursor pagination contract
session/CSRF application representation
upload allocation/completion/attachment wire
exact-byte resources
generated Go + TypeScript boundary
runtime OpenAPI conformance
```

### T8-E does not own

```text
semantic lifecycle/owners                    T1→T6
persistence/constraints/locks                T8-D
frontend realization                         T8-F
runtime/process/deploy                       T8-G
Golden Flow proof architecture               T9
current→target cutover                       T10
implementation task graph                    T11
```

### Exact next action

```text
reconstruct complete T6 semantic operation census
→ map every operation to T8-D VersionToken/generation/idempotency/content consequences
→ inspect current OpenAPI/generated code only as evidence
→ classify current wire PRESERVE/REFINE/REWRITE/DELETE
→ derive exact success schema/status/header matrix
→ derive RFC 9457 problem-code matrix
→ derive pagination / ETag / If-Match / Idempotency-Key matrices
→ derive upload/exact-byte contract
→ prove generated Go/TypeScript/runtime-conformance feasibility
→ compare wire alternatives
→ Method + Structural Inversion + subtractive pass
→ independent Fable review
→ explicit operator ratification
```

Implementation remains **BLOCKED**.
