# Shared brief — total engineering inventory, discovery lanes

You are one of ten parallel discovery lanes producing the evidence base for a whole-repo engineering
remediation program. Repo: MetalDocs, a Go + TypeScript + Postgres multi-tenant modular monolith
(15 backend modules under `internal/modules/`, ~180k Go LOC, frontend at `frontend/apps/web`).

## The goal, in the operator's words

Not Google-scale gold-plating. A codebase at a **solid professional engineering level**: clear
dependencies, clear modules, clear consumable surfaces, no duplication, no hand-maintained
redundancy, no "AI slop" where everything is implemented ad hoc and by hand. Following software
rules that have existed for decades. Optimised for efficiency, security, and above all
**maintainability and the ability to scale later**.

## YOUR JOB IS DISCOVERY, NOT VERDICT

Report **what is there**, with evidence. Do **not** recommend restructuring, do not propose a target
architecture, do not say what "should" be merged, split, or rewritten. A separate synthesis step owns
those judgments and has a method for them that your evidence would corrupt.

The one exception: where a thing is *objectively* a defect against a rule the repo already states
(CLAUDE.md, `wiki/architecture/backend-target-architecture.md`, an ADR), say so and cite the rule.

## Report classes

Tag every finding with exactly one:

- `duplication` — the same logic implemented more than once
- `hand-sync` — a fact that must be kept identical across N places by human discipline
- `layering` — a dependency that crosses a boundary the repo's own rules forbid
- `drift` — two sources of one truth that can silently disagree
- `gap` — a control, test, or mechanism that is absent or does not fire
- `idiom` — a language- or framework-level practice that a competent Go/TS reviewer would flag
- `hazard` — a correctness, concurrency, or security risk

## Sizing is mandatory

Every finding carries a **scale**: how many call sites, files, lines, or occurrences. "Duplicated
error handling" is not a finding; "`writeProblem` defined 10× with an identical signature across
audit/auth/security and 7 files in iam" is. Count with a command and show the command.

## Evidence discipline

- Every load-bearing claim cites `file:line` or a command you ran and its output.
- Where you could not verify, write `unverified` — never assert.
- Do not trust comments, doc-comments, or wiki docs as evidence about the code. Read the code.
- Prefer `rg`/`grep` counts over impressions. Prefer reading 3 representative sites over skimming 30.

## What is ALREADY KNOWN — do not re-derive, do not re-report

These are established. Reference them if your lane intersects, but spend no budget rediscovering:

- 7 module-level import cycles: `iam↔auth`, `iam↔taxonomy`, `iam↔security`, `documents↔approval`,
  `documents↔controlleddocuments`, `controlleddocuments↔approval`, `taxonomy↔approval`.
- `scripts/check-module-boundaries.ps1` asserts the *layer* of the import target only; it builds no
  graph, so cycles are invisible to it.
- `internal/platform` imports modules on 20 edges across 6 packages; `platform/tenantdata` imports 12
  of 15 modules. This inverts the rule in `wiki/architecture/backend-target-architecture.md`.
- 5 hand-maintained allowlist files in `scripts/api-lint/` totalling 221 lines.
- `problem.New(` 232 call sites vs `problem.NewFor(` 11, where `NewFor` is the ADR-0089 default.
- `func writeProblem` defined 10× with identical signature.
- Only 2 of 15 modules call `otel.Tracer`; events persist a plain `TraceID`, not W3C span context.
- ADR 0093 (2026-08-07) rules that documents/controlleddocuments/templates is one bounded context and
  that a template is a version-scoped role. **Do not re-litigate this and do not use it to justify
  findings** — it is a design ruling, not yet implemented.
- The register `docs/engineering/mechanical-enforcement-register.md` (ME-01..ME-13) catalogues known
  hand-enforcement defects. Read it so you do not duplicate an entry; say "already ME-nn" if you land
  on one.

## Output contract — follow exactly

**1. Write your full report** to `docs/superpowers/analysis/inventory/<LANE>.md` with this shape:

```markdown
# Lane: <name>

## Findings
| ID | Class | Finding | Evidence | Scale |
|----|-------|---------|----------|-------|
| <LANE>-01 | duplication | one sentence | `file:line` | 14 sites |

## The five heaviest, with detail
(one short paragraph each: what it is, why it costs, what it blocks)

## What is actually fine
(things a remediation program should NOT touch, with why — this is as valuable as the findings)

## Unverified / needs judgment
(questions your lane could not settle, phrased as questions)

## Commands run
```

**2. Return, as your final message text, a compact summary** — the findings table plus the five
heaviest in two sentences each, and the "actually fine" list. Under 60 lines. The file is the record;
the message is what the synthesis step reads. **Both are required** — a past run silently wrote an
empty file.

Terse. No preamble, no praise, no restating this brief.
