---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE                       CLEAN-SLATE / ARCHITECTURE-FIRST
REPOSITORY RESET                      MERGED / OPERATOR-RATIFIED
REPOSITORY STANDARD V1 ALIGNMENT      MERGED
PRODUCT / OWNERSHIP                   OPERATOR-APPROVED
T1 → T8-D                             CLOSED / OPERATOR-RATIFIED
T8-E EXECUTABLE WIRE CONTRACT         ACTIVE / DRAFT PR PENDING
T8-F → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-E — Executable Wire Contract.

Accepted work through the previous four T8-E layers is preserved at `docs/reference/t8e-checkpoint.md`. The current application census is 78 operations and is owned by `docs/product/journeys.md` plus `docs/decisions/api-operation-census.md`.

## Exact next action

```text
freeze the remaining executable ledger for all 78 operations
→ exact request and success schemas
→ required / optional / nullable fields
→ closed enum vocabularies
→ success status + response-header matrix
→ operation-specific RFC 9457 problem-code matrix
→ list filters + deterministic ordering
→ allowed_actions vocabularies
→ request/body/document limits from evidence
→ Go + TypeScript generation feasibility
→ runtime contract-conformance proof design
→ subtractive / global-coherence pass
→ isolated final Fable review
→ Lead adjudication
→ explicit operator ratification
```

Do not reopen the accepted T8-E checkpoint by preference. A material contradiction follows the DevelopmentConexus Engineering Method reopen law.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | ACTIVE; exits by operator ratification |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | Opens after T8-E ratification |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/provider boundary, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | Opens after T8-F supplies its concrete runtime consumers; exits by ratification |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, and runtime realization as one system | Opens after T8-A→T8-G close; exits with no unresolved material contradiction |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes: contract, integration, concurrency, security, recovery, E2E, restore | Opens after Whole T8 coherence; exits by ratified validation baseline |
| T10 — Transition / Cutover | Any real current→target transition, data/schema transition, switch-over, rollback barriers, cutover readiness | Opens after T9 baseline is known; no compatibility mechanism survives without a named property and deletion condition |
| T11 — Implementation Program & Execution Graph | Bounded work graph, dependencies, touched authorities/contracts, proof obligations, rollback/replan triggers | Opens after T1→T10 accepted; may translate architecture but may not invent it |
| T12 — Adversarial Implementation-Readiness | Independent attack on ownership, dependencies, persistence, API/frontend/runtime parity, proof completeness, security/recovery, transition, and hidden Writer decisions | Opens after T11; exits only when material disagreement is closed or ratified |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED
T9  CLOSED / OPERATOR-RATIFIED
T10 CLOSED / OPERATOR-RATIFIED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, or hypothetical future capability are not reopen triggers.