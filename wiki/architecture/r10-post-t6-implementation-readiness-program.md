# R10 Post-T6 Implementation Readiness Program

> **Status:** ACTIVE / OPERATOR-RATIFIED PROGRAM AUTHORITY  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Parent authority:** `wiki/architecture/r10-technical-architecture.md`  
> **Implementation:** BLOCKED

This program authority corrects the post-T6 stage decomposition after a Global Coherence finding proved that the prior sequence could reach implementation planning while the physical technical realization was still materially undefined.

It does **not** reopen Product Contract REV001 or T1→T6 by default. It preserves their accepted semantic/product decisions and inserts the missing realization, validation, transition and execution-planning layers required before code.

The operator ratified the restructuring direction on 2026-08-19.

---

## 1. Global Coherence finding

### Finding

The former post-T6 path was:

```text
T7 Historical Migration / Cutover
→ Whole-R10 Global Coherence Review
→ cold independent review
→ operator final ratification
→ implementation spec/plan
→ code
```

That path was insufficient because it could delegate material architecture decisions to implementation planning or Writers, including:

```text
backend module/package topology
allowed/forbidden dependency graph
internal owner-to-owner communication realization
physical persistence/table/constraint ownership
exact executable OpenAPI wire schemas
frontend feature/query/cache realization
runtime binaries/processes/jobs/deployment topology
operational readiness/observability/recovery realization
legacy code/schema/frontend disposition
cross-system Golden Flow proof design
implementation dependency/decomposition graph
```

These are material under the DevelopmentConexus Engineering Method because they affect boundaries, persistent meaning, contracts, security, concurrency/recovery, topology and proof.

### Root cause

T1→T6 correctly descended through semantic architecture, correctness, authorization, content, async/search and public API/frontend journeys, but the program treated **semantic design completion** as if it also implied **physical realization completion**.

Current technical documents cannot fill that gap by inheritance:

```text
backend-target-architecture.md = HISTORICAL prior target
data-model.md                  = current-state evidence; target physical model not designed
backend-api-structure.md       = current/legacy implementation structure evidence
frontend-structure.md          = current/legacy implementation structure evidence
repo-topology.md               = current runtime/repository evidence
```

Using those documents as silent target defaults would preserve the legacy local maximum.

### Method outcome

```text
RESTRUCTURE NOW
```

Scope of restructure:

```text
post-T6 stage decomposition
implementation-readiness definition
technical realization authority
legacy technical-document routing
```

Formal semantic reopen:

```text
Product Contract = NO
T1 = NO
T2 = NO
T3 = NO
T4 = NO
T5 = NO
T6 = NO
```

A later stage may reopen only the exact upstream decision materially contradicted by new evidence.

---

## 2. Program invariant

> **No implementation task may contain a material architecture decision that should have been decided by the architecture program.**

Corollaries:

1. A semantic owner is not assumed to equal one Go package, database schema, frontend feature or process.
2. Current packages/tables/routes/features/binaries receive no survival entitlement from existence.
3. A Writer may choose local tactics inside accepted realization, but may not choose ownership, dependency direction, persistent meaning, wire contract, trust boundary, recovery model or execution topology.
4. Target realization must remain independently understandable from legacy implementation shape.
5. Proof architecture precedes implementation decomposition.
6. Existing safety controls remain in force until a later accepted stage explicitly replaces them with an equal-or-stronger target property.

---

## 3. Corrected technical descent

```text
Product Contract REV001                          CLOSED / OPERATOR-APPROVED
T1 — Semantic State & Invariants                CLOSED / OPERATOR-RATIFIED
T2 — Governance / Effectivity / Transactions    CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit                      CLOSED / OPERATOR-RATIFIED
T4 — Exact Content / Storage / Restore          CLOSED / OPERATOR-RATIFIED
T5 — Durable Async / Search / Effects           CLOSED / OPERATOR-RATIFIED
T6 — Canonical API / Frontend Journeys          CLOSED / OPERATOR-RATIFIED

Post-T6 Stage-Decomposition GCR                 RESTRUCTURE NOW / OPERATOR-RATIFIED
Technical Realization Reconciliation Baseline  OPERATOR REVIEW NEXT

T7 — Historical Migration Truth & Mapping       NOT OPEN
T8 — Technical Realization Architecture         NOT OPEN
T9 — Golden Flows & Validation Baseline         NOT OPEN
T10 — Transition / Refactor / Migration/Cutover NOT OPEN
T11 — Implementation Program & Execution Graph  NOT OPEN
T12 — Adversarial Implementation-Readiness      NOT OPEN

implementation                                  BLOCKED
```

