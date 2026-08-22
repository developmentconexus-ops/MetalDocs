---
id: frontend-product-experience-planning-method
kind: methodology
owner: development
summary: Reusable authority-to-UX planning method for deriving coherent information architecture, screen structure, interaction patterns and functional prototypes before production frontend implementation.
---

# Frontend Product Experience Planning Method

**Version:** 2.1  
**Scope:** reusable across products and repositories  
**Purpose:** make frontend implementation a realization of already-reviewed product, UX and system decisions instead of a new design/architecture phase performed while coding.

---

## 1. Why this method exists

A frontend can be technically connected to the backend and still be a poor product. It can also be visually attractive and still be architecturally wrong.

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

The target is not a collection of wireframes. The target is a reviewed chain of reasoning:

```text
accepted product/system authority
→ actors + user needs
→ end-to-end user flows
→ frontend coverage
→ candidate information architecture
→ screen/surface inventory
→ reference study
→ competing layout hypotheses
→ operator-reviewed visual structural wireframes
→ exact screen/backend contracts
→ derived reusable interaction patterns
→ interactive low-fidelity prototype
→ adversarial walkthrough
→ visual-design handoff
→ structural-conformance review
→ implementation readiness
```

---

## 2. Core principle

> The frontend is the human-operable projection of accepted product architecture, but its information architecture and interaction structure must still be deliberately designed for humans.

Backend authority does **not** uniquely determine good UX.

A backend collection may legitimately become a table, grid, structured list, master-detail view, categorized browse view, search-first interface or a justified combination. The choice depends on user goals, frequency, scale, recognition needs, comparison needs, density, preview needs, context and accepted future seams.

Therefore:

```text
backend coherence != UX coherence
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
what alternatives were considered when ambiguity was real?
what accepted product capability does it serve?
what backend/system truth supplies it?
what operation/action does it invoke?
where does every required identity come from?
what state is authoritative?
what happens on success?
what happens on material failure?
what must the user understand after that failure?
what screen/state follows?
which reusable interaction pattern does it use, if any?
what changes responsively?
what accessibility constraints shaped the structure?
what may the frontend NOT infer or own?
```

If a material answer is missing, the frontend is **not implementation-ready**.

---

# 4. Method laws

## 4.1 Human needs before screens

Never start with a screen inventory merely because backend capabilities are known.

```text
actor
→ context
→ user need / job
→ desired outcome
→ end-to-end flow
```

A user need describes an outcome/problem, not a predetermined interface.

## 4.2 Coverage before layout

Before choosing tables, cards, drawers or page structures, prove every accepted human capability has a coherent frontend home.

```text
accepted capability / human goal
→ semantic owner
→ admitted read/write contracts
→ candidate frontend context
```

Coverage tells us **what must be representable**, not how it should look.

## 4.3 Information architecture before screen composition

Navigation and grouping are product decisions. Before drawing screens, create a candidate model for how users understand and find tasks/objects.

Consider:

```text
user mental models
core objects
frequent tasks
relationships
browse hierarchy
search
filters
cross-links
global vs contextual navigation
primary vs secondary destinations
accepted future seams
```

Do not expose backend package/domain topology as navigation merely because it exists.

## 4.4 Reference study before layout commitment

For material ambiguous blocks, study relevant mature products and design systems. References are **evidence**, not authority.

Analyze task patterns, not visual fashion.

## 4.5 Competing hypotheses only when ambiguity is real

For consequential blocks with real structural ambiguity, create 2–3 plausible alternatives and compare them against explicit criteria.

Do **not** manufacture alternatives for ceremony. When one conventional pattern is clearly adequate, record one sentence in the block ledger explaining why no competing hypothesis is needed.

## 4.6 Operator-only LOCK and progression

`LOCKED` is an **operator-only** decision status.

```text
assistant / reviewer / tool
  may propose CANDIDATE, FINDING or REJECTED
  MUST NOT set LOCKED

operator / designated human product decision owner
  is the only actor that may set LOCKED
```

The next material block must not be presented or generated as baseline until the current material block is `LOCKED`.

Exception:

```text
operator explicitly authorizes parallel progression
→ current block remains CANDIDATE
→ dependent decisions may not be treated as locked authority
```

There is no assistant-judged substitute such as “sufficiently coherent.”

## 4.7 No all-at-once wireframing

For non-trivial products, this is prohibited:

```text
screen inventory
→ generate every screen in one pass
→ review finished prototype afterwards
```

The operator must visually review important blocks before downstream structure inherits them.

