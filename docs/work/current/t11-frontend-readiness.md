# T11 — Frontend Implementation Readiness

> **TEMPORARY T11 CANDIDATE COMPANION / BRANCH-ONLY WORK.** This file derives implementation-readiness detail from accepted Product/T6/T8-E/T8-F authority. It is not durable authority and must be absorbed or removed before T11 integration. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose

T8-F already closes frontend semantic realization: accepted human goals, stable route/lens meanings, 78/78 operation consumers, generated transport consumption, TanStack Query/state behavior, ETag/idempotency/Problem handling and editor/viewer boundaries.

T11 closes a different question:

```text
Can an implementer build every material MetalDocs screen/interaction
without inventing Product meaning, guessing a backend relationship,
creating screen-shaped API, or deciding material UX behavior while coding?
```

This is implementation-readiness precision, **not a T8-F reopen**. A material gap that proves accepted authority insufficient routes to the smallest implicated owner explicitly.

## 2. Fixed envelope

```text
stable SPA route meanings             exact accepted T6/T8-F set
application operations                78
orphaned operations                   0
invented operations                   0
operation 79                          absent
frontend semantic owner               none
frontend Authorization engine         absent
parallel DTO/API authority             absent
parallel global server-truth store     absent
React SPA                              accepted
TanStack Query                         accepted server-state mechanism
generated TypeScript wire projection  required
interactive DOCX boundary              one adapter boundary
Product implementation                 BLOCKED
```

Stable Product routes remain:

```text
/documents
/documents/:document_id
/documents/:document_id/work
/documents/:document_id/history
/work
/work/governance/:attempt_id
/audit
/admin/organization
/admin/access
/admin/document-governance
```

`/auth/login` and `/auth/callback` remain browser integration routes outside the 78-operation application census.

Route count is not screen count. One stable route may contain several material interaction states and may be progressively enriched by later T11 implementation tranches where the exact progression is declared.

## 3. Method

```text
F0 accepted authority recovery
→ F1 Frontend Coverage Matrix
→ F2 material interaction-surface inventory
→ F3 Screen Contracts
→ F4 Navigation/Data Graph
→ F5 functional low/mid-fidelity wireframes
→ F6 Material Interaction Ledger
→ F7 bidirectional frontend↔backend trace
→ F8 finding classification / smallest justified reopen
→ F9 frontend readiness closure
```

Reusable idea from Marketplace Central: **coverage before drawing**. Marketplace-specific Product ontology is not imported into MetalDocs.

## 4. F0 — Authority baseline — COMPLETE

The frontend planning input must always retain:

```text
accepted human goal / Product journey
semantic owner(s)
stable route/lens home
read operationIds + read models
mutation operationIds
permission/relationship/disclosure constraints
ETag/idempotency/exact-content requirements
T9 browser/composed proof obligations
```

F0 is represented by current Product/T6/T8-E/T8-F authority and the derived coverage artifact.

## 5. F1 — Frontend Coverage Matrix — COMPLETE

Current artifact:

```text
docs/work/current/t11-frontend-coverage.md
```

Closed result:

```text
accepted human goals mapped                 16 / 16
stable route/lens homes                     covered
T8-F frontend consumer reconciliation       78 / 78
T11 implementation assignment               78 / 78 exactly once
operation 79                                absent
material T11 coverage findings              5 found / 5 adjudicated
unresolved MATERIAL coverage finding        0
Product/T8-F semantic reopen                not justified
```

F1 materially corrected the open T11 graph:

```text
old artificial order
  Organization → Authentication/Authorization

corrected first semantic tranche
  Identity + Organization + Access
```

It also closed dead-target risks by assigning `createDocumentRevision` + `listAuthoringWork` with the real Document Work target and `listGovernanceWork` with the real Governance Case target.

## 6. F2 — Material interaction-surface inventory — NEXT

A **surface** is a coherent user decision context, not necessarily a URL or React component.

Create a distinct material surface/state when at least one changes materially:

```text
primary semantic truth shown
safe user action
mutation owner/operation
required target identity
OCC/idempotency/exact-byte behavior
lifecycle context
security/disclosure outcome
recovery path after material failure
editor/viewer mode
```

Do not split merely for cosmetic loading spinners or visually different but semantically identical presentation.

Seed only — not final inventory:

```text
Application shell / authenticated navigation
Library / official discovery
Document creation interaction
Document Official
Document Work
My Work
Governance Case
Document History
Audit
Admin / Organization
Admin / Access
Admin / Document Governance
```

F2 must expand this seed until every F1 human goal/material state has a home and no surface requires invented backend truth.

## 7. F3 — Screen Contract

Every material surface MUST answer all claim-relevant questions before its wireframe can be accepted:

1. Which accepted human goal/journey does it serve?
2. Which stable route/lens contains it?
3. Which semantic owner supplies each material truth block?
4. Which exact `operationId` supplies each read?
5. Which exact `operationId` receives each write/control?
6. Which generated schema/read model supplies displayed fields?
7. Where does every identity needed for child navigation/action come from?
8. Which state belongs to TanStack Query, URL/router, form state or ephemeral React state?
9. Which interactions require CSRF, `If-Match`, `Idempotency-Key`, cursor semantics or exact-byte handling?
10. Which Product/lifecycle states materially alter safe action/presentation?
11. Which Problem/precondition/dependency outcomes change the next safe action?
12. What authoritative query/navigation consequence follows each successful mutation?
13. Which current permission/relationship/disclosure constraints affect the interaction while remaining server authority?
14. Which browser/composed proof demonstrates the correct backend path?
15. What must the surface explicitly NOT own/infer?
16. Does any required display/action need information absent from accepted 78-operation/read-model authority?

