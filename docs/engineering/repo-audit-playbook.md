# Repo audit playbook — from "this codebase is a mess" to a sequenced program

A portable method for auditing a large, degraded, or AI-authored codebase and converting it into a
small number of executable workstreams. Nothing here is specific to one repository; substitute your
own lane list and stack.

The method exists because the naive approach fails in three predictable ways:

1. **The thousand-issue list.** An audit that emits 150 findings produces no action, because nobody
   can sequence 150 things. Findings must collapse into a handful of *causes*.
2. **The status-quo trap.** An analyst who reasons from the code as it exists can only ever conclude
   "leave it as it is", because the evidence was produced by the thing under examination.
3. **Churn against working code.** An audit with no explicit *don't touch this* list turns a
   remediation program into a rewrite, and the rewrite loses.

The method is four phases. Do not merge them — each one's discipline is what makes the next honest.

---

## Phase 0 — Decide what "good" means, in the operator's words

Before any tooling, write down the target in one paragraph, in the language of whoever owns the
codebase. Not "clean architecture" — something falsifiable and calibrated.

Calibration matters more than ambition. "Google-tier" and "solid professional level" produce
completely different programs, and the second is usually what is actually wanted: clear dependencies,
clear modules, clear consumable surfaces, no hand-maintained redundancy, following rules that have
existed for decades, optimised for maintainability and future scaling.

Write it once. Paste it verbatim into every brief that follows. Every agent that reads it makes the
same trade-offs, which is what makes ten parallel outputs comparable.

Also record the hard constraints now: budget (zero-spend changes every tooling answer), team size,
whether the repo is public, what cannot be broken.

---

## Phase 1 — Discovery in parallel lanes

### The core rule: the finder does not judge

Split the audit into lanes and give each lane one dimension. **Lanes report what is there. They do
not recommend restructuring, do not propose a target, and do not say what should be merged, split or
rewritten.**

This is not process ceremony. If the finder also judges, it judges using the evidence the status quo
produced — which is the failure mode in the introduction. Separating discovery from verdict is what
lets you apply a different evidence standard to the verdict later.

One exception: where something is objectively a defect against a rule the repo *already states*, the
lane says so and cites the rule.

### Lane list (adapt to your stack)

Ten is a good number: broad enough to cover the surface, small enough to synthesise.

| Lane | Question it answers |
|---|---|
| duplication | What logic is implemented more than once? |
| layering | What dependency crosses a boundary the repo's own rules forbid? |
| CI/CD topology | Which gates exist, which actually fire, and what blocks merge? |
| delivery / HTTP surface | How many dialects solve the same request-boundary concern? |
| persistence | How is data-access correctness maintained — by machinery or by hand? |
| testing | What does the test suite actually prove, and when does it prove it? |
| observability / ops | Can you tell what the system is doing in production? |
| security / config | Which controls are real, which are contingent on configuration? |
| language idiom | What would a competent reviewer of this language flag? |
| frontend (or second stack) | Same questions, other side of the wire |

Run them **concurrently** — they are read-only and independent. A capable model per lane, not the
cheapest; the quality of the evidence base determines everything downstream.

### The shared brief — copy this

Every lane gets the *same* brief with only its dimension swapped. Verbatim template:

```markdown
# Shared brief — <repo> engineering inventory, discovery lanes

You are one of N parallel discovery lanes producing the evidence base for a whole-repo
remediation program. Repo: <one-paragraph description: stack, size, module count, domain>.

## The goal, in the operator's words
<paste Phase 0 verbatim>

## YOUR JOB IS DISCOVERY, NOT VERDICT
Report what is there, with evidence. Do NOT recommend restructuring, do not propose a target
architecture, do not say what "should" be merged, split, or rewritten. A separate synthesis
step owns those judgments and has a method for them that your evidence would corrupt.

The one exception: where a thing is objectively a defect against a rule the repo already
states (<name the files>), say so and cite the rule.

## Report classes
Tag every finding with exactly one:
- duplication — the same logic implemented more than once
- hand-sync — a fact kept identical across N places by human discipline
- layering — a dependency crossing a boundary the repo's own rules forbid
- drift — two sources of one truth that can silently disagree
- gap — a control, test, or mechanism that is absent or does not fire
- idiom — a language/framework practice a competent reviewer would flag
- hazard — a correctness, concurrency, or security risk

## Sizing is mandatory
Every finding carries a scale: call sites, files, lines, occurrences. "Duplicated error
handling" is not a finding; "`writeProblem` defined 10x with identical signature across 3
modules" is. Count with a command and show the command.

## Evidence discipline
- Every load-bearing claim cites file:line or a command you ran and its output.
- Where you could not verify, write `unverified` — never assert.
- Do NOT trust comments, doc-comments, or docs as evidence about the code. Read the code.
- Prefer rg/grep counts over impressions. Prefer reading 3 representative sites over
  skimming 30.

## What is ALREADY KNOWN — do not re-derive
<list established facts so lanes spend no budget rediscovering them>

## Output contract — follow exactly
1. Write your full report to <path>/<LANE>.md with this shape:
   # Lane: <name>
   ## Findings          (table: ID | class | finding | evidence | scale)
   ## The five heaviest, with detail
   ## What is actually fine     <- as valuable as the findings
   ## Unverified / needs judgment
   ## Commands run
2. Return, as your final message text, a compact summary under 60 lines.
Both are required — a past run silently wrote an empty file.

Terse. No preamble, no praise, no restating this brief.
```

### Four details that carry most of the value

**Sizing converts opinion into engineering.** "Lots of duplication" is unactionable. "53 hand-written
clone pairs after excluding generated code and tests" is a decision. Requiring the command makes the
number auditable and catches malformed searches.

**"What is actually fine" is not politeness.** It is the do-not-touch list, and without it the
program becomes a rewrite. Some of the most useful outputs are of this shape: *interfaces are
consumer-defined, 37 in application vs 17 in infrastructure* — that is the correct pattern already in
place, and a remediation that churns it destroys value.

**Comments are not evidence.** A doc-comment describing what a module does is the code describing
itself, and in an AI-authored repo it is frequently confident and wrong.

**Demand both a file and a returned summary.** Agent runs write empty files. If the summary comes
back only in the file, you find out an hour later.

---

## Phase 2 — Synthesis under an inverted evidence standard

This is the phase where the method differs from an ordinary audit, and it exists because of a real
failure.

### The failure, stated plainly

Asked whether a three-module split was domain truth or an implementation accident, two independent
advisors both answered with: the database tables are disjoint; there is no route that would exist if
they were one; a prior design document already ruled against merging; the current transaction creates
both together.

Every one of those is a *consequence of the split being examined*. An argument of that shape can only
conclude "leave it as it is". It is a circular argument wearing evidence as a costume, and it is
uniquely hard to catch because it is **selective** — the analyst reasons correctly about everything
else, so the surrounding rigour vouches for the circular section.

When the question was re-asked with that evidence ruled inadmissible, both arms reversed.

### The rule

For any **structural** conclusion — a boundary, an ownership, a target shape:

**Inadmissible as an argument FOR a structure:**
- that the code is currently organised that way
- that a prior decision document said so
- that migration would be expensive
- that the import graph, route topology, or schema reflects it
- doc-comments describing the current design

**Admissible:**
- the problem domain and any standards or regulation governing it
- how mature systems in the same field solve it
- design principle with a **named failure mode** (not "clean architecture says so")
- the product's observable user-facing behaviour
- measured cost and measured scale

**Migration cost is admissible for *sequencing* and never as an argument that a target is right.**
"Do A before B because B is cheaper afterwards" is valid. "Keep A because changing it is expensive"
is the trap.

### The inversion test — mandatory, one line per conclusion

> State what would survive if the current implementation were the opposite in every respect.

A conclusion that cannot pass this is not a conclusion. In practice it takes one sentence and kills a
surprising share of confident recommendations:

> *One authorization relation: survives an opposite table arrangement, because route admission and
> in-transaction enforcement answer the same product question and must be views of one relation.*

> *Boot-proven non-bypass DB identity: survives an opposite schema design, because row-level security
> is ineffective for a bypassing connection in every topology.*

Write these down in the output. They are the audit trail for why a target is a target.

### Second rule: do not confuse a control with its effect

A control that exists and does not fire is **absent**. A linter scoped to new lines only, a gate
script referenced by zero pipelines, a scanner whose skip flag defaults to true, a database policy
inert against the connecting role. Counting these as "we have a control" is the same error class as
the circular argument — reasoning from an artifact rather than from an effect.

