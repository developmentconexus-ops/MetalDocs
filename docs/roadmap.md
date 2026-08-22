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
T9 GOLDEN FLOWS / VALIDATION          OPEN / ACTIVE — REVIEW CONVERGED / RATIFICATION PENDING
T10 → T12                             NOT OPEN
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

The active branch-only T9 candidate is `docs/work/current/t9-golden-flows.md`. It is temporary work/Evidence, not durable architecture authority. It preserves exactly **6 composed Golden Flows**, **10 cross-cutting validation properties**, **6 evidence classes**, the **78-operation** application census and operation 79 absent.

## T9 review state

The operator explicitly approved the Lead candidate direction at:

```text
approved Lead candidate            2d5d127e95821eac355296e0a7f09c93aef6cef3
approved candidate required CI     #1127 SUCCESS
```

Independent Fable Round 1 reviewed that exact candidate through isolated Draft PR #155:

```text
review PR                           #155 CLOSED / UNMERGED
review branch                       review/t9-fable
review HEAD                         47483960e596539c69dc32139eb069dcc696694f
review CI                           #1128 SUCCESS
review delta                        docs/work/current/ai-dialog.md only
verdict                             NOT CONVERGED
MATERIAL findings                   2
Round 2 justified                   YES
```

Lead adjudication accepted both MATERIAL findings and four bounded MINOR precisions without reopening T1→T8:

```text
F1 MATERIAL  ACCEPT — V1 owns runtime execution of T8-E §9.4 wire-conformance classes
F2 MATERIAL  ACCEPT — GF1 attacks the OIDC callback/binding admission boundary
F3 MINOR     ACCEPT — session expiry/endSession/binding-replacement revocation explicit
F4 MINOR     ACCEPT — managed-content GC + backup-pin/GC race explicit
F5 MINOR     ACCEPT — V4 enumeration bound to closed T3 §15 Audit census
F6 MINOR     ACCEPT — concurrent distinct Document-code allocation explicit
F7 NOTE      traceability only
F8 NOTE      no change
```

Corrected candidate after Round-1 adjudication:

```text
technical correction commit         ca3a72d3f92eacea734bd1c583cd981e6e787bce
reviewed corrected candidate HEAD    eb7e0147cf575fe69290c231ea360af229917eeb
corrected candidate required CI      #1130 SUCCESS
Golden Flows                         6
cross-cutting properties             10
evidence classes                     6
application operations               78
operation 79                          ABSENT
new Product authority                NONE
T1→T8 reopen                          NONE
```

Bounded Fable Round 2 reviewed that exact corrected candidate through isolated Draft PR #156:

```text
review PR                           #156 CLOSED / UNMERGED
review branch                       review/t9-fable-r2
final review HEAD                   27b7ce63a8c63169b6ac8b582ee49621e7c86355
review CI                           #1132 SUCCESS
review delta                        docs/work/current/ai-dialog.md only
verdict                             CONVERGED
MATERIAL findings                   0
Round 3 justified                   NO
```

Round 2 confirmed F1/F2 closure, F3–F6 boundedness and the complete fixed envelope. One non-blocking MINOR remains: one §7 closure sentence is phrased in execution mode while adjacent T9 requirements are plan-mode. It fails closed and changes no accepted property; align that wording during the next natural candidate promotion/closure edit. No additional Fable round is justified.

T9 must prove accepted composition rather than manufacture implementation. A material contradiction reopens only the smallest owning upstream authority; test convenience may not create operation 79, new Product state, new Permission, new semantic owner or new runtime capability.

## Ratified T8-H baseline preserved by T9

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
operator ratification of the corrected, independently converged T9 candidate
→ if ratified, promote/absorb the T9 candidate into durable routed authority
→ align the non-blocking Fable Round-2 §7 wording precision during that promotion
→ remove temporary docs/work/current/t9-golden-flows.md before merge candidacy
→ record immutable T9 ratification evidence and update this roadmap
→ run required CI on the exact closure candidate
→ do not open T10 until T9 is integrated and the operator separately authorizes T10
→ do not begin T11 or T12
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
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | OPEN / ACTIVE; independent review CONVERGED; operator ratification pending |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NOT OPEN; requires T9 integration and separate operator authorization |
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
