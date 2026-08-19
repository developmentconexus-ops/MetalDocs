# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T8-C CLOSED / OPERATOR-RATIFIED; T8-D ACTIVE / PERSISTENCE REALIZATION; T8-E→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file is the **sole R10 current stage/status/next-action router**. Detailed meaning lives in the durable authorities it routes to.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T8-C durable R10 authorities
6. Decision Registry + D4/T6/post-T6/T7/T8-A/T8-B/T8-C amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. this router
10. active T8-D bootstrap listed in §7
11. current schema/SQL/code only for a concrete T8-D evidence/reuse claim

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

concrete semantic-owner APIs
+ real consumer-owned mechanism/resolver ports
+ application-routed cross-owner facts
+ database/sql-family shared txscope
+ owner-authored same-tx Audit evidence
+ Authorization sole final ALLOW/default-DENY
+ named transaction-coupled durable intents
+ self-contained PII-free ReplaySnapshot
+ bounded owner facts + application read composition
-
shared/common semantic contracts
-
generic UnitOfWork/EventBus/policy language/ServiceLocator
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
Decision Registry                                CURRENT + amendments through T8-C
TRRB                                             CLOSED / OPERATOR-RATIFIED / PROMOTED

T8 — Technical Realization Architecture          ACTIVE
  T8-D Persistence Realization                   ACTIVE
  T8-E Executable Wire Contract                  NOT OPEN
  T8-F Frontend Realization                      NOT OPEN
  T8-G Runtime / Process / Deployment            NOT OPEN
  T8-H Whole-T8 Global Coherence Review          NOT OPEN

T9 — Golden Flows & Validation Baseline          NOT OPEN
T10 — Transition / Refactor / Migration/Cutover  NOT OPEN
T11 — Implementation Program & Execution Graph   NOT OPEN
T12 — Adversarial Implementation-Readiness       NOT OPEN

implementation                                    BLOCKED
```

## 4. T8-A closure

Durable authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8a-amendment.md`

Binding consequences remain:

```text
current implementation/schema/API/frontend/runtime = evidence only
legacy package/schema/table/query shape = no inheritance entitlement
non-Launch implementation = DELETE / DEFER absent named Launch consumer
selective reuse requires all five T8-A proofs
```

## 5. T8-B closure

Durable authority:

`wiki/architecture/r10-t8b-backend-module-package-topology.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8b-amendment.md`

Binding consequences:

```text
semantic owner roots = authentication / organization / authorization /
                       controlleddocs / audit
one importable public surface per owner
owner-private decomposition ungated
transport → application = only semantic inbound door
application = stateless choreography
owner → owner imports forbidden
platform = mechanism only
composition = wiring only
first-party package/edge classification = closed-world/default-deny
foreign SQL / hidden shared semantic write authority forbidden
```

T8-C later refined the application class with `internal/application/maintenance` for T5-J GC choreography; this adds no new architecture class/owner/dependency direction and did not reopen T8-B.

## 6. T8-C closure

Durable authority:

`wiki/architecture/r10-t8c-internal-communication-contracts.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8c-amendment.md`

Status:

```text
T8-C Internal Communication Contracts = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

Independent review convergence:

```text
Round 1  Global Maximum class CONFIRMED
Round 2  BLOCKER 0 / surviving material contradiction 0
B5       Round-1 blocker NOT SUSTAINED
PII-free replay selection independently UPHELD
third Fable round NOT REQUIRED
```

Binding T8-C consequences carried into persistence:

```text
database/sql-family txscope
T2 READ COMMITTED posture remains binding
protected eligibility serialization semantics
owner VersionToken / expected-version contract
same-Scope Audit append
Authorization Decide/DecideMany + AuthorizedScopes
owner-private facts/queries only; no foreign SQL
ManagedContent create-once + AdmissionClaims + two-phase GC
named OfficialRendition River intent shares semantic commit
idempotency same-key concurrency must not poison Scope
ReplaySnapshot = versioned / self-contained / PII-free / snapshot-only reconstruction
```

T8-C staging/reviewer artifacts are historical provenance only after promotion; they are not status/target authority.

## 7. T8-D — ACTIVE / PERSISTENCE REALIZATION

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-bootstrap.md`

T8-D answers:

> **What is the smallest PostgreSQL persistence realization that makes T1→T8-C invariants and internal contracts structurally enforceable, assigns every persistent fact to its ratified owner/mechanism, and maps required ACID/OCC/serialization behavior to explicit schema/constraint/query/lock rules without foreign SQL, duplicate truth or speculative persistence?**

T8-D freezes:

```text
PostgreSQL schema namespace strategy
tables / persistent state ownership
material columns/types
PK/FK/unique/check/partial/exclusion constraints required for correctness
immutable/history relational shapes
owner-private SQL/query realization
WorkingContent OCC + owner VersionToken persistence
Submission/governance/Release/effectivity/obsolescence persistence
Organization/AuthZ/ApplicationSession persistence
Audit persistence/history visibility query
exact-content descriptor persistence
managed-content technical state + AdmissionClaims + GC_PENDING
idempotency claim/replay persistence + concurrency realization
River technical persistence boundary
canonical PostgreSQL Search/query/view realization where material
transaction/serialization/lock mapping + lock ordering
```

T8-D must deliberately classify each persistent family as:

```text
PERSIST — semantic owner
PERSIST — technical mechanism
STATIC / CODE AUTHORITY
DERIVED / QUERY-ONLY
DEFER / NOT LAUNCH
```

Current schema/migrations/SQL are evidence only and pass the T8-A five-part reuse gate before survival.

### Exact next action

```text
reconstruct complete persistent-state/invariant census from T1→T8-C
→ map every persistent fact to semantic owner or technical mechanism
→ classify PERSIST / STATIC / DERIVED / DEFER
→ derive correctness constraints before table convenience
→ derive transaction/serialization/lock matrix
→ derive owner-private query/persistence boundaries
→ compare credible PostgreSQL namespace/schema/version/history alternatives
→ remeasure current schema/query evidence only for concrete reuse claims
→ apply T8-A selective-reuse gate
→ apply Method + Structural Inversion + subtractive pass
→ adversarial challenge
→ operator-ratifiable T8-D candidate
```

No T8-D candidate is yet ratified.

## 8. Stage boundaries

```text
T8-D = relational persistence / constraints / queries / locks
T8-E = exact executable OpenAPI/wire contract
T8-F = frontend realization
T8-G = runtime/process/deployment realization
T8-H = Whole-T8 coherence
T9   = Golden Flows + falsifiable Validation Baseline
T10  = current→target technical transition/cutover/rollback/deletion
T11  = implementation Execution Graph
T12  = adversarial implementation-readiness
```

T8-D may not change semantic owners/contracts by schema convenience. It may identify a real contradiction and reopen only the exact upstream decision implicated.

T8-D may name a wire/runtime consequence but must not decide exact HTTP representation, frontend topology or process/deployment topology.

## 9. Final implementation gate

Implementation remains blocked until:

```text
T8→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 GCR PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted equal-or-stronger target realization.