Where a control is inert, size the axis as if the control were absent, because it is.

### Run two independent arms

Give the same synthesis brief to two different models with no visibility into each other. Then
reconcile, and record the divergences and the resolution.

Convergence between independent arms is weak evidence of correctness — but **divergence is strong
evidence of an unsettled question**, and those are exactly the ones worth the operator's attention.
In practice the arms agreed on the sequence and disagreed on grouping and on the size of four
findings; every one of those disagreements was worth resolving explicitly.

### The synthesis brief — what to demand

1. **Axes, not a list.** Group findings into 6–10 axes where "connected" means *fixing them together
   is cheaper than separately, because they share a root cause, a mechanism, or a blast radius*. Name
   each axis for its **cause**, not its symptom. "The verifier is not one trusted product" beats
   "CI gaps". A finding belongs to exactly one axis — **if two axes want the same finding, they are
   one axis or the cause is wrong.**
2. **Kill the noise.** Which real findings do not earn a slot, and why — cost of fix versus cost of
   leaving it. Be willing to say "this is fine, leave it forever". A program that treats 150 findings
   as equal is how programs die.
3. **Consolidated do-not-touch list**, merged from every lane's "actually fine".
4. **The target methodology** — the gate topology, below. This is the half that outlives the refactor.
5. **The sequence** — dependency order, rough sizes in days, and which axis makes every later axis
   cheaper.
6. **Inversion tests** — one line per structural conclusion.
7. **Where you disagree with a lane** — a finding that is wrong or mis-sized.

---

## Phase 3 — The gate topology (the durable half)

A remediation that only fixes code decays back. What prevents decay is where the rule *fires*.

### The firing hierarchy — always climb as high as you can

| Level | Mechanism | Example |
|---|---|---|
| 1 | **Unrepresentable** | Generated types; a DB constraint; an API that cannot express the wrong state |
| 2 | **Boot-fatal** | The process refuses to start unless a precondition holds |
| 3 | **Red build** | A deterministic check fails the build |
| 4 | **Runtime assertion** | Fail closed at the owning boundary |
| 5 | **Discipline** | A doc, a comment, a checklist, a review convention |

**Level 5 is not a control.** Treat every level-5 rule with suspicion and record which level each
axis can actually reach and why not higher. "Make the defect unwritable" beats "detect the defect"
beats "document that the defect is forbidden".

### Blocking versus annotating

A blocking check must be **deterministic, locally reproducible, and carry a demonstrated failing
fixture**. Otherwise it will be wrong once, and being wrong once teaches everyone to bypass gates —
which costs more than the check was ever worth.

Everything judgment-shaped or false-positive-prone annotates: duplication ratios, dead-code
detection, coverage trends, complexity, performance until budgets are validated.

**The negative-fixture rule.** Every guard ships with an input that makes it fail, in the same
change. A guard nobody has ever seen fail is a guard nobody knows works. This is the single cheapest
high-value rule in the playbook — apply it to shell and CI scripts too, which is exactly where it is
usually skipped.

### The in-loop gate, when the author is a machine

The PR-review-bot model assumes a human author caught by a machine. When agents write the code, the
author *is* a machine, and its failure mode is **fluent, internally consistent wrongness** — the
wrong noun used correctly everywhere, including inside the guard written to catch it. Post-hoc
comments do not correct that. Two things do:

**Single-source vocabularies, compiled.** Wherever a noun matters — permissions, error codes, table
ownership, queue names — it lives in exactly one registry and everything else is generated from it.
An agent cannot consistently misuse a noun the compiler owns.

**One verify manifest, run during authoring.** A single entry point with profiles (`fast`, `changed`,
`pr`, `full`) that CI calls and *nothing else*, so "green locally" and "green in CI" are literally the
same claim. A fast tier cheap enough to run mid-edit — under about five minutes — wired as an editor
or agent hook. This is the highest-leverage artifact in the whole program.

Plus: an agent may add a test in its own change, but **that self-authored test may never be the sole
evidence** for a security, boundary, or migration invariant.

### Separation of powers

Ask one question: **can a single change modify both the code and the thing that judges the code?**

If yes, the path of least resistance for a stuck agent is to loosen the judge. Minimum viable
separation, in descending order of strength:

1. Platform enforcement — no direct push to the main branch, required checks (free on public repos;
   on most hosts, paid for private ones).
