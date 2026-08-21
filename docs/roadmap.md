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
T9 GOLDEN FLOWS / VALIDATION          OPEN / ACTIVE
T10 → T12                              NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T9 — Golden Flows & Validation Baseline is **OPEN / ACTIVE** as of 2026-08-21 by explicit operator authorization.

Opening proof:

```text
opening main                       82832cce62d11ea90575fb484b97e3c934c03e37
opening branch                     arch/t9-golden-flows
open PRs before T9                 0
last integrated closeout           PR #151 / MERGED
closeout merge-candidate CI        #1121 SUCCESS
T1 → T8-H                          CLOSED / OPERATOR-RATIFIED / INTEGRATED
application operations             78
operation 79                       ABSENT
T10 → T12                          NOT OPEN
Product implementation             BLOCKED
legacy implementation              ABSENT
```

The active branch-only Lead candidate is `docs/work/current/t9-golden-flows.md`. It is temporary Evidence/work, not durable architecture authority. It currently derives exactly **6 composed Golden Flows** and **10 cross-cutting validation properties** from the accepted T1→T8 authority, with causal falsifiers and real-dependency proof classes.

T9 must prove accepted composition rather than manufacture implementation. A material contradiction reopens only the smallest owning upstream authority; test convenience may not create operation 79, new Product state, new Permission, new semantic owner or new runtime capability.

## Ratified T8-H result

The Whole-T8 review closed three material coherence seams without reopening Product or any T1→T8-G semantic authority:

```text
H1  mutable program state
    -> docs/roadmap.md is the sole current stage/status/implementation/next-action authority

H2  executable DocumentOfficialView meaning
    -> complete effective schema/presence law consolidated in docs/architecture/wire-contract.md

H3  T5-J managed-content GC topology
    -> internal/application/maintenance reflected inside the existing non-semantic application class
```

Ratified technical candidate and independent proof:

```text
corrected technical candidate      b940d4e105a8b837ecdac7f71233ff10d735cd5e
candidate required CI              #1108 SUCCESS
Fable Round 1 PR                    #149 CLOSED / UNMERGED / 1 MATERIAL / adjudicated
Fable Round 2 PR                    #150 CLOSED / UNMERGED / CONVERGED / MATERIAL=0
Fable Round 2 final HEAD            5564612d07dc0325ac9b81e441f551340872e59d
Fable Round 2 CI                    #1110 SUCCESS
Round 3                             NOT JUSTIFIED
post-review status carrier          da0ffffc386a1335a866a9416cdcf7625de2ac02
status-carrier required CI          #1112 SUCCESS
operator ratification               EXPLICIT / 2026-08-21
closure candidate                   8c2ae8515fecf513cfd699e9d0e53eb2551fd835
closure required CI                 #1117 SUCCESS
final candidate                     213e3d7cb84130e282eec383b060577ca7580b48
final merge-candidate CI            #1119 SUCCESS
T8-H integration merge              d7f5d59f5dab6bc369483f88d44f69b9f2712c27
T8-H roadmap closeout/main          82832cce62d11ea90575fb484b97e3c934c03e37
```

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

The bounded T8-E-FR read-symmetry meaning remains executable only through the T8-E wire SSOT. T9 validates it; it does not create another wire/read authority.

## Exact next action

```text
complete the Lead T9 candidate on arch/t9-golden-flows
→ run required CI on the exact candidate HEAD
→ operator reviews/adjudicates the 6-flow / 10-property baseline
→ if operator accepts the Lead direction, open isolated Fable challenge from that exact candidate
→ adjudicate any material falsifier against the smallest owning authority
→ do not begin T10, T11 or T12
→ do not implement Product code
```

Cross-repository Marketplace review is complete and is not a T9 blocker. Candidate/review branch cleanup is non-authoritative housekeeping.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR meaning retained and executable representation consolidated in wire SSOT |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | OPEN / ACTIVE; Lead candidate under operator review path |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NOT OPEN; opens only after T9 is closed/operator-ratified/integrated |
| T11 — Implementation Program & Execution Graph | Bounded work graph and proof obligations | NOT OPEN; opens after T1→T10 accepted |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | NOT OPEN; opens after T11 |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED / INTEGRATED
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
