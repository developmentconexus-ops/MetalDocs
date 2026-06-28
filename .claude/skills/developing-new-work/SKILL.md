---
name: developing-new-work
description: >-
  Run BEFORE brainstorming or designing any new MetalDocs backend module or feature — the
  pre-design system-impact orientation gate. Walks a static, pre-baked checklist of the whole
  system (the 14 bounded-context modules, the 6 non-negotiable invariants, capability and module
  wiring touchpoints, reusable platform frameworks, contract + DB conventions, test/QA gates, and
  docs/ADR governance), judges whether the work fits the architecture and whether the foundation is
  sound, then emits a written system-impact analysis plus a Green / Yellow / Red verdict. Red is a
  hard block on design. Use whenever the operator says "new module", "new feature", "add X",
  "build Y", "implement Z", "should we build…", "does this fit our architecture / our system", or
  kicks off an SP-/increment of a program — even when they never say "orientation" or "impact".
  This is architecture orientation BEFORE design, distinct from code blast-radius
  (gitnexus-impact-analysis) and from mission/milestone program planning.
---

# Developing New Work — system-impact orientation

CLAUDE.md carries two meta-rules that are easy to nod at and skip: the **Orientation rule**
(before planning new work, state owning module, invariants, read the owning wiki doc) and
**Global Maximum, Not Local Maximum** (judge the foundation first; never optimize inside a patch).
This skill turns both into a binding gate that runs *before* design and leaves an auditable record.

Given a one-line intent ("add a tenant token dictionary"), you walk a **static, pre-baked map** of
the system, judge fit and foundation, and write a **system-impact analysis** with a verdict. On
Green/Yellow you hand the analysis to `superpowers:brainstorming` as the rails it designs within.
On Red you stop — design cannot begin until the named redesign gate clears.

## The engine: a static checklist, not a rescan

**Do not re-analyze the codebase on every run.** The MetalDocs backend base is mature and frozen
(Grade-A, signed off 2026-06-21; module boundaries A). The 14 modules, the middleware chain, the
invariants, and the wiring touchpoints are settled contracts. Re-deriving them per run buys nothing
and burns thousands of tokens.

Everything you need is **inline in the `references/` files** — read them, not the code. Each item
carries a `file:line` or wiki-REQ anchor for one reason: so that when (and only when) a single item
is genuinely uncertain, you can verify *that one anchor* — never re-map the system. The references
mirror guards the repo already trusts (`TestCapabilityRegistrySize`, `TestEveryCapabilityClassified`,
`module-boundaries.yml`, `check-test-discipline.sh`), so they are agent-facing copies of machine
checklists, not a second source of truth.

When the checklist and the code disagree, **the code wins** (CLAUDE.md runtime-truth rule) — and the
disagreement is a signal to refresh the checklist (bump its `Last verified` stamp), not to patch
around it. CI guards are the backstop for anything a stale item missed.

## Workflow

Create a TodoWrite item per phase and work them in order.

1. **Orient.** Classify the work: **module** or **feature**. Name the owning module(s), the modules
   that explicitly do *not* own it, and the cross-module edges (with direction — who depends on
   whom). Read `references/` checklist files for the branch you're on. **No code reads here.**
   If the owning module is ambiguous → **AS-3**.

2. **Foundation.** Judge the base you'd build on. Is it sound, or is it legacy / a patch / a
   workaround? If you'd be optimizing inside a patch, name the global-maximum structure and its
   trade-off — or **AS-2**. (This is the Global-Maximum rule made concrete.)

3. **Invariants.** Walk `references/invariant-checklist.md` — the 6 non-negotiables. For each:
   touched? how satisfied? which helper to reuse? Any violation → **AS-1**.

4. **Wiring.** Walk `references/capability-wiring.md` if the work adds a capability, and/or
   `references/module-wiring.md` if it births a module. Skip the branch that doesn't apply.

5. **Frameworks.** Walk `references/frameworks-catalog.md` — reuse the platform primitives, do not
   reinvent them. Every hand-rolled equivalent of `TxRunner`, `problem.Write`, the outbox repo, or
   the `testdb` factory is a defect, not a choice.

6. **Test/QA + Docs/ADR.** Walk `references/test-qa-gates.md` and `references/docs-adr-governance.md`.
   Decide the canonical test framework, which QA gates apply, the evidence shape, the wiki docs to
   touch, and whether an ADR is required (a MUST-deviation or a policy change ⇒ yes).

7. **Targeted verify.** For any item the run cannot answer with confidence, read its *single* anchor
   (1–2 files maximum). This is the only code you read. If you find yourself opening a third file,
   stop — you're rescanning, not verifying.

8. **Verdict + handoff.** Fill `templates/system-impact-analysis.md`, write it to
   `docs/superpowers/analysis/YYYY-MM-DD-<slug>-system-impact.md`, and commit it.
   - **Green / Yellow** → invoke `superpowers:brainstorming`, passing the analysis path as the
     locked constraints it designs within.
   - **Red** → stop. Surface the redesign gate. Do not invoke brainstorming.

## Module vs feature

The artifact has the **same ten sections** either way. A **feature** marks module-only rows
(section 5, and the module-birth parts of 4 and 9) **N/A** with a one-line reason, and is mostly
invariants + frameworks + test/QA + verdict. A **new module** fills everything. Don't drop sections —
mark them N/A so the record shows the question was asked and answered.

## Hard-stops

A hard-stop means: stop the run, record the reason in the artifact, and resolve it before design.
Any unresolved hard-stop forces a **Red** verdict, which blocks the brainstorming handoff.

| ID | Trigger | Action |
|----|---------|--------|
| AS-1 | A non-negotiable invariant (§3) would be violated | STOP. Record the violation. Require an ADR or a redesign before any design. |
| AS-2 | The foundation is a patch/legacy and the work would optimize *inside* it | STOP. Propose the global-maximum structure + trade-off. The operator decides. |
| AS-3 | The owning module is ambiguous | STOP. Resolve the boundary (targeted-verify the candidate anchors) before continuing. |

These exist because the cheapest place to catch an architecture mistake is before a line of design
is written. A wrong owning module or a violated invariant discovered during implementation costs a
rewrite; discovered here it costs a sentence.

## Verdict semantics

- **Green** — fits cleanly; proceed to brainstorming.
- **Yellow** — proceed, but a named risk or an ADR is flagged and carried into the design as a locked
  constraint (e.g. "new capability ⇒ bump `TestCapabilityRegistrySize`", "supersedes ADR 0008").
- **Red** — a hard-stop is unresolved. Design is blocked until the redesign gate clears.

## Reference files

Read the ones relevant to the branch; don't bulk-load all six.

| File | Read when |
|------|-----------|
| `references/invariant-checklist.md` | Always — the 6 non-negotiables + the helper to reuse for each. |
| `references/capability-wiring.md` | The work adds or changes an IAM capability (10 ordered touchpoints). |
| `references/module-wiring.md` | The work births a new module (ordered birth checklist). |
| `references/frameworks-catalog.md` | Always — the reuse-don't-reinvent table. |
| `references/test-qa-gates.md` | Always — canonical test framework, R1–R4, the 6 QA gates, evidence shape. |
| `references/docs-adr-governance.md` | Always — wiki doc structure, REQ-ID citation, when an ADR is required. |

Each reference carries a `Last verified` stamp. If you targeted-verify an anchor and find it moved,
fix the anchor and bump the stamp — the references are living, governed the same way the wiki is.

## Output template

`templates/system-impact-analysis.md` is the exact ten-section shape to fill. Don't improvise the
structure — the fixed shape is what makes the record auditable and what brainstorming reads back.
