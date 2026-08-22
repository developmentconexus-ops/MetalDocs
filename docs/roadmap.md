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

Current branch-only candidate pack:

```text
docs/work/current/t11-implementation-program.md
docs/work/current/t11-frontend-blueprint.md
docs/work/current/t11-wireframes.md
docs/work/current/t11-interaction-ledger.md
```

Reusable frontend implementation-readiness methodology:

```text
docs/development/functional-html-wireframe-method.md
```

Operator-approved bounded precision discovered during T11 frontend-readiness analysis:

```text
decisions/responsible-owner-selection-read.md
```

The T11 work pack is Evidence/candidate material, not durable ratified T11 authority. `t11-implementation-program.md` owns the consolidated execution DAG and node exit contracts. `t11-frontend-blueprint.md` owns the consolidated frontend implementation-readiness result; wireframes and interaction ledger provide bounded detail. All temporary work files must be promoted/absorbed or removed before T11 integration.

T11 does not authorize Product implementation, does not begin T12 and does not reopen accepted Product/T1→T10 authority by preference.

## T11 current candidate proof

```text
opening integrated main               cae6ba48df5d611959c0390e0f2b9b8194d62a9d
opening branch                         arch/t11-implementation-program
Draft PR                               #162
operator T11 authorization             EXPLICIT / 2026-08-22
operator node/frontend precision       EXPLICIT / 2026-08-22
operator F3-F01 bounded correction     APPROVED / 2026-08-22
Functional HTML Wireframe Method       ADDED / ROUTED / CI #1212 SUCCESS
application operations                 78
orphaned operations                    0
invented operations                    0
operation 79                           ABSENT
Idempotency-Key creations             exact 10
ETag read / mutation domains          13 / 13
exact-byte resources                  exact 4
accepted human goals                  16 / 16
material frontend surfaces            36 / 36
Screen Contracts                      36 / 36 READY
Navigation/Data Graph                 COMPLETE
markdown functional wireframes         COMPLETE CANDIDATE
Material Interaction Ledger           COMPLETE CANDIDATE
bidirectional frontend↔backend trace  78 / 78
HTML functional prototype              REVIEW CANDIDATE / EXTERNAL ARTIFACT
HTML stable Product routes             10 / 10
HTML material surfaces                 36 / 36
HTML operation manifest                78 / 78 unique
HTML pattern vocabulary                PRESENT
HTML trace metadata                    PRESENT
HTML candidate SHA-256                 37378abfb7671767823f07552d9ef2feabad107d33451466d79159d7d7728a12
unresolved MATERIAL frontend finding  0 before visual walkthrough
Product implementation                BLOCKED
```

The HTML prototype is deliberately an **external review artifact** while the repository remains architecture-first. Current Repository Standard v1 CI admits durable documentation as Markdown and blocks implementation surfaces. T11 does not weaken that allowlist merely to host a prototype. The prototype contains no real backend calls and is not Product runtime/frontend implementation. Its reviewed behavior must remain traceable to the repository-owned blueprint/wireframes/interaction ledger; if durable HTML hosting later becomes a real repository requirement, that policy change must be explicit rather than smuggled in through T11.

Coverage-first frontend planning corrected the T11 decomposition twice without silently changing Product authority: first by removing an impossible Organization-before-Authentication ordering, then by proving that Document creation and Document Work cannot truthfully close as separate user-facing implementation nodes. The final semantic execution partition is:

```text
S1  Identity + Organization + Access                         33
S2  Document Governance configuration                       10
S3  Document core + creation + authoring + Submission       22
    + Library + My Work authoring + History
S4  Governance work + Governance Case + Release/rendition    9
S5  Obsolescence + Audit                                     4
TOTAL                                                        78
```

S3 closes only when the real user journey is complete:

```text
Library
→ create/open Document
→ real Document Work target
→ DRAFT edit/upload
→ Submission
```

No node may close as “backend complete; frontend later” when it owns a material user-facing claim.

## Functional HTML prototype gate

The reusable method now requires the frontend implementation-readiness candidate to survive an interactive HTML walkthrough before independent T11 review.

The current HTML candidate is a single-file HTML/CSS/vanilla-JavaScript prototype derived from the accepted 36 Screen Contracts and Material Interaction Ledger. It provides:

```text
one AppShell vocabulary
accepted stable route simulation only
component-pattern catalog
material tables/cards/forms/dialogs/drawers/viewer/editor regions
interactive navigation
material success/failure scenario simulation
403 / 404 / 412 / CSRF / ambiguous-idempotency / upload-expiry / dependency / integrity states
machine-readable data-* trace metadata
10-route / 36-surface / 78-operation manifest
zero real backend calls
zero Product authority
```

The HTML prototype may reveal one of four classes of evidence:

```text
presentation/pattern defect
  → correct prototype/T11 frontend detail only

trace/navigation defect with accepted backend truth available
  → correct T11 Screen Contract / blueprint / ledger

accepted frontend realization contradiction
  → smallest bounded T8-F reopen

required accepted human goal not representable from current Product/wire truth
  → smallest Product/T6/T8-E reopen; never a screen-shaped convenience API
```

No visual preference by itself reopens Product/architecture authority.

## Approved bounded responsible-owner precision — T8-E-RO

T11 F3 proved one accepted later responsible-owner journey lacked a complete least-privilege human candidate read. The operator approved the smallest correction on 2026-08-22.

Existing operation 47 remains the sole Document Official read:

```text
GET /api/v1/documents/{document_id}
operationId: getDocument
```

`DocumentOfficialView` gains exactly one optional derived member:

```text
responsible_owner_candidates?: UserReference[]
```

Binding law:

```text
present iff
  getDocument is otherwise disclosable
  AND current canonical document.owner.manage = ALLOW for the exact Document

when present
  complete current D4-eligible target set
  = existing + same Company + ENABLED Users
  ordered user_id ASC

when absent
  candidate existence/reason cannot be inferred
```

The candidate projection grants no authority and is not part of the ResponsibleOwner ETag domain. Replacement still uses:

```text
getDocumentResponsibleOwner
→ ResponsibleOwnerView + strong ETag
→ replaceDocumentResponsibleOwner(target user_id, If-Match)
→ current AuthZ + D4 eligibility/offboarding revalidation
```

Effect:

```text
new Product capability     none
new Permission             none
new semantic owner         none
new stable SPA route       none
new application operation  none
operation 79               absent
ETag domains               13 / 13 unchanged
Idempotent creations       exact 10 unchanged
exact-byte resources       exact 4 unchanged
```

The approved precision is recorded in `decisions/responsible-owner-selection-read.md`. Before T11 integration, the effective T6/T8-E/T8-F owners must be consolidated to match that exact decision so no contradictory second authority remains.

## Predecessor T10 authority

Durable T10 authority:

```text
architecture/transition.md
```

Immutable T10 ratification evidence:

```text
decisions/t10-ratification.md
```

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

## Ratified T10 result

```text
B0  source truth classified
B1  target privately prepared
B2  exact production candidate proven + verified clean seal
B3  first post-seal authoritative R10 Product mutation / point of no return
B4  authoritative recovery point exists + disposable serving estate fenced + canonical R10 serving activated
```

Core transition law remains:

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
operator visual walkthrough of the current Functional HTML prototype candidate
→ inspect AppShell, route composition, tables/cards/forms/drawers/dialogs, reusable patterns and all material failure/recovery scenarios
→ record any finding against the owning Screen Contract / interaction / route / backend trace; do not redesign by visual preference alone
→ iterate the HTML candidate until the operator accepts the functional prototype behavior
→ re-run HTML ↔ frontend blueprint ↔ Interaction Ledger reconciliation; require 36/36 surfaces and 78/78 operations with zero unbound material control
→ record the final reviewed HTML artifact SHA-256 in this roadmap/PR checkpoint
→ self-review the resulting consolidated T11 candidate and verify documentation-only repository scope
→ run required CI on the exact repository candidate HEAD
→ present the exact repository HEAD + CI + final HTML artifact hash for separate explicit operator approval before independent review
→ only after that approval create isolated review/t11-fable from the exact approved candidate following Repository Standard v1 and provide the exact approved HTML artifact to the reviewer as Evidence
→ adjudicate reviewer Evidence against current authority; correct only material defects
→ require bounded re-review until MATERIAL=0 or route the smallest justified reopen
→ after convergence consolidate the approved T8-E-RO precision into the effective T6/T8-E/T8-F owning authorities
→ prepare durable T11 implementation-program authority + minimum durable frontend blueprint required by T12/future implementation
→ record immutable T11 ratification evidence and remove temporary T11 work files
→ do not begin T12
→ do not implement Product code while roadmap implementation gate remains BLOCKED
```

T11 is active only as architecture/planning. T12 remains not open. Product implementation remains blocked.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-RO bounded precision approved and pending consolidation before T11 integration |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-RO frontend consumption precision approved and pending consolidation before T11 integration |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T10 — Transition / Cutover | Real current→target transition, authority edge, recovery and rollback barriers | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T11 — Implementation Program & Execution Graph | Bounded work graph, exact node-exit states, frontend implementation-readiness and proof obligations | OPEN / ACTIVE candidate on Draft PR #162; Functional HTML prototype walkthrough is current gate |
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
