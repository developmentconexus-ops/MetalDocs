---
name: developing-new-work
description: >-
  Run before brainstorming or designing a new MetalDocs module or feature. This is the pre-design
  system-impact gate: establish ownership, invariants, foundation quality, reusable boundaries,
  proof strategy, and a Green/Yellow/Red verdict before design begins.
---

# Developing New Work — system-impact orientation

## Canonical engineering doctrine

Before judging foundation or design direction, read `wiki/standards/root-cause-global-maximum-method.md`.

That document is the authority for root cause, local/global maximum, YAGNI, enforcement hierarchy, transitional design, legal outcomes, and the Engineering Decision Record. This skill only applies those definitions to MetalDocs pre-design work.

MetalDocs-specific invariants remain governed by `CLAUDE.md` and the target architecture/wiki.

## Purpose

Catch wrong ownership, violated invariants, local-maximum foundations, and duplicated platform primitives before design and implementation make them expensive.

On **Green/Yellow**, pass the system-impact analysis into design as locked constraints. On **Red**, stop until the named prerequisite or redesign gate clears.

## Workflow

1. **Canonical method.** Start the Engineering Decision Record from `wiki/standards/root-cause-global-maximum-method.md`.
2. **Orient.** Classify module vs feature. Name owning module(s), modules that do not own it, and cross-module edge direction. Read the owning module wiki. Ambiguous ownership -> **AS-3**.
3. **Foundation.** Record target property, authority/boundary, local-maximum candidate, global-maximum candidate, proposed outcome, enforcement, and proof. Building inside a known patch/local maximum -> **AS-2**.
4. **Invariants.** Walk `references/invariant-checklist.md`. For every touched invariant, state how the design preserves it and which existing primitive/boundary it reuses. Unresolved violation -> **AS-1**.
5. **Wiring.** Read `references/capability-wiring.md` for capability work and `references/module-wiring.md` for a new module.
6. **Frameworks.** Walk `references/frameworks-catalog.md`. Do not create a parallel transaction, problem, outbox, authorization, test, or other platform path when an established authority already owns it. Any proposed new guard/framework must satisfy the canonical method.
7. **Test/QA + docs.** Walk `references/test-qa-gates.md` and `references/docs-adr-governance.md`. Define proof, QA, documentation, and ADR needs before implementation.
8. **Targeted verify.** Read source only where a material premise is uncertain. Repository/runtime truth wins over stale reference text. If new evidence changes root cause or target structure, update the decision record.
9. **Verdict + handoff.** Fill `templates/system-impact-analysis.md` at `docs/superpowers/analysis/YYYY-MM-DD-<slug>-system-impact.md` and include the canonical Engineering Decision Record fields.
   - **Green / Yellow** -> proceed to design.
   - **Red** -> stop and surface the prerequisite/redesign.

## Hard stops

| ID | Trigger | Action |
|---|---|---|
| AS-1 | A non-negotiable MetalDocs invariant would be violated | STOP. Redesign or obtain the governing decision before design. |
| AS-2 | Work would optimize inside a patch/local maximum | STOP. Apply the canonical local-vs-global decision flow; operator chooses the legal outcome. |
| AS-3 | Owning module/boundary is ambiguous | STOP. Resolve ownership before design. |

## Verdicts

- **Green** — fits cleanly; proceed.
- **Yellow** — proceed with a named bounded risk/decision carried into design.
- **Red** — unresolved hard stop; design blocked.

## References

Read only what applies:

- `references/invariant-checklist.md` — always.
- `references/frameworks-catalog.md` — always.
- `references/test-qa-gates.md` — always.
- `references/docs-adr-governance.md` — always.
- `references/capability-wiring.md` — capability work.
- `references/module-wiring.md` — new module.

If a reference is stale, repair governed memory rather than compensating with a local patch.

## Output

Use `templates/system-impact-analysis.md`. Do not invent another root-cause/global-maximum vocabulary inside the analysis; use the canonical Engineering Decision Record fields.
