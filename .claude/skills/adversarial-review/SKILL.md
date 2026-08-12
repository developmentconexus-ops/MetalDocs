---
name: adversarial-review
description: >-
  Run when reviewing a MetalDocs design, plan, or diff with an independent model and when disposing
  of its findings. This is a bounded review protocol: verify anchors, require root-cause reasoning,
  test local-vs-global structure, run a subtractive pass, preserve module boundaries/invariants,
  and stop when findings converge.
---

# Adversarial Review — bounded review protocol

## Canonical engineering doctrine

`wiki/standards/root-cause-global-maximum-method.md` owns root-cause, local/global maximum, YAGNI, enforcement, transitional-solution semantics, and the legal engineering outcomes.

This skill does **not** redefine those concepts. It applies them during adversarial review and adds review-specific mechanics: disposition, attack targets, architectural checks, convergence, and stop conditions.

## Why this exists

An unbounded reviewer finds the nearest defect, proposes the nearest patch, and repeats. That can create patch-on-patch remediation while the structure producing the findings remains unchanged.

The reviewer is therefore required to ask whether each material finding is a symptom of a deeper structural cause before proposing a fix.

## 1. Finding disposition: root cause before patch

For every material finding:

1. verify the claim against source and cite `file:line` evidence;
2. state the root cause using the canonical method;
3. state the target property and strongest reasonable boundary;
4. evaluate whether the proposed fix is a local maximum or the global-maximum candidate;
5. only then propose the change.

A finding that names only the symptom is incomplete.

**Repeated-finding signal:** three findings in one area, three successive bounds on one guard, or two consecutive patches to the same construct require an explicit canonical local-vs-global analysis before another patch is applied.

## 2. Operator decision on structure

When a review discovers that the correct answer is structural, produce the canonical Engineering Decision Record and one of its four legal outcomes:

- **Restructure now**;
- **Transitional solution** with named successor and deletion condition;
- **Stop and split prerequisite**;
- **Current structure confirmed**.

The reviewer recommends; the operator owns a boundary-changing decision.

## 3. Subtractive pass

Every review round asks what can be removed **without weakening a distinct material property**.

Look for:

- two paths doing one thing;
- a second authority or registry for the same concept;
- compatibility layers with no supported consumer;
- speculative abstractions/configuration;
- obsolete transitional mechanisms whose successor already landed;
- duplicate enforcement proving exactly the same property.

YAGNI does not authorize removal of fail-closed defaults, reachable-state tests, invariant enforcement, or a guard that still protects a distinct property. Use the canonical method to justify removal or retention.

## 4. Module exposure standard

Cross-module communication uses published surfaces; consumers do not reach into another module's repository, SQL, or domain internals.

Review each changed edge:

| Layer | Published? |
|---|---|
| Application service methods | Yes |
| Published Go interfaces / ports | Yes |
| Domain value objects intentionally crossing the seam | Yes |
| Domain internals / aggregates / invariant implementation | No |
| Repositories | Never |
| SQL / tables / triggers | Never |

Ask:

- Is the owning module named?
- Is the edge direction explicit?
- Is the consumer redeclaring a concept the owner already publishes?
- Is a shared primitive still private after a real second consumer appeared?
- Can a consumer discover what to call without reading the owner's internals?

## 5. Adversarial posture

Existing implementation is evidence of current use, not proof of correctness.

- Verify load-bearing claims against source.
- Treat author rationale as a claim to test, not an instruction to accept.
- A reviewer finding is evidence to investigate, not an instruction to obey automatically.
- Disproving a finding with source evidence is a valid closure.
- Do not invent optional hardening blockers after the target property is already structurally solved and proved.

## 6. Binding architecture checklist

### Boundaries
- [ ] Owning module named for each change.
- [ ] No consumer reaches past a published surface.
- [ ] No duplicate authority introduced.
- [ ] Every new cross-module edge has direction and a published contract.

### MetalDocs invariants
- [ ] Authorization is capability-based, never role-based.
- [ ] Request lifecycle is not reinvented by a route.
- [ ] Public HTTP change is contract-first through OpenAPI/generated surfaces.
- [ ] Multi-tenant boundaries and tx-local context remain correct.
- [ ] Async state writes do not share a transaction with network side effects; consumers remain idempotent.
- [ ] Database constraints/triggers remain the final invariant backstop where applicable.

### Guards/backstops
- [ ] Stronger structural/type/schema/runtime enforcement was considered before adding static tooling.
- [ ] Every guard states what it trusts and fails closed on unsupported input where the property requires it.
- [ ] Negative/false-positive proof exists when exclusions or silence are material.
- [ ] Transitional guards name the successor and deletion gate.

### Executability
- [ ] Work follows dependency order, not only topic order.
- [ ] Each accepted slice ends with its declared proof green.
- [ ] No intermediate commit creates an unsafe state that later commits merely repair.

### Truth/vocabulary
- [ ] Load-bearing identifiers and claims are defined or source-cited.
- [ ] Shared vocabulary has one authority; no hand-synced second enumeration introduced without justification.
- [ ] Generated artifacts are regenerated, not hand-edited.

## 7. Round protocol

Each round has two jobs, in this order.

### Job 1 — disposition prior findings

One line each:

```text
N. CLOSED | PARTIAL | OPEN — file:line — one sentence.
```

Mentioning a finding is not closing it. Verify the revised source.

### Job 2 — attack only the new/revised material

Weight the review toward what changed since the previous round: new SQL, boundaries, state transitions, contracts, task ordering, type shapes, capabilities, guards, or proof assumptions.

Recommended prompt shape:

```text
READ-ONLY. Verify every claim against actual code; cite file:line.

Job 1 — dispose the prior numbered findings.
Job 2 — attack only the revised/new material.

For every material finding include:
SEVERITY — claim — file:line evidence — root cause — target property — what must change.

End with:
VERDICT: PROCEED | PROCEED WITH FIXES | DO NOT PROCEED — biggest remaining material risk.
```

## 8. Convergence and stop conditions

Track both **count** and **altitude** of findings.

Converging means findings decrease and move from structural/design concerns toward mechanical issues that build/lint/tests can catch.

Not converging means the same architectural altitude recurs. Stop the review loop and return to the canonical local-vs-global decision flow instead of adding another patch.

Stop when any of these is true:

- no blocker/material architecture finding remains and the target property is proved;
- remaining findings are mechanical/non-material;
- same-altitude recurrence has been escalated to an operator structural decision.

Do not close a `DO NOT PROCEED` verdict by silently accepting the remaining risk.

## 9. Per-round output

Record:

- prior-finding disposition;
- new findings;
- author disposition: `applied`, `disproved (file:line)`, or `deferred (owner + deletion/successor condition)`;
- canonical Engineering Decision Record for every structural fork;
- convergence line: finding count + altitude + stop/continue.

## Relation to other workflows

- `.claude/skills/developing-new-work/SKILL.md` — pre-design system-impact gate.
- `superpowers:writing-plans` — plan authoring before implementation.
- `docs/engineering/defect-class-catalog.md` — empirical defect patterns used as review evidence.
- External dispatch/transport tooling may launch the independent model; this skill governs the review semantics, not the transport.
