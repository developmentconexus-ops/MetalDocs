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
T1 → T9                               CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10                                   OPEN / ACTIVE
T11 → T12                             NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T10 — Transition / Cutover is **OPEN / ACTIVE** as of 2026-08-22 by explicit operator authorization.

Opening proof:

```text
opening main                           fc7030e98021bdb55fa806df68821cf19ed1a40c
opening branch                         arch/t10-transition-cutover
open PRs before T10                    0
T1 → T9                                CLOSED / OPERATOR-RATIFIED / INTEGRATED
application operations                 78
operation 79                           ABSENT
T11 → T12                              NOT OPEN
Product implementation                 BLOCKED
legacy implementation                  ABSENT FROM LIVE TREE
```

The active branch-only Lead candidate is `docs/work/current/t10-transition-cutover.md`. It is temporary work/Evidence, not durable architecture authority.

T10 must derive the smallest truthful technical activation/cutover/recovery contract from accepted T1→T9 authority. T7 already establishes that Launch has no historical business corpus to migrate; T10 may not manufacture ETL, dual-write, compatibility or old/new authority merely for transition convenience.

The current Lead direction is one-way greenfield activation with five monotonic barriers:

```text
B0  source truth classified
B1  target privately prepared
B2  target proven before normal serving
B3  canonical serving authority activated
B4  first authoritative R10 Product mutation committed
```

Before B4, activation may be reversed only while no authoritative R10 Product mutation has committed. After B4, destructive return to disposable pre-R10/DEV/test state is forbidden; incidents become R10 recovery.

## Integrated T9 result preserved by T10

T9 — Golden Flows & Validation Baseline is **CLOSED / OPERATOR-RATIFIED / INTEGRATED**.

Integrated proof:

```text
opening main                           82832cce62d11ea90575fb484b97e3c934c03e37
candidate PR                           #154
operator-approved Lead candidate       2d5d127e95821eac355296e0a7f09c93aef6cef3
Lead candidate required CI             #1127 SUCCESS
Round-1 Evidence PR                    #155 CLOSED / UNMERGED
Round-1 review CI                      #1128 SUCCESS
Round-1 verdict                        NOT CONVERGED / MATERIAL=2
independently reviewed candidate HEAD  eb7e0147cf575fe69290c231ea360af229917eeb
corrected candidate required CI        #1130 SUCCESS
Round-2 Evidence PR                    #156 CLOSED / UNMERGED
Round-2 final review HEAD              27b7ce63a8c63169b6ac8b582ee49621e7c86355
Round-2 review CI                      #1132 SUCCESS
Round-2 verdict                        CONVERGED / MATERIAL=0
Round 3                                NOT JUSTIFIED
operator ratification                  EXPLICIT / 2026-08-21
merge authorization                    EXPLICIT / 2026-08-21
final authorized candidate HEAD        e8ee5f9e12cd9a933cd732b12549c7e48a42be52
Draft required CI                      #1146 SUCCESS
merge-candidate required CI            #1147 SUCCESS
candidate tree                         3e0f9d494ea577310e632633c17dfd621f75bf1e
squash merge / integrated main         29c0c87c3f659ce889b4210d487ee89a43d43d55
integrated main tree                   3e0f9d494ea577310e632633c17dfd621f75bf1e
T9 integration                         VERIFIED
```

The durable T9 technical authority is `architecture/validation-baseline.md`. The immutable ratification snapshot is `decisions/t9-ratification.md`.

Ratified T9 envelope preserved by T10:

```text
Golden Flows                         exactly 6
cross-cutting validation properties exactly 10
evidence classes                     exactly 6
application operations               exactly 78
orphaned operations                  0
invented application operations      0
operation 79                         absent
new Permission                       none
new semantic owner                   none
Product implementation               BLOCKED
```

## T10 source truth

T7 remains binding:

```text
pre-R10 MetalDocs business history   NONE
required pre-R10 business corpus     NONE
historical business migration        NOT REQUIRED
DEV/test state preservation          REJECTED
```

A surviving external DEV/test database, managed-content store, OIDC client, deployment, ingress or secret/config resource receives no business-authority status merely because it exists. T10 inventories such technical estate only to make activation/cleanup truthful.

If concrete evidence proves that real pre-R10 business truth or a required compatibility consumer exists, T10 stops and routes a bounded reopen to the smallest owning authority before migration design proceeds.

## Preserved integrated baseline

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
one modular-monolith application runtime
one PostgreSQL product-state database
River workers in-process
one active ManagedContentStore
private conditional renderer + MalwareInspector
verified ephemeral exact-byte spool
Redis / BFF / realtime / external Search / generic event bus absent
Product implementation               BLOCKED
```

## Exact next action

```text
complete the Lead T10 candidate on arch/t10-transition-cutover
→ run required CI on the exact candidate HEAD
→ operator reviews/adjudicates the one-way activation + B0→B4 rollback-barrier contract
→ if operator accepts the Lead direction, open isolated Fable challenge from that exact candidate
→ adjudicate any material falsifier against the smallest owning authority
→ do not begin T11 or T12
→ do not implement Product code
```

No T10 implementation plan is authorized or created while Product implementation remains blocked. T10 owns transition architecture only.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | OPEN / ACTIVE; Lead candidate under operator-review path |
| T11 — Implementation Program & Execution Graph | Bounded work graph and proof obligations | NOT OPEN; opens after T1→T10 accepted |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | NOT OPEN; opens after T11 |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED / INTEGRATED
T9  CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10 CLOSED / OPERATOR-RATIFIED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or generic infrastructure fashion are not reopen triggers.
