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

The active candidate contains the closed 78-operation request/success/header/problem/filter/action ledger. The two material upstream contradictions exposed by that ledger were operator-approved and reconciled on 2026-08-20:

```text
T8-D  Governance Step label + immutable attempt label snapshot
T3    unreachable ProviderSubjectBinding-disabled Audit census entry removed
```

Document-admission measurement/probes and executable generator/provider feasibility evidence are now complete enough for the Launch candidate:

```text
DOC_RAW_MAX_BYTES       100 MiB
DOCX_EXPANDED_MAX_BYTES 256 MiB
DOCX_MAX_ZIP_ENTRIES    4096
Go boundary             oapi-codegen v2.8.0 probe PASS
TypeScript boundary     openapi-typescript 7.13.0 probe PASS
direct S3 PUT           signed exact Content-Length + If-None-Match:* probe PASS
strict request split        kin-openapi + minimal envelope-guard probe PASS
```

The direct-PUT concern remains resolved subtractively without reopening T8-C: the existing `PresignCreate(handle,maxBytes,ttl)` seam is sufficient when T8-E supplies `maxBytes=expected_size_bytes`; the portable property is an at-most bound, while the reference S3 profile is stronger and signs exact `Content-Length`. Completion derives the actual descriptor independently, so no client-size truth is persisted.

The final Lead coherence attack exposed one consolidated bounded upstream package and the operator approved it on **2026-08-21**. The implicated T4/T5/T8-C/T8-D lines are now reconciled:

```text
T8-D               transaction/Audit + idempotency precision
T4/T5/T8-C/T8-D    no renderer/job for already-PDF required PDF rendition
T8-C/T8-D           reconstructible server-side CSRF synchronizer secret for GET /session
```

The correction remained subtractive/precision-only: no Product operation, owner, lifecycle state, permission, table family, generic worker, second cookie or new API was added.

## Exact next action

```text
rerun exact whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence delta
→ revalidate main/base + exact candidate HEAD + intended authority/work diff + required CI
→ if and only if the Lead delta converges, create isolated review/t8e-fable from exact candidate HEAD
→ independent Fable challenge
→ Lead adjudication
→ explicit operator ratification
```

Do not reopen the accepted T8-E checkpoint or completed T1→T8-D decisions by preference. Product/T6 operation census remains closed at 78; new material evidence reopens only the authority it actually implicates.

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
