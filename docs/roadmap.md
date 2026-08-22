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
T1 → T10                              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                                   OPEN / ACTIVE CANDIDATE
T12                                   NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T11 — Implementation Program & Execution Graph is **OPEN / ACTIVE** on Draft PR #162 after explicit operator authorization on 2026-08-22.

Temporary branch-only T11 candidate pack:

```text
docs/work/current/t11-implementation-program.md
docs/work/current/t11-node-completion-contracts.md
docs/work/current/t11-frontend-readiness.md
docs/work/current/t11-frontend-coverage.md
```

The pack is Evidence/work product, not durable authority. The implementation-program Lead owns graph/closure; companions define mandatory node-exit and frontend implementation-readiness precision. All temporary files must be absorbed or removed before integration.

T11 does not authorize Product implementation, does not begin T12 and does not reopen accepted Product/T1→T10 authority by preference.

Opening / current precision proof:

```text
opening integrated main               cae6ba48df5d611959c0390e0f2b9b8194d62a9d
opening branch                         arch/t11-implementation-program
Draft PR                               #162
operator T11 authorization             EXPLICIT / 2026-08-22
operator node/frontend precision       EXPLICIT / 2026-08-22
application operations                 78
operation 79                           ABSENT
F1 frontend coverage                   COMPLETE / 16 human goals / 78 operations
F1 material T11 findings               5 found / 5 adjudicated / 0 unresolved
Product/T8-F semantic reopen           NOT JUSTIFIED by current evidence
Product implementation                 BLOCKED
```

F1 corrected the open T11 implementation graph without changing Product/T8 authority:

```text
S1  Identity + Organization + Access                              33
S2  Document Governance configuration                            10
S3  Library + Document core + template-role + History             9
S4  Revision authoring + My Work authoring + content + Submission 13
S5  Governance work + Governance Case + Release/rendition          9
S6  Obsolescence + Audit                                           4
TOTAL                                                             78
```

Predecessor durable T10 authority:

```text
architecture/transition.md
```

Immutable T10 ratification evidence:

```text
decisions/t10-ratification.md
```

Current progression, implementation permission and exact next action remain owned here.

## T10 integrated proof

```text
opening main                           fc7030e98021bdb55fa806df68821cf19ed1a40c
candidate PR                           #158
operator-approved original Lead        0b90f26690b2b2bbf627f0c72283ff14c0ce9b84
original Lead required CI              #1153 SUCCESS
Round-1 Evidence PR                    #159 CLOSED / UNMERGED
Round-1 final review HEAD              0f47dfc2365433b5950fccac4b48106e7a7fa453
Round-1 review CI                      #1155 SUCCESS
Round-1 verdict                        NOT CONVERGED / MATERIAL=3
technical correction commit            7c5bb3e0106657c6e0db993afbe8d646b0ac09d1
independently reviewed candidate HEAD  c1afc292bc94f48bfd2146c3b4374342ff5c2701
corrected candidate required CI        #1157 SUCCESS
Round-2 Evidence PR                    #160 CLOSED / UNMERGED
Round-2 final review HEAD              937aebf9688516d1b0b1245eb014c0a6c03d6e7e
Round-2 review CI                      #1159 SUCCESS
Round-2 verdict                        CONVERGED / MATERIAL=0
Round 3                                NOT JUSTIFIED
post-review status carrier             aadb2a81136dcf5020804c86738dc84c263d52f8
status-carrier required CI             #1160 SUCCESS
operator ratification                  EXPLICIT / 2026-08-22
closure candidate HEAD                 cc408964e4e9e4719e9bc0808b9ec49a076df89f
Draft required CI                      #1166 SUCCESS
merge authorization                    EXPLICIT / 2026-08-22
merge-candidate required CI            #1167 SUCCESS
candidate tree                         c3de41e73ee153278e0869ac80640cc945ae26b2
squash merge / integrated main         e8f415ec16df9cc2d4623981412e1ac21c3c6647
integrated main tree                   c3de41e73ee153278e0869ac80640cc945ae26b2
T10 integration                        VERIFIED
T10 closeout PR                        #161
T10 closeout required CI               #1168 SUCCESS / #1169 SUCCESS
T10 closeout / current main            cae6ba48df5d611959c0390e0f2b9b8194d62a9d
```

The T10 squash merge integrated the exact authorized closure tree. PR #161 recorded completed integration and produced the current T11 opening base.

## Ratified T10 result

```text
B0  source truth classified
B1  target privately prepared
B2  exact production candidate proven + verified clean seal
B3  first post-seal authoritative R10 Product mutation / point of no return
B4  authoritative recovery point exists + disposable serving estate fenced + canonical R10 serving activated
```

Core transition law:

```text
one-way greenfield activation
proof before authority
operations/provenance clean seal, never Product activation state
first authoritative Product commit = point of no return
authoritative recovery point before ordinary serving
DEV/test user-serving paths fenced before ordinary serving
single business authority
post-B3 forward recovery only
```

Explicitly absent:

```text
historical business migration
generic ETL/import framework
dual write
dual Product authority
legacy read fallback
schema/API compatibility bridge
Product activation marker/table/endpoint
operation 79
```

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

T7 remains binding: Launch has no historical business corpus to migrate. Contrary concrete evidence triggers the smallest bounded reopen rather than silent preservation/compatibility machinery.

## Exact next action

```text
continue T11 frontend implementation-readiness on Draft PR #162
→ F2 derive the complete material interaction-surface/state inventory from the now-closed F1 Coverage Matrix
→ prove every F2 surface is justified by an accepted human goal and backend contract; reject cosmetic over-fragmentation
→ F3 derive a Screen Contract for every material surface
→ F4 derive the Navigation/Data Graph and prove every target identity/read path
→ classify any new material finding; never invent operation 79 or a screen-shaped API
→ only after F2→F4 are coherent, F5 draw functional low/mid-fidelity wireframes for every material screen/safe-action state
→ F6 derive Material Interaction Ledger
→ F7 execute final bidirectional frontend↔backend trace and reconcile 78/78 coverage
→ feed the completed frontend pack into S1→S6 Node Completion Contracts
→ run required CI on the exact completed T11 candidate HEAD
→ present exact completed candidate for explicit operator approval before independent review
→ only then create isolated review/t11-fable from the exact approved candidate
→ adjudicate reviewer Evidence; re-review until MATERIAL=0 or smallest justified reopen
→ after convergence prepare durable T11 authority + ratification and remove temporary work files
→ do not begin T12
→ do not implement Product code while roadmap gate remains BLOCKED
```

T11 is active only as architecture/planning. T8-F remains closed unless frontend-readiness evidence materially falsifies it. T12 remains not open. Product implementation remains blocked.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T10 — Transition / Cutover | Real current→target transition, authority edge, recovery and rollback barriers | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T11 — Implementation Program & Execution Graph | Bounded work graph, exact node-exit states, frontend implementation-readiness and proof obligations | OPEN / ACTIVE candidate on Draft PR #162 |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | NOT OPEN; opens only after T11 closure |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED / INTEGRATED
T9  CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10 CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or generic infrastructure fashion are not reopen triggers.