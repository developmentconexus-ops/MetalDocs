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

The active branch-only T10 candidate is `docs/work/current/t10-transition-cutover.md`. It is temporary work/Evidence, not durable architecture authority.

T10 must derive the smallest truthful technical activation/cutover/recovery contract from accepted T1→T9 authority. T7 already establishes that Launch has no historical business corpus to migrate; T10 may not manufacture ETL, dual-write, compatibility or old/new authority merely for transition convenience.

The corrected Lead direction remains one-way greenfield activation with exactly five monotonic barriers:

```text
B0  source truth classified
B1  target privately prepared
B2  exact production candidate proven + fixture state clean-sealed
B3  first post-seal authoritative R10 Product mutation committed / point of no return
B4  authoritative recovery point exists + disposable serving estate fenced + canonical R10 serving activated
```

Before B3, destructive technical reversal is permitted only while no post-seal Product mutation has crossed or ambiguously approached the authority boundary. After B3, destructive return to disposable pre-R10/DEV/test state is forbidden; incidents become R10 recovery. B4 may expose normal serving only after the authoritative R10 baseline is recoverable and every inventoried disposable user-serving path is fenced.

## T10 independent review state

The operator approved the original Lead candidate at:

```text
approved Lead candidate                0b90f26690b2b2bbf627f0c72283ff14c0ce9b84
approved candidate required CI         #1153 SUCCESS
candidate Draft PR                     #158
```

Independent Fable Round 1 reviewed that exact candidate through isolated Draft PR #159:

```text
review PR                              #159
review branch                          review/t10-fable
final review HEAD                      0f47dfc2365433b5950fccac4b48106e7a7fa453
review CI                              #1155 SUCCESS
review delta                           docs/work/current/ai-dialog.md only
verdict                                NOT CONVERGED
MATERIAL findings                      3
Round 2 justified                      YES
```

Lead adjudication accepts all three MATERIAL failure classes with bounded candidate-only corrections, while refining two proposed remedies:

```text
F1 MATERIAL  ACCEPT / REMEDY REFINED
  → B2 verified-clean seal is operations/provenance evidence only
  → proof mutation paths are fenced before bootstrap
  → B3 stays the first post-seal Product bootstrap commit
  → no Product activation marker/table/endpoint is created

F2 MATERIAL  ACCEPT / REMEDY REFINED
  → at least one complete authoritative R10 recovery point must cover B3 before B4
  → total loss of canonical authority + all coherent recovery points is catastrophic authority loss
  → no automatic re-bootstrap or disposable-state promotion

F3 MATERIAL  ACCEPT
  → all inventoried user-reachable disposable serving estate is fenced before/at B4
  → DNS change alone is insufficient fencing
  → cleanup may delete only resources containing no business truth, regardless of timestamp

F4 MINOR     ACCEPT
  → B2 evidence binds to the exact production candidate/profile
  → resets re-arm affected state-dependent proof

F5 MINOR     PARTIAL ACCEPT
  → T3 non-serving bootstrap concern is the semantic anchor
  → T8-D bootstrap/provisioner remains DDL/provisioning-only, not semantic Product bootstrap
  → if T11 proves accepted T8-G surfaces cannot realize bootstrap, bounded T8-G reopen is required before implementation

F6 NOTE      WORDING ALIGNED
F7 NOTE      NO CHANGE
```

Technical correction commit:

```text
7c5bb3e0106657c6e0db993afbe8d646b0ac09d1
```

No Product/T1→T9 authority is reopened by the current correction. No sixth barrier, Product state, Permission, semantic owner, runtime capability or operation 79 is introduced.

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

Any surviving external database, managed-content store, OIDC client, deployment, ingress or secret/config resource remains unclassified until B0 proves whether it contains authority. The expected DEV/test classification never grants disposal by itself.

If concrete evidence proves that real business truth or a required compatibility consumer exists, T10 stops and routes a bounded reopen to the smallest owning authority before migration design proceeds.

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
run required CI on the exact corrected T10 candidate HEAD
→ close Fable Round-1 Evidence PR #159 unmerged after the correction is proven
→ open isolated bounded Fable Round 2 from that exact corrected candidate
→ Round 2 confirms F1–F3 closure and fixed-envelope regression only; no unconstrained redesign
→ adjudicate any surviving MATERIAL falsifier against the smallest owning authority
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
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | OPEN / ACTIVE; corrected candidate awaiting bounded Fable Round 2 |
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