This sequence is a repository-specific specialization of the Method, not a second organizational method.

---

## 4. Prerequisite — Technical Realization Reconciliation Baseline

Before opening the redefined T7, the program must establish one reviewed census of current technical evidence and technical-authority status.

The baseline must cover:

```text
repository/runtime topology
backend modules/packages/import graph
persistence/table/SQL ownership
API/OpenAPI/codegen/runtime contract mechanisms
frontend routes/features/API/query/state topology
runtime async/jobs/rendering/processes
configuration/deploy/network/trust/operations
verification/tests/CI/tools/verify
durable vs stale/superseded technical documentation
```

Every claim is classified as:

```text
CURRENT-PROVEN
LAST-REPRODUCED
STALE / SUPERSEDED
UNKNOWN / REMEASURE
```

The baseline is **evidence and routing**, not target design. It must not choose the T8 architecture by stealth.

---

## 5. T7 — Historical Migration Truth & Semantic Mapping

T7 answers:

> **What source truth can be migrated honestly into the ratified target semantics, and how is that truth represented without fabricating native history?**

T7 owns only:

```text
actual source evidence census
PROVEN / INFERABLE WITH EXPLICIT RULE / UNKNOWN classification
smallest real migration-mode set
source→target semantic mapping
imported target-owned facts vs provenance-only evidence
revision/ordinal mapping
exact-content provenance
actor/governance provenance quality
semantic migration unit
truthful representation of partial/unknown historical evidence
```

T7 does **not** decide:

```text
final target tables/packages
migration SQL/scripts
production process topology
concrete deployment cutover
concrete rollback choreography
concrete restore orchestration
implementation decomposition
```

Those require T8/T10.

Reopen upstream only if real source evidence proves a ratified target semantic cannot represent a required truthful Launch migration without distortion.

---

## 6. T8 — Technical Realization Architecture

T8 converts T1→T7 semantics into one executable technical architecture. It is intentionally split to prevent a monolithic “technical design” from hiding contradictions.

### T8-A — Technical Authority & Legacy Census

Decide disposition of current technical structures:

```text
PRESERVE
REFINE
REHOME
REWRITE
DELETE
CURRENT-STATE ONLY
SUPERSEDED
```

Scope includes packages/modules, SQL/table access, OpenAPI/codegen, frontend features, binaries, deploy, verification and technical docs/ADRs.

### T8-B — Backend Module & Package Topology

Freeze:

```text
target repository/package layout
semantic-owner realization boundaries
layering within owners
public/internal Go package surfaces
allowed dependency graph
forbidden dependency graph
composition root / dependency injection
location of shared mechanisms
```

A semantic owner may realize through multiple cohesive packages. Package count follows isolation and clarity, not owner count or legacy module count.

### T8-C — Internal Communication Contracts

Freeze how accepted owners interact inside the realization:

```text
owner queries
owner capabilities
read projections
same-process/local calls
transaction-coupled intents
River/durable job seams where already justified
consumer/producer contract ownership
```

No private implementation import, cross-owner table access, foreign SQL or hidden shared write authority.

### T8-D — Persistence Realization

Freeze target persistent structure required for correctness:

```text
schemas/tables
material PK/FK/unique/check constraints
persistent state ownership
immutable/history shapes
WorkingContent/OCC realization
Submission/Release/effectivity constraints
Organization/AuthZ/ApplicationSession state
Audit evidence
managed-content technical state
idempotency replay state
River technical persistence boundary
transaction and serialization/lock mapping
```

Performance indexes that do not affect correctness may remain implementation tuning unless evidence makes them material.

### T8-E — Executable Wire Contract

Turn T6 semantic operations into exact OpenAPI contract authority:

```text
paths + operationIds
request/response schemas
fields/enums/nullability
headers
ETag / If-Match
Idempotency-Key
pagination cursors
RFC 9457 problem codes
upload/admission contract
exact-byte resources
generated Go boundary
generated TypeScript boundary
runtime validation/conformance
```

No implementation task may invent a missing wire field or enum.

### T8-F — Frontend Realization

Freeze:

```text
route tree
feature/package topology
public feature surfaces
generated transport consumption
TanStack Query ownership
query keys / invalidation / stale behavior
session/bootstrap/error behavior
local-vs-server state
DocumentWork/GovernanceCase/History/Admin read models
editor/viewer adapter boundaries
legacy feature disposition
```

The frontend remains a client, never a second lifecycle/AuthZ authority.

### T8-G — Runtime / Process / Deployment Realization

