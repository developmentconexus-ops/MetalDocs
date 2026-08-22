# T11 — Frontend Implementation Readiness

> **TEMPORARY T11 CANDIDATE COMPANION / BRANCH-ONLY WORK.** This file derives implementation-readiness detail from accepted Product/T6/T8-E/T8-F authority. It is not durable authority and must be absorbed or removed before T11 integration. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose

T8-F already closes frontend semantic realization: accepted human goals, stable route/lens meanings, 78/78 operation consumers, generated transport consumption, TanStack Query/state behavior, ETag/idempotency/Problem handling and editor/viewer boundaries.

T11 must now close a different question before implementation is authorized:

```text
Can an implementer build every material MetalDocs screen and interaction
without inventing Product meaning, guessing a backend relationship,
creating screen-shaped API, or deciding material UX behavior while coding?
```

The answer is acceptable only after a complete frontend implementation-readiness pack exists and every material screen/control can be traced bidirectionally to accepted backend authority.

This is implementation-readiness precision, **not a T8-F reopen**. If the exercise produces material evidence that accepted T8-F/T8-E/Product authority cannot serve a required human goal, the smallest implicated authority reopens explicitly; T11 never patches the gap silently.

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

Stable Product route meanings remain:

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

`GET /auth/login` and `GET /auth/callback` remain browser AuthN integration routes outside the SPA Product tree and outside the 78-operation census.

A route is not assumed to equal one screen. One route may contain multiple material interaction surfaces/states when the user's safe next action materially changes.

## 3. Reusable planning principle from Marketplace Central

The useful reusable idea is **coverage before drawing**, not Marketplace-specific ontology.

MetalDocs therefore uses this order:

```text
accepted Product/human goals
→ accepted semantic owners
→ accepted 78-operation wire/read models
→ T8-F route/lens homes
→ frontend coverage matrix
→ material interaction-surface inventory
→ screen contracts
→ navigation/data graph
→ wireframes
→ material interaction ledger
→ bidirectional frontend↔backend trace
→ coverage findings / smallest justified reopen
→ zero unresolved material frontend gap
```

Marketplace-specific concepts such as Operational Work, external convergence or generic known/partial/unknown state taxonomies are not imported unless MetalDocs authority independently contains them.

## 4. F0 — Authority recovery and coverage baseline

Before screen design begins, establish one bounded source table from current accepted authority containing at least:

```text
human goal / journey
semantic owner(s)
route/lens home
read operationIds
mutation operationIds
read-model/schema authority
permission/relationship/disclosure constraints
ETag/idempotency requirements where reachable
exact-content/editor/viewer requirements where reachable
material T9 browser/composed proof obligations
```

This is a derived implementation-planning table. It does not redefine operation or Product meaning.

F0 fails if a claimed frontend capability has no accepted Product journey/human goal, or if an accepted human goal has no frontend home.

## 5. F1 — MetalDocs Frontend Coverage Matrix

The coverage matrix must reconcile **all accepted human interactions**, not just route count.

Minimum columns:

| Field | Meaning |
|---|---|
| Accepted human goal | exact T6/T8-F user intent being served |
| Semantic owner(s) | authority supplying/receiving meaning |
| Primary route/lens | stable T8-F home |
| Read operations | admitted operationIds/read models feeding the surface |
| Write operations | exact owner operations behind actions |
| Material state/concurrency | lifecycle/OCC/idempotency state that changes safe UX |
| Exact-content boundary | when bytes/editor/viewer behavior matters |
| Browser proof | T9 E4/composed obligation where claim-relevant |
| Coverage status | covered / explicit wireframe obligation / material finding |

The matrix must prove 78-operation frontend coverage against T8-F's accepted consumer mapping while keeping semantic ownership unchanged.

## 6. F2 — Material interaction-surface inventory

A **surface** is a coherent user decision context, not necessarily a URL or React component.

Create a distinct surface/state when at least one of these changes materially:

```text
primary semantic truth shown
available safe user action
mutation owner/operation
required target identity
OCC/idempotency/exact-byte behavior
current lifecycle context
security/disclosure outcome
recovery path after a material failure
editor/viewer mode
```

Do not create separate wireframes merely for cosmetic loading spinners or visually different but semantically identical states.

Initial inventory seed from T8-F route meanings, to be expanded by coverage analysis rather than treated as final screen count:

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

Material dialogs/drawers/forms may remain within a parent surface when their owner/data/action contract is explicit. They become separate surface contracts only when separation materially improves traceability or safe interaction reasoning.

