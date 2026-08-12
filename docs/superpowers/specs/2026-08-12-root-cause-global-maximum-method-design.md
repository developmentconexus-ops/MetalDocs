# Root-Cause / Global-Maximum Engineering Method — Design

**Status:** Proposed for operator review  
**Date:** 2026-08-12  
**Scope:** MetalDocs first; designed for later promotion into a shared MNFS engineering doctrine.

## 1. Purpose

MetalDocs needs one explicit engineering doctrine for non-trivial changes so correctness does not depend on an individual agent session, prompt, model, or reviewer.

The doctrine must preserve this binding principle:

> **Always simplify the code, never simplify correctness. Find the root cause, test whether the proposed solution is only a local maximum, and prefer the global structure that makes the defect class unrepresentable or mechanically impossible at the strongest reasonable boundary.**

This is not a mandate for maximum sophistication. It is a mandate to distinguish essential complexity from accidental complexity and to remove the latter without weakening invariants.

## 2. Problems this method prevents

The method exists to prevent five recurring failure modes:

1. **Patch-on-patch:** a visible symptom is corrected without changing the structural fact that made it possible.
2. **Local-maximum optimization:** engineering effort improves a legacy/workaround structure instead of replacing the faulty foundation.
3. **False simplification:** enforcement or proof is deleted because it is large or inconvenient, while the invalid state remains representable.
4. **Overengineering:** every defect creates a new framework, parser, guard, configuration layer, or second source of truth when a simpler existing boundary could enforce the same property.
5. **Infinite review:** reviewers keep inventing hypothetical hardening after the material property is already structurally solved and proved.

## 3. Canonical definitions

### 3.1 Symptom

The observable failure: wrong response, race, duplicated implementation, CI escape, incorrect write, broken import boundary, etc.

A symptom identifies where the defect appeared. It does not by itself identify where the defect should be fixed.

### 3.2 Root cause

The structural fact that made the defect possible.

Examples:

- not “taxonomy ignored a tenant error,” but “multiple modules were allowed to define independent idempotency identity policies”;
- not “one worker forgot rollback,” but “application code can own transaction lifecycle in dozens of places”;
- not “one request DTO drifted,” but “the OpenAPI contract and hand-authored runtime shapes were both treated as authorities.”

### 3.3 Target property / invariant

The statement that must remain true for every valid implementation.

Examples:

- every idempotency replay is scoped by a valid tenant and actor;
- application transaction lifecycle has one owner;
- generated contract types are the runtime request authority;
- a foreign module cannot write another module's persistence directly.

### 3.4 Local maximum

A solution that meaningfully improves the current implementation but preserves the structural limitation that produced the defect class.

A local maximum is legal only when explicitly transitional, with a named successor and a deletion condition.

### 3.5 Global maximum

The best sustainable structure for the actual system constraints: one that resolves the root cause, converges authorities, preserves required invariants, minimizes accidental complexity, and makes invalid states impossible or mechanically detectable at the strongest reasonable boundary.

Global maximum does **not** mean:

- maximum abstraction;
- maximum code;
- copying big-tech infrastructure without need;
- perfect future-proofing;
- indefinite redesign.

### 3.6 Essential vs accidental complexity

**Essential complexity** comes from real domain or system constraints: multi-tenancy, transactional consistency, authorization, asynchronous delivery, immutable evidence, contract evolution, etc.

**Accidental complexity** comes from the chosen implementation: duplicate policies, parallel abstractions, hand-synced registries, compatibility layers with no current need, redundant guards, speculative configuration, and repeated lifecycle code.

The method removes accidental complexity. It does not weaken essential complexity.

### 3.7 YAGNI

YAGNI removes speculative capability, not required correctness.

YAGNI can justify deleting:

- unused extensibility;
- a second abstraction with no second consumer;
- configuration nobody sets;
- compatibility paths with no supported legacy consumer;
- duplicate enforcement that proves exactly the same property.

YAGNI does **not** justify deleting:

- an invariant enforcement point;
- a fail-closed default;
- a test for a reachable state;
- a boundary that prevents a known defect class.

## 4. Enforcement hierarchy