Any material unanswered item blocks the wireframe.

## 8. F4 — Navigation / Data Graph

For every meaningful transition:

```text
source surface
→ accepted source of target identity/reference
→ target stable route
→ target initial read operation
→ absent/non-disclosable/unauthorized behavior
```

Navigation may not use History, Audit, provider identity, DOM state or client-only inference as substitute for accepted current-resource truth.

A missing required target identity is a material finding, not permission for a convenience endpoint.

## 9. F5 — Wireframe law

All material screens/surfaces and material safe-action state variants must be wireframed before T11 closes.

Wireframes are **functional low/mid-fidelity implementation contracts**, not visual-brand authority.

They MUST show where relevant:

```text
page/surface hierarchy
material data blocks/labels
tables/lists/cards/editor/viewer regions
primary/secondary actions
forms/dialogs/drawers carrying semantic writes
navigation destinations
material absent/denied/lifecycle states
OCC conflict/reconciliation
idempotent ambiguous retry when user-visible
upload/admission/recovery
exact-content editor/viewer mode
dependency failure when it changes safe action
read-only vs actionable regions
```

They do not need to freeze ornamental pixel choices such as exact spacing, final palette, shadows/radii or micro-animation unless a material requirement depends on them.

**Freeze interaction meaning, not ornamental pixels.**

## 10. F6 — Material Interaction Ledger

Every material control/action in accepted wireframes gets one ledger row:

| Field | Required content |
|---|---|
| Surface/control | exact user context/control |
| Owner | semantic owner receiving meaning |
| operationId | exact operation or admitted AuthN browser route |
| Input source | form/current representation/exact bytes/route identity |
| Wire mechanics | CSRF / If-Match / Idempotency-Key / exact byte / cursor as applicable |
| Success truth | authoritative result/read controlling presentation |
| Material failures | exact Problem/precondition/state classes changing safe UX |
| Client consequence | query replacement/invalidation/refetch, local input, navigation |
| Retry law | same logical command vs new semantic command |
| Forbidden inference | authority client must not manufacture |

This removes “button exists; API behavior decided later” ambiguity.

## 11. F7 — Bidirectional trace

### Product/backend → frontend

```text
every accepted human goal has UX home
every 78 application operation has accepted frontend consumer context
every material read model has identified surface
every user-relevant ETag/idempotency/exact-content/browser behavior has interaction home
```

### Frontend → Product/backend

```text
every surface traces to accepted human goal
every material data block traces to accepted read truth
every material write control traces to exactly one owner operation
every navigation identity traces to server-returned truth
every client state is classified by T8-F authority model
zero invented Product state
```

Final reconciliation:

```text
application operations   78 / 78
orphaned                 0
invented                 0
operation 79             absent
```

## 12. F8 — Findings / reopen law

```text
frontend placement/presentation issue only
  → correct T11 readiness; no Product/T8 reopen

missing trace precision but accepted backend truth exists
  → correct Screen Contract/trace

accepted frontend mapping materially contradictory
  → smallest bounded T8-F reopen

required accepted human goal not safely representable from current wire/read models
  → smallest Product/T6/T8-E owner reopen

convenience endpoint without accepted Product need
  → reject; redesign with accepted truth

operation 79 proposal
  → hard STOP; explicit material Product/T6/T8-E reopen required
```

**Hard no screen-shaped API.** Wireframe/component convenience never authorizes a new Product operation.

Conversely, a real accepted human goal must not be forced into frontend guessing merely to avoid a justified bounded reopen.

## 13. F9 — Frontend closure

Complete only when:

```text
F1 Coverage Matrix complete
F2 material surface inventory complete
Screen Contract for every material surface
Navigation/Data Graph complete
wireframes for all material screens/safe-action states
Material Interaction Ledger complete
backend→frontend trace complete
frontend→backend trace complete
78/78 frontend operation coverage reconciled
material state/concurrency/idempotency/exact-content behavior has implementation home
no invented lifecycle/Authorization/server truth
no unresolved MATERIAL coverage finding
operation 79 absent
```

This closes **during T11**, before T12 and before Product implementation authorization.

T12 can then attack:

```text
Product journey
↕ semantic owner
↕ operation/read model
↕ route/Screen Contract
↕ wireframe/control
↕ mutation/navigation/state behavior
↕ implementation node
↕ proof obligation
```

## 14. Implementation-node linkage

Future semantic tranches implement the completed T11 frontend pack against real generated client/server paths. No S node may close as “backend complete, frontend follows later”.

Stable route/lenses may be progressively enriched only as named in the Node Completion Contracts; no actionable dead target is allowed.

P5 later proves realized frontend did not drift from reviewed T11 screen/wireframe contracts.

## 15. Current status

```text
F0 authority baseline                  COMPLETE
F1 coverage matrix                     COMPLETE / 5 findings adjudicated
F2 material surface inventory          NEXT / NOT STARTED
F3 Screen Contracts                    NOT STARTED
F4 Navigation/Data Graph               NOT STARTED
F5 wireframes                          NOT STARTED
F6 Material Interaction Ledger         NOT STARTED
F7 bidirectional reconciliation        PARTIAL via F1 / final not executed
F8 finding classification              ACTIVE as evidence appears
F9 frontend readiness closure          NOT REACHED
```

The method remains intentionally ahead of drawing: wireframes are evidence-driven rather than design-first invention.

## 16. Promotion law

This file is temporary review structure. If T11 converges, its binding implementation-readiness outcome is absorbed into durable T11 authority and the reviewed screen/wireframe pack is retained only in the minimum repository location required by T12/future implementation. This temporary method file is removed before integration.