Freeze the smallest justified runtime topology:

```text
binaries/processes/services
River worker/job ownership
renderer/provider execution boundary
startup/readiness/shutdown
configuration/secrets
network/trust boundaries
timeouts
dependency degradation
logs/metrics/tracing
backup/restore runtime roles
dev/test/prod realization profiles
```

Legacy API/worker/jobs/renderer process count is evidence only.

### T8-H — Whole-T8 Global Coherence Review

Cross-check backend + persistence + wire + frontend + runtime together before T8 ratification.

---

## 7. T9 — Golden Flows & Validation Baseline

T9 answers:

> **What composed system behavior must be proven before implementation may be accepted?**

At minimum evaluate materially applicable flows such as:

```text
OIDC login → Library
blank create → DRAFT edit/autosave → submit → Release
create from Template
RETURN_FOR_CHANGES → edit → resubmit
exact immutable Submission governance
required OfficialRendition success/failure
replacement Revision/Release
obsolescence
User offboarding during authoring/governance
stale ETag race
Idempotency-Key uncertainty/replay
tampered/malicious upload / scanner outage
object-store/renderer outage
backup/restore readiness
historical migration
```

Each flow traces:

```text
frontend
→ API contract
→ semantic owner/use case
→ transaction/persistence
→ async/external mechanism where applicable
→ read model
→ final user-visible truth
```

Each material claim receives a falsifiable proof class, e.g. unit, schema, contract, integration, concurrency, restart/recovery, security, E2E or restore drill.

T9 becomes the implementation Validation Baseline.

---

## 8. T10 — Transition / Refactor / Migration / Cutover

T10 designs how the current product becomes the T8 target safely.

It owns:

```text
current→target code/package disposition
current→target table/schema disposition
current→target API route disposition
current→target frontend feature disposition
hard-cutover vs staged-transition choices
schema/data migration order
historical-data import execution mechanics
API/frontend/runtime switch-over
rollback windows and barriers
legacy deletion map
restore/offboarding security reconciliation choreography
deployment/cutover readiness
```

T10 may use transitional mechanisms only when they preserve a named property, have a successor and have a deletion condition.

No compatibility shim exists merely because current code exists.

---

## 9. T11 — Implementation Program & Execution Graph

T11 translates ratified architecture into bounded executable work. It may not decide architecture.

Planning layers:

```text
L0 Validation Baseline   = T9
L1 Realization Baseline  = T1→T10 accepted authority
L2 Execution Graph       = T11 bounded work/dependencies/proofs
L3 Tactical Plan         = adaptive Actor reasoning during execution
```

Every bounded implementation tranche states, where applicable:

```text
outcome
applicable authority
scope/non-goals
expected package/file loci or localization rule
interfaces/contracts touched
persistent changes
API/frontend/runtime effects
proof obligations
review/QA expectation
dependencies
rollback or Replan trigger
```

A task that asks a Writer to choose module topology, persistent authority, wire meaning, security boundary, runtime topology or migration semantics is invalid and returns to the owning architecture stage.

---

## 10. T12 — Adversarial Implementation-Readiness Review

T12 independently attacks the complete realization before implementation authorization.

Mandatory challenge surfaces:

```text
semantic owner → package mapping
allowed dependency graph
persistent ownership/constraints
API ↔ backend parity
API ↔ frontend parity
frontend ↔ journey parity
runtime/job/effect ownership
legacy disposition completeness
Golden Flow/proof completeness
security/trust boundaries
concurrency/recovery
migration/cutover/rollback
deploy/readiness/restore
documentation authority drift
future seams vs dormant implementation
Execution Graph hidden decisions
```

Use fresh/independent challenge for material authority/trust/topology decisions per the Method.

T12 closes only when material disagreement is empty or explicitly adjudicated/ratified.

---

## 11. Final implementation gate

Implementation remains blocked until all of:

```text
T7  CLOSED / OPERATOR-RATIFIED
T8  CLOSED / OPERATOR-RATIFIED
T9  CLOSED / OPERATOR-RATIFIED
T10 CLOSED / OPERATOR-RATIFIED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 GCR = PASS
fresh independent/cold review = converged
operator final implementation authorization = explicit
```

Only then may implementation execute the T11 graph.

---

## 12. Current exact next gate

```text
Technical Realization Reconciliation Baseline
→ operator reviews evidence classifications and coverage
→ correct any census gap/material misclassification
→ reconcile technical-document routing
→ only then open redefined T7
```

No substantive T7 design, implementation plan or product code is authorized yet.
