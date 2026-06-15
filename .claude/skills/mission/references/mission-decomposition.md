# Mission Decomposition — intent → milestones → features

Read this in Phase 3, when turning the discovery brief + locked decisions into the §7 milestone table.
The goal is a decomposition that is **stable** (you can validate against it later without it having drifted)
and **traceable** (every milestone earns its place from a finding; every finding lands somewhere).

## What a milestone is

A **bounded slice validatable in one pass**. Big enough to be a coherent deliverable an independent judge
can rule on; small enough that one `milestone-validator` run can actually re-run its gates from clean state.
If you can't imagine the validator re-running a milestone's acceptance in a single sitting, it's two
milestones.

A milestone moves something observable — a defect class to zero, a capability to live, a dimension up a
grade. Name that bar in the milestone's objective and make the close-gate prove it.

## What a feature is

The **atomic unit of work** inside a milestone. One feature → one home (`f<n>.x-<slug>/`) downstream, with
its own consumer-contract `spec.md`, `plan.md`, and `evidence.md` (the `milestone` skill owns that
lifecycle). In `mission.md` a feature is just two cells: *what to implement* (by outcome) and *what to
validate* (objectively checkable). No steps.

## Ordering — the part that's easy to get wrong

1. **Dependencies first.** If milestone B consumes something milestone A produces, A comes first. For
   greenfield, the consumer-contract-first rule still applies at the program scale: sequence so a consumer's
   required shape is known before its producer is built.
2. **De-stale / clear-the-ground early.** If stale docs, a broken baseline, or a missing prerequisite would
   pollute later work (grade-a's M0 docs-destaling), do it first.
3. **Risk-isolating work last.** Put the change most likely to regress the quality bar at the end, so it
   can't silently undo earlier milestones (grade-a put the systemic ports last "so it cannot regress the
   grade"). The terminal gate then judges the *whole* thing in its final state.

## Fixing classes, not instances

When discovery found a *class* of defect (the same shape at many sites), the milestone that closes it must
close the **class** — and its gate must prove there are zero remaining instances anywhere in scope, not just
at the sites discovery happened to list. This is the difference between "patched the 3 we saw" and "the
class is gone". Say so in the milestone's "what to validate".

## Traceability check (do this before Phase 4)

- Every milestone/feature ↔ at least one discovery finding (why does this exist?).
- Every discovery finding ↔ a milestone **or** an explicit out-of-scope row in §5 (where did this go?).
- Orphans on either side are a defect: an unmotivated milestone (scope creep) or a dropped finding (silent
  cap). Fix before you present.

## Anti-patterns

- **Milestones as a to-do list.** A milestone is a validatable slice with a bar, not "the next batch of
  tasks". If a milestone has no bar to re-measure, it's probably mis-cut.
- **Execution detail leaking up.** "First we'll edit X, then run Y" belongs in a feature `plan.md`, never in
  `mission.md`. Catch yourself writing *how* and move it down.
- **One giant milestone.** If M0 is the whole mission, you've lost the per-milestone gate — the thing that
  makes the structure worth more than a single plan. Split until each slice is independently judgeable.
- **Forcing the structure on small work.** If the whole thing is one validatable slice, a mission is
  overkill — recommend a single `writing-plans` plan instead and stop.