## 4.8 Hard no screen-shaped API

A screen being inconvenient to implement is not authority to create a backend operation.

```text
prove user need
→ identify missing semantic truth
→ find accepted owner
→ classify whether authority already exists
→ reopen only the smallest owning decision when evidence demands it
```

## 4.9 Frontend never becomes parallel business authority

Unless explicitly accepted, frontend planning must not create:

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

Repeated protected behavior may become a shared pattern. Cosmetic similarity alone is not enough.

Pattern consolidation happens only after reviewed/locked evidence exists.

## 4.11 Visual design cannot silently redesign UX

Visual design may refine palette, typography, spacing, radius, shadow, iconography, tone and motion polish.

A proposal that changes navigation, IA, material fields, action meaning, reading order, region priority, density class or workflow returns to the smallest affected UX stage.

## 4.12 Accessibility and responsive behavior are structural

A candidate cannot be locked if its interaction model has no plausible accessible or responsive realization.

---

# 5. Evidence posture and assumption register

The method distinguishes four inputs.

## 5.1 Accepted authority

Binding product/system decisions: product boundary, capabilities, ownership, lifecycle, identity, permissions/disclosure, API/wire, persistence/concurrency and runtime boundaries.

## 5.2 User evidence

Examples:

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

Where direct evidence is unavailable, assumptions are allowed only when explicitly registered.

Required assumption register:

| Assumption | Evidence level | Phases/blocks influenced | P12 probe | Status |
|---|---|---|---|---|
| ... | operator/domain/direct-user | ... | ... | OPEN / VALIDATED / REJECTED |

A material assumption may influence a candidate, but it must be explicitly probed during P12 or remain an open finding. Unresolved material assumptions block P14 closure.

## 5.3 Reference evidence

External products/design systems inform alternatives but never become product authority.

## 5.4 Operator adjudication

The designated human operator/product decision owner chooses among evidence-backed candidates and alone may set `LOCKED`.

---

# 6. Decision vocabulary

```text
LOCKED
  operator-approved current planning baseline; operator-only status

CANDIDATE
  plausible leading design; not approved

FINDING
  material unresolved question/contradiction

REJECTED
  alternative considered and deliberately not selected

DEFERRED
  known future seam with no current consumer; not implemented now

NOT-HUMAN-FACING
  backend/system behavior explicitly adjudicated as having no direct human UI consumer
```

Never write a `CANDIDATE` as though it were accepted authority.

---

# 7. Proportionality / tailoring

The method scales with risk and ambiguity; its core protections do not disappear on small work.

## Always required for any material frontend change

```text
method laws in §4
bounded authority recovery
explicit user need/outcome
operator adjudication before LOCK
visual structural artifact for the changed screen/block
Screen Contract + vertical backend trace
no screen-shaped API / no parallel frontend authority
```

## Activate only on named triggers

| Phase/instrument | Trigger |
|---|---|
| P4 IA work | new/changed navigation context, grouping, findability model or major collection |
| P6 reference study | unfamiliar/high-impact UX problem or meaningful pattern uncertainty |
| P7 competing hypotheses | genuine structural ambiguity with more than one credible solution |
| card sorting/tree testing | high-impact IA uncertainty that cannot be resolved confidently from existing evidence |
| multi-block ledger | more than one material block or dependencies between screens |
| full assembled P11 prototype | cross-screen flow/interaction needs realistic integrated testing |

## Small single-block delta

A small change to a converged product may collapse to:

```text
P0 bounded authority
→ P1 user-need delta
→ P8 rendered structural delta + operator LOCK
→ P9 Screen Contract/trace delta
→ P11 interaction delta when material
→ affected tests/review
```

Do not force product-wide matrices or reference research when no trigger exists.

---

# 8. Phase P0 — Recover accepted authority

Recover only the smallest authority pack needed for the current frontend problem.

Collect where applicable:

```text
product scope
actors/human capabilities
semantic/business owners
state/lifecycle
permissions/disclosure
backend/module boundaries
API/read models
concurrency/idempotency
external dependency semantics
accepted route/lens constraints
proof obligations
```

Exit: every known frontend requirement traces to current authority or is explicitly classified as unknown/assumption.

---

# 9. Phase P1 — Actors, jobs and user needs

For each human actor capture:

```text
actor / role context
trigger / situation
need / job-to-be-done
desired outcome
frequency / urgency
information needed to decide
common friction
handoffs
```

Recommended need format:

```text
When <context>,
I need to <goal>,
so that <outcome>.
```

Do not confuse access roles with personas unless evidence supports it.

Exit: human goals are independent from proposed pages/components, and material assumptions are in the register.

---

# 10. Phase P2 — End-to-end user-flow inventory

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

Capture alternate/failure branches when they change safe behavior.

Do not split implementation planning at a point that leaves a real human goal unfinished.

Exit: each accepted human goal has at least one complete end-to-end flow.

---

# 11. Phase P3 — Frontend Coverage Matrix

Minimum matrix:

| Capability / user need | Owner | User flow | Candidate context | Reads | Writes | Access/security | UX obligations | Status |
|---|---|---|---|---|---|---|---|---|

Cross-cutting system invariants become UX obligations, e.g.:

```text
unknown != known-empty
projection != mutation authority
hidden control != authorization
ambiguous outcome != known failure
stale write != silent overwrite
external mechanism success != product success
```

## Orphan backend human operation disposition

A backend/application operation with a human trust class but no discovered user need must resolve by exactly one of:

```text
A. real evidenced user need + frontend home
B. operator-adjudicated NOT-HUMAN-FACING or DEFERRED disposition in the ledger
C. UPSTREAM finding that accepted capability is excess/misaligned
```

Never invent a screen or user need merely to make an orphan count reach zero.

Exit:

```text
accepted human capabilities mapped
human-consumer backend operations mapped or explicitly dispositioned
no invented capability
material findings named
```

---

# 12. Phase P4 — Candidate Information Architecture

Design how users find, understand and move among product information/tasks **before final screen composition**.

## 12.1 Durable terminology glossary

Create and maintain a user-language glossary for material objects, actions and states.

| Concept/action | User-facing term | Avoid/synonyms | Notes |
|---|---|---|---|
| ... | ... | ... | ... |

This prevents implementation-time terminology drift. Final marketing/content polish is not required here.

## 12.2 Object/task inventory

For each concept:

```text
name in user language
why users care
primary actions
relationships
frequency
browse/search/filter role
whether independent navigation is justified
```

Do not mirror domain aggregates, DB entities or API families automatically.

## 12.3 Navigation model

Evaluate global navigation, local/context navigation, home/default landing, primary work queues, browse hierarchy, search entry, cross-context links and breadcrumbs only where hierarchy genuinely exists.

## 12.4 Findability

For important collections decide intentionally among:

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

## 12.5 Mental-model validation

When IA is uncertain/high-impact, use the smallest useful technique: category naming review, card sorting, tree testing, first-click testing or operator/domain walkthrough.

## 12.6 Future seams

Future concepts may shape extensibility but must not become live screens/routes without current authority.

## P4 exit law

P4 exits with **CANDIDATE IA**, never `LOCKED` IA.

Global IA becomes `LOCKED` only through the first global-frame block cycle (for example an App Shell + global IA block) after reference study when triggered, structural hypotheses when ambiguity is real, a rendered visual artifact and explicit operator adjudication.

This first global-frame lock is the first mandatory whole-product coherence checkpoint because later blocks inherit it.

---

# 13. Phase P5 — Screen and material-surface inventory

Derive screens from user flows + candidate IA, not endpoints.

Distinguish:

```text
route/page
material surface inside a route
drawer/modal
inline composition region
alternate view of one collection
material state variant
```

Create a separate material surface when one changes:

```text
primary semantic truth
safe user action
write owner
identity needed
concurrency/idempotency behavior
content-integrity/exactness guarantees
security/disclosure context
recovery path
editor/viewer mode
```

Do not split screens merely because loading style, spacing, component file or responsive arrangement differs.

Exit: every human-flow step/material decision has a candidate screen/surface home.

---

# 14. Phase P6 — Reference Study (per block, conditional)

Study how mature products solve the **same user task** before committing to a layout.

Select approximately 3–6 useful references when available; fewer is fine when the problem is conventional or evidence is limited. Stop when additional references are repeating the same useful patterns.

Potential sources:

```text
direct competitors
adjacent enterprise/SaaS products
high-quality products with analogous jobs
design systems with relevant task patterns
platform conventions familiar to target users
```

Analyze:

```text
user problem solved
navigation model
information hierarchy
primary/secondary actions
collection representation
search/filter strategy
list/grid/master-detail usage
progressive disclosure
selection/action behavior
empty/loading/error patterns
responsive behavior
density
strong pattern
mismatch/risk for our product
```

Evidence matrix:

| Reference | Relevant job | Pattern | Benefit | Risk/mismatch | Candidate lesson |
|---|---|---|---|---|---|

Where useful, include at least one disconfirming observation: where the leading pattern performs poorly or does not fit. Research should challenge, not merely confirm.

Always separate:

```text
SOURCE OBSERVATION
INFERENCE
PRODUCT DECISION
```

Exit: enough evidence-informed structural possibilities exist for the current block; no unbounded browsing requirement remains.

---

# 15. Phase P7 — Competing Layout Hypotheses + lightweight data feasibility

For consequential blocks with genuine ambiguity, compare 2–3 plausible structures.

Examples:

```text
dense table
cards/grid
structured list
master-detail
category-first browse
search-first
mixed browse + list
```

Evaluation variables should be chosen for the block and may include:

```text
task completion speed
recognition
comparison
scanability
information density
cognitive load
frequency
scale / number of records
preview needs
bulk actions
context preservation
error recovery
accessibility
responsive viability
accepted future extensibility
implementation complexity
data volatility / refresh behavior
realistic cold-start / small-collection behavior
export/print/reporting needs when current product need exists
backend truth fit
```

## 15.1 Data-feasibility line — required before P8 LOCK

Each **leading** hypothesis must explicitly state the material data/action assumptions it needs:

```text
fields / summaries
identity sources
scale/pagination expectations
sort/filter needs
preview/thumbnail/content truth
material writes/actions
```

For each assumption mark:

```text
PRESENT-IN-AUTHORITY
or
FINDING
```

This is a lightweight feasibility gate, not a full Screen Contract. The full exact contract remains P9 because its detail depends on the structure selected.

## 15.2 Example: table vs cards

Table leads when users primarily compare structured attributes, scan many rows, sort/filter repeatedly or perform row/batch actions.

Cards/grid lead when recognition/preview dominates cross-row comparison and items need more heterogeneous summaries.

Alternate list/grid views exist only when both tasks are genuinely important.

Exit: one leading candidate is ready for structural wireframing, or a named finding remains. If no real ambiguity existed, the ledger records the one-line reason for using a single hypothesis.

---

# 16. Phase P8 — Block-by-block rendered Structural Wireframing

## Goal

Turn the selected hypothesis into a low-fidelity **viewable structural design** that the operator can actually inspect.

P8 freezes structure, not brand styling.

## 16.1 Wireframe medium — mandatory visual artifact

A P8 candidate MUST be rendered/viewable by the operator. Prose, markdown tables or textual boxes alone do not satisfy P8.

Acceptable media include:

```text
schematic image
SVG / box-layout diagram
wireframing tool capture
unstyled grayscale HTML/CSS structural skeleton
other rendered visual equivalent
```

A deliberately unstyled per-screen HTML skeleton is permitted at P8 when it is the easiest shared medium.

### P8 vs P11 boundary

The boundary is **scope/fidelity**, not technology.

```text
P8
  one bounded screen/block
  structural hierarchy and proportions
  placeholder/local fixture content
  no brand styling
  no obligation for complete cross-flow navigation
  used to explore and LOCK structure

P11
  realizes already-locked screens together
  material interaction/navigation works
  negative/recovery states are inspectable
  supports realistic end-to-end flow testing
```

HTML may therefore appear in P8 without turning P8 into the final functional prototype phase.

## 16.2 Structural wireframe decisions

```text
screen hierarchy
major regions
relative width/height
primary reading order
navigation placement
information grouping
card/table/list/master-detail choice
density class
primary/secondary action placement
progressive disclosure
dialog/drawer vs inline editing
empty/error/conflict region placement
responsive transformation rules
keyboard/focus-order plausibility
heading/label structure
disclosure reachability
non-drag alternative for essential drag interactions
```

## 16.3 Block review cycle

```text
BLOCK INPUT
  user goals + candidate IA + coverage + local authority

REFERENCE STUDY when triggered

LAYOUT HYPOTHESES when ambiguity is real

LEADING HYPOTHESIS DATA-FEASIBILITY CHECK

RENDERED CANDIDATE STRUCTURAL WIREFRAME

OPERATOR VISUAL WALKTHROUGH
  layout, hierarchy, position, size, density, discoverability, task flow

FINDINGS / REVISION

OPERATOR LOCK
or continue iterating
```

The next material block does not become baseline until this block is `LOCKED`, unless the operator explicitly authorizes parallel candidate work under §4.6.

## 16.4 Operator walkthrough questions

```text
Can I find what I came here to do?
Is the most important information visible first?
Does the screen expose too much at once?
Do I understand where I am and where I can go next?
Are primary actions obvious without being dangerous?
Would I rather compare, recognize or preview these items?
Does density match the task?
Should this be a page, drawer or inline region?
Do I need prior-screen context while acting?
Would users repeatedly open records merely to find the right one?
Does backend structure leak into the UI?
Can the primary workflow be completed by keyboard without relying on hover-only disclosure?
Is focus/read order plausible and understandable?
If an interaction uses drag, is an equivalent non-drag path clear?
What happens structurally at narrow widths without hiding the only material action?
```

Exit: operator explicitly sets `LOCKED`, or findings remain open.

---

# 17. Phase P9 — Screen Contract and vertical backend trace

After structural LOCK, prove every material region/control is realizable by accepted authority.

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
FAILURE MESSAGE INTENT
SUCCESS CONSEQUENCE
AUTHZ / DISCLOSURE
PROOF
FORBIDDEN FRONTEND AUTHORITY
BACKEND SUFFICIENCY
```

### Failure message intent

For each material failure define what the user must understand and what safe next action must be possible. This is **not** final copywriting; it prevents the implementer from inventing contradictory error semantics.

Example abstraction:

```text
state: stale write
intent: user must understand that server truth changed, local input is preserved, and review is required before resubmission
```

## Bidirectional law

```text
Product/backend → frontend
capability → owner → operation/read model → screen/region/control

Frontend → Product/backend
screen/control → operation/read truth → owner → accepted capability
```

A one-way trace can hide orphan operations or invented UI behavior.

Exit: each material screen/control is `READY` or blocked by explicit finding.

---

# 18. Phase P10 — Incremental + terminal Component/Interaction Pattern Vocabulary

Pattern vocabulary evolves from **locked evidence**, not upfront design-system speculation.

## 18.1 After each block LOCK

Run a bounded consolidation pass:

```text
local patterns observed
→ compare against already locked screens
→ if ≥2 locked screens share the same purpose/state ownership/protected semantics/accessibility/failure class
→ candidate shared pattern may graduate
```

Do not graduate a shared pattern mid-block.

## 18.2 Pattern entry

```text
purpose
locked screens using it
required inputs
states/variants
material interactions
validation timing semantics when form-like
accessibility expectation
responsive behavior
what it does NOT own
```

## 18.3 Terminal P10 reconciliation

Before whole-product prototype closure, reconcile the vocabulary globally:

```text
unexplained semantic duplicates = 0
false abstractions = 0
local-only patterns remain local when sharing would be wrong
```

When later evidence falsifies an earlier abstraction, refine it deliberately rather than preserving a bad component boundary for consistency.

---

# 19. Phase P11 — Interactive Low-Fidelity Prototype

Create a navigable representation of **already-locked** structures.

Default technical baseline when code is useful:

```text
HTML
CSS
vanilla JavaScript
local deterministic fixtures/state simulation
```

Avoid the production framework unless plain browser technology is genuinely insufficient. Prototype code is evidence, not production code.

P11 may be realized per locked block and then assembled.

It must preserve:

```text
accepted navigation
locked hierarchy/proportions
material fields/regions
action placement
pattern identity
material dialogs/drawers
negative/recovery states
read-only vs editable regions
responsive structural rules
```

Material buttons/links/forms/dialogs must work when they change state/navigation; no decorative dead controls.

Use deterministic scenarios for success, empty, denied, not-found/non-disclosable, stale concurrency, ambiguous outcome, dependency failure, admission/integrity failure and lifecycle transitions as relevant.

Where useful, material controls carry machine-readable trace metadata such as:

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

P11 does not reopen visual layout exploration by default. If realistic interaction falsifies a locked structure, record a finding and return to the smallest affected phase.

Exit: reviewed flows are realistically navigable without the prototype inventing product authority.

---

# 20. Phase P12 — Adversarial UX + Architecture Walkthrough

Attack the assembled experience from target-user, product-owner, principal-designer, IA, frontend, backend/domain, accessibility and adversarial-reviewer perspectives.

UX attack examples:

```text
Can users find the right place without backend terminology?
Are frequent tasks unnecessarily deep?
Are decision-critical facts hidden by progressive disclosure?
Does table/card/master-detail choice match the task?
Does the interface force recall where recognition is possible?
Are similar interactions inconsistent without reason?
Are screens optimized for a demo rather than repeated use?
```

Architecture attack examples:

```text
Does every material field have a source?
Does every write have accepted owner/operation?
Does fixture state accidentally become product truth?
Did convenience API sneak in?
Does UI visibility pretend to authorize?
Are concurrency/idempotency/recovery semantics preserved?
Does navigation depend on unavailable IDs?
```

## Mandatory assumption probe

Every material assumption in the register must be explicitly:

```text
VALIDATED
REJECTED
or carried as FINDING
```

Any material assumption left OPEN blocks P14.

## Finding classes

```text
UX-LAYOUT
IA
PATTERN
SCREEN-CONTRACT
UPSTREAM
VISUAL-DESIGN
```

Upstream reopen remains the last resort after layout, IA, pattern, contract and allowed read-composition options are exhausted.

Exit: no unresolved material UX/IA/contract contradiction remains in reviewed scope.

---

# 21. Phase P13 — Visual Design Handoff + structural-conformance review

Handoff includes:

```text
locked IA
locked flows
locked structural wireframes
terminology glossary
pattern vocabulary
functional prototype
Screen Contracts
material state/message-intent inventory
responsive structure
```

Visual design may change aesthetic treatment but may not silently change navigation, IA, material fields, action semantics, reading order, region priority or density class.

## Required structural-conformance check

After visual design, compare designed screens against locked structural wireframes for at least:

```text
reading order
region priority
primary/secondary action placement
density class
navigation meaning
material information visibility
responsive structural behavior
```

Mismatch is classified using the normal finding law and returns to the smallest affected phase.

If visual design occurs during production implementation (for example an LLM-built product), this structural-conformance check transfers into implementation review as an explicit obligation; it does not disappear.

---

# 22. Phase P14 — Frontend Implementation-Readiness Closure

Close only when project-specific counts reconcile exactly.

Generic closure requirements:

```text
accepted human goals                         complete
end-to-end flows                             complete
information architecture                    operator-adjudicated / locked where applicable
terminology glossary                        coherent for material concepts/actions
material screens/surfaces                   complete
reference study                             complete where triggered
layout hypotheses                           adjudicated where triggered
structural visual wireframes                operator-locked
Screen Contracts                            complete
material controls                           100% bound
navigation identities                        100% sourced
component/interaction patterns              reconciled
interactive prototype                       complete for accepted scope
negative/material states                    represented
material failure message intent             defined
frontend ↔ backend trace                     complete
backend human operations without disposition 0
invented frontend Product operations        0
screen-shaped APIs                          0
material assumptions left OPEN              0
post-design structural-conformance defects  0 unresolved
unresolved material UX findings             0
unresolved architecture findings            0
```

Product-specific invariants are added by the consuming repository.

---

# 23. Block operating protocol

For non-trivial products maintain a block ledger.

Example names only:

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

Actual blocks are product-derived.

Required block record:

| Field | Required |
|---|---|
| Block | stable planning ID |
| User goals | explicit |
| Authority pack | bounded |
| Assumptions | linked |
| Reference study | complete when triggered |
| Hypotheses | 1–3 depending on ambiguity |
| Single-hypothesis justification | required when alternatives are skipped |
| Leading-candidate data feasibility | present/finding |
| Rendered visual wireframe | required |
| Operator walkthrough | explicit |
| Findings | explicit |
| Decision | LOCKED / CANDIDATE / FINDING |
| Screen Contracts | after structure |
| Pattern consolidation | after LOCK |
| Interactive realization | after LOCK when material |

Progression follows §4.6 exactly: next material baseline block requires current `LOCKED`, unless explicit operator authorization permits parallel candidate work.

---

# 24. Phase scope — global vs per-block

| Phase | Scope |
|---|---|
| P0–P3 | global/bounded product foundation |
| P4 | global candidate IA; exits CANDIDATE |
| P5 | global candidate screen/surface inventory |
| P6–P9 | **per block** inside the operator review loop |
| P10 | incremental consolidation after each block LOCK + terminal global reconciliation |
| P11 | per locked block, then assembled when integrated flow testing is needed |
| P12–P14 | global/assembled closure |

This table governs over any linear shorthand diagram.

First whole-product coherence checkpoint: global-frame/B01 lock.  
Terminal whole-product coherence checkpoint: P12 assembled walkthrough.

---

# 25. Minimum artifact inventory

Repositories may consolidate files, but the following logical records must exist where triggered:

| Artifact | Minimum content |
|---|---|
| Authority map | bounded current sources |
| Need/assumption register | actors, goals, assumptions, evidence |
| User-flow inventory | complete outcomes + material branches |
| Coverage matrix | capability ↔ owner ↔ frontend home/disposition |
| Candidate IA + glossary | navigation/findability + user terms |
| Block ledger | status, evidence, decisions |
| Reference notes | only when P6 triggered |
| Hypothesis comparison | only when P7 triggered |
| Rendered structural wireframe | operator-viewable candidate/locked artifact |
| Screen Contract/trace | material regions/controls ↔ authority |
| Pattern vocabulary | derived shared/local patterns |
| Interactive prototype | when P11 interaction testing is material |
| Adversarial findings | P12 outcome |
| Structural-conformance result | post-visual-design check |

The method does not require one file per row.

---

# 26. Accessibility is structural, not polish

Wireframes/prototypes must consider keyboard navigation, semantic control choice, focus order, heading hierarchy, labels/instructions, error association, non-color-only meaning, target viability, responsive reflow, screen-reader comprehensibility and non-drag alternatives for essential interactions.

A layout that cannot reasonably become accessible is not a valid locked candidate merely because it looks efficient.

---

# 27. Responsive planning

Do not defer structural responsive behavior to production CSS.

Define:

```text
what remains primary
what stacks
what collapses
what becomes a drawer/menu
what becomes locally scrollable
what must stay visible
how tables/lists/cards transform
```

Responsive transformation must not change product semantics or hide the only material action.

---

# 28. Density and progressive disclosure

Density is a user-task decision.

Use higher density for rapid scanning/comparison across many records; more whitespace/preview when recognition/comprehension dominate.

Progressive disclosure may hide secondary information but must not hide facts required for the current decision merely to make a screen look clean.

---

# 29. Search, browse and filters

Do not assume search replaces IA or vice versa.

```text
known-item lookup → search may dominate
exploratory discovery → browse/grouping may dominate
large structured collection → filter/sort may dominate
mixed use → combine carefully
```

Filter controls require human-understandable value sources. An API parameter accepting an opaque ID does not prove a coherent selector exists.

---

# 30. Tables, cards, lists and master-detail

These are task patterns, not aesthetics.

### Data table
Strong for comparable attributes, many rows, repeated sort/filter, precise comparison and row/batch actions.

### Card/grid
Strong for recognition, preview/thumbnail, heterogeneous summaries and lower cross-row comparison needs.

### Structured list
Strong for compact simple records with one main label plus small metadata/action.

### Master-detail
Strong when users inspect many records sequentially and should preserve list/context while doing so.

### Multiple views
Offer alternate views only when distinct important user tasks justify them.

---

# 31. Findings and smallest-reopen law

When a screen fails, classify in order:

```text
1. UX/layout issue?
2. IA issue?
3. wrong candidate pattern?
4. incomplete Screen Contract?
5. missing read composition already allowed by authority?
6. excess/misaligned accepted backend capability?
7. genuinely missing Product/API capability?
```

Only the smallest evidence-backed owner is reopened.

For excess capability, use the orphan disposition law rather than manufacturing UI.

After correction, redraw/re-review only affected blocks unless the finding invalidates global assumptions.

---

# 32. Adversarial independent review protocol

When governance supports it, review the methodology/major frontend plan independently from the stance of Principal Product Designer + Information Architect + Senior UX/Service Design + Senior Frontend Architect + accessibility-aware adversarial reviewer.

Attack:

```text
missing user-centered discovery
weak IA
premature screen inventory
premature pattern vocabulary
single-solution wireframing
lack of useful reference study
poor operator feedback loop
all-at-once generation
screen-shaped API risk
frontend authority duplication
component duplication
YAGNI / overengineering
unprovable backend-screen mappings
missing failure/recovery UX
accessibility/responsive deferral
untracked assumptions
visual design structural drift
```

Severity:

```text
MATERIAL
IMPORTANT
OPTIONAL
UNSUPPORTED PREFERENCE
```

Correct only evidence-backed findings.

---

# 33. Reusable closure checklist

## User/product
- [ ] Human actors/needs are explicit before screens.
- [ ] User needs describe outcomes, not predetermined UI solutions.
- [ ] End-to-end flows cover complete human goals.
- [ ] Material assumptions are registered.

## IA
- [ ] Navigation reflects tasks/mental models, not backend topology.
- [ ] P4 IA remained CANDIDATE until operator visual lock in the global-frame block.
- [ ] Browse/search/filter strategy is deliberate.
- [ ] Terminology glossary is coherent.

## Reference/layout
- [ ] P6/P7 ran only where triggered.
- [ ] References are analyzed by task/pattern, not copied visually.
- [ ] A disconfirming observation was considered where useful.
- [ ] Consequential alternatives were compared when real ambiguity existed.
- [ ] Single-hypothesis decisions have a short justification.
- [ ] Leading hypothesis passed the lightweight data-feasibility line.
- [ ] Table/card/list/master-detail choice is task-justified.
- [ ] Density/progressive disclosure are deliberate.

## Operator visual loop
- [ ] Every material block has a rendered visual artifact.
- [ ] Major blocks were reviewed individually with the operator.
- [ ] Only the operator set LOCKED.
- [ ] No next material block became baseline before LOCK or explicit parallel authorization.

## Architecture
- [ ] Every material read/write/identity has accepted source/owner.
- [ ] No client state became Product authority.
- [ ] No screen-shaped API was added for convenience.
- [ ] Orphan backend human operations are mapped or explicitly dispositioned.
- [ ] Concurrency/idempotency/recovery semantics are represented.
- [ ] Material failure message intent is defined.

## Accessibility/responsive
- [ ] Accessibility was checked during P8, not only after prototype.
- [ ] Keyboard/focus/non-drag paths are structurally plausible.
- [ ] Responsive transformation is defined.

## Patterns
- [ ] Pattern consolidation happened after block locks, not mid-block.
- [ ] Shared patterns have ≥2 locked semantic consumers or equivalent evidence.
- [ ] Duplicate semantic patterns are reconciled.
- [ ] Cosmetic similarity did not force abstraction.

## Prototype/design
- [ ] P11 realizes locked structure rather than inventing it.
- [ ] Material interactions/negative states are inspectable.
- [ ] Material assumptions were probed in P12.
- [ ] Visual design passed structural-conformance review.

## Closure
- [ ] Frontend ↔ backend trace is complete.
- [ ] Backend human operations without disposition = 0.
- [ ] Invented operations = 0.
- [ ] Material assumptions left OPEN = 0.
- [ ] Unresolved material UX/IA/architecture findings = 0.
- [ ] Implementation can proceed without material UX invention in code.

---

# 34. Compact process

```text
GLOBAL FOUNDATION
P0  Recover accepted authority
 ↓
