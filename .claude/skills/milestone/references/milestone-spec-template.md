# Milestone Spec Authoring — Distilled Framework

> Companion to `templates/milestone.md`. Template = shape. This file = quality bar.
> Read before authoring any non-trivial `milestone.md`. Distills three industry frames
> into the checks a milestone spec must survive **before** features start.

## Why this exists

`milestone.md` is a **contract you validate against** at close. Weak spec → validator
has nothing to judge → "done" decays into "whatever we built". Three frames cover the
failure modes:

- **Shape Up (Basecamp)** — kills scope drift via fixed appetite + named rabbit holes.
- **Amazon PR/FAQ (Working Backwards)** — kills outcome ambiguity via consumer-first
  narrative + FAQ of hard questions.
- **arc42** — kills missing architectural context via quality goals + constraints +
  risks.

Each frame answers one question the validator will ask. Skip a frame → expect that
question to surface as a C-check failure later.

## The five sections every milestone spec needs (mapped to the template)

### 1. Objective — outcome, not output

Frame: **PR/FAQ Working Backwards**. Write the milestone as the change a downstream
*consumer* (operator, end user, next milestone) experiences. Not "we refactor X" —
"after this milestone, X behaves as Y, observable via Z".

Checks:
- Could a stranger read it and name the visible change? No → rewrite.
- Does it state the **quality bar moved** (grade, closed defect class, contract cleaned)?
  If yes, name the exact metric and how it's re-measured (C5 binds on this).
- Is the audience the *consumer* of the change, not the implementer?

Anti-patterns: "improve auth", "tidy the database", "modernize the screen".
Replacements: "auth refuses expired sessions on every protected route", "DB has one
source of truth for tenant tier", "screen renders without a flash of mock data".

### 2. Appetite + rabbit holes — what you will NOT chase

Frame: **Shape Up**. A milestone has a **fixed appetite** (rough time/effort the
operator authorizes) and a **named rabbit hole list** (tempting expansions you refuse
in advance). Without these, every milestone grows until something else stops it.

Author into `milestone.md` under **Dependencies & constraints** (constraints) or a
short "Appetite" line above Features:

- **Appetite:** <"1 week of focused work" / "5 features cap" / "no schema changes">.
- **Rabbit holes (do not chase):** named temptations + the reason each is out of scope
  for *this* milestone. These become the validator's scope-drift baseline (C6).

Checks:
- Is there at least one rabbit hole? Zero → you haven't thought about scope yet.
- Each rabbit hole has a reason (defer / wrong milestone / no consumer yet)?

### 3. Features — atomic, ordered, each with a consumer

Frame: **PR/FAQ** (every feature has a named consumer) + **arc42** (each is a coherent
building block).

The Features table is the milestone's payload. Quality bars:

- **One outcome per row.** A row that says "X and Y and Z" hides three features.
- **Named consumer.** "What to implement" is meaningless without "who calls this".
  Encode it: `<route|hook|module|screen|job>` consumes `<shape>`.
- **Objectively checkable acceptance.** "What to validate" cell holds a test name, a
  command, a response shape — never "works".
- **Order is intentional.** First feature unblocks the rest. Last feature is the one
  whose evidence closes the milestone's quality-bar claim.

If a row's acceptance is fuzzy, the feature isn't ready — break it down or interview
the operator before writing the row.

### 4. Quality goals + constraints — what the build must respect

Frame: **arc42 §1.2 Quality Goals + §2 Constraints**. Architectural context that
governs every feature, recorded once at milestone scope so per-feature plans don't
have to re-derive it.

Author into **Dependencies & constraints**:

- **Quality goals** (top 3, ranked): e.g. *correctness > simplicity > performance*,
  or *contract stability > velocity*. The validator uses this rank when judging trade-offs.
- **Architectural constraints** (hard rules): no migrations, reads stay live,
  advisory-lock hazard rules, contract-first regen order, separation-of-powers
  boundaries, root-cause-over-symptom-patch (binds C5/C6).
- **Risks** (named, not "various"): each risk + its mitigation or accepted-defer.

Checks:
- Top-3 quality goals ranked? Equal-priority list = no priority.
- Each constraint phrased as a rule the validator can fail on?
- Each risk has an owner (mitigate now / defer with trigger / accept)?

### 5. Validation definition — what the close gate enforces

Frame: **arc42 §11 Risks + Technical Debt** + Shape Up "shaped work has a definition
of done baked in".

The template already lists 1–5; the quality bar is that each line is **concrete enough
for a fresh subagent to execute without asking you anything**:

- Per-feature acceptance — points at the spec.md Validation Gate of each feature
  (transitive — milestone gate verifies the feature gates, doesn't re-invent them).
- Workflow-class QA checklist — name the file path, not a category.
- Regression — name the prior milestones whose gates must still pass.
- Quality-bar re-measure — name the metric and the command/command-set.
- No unplanned scope — anchored to the Features table + rabbit-hole list above.

## Authoring loop (one pass, then revise)

1. Draft Objective in PR/FAQ voice (consumer narrative + FAQ of 3 hardest objections).
2. Set Appetite + Rabbit holes — write the refusals first; they constrain everything below.
3. List Features — one outcome per row; if a row is fuzzy, interview the operator
   (`Skill(superpowers:brainstorming)`) before committing it.
4. Fill Quality goals / Constraints / Risks (arc42 lens).
5. Tighten Validation definition — every line executable by a fresh subagent.
6. Re-read top-to-bottom: does the validator have enough to judge PASS/FAIL on each
   line without phoning you? No → tighten. Yes → ship to operator for approval.

## Anti-pattern catalog (the spec is too weak if…)

- Features table has rows whose acceptance is an adjective ("clean", "stable", "good").
- No rabbit holes — scope is implicitly infinite.
- No named consumer per feature — producer-first thinking has leaked in.
- Quality goals listed unranked — first trade-off will reopen the spec.
- Validation lines reference "all tests" / "no regressions" without naming which.
- Risks section says "none" — either dishonest or unexamined.
- "Improve" / "modernize" / "refactor" in the Objective — output, not outcome.

## Source frames (one-line each)

- Shape Up — Singer 2019, Basecamp. Fixed appetite + rabbit holes + circuit breaker.
- Working Backwards / PR/FAQ — Bryar & Carr 2021, Amazon. Consumer narrative + FAQ.
- arc42 — Starke. Architectural communication template; §1.2 quality goals, §2
  constraints, §11 risks bind here.

These are framings, not religion — adapt to the milestone. A schema-migration milestone
leans hard on arc42 constraints + risks. A user-visible-feature milestone leans on
PR/FAQ. A scope-fragile cleanup milestone leans on Shape Up appetite. Pick the lens
each section needs and move on.
