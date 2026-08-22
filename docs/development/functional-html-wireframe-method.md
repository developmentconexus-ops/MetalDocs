---
id: functional-html-wireframe-method
kind: methodology
owner: development
summary: Reusable coverage-first method for deriving implementation-ready frontend prototypes from accepted product, architecture, backend and API authority before production UI implementation.
---

# Functional HTML Wireframe Method

**Version:** 1.0  
**Scope:** reusable across products and repositories  
**Purpose:** make frontend implementation a realization of already-validated product/system decisions instead of a second design phase performed while coding.

---

## 1. Problem this method solves

A frontend can be visually plausible and still be architecturally wrong.

Common failure modes include:

```text
screen exists but backend cannot supply required truth
button exists but no accepted operation owns the action
route exists only because the UI wanted one
frontend invents lifecycle/business state
frontend authorization diverges from server authorization
the same semantic component is implemented several different ways
API is changed merely to make one screen easier
important failure/concurrency states are discovered only during implementation
backend is declared complete before its real user-facing consumer exists
wireframe and implementation drift because the wireframe contains no machine-readable trace
```

The method prevents those failures by deriving the frontend from accepted authority **before production UI code is written** and by ending with a navigable functional HTML prototype whose material controls are traceable to real system contracts.

The goal is not a pretty mockup.

The goal is:

```text
accepted product/system authority
→ complete user-capability coverage
→ implementation-relevant surfaces
→ exact screen contracts
→ navigation/data graph
→ reusable component-pattern vocabulary
→ functional HTML prototype
→ interaction ledger
→ bidirectional frontend↔backend trace
→ adversarial review
→ implementation-ready frontend
```

---

## 2. Core invariant

Frontend planning is complete only when an implementer can build the production frontend without inventing material product behavior, backend relationships, navigation meaning, state authority, component semantics or failure handling while coding.

For every material UI element, the implementation team must be able to answer:

```text
why does this exist?
what accepted user goal does it serve?
what backend/system truth supplies it?
what operation/action does it invoke?
what identity is used and where did that identity come from?
what state is authoritative?
what happens on success?
what happens on material failure?
what screen/state follows?
what may the frontend NOT infer or own?
```

If a material question is unanswered, the frontend is **not implementation-ready**.

---

## 3. Method principles

### 3.1 Coverage before drawing

Do not start by inventing pages.

Start by proving that every accepted user-facing capability and human goal has a coherent frontend home.

```text
accepted capability / journey
→ semantic/business owner
→ admitted read/write contract
→ frontend home
```

Only after coverage is coherent should screen contracts and wireframes be produced.

### 3.2 Bidirectional derivation

Every trace must work in both directions.

```text
backend/product → frontend
accepted capability → owner → operation/read model → screen/control
```

and:

```text
frontend → backend/product
screen/control → operation/read truth → owner → accepted capability
```

A one-way mapping is insufficient because it can hide orphan backend operations or invented frontend behavior.

### 3.3 Hard no screen-shaped API

A screen being inconvenient to implement is not authority to create a new endpoint, DTO, state, permission or backend capability.

When a screen requires information that is not available:

```text
prove the human/product need
→ trace the missing truth to its owner
→ classify whether accepted authority already contains that truth
→ if necessary, reopen the smallest owning contract explicitly
```

Never silently repair the frontend with a convenience API.

### 3.4 Server/business authority stays authoritative

The prototype may simulate states, but it must not create a second semantic authority.

Examples of forbidden frontend authority unless explicitly accepted by the product architecture:

```text
client lifecycle state machine
client authorization evaluator
parallel DTO/schema registry
parallel normalized business-entity truth store
provider/external mechanism state treated as product truth
history/audit projection used as current-resource truth
optimistic business-state fabrication
```

### 3.5 Prepare the seam, not hypothetical capability

Do not add generic directories, workflow engines, component frameworks, state stores, routes or endpoints because they may be useful later.

Every material pattern must have a current consumer.

### 3.6 Implementation nodes close user journeys, not technical layers

A backend slice must not declare itself complete when the user-facing journey it owns still points to a dead or unimplemented frontend target.

When an accepted journey is:

```text
A → B → C → D
```

implementation planning must not create a completion boundary between B and C if that boundary leaves B claiming success while C is required to complete the real user goal.

---

## 4. Required inputs

The method does not require a specific stack, framework or API technology, but it does require an accepted planning baseline.

Before starting, recover the smallest authority pack that defines, where applicable:

```text
product capabilities and human goals
semantic/business ownership
user journeys and lifecycle rules
permissions/authorization/disclosure rules
backend/module boundaries
persistence/concurrency rules
API or application-operation contract
read models / DTO wire shapes
generated-client boundaries if any
frontend route/lens decisions already accepted
runtime/external dependency constraints
validation/proof obligations
```

If these are not yet decided, the wireframe method must not pretend they are.

Unknowns are recorded and routed to the owning planning stage.

---

## 5. Outputs

The complete method produces these logical artifacts. They may be separate files or consolidated according to repository standards.

```text
1. Frontend Coverage Matrix
2. Material Surface Inventory
3. Screen Contracts
4. Navigation / Data Graph
5. Component-Pattern Vocabulary
6. Functional HTML Wireframe Prototype
7. Material Interaction Ledger
8. Bidirectional Frontend ↔ Backend Trace
9. Finding / Reopen Record
10. Frontend Implementation-Readiness Closure
```

The artifacts are not independent authorities. They derive implementation detail from accepted product/system authority.

---

# 6. Phase F0 — Authority recovery

## Goal

Create the bounded evidence set needed to reason about the frontend without rediscovering the whole repository.

For every user-facing capability collect, where applicable:

```text
accepted human goal / journey
semantic/business owner
stable route/lens or navigation meaning
read operations + read models
write operations
identity sources
permission / relationship / disclosure predicates
concurrency / idempotency requirements
exact-content/file behavior
material dependency/failure behavior
proof obligations
```

## Exit condition

The team can trace every known frontend requirement back to accepted source authority, and unknowns are explicitly labeled rather than guessed.

---

# 7. Phase F1 — Frontend Coverage Matrix

## Goal

Prove that the planned frontend represents the product already accepted rather than inventing a new one.

Build a matrix similar to:

| Accepted capability / human goal | Frontend home | Backend/system obligations inherited | Status |
|---|---|---|---|
| capability A | route/surface | owner, reads, writes, state, security | Covered / Wireframe / Finding |

Recommended status vocabulary:

```text
COVERED
  accepted authority provides a coherent frontend home

WIREFRAME OBLIGATION
  backend/product contract is coherent but a material interaction/state must still be drawn

FINDING
  a material product/API/architecture/frontend question is unresolved
```

## Cross-cutting coverage

Also map architecture/system invariants into UX obligations.

Examples:

```text
server authorization is final
→ hidden UI may improve usability but never become security authority

optimistic concurrency exists
→ stale edits must have an explicit reconciliation UX

idempotent command intake exists
→ same logical retry preserves the same key; changed command receives a new key

external effect can be ambiguous
→ UI must not expose blind retry as though failure were known

projection/read model is not source truth
→ list row cannot be carried forward as mutation authority

content integrity is protected
→ partial/corrupt content cannot be rendered as successful authoritative content
```

## Exit condition

```text
all accepted human goals mapped
all material backend capabilities with human consumers have a frontend home
no invented product capability
all findings explicitly classified
```

Do not proceed to drawing while a material coverage finding is unresolved.

---

# 8. Phase F2 — Material Surface Inventory

## Goal

Identify every material user decision context that requires a Screen Contract.

A **surface** is not necessarily a page or URL.

Create a distinct surface when at least one materially changes:

```text
primary semantic truth shown
safe user action
write owner / operation
target identity required
concurrency behavior
idempotency behavior
exact-file/content behavior
lifecycle context
security/disclosure outcome
recovery path
editor/viewer mode
```

Do **not** split merely because:

```text
loading spinner differs
spacing/layout differs
responsive arrangement differs
component file differs
cosmetic empty state differs without semantic consequence
```

One route may contain many surfaces. Several surfaces may be represented in one composed HTML screen.

## Exit condition

Every material human decision state from F1 has exactly one coherent surface home and no surface requires invented backend truth.

---

# 9. Phase F3 — Screen Contracts

## Goal

Freeze the meaning of every material surface before visual implementation begins.

Every Screen Contract must answer the following when relevant:

1. What accepted human goal/journey does this surface serve?
2. Which route/lens contains it?
3. Which semantic/business owner supplies each material block?
4. Which exact read operation/query supplies each block?
5. Which exact write operation/action receives each material control?
6. Which schema/read model supplies displayed fields?
7. Where does every identifier needed for a child action/navigation come from?
8. Which state is server state, URL/navigation state, form draft or ephemeral UI state?
9. Which interactions require concurrency tokens, idempotency keys, CSRF/trust material, cursor rules, file-integrity rules or other wire mechanics?
10. Which business/lifecycle states materially alter safe action or presentation?
11. Which failures change the user’s next safe action?
12. What authoritative query/refetch/navigation follows success?
13. Which permission/relationship/disclosure constraints affect the interaction while remaining server authority?
14. What proof demonstrates the real implementation path?
15. What must the frontend explicitly NOT own or infer?
16. Does any required display/action depend on information absent from accepted backend authority?

Recommended compact contract form:

```text
GOAL / ROUTE
OWNER + READ TRUTH
WRITE CONTROLS
IDENTITY / NAVIGATION
CLIENT STATE
WIRE MECHANICS
MATERIAL STATES / FAILURES
SUCCESS CONSEQUENCE
AUTHZ / DISCLOSURE
PROOF
FORBIDDEN
BACKEND SUFFICIENCY
```

## Exit condition

Every material surface is `READY` or explicitly blocked by a named finding. No blocked surface may proceed to final HTML realization.

---

# 10. Phase F4 — Navigation / Data Graph

## Goal

Prove every meaningful transition between surfaces without relying on magic IDs, client inference or history scans.

For each cross-surface navigation edge record:

```text
source surface
→ source of target identity/reference
→ target route/state
→ target initial read
→ 401/403/404/non-disclosure behavior
→ stale-target behavior
```

Example abstract form:

```text
List row
→ returned resource_id
→ /resource/:resource_id
→ getResource(resource_id)
→ server decides current disclosure
```

## Rules

Target identity must come from admitted current truth or from the authoritative result of an accepted operation.

Do not use as current-resource resolvers unless explicitly accepted:

```text
Audit history
semantic history timeline
DOM state
localStorage
provider identity
projection row cached indefinitely
human-entered opaque IDs
```

## Exit condition

Every link/button/navigation has a provable target identity source and real destination read. There are no dead navigation targets.

---

# 11. Phase F5 — Component-Pattern Vocabulary

## Goal

Prevent duplicated semantics and visually different implementations of the same interaction pattern.

Before building the HTML prototype, define the smallest reusable vocabulary of **behavioral UI patterns** actually required by the surfaces.

Typical examples may include:

```text
AppShell
PageHeader
SectionNav
FilterBar
DataTable
StatusBadge
FormField
SelectField
ActionMenu
Drawer
Modal
ConfirmDialog
EmptyState
DeniedState
NotFoundState
UnavailableState
ConflictReconciliation
IdempotentRetry
UploadProgress
InlineViewer
EditableDocumentRegion
Pagination / LoadMore
Timeline
Toast / non-material notification
```

Names are local conventions, not universal requirements.

## Pattern law

Two surfaces should reuse one pattern when they share the same protected interaction semantics.

Create a distinct pattern only when a real difference exists in at least one of:

```text
data/authority semantics
interaction behavior
accessibility requirement
failure behavior
state ownership
layout responsibility
```

Cosmetic difference alone should normally be a variant, not a new component concept.

## Pattern catalog entry

For each reusable pattern record:

```text
purpose
where used
required inputs
states/variants
material interactions
accessibility expectation
what it does NOT own
```

## Exit condition

Every repeated material interaction points to a shared pattern or has a documented reason why reuse would be semantically wrong.

---

# 12. Phase F6 — Functional HTML Wireframe Prototype

## Goal

Create a navigable, executable representation of the accepted frontend behavior **without production framework code**.

The prototype is the visual implementation contract.

It should allow reviewers to use the product flow before the product frontend exists.

## 12.1 Technical baseline

Default to the smallest static implementation that supports the required behavior:

```text
HTML
CSS
vanilla JavaScript
local deterministic fixtures/state simulation
```

React/Vue/Svelte or the eventual production framework is intentionally avoided unless the prototype itself has a proven requirement that plain browser technology cannot satisfy.

Reason:

```text
prototype behavior must not accidentally become production implementation
```

The HTML prototype may be one file for a small product or a small static bundle such as:

```text
index.html
wireframe.css
wireframe.js
fixtures.js
```

Choose the smallest structure that remains understandable.

## 12.2 Navigation

The prototype MUST represent every accepted user-facing route/lens and every material navigation edge from F4.

It may use hash/history simulation or another static routing technique.

Prototype navigation must not invent production routes.

## 12.3 State simulation

Use deterministic fixture scenarios to demonstrate material states without a live backend.

Examples:

```text
normal success
known empty
permission denied
not found / non-disclosable
stale concurrency token
ambiguous command outcome
external dependency unavailable
upload/file admission failure
content integrity failure
lifecycle transition
```

Simulation is presentation evidence only. Fixture state is never treated as product authority.

## 12.4 Material interaction

Buttons, links, tabs, dialogs, drawers and forms should be interactive whenever the interaction changes the user's material state or navigation.

Do not leave implementation-ready controls as dead decorative elements.

## 12.5 Machine-readable trace metadata

Material elements SHOULD carry trace metadata so humans and LLMs can map the prototype directly to accepted contracts.

Example:

```html
<button
  data-surface="SURFACE-ID"
  data-owner="semantic-owner"
  data-operation="operationId"
  data-concurrency="resource-etag"
  data-idempotency="required"
>
  Submit
</button>
```

Navigation example:

```html
<a
  data-from="LIST-SURFACE"
  data-to="DETAIL-SURFACE"
  data-id-source="ReadModel.resource_id"
  href="#/resources/example-id"
>
  Open
</a>
```

Useful metadata categories:

```text
data-surface
data-owner
data-operation
data-read
data-id-source
data-target
data-concurrency
data-idempotency
data-content-mode
data-proof
data-pattern
```

These attributes are prototype trace aids; production code is not required to preserve them.

## 12.6 Semantic HTML and accessibility

Even low-fidelity prototypes should use meaningful browser semantics:

```text
button for actions
a for navigation
label/input association
native form controls where practical
table semantics for real tabular data
heading hierarchy
keyboard-reachable dialogs/drawers where simulated
```

Do not use visual convenience that would hide whether the intended production interaction is actually coherent.

## 12.7 Design boundary

The functional prototype freezes interaction meaning, not final visual design.

It SHOULD freeze:

```text
screen hierarchy
material regions
fields/data blocks
actions
navigation
forms/dialogs/drawers
component-pattern identity
material status/failure/conflict states
read-only vs editable regions
```

It SHOULD NOT freeze unless materially required:

```text
final brand palette
exact typography
pixel-perfect spacing
radii/shadows
final iconography
micro-animation
ornamental illustration
```

## Exit condition

A reviewer can navigate every accepted human flow and exercise every material state/control without encountering a dead element or behavior that requires an invented backend contract.

---

# 13. Phase F7 — Material Interaction Ledger

## Goal

Bind every material prototype control to the system contract that makes it real.

Recommended ledger columns:

| Field | Required meaning |
|---|---|
| Surface/control | exact prototype context and control |
| Pattern | reusable UI pattern used |
| Owner | semantic/business owner of the action/truth |
| Operation/action | exact operation or admitted external/browser mechanism |
| Input source | form/current read/route identity/local exact bytes/etc. |
| Wire mechanics | concurrency/idempotency/trust/file/cursor requirements |
| Success truth | authoritative result/read that controls presentation |
| Material failures | failures that change the next safe UX |
| Client consequence | cache/refetch/local-input/navigation behavior |
| Retry law | same logical command vs new command |
| Forbidden inference | what the client must not manufacture |

## Rule

There must be no material prototype button whose implementation contract is “figure it out later.”

## Exit condition

Every material control is bound to an accepted action/read or explicitly identified as a presentation-only local interaction.

