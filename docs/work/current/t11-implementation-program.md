# T11 — Implementation Program & Execution Graph

> **TEMPORARY T11 CANDIDATE / BRANCH-ONLY WORK.** This file is not durable authority and must be absorbed or removed before merge. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose and boundary

T11 derives the smallest bounded implementation work graph and proof obligations from accepted T1→T10 authority so later Product implementation can proceed without architectural improvisation, duplicate authority, late integration, frontend/backend drift or proof-by-ceremony.

T11 does **not** implement Product code, create schema/OpenAPI/frontend/runtime/deploy implementation, begin T12, choose a new Product capability, add an application operation, or reopen accepted authority by preference.

Fixed opening state:

```text
opening main                          cae6ba48df5d611959c0390e0f2b9b8194d62a9d
T1 → T10                               CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                                   OPEN / ACTIVE candidate
T12                                   NOT OPEN
Product implementation                BLOCKED
legacy implementation in live tree    ABSENT
application operations                78
operation 79                          ABSENT
Idempotency-Key creations             exact 10
ETag read / mutation domains          13 / 13
exact-byte resources                  exact 4
```

The future execution graph is inert while the roadmap implementation gate is BLOCKED. It becomes executable only after every final implementation-gate condition in `docs/roadmap.md` is satisfied.

## 2. Candidate composition

Current temporary T11 work:

```text
docs/work/current/t11-implementation-program.md
  → execution DAG, operation assignment, proof/cutover integration, T11 closure

docs/work/current/t11-node-completion-contracts.md
  → exact MUST-BE-IMPLEMENTED exit state for every P/S node

docs/work/current/t11-frontend-readiness.md
  → frontend readiness method

docs/work/current/t11-frontend-coverage.md
  → Product/backend/frontend coverage reconciliation + current T11 graph findings
```

These are one branch-only candidate pack, not parallel durable authorities. The companions refine this Lead and are mandatory parts of the candidate. Durable promotion consolidates their binding outcome and removes temporary work files.

## 3. Decision frame

Accepted authority already fixes:

```text
Product / ownership / lifecycle / Authorization / Audit / content / async semantics
backend + internal-interface + persistence topology
closed-world/default-deny first-party import graph
executable 78-operation application wire
frontend route/lens/state/transport realization
runtime / process / deployment realization
Whole-T8 coherence
T9 Golden Flows + evidence classes + cross-cutting falsifiers
T10 B0→B4 one-way cutover law
```

The live tree contains no Product implementation to preserve.

The remaining root risk is **implementation ambiguity across boundaries**:

```text
execution-order ambiguity
node-exit ambiguity
late backend/frontend integration
screen-shaped API invention
protected HTTP work attempted before real AuthN/AuthZ exists
actionable navigation exposed before its target surface exists
proof deferred until defects are expensive to localize
```

Target invariant:

```text
every future implementation increment
→ has one bounded graph position
→ consumes only accepted authority
→ has explicit prerequisites
→ has an exact observable completion contract
→ implements backend/persistence/wire/frontend coherently where applicable
→ satisfies the accepted closed-world dependency graph
→ has real positive + causal negative proof
→ never creates a second semantic authority
```

T11 does not freeze arbitrary private file splits, patch versions, secret values, environment identifiers, mechanical commit counts or ornamental UI pixels. It **does** freeze every material outcome needed to prevent semantic guessing during implementation.

## 4. Execution-shape decision

Rejected:

```text
A technical-layer waterfall
  contract → database → backend → frontend → runtime → tests

B Golden-Flow-only vertical slices
  GF1 → GF2 → ... → GF6
```

Selected:

```text
C minimal shared spines
  + coherent vertical user/capability tranches
  + exact node-exit contracts
  + global proof closure
```

T9 Golden Flows remain validation composition, not the complete implementation census.

The initial T11 Lead split Organization before Authentication/Authorization. The F0/F1 frontend coverage pass falsified that ordering because all 78 operations require `MetalDocsSession` and the Admin Organization/Access surfaces have cross-dependencies. The graph is therefore corrected **inside open T11**, not by weakening proof or reopening Product/T8 authority.

## 5. Corrected global execution law

After future implementation authorization:

