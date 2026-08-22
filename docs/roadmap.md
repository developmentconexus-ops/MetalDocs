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
T10                                   OPEN / ACTIVE / REVIEW-CONVERGED
T11 → T12                             NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T10 — Transition / Cutover is **OPEN / ACTIVE** and its corrected technical candidate has completed bounded independent review with **CONVERGED / MATERIAL=0 / Round 3 NOT JUSTIFIED**.

The next gate is **explicit operator ratification of T10**. Review convergence is Evidence, not ratification, closure, merge authorization or permission to open T11.

## T10 opening proof

```text
opening main                           fc7030e98021bdb55fa806df68821cf19ed1a40c
opening branch                         arch/t10-transition-cutover
candidate Draft PR                     #158
open PRs before T10                    0
T1 → T9                                CLOSED / OPERATOR-RATIFIED / INTEGRATED
application operations                 78
operation 79                           ABSENT
T11 → T12                              NOT OPEN
Product implementation                 BLOCKED
legacy implementation                  ABSENT FROM LIVE TREE
```

The active branch-only technical candidate is `docs/work/current/t10-transition-cutover.md`. It remains temporary work, not durable authority, until explicit operator ratification and durable promotion.

## T10 selected architecture

T10 derives the smallest truthful technical activation/cutover/recovery contract from accepted T1→T9 authority. T7 remains binding: Launch has no historical business corpus to migrate, so T10 must not manufacture ETL, dual-write, compatibility or old/new Product authority merely for transition convenience.

Exactly five monotonic barriers remain:

```text
B0  source truth classified
B1  target privately prepared
B2  exact production candidate proven + fixture state clean-sealed
B3  first post-seal authoritative R10 Product mutation committed / point of no return
B4  authoritative recovery point exists + disposable serving estate fenced + canonical R10 serving activated
```

Core laws:

```text
historical business migration                         ABSENT
business authority                                    SINGULAR
B2 clean seal                                         operations/provenance evidence only
Product activation marker/table/endpoint              ABSENT
pre-B3 destructive reset                              allowed only while authority has not begun or become ambiguous
post-B3 destructive return to DEV/test truth          FORBIDDEN
post-B3 incident model                                R10 recovery / fail closed
B4 serving before authoritative recovery point        FORBIDDEN
DNS switch alone as legacy-serving fence              INSUFFICIENT
cleanup of any resource containing business truth     FORBIDDEN pending bounded adjudication
```

## Independent review evidence

Original operator-approved Lead candidate:

```text
candidate HEAD                           0b90f26690b2b2bbf627f0c72283ff14c0ce9b84
required CI                              #1153 SUCCESS
operator Lead approval                   EXPLICIT
```

Fable Round 1 — isolated Evidence PR #159:

```text
review branch                            review/t10-fable
final review HEAD                        0f47dfc2365433b5950fccac4b48106e7a7fa453
review CI                                #1155 SUCCESS
review delta                             docs/work/current/ai-dialog.md only
verdict                                  NOT CONVERGED
MATERIAL                                 3
Round 2                                  JUSTIFIED
PR                                       CLOSED / UNMERGED
```

Round-1 adjudication:

```text
F1 MATERIAL  ACCEPT / REMEDY REFINED
  → B2 closes with mechanical clean-baseline verification and an operations/provenance-only seal
  → proof mutation paths are fenced before bootstrap
  → B3 remains the first post-seal Product bootstrap commit
  → no Product activation marker/table/endpoint/Permission/owner

F2 MATERIAL  ACCEPT / REMEDY REFINED
  → B4 requires at least one complete authoritative R10 recovery point covering the current B3 baseline
  → total loss of canonical authority plus every coherent recovery point is catastrophic authority loss
  → fail closed; no automatic re-bootstrap or disposable-state promotion

F3 MATERIAL  ACCEPT
  → every disposable user-reachable DEV/test serving path is stopped/fenced before or at B4
  → DNS change alone is insufficient
  → cleanup requires absence of any business truth regardless of timestamp

F4 MINOR     ACCEPT
  → proof binds to the exact production candidate/profile
  → reset/rebuild re-arms affected state-dependent evidence

F5 MINOR     PARTIAL ACCEPT
  → T3 non-serving operations concern is semantic-bootstrap anchor
  → T8-D bootstrap/provisioner remains DDL/provisioning-only
  → only concrete implementation evidence may trigger bounded T8-G reopen

F6 NOTE      WORDING ALIGNED
F7 NOTE      NO CHANGE
```

Corrected technical candidate:

```text
technical correction                     7c5bb3e0106657c6e0db993afbe8d646b0ac09d1
independently reviewed candidate HEAD     c1afc292bc94f48bfd2146c3b4374342ff5c2701
required CI                               #1157 SUCCESS
```

Fable Round 2 — isolated bounded Evidence PR #160:

```text
review branch                            review/t10-fable-r2
review base                              c1afc292bc94f48bfd2146c3b4374342ff5c2701
final review HEAD                        937aebf9688516d1b0b1245eb014c0a6c03d6e7e
review CI                                #1159 SUCCESS
review delta                             docs/work/current/ai-dialog.md only
verdict                                  CONVERGED
MATERIAL                                 0
Round 3                                  NOT JUSTIFIED
PR                                       CLOSED / UNMERGED
```

Round 2 confirmed F1/F2/F3 closure, bounded F4–F6 handling and no regression of the fixed envelope. Two non-blocking precision notes remain:

```text
R2-N1  clean-baseline verification should be current at seal completion after proof-path fencing
R2-N2  pre-B3 reset wording relies on the B3 section's absolute prohibition after an authoritative mutation
```

Neither note changes architecture, creates a new barrier or justifies Round 3. They may be aligned during durable promotion if doing so preserves the independently reviewed meaning.

## Preserved T10 envelope

```text
barriers                         exactly 5 / B0→B4
historical business migration   absent
business authority              singular
application operations          78
operation 79                    absent
new Permission                  none
new semantic owner              none
new Product state               none
new current runtime capability  none
T1→T9 reopen                    none
T11/T12                         not open
Product implementation          blocked
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

If concrete evidence proves real business truth or a required compatibility consumer, T10 stops and routes a bounded reopen to the smallest owning authority before migration design proceeds.

## Integrated T9 baseline preserved

T9 remains **CLOSED / OPERATOR-RATIFIED / INTEGRATED**. Its durable technical authority is `architecture/validation-baseline.md` and immutable ratification snapshot is `decisions/t9-ratification.md`.

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

Integrated technical baseline remains:

```text
accepted application operations      78
Idempotency-Key creations            exact 10
ETag read / mutation domains         13 / 13
exact-byte resources                 exact 4
stable SPA route meanings            exact accepted T6 route set
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
operator explicitly ratifies or rejects the corrected independently converged T10 candidate
→ if ratified, promote the reviewed five-barrier contract into durable routed authority
→ align only non-semantic Round-2 precision notes where useful
→ record immutable T10 ratification evidence
→ remove temporary T10 work/review content from the integration tree
→ run required CI on the exact closure candidate
→ do not open T11 until T10 is ratified and integrated
→ do not implement Product code
```

No T10 implementation plan is authorized or created while Product implementation remains blocked.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | OPEN / ACTIVE / REVIEW-CONVERGED; OPERATOR RATIFICATION PENDING |
| T11 — Implementation Program & Execution Graph | Bounded work graph and proof obligations | NOT OPEN; opens only after T1→T10 accepted/integrated and separate authorization |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | NOT OPEN; opens after T11 |

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