When choosing how to preserve a property, prefer the strongest reasonable mechanism:

1. **Structure / API makes the state unrepresentable**
2. **Type system makes the state invalid**
3. **Database/schema constraint makes the state invalid**
4. **Runtime boundary fails closed**
5. **Test proves the reachable behavior**
6. **Lint / static guard detects the violation**
7. **Documentation / convention**

This is a preference order, not a rigid requirement. A lower layer is valid when a stronger layer cannot represent the property without disproportionate cost or loss of legitimate flexibility.

Whenever a lower layer is chosen, the decision record must explain why a stronger one was not suitable.

## 5. Required decision flow

For every non-trivial change covered by this method, work in this order.

### Step 1 — Observe and reproduce

State the actual symptom and obtain evidence where practical.

Do not start from the proposed patch.

### Step 2 — Identify root cause

Ask:

- What structural fact made this possible?
- Is the same structural fact capable of producing other symptoms?
- Are multiple findings in this area evidence of one shared cause?

Three repeated patches/findings around the same construct are a mandatory local-vs-global review signal.

### Step 3 — State the target property

Write the invariant independently of the current code.

Bad: “replace this BeginTx.”  
Good: “application transaction lifecycle is owned by one unit-of-work boundary.”

### Step 4 — Name authority and boundary

Identify:

- the owning module or platform primitive;
- the single source of truth;
- the strongest reasonable enforcement boundary.

A solution that creates a second authority is presumed wrong until justified.

### Step 5 — Evaluate local and global candidates

For each credible candidate, answer:

- Does it remove the root cause or only this symptom?
- What defect class remains representable afterward?
- What complexity does it add now?
- What complexity does it avoid later?
- Is the complexity essential or accidental?

### Step 6 — Choose one legal outcome

Exactly one of these outcomes must be recorded:

1. **Restructure now** — implement the global-maximum structure in the current work.
2. **Transitional solution** — use a bounded local maximum while explicitly naming the successor and deletion gate.
3. **Stop and split prerequisite** — the correct fix crosses the current boundary; do not patch around it.
4. **Current structure confirmed** — analysis shows the existing architecture is sound and a local correction is appropriate.

### Step 7 — Define proof before implementation

Define how the property will be demonstrated:

- RED/counterfactual proof where useful;
- GREEN behavior;
- targeted tests;
- contract/schema/static checks as appropriate;
- broader regression only when boundaries were crossed.

Proof should validate the target property, not merely the implementation detail.

### Step 8 — Implement and simplify

After correctness is established, perform a subtractive pass:

- remove duplicate paths;
- remove obsolete transitional mechanisms;
- remove dead compatibility layers;
- consolidate authorities;
- keep only enforcement that still protects a distinct property.

### Step 9 — Close with evidence

A change is complete when:

- root cause disposition is explicit;
- target property holds;
- no known material contradiction remains;
- tests/verifiers relevant to the property are green;
- transitional debt has a named deletion condition;
- review findings are resolved or explicitly disproved/deferred with ownership.

## 6. When this method is mandatory

A written decision record is required for:

- non-trivial bug fixes;
- repeated defects/findings in one area;
- refactors with architectural consequences;
- simplification of existing enforcement or abstractions;
- new guards, linters, verifiers, frameworks, or platform primitives;
- cross-module changes;
- authorization, tenant, transaction, idempotency, async, persistence, contract, or schema changes;
- any explicitly transitional design;
- remediation planning or issue consolidation by root cause;
- changes where the proposed fix crosses an ownership boundary.

A formal record is normally unnecessary for pure spelling/formatting, deterministic generated updates, tiny factual comment fixes, or prescribed mechanical edits with no semantic choice. If such work exposes an architectural contradiction, the method becomes mandatory.

## 7. Decision record template

Keep the record proportional to the change. A small bug may need ten lines; a redesign may become a spec or ADR.

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

## 8. Guard / lint / verifier policy

A custom guard is justified when all are true:

1. it protects a material property;
2. the defect class is reachable or has occurred;
3. a stronger structural/type/schema/runtime boundary cannot reasonably express the complete property yet;
4. an existing standard tool does not already provide the needed enforcement;
5. the guard's maintenance cost is lower than the recurring defect risk.