2. Credential custody — the author has no merge credential; only the operator merges.
3. Policy in a separate repository the author can only read, pinned by commit.
4. **Mixed-change ban** — a check that fails any change touching both product code and gate
   configuration. Cheap, works anywhere, and turns loosening into a visible standalone event.
5. Shrink-only ratchets — every allowlist carries a committed count; growth requires an explicit
   approved marker, making it a logged decision instead of a silent edit.

State honestly what each level does *not* buy. Detection is not prevention; say which one you have.

---

## Phase 4 — Issues and execution structure

### One issue per axis, not per finding

Child findings live *inside* the axis issue as sized evidence. Issue titles name the **cause**:

> A5 — Persistence correctness is maintained by visual agreement

not "Refactor the repository layer". The title should make someone who has never seen the codebase
understand what is broken.

### Issue body shape

```markdown
**Root cause.** One sentence. Why these findings are one thing.

**Evidence (sized).** Bullets, each with a number and a file:line or command.

**Deliverable.** What the target is. Concrete.

**Explicitly not in scope.** The adjacent tempting work this issue does NOT authorize.

**Acceptance.** The firing level reached, and why not higher.

**Sequencing.** What it depends on, what it blocks, rough size in days.
```

Two sections earn their place beyond the obvious:

**"Explicitly not in scope"** is what stops an axis becoming a rewrite. If a god-package is 8,755
lines, say that this proves a *hotspot*, not a decomposition, and that cohesion decides subpackages —
otherwise someone splits it by line count.

**Corrections recorded, not silently applied.** When synthesis resizes or overturns a lane's finding,
write that down in the issue and the program doc. The reasoning is the durable artifact; a silently
corrected claim teaches nobody and gets re-derived wrongly next time.

### Sequencing rules

- **The verification axis goes first, always.** Every later fix lands as *mechanism* if its gate
  fires before handoff, and as *discipline* if it does not. Discipline regresses. This is also the
  argument that survives inversion: it does not depend on what the code currently looks like.
- **Small, urgent, independent axes run parallel** with the first big one. Do not serialise work with
  no dependency.
- **Contract and data-access spines before boundary moves.** Moving modules is mechanical once
  request handling and queries are generated; before that it is hand-editing hundreds of call sites.
- **The largest structural merge goes last** among structural work, and only migration *cost*
  justifies that placement — never a claim that the target changed.
- Mark which axes are genuinely independent so they can be parallelised.

### Execution loop per axis

Open the issue → work it → open the PR referencing it → the PR closes it. One axis at a time per
person or agent; parallel only where the sequence marked it independent.

---

## Failure modes to expect (all observed in a real run)

| Failure | Fix |
|---|---|
| An advisor answers a "should this be different" question with evidence produced by the current design | Rule status-quo evidence inadmissible in the brief; require the inversion test |
| An agent writes an empty output file | Demand the answer both written **and** returned as final message text |
| A malformed search silently undercounts (`grep -o 'x\.NewFor\?'` requires the literal "Fo") | Require the command in the report; re-run load-bearing counts yourself |
| A count misses method receivers (`func writeProblem` vs `func (h *H) writeProblem`) | Widen the pattern; verify the heaviest findings by hand |
| An agent is dispatched read-only but its brief requires writing a file | Match the sandbox to the contract; check the tail of the log when a run produces nothing |
| Killing a wrapper process orphans or kills the real worker | Monitor a sentinel file or the process itself, not the log's freshness |
| A finding is stated more strongly than its evidence supports | Verify the top findings personally; resize in writing and say you resized |
| The audit's own memory or prior notes are treated as current fact | Re-verify before asserting; a note is a point-in-time observation |

The last two are the ones that matter. **Verify the load-bearing claims yourself** — not all of them,
but every one that a decision rests on. In a real run this changed several conclusions: one security
finding was overstated and had to be publicly resized; one "unverified" risk turned out to have been
verified months earlier in a record nobody consulted.

---

## Cost

Ten discovery lanes plus two synthesis arms is roughly 1.2–1.5M tokens for a ~200k-LOC repository,
running a few hours wall-clock with the lanes concurrent. That is the entire audit — the alternative
is weeks of reading, and it produces a *list* rather than a sequence.

Do not run this against a small codebase. Below roughly 20k LOC, read it.
