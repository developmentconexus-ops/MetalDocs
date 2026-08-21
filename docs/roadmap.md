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
T1 → T8-G                             CLOSED / OPERATOR-RATIFIED / INTEGRATED
T8-H WHOLE-T8 GLOBAL COHERENCE        NEXT / NOT STARTED
T9 → T12                              NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-G — Runtime / Process / Deployment is **CLOSED / OPERATOR-RATIFIED / INTEGRATED** as of 2026-08-21.

PR #144 was squash-merged into `main` as:

```text
1a5b3b5e0e05794175a724eed9d92802ee1bf705
```

The merge commit carries tree:

```text
87434c378aa622e2ef1c73c30f9e26e69cd1a55d
```

That tree is exactly the tree of the final authorized T8-G HEAD:

```text
01b1f021665a100fffc71d7387c5f4672e1323b9
```

The same authorized tree passed required CI **#1077** while Draft and required CI **#1078** after PR #144 was marked ready. The squash merge therefore integrated the exact authorized and verified tree.

Ratified durable authority:

```text
docs/architecture/runtime.md
+ docs/decisions/t8g-ratification.md
```

Independent review closure remains:

```text
Fable Round 1 PR #145  CLOSED / UNMERGED / F1 MATERIAL + F2-F4 MINOR / adjudicated
Fable Round 2 PR #146  CLOSED / UNMERGED / F1-F4 CLOSED / CONVERGED / 0 MATERIAL
Round 3                   NOT JUSTIFIED
```

The integrated T8-G Global Maximum remains:

```text
one modular-monolith application runtime
+ one PostgreSQL product-state database
+ River workers in the application process
+ one active ManagedContentStore
+ one private MalwareInspector
+ one private conditional DOCX→PDF renderer
+ external OIDC provider boundary
+ verified ephemeral exact-byte spool
+ fail-closed recovery profile
+ OpenTelemetry / OTLP observability baseline
+ one-shot migration / job / recovery operations
+ proven third-party mechanisms before local infrastructure
```

The integrated subtraction remains:

```text
Redis                             absent
separate worker deployment         absent
BFF / SSR                         absent
WebSocket/realtime                absent
external Search                   absent
custom scheduler/event bus         absent
service mesh                      absent
Kubernetes requirement             absent
custom telemetry framework         absent
custom queue/migration/OIDC/S3     absent
operation 79                       absent
Product implementation             BLOCKED
```

## Integrated T8-E / T8-F baseline

T8-E and T8-F remain **CLOSED / OPERATOR-RATIFIED / INTEGRATED**.

Application wire/frontend closure remains:

```text
accepted application operations      78
orphaned operations                  0
invented application operations      0
operation 79                         absent
Idempotency-Key creations            exact 10
ETag read / mutation domains         13 / 13
exact-byte resources                 exact 4
stable SPA route meanings            exact accepted T6 route set
frontend semantic owner added        none
frontend Authorization engine        absent
parallel global server store         absent
interactive DOCX runtime             one adapter boundary
```

The bounded T8-E-FR read-symmetry precision ratified with T8-F remains in force and changes none of those counts.

## Exact next action

```text
explicit operator authorization to open T8-H — Whole-T8 Global Coherence Review
→ start fresh from current `main` and revalidate its exact SHA at execution time
→ read AGENTS.md → docs/index.md → docs/roadmap.md → only the smallest T8-H authority pack required for a concrete cross-T8 check
→ cross-check T8-A→T8-G as one coherent realization without reopening accepted authority by preference
→ do not begin T9 and do not implement Product code
```

Candidate/review branch cleanup is non-authoritative housekeeping and does not open or block T8-H.

Do not reopen completed T1→T8-G or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR precision ratified with T8-F |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and their accepted upstream authorities as one system | NEXT / NOT STARTED; opens only on explicit operator authorization from current `main` |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | NOT OPEN; opens only after Whole-T8 coherence closes |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NOT OPEN; opens after T9 baseline |
| T11 — Implementation Program & Execution Graph | Bounded work graph and proof obligations | NOT OPEN; opens after T1→T10 accepted |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | NOT OPEN; opens after T11 |

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

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or generic infrastructure fashion are not reopen triggers.