Repeated syntax-specific patches to a guard are a structural signal. Re-evaluate the enforcement boundary rather than indefinitely hardening source spelling.

A sophisticated semantic analyzer can still be the correct solution when language semantics are genuinely part of the property. Complexity alone is not grounds for deletion.

## 9. Transitional enforcement

Every transitional mechanism must state:

- the property it protects now;
- why the global maximum cannot land in the current slice;
- the named global-maximum successor;
- the milestone/slice that removes it;
- deletion as part of that successor's definition of done.

A transitional mechanism without an explicit removal condition is treated as permanent design and must meet the same bar as permanent architecture.

## 10. Review and convergence

Review is adversarial but bounded.

- Verify findings against source before applying them.
- A reviewer finding is evidence to investigate, not an instruction to obey automatically.
- Repeated findings at the same architectural altitude trigger root-cause/global-maximum analysis instead of another patch round.
- Optional hardening does not become a blocker without a material property at risk.

Stop reviewing when:

- the root cause and target property are settled;
- remaining findings are mechanical or non-material;
- the chosen boundary is coherent with target architecture;
- proofs are green;
- no material contradiction remains.

Global maximum is not permission for endless perfection search. Reopen a settled decision only on a new material finding or changed constraint.

## 11. Documentation architecture

This design uses one canonical doctrine and several contextual bridges.

### Canonical method

Implementation target:

`docs/engineering/root-cause-global-maximum-method.md`

This file owns the definitions and decision flow.

### AGENTS.md

`AGENTS.md` must contain a short mandatory routing rule explaining when this method is required and what minimal decision record must exist before implementation. It must not duplicate the full doctrine.

### CLAUDE.md

`CLAUDE.md` keeps project invariants and a short bridge to the canonical method. Its existing Global Maximum section should stop being a separate competing definition.

### developing-new-work

The skill remains the pre-design workflow for new features/modules. It consumes this doctrine for foundation/root-cause/local-vs-global/YAGNI decisions and adds MetalDocs-specific system-impact checks.

### adversarial-review

The skill remains the review/convergence protocol. It consumes this doctrine rather than redefining root cause/global maximum/YAGNI independently.

## 12. Existing routing drift

Current `AGENTS.md` references `.agents/skills/...`, while the repository currently exposes the relevant project skills under `.claude/skills/...`.

Implementation must reconcile this routing drift rather than adding another dead link. The exact fix must follow current repository/tooling conventions discovered during implementation planning; this design does not assume whether `.agents/skills` should be restored or `AGENTS.md` should point to `.claude/skills`.

## 13. Cross-project promotion

MetalDocs is the first proving ground, not the permanent owner of the generic doctrine.

After the method has been used successfully on several real changes, extract the project-neutral core into MNFS/harness as a versioned shared engineering skill/method.

Target shape:

```text
Shared MNFS engineering doctrine
        |
        +-- MetalDocs invariants + bridges
        +-- Aurora invariants + bridges
        +-- MNFS project invariants + bridges
        +-- other projects
```

Project repositories should eventually own only their specific invariants, truth hierarchy, routing bridges, and any deliberate project-level extensions.

## 14. Non-goals

This work does not:

- redesign MetalDocs modules by itself;
- automatically rewrite current remediation issues;
- mandate a new custom verifier for the doctrine;
- add process gates to trivial edits;
- require every decision to become an ADR;
- replace existing domain-specific skills;
- make review or planning slower for mechanically obvious work.

## 15. Acceptance for implementation

The implementation is accepted when:

1. one canonical method exists under `docs/engineering/`;
2. `AGENTS.md` routes all non-trivial engineering work to it without duplicating the doctrine;
3. `CLAUDE.md`, `developing-new-work`, and `adversarial-review` reference the canonical definitions instead of competing definitions;
4. skill-path routing drift in `AGENTS.md` is resolved correctly;
5. the decision record and mandatory-use rules are explicit;
6. no new verifier/framework is created solely to enforce that agents read the document;
7. the result remains usable by Codex, Claude, GPT, and future agents;
8. a future promotion path to MNFS is documented but not prematurely implemented.
