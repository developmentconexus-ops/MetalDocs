# Root-Cause / Global-Maximum Engineering Method

> **Status:** Canonical engineering standard  
> **Owner:** `wiki/standards/`  
> **Last verified:** 2026-08-12

## Binding principle

> **Always simplify the code, never simplify correctness. Find the root cause, test whether the proposed solution is only a local maximum, and prefer the global structure that makes the defect class unrepresentable or mechanically impossible at the strongest reasonable boundary.**

This method governs non-trivial engineering work. It is not a mandate for maximum sophistication. It distinguishes **essential complexity** from **accidental complexity** and removes the latter without weakening invariants.

## Problems this method prevents

1. **Patch-on-patch:** fixing the visible symptom while preserving the structural fact that made it possible.
2. **Local-maximum optimization:** improving a workaround, legacy seam, or faulty foundation instead of replacing it.
3. **False simplification:** deleting enforcement or proof because it is large or inconvenient while the invalid state remains representable.
4. **Overengineering:** creating a new framework, parser, guard, registry, configuration layer, or second source of truth when a stronger existing boundary can enforce the same property.
5. **Infinite review:** continuing hypothetical hardening after the material property is structurally solved and proved.

## Definitions

### Symptom

The observable failure: wrong response, race, duplicated implementation, CI escape, incorrect write, broken import boundary, or similar evidence.

A symptom identifies where a defect appeared. It does not by itself identify where the defect should be fixed.

### Root cause

The structural fact that made the defect possible.

Examples:

- not “taxonomy ignored a tenant error,” but “multiple modules were allowed to define independent idempotency identity policies”;
- not “one worker forgot rollback,” but “application code can own transaction lifecycle in many places”;
- not “one request DTO drifted,” but “OpenAPI and hand-authored runtime shapes were both treated as authorities.”

### Target property / invariant

The statement that must remain true for every valid implementation, independent of the current code shape.

Examples:

- every idempotency replay is scoped by a valid tenant and actor;
- application transaction lifecycle has one owner;
- generated contract types are the runtime request authority;
- a foreign module cannot write another module's persistence directly.

### Local maximum

A solution that meaningfully improves the current implementation but preserves the structural limitation that produced the defect class.

A local maximum is legal only when explicitly transitional, with a named successor and deletion condition.

### Global maximum

The best sustainable structure for the system's actual constraints: it resolves the root cause, converges authorities, preserves required invariants, minimizes accidental complexity, and makes invalid states impossible or mechanically detectable at the strongest reasonable boundary.

Global maximum does **not** mean maximum abstraction, maximum code, copying big-tech infrastructure without need, perfect future-proofing, or indefinite redesign.

### Essential vs accidental complexity

**Essential complexity** comes from real domain or system constraints such as multi-tenancy, transactional consistency, authorization, asynchronous delivery, immutable evidence, and contract evolution.

**Accidental complexity** comes from the chosen implementation: duplicate policies, parallel abstractions, hand-synced registries, unnecessary compatibility layers, redundant enforcement, speculative configuration, and repeated lifecycle code.

The method removes accidental complexity. It does not weaken essential complexity.

### YAGNI

YAGNI removes speculative capability, not required correctness.

YAGNI may justify deleting unused extensibility, abstractions with no current second consumer, configuration nobody sets, unsupported compatibility paths, and duplicate enforcement that proves exactly the same property.

YAGNI does **not** justify deleting an invariant enforcement point, a fail-closed default, a test for a reachable state, or a boundary that prevents a known defect class.

## Enforcement hierarchy

When preserving a property, prefer the strongest reasonable mechanism:

1. **Structure / API makes the state unrepresentable**
2. **Type system makes the state invalid**
3. **Database/schema constraint makes the state invalid**
4. **Runtime boundary fails closed**
5. **Test proves reachable behavior**
6. **Lint / static guard detects the violation**
7. **Documentation / convention**

This is a preference order, not a rigid rule. A lower layer is valid when a stronger layer cannot represent the complete property without disproportionate cost or loss of legitimate flexibility.

When a lower layer is chosen, the decision record must explain why a stronger layer is not suitable.

## Required decision flow

### 1. Observe and reproduce

State the actual symptom and obtain evidence where practical. Do not start from the proposed patch.

### 2. Identify root cause

Ask:

- What structural fact made this possible?
- Can the same fact produce other symptoms?
- Are several findings in this area evidence of one shared cause?

Three repeated patches/findings around the same construct are a mandatory local-vs-global review signal.

### 3. State the target property

Write the invariant independently of current code.

Bad: “replace this `BeginTx`.”  
Good: “application transaction lifecycle is owned by one unit-of-work boundary.”

### 4. Name authority and boundary

