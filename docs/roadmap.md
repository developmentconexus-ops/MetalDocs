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
T8-E EXECUTABLE WIRE CONTRACT         ACTIVE / DRAFT PR #136
T8-F → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-E — Executable Wire Contract, active in Draft PR #136 on branch `arch/t8e-wire-contract`.

Accepted work through the previous four T8-E layers is preserved at `docs/reference/t8e-checkpoint.md`. The current application census is 78 operations and is owned by `docs/product/journeys.md` plus `docs/decisions/api-operation-census.md`.

The active candidate now contains the closed 78-operation request/success/header/problem/filter/action ledger. Final promotion is blocked only by evidence-triggered bounded upstream reconciliation and the remaining executable/measurement proofs recorded in `docs/work/current/proposal.md`.

## Exact next action

```text
adjudicate the three bounded contradictions exposed by the executable ledger
→ T8-D Governance Step label + immutable label snapshot
→ T3 unreachable ProviderSubjectBinding-disabled Audit census entry
→ T8-C direct-PUT presign seam: exact expected bytes + max guard
→ reconcile only the implicated durable authorities after explicit operator approval
→ measure representative DOCX/PDF corpus and freeze raw/expanded/ZIP ceilings
→ run pinned Go + TypeScript generation/compile/type probe
→ prove exact request/response fixtures across all 78 operation rows
→ final Structural Inversion / YAGNI / overengineering / global-coherence pass
→ only then create isolated review/t8e-fable from exact candidate HEAD
→ Lead adjudication
→ explicit operator ratification
```

Do not reopen the accepted T8-E checkpoint or any completed T1→T8-D decision by preference. Only the three named material contradictions above are bounded reopen candidates; Product/T6 operation census remains closed at 78.

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
