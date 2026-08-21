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
T1 → T8-E                             CLOSED / OPERATOR-RATIFIED
T8-F FRONTEND REALIZATION             NEXT / NOT STARTED
T8-G → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-E — Executable Wire Contract is **CLOSED / OPERATOR-RATIFIED / INTEGRATED** as of 2026-08-21.

PR #136 was squash-merged into `main` as `5568788d6322396f230db82e0cd0da027778f55e`. The merged commit carries tree `41b0742628d927c65b4e4a841c125b33ed2fedca`, exactly matching the final ratified candidate tree that passed required CI #1032. The absorbed `arch/t8e-wire-contract` branch and the earlier `review/t8e-fable` Evidence branch have both been deleted.

T8-F — Frontend Realization is **NEXT / NOT STARTED**. This closeout does not open T8-F and does not authorize implementation.

The reconciled T8-E contract is the ratified durable authority at `docs/architecture/wire-contract.md`. The application census remains 78 operations and is owned by `docs/product/journeys.md` plus `docs/decisions/api-operation-census.md`.

The ratified authority contains the closed 78-operation request/success/header/problem/filter/action ledger. The two material upstream contradictions exposed by that ledger were operator-approved and reconciled on 2026-08-20:

```text
T8-D  Governance Step label + immutable attempt label snapshot
T3    unreachable ProviderSubjectBinding-disabled Audit census entry removed
```

Document-admission measurement/probes and executable generator/provider feasibility evidence are part of the ratified Launch baseline:

```text
DOC_RAW_MAX_BYTES       100 MiB
DOCX_EXPANDED_MAX_BYTES 256 MiB
DOCX_MAX_ZIP_ENTRIES    4096
Go boundary             oapi-codegen v2.8.0 probe PASS
TypeScript boundary     openapi-typescript 7.13.0 probe PASS
direct S3 PUT           signed exact Content-Length + If-None-Match:* probe PASS
strict request split    kin-openapi + minimal envelope-guard probe PASS
```

The direct-PUT concern remains resolved subtractively without reopening T8-C: the existing `PresignCreate(handle,maxBytes,ttl)` seam is sufficient when T8-E supplies `maxBytes=expected_size_bytes`; the portable property is an at-most bound, while the reference S3 profile is stronger and signs exact `Content-Length`. Completion derives the actual descriptor independently, so no client-size truth is persisted.

The final Lead coherence attack exposed one consolidated bounded upstream package and the operator approved it on **2026-08-21**. The implicated T4/T5/T8-C/T8-D lines are reconciled:

```text
T8-D               transaction/Audit + idempotency precision
T4/T5/T8-C/T8-D    no renderer/job for already-PDF required PDF rendition
T8-C/T8-D           reconstructible server-side CSRF synchronizer secret for GET /session
```

The correction remained subtractive/precision-only: no Product operation, owner, lifecycle state, permission, table family, generic worker, second cookie or new API was added.

Independent Fable review PR #137 challenged candidate `ef329534fc9d5df3254d59c3787197fefa8435e6`. Lead adjudication accepted the material promotion/presence/bounds/profile findings, restored the current T6 Problem namespace rather than reopening it, and preserved T3 configuration Audit via closed typed facts rather than deleting auditability. No material contradiction survived to justify Round 2. PR #137 is closed and unmerged; its Evidence branch was deleted after adjudication.

The ratified T8-E authority absorbed the former checkpoint/work contract; `DOC-12` is consumed and the router/register point at the durable wire authority. Final verification evidence:

```text
ledger rows                         78 / exact 1..78
operationId                         78 unique
method + path                       78 unique
Idempotency-Key creations           exact 10
ETag read / mutation domains        13 / 13
exact-byte resources                exact 4
Audit operation codes               37 unique
Problem namespace                   https://errors.metaldocs.io/{code}
ShortText / LongText                256 / 4096
attention_required                  absent
PROFILE_REPLACE If-Match+absent     412 precondition.resource_changed
rows 3 / 45 validation.failed       absent
row 42 validation.failed            present
forward obligations                 21 / 3 / 27 = 51
docs/work/**                         absent
required CI                          #1032 SUCCESS on exact merged tree
main merge                           5568788d6322396f230db82e0cd0da027778f55e
```

Operator ratification occurred explicitly on **2026-08-21**.

## Exact next action

```text
explicit operator authorization to open T8-F
→ start fresh from current main @ 5568788d6322396f230db82e0cd0da027778f55e
→ read AGENTS.md → docs/index.md → docs/roadmap.md → only the bounded T8-F authority pack routed from there
→ derive the smallest Frontend Realization contract
→ do not open T8-G and do not implement Product code
```

Do not reopen completed T1→T8-E or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | NEXT / NOT STARTED; opens only on explicit operator authorization from current `main` |
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