## 7. F3 — Screen Contract

Every material frontend surface MUST have a Screen Contract before its wireframe can be accepted.

For each surface answer all claim-relevant questions:

1. **Which accepted human goal/journey does this surface serve?**
2. **Which stable route/lens contains it?**
3. **Which semantic owner supplies each material block of truth?**
4. **Which exact `operationId` supplies each read?**
5. **Which exact `operationId` receives each write/control?**
6. **Which generated schema/read model supplies the displayed fields?**
7. **Where does every identity needed for a child route/action come from?**
8. **Which state belongs to TanStack Query, URL/router, local form state or ephemeral React state?**
9. **Which interactions require CSRF, `If-Match`, `Idempotency-Key`, cursor semantics or exact-byte handling?**
10. **Which Product/lifecycle states materially alter what the user sees or may safely do?**
11. **Which Problem/precondition/dependency outcomes materially change the user's next safe action?**
12. **What happens to authoritative queries/navigation after each successful mutation?**
13. **Which current permission/relationship/disclosure rules affect the interaction while remaining server authority?**
14. **Which browser/composed proof demonstrates that the surface is connected to the correct backend path?**
15. **What must this surface explicitly NOT own or infer?**
16. **Does any required display/action need information absent from the accepted 78-operation/read-model contract?**

A material unanswered question blocks the wireframe. Do not substitute generic UI convention for missing Product/API authority.

## 8. F4 — Navigation / Data Graph

For every meaningful frontend transition, record:

```text
source surface
→ accepted source of target identity/reference
→ target stable route
→ target initial read operation
→ disclosure/authorization behavior when target is absent/non-disclosable
```

Examples of the type of relation that must be explicit, without creating new authority:

```text
Library item
→ document_id returned by accepted DocumentSummary/read
→ /documents/:document_id
→ getDocument

Document Official with disclosed open_revision
→ disclosure-safe open Revision identity returned by DocumentOfficialView
→ /documents/:document_id/work
→ accepted Document Work reads
```

Navigation may not use History, Audit, provider object identity, DOM state or client-only inference as a substitute for accepted current-resource truth.

If a necessary navigation edge cannot obtain its target identity from accepted truth, record a material coverage finding. Do **not** add a screen-shaped read endpoint locally.

## 9. F5 — Wireframe law

All material MetalDocs screens/surfaces and material state variants must be wireframed before T11 can close.

Wireframes are **functional low/mid-fidelity implementation contracts**, not visual-brand authority.

They MUST make clear when claim-relevant:

```text
page/surface structure and hierarchy
material data blocks and labels
tables/lists/cards/editor/viewer regions
primary and secondary actions
forms/dialogs/drawers that carry semantic writes
navigation destinations
empty/absent/denied states when user behavior differs
material lifecycle states
OCC conflict/reconciliation state
idempotent ambiguous-retry behavior when user-visible
upload/admission/recovery state
exact-content viewer/editor mode
external-dependency failure state when it changes safe action
what is read-only vs actionable
```

They do NOT need to freeze before implementation unless a material requirement depends on them:

```text
exact spacing
final color palette
shadow/radius polish
micro-animation
pixel-perfect brand styling
non-functional decorative choices
```

The rule is: **freeze interaction meaning, not ornamental pixels**.

One wireframe may cover several non-material presentation variants. A separate state/frame is required when a different action could be unsafe if the distinction were hidden.

## 10. F6 — Material Interaction Ledger

Every material control/action in the accepted wireframes must have one ledger row.

Minimum fields:

| Field | Required content |
|---|---|
| Surface/control | exact user control/context |
| Owner | semantic owner receiving the action |
| operationId | exact accepted application operation, or AuthN browser route where applicable |
| Input source | local form, current server representation, exact bytes, route identity, etc. |
| Wire mechanics | CSRF / If-Match / Idempotency-Key / exact byte / cursor as applicable |
| Success truth | authoritative result/read that determines success presentation |
| Material failures | exact Problem/precondition/state classes that change safe UX |
| Client consequence | query replacement/invalidation/refetch, local-input preservation, navigation |
| Retry law | same logical command vs new semantic command behavior where applicable |
| Forbidden inference | authority the client must not manufacture |

Illustrative shape only; exact operation semantics remain T8-E:

```text
Save DRAFT
→ updateRevisionDraft
→ current local DRAFT buffer + loaded exact ETag
→ CSRF + If-Match
→ authoritative returned/current DRAFT truth
→ stale 412 preserves local input + refetch + explicit reconciliation
→ never silent LWW/auto-merge
```

The ledger prevents “button exists but backend behavior will be decided during React implementation”.

## 11. F7 — Bidirectional frontend ↔ backend trace

Closure requires both directions.

### Product/backend → frontend

```text
every accepted human goal has a UX home
every one of 78 application operations has its accepted frontend consumer context
every accepted T8-F read model used by a lens has an identified surface
every material ETag/idempotency/exact-content/browser behavior has a visual/interaction home when user-relevant
```

### Frontend → Product/backend

```text
every screen/surface traces to an accepted human goal
every material data block traces to an accepted read model/operation
every material write control traces to exactly one accepted owner operation
every navigation identity traces to accepted server-returned truth
every material client state is classified by the T8-F state-authority model
zero button or display block depends on invented Product state
```

Final reconciliation must retain:

```text
application operations   78 / 78
orphaned                 0
invented                 0
operation 79             absent
```

## 12. F8 — Coverage findings and reopen law

Every gap discovered while doing coverage/screens/wireframes is classified before correction.

```text
frontend placement/presentation issue only
  → correct inside T11-derived frontend readiness; no Product/T8 reopen

missing trace/documentation precision but accepted backend truth already exists
  → correct the T11 Screen Contract/interaction trace; no semantic reopen

accepted read/model exists but T8-F mapping is materially contradictory
  → STOP affected surface; smallest bounded T8-F reopen

required human goal cannot be represented from accepted wire/read models
  → STOP affected surface; trace to Product journey and smallest Product/T6/T8-E owner

proposed convenience endpoint with no accepted Product need
  → reject; redesign the surface using accepted truth

operation 79 proposal
  → hard STOP; explicit material Product/T6/T8-E reopen required
```

**Hard no screen-shaped API:** wireframe inconvenience, component convenience or desire to avoid client composition never by itself authorizes a new Product operation/read model.

Conversely, if a required accepted human goal genuinely cannot be implemented safely from current backend authority, do not force the frontend to guess. The wireframe becomes material evidence for the smallest bounded reopen.

## 13. F9 — Frontend readiness closure

The T11 frontend pack is complete only when all are true:

```text
accepted human-goal coverage matrix complete
material interaction-surface inventory complete
Screen Contract exists for every material surface
Navigation/Data Graph complete for every material transition
wireframes exist for all material screens and materially different safe-action states
Material Interaction Ledger complete
backend→frontend trace complete
frontend→backend trace complete
78/78 frontend operation coverage reconciled
all material state/concurrency/idempotency/exact-content behavior has an implementation home
no screen requires invented lifecycle/Authorization/server truth
no unresolved MATERIAL coverage finding
operation 79 absent
```

This closure occurs **during T11 planning**, before T12 and before Product implementation authorization.

T12 will then be able to adversarially attack the complete chain:

```text
Product journey
↕
semantic owner
↕
operation / read model
↕
route / screen contract
↕
wireframe / control
↕
mutation / navigation / state behavior
↕
future implementation node
↕
proof obligation
```

## 14. Implementation-node linkage

The T11 Node Completion Contracts must bind each future semantic S tranche to the frontend surfaces assigned by the completed frontend readiness pack.

A future S node cannot close with “backend complete, frontend follows later”. Its relevant accepted screen contracts/wireframes must be implemented against the real generated client and real server path in that same tranche.

P5 later proves that the realized frontend has not drifted from the T11 coverage/screen/wireframe contract.

## 15. Current status of this companion

```text
method                             DEFINED
coverage matrix                    NOT YET DERIVED
material surface inventory        SEEDED ONLY / NOT CLOSED
Screen Contracts                  NOT YET DERIVED
Navigation/Data Graph             NOT YET DERIVED
wireframes                        NOT STARTED
Material Interaction Ledger       NOT YET DERIVED
bidirectional reconciliation      NOT YET EXECUTED
```

This is intentional: the method is frozen before drawing so wireframes are evidence-driven rather than design-first invention.

## 16. Promotion law

This companion is temporary review structure. If T11 converges, its binding implementation-readiness outcome is absorbed into durable T11 authority and the reviewed screen/wireframe pack is retained only in the minimum repository location needed for future implementation/T12 consumption. This temporary method file is removed before integration so it cannot become a competing frontend authority.