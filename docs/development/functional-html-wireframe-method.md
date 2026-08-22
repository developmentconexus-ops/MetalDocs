---
id: frontend-product-experience-planning-method
kind: methodology
owner: development
summary: Reusable authority-to-UX planning method for deriving coherent information architecture, screen structure, interaction patterns and functional HTML prototypes before production frontend implementation.
---

# Frontend Product Experience Planning Method

**Version:** 2.0  
**Scope:** reusable across products and repositories  
**Purpose:** make frontend implementation a realization of already-validated product, UX and system decisions instead of a new design/architecture phase performed while coding.

---

## 1. Why this method exists

A frontend can be technically connected to the backend and still be a poor product.

It can also be visually attractive and still be architecturally wrong.

Typical failures include:

```text
screens generated before understanding user goals
navigation shaped by backend nouns instead of user mental models
a table chosen only because data is list-shaped
cards chosen only because they look modern
important information hidden because hierarchy was never studied
screen exists but backend cannot supply required truth
button exists but no accepted operation owns the action
route exists only because the UI wanted one
frontend invents lifecycle/business state
frontend authorization diverges from server authorization
same interaction pattern is implemented several different ways
component abstractions are created before repeated semantics are proven
API is changed merely to make one screen easier
failure/concurrency states appear only during implementation
backend is declared complete before its user-facing journey exists
visual design silently changes information architecture or behavior
LLM implements the product by improvising missing UX decisions in code
wireframe and implementation drift because no vertical trace exists
```

The method exists to stop those failures **before production frontend code is written**.

The target is not merely a collection of wireframes.

The target is a reviewed chain of reasoning:

```text
accepted product/system authority
→ actors + user needs
→ end-to-end user flows
→ frontend coverage
→ information architecture
→ screen/surface inventory
→ reference study
→ competing layout hypotheses
→ operator-reviewed structural wireframes
→ exact screen/backend contracts
→ derived reusable interaction patterns
→ interactive low-fidelity HTML
→ adversarial walkthrough
→ visual-design handoff
→ implementation readiness
```

---

## 2. Core principle

> The frontend is the human-operable projection of the accepted product architecture, but its information architecture and interaction structure must still be deliberately designed for humans.

Backend authority does **not** uniquely determine good UX.

For example, a backend collection does not automatically imply:

```text
data table
```

The same accepted resource set could legitimately be presented as:

```text
table
grid/cards
master-detail
categorized browse view
search-first interface
mixed browse + search
multiple alternate views over the same truth
```

The correct presentation depends on user goals, frequency, scale, recognition needs, comparison needs, information density, context and future product seams.

Therefore:

```text
backend coherence
≠
UX coherence
```

Both must be proven before implementation.

---

## 3. Success condition

Frontend planning is complete only when a future implementer can build the production frontend without inventing material decisions in code.

For every material screen, region and interaction, the team must be able to answer:

```text
who is the user?
what are they trying to accomplish?
why does this screen/region exist?
why is the information organized this way?
why is this presentation pattern appropriate?
what alternatives were considered?
what accepted product capability does it serve?
what backend/system truth supplies it?
what operation/action does it invoke?
where does every required identity come from?
what state is authoritative?
what happens on success?
what happens on material failure?
what screen/state follows?
which reusable interaction pattern does it use, if any?
what may the frontend NOT infer or own?
```

If a material answer is missing, the frontend is **not implementation-ready**.

---

# 4. Method laws

## 4.1 Human needs before screens

Never start with a screen inventory merely because backend capabilities are known.

Start from:

```text
actor
→ context
→ user need / job
→ desired outcome
→ end-to-end flow
```

A user need describes the outcome/problem, not a predetermined interface.

Bad:

```text
"I need a dashboard with cards"
```

Better:

```text
"I need to quickly identify which resources require my attention so I can act without opening every item"
```

Interface form is derived later.

## 4.2 Coverage before layout

Before choosing tables, cards, drawers or page structures, prove every accepted human capability has a coherent frontend home.

```text
accepted capability / human goal
→ semantic owner
→ admitted read/write contracts
→ candidate frontend context
```

Coverage tells us **what must be representable**.

It does not yet tell us **how the screen should look**.

## 4.3 Information architecture before screen composition

Navigation and grouping are product decisions.

Before drawing screens, decide how users should understand and find the product's objects/tasks.

Information architecture must consider:

```text
user mental models
core objects
frequent tasks
relationships
browse hierarchies
search
filters
saved/persistent views when justified
cross-links
global vs contextual navigation
primary vs secondary destinations
future extension seams already accepted by product direction
```

Do not expose backend package/domain topology as navigation merely because it exists.

## 4.4 Reference study before layout commitment

Do not invent every interaction from scratch.

For each material functional block, study relevant mature products and design systems.

References are **evidence**, not authority.

Use them to understand:

```text
information hierarchy
navigation models
density
selection patterns
list vs grid vs master-detail
search/filter behavior
action placement
progressive disclosure
empty/error states
responsive behavior
interaction conventions
```

Never copy visual appearance or foreign product semantics blindly.

## 4.5 Competing hypotheses before commitment

For consequential screens/blocks, produce **2–3 plausible structural alternatives** before choosing one.

Examples:

```text
dense table
visual cards/grid
master-detail
category-first browse
search-first
mixed browse + list
```

Evaluate alternatives against explicit criteria instead of taste.

## 4.6 Screen-by-screen / block-by-block adjudication

Never generate the whole product interface and ask for review afterwards.

Large products are reviewed in bounded blocks.

For each block:

```text
recover local authority
→ review user need/flow
→ study references
→ produce alternatives
→ discuss with operator/product owner
→ select or revise structure
→ record decision
→ only then progress
```

The human operator must be able to **see and discuss each important screen/block before it becomes the baseline**.

## 4.7 No all-at-once wireframing

The following workflow is prohibited for non-trivial products:

```text
screen inventory
→ generate every screen in one pass
→ review finished prototype afterwards
```

Why:

```text
weak early assumptions propagate to every later screen
component vocabulary freezes prematurely
poor navigation becomes expensive to unwind
screens become internally consistent with the wrong structure
reviewer cognitive load becomes too high
```

## 4.8 Hard no screen-shaped API

A screen being inconvenient to implement is not authority to create a backend operation.

If a needed fact/action is absent:

```text
prove the user need
→ identify the missing semantic truth
→ find its accepted owner
→ classify whether authority already exists
→ reopen only the smallest owning decision when evidence demands it
```

Never repair layout convenience with an arbitrary endpoint.

## 4.9 Frontend never becomes parallel business authority

Unless explicitly accepted by product architecture, frontend planning must not create:

```text
client lifecycle state machine
client authorization evaluator
parallel DTO/schema registry
parallel normalized business-entity truth store
provider mechanism state as product truth
history/audit projection as current-resource truth
optimistic fabrication of consequential business state
```

## 4.10 Patterns are derived, not invented upfront

Do not start by declaring a universal component library for a product that has not been wireframed.

First observe repeated interaction semantics across reviewed screens.

Then consolidate them.

```text
repeated protected behavior
→ shared pattern

cosmetic similarity only
→ not enough evidence for semantic abstraction
```

This avoids both duplication and premature abstraction.

## 4.11 Visual design cannot silently redesign UX

After structural wireframes are accepted, visual design may refine:

```text
brand palette
typography
spacing rhythm
radii
shadows
icons
visual hierarchy emphasis
responsive polish
microinteraction polish
```

A visual-design proposal that changes any of these is a UX/contract finding and must return to the appropriate stage:

```text
information architecture
navigation
screen composition
material region priority
action semantics
required data
workflow
state model
```

---

# 5. Evidence posture

The method distinguishes four kinds of input.

## 5.1 Accepted authority

Binding product/system decisions such as:

```text
product boundary
user capabilities
semantic/domain ownership
identity rules
lifecycle/state
permissions/disclosure
API/wire contract
persistence/concurrency meaning
runtime/external boundaries
```

Frontend planning may derive from these but cannot casually supersede them.

## 5.2 User evidence

Evidence about real user goals and behavior:

```text
interviews
observation
existing workflows
analytics/search logs
support tickets
operational reports
known user feedback
validated domain/operator experience
```

Where direct user research is not yet available, assumptions must be explicitly labeled as assumptions and tested during prototype review.

## 5.3 Reference evidence

External products, SaaS systems, enterprise systems and design systems used to study established patterns.

They inform alternatives but never become product authority.

## 5.4 Operator adjudication

A human product/operator decision after reviewing evidence and candidate layouts.

The operator does not replace user evidence, but owns the explicit product decision when the planning process requires a choice.

---

# 6. Decision status vocabulary

Use a small explicit vocabulary through the planning process.

```text
LOCKED
  accepted for the current planning baseline

CANDIDATE
  plausible leading design; not yet approved

FINDING
  material unresolved question or contradiction

REJECTED
  alternative considered and deliberately not selected

DEFERRED
  known future seam with no current consumer; not implemented now
```

Never write a `CANDIDATE` as though it were accepted authority.

---

# 7. Phase P0 — Recover accepted authority

## Goal

Recover only the smallest authority pack required to plan the current frontend problem.

Collect, where applicable:

```text
product mission / scope
actors
human capabilities
semantic/business owners
identity/data semantics
state/lifecycle rules
permissions/disclosure
backend/module boundaries
API operations/read models
concurrency/idempotency
external dependency semantics
accepted frontend route/lens constraints
proof/validation obligations
```

Do not recursively ingest historical documentation.

## Exit

Every known frontend requirement can be traced to current authority or is explicitly classified as an unknown/assumption.

---

# 8. Phase P1 — Actors, jobs and user needs

## Goal

Understand what people need to accomplish before deciding screens.

For each human actor capture:

```text
actor / role context
trigger
current situation
need / job-to-be-done
desired outcome
frequency
urgency
information needed to decide
common failure/friction
handoffs to other actors/systems
```

Recommended need format:

```text
When <context/trigger>,
I need to <goal>,
so that <outcome>.
```

Do not confuse access roles with personas unless evidence shows they map cleanly.

Do not create screens for machine actors unless a real human operational need exists.

## Exit

Human goals are written independently from proposed pages/components.

---

# 9. Phase P2 — End-to-end user-flow inventory

## Goal

Describe how users accomplish goals across the whole product.

For each goal:

```text
entry context
→ understand state
→ decision
→ action
→ system response
→ handoff if any
→ outcome
→ next likely task
```

Capture alternate/failure branches where they materially change safe behavior.

A user flow is not a route graph and not a component diagram.

## Flow completeness rule

Do not split implementation planning at a point that leaves an accepted human goal unfinished.

If the user goal is:

```text
A → B → C → D
```

an implementation slice must not call itself complete at B if C/D are required for the real outcome and B exposes a dead target.

## Exit

Every accepted human goal has at least one complete end-to-end flow.

---

# 10. Phase P3 — Frontend Coverage Matrix

## Goal

Prove that accepted product/system semantics reach the frontend before layout work begins.

Minimum matrix:

| Capability / user need | Owner | User flow | Candidate frontend context | Reads | Writes | Access/security | UX obligations | Status |
|---|---|---|---|---|---|---|---|---|

Cross-cutting system invariants must also become UX obligations.

Examples:

```text
unknown != known-empty
accepted != applied != converged
projection != mutation authority
hidden control != authorization
ambiguous outcome != known failure
stale write != silent overwrite
external provider success != product success
```

## Exit

```text
all human-facing accepted capabilities mapped
all backend/application operations with human consumers have a frontend home
no invented capability
material findings explicitly named
```

Do not proceed to structural layout for a materially blocked capability.

---

# 11. Phase P4 — Information Architecture

## Goal

Design how users find, understand and move among the product's information and tasks.

This phase happens **before final screen inventory and layout**.

## 11.1 Object/task inventory

List the concepts users must recognize.

For each:

```text
name in user language
why users care
primary actions
relationships to other objects
frequency of use
whether it is browse/search/filter target
whether it deserves independent navigation
```

Do not automatically mirror domain aggregates, database entities or API resource families.

## 11.2 Navigation model

Evaluate:

```text
global navigation
local/context navigation
home/default landing
recents/favorites when evidence justifies them
primary work queue
browse hierarchy
search entry
cross-context links
breadcrumbs where hierarchy genuinely exists
```

Navigation links are appropriate when users repeatedly perform multiple unordered tasks. Sequential journeys should favor task/journey guidance rather than pretending everything is peer navigation.

## 11.3 Findability strategy

For every important collection decide how users should find an item:

```text
browse
search
filter
group
sort
alternate view
saved view
recent context
```

Do not add all mechanisms by default.

Choose only those supported by user need/scale.

## 11.4 Mental-model validation

When IA is uncertain or high-impact, use lightweight validation such as:

```text
card sorting
category naming review
tree testing
first-click testing
operator/domain walkthrough
```

The exact research technique depends on project maturity and access to users.

## 11.5 Future seams

Future known product concepts may influence extensibility of the IA, but they must not become live screens/routes without current Product authority.

Correct:

```text
leave a navigation grouping extensible
```

Incorrect:

```text
add empty future modules because they may exist later
```

## Exit

The team has a reviewed model for how users understand/find product contexts, independent of final visual design.

---

# 12. Phase P5 — Screen and material-surface inventory

## Goal

Derive screens from user flows + IA, not from backend endpoints.

Distinguish:

```text
route/page
material surface within a route
drawer/modal
inline composition region
alternate view of the same collection
material state variant
```

Create a separate material surface when one of these changes:

```text
primary semantic truth
safe user action
write owner
identity required
concurrency/idempotency behavior
exact-content behavior
security/disclosure context
recovery path
editor/viewer mode
```

Do not create a separate screen merely because:

```text
loading style differs
spacing differs
component file differs
responsive arrangement differs
cosmetic empty state differs
```

## Exit

Every human-flow step and material decision context has a candidate screen/surface home.

---

# 13. Phase P6 — Reference Study

## Goal

Study how mature products solve the **same UX problem** before committing to a layout.

Reference study is performed per functional block, not once for the whole product.

## 13.1 Choose references intentionally

Select approximately 3–6 useful references when available, such as:

```text
direct competitors
adjacent enterprise/SaaS products
high-quality products with analogous jobs
design systems with relevant task patterns
platform conventions already familiar to target users
```

Do not optimize for visual similarity.

Optimize for similarity of **user task**.

## 13.2 Analyze patterns, not screenshots

For each reference record:

```text
what user problem it appears to solve
navigation model
information hierarchy
primary vs secondary actions
collection representation
search/filter strategy
list/grid/master-detail usage
progressive disclosure
selection/action behavior
empty/loading/error patterns
responsive behavior
density
what appears strong
what would be wrong for our product
```

## 13.3 Evidence matrix

Recommended form:

| Reference | Relevant job | Pattern observed | Benefit | Risk / mismatch | Candidate lesson |
|---|---|---|---|---|---|

## 13.4 Reference discipline

Do not say:

```text
"Product X uses cards, therefore we use cards."
```

Say:

```text
"Product X uses cards because recognition/preview dominates comparison. Our users have a similar/different need, therefore this pattern is/is not a useful hypothesis."
```

## Exit

The team has evidence-informed structural possibilities for the current block.

---

# 14. Phase P7 — Competing Layout Hypotheses

## Goal

Create and compare plausible structures before choosing one.

For a consequential block, propose 2–3 alternatives whenever real ambiguity exists.

Examples:

```text
A — dense table
B — visual cards/grid
C — master-detail
D — category-first browse + list
```

Do not manufacture alternatives when there is genuinely only one reasonable conventional pattern.

## 14.1 Evaluation criteria

Use criteria relevant to the block, commonly:

```text
task completion speed
recognition
comparison
scanability
information density
cognitive load
frequency of action
scale / number of records
preview needs
bulk action needs
context preservation
error recovery
accessibility
responsive viability
future accepted extensibility
implementation complexity
backend truth fit
```

Recommended comparison:

| Criterion | Hypothesis A | Hypothesis B | Hypothesis C |
|---|---|---|---|
| Recognition | | | |
| Comparison | | | |
| Scale | | | |
| Density | | | |
| Task speed | | | |
| Context | | | |
| Accessibility | | | |
| Backend fit | | | |

## 14.2 Example decision rule: table vs cards

Use a table as a leading hypothesis when users primarily need to:

```text
compare several attributes across many rows
scan structured records quickly
sort/filter by column-like data
perform repeated row/batch actions
find a specific record in a dense collection
```

Use cards/grid as a leading hypothesis when users primarily need:

```text
recognize items visually
consume previews/thumbnails
scan a small number of heterogeneous attributes
choose among items where identity/summary matters more than cross-row comparison
```

Use alternate list/grid views only when both tasks are genuinely important, not because a switcher is fashionable.

## Exit

One candidate is selected for structural wireframing, or the block remains a named finding.

---

# 15. Phase P8 — Block-by-block Structural Wireframing

## Goal

Turn the selected layout hypothesis into a low-fidelity structural design.

This phase freezes **structure**, not brand styling.

Structural wireframes decide:

```text
screen hierarchy
major regions
relative width/height
primary reading order
navigation placement
information grouping
card/table/list/master-detail choice
density
primary/secondary action placement
progressive disclosure
dialog/drawer vs inline editing
empty/error/conflict region placement
responsive transformation rules
```

## 15.1 Review one block at a time

Recommended block cycle:

```text
BLOCK INPUT
  user goals + IA + coverage + local authority

REFERENCE STUDY

LAYOUT HYPOTHESES

CANDIDATE STRUCTURAL WIREFRAME

OPERATOR WALKTHROUGH
  discuss layout, hierarchy, position, size, discoverability, task flow

FINDINGS / REVISION

LOCK or continue iterating
```

Do not start the next high-impact block until the current block is sufficiently coherent to serve as context.

## 15.2 Operator review questions

For each screen/block ask:

```text
Can I find what I came here to do?
Is the most important information visible first?
Does the screen expose too much at once?
Do I understand where I am?
Do I understand where I can go next?
Are primary actions obvious without being dangerous?
Would I rather compare, recognize or preview these items?
Does the information density match the task?
Would I expect this region to be a page, drawer or inline panel?
Do I need context from the previous screen while acting?
Would a real user repeatedly open records just to find the right one?
Does anything feel like backend structure leaking into the UI?
```

## Exit

The current block is explicitly `LOCKED`, or named findings remain open.

---

# 16. Phase P9 — Screen Contract and vertical backend trace

## Goal

After a screen structure is coherent, prove every material region/control is actually realizable by accepted system authority.

Every material Screen Contract answers, where relevant:

```text
GOAL / USER FLOW
ROUTE / SURFACE
INFORMATION HIERARCHY ROLE
OWNER + READ TRUTH
WRITE CONTROLS
IDENTITY SOURCE
CLIENT STATE CLASS
WIRE MECHANICS
MATERIAL STATES / FAILURES
SUCCESS CONSEQUENCE
AUTHZ / DISCLOSURE
PROOF
FORBIDDEN FRONTEND AUTHORITY
BACKEND SUFFICIENCY
```

## Bidirectional law

Trace both directions:

```text
Product/backend → frontend
capability → owner → operation/read model → screen/region/control
```

and:

```text
Frontend → Product/backend
screen/control → operation/read truth → owner → accepted capability
```

A one-way trace can hide orphan operations or invented UI behavior.

## Exit

Every material screen/control is `READY` or blocked by an explicit finding.

---

# 17. Phase P10 — Derive Component and Interaction Pattern Vocabulary

## Goal

Consolidate repeated semantics after enough screens have been reviewed to provide evidence.

Typical patterns may emerge such as:

```text
AppShell
PageHeader
FilterBar
DataTable
CardGrid
MasterDetail
StatusBadge
FormField
ActionMenu
Drawer
Modal
ConfirmDialog
EmptyState
DeniedState
ConflictReconciliation
IdempotentRetry
UploadProgress
InlineViewer
EditableRegion
Timeline
Pagination / LoadMore
```

These names are examples, not mandatory taxonomy.

## Pattern creation rule

Create a shared pattern when multiple accepted screens share:

```text
same purpose
same state ownership
same protected interaction semantics
same accessibility behavior class
same failure/recovery behavior
```

Do not create a shared pattern only because two boxes look alike.

## Pattern entry

For each pattern record:

```text
purpose
screens using it
required inputs
states/variants
material interactions
accessibility expectation
responsive behavior
what it does NOT own
```

## Refactoring rule

When a later screen reveals that an earlier pattern was too specific or too generic, refine the vocabulary deliberately.

Do not preserve a bad abstraction for consistency.

## Exit

Repeated semantics are represented by a bounded reusable vocabulary with no unexplained duplicate pattern.

---

# 18. Phase P11 — Interactive Low-Fidelity HTML Prototype

## Goal

Create a navigable representation of the **already-reviewed** structures.

HTML is not where layout exploration starts.

It is where selected structural decisions become realistically reviewable together.

## 18.1 Technical baseline

Default to the smallest static technology sufficient for realistic interaction:

```text
HTML
CSS
vanilla JavaScript
local deterministic fixtures/state simulation
```

Avoid the eventual production framework unless plain browser technology is genuinely insufficient.

Prototype code is evidence, not production code.

## 18.2 What the prototype must preserve

```text
accepted navigation
reviewed screen hierarchy
reviewed relative layout
material fields/regions
action placement
component-pattern identity
material dialogs/drawers
negative/recovery states
read-only vs editable regions
responsive structural rules
```

## 18.3 Material interaction

Buttons, links, tabs, forms, drawers and dialogs must work when they represent a material state/navigation change.

No decorative dead controls.

## 18.4 Fixture simulation

Use deterministic scenarios to inspect states such as:

```text
normal success
empty
permission denied
not found / non-disclosable
stale concurrency
ambiguous command outcome
dependency unavailable
upload/admission failure
integrity failure
lifecycle transition
```

Fixture state never becomes product authority.

## 18.5 Machine-readable trace

Material controls should carry prototype metadata where useful:

```html
<button
  data-surface="..."
  data-owner="..."
  data-operation="..."
  data-pattern="..."
  data-id-source="..."
  data-concurrency="..."
  data-idempotency="..."
>
  Action
</button>
```

This helps human reviewers and LLM implementers connect visual structure to accepted contracts.

## 18.6 Prototype design boundary

The HTML SHOULD freeze:

```text
layout structure
relative region size
information hierarchy
interaction placement
patterns
states
navigation
```

It SHOULD NOT freeze unless materially necessary:

```text
brand palette
final font
pixel-perfect spacing
ornamental shadows
final icon set
micro-animation
```

## Exit

A reviewer can navigate the approved flows and exercise material interactions without the prototype inventing Product authority.

---

# 19. Phase P12 — Adversarial UX + Architecture Walkthrough

## Goal

Attack the prototype as both a user experience and a system realization.

Review from these roles:

```text
target user
product owner
principal product designer
information architect
senior frontend architect
backend/domain owner
accessibility reviewer
adversarial reviewer
```

## UX attack questions

```text
Can users find the right place without knowing backend terminology?
Are frequent tasks unnecessarily deep?
Are important decisions hidden by progressive disclosure?
Are lists too dense or too sparse?
Does a card/table choice match the actual task?
Does the interface force memory instead of recognition?
Do actions preserve necessary context?
Are similar interactions inconsistent without reason?
Are screens optimized for a demo rather than repeated real use?
```

## Architecture attack questions

```text
Does every material field have a source?
Does every write have an accepted owner/operation?
Does any fixture state accidentally become Product truth?
Does any screen imply a capability the backend does not own?
Did a convenience endpoint sneak in?
Does UI visibility pretend to authorize?
Are concurrency/idempotency/recovery semantics preserved?
Does navigation depend on unavailable IDs?
```

## Finding classes

```text
UX-LAYOUT
  hierarchy/position/density/presentation problem; authority unchanged

IA
  grouping/findability/navigation problem

PATTERN
  duplicate/inconsistent/premature component pattern

SCREEN-CONTRACT
  screen behavior/trace incomplete but upstream product remains sufficient

UPSTREAM
  accepted product/API/architecture lacks truth required by a proven user need

VISUAL-DESIGN
  purely aesthetic refinement; does not block structural readiness
```

## Exit

No unresolved material UX/IA/contract contradiction remains in the reviewed blocks.

---

# 20. Phase P13 — Visual Design Handoff Contract

## Goal

Allow visual design to improve the product without silently changing accepted structure.

Handoff includes:

```text
locked IA
locked user flows
locked structural wireframes
interaction pattern vocabulary
functional HTML prototype
Screen Contracts
material state inventory
responsive structure
```

Visual design is free to explore appearance but must raise a finding when it discovers that good design requires a structural UX change.

## Visual design may change

```text
color
typography
spacing scale
radius
shadow
iconography
visual tone
motion polish
illustration
fine responsive styling
```

## Visual design must not silently change

```text
navigation model
information architecture
material fields
business actions
action consequences
backend operation mapping
state semantics
screen hierarchy that affects task meaning
```

---

# 21. Phase P14 — Frontend Implementation-Readiness Closure

## Goal

Prove that implementation can be realization rather than rediscovery.

Close only when the project-specific counts reconcile exactly.

Generic closure requirements:

```text
accepted human goals                    complete
end-to-end flows                        complete
information architecture               adjudicated
material screens/surfaces               complete
reference study                         complete for material ambiguous blocks
layout hypotheses                       adjudicated
structural wireframes                   locked
Screen Contracts                        complete
material controls                       100% bound
navigation identities                   100% sourced
component/interaction patterns          derived and reviewed
interactive HTML                        complete for accepted scope
negative/material states                represented
frontend ↔ backend trace                complete
orphan backend human operations         0
invented frontend Product operations    0
screen-shaped APIs                      0
unresolved material UX findings         0
unresolved architecture findings        0
```

Product-specific invariants such as operation counts, concurrency domains or exact-file surfaces are added by the consuming repository.

---

# 22. Block operating protocol

For a non-trivial product, maintain a block ledger.

Example generic sequence:

```text
B01 — App shell + global IA
B02 — primary discovery/browse
B03 — primary resource detail
B04 — authoring/editing
B05 — personal work/queue
B06 — decision/governance
B07 — history/evidence
B08 — administration
```

Actual blocks are product-derived; these names are examples.

For each block:

| Field | Required |
|---|---|
| Block | stable planning ID |
| User goals | explicit |
| Authority pack | bounded |
| Reference study | complete when material |
| Hypotheses | 1–3 depending on ambiguity |
| Leading candidate | named |
| Operator review | date/result |
| Findings | explicit |
| Decision | LOCKED / CANDIDATE / FINDING |
| Screen Contracts | ready after structure |
| HTML realization | only after lock |

The next high-impact block should not proceed while the previous block has unresolved structural findings that materially affect global IA/patterns.

---

# 23. Research and reference discipline

This method is intentionally compatible with iterative product discovery.

Use external research to challenge layout assumptions, not to outsource product decisions.

Useful research questions include:

```text
How do mature products present this kind of collection?
When do enterprise systems prefer table vs grid?
How do users navigate similar object hierarchies?
How is context preserved during editing?
How are high-risk actions disclosed?
How do established design systems distinguish tabs, content switchers, accordions and navigation?
How do comparable products surface search/filter/saved views?
```

For each external conclusion separate:

```text
SOURCE OBSERVATION
what the reference actually does/recommends

INFERENCE
why it may apply here

PRODUCT DECISION
what we choose after considering our users/authority
```

Never collapse these three into one statement.

---

# 24. Accessibility is structural, not polish

Accessibility enters during layout and interaction planning, not after visual design.

Wireframes/prototypes must consider:

```text
keyboard navigation
semantic control choice
focus order
heading hierarchy
labels/instructions
error association
contrast-dependent meaning avoided at wireframe level
minimum interactive target viability
responsive reflow
screen-reader comprehensibility of tables/forms/dialogs
non-drag alternatives for essential interactions
```

A layout that cannot reasonably become accessible is not a valid candidate simply because it looks efficient.

---

# 25. Responsive planning

Do not defer responsive behavior to production CSS.

Structural wireframes should define what happens when width changes:

```text
what remains primary
what stacks
what collapses
what becomes a drawer/menu
what becomes locally scrollable
what must stay visible
whether table transforms or remains scrollable
whether cards change columns
```

Responsive transformation must not change Product semantics or hide the only path to a material action.

---

# 26. Density and progressive disclosure

Information density is a user-task decision.

Use higher density when users need rapid scanning/comparison across many records.

Use more whitespace/preview when recognition and comprehension dominate.

Progressive disclosure is appropriate when secondary information would otherwise overwhelm the primary task.

Do not hide information that users need to make the current decision merely to make the screen look clean.

---

# 27. Search, browse and filters

Do not assume search replaces IA, or IA replaces search.

Choose based on user behavior.

```text
known-item lookup
  search may dominate

exploratory discovery
  browse/grouping may dominate

large structured collection
  filter/sort may dominate

mixed use
  combine carefully
```

Filter controls require human-understandable value sources. A backend parameter accepting an opaque ID is not proof that a good selector exists.

If selector discovery is missing, classify whether the user need is real before requesting new backend support.

---

# 28. Tables, cards, lists and master-detail

These are task patterns, not aesthetics.

## Data table

Strong when:

```text
records share comparable attributes
users scan many rows
users sort/filter repeatedly
precise comparison matters
row/batch actions matter
```

## Card/grid

Strong when:

```text
visual/summary recognition matters
preview/thumbnail is useful
items have less uniform metadata
cross-row comparison is secondary
```

## Structured/contained list

Strong when:

```text
space is limited
items share a simple structure
one main label plus small metadata/action is enough
```

## Master-detail

Strong when:

```text
users inspect many records sequentially
context/list position should remain visible
opening a separate page repeatedly would be expensive
```

## Multiple views

Offer grid/list or equivalent switch only when distinct important user tasks justify both.

Do not add view switchers as generic polish.

---

# 29. Findings and smallest-reopen law

A visual inconvenience is not automatically an upstream defect.

When a screen fails, classify in order:

```text
1. UX/layout issue?
2. IA issue?
3. wrong candidate pattern?
4. incomplete Screen Contract?
5. missing read composition already allowed by authority?
6. genuinely missing Product/API capability?
```

Only item 6 justifies an upstream Product/API reopen.

If a reopen is required:

```text
reopen smallest owner
→ correct authority
→ update coverage/IA/screen contract
→ redraw affected blocks only
→ rerun trace
```

Do not redesign unrelated screens.

---

# 30. Adversarial independent review protocol

Before declaring the methodology or a major frontend plan ready, use an independent reviewer when project governance supports it.

Recommended reviewer stance:

```text
Principal Product Designer
+ Information Architect
+ Senior Frontend Architect
+ adversarial architecture reviewer
```

Review question:

> Does this method/plan leave any material UX, IA, interaction, component-boundary or backend/frontend decision to be improvised during implementation?

Reviewer must specifically attack:

```text
missing user-centered discovery
weak information architecture
premature screen inventory
premature component vocabulary
single-solution wireframing
lack of reference study
poor operator/user feedback loop
all-at-once prototype generation
screen-shaped API risk
frontend authority duplication
component duplication
YAGNI violations
overengineering
unprovable backend-to-screen mappings
missing failure/recovery UX
accessibility deferred as visual polish
responsive behavior deferred to implementation
```

Finding severity:

```text
MATERIAL
  method/plan can plausibly produce a materially wrong product or implementation

IMPORTANT
  meaningful weakness that should be corrected before closure but does not invalidate the whole model

OPTIONAL
  useful refinement; not required for correctness/readiness

UNSUPPORTED PREFERENCE
  reviewer taste without evidence of a protected user/system property
```

Correct only evidence-backed findings.

---

# 31. Reusable review checklist

## User / product

- [ ] Human actors and needs are explicit before screens.
- [ ] User needs describe outcomes, not predetermined UI solutions.
- [ ] End-to-end flows cover the whole accepted human goal.
- [ ] Every accepted human capability has a frontend home.

## Information architecture

- [ ] Navigation reflects user tasks/mental models, not backend topology.
- [ ] Browse/search/filter strategy is deliberate.
- [ ] Important collections have a justified findability model.
- [ ] Future seams do not create present speculative features.

## Reference study

- [ ] Material ambiguous blocks use relevant references.
- [ ] References are analyzed by task/pattern, not copied visually.
- [ ] Source observation, inference and product decision remain distinct.

## Layout

- [ ] Consequential screens considered real alternatives when ambiguous.
- [ ] Table/card/list/master-detail choice is justified by user task.
- [ ] Information hierarchy and density are deliberate.
- [ ] Primary and secondary actions are deliberately placed.
- [ ] Progressive disclosure does not hide decision-critical information.
- [ ] Responsive transformation is defined.

## Human review loop

- [ ] Major screens/blocks are reviewed individually with the operator/product owner.
- [ ] Candidate vs locked decisions are explicit.
- [ ] The whole product was not generated in one unreviewed pass.

## Architecture

- [ ] Every material read has an accepted source.
- [ ] Every material write maps to an accepted owner/operation.
- [ ] Every navigation identity has a source.
- [ ] No client state becomes Product authority.
- [ ] No screen-shaped API was introduced for convenience.
- [ ] Material concurrency/idempotency/recovery semantics are represented.

## Patterns

- [ ] Reusable UI patterns were derived from reviewed repetition.
- [ ] Duplicate semantic patterns have been reconciled.
- [ ] Cosmetic similarity did not force false abstraction.
- [ ] Pattern vocabulary does not become a speculative design-system project.

## Prototype

- [ ] Interactive HTML realizes reviewed structure rather than inventing it.
- [ ] Material controls work.
- [ ] Material negative states are inspectable.
- [ ] Trace metadata exists where it materially improves implementation handoff.
- [ ] Prototype code is not treated as production code.

## Visual design handoff

- [ ] Structural UX baseline is explicit before visual styling.
- [ ] Visual design is free to improve aesthetics.
- [ ] Structural changes discovered by design return to the correct UX stage.

## Closure

- [ ] Frontend ↔ backend trace is complete.
- [ ] No orphan human-facing operations.
- [ ] No invented operations.
- [ ] No unresolved material UX/IA findings.
- [ ] No unresolved material architecture findings.
- [ ] Implementation can proceed without material UX invention in code.

---

# 32. Compact process

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
P8  Structural Wireframe — block-by-block + human adjudication
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

The loop is deliberately iterative:

```text
finding
→ smallest affected phase
→ correction
→ re-review affected block
```

Do not restart the entire method unless the finding actually invalidates global assumptions.

---

# 33. Research basis for the method

The methodology deliberately synthesizes several established principles without making any external product its authority.

Useful reference families include:

```text
GOV.UK Service Manual
  understand users and their whole problem
  prototype before committing to build
  test multiple approaches
  use realistic code prototypes for interaction research

GOV.UK Design System
  choose navigation/patterns according to the user task
  patterns are task solutions, not decorative components

Nielsen Norman Group / established IA practice
  user mental models
  card sorting / tree testing where useful
  recognition over recall
  findability and scanability

Enterprise design systems such as Carbon
  data tables for dense structured comparison
  content switchers for alternate views of the same content
  tabs for distinct related sections
  progressive disclosure for secondary information
```

External reference material informs hypotheses. The consuming product's accepted authority + user evidence + operator adjudication determine the actual frontend.

---

# 34. Final principle

A strong frontend plan should make production coding feel almost boring.

Implementation should mostly be:

```text
realize accepted structure
→ bind generated/accepted contracts
→ implement reviewed interaction patterns
→ prove states and failures
```

not:

```text
invent navigation
→ invent screens
→ invent components
→ discover missing APIs
→ redesign workflows
→ reconcile backend/frontend after the fact
```

> **Frontend implementation readiness means the important product, IA, layout, interaction and system-contract decisions have already been made visibly, reviewed deliberately and traced to evidence.**
