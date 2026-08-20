# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T8-D CLOSED / OPERATOR-RATIFIED; T8-E ACTIVE / EXECUTABLE WIRE CONTRACT; T8-F→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-20  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file is the **sole R10 current stage/status/next-action router**. Detailed target meaning lives in durable authorities. Staging/reviewer artifacts are provenance only after promotion.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T8-D durable R10 authorities
6. Decision Registry + amendments through T8-D
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. this router
10. active T8-E bootstrap listed in §7
11. current OpenAPI/generated/runtime/frontend code only for a concrete T8-E evidence/reuse claim

Legacy implementation proves what exists, not what survives.

## 2. Binding Method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
Structural Inversion
unknown remains unknown
revalidation does not mean reinvention
prepare the seam, not dormant future capability
```

Program law:

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

Ratified T8-A realization law:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Ratified T8-B topology law:

```text
ONE GO MODULE FOR BACKEND GO CODE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
```

Ratified T8-C contract law:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

Ratified T8-D persistence law:

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

## 3. Current descent

```text
Product Contract REV001                          CLOSED / OPERATOR-APPROVED
Whole-Product GCR A1→A10                         CLOSED / OPERATOR-APPROVED
Launch ownership topology                        CLOSED / OPERATOR-APPROVED / 4+1
T1 — Semantic State & Invariants                 CLOSED / OPERATOR-RATIFIED
T2 — Governance / Effectivity / Tx               CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit                       CLOSED / OPERATOR-RATIFIED + D4
T4 — Exact Content / Storage / Restore           CLOSED / OPERATOR-RATIFIED
T5 — Durable Async / Search / Effects            CLOSED / OPERATOR-RATIFIED
T6 — Canonical API / Frontend Journeys           CLOSED / OPERATOR-RATIFIED / PROMOTED
T7 — Historical Migration Truth & Mapping        CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-A — Technical Authority & Legacy Census       CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-B — Backend Module & Package Topology         CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-C — Internal Communication Contracts          CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-D — Persistence Realization                   CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                                CURRENT + amendments through T8-D
TRRB                                             CLOSED / OPERATOR-RATIFIED / PROMOTED

T8 — Technical Realization Architecture          ACTIVE
  T8-E Executable Wire Contract                  ACTIVE
  T8-F Frontend Realization                      NOT OPEN
  T8-G Runtime / Process / Deployment            NOT OPEN
  T8-H Whole-T8 Global Coherence Review          NOT OPEN

T9 — Golden Flows & Validation Baseline          NOT OPEN
T10 — Transition / Refactor / Migration/Cutover  NOT OPEN
T11 — Implementation Program & Execution Graph   NOT OPEN
T12 — Adversarial Implementation-Readiness       NOT OPEN

implementation                                    BLOCKED
```

## 4. T8-A / T8-B / T8-C closure

Durable authorities:

```text
wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md
wiki/architecture/r10-t8b-backend-module-package-topology.md
wiki/architecture/r10-t8c-internal-communication-contracts.md
```

Their amendments remain binding. T8-E may not redesign semantic ownership, package topology, internal contract placement or transaction participation by wire convenience.

## 5. T8-D — CLOSED / PROMOTED

Durable authority:

`wiki/architecture/r10-t8d-persistence-realization.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8d-amendment.md`

Status:

```text
T8-D Persistence Realization = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

Binding T8-D consequences carried into T8-E include:

```text
one PostgreSQL database
owner/mechanism schemas authn/org/authz/controlled_docs/audit/platform/river
owner-private fully-qualified SQL
PostgreSQL READ COMMITTED + narrow locks + explicit CAS
explicit BIGINT VersionToken and WorkingContent generation
Revision.state lifecycle + immutable Release effectivity fact
no persisted Document.current_status/current pointer authority
closed relational governance persistence and frozen candidate FK
immutable Submission/Decision/Release/Rendition/history grant classes
semantic exact descriptor ownership
immutable ManagedContent descriptor/malware proof relations
AdmissionClaim allocation-through-consume lifecycle
universal semantic attach ManagedContent FOR SHARE
GC ManagedContent FOR UPDATE + repeated non-locking reference proofs
paired Idempotency Key↔Replay deferred-completion invariant
HMAC fingerprint + drain-before-rotation equality law
River under river.*; self-REINDEX OFF on PG16; serving non-owner
four DB trust classes with runtime/verifier grant parity
materialized Search OFF
```

Independent-review convergence:

```text
Round 1: BLOCKER 2 / MAJOR 11 / LOW 10; Global Maximum CONFIRMED
Round 2: BLOCKER 0 / MAJOR 7 / LOW 6; both Round-1 blockers CLOSED
Final Lead adjudication: all Round-2 findings closed; material contradiction 0
T1→T7/T8-B/T8-C reopen: NO
third full review: NOT REQUIRED
operator ratification: explicit 2026-08-20
```

All live T8-D candidate/reviewer/adjudication artifacts are historical tombstones. Git history retains full provenance.

## 6. Stage boundaries

```text
T8-D = relational persistence / constraints / queries / locks              CLOSED
T8-E = exact executable OpenAPI/wire contract                              ACTIVE
T8-F = frontend realization                                                NOT OPEN
T8-G = runtime/process/deployment realization                              NOT OPEN
T8-H = Whole-T8 coherence                                                  NOT OPEN
T9   = Golden Flows + falsifiable Validation Baseline                      NOT OPEN
T10  = current→target technical transition/cutover/rollback/deletion        NOT OPEN
T11  = implementation Execution Graph                                      NOT OPEN
T12  = adversarial implementation-readiness                                NOT OPEN
```

T8-E may identify a real contradiction requiring a bounded reopen but may not silently change T1→T8-D meaning.

## 7. T8-E — ACTIVE / EXECUTABLE WIRE CONTRACT

Active bootstrap:

`docs/superpowers/analysis/2026-08-20-r10-t8e-executable-wire-contract-bootstrap.md`

T8-E answers:

> **What is the smallest exact OpenAPI 3.0.3 wire contract that realizes the already-ratified T6 semantic journeys and T8-D persistence/concurrency laws without inventing new state, leaking internal realization, leaving Writers to choose missing fields/enums/errors, or creating a second lifecycle/AuthZ authority?**

T8-E freezes:

```text
/api/v1 paths + operationIds
request/response schemas
fields/enums/nullability/requiredness
success statuses/bodies/headers
RFC 9457 Problem Details + application problem codes
strong ETag / If-Match representation
Idempotency-Key matrix
cursor pagination contract
session/CSRF application representation
upload allocation/completion/attachment contract
exact-byte resource responses
generated Go boundary
generated TypeScript boundary
runtime OpenAPI conformance
```

### Exact next action

```text
reconstruct complete T6 semantic operation census
→ map every operation to T8-D VersionToken/generation/idempotency/content consequences
→ inventory current OpenAPI/generated surfaces only as evidence
→ classify current paths/schemas/problem codes PRESERVE/REFINE/REWRITE/DELETE
→ derive exact success schema/header/status matrix
→ derive RFC 9457 problem-code matrix
→ derive pagination + ETag/If-Match + Idempotency-Key matrices
→ derive upload/exact-byte contract
→ prove generated Go/TypeScript/runtime-conformance feasibility
→ compare credible wire alternatives
→ apply Method + Structural Inversion + subtractive pass
→ independent Fable challenge
→ explicit operator ratification
```

No T8-F design, implementation planning, schema migration or product code is authorized.

## 8. Final implementation gate

Implementation remains blocked until:

```text
T8→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 GCR PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted equal-or-stronger target realization.
