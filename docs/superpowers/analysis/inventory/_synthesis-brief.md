# Synthesis brief — from 10 discovery lanes to a sequenced remediation program

Ten parallel discovery lanes have finished. Their reports are on disk, ~900 lines total:

```
docs/superpowers/analysis/inventory/
  duplication.md  cicd.md  layering.md  http-delivery.md  persistence.md
  frontend.md  testing.md  observability.md  security-config.md  go-idiom.md
```

Read all ten. `_shared-brief.md` in the same directory is the brief they answered — read it too, so
you know what they were forbidden to do. **They were discovery-only: explicitly barred from
recommending restructuring or proposing a target.** That judgment is yours. Their evidence is
sized and `file:line`-cited; treat the sizes as load-bearing and the interpretations as absent.

Repo: MetalDocs. Go + TypeScript + Postgres multi-tenant modular monolith, 15 backend modules under
`internal/modules/`, ~180k Go LOC, frontend at `frontend/apps/web`. Regulated domain (ISO 9001 /
ISO 13485 / 21 CFR Part 11 eQMS). Nearly all code was written by AI agents under human direction.

## The operator's goal, in their words

Not Google-scale gold-plating. A codebase at a **solid professional engineering level**: clear
dependencies, clear modules, clear consumable surfaces, no duplication, no hand-maintained
redundancy, no "AI slop" where everything is implemented ad hoc and by hand. Following software
rules that have existed for decades. Optimised for efficiency, security, and above all
**maintainability and the ability to scale later**.

Equally weighted, and stated as its own requirement: **software-development methodology — CI, CD,
lint — and not merely having them, but ensuring they actually operate the way they must operate in a
repository as large as this one.** Plus anything Go-specific that helps.

Hard constraint: **zero spend**. Free / OSS tooling only. No paid SaaS, no paid tiers.

## What you must produce

### 1. Axes, not a list

The operator explicitly rejected "a thousand issues". Group every finding into a small number of
**connected axes** — where "connected" means: fixing them together is cheaper than fixing them
separately, because they share a root cause, a mechanism, or a blast radius. Name each axis for its
**cause**, not its symptom. Target roughly 6–10 axes. State for each: the root cause in one
sentence, which lane findings belong to it, its total measured scale, and what it blocks.

A finding may appear in only one axis. If two axes want the same finding, they are one axis or your
cause is wrong.

### 2. Kill the noise

Some findings are real but not worth a program slot. Say which, and why — cost of fix vs. cost of
leaving it. A remediation program that treats all 100+ findings as equal is how programs die. Be
willing to say "this is fine, leave it forever."

Also honour the reverse: every lane reported an "actually fine" section. Produce a consolidated
**do-not-touch list**. Churn against working code is a defect.

### 3. The target methodology — this is the half that outlives the refactor

Design the gate topology this repo should have. Concretely:

- **What blocks vs. what annotates.** Which checks must fail a build, which must merely report.
  Justify each: a blocking check that is wrong once teaches people to bypass gates.
- **Where each gate fires.** Unrepresentable-in-code > boot-fatal > red build > runtime assertion >
  discipline. For each axis, name the highest level actually achievable and why not higher.
- **The in-loop agent gate.** Post-hoc PR review assumes a human author caught by a machine. Here
  the author IS a machine whose failure mode is confident, internally consistent wrongness — the
  wrong noun used fluently everywhere, including inside the guard written to catch it. Design for
  that. What fires *during* authoring, not after?
- **Separation of powers.** Today one commit can change the code AND the lint config, the workflow,
  and the allowlist that judge it. `gh api .../protection` returns 403, so branch protection is
  unverifiable from here. Design the separation. Name what enforces it given a solo operator with
  agent authors and no paid tooling.
- **Concrete free tooling**, named, with what it replaces and what it costs to run.

### 4. The sequence

Order the axes by dependency: which must land before which, and why. Where two are independent, say
so — the operator can parallelise. Give each a rough size (days, not story points) and name the
axis that, if done first, makes every later axis cheaper.

Two known pre-existing items must be placed in this order, not ignored:
- ADR 0093 (2026-08-07) rules that `documents`/`controlleddocuments`/`templates` is one bounded
  context and a template is a version-scoped role. **Ruled, not implemented.** It is an axis.
- ADR 0092 grant-model unification (authz) has a written spec awaiting its slot. Both advisory arms,
  in both prior rounds, held it **must not be shelved** behind the artifact axis.

## Method constraints — these are binding, and one of them has already burned this program

**ME-13 (`docs/engineering/mechanical-enforcement-register.md`).** An analysis that takes its
subject as its own premise. Earlier in this same program, two independent advisory arms were asked
whether a three-module split was domain truth or implementation accident, and both answered with
evidence *produced by the split* — table ownership, route topology, an existing ADR. Every such
argument can only ever conclude "leave it as it is". Both answers were rejected and the question was
re-asked with status-quo evidence ruled inadmissible; both arms then reversed.

So, for any **structural** conclusion you reach (a boundary, an ownership, a target shape):

- **Inadmissible as an argument FOR a structure:** that the code is currently organised that way;
  that a prior ADR said so; that migration would cost; that the import graph or route topology
  reflects it.
- **Admissible:** the regulated domain and its standards; how mature systems in this field solve it;
  design principle with a named failure mode; the product's observable behaviour; measured cost.
- **Mandatory inversion test.** For each structural conclusion, state what would survive if the
  current implementation were the opposite in every respect. A conclusion that cannot pass this is
  not a conclusion.

Migration cost IS admissible for **sequencing** — just never as an argument that a target is right.

**Second constraint: do not confuse a control with its effect.** Several findings are controls that
exist and do not fire (13 unwired scripts, `only-new-issues: true` lint, govulncheck defaulted off,
99% of integration tests post-merge-only, RLS FORCE against a superuser connection). Counting these
as "we have a control" is the same error class. Where a control is inert, treat the axis as if the
control were absent, because it is.

## Output shape

Terse. Evidence-dense. No preamble, no restating this brief, no praise.

1. **Axes** — table (name, root cause, member findings, scale, blocks-what), then one paragraph each.
2. **Noise** — what to drop, one line each with the reason.
3. **Do-not-touch** — consolidated, from the lanes' "actually fine" sections.
4. **Target methodology** — gate topology per §3 above, including the in-loop agent gate and the
   separation-of-powers design, with named free tools.
5. **Sequence** — ordered, with dependencies, rough sizes, and the highest-leverage first axis.
6. **Inversion tests** — one line per structural conclusion.
7. **Where you disagree with a lane** — if a lane's finding is wrong or mis-sized, say so.

Return the full answer as your final message text. Also write it to
`docs/superpowers/analysis/inventory/_synthesis-<ARM>.md` where `<ARM>` is the arm name you were
given. **Both are required** — a past run silently wrote an empty file.
