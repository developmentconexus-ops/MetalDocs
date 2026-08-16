---
name: developing-new-work
description: >-
  Run before brainstorming or designing a new MetalDocs module or feature. This is the repo-specific
  pre-design system-impact gate: establish ownership, invariants, foundation quality, reusable
  boundaries, proof strategy, and a Green/Yellow/Red verdict before design begins.
---

# Developing New Work — MetalDocs system-impact specialization

> **Scope:** repo-specific specialization of the DevelopmentConexus Engineering Method, **not a second organizational engineering method**.

## Canonical engineering method

Before judging foundation or design direction, read:

`docs/engineering/standards/root-cause-global-maximum-method.md`

Canonical authority remains `developmentconexus-ops/conexus-methodology/METHOD.md`; MetalDocs currently consumes version `1.0.0` through that local mirror.

This skill operationalizes the canonical Method for MetalDocs pre-design work. It may add a repository-specific system-impact gate, but it MUST NOT redefine the Method's decision vocabulary, outcomes, YAGNI rule, authority model, proof obligations, or reopen rules. If this skill conflicts with the Method inside the Method's scope, surface the conflict instead of reinterpreting it locally.

MetalDocs-specific product invariants and current architecture/status remain governed by the authorities routed from `AGENTS.md`.

## Purpose

Catch wrong ownership, violated invariants, local-maximum foundations, and duplicated platform primitives before design and implementation make them expensive.

On **Green/Yellow**, pass the system-impact analysis into design as repository-specific constraints. On **Red**, stop until the named prerequisite or redesign gate clears.

## Workflow

1. **Canonical method.** Apply the Method's Decision Core proportionally; do not invent a parallel root-cause/global-maximum vocabulary.
2. **Orient.** Classify module vs feature. Name owning module(s), modules that do not own it, and cross-module edge direction. Read the owning module/wiki authority. Ambiguous ownership -> **AS-3**.
3. **Foundation.** State the target invariant/property, authority/boundary, credible local/global candidates, chosen Method outcome, enforcement and proof. Building inside a known patch/local maximum -> **AS-2**.
4. **Invariants.** Walk `references/invariant-checklist.md`. For every touched invariant, state how the design preserves it and which existing primitive/boundary it reuses. Unresolved violation -> **AS-1**.
5. **Wiring.** Read `references/capability-wiring.md` for capability work and `references/module-wiring.md` for a new module.
6. **Frameworks.** Walk `references/frameworks-catalog.md`. Do not create a parallel transaction, problem, outbox, authorization, test, or other platform path when an established authority already owns it. Any proposed new mechanism must satisfy the canonical Method.
7. **Test/QA + docs.** Walk `references/test-qa-gates.md` and `references/docs-adr-governance.md`. Define proportional proof, QA, documentation, and ADR needs before implementation.
8. **Targeted verify.** Read source only where a material premise is uncertain. Repository/runtime truth wins for current-state claims. If evidence changes the root cause or target structure, return to the Method's Decision Core.
9. **Verdict + handoff.** When this repo-specific gate is required, record the system-impact analysis at `docs/superpowers/analysis/YYYY-MM-DD-<slug>-system-impact.md` using `templates/system-impact-analysis.md`.
   - **Green / Yellow** -> proceed to design.
   - **Red** -> stop and surface the prerequisite/redesign.

## Hard stops

| ID | Trigger | Action |
|---|---|---|
| AS-1 | A non-negotiable MetalDocs invariant would be violated | STOP. Redesign or obtain the governing decision before design. |
| AS-2 | Work would optimize inside a patch/local maximum | STOP. Apply the canonical Method; use one of its legal outcomes. |
| AS-3 | Owning module/boundary is ambiguous | STOP. Resolve ownership before design. |

## Verdicts

These verdicts are **MetalDocs workflow routing**, not replacements for the Method's decision outcomes:

- **Green** — fits cleanly; proceed.
- **Yellow** — proceed with a named bounded risk/decision carried into design.
- **Red** — unresolved hard stop; design blocked.

## References

Read only what applies:

- `references/invariant-checklist.md` — always when this gate is invoked.
- `references/frameworks-catalog.md` — always when this gate is invoked.
- `references/test-qa-gates.md` — always when this gate is invoked.
- `references/docs-adr-governance.md` — always when this gate is invoked.
- `references/capability-wiring.md` — capability work.
- `references/module-wiring.md` — new module.

If a reference is stale, repair governed memory rather than compensating with a local patch.

## Output

Use `templates/system-impact-analysis.md` only as this repo's system-impact handoff format. It is not an organizational Method template or a second methodology authority. Keep it proportional and use the canonical Method's vocabulary/outcomes rather than recreating them.