P1  Actors / jobs / user needs + assumptions
 ↓
P2  End-to-end user flows
 ↓
P3  Frontend Coverage Matrix
 ↓
P4  Candidate Information Architecture + glossary
 ↓
P5  Candidate screen/material-surface inventory

PER-BLOCK LOOP
P6  Reference Study when triggered
 ↓
P7  Competing Layout Hypotheses when triggered
    + lightweight data-feasibility line
 ↓
P8  Rendered Structural Wireframe
    + operator visual adjudication / LOCK
 ↓
P9  Screen Contract + bidirectional backend trace
 ↓
P10 bounded pattern consolidation
 ↓
P11 interactive realization when material
 ↓
NEXT BLOCK only after operator LOCK (or explicit parallel-candidate authorization)

ASSEMBLED CLOSURE
P10 terminal pattern reconciliation
 ↓
P11 assembled prototype when needed
 ↓
P12 Adversarial UX + Architecture Walkthrough
 ↓
P13 Visual Design Handoff + structural-conformance review
 ↓
P14 Frontend Implementation-Readiness Closure
```

A finding returns only to the smallest affected phase. Do not restart the entire method unless the finding invalidates global assumptions.

---

# 35. Research basis

This methodology synthesizes established principles without making any external product authority:

```text
GOV.UK Service Manual
  understand users and their whole problem
  prototype before committing to build
  test multiple approaches when uncertainty is real

Nielsen Norman Group / established IA practice
  user mental models
  recognition over recall
  findability / scanability
  card sorting / tree testing when useful

Enterprise design systems such as Carbon
  data tables for dense structured comparison
  alternate views for distinct user tasks
  progressive disclosure for secondary information
```

External evidence informs hypotheses. Accepted product authority + user evidence + operator adjudication determine the actual frontend.

---

# 36. Final principle

A strong frontend plan should make production coding feel almost boring.

Implementation should mostly be:

```text
realize operator-locked structure
→ bind accepted contracts
→ implement reviewed patterns
→ preserve responsive/accessibility behavior
→ prove states and failures
```

not:

```text
invent navigation
→ invent screens
→ invent components
→ discover missing APIs
→ invent user-facing failure semantics
→ redesign workflows
→ reconcile backend/frontend after the fact
```

> **Frontend implementation readiness means the important product, IA, layout, interaction, language-intent and system-contract decisions have already been made visibly, reviewed deliberately, operator-locked where required and traced to evidence.**