---

# 14. Phase F8 — Bidirectional Frontend ↔ Backend Trace

## Backend/product → frontend

Prove:

```text
every accepted human goal has a frontend home
every user-facing application operation has a consumer context
every material read model has a surface
every material concurrency/idempotency/file behavior has a visible interaction home
every material backend failure has an appropriate safe UX where user-visible
```

## Frontend → backend/product

Prove:

```text
every surface traces to an accepted human goal
every material data block traces to accepted read truth
every material write control traces to one accepted owner/action
every target identity traces to server-returned/current truth
every client state belongs to an accepted state class
no frontend-created product lifecycle/authorization truth
```

## Exit condition

```text
orphan accepted human operations = 0
invented frontend product operations = 0
unbound material controls = 0
unresolved navigation identities = 0
unexplained duplicated semantic patterns = 0
```

Exact operation counts are product-specific and must be reconciled against that product's accepted census rather than copied from another repository.

---

# 15. Phase F9 — Finding classification and reopen law

The method is intended to discover contradictions **before implementation**.

Do not treat findings as failure of the method. Finding them early is the purpose.

Classify each finding before changing authority.

```text
presentation/layout issue only
→ correct prototype/screen contract

missing trace precision but backend truth already exists
→ correct Screen Contract / navigation trace

prototype duplicated a semantic pattern
→ consolidate component vocabulary

accepted frontend mapping is contradictory
→ reopen the smallest frontend planning authority

accepted human goal cannot be represented from current backend/API truth
→ reopen the smallest Product/API/backend owner required to supply that truth

new desired capability discovered during design
→ Product decision/reopen; never smuggle through UI

convenience endpoint with no accepted product need
→ reject and redesign against current truth

scale/performance evidence invalidates a previously bounded projection
→ reopen only that projection/contract with evidence
```

### Hard stops

```text
never invent a new backend operation silently
never add a generic endpoint because HTML wants a selector
never infer permission from navigation visibility
never create a second lifecycle state machine in JavaScript
never use history/audit as current truth merely because it contains an id
never keep a dead button as placeholder in an implementation-ready prototype
```

---

# 16. Phase F10 — Adversarial visual walkthrough

Before declaring readiness, review the prototype as a hostile user/implementer rather than as its author.

Walk at least these classes where relevant:

```text
fresh direct navigation to every route
no-data / empty result
unauthenticated session
permission denied
resource absent/non-disclosable
stale edit/concurrency conflict
ambiguous retry
external dependency unavailable
file/upload failure
state changes between list and detail navigation
state changes between form load and submit
browser refresh on nested/material state
back/forward navigation
long text / large table / pagination
admin vs ordinary-user visibility
read-only vs editable content
```

Ask repeatedly:

```text
Does this screen know something the backend never said?
Can this button execute without an accepted operation?
Can this route obtain every id it needs?
Would two developers reasonably create two different components for the same pattern?
Is any current truth being reconstructed from stale projection/history?
Is any error collapsed into a generic state that changes the user's safe next action?
```

All material findings must be resolved or routed before closure.

---

# 17. Frontend Implementation-Readiness closure

The frontend is implementation-ready only when all product-relevant checks are true.

Generic closure contract:

```text
accepted human goals                     complete
capability coverage                       complete
material surface inventory                complete
Screen Contracts                          complete
Navigation/Data Graph                     complete
component-pattern vocabulary              complete
functional HTML prototype                 complete and navigable
material prototype states                 represented
Material Interaction Ledger               complete
backend→frontend trace                     complete
frontend→backend trace                     complete
unbound material controls                 0
unresolved navigation identities          0
invented Product routes                   0
invented Product operations               0
unexplained duplicate semantic patterns   0
frontend semantic authority added         0 unless explicitly accepted
unresolved MATERIAL findings              0
```

Product-specific censuses should be added to this closure, for example:

```text
accepted operations x/x
concurrency domains x/x
idempotent commands x/x
exact-content/file resources x/x
required proof flows x/x
```

Do not copy numbers from another product.

---

# 18. Design handoff after prototype closure

Once the functional HTML wireframe is accepted, visual design may refine:

```text
brand
palette
typography
spacing
density
icons
radius/shadow
visual hierarchy
responsive presentation
micro-interactions
```

Visual design must not silently change:

```text
screen meaning
accepted fields/data blocks
business actions
operation ownership
navigation semantics
required states
concurrency/idempotency behavior
security/disclosure meaning
backend data requirements
```

If design discovers that a material semantic change is needed, route it back through the smallest Screen Contract/authority finding instead of treating it as styling.

This keeps:

```text
design = visual realization
not
product/backend redesign hidden inside Figma/HTML
```

---

# 19. Production implementation handoff

The HTML prototype is **not production code**.

Production implementation consumes it as an interaction/component contract.

The implementation plan should map:

```text
HTML pattern → production component
HTML surface → production feature/lens
trace metadata → generated/API client call
fixture scenario → test scenario
wireframe state → server/query/form/ephemeral state classification
```

Production components may be internally refactored, but semantic reuse must be preserved.

When several prototype surfaces use the same behavioral pattern, production code should normally realize one reusable component/pattern rather than independent copies.

A production deviation is valid only when it has a material reason such as:

```text
different accessibility contract
different authority/state ownership
different failure behavior
different performance/rendering boundary
```

“Different developer,” “different page,” or “easier to copy” are not sufficient reasons.

---

# 20. Anti-pattern catalog

## Screen-first design

```text
open design tool
→ invent screens
→ later search for APIs
```

**Reject.** Coverage and authority tracing come first.

## Endpoint-shaped UI

```text
one backend endpoint = one page/component
```

**Reject.** UI surfaces follow human decision contexts, not transport structure.

## Page-shaped API

```text
wireframe needs one convenient payload
→ create /dashboard-page-data
```

**Reject unless a real accepted product/read-model need justifies it.**

## Admin-directory shortcut

Using a privileged administrative directory merely to populate an ordinary user selector can create an invalid permission dependency.

Use a purpose-built least-privilege projection only when the accepted journey actually requires one.

## Generic global client store

Do not introduce a second global server-truth store merely because several screens need the same data.

Preserve the accepted frontend state model.

## Duplicate semantic components

Do not independently implement multiple tables, status systems, conflict dialogs, upload flows or command retry patterns that protect the same semantics.

Use the component vocabulary.

## Pretty but inert prototype

A static screenshot with decorative buttons does not prove implementation readiness.

Material controls must navigate or simulate their accepted state consequences.

## Prototype-as-production

Do not choose the production framework merely to make the wireframe look closer to final code.

The prototype should be cheap to change and safe to discard.

## Missing negative states

A prototype that only demonstrates happy paths is incomplete when safe behavior depends on conflicts, permission, external failure or stale state.

---

# 21. Minimal reusable checklist

Use this when applying the method to another repository.

```text
[ ] recover accepted product/backend/frontend authority
[ ] enumerate accepted human goals/capabilities
[ ] build capability coverage matrix
[ ] translate cross-cutting architecture into UX obligations
[ ] resolve material coverage findings
[ ] enumerate material decision surfaces
[ ] create Screen Contract for every surface
[ ] prove every navigation/data identity edge
[ ] define smallest component-pattern vocabulary
[ ] build functional HTML/CSS/JS prototype
[ ] make every material control interactive or explicitly presentation-only
[ ] add trace metadata to material HTML controls
[ ] simulate material positive + negative states
[ ] bind every material control in Interaction Ledger
[ ] execute backend→frontend trace
[ ] execute frontend→backend trace
[ ] reconcile product-specific operation/concurrency/idempotency/file censuses
[ ] perform adversarial walkthrough
[ ] resolve or route every MATERIAL finding
[ ] freeze functional prototype before visual design
[ ] hand design only ornamental/visual freedom unless a semantic finding is reopened
[ ] map prototype patterns to reusable production components before coding
```

---

# 22. Definition of success

The method succeeds when frontend implementation no longer begins with:

> “What screens should we build and how should they call the backend?”

Instead, implementation begins with:

> “Here is the reviewed functional prototype, here are the accepted component patterns, here is every screen/control/backend trace, and here are the exact states and proof obligations we must realize.”

At that point, production frontend development is primarily **realization**, not architecture discovery.
