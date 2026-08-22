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
T1 → T8-H                             CLOSED / OPERATOR-RATIFIED / INTEGRATED
T9 GOLDEN FLOWS / VALIDATION          CLOSED / OPERATOR-RATIFIED / INTEGRATION AUTHORIZED
T10 → T12                             NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T9 — Golden Flows & Validation Baseline is **CLOSED / OPERATOR-RATIFIED** as of 2026-08-21. The operator separately authorized squash integration of PR #154 after bounded independent Fable convergence.

The durable T9 authority is `architecture/validation-baseline.md`; immutable ratification evidence is `decisions/t9-ratification.md`. Temporary review/work Evidence is not authority and must not land in `main`.

Ratification proof:

```text
opening main                           82832cce62d11ea90575fb484b97e3c934c03e37
candidate PR                           #154
operator-approved Lead candidate       2d5d127e95821eac355296e0a7f09c93aef6cef3
Lead candidate required CI             #1127 SUCCESS
Round-1 Evidence PR                    #155 CLOSED / UNMERGED
Round-1 review CI                      #1128 SUCCESS
Round-1 verdict                        NOT CONVERGED / MATERIAL=2
technical correction commit            ca3a72d3f92eacea734bd1c583cd981e6e787bce
independently reviewed candidate HEAD  eb7e0147cf575fe69290c231ea360af229917eeb
corrected candidate required CI        #1130 SUCCESS
Round-2 Evidence PR                    #156 CLOSED / UNMERGED
Round-2 final review HEAD              27b7ce63a8c63169b6ac8b582ee49621e7c86355
Round-2 review CI                      #1132 SUCCESS
Round-2 verdict                        CONVERGED / MATERIAL=0
Round 3                                NOT JUSTIFIED
post-review status carrier             c5fba2b179e1e0a9a806df83654ea6daf6e67513
status-carrier required CI             #1133 SUCCESS
operator ratification                  EXPLICIT / 2026-08-21
merge authorization                    EXPLICIT / 2026-08-21
```

Round-1 adjudication closed two material proof gaps without reopening T1→T8:

```text
F1  runtime wire-conformance ownership
    → T9 V1 owns the future executable proof lane for accepted T8-E §9.4 fixture classes

F2  authentication-boundary falsification
    → GF1 causally attacks OIDC callback verification and ProviderSubjectBinding admission
```

Four bounded MINOR precisions also made session lifecycle revocation, managed-content GC/backup-pin race, closed T3 §15 Audit-census coverage and concurrent Document-code allocation explicit. Round 2 confirmed all corrections and found no MATERIAL regression. Its only MINOR was a safe-direction execution-mode sentence; durable promotion resolves it by distinguishing T9 baseline ratification from future runtime proof execution.

## Ratified T9 baseline

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

The six Golden Flows cover:

```text
GF1 identity / session / access / revocation
GF2 governance configuration → atomic Document creation
GF3 Revision authoring / upload admission / concurrency
GF4 governance → Release / OfficialRendition
GF5 official discovery / obsolescence / disclosure-safe routing
GF6 runtime failure isolation / shutdown / backup-restore readiness
```

The ten cross-cutting properties cover canonical wire/runtime conformance, architecture dependency closure, AuthN/AuthZ/CSRF, transaction+Audit atomicity, idempotency/replay, ETag/concurrency, exact-content/malware/rendition integrity, River durable work, runtime readiness/resources/observability and backup/restore privacy/security readiness.

T9 ratifies the **validation contract**, not an assertion that Product implementation already exists or has passed runtime tests. Later implementation/readiness work must execute or mechanically inspect the real production subject as required; mock-only, fixture-only and self-proving probes cannot close claims about real runtime/dependency behavior.

## Integrated T8 baseline preserved by T9

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

The bounded T8-E-FR read-symmetry meaning remains executable only through the T8-E wire SSOT. T9 validates that contract; it creates no second wire/read authority.

## Exact next action

```text
finish T9 closure candidate on PR #154
→ remove temporary docs/work/current/t9-golden-flows.md
→ required CI must pass on the exact closure HEAD
→ mark PR #154 Ready and obtain merge-candidate required CI on the same exact HEAD
→ squash merge PR #154 using expected HEAD protection
→ verify integrated tree / main
→ record T9 integration status without opening T10
```

T10 requires separate explicit operator authorization after T9 integration. T11/T12 remain closed. Product implementation remains blocked.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | CLOSED / OPERATOR-RATIFIED / INTEGRATION AUTHORIZED |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NOT OPEN; requires separate operator authorization after T9 integration |
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