Identify the owning module or platform primitive, the single source of truth, and the strongest reasonable enforcement boundary.

A solution that creates a second authority is presumed wrong until justified.

### 5. Evaluate local and global candidates

For each credible candidate, answer:

- Does it remove the root cause or only this symptom?
- What defect class remains representable afterward?
- What complexity does it add now?
- What complexity does it avoid later?
- Is that complexity essential or accidental?

### 6. Choose one legal outcome

Exactly one outcome must be recorded:

1. **Restructure now** — implement the global-maximum structure in the current work.
2. **Transitional solution** — use a bounded local maximum with a named successor and deletion gate.
3. **Stop and split prerequisite** — the correct fix crosses the current boundary; do not patch around it.
4. **Current structure confirmed** — the existing architecture is sound and a local correction is appropriate.

### 7. Define proof before implementation

Define how the target property will be demonstrated: RED/counterfactual proof where useful, GREEN behavior, targeted tests, contract/schema/static checks as appropriate, and broader regression only when boundaries were crossed.

Proof validates the property, not merely the implementation detail.

### 8. Implement and simplify

After correctness is established, run a subtractive pass:

- remove duplicate paths;
- remove obsolete transitional mechanisms;
- remove dead compatibility layers;
- consolidate authorities;
- keep only enforcement that still protects a distinct property.

### 9. Close with evidence

Work is complete when root-cause disposition is explicit, the target property holds, no known material contradiction remains, relevant proofs are green, transitional debt has a named deletion condition, and review findings are resolved, disproved, or explicitly deferred with ownership.

## Mandatory-use rules

A written decision record is required for:

- non-trivial bug fixes;
- repeated defects/findings in one area;
- refactors with architectural consequences;
- simplification of enforcement or abstractions;
- new guards, linters, verifiers, frameworks, or platform primitives;
- cross-module changes;
- authorization, tenant, transaction, idempotency, async, persistence, contract, or schema changes;
- explicitly transitional designs;
- remediation planning or issue consolidation by root cause;
- changes where the correct fix may cross an ownership boundary.

A formal record is normally unnecessary for pure spelling/formatting, deterministic generated updates, tiny factual comment fixes, or prescribed mechanical edits with no semantic choice. If such work exposes an architectural contradiction, this method becomes mandatory.

Keep the record proportional to the change. A small bug can require only a short record; a redesign may become a spec or ADR.

## Engineering Decision Record

```markdown
## Engineering Decision Record

### Symptom

### Root cause

### Target property

### Authority and boundary

### Local-maximum candidate

### Global-maximum candidate

### Decision
Restructure now | Transitional solution | Stop and split prerequisite | Current structure confirmed

### Enforcement
Strongest reasonable layer and why.

### Proof
RED/GREEN/contract/schema/static/regression evidence.

### Transitional exit
Successor + deletion condition, or N/A.
```

## Guard / lint / verifier policy

A custom guard is justified only when all are true:

1. it protects a material property;
2. the defect class is reachable or has occurred;
3. a stronger structural/type/schema/runtime boundary cannot reasonably express the complete property yet;
4. an existing standard tool does not already provide the needed enforcement;
5. the guard's maintenance cost is lower than the recurring defect risk.

Repeated syntax-specific patches to a guard are a structural signal. Re-evaluate the enforcement boundary rather than indefinitely hardening source spelling.

A sophisticated semantic analyzer can still be the correct solution when language semantics are genuinely part of the property. Complexity alone is not grounds for deletion.

## Transitional enforcement

Every transitional mechanism must state:

- the property it protects now;
- why the global maximum cannot land in the current slice;
- the named global-maximum successor;
- the milestone/slice that removes it;
- deletion as part of that successor's definition of done.

A transitional mechanism without an explicit removal condition is treated as permanent design and must meet the same bar as permanent architecture.

## Review and convergence

Review is adversarial but bounded.

- Verify findings against source before applying them.
- A reviewer finding is evidence to investigate, not an instruction to obey automatically.
- Repeated findings at the same architectural altitude trigger root-cause/global-maximum analysis instead of another patch round.
- Optional hardening does not become a blocker without a material property at risk.

Stop reviewing when the root cause and target property are settled, remaining findings are mechanical or non-material, the chosen boundary is coherent with target architecture, proofs are green, and no material contradiction remains.

**Global maximum is not permission for endless perfection search. Reopen a settled decision only on a new material finding or changed constraint.**

## Cross-project direction

MetalDocs is the proving ground for this method, not the permanent owner of the generic doctrine. After it has been exercised successfully on real changes, the project-neutral core should be promoted into MNFS/harness as a versioned shared engineering method. Project repositories should then retain their specific invariants, truth hierarchy, routing bridges, and deliberate extensions.