```text
P0  authority/admission pin
 ↓
P1  structural + executable-contract spine
 ├────────────────┐
 ↓                ↓
P2 persistence    P3 runtime/dependency + non-serving bootstrap shell
 └────────┬───────┘
          ↓
S1 Identity + Organization + Access — 33 ops
 ↓
S2 Document Governance configuration — 10 ops
 ↓
S3 Library + Document core + template-role + History — 9 ops
 ↓
S4 Revision authoring + My Work authoring + content + Submission — 13 ops
 ↓
S5 Governance work + Governance Case + Release/rendition — 9 ops
 ↓
S6 Obsolescence + Audit — 4 ops
          ↓
P4 runtime / recovery closure
          ↓
T10 B1 private target
          ↓
P5 whole implementation proof closure on exact private candidate
          ↓
T10 B2 real proof + clean seal
          ↓
T10 B3 first authoritative Product mutation / point of no return
          ↓
T10 B4 recovery point + serving fence + canonical activation
```

T10 B0 remains a prerequisite to treating target preparation as cutover and is revalidated before B1.

Count proof:

```text
33 + 10 + 9 + 13 + 9 + 4 = 78
```

Why S1 is intentionally larger than the old split:

```text
all application operations require current session
Admin Organization uses Authentication-owned provider-subject lookup
Admin Access needs current Organization identity/reference truth
current authorization must exist before later protected Product E3/E4 proof
T3/T10 already admit explicit non-serving bootstrap/recovery concern for the first authority baseline
```

This removes an artificial cycle instead of creating mocks/bypasses between Identity, Organization and Access.

## 6. Program-node summary

| Node | Work boundary | Depends on | App ops | Required exit basis |
|---|---|---:|---:|---|
| P0 | implementation admission + authority snapshot | final roadmap gate | 0 | exact current authority/census pinned; no Product code |
| P1 | structural/package + executable contract + SPA transport spine | P0 | 0 semantic | E1/V1/V2 structural; generated Go/TS; closed-world import verifier fires |
| P2 | PostgreSQL/transaction/idempotency correctness spine | P1 | 0 | real E2 shared persistence mechanics; no database-first semantic prebuild |
| P3 | runtime/config/dependency shell + accepted non-serving bootstrap realization | P1 | 0 | one runtime; bootstrap fenced from serving; no operation 79 |
| S1 | AuthN/session + Organization + Authorization + Admin Organization/Access | P2 + P3 | 33 | GF1/V3; real protected HTTP/browser baseline established |
| S2 | DocumentType/governance/eligible-template/numbering config | S1 | 10 | Admin Document Governance base config real; OCC/disclosure proof |
| S3 | Library/create/Official core/responsible-owner/template-role/History | S2 | 9 | Product core + concrete template-role enrichment + honest progressive lenses |
| S4 | create/enter Revision + authoring work + DRAFT/upload/Submission | S3 | 13 | GF3; complete official→work authoring path with exact content/OCC |
| S5 | governance work/case + Release/source/rendition | S4 | 9 | GF4; complete My Work governance target + Release presentation |
| S6 | obsolescence + Audit | S5 | 4 | GF5; complete accepted user-facing Product surface |
| P4 | runtime failure isolation/shutdown/recovery | S1→S6 + P3 | 0 | GF6 + V8/V9/V10 real failure/recovery evidence |
| P5 | exact private whole-implementation proof candidate | all prior nodes + B1 | 78 closed | 6/6 GF; 10/10 V; wire runtime closure; censuses; frontend drift check |

The mandatory exact exit meaning of every row is `t11-node-completion-contracts.md`. This summary cannot weaken that companion.

## 7. Exact 78-operation implementation assignment

The canonical T8-E operation ledger remains the wire authority. T11 assigns **implementation rows only**, by exact ledger number; it does not duplicate paths/schemas/problems.

```text
S1 Identity + Organization + Access
  T8-E operations 1–33                                  = 33
  + GET /auth/login and GET /auth/callback outside census

S2 Document Governance configuration
  T8-E operations 34–43                                 = 10

S3 Library + Document core + template-role + History
  T8-E operations 44–51 + 53                             = 9

S4 Revision authoring + My Work authoring + content + Submission
  T8-E operation 52
  + operation 54
  + operations 56–66                                    = 13

S5 Governance work + Governance Case + Release/rendition
  T8-E operation 55
  + operations 67–74                                     = 9

S6 Obsolescence + Audit
  T8-E operations 75–78                                  = 4
```

Coverage proof:

```text
rows 1–78 assigned exactly once
orphaned rows = 0
multiply-assigned rows = 0
invented rows = 0
operation 79 = absent
```

Important non-contiguous assignments are deliberate user-flow closures:

```text
50–51 concrete Document template-role
  → S3, after ordinary Document creation exists; frontend home remains Admin Document Governance

52 createDocumentRevision + 54 listAuthoringWork
  → S4 with the real Document Work target; no dead authoring navigation

55 listGovernanceWork
  → S5 with the real Governance Case target; no dead governance navigation
```

A path/method outside the canonical 78 proposed as a new application operation is a STOP requiring the smallest Product/T6/T8-E reopen.

## 8. Node completion law

`t11-node-completion-contracts.md` defines mandatory:

```text
ENTRY
PRODUCTION STATE
PERSISTENCE STATE
WIRE STATE
DEPENDENCY GRAPH STATE
FRONTEND STATE
PROOF STATE
EXPLICIT ABSENCES
EXIT
```

A node cannot close as:

```text
backend done / frontend later
schema exists / behavior later
endpoint exists / owner path not integrated
UI green against mocks / real API later
imports compile / architecture edge unverified
test exists / causal falsifier absent
```

P1 makes T8-B closed-world dependency enforcement executable; every downstream node keeps the live package graph classified/legal.

## 9. Progressive stable-lens law

A stable T8-F route may be enriched by later tranches when new accepted capability becomes reachable, but the exact progression is part of the node contract.

Current planned progression:

```text
/admin/document-governance
  S2 base DocumentType/governance/eligible-template/numbering/config
  → S3 concrete Document template-role administration

/documents/:document_id
  S3 official/core + responsible-owner
  → S4 create/enter Revision
  → S5 Release/source/OfficialRendition presentation
  → S6 obsolescence state/actions

/work
  S4 authoring projection + live Work target
  → S5 governance projection + live Governance Case target

/documents/:document_id/history
  S3 operation/surface for reachable history
  → S4/S5/S6 regression/enrichment as additional accepted facts become reachable
```

No tranche may expose an actionable navigation edge whose admitted target surface is knowingly absent.

## 10. Frontend implementation-readiness program

T8-F remains closed. T11 derives implementation-ready screen/action/data realization from it before Product implementation.

Mandatory sequence:

```text
F0 authority recovery
→ F1 Frontend Coverage Matrix
→ F2 material interaction-surface inventory
→ F3 Screen Contracts
→ F4 Navigation/Data Graph
→ F5 functional low/mid-fidelity wireframes
→ F6 Material Interaction Ledger
→ F7 bidirectional frontend↔backend trace
→ F8 material finding classification / smallest justified reopen
→ F9 frontend readiness closure
```

Current F1 output is `t11-frontend-coverage.md`. It has already produced the graph correction recorded in §§5–7 without requiring a Product/T8-F reopen.

### Hard no screen-shaped API

If a screen needs absent information/action:

```text
STOP affected surface
→ prove whether need follows from accepted Product/human goal
→ classify gap
→ if material, reopen only smallest Product/T6/T8-E/T8-F owner
→ never add operation 79/convenience endpoint silently
```

Frontend may not be forced to guess material backend truth merely to avoid a justified bounded reopen.

Wireframes freeze interaction meaning, not ornamental pixels.

## 11. Future frontend implementation binding

The completed T11 frontend pack binds to tranches, not a late frontend phase:

```text
P1  SPA/router/generated TypeScript transport shell
S1  auth/session + Admin Organization + Admin Access
S2  Admin Document Governance base configuration
S3  Library/create/Document Official core/History + template-role admin enrichment
S4  Document Official revise entry + My Work authoring + Document Work/Submission
S5  My Work governance + Governance Case + Release/Official presentation
S6  Document Official obsolescence + Audit + final accepted lens enrichments
```

Every browser mutation relies on server authority. `allowed_actions` remain hints. No frontend Authorization engine, parallel global server store, handwritten application DTO contract or duplicate lifecycle state is permitted.

## 12. Proof law per implementation node

A node closes only when all applicable claims satisfy T9:

```text
real production subject executed/mechanically inspected
+
positive invariant observed
+
causal negative/fault case proves control fires
```

Mocks/fakes may support local tests but cannot close PostgreSQL/composed HTTP/browser/external-mechanism/recovery claims.

Each future mergeable increment identifies:

```text
accepted authority consumed
node/tranche owned
completion-contract clauses advanced/closed
production subjects changed
positive proof
causal negative/fault proof
T9 GF/V obligations advanced/closed
frontend Screen Contracts/interactions advanced/closed
application-census effect limited to assigned rows
cross-cutting census remains within accepted counts
remaining downstream prerequisites
```

No baseline waiver/placeholder/dormant capability may make an unresolved material contradiction appear green.

## 13. PR/integration slicing law

A graph node may use multiple PRs when that reduces review blast radius, but every PR leaves a coherent independently testable increment. Split at material ownership/proof seams, not arbitrary file count.

```text
no direct commits to main
no Product implementation before P0
no speculative cross-node scaffolding without current consumer
no dormant tables/endpoints/workers/frameworks
no hand-edited generated application projection
no merge knowingly breaking a closed invariant
no proof deferred when current claim is already testable
no semantic tranche declared complete while its assigned reviewed frontend contract is incomplete
```

P2 and P3 are the principal intentional early parallel branches.

## 14. T10 barrier overlay

T11 preserves exactly:

```text
B0 source truth
→ B1 private target
→ B2 real proof + clean seal
→ B3 first authoritative Product mutation / point of no return
→ B4 recovery point + serving fence + canonical activation
```

P3's non-serving bootstrap realization is technical implementation of already-accepted T3/T10 concern. It does **not** cross B3 during ordinary development/proof. B3 occurs only after B2 under T10 launch authority when the first authoritative post-seal Product mutation commits through the admitted non-serving bootstrap/administrative concern.

P5 proves the exact B1 private candidate. Post-B3 recovery is forward on R10 only.

## 15. Stop / reopen routing

```text
implementation/proof defect
  → fix owning implementation node

frontend placement/presentation issue with adequate accepted backend truth
  → correct T11 frontend readiness only

T11 graph/dependency boundary falsified by coverage/proof evidence
  → correct T11 while open; bounded T11 reopen after ratification

accepted mechanism assumption falsified
  → smallest implicated T8 technical reopen

required accepted human goal cannot be safely represented from wire/read models
  → smallest Product/T6/T8-E/T8-F reopen justified by evidence

missing/extra application operation or consumer
  → STOP; smallest Product/T6/T8-E reopen; never operation 79 silently

semantic owner/lifecycle contradiction
  → smallest accepted semantic-owner reopen

real pre-R10 authoritative business truth discovered
  → smallest T7/T10 reopen
```

Version churn, private file splits, ornamental UI choice and mechanical refactoring do not reopen architecture by themselves.

## 16. T11 closure contract

T11 can be ratified only when all are true:

```text
implementation DAG has no unresolved cycle/unowned material work
Node Completion Contract exists for every P0→P5 / S1→S6 node
all node exit contracts specify exact applicable production/persistence/wire/dependency/frontend/proof state
78/78 application operations assigned exactly once
operation 79 absent
Idempotency-Key creations exact 10
ETag read/mutation domains 13/13
exact-byte resources exact 4
6/6 Golden Flows have implementation/proof owner
10/10 T9 cross-cutting properties have implementation/proof owner
T8-E runtime wire-conformance closure assigned to P5 real composed path
real external-mechanism claims retain E5 lanes
failure/recovery claims retain E6 lanes
T8-B closed-world import graph has executable verifier/negative-proof owner
Frontend Coverage Matrix complete and all findings adjudicated
material interaction-surface inventory complete
Screen Contract exists for every material surface
Navigation/Data Graph complete
functional wireframes cover every material screen/safe-action state
Material Interaction Ledger complete
backend→frontend trace complete
frontend→backend trace complete
78/78 frontend operation coverage reconciled
zero unresolved MATERIAL frontend/backend coverage finding
frontend remains semantic consumer only
T10 B0→B4 preserved exactly
no T12 work begun
no Product implementation begun
independent bounded T11 review converged MATERIAL=0
operator ratification explicit
```

Expected durable promotion after convergence:

```text
docs/architecture/implementation-program.md
docs/decisions/t11-ratification.md
docs/index.md route update
minimum retained implementation-readiness/wireframe pack required by T12/future implementation
```

All temporary T11 work files are removed/absorbed before integration. T11 closure still does **not** authorize Product implementation; T12, Whole-R10 coherence, fresh independent challenge and explicit operator implementation authorization remain downstream roadmap gates.