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

T11 — Implementation Program & Execution Graph remains **OPEN / ACTIVE** on Draft PR #162.

Current branch-only T11 work:

```text
docs/work/current/t11-implementation-program.md
docs/work/current/t11-frontend-blueprint.md
docs/work/current/t11-wireframes.md
docs/work/current/t11-interaction-ledger.md
```

Reusable frontend method:

```text
docs/development/functional-html-wireframe-method.md
```

Approved bounded precision:

```text
decisions/responsible-owner-selection-read.md
```

The work pack is candidate Evidence, not durable T11 authority. T11 does not authorize Product implementation or T12.

## T11 fixed system invariants

```text
opening integrated main               cae6ba48df5d611959c0390e0f2b9b8194d62a9d
branch                                 arch/t11-implementation-program
Draft PR                               #162
application operations                 78
orphaned / invented                    0 / 0
operation 79                           ABSENT
Idempotency-Key creations             10
ETag read / mutation domains          13 / 13
exact-byte resources                  4
Product implementation                BLOCKED
```

The operator-approved T8-E-RO precision remains binding and adds no operation, Permission, owner or route.

## Frontend planning correction — 2026-08-22

The first T11 frontend pass proved backend/screen coverage but moved too quickly from semantic Screen Contracts to a whole-product HTML prototype.

Operator feedback rejected that workflow as insufficient for implementation readiness because it skipped deliberate UX/product-design work for:

```text
user needs / jobs
information architecture
navigation mental model
reference study
layout alternatives
screen hierarchy
relative size / density
cards vs tables vs master-detail
progressive disclosure
screen-by-screen operator walkthrough
pattern derivation after evidence of repetition
```

Therefore the previously generated all-at-once HTML prototype is **SUPERSEDED / NOT ACCEPTED** and is not a T11 review candidate or implementation baseline.

No Product/T1→T10 authority is reopened by this UX-method correction.

## Frontend Product Experience Planning Method v2

The reusable method is now structured as:

```text
P0  Recover accepted authority
 ↓
P1  Actors / jobs / user needs
 ↓
P2  End-to-end user flows
 ↓
P3  Frontend Coverage Matrix
 ↓
P4  Information Architecture
 ↓
P5  Screen / material-surface inventory
 ↓
P6  Reference Study — per functional block
 ↓
P7  Competing Layout Hypotheses
 ↓
P8  Structural Wireframe — block-by-block + operator adjudication
 ↓
P9  Screen Contract + bidirectional backend trace
 ↓
P10 Derive reusable component/interaction patterns
 ↓
P11 Interactive Low-Fidelity HTML
 ↓
P12 Adversarial UX + Architecture Walkthrough
 ↓
P13 Visual Design Handoff
 ↓
P14 Frontend Implementation-Readiness Closure
```

Hard method law:

```text
no whole-product wireframe generation in one pass
no high-impact block advances past unresolved structural findings
no layout pattern chosen merely from backend shape
no component vocabulary frozen before reviewed repetition exists
no screen-shaped API for frontend convenience
operator must see/discuss material screens block-by-block before lock
```

External products/design systems are reference Evidence only. Accepted Product/system authority + user evidence + explicit operator adjudication determine MetalDocs UX.

## MetalDocs application order after method convergence

After independent review of the method, MetalDocs frontend planning restarts from UX structure rather than editing the rejected all-at-once HTML.

Candidate block order:

```text
B01  App Shell + global Information Architecture
B02  Library / discovery
B03  Document Official
B04  Document Work / authoring
B05  My Work
B06  Governance
B07  History / Audit
B08  Administration
```

Block names/order remain planning candidates until B01 global IA proves the final grouping.

Each material block follows:

```text
bounded accepted authority
→ user goals / local flow
→ reference study where material
→ 2–3 layout hypotheses when genuine ambiguity exists
→ structural candidate
→ operator visual walkthrough
→ finding/adjudication
→ LOCK
→ vertical Screen Contract / backend trace
→ only then HTML realization
```

The operator and assistant explicitly converse screen-by-screen/block-by-block; the next material block is not generated automatically.

## Final implementation DAG candidate

The backend/implementation partition remains:

```text
P0 authority / implementation-admission pin
 ↓
P1 structural + executable-contract spine
 ├────────────────────┐
P2 persistence        P3 runtime/dependency/non-serving bootstrap
 └──────────┬─────────┘
            ↓
S1 Identity + Organization + Access                         33
 ↓
S2 Document Governance configuration                       10
 ↓
S3 Document core + creation + authoring + Submission       22
   + Library + My Work authoring + History
 ↓
S4 Governance work + Governance Case + Release/rendition    9
 ↓
S5 Obsolescence + Audit                                     4
 ↓
P4 runtime / durable-work / recovery closure
 ↓
T10 B1 private target
 ↓
P5 whole implementation proof closure
 ↓
T10 B2 → B3 → B4
```

`33 + 10 + 22 + 9 + 4 = 78`.

Frontend UX replanning may refine how accepted capability is presented but may not change this operation census or Product semantics without a material bounded reopen.

## T8-E-RO — approved responsible-owner precision

Existing operation 47 remains `getDocument`. `DocumentOfficialView` gains:

```text
responsible_owner_candidates?: UserReference[]
```

Binding law:

```text
present iff current document.owner.manage = ALLOW for the exact Document
contents = complete existing + same-Company + ENABLED Users
order = user_id ASC
absence discloses neither candidate existence nor reason
```

The list grants no authority and is outside the ResponsibleOwner ETag domain. Replacement still rechecks current AuthZ, D4 eligibility/offboarding serialization and `If-Match`.

Before T11 integration, effective T6/T8-E/T8-F owners must consolidate this approved precision.

## T10 preserved authority

Binding barriers remain:

```text
B0 source truth
→ B1 private target
→ B2 exact candidate proof + verified clean seal
→ B3 first authoritative Product mutation / point of no return
→ B4 authoritative recovery point + serving fence + canonical activation
```

No historical business migration, dual Product authority, legacy fallback, compatibility bridge, Product activation marker or operation 79 is introduced.

## Exact next action

```text
self-review Frontend Product Experience Planning Method v2
→ run required CI on exact candidate HEAD
→ freeze exact methodology HEAD for isolated adversarial review
→ create review/t11-frontend-method-fable targeting arch/t11-implementation-program
→ Fable reviews methodology only as Principal Product Designer + Information Architect + Senior Frontend Architect + adversarial architecture reviewer
→ classify findings MATERIAL / IMPORTANT / OPTIONAL / UNSUPPORTED PREFERENCE
→ adjudicate evidence; correct only justified findings
→ bounded re-review until methodology MATERIAL=0
→ then apply the converged method to MetalDocs from B01 App Shell + global IA
→ research/reference study + layout hypotheses
→ operator visual walkthrough and explicit block lock
→ proceed block-by-block; never generate the remaining screens automatically
→ after all blocks converge, reconcile complete frontend↔backend trace and T11 implementation graph
→ only then perform whole-T11 independent review / ratification work
→ do not begin T12
→ do not implement Product code
```

## Remaining architecture program

| Stage | Owns | State |
|---|---|---|
| T8-E | Executable application wire | CLOSED / INTEGRATED; T8-E-RO consolidation pending T11 close |
| T8-F | Frontend realization | CLOSED / INTEGRATED; T8-E-RO consolidation pending T11 close |
| T8-G | Runtime / process / deployment | CLOSED / INTEGRATED |
| T8-H | Whole-T8 coherence | CLOSED / INTEGRATED |
| T9 | Golden Flows / validation baseline | CLOSED / INTEGRATED |
| T10 | Transition / cutover | CLOSED / INTEGRATED |
| T11 | Implementation graph + implementation-readiness | OPEN / frontend method adversarial-review gate |
| T12 | Adversarial implementation-readiness | NOT OPEN |

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

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or infrastructure fashion are not reopen triggers.
