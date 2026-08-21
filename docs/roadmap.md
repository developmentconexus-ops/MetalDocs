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
T9                                     NEXT / NOT STARTED
T10 → T12                              NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-H — Whole-T8 Global Coherence Review is **CLOSED / OPERATOR-RATIFIED / INTEGRATED** as of 2026-08-21.

Integrated proof:

```text
opening main                       0b4ef6ef891b01f907804cff4bd3c0022aebad80
final candidate PR                 #148
final authorized candidate HEAD    213e3d7cb84130e282eec383b060577ca7580b48
merge-candidate required CI        #1119 SUCCESS
squash merge                       d7f5d59f5dab6bc369483f88d44f69b9f2712c27
integrated main                    d7f5d59f5dab6bc369483f88d44f69b9f2712c27
candidate tree                     a536233b634abec0874d2d488e470f057be2b759
integrated main tree               a536233b634abec0874d2d488e470f057be2b759
T8-H ratification                  EXPLICIT / 2026-08-21
T8-H integration                   VERIFIED
T9                                 NEXT / NOT STARTED
Product implementation             BLOCKED
```

The squash merge integrated the exact authorized tree: the final candidate HEAD and `main @ d7f5d59f...` both resolve to tree `a536233b634abec0874d2d488e470f057be2b759`.

The immutable T8-H ratification record is `docs/decisions/t8h-ratification.md`. It owns ratification evidence only; current progression, implementation permission and exact next action remain here.

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
integrated main                     d7f5d59f5dab6bc369483f88d44f69b9f2712c27
```

The post-review status carrier changed only roadmap plus the temporary T8-H ledger; it changed no independently reviewed technical authority. The temporary ledger was removed before merge candidacy. CI #1116 correctly rejected the first closure attempt because the new durable T8-H ratification record was not yet reachable; the decision register now routes it explicitly and #1117 passed.

## Integrated T8 baseline preserved by T8-H

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

The bounded T8-E-FR read-symmetry meaning remains unchanged and is executable only through the T8-E wire SSOT. The T8-H CI proportionality correction keeps material repository-envelope/routing/provenance protections blocking, keeps valid isolated Draft Evidence green, treats Markdown whitespace as warning-only, and keeps leftover merge-conflict markers blocking.

## Exact next action

```text
operator authorization to open T9 — Golden Flows & Validation Baseline
→ if authorized, start fresh from integrated main @ d7f5d59f5dab6bc369483f88d44f69b9f2712c27
→ derive the smallest falsifiable composed-system validation contract from accepted T1→T8 authority
→ do not begin T10, T11 or T12
→ do not implement Product code
```

No additional T8-H Fable round is justified. The two Round-2 MINOR safe-direction `Implementation remains BLOCKED` echoes are explicitly non-blocking and do not authorize implementation or contradict this roadmap.

Candidate/review branch cleanup is non-authoritative housekeeping and does not open T9.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR meaning retained and executable representation consolidated in wire SSOT |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | NEXT / NOT STARTED; requires explicit operator authorization to open |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NOT OPEN; opens after T9 baseline |
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
