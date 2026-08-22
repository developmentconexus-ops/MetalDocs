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
T10                                   NEXT / NOT STARTED
T11 → T12                             NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T9 — Golden Flows & Validation Baseline is **CLOSED / OPERATOR-RATIFIED / INTEGRATED** as of 2026-08-21.

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

The squash merge integrated the exact authorized tree: final candidate `e8ee5f9e...` and `main @ 29c0c87c...` both resolve to tree `3e0f9d494ea577310e632633c17dfd621f75bf1e`.

The durable T9 technical authority is `architecture/validation-baseline.md`. The immutable ratification snapshot is `decisions/t9-ratification.md`. Current progression, implementation permission and exact next action remain here.

## Ratified T9 result

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
T1→T8 reopen                         none
Product implementation               BLOCKED
```

The six Golden Flows cover identity/session/access/revocation; governance configuration→atomic Document creation; Revision authoring/upload/concurrency; governance→Release/OfficialRendition; official discovery/obsolescence/disclosure-safe routing; and runtime failure/shutdown/backup-restore readiness.

The ten cross-cutting properties cover wire/runtime conformance, architecture dependency closure, AuthN/AuthZ/CSRF, transaction+Audit atomicity, idempotency/replay, ETag/concurrency, exact-content/malware/rendition integrity, River durable work, runtime readiness/resources/observability and backup/restore privacy/security readiness.

T9 ratifies a **falsifiable validation contract**, not an assertion that Product implementation already exists or has passed runtime tests. Later execution must target real production subjects/boundaries; mock-only, fixture-only and self-proving probes cannot close real runtime/dependency claims.

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
operator authorization to open T10 — Transition / Cutover
→ if authorized, start fresh from the then-current integrated `main` containing this T9 closeout; revalidate its exact SHA before work
→ derive the smallest truthful current→target transition and rollback-barrier contract from accepted T1→T9 authority
→ do not begin T11 or T12
→ do not implement Product code
```

T10 is **NEXT / NOT STARTED** and is not open without separate explicit operator authorization.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NEXT / NOT STARTED; requires explicit operator authorization to open |
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
