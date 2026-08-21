# T8-H Fable Round 2 — bounded convergence review

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t8h-global-coherence`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Candidate PR: #148 — `docs(t8h): open Whole-T8 global coherence review`

Candidate branch: `arch/t8h-global-coherence`

Exact corrected candidate HEAD under review:

```text
b940d4e105a8b837ecdac7f71233ff10d735cd5e
```

Required candidate CI:

```text
#1108 SUCCESS
```

Review branch:

```text
review/t8h-fable-r2
```

This Round 2 is **bounded**. Do not restart the entire T8-H review from preference.

Repository current authority wins over this handoff/history/chat.

Read strictly:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ this file
→ only the smallest owning authority required to falsify one of the bounded questions below
```

Do not recursively read the repository unless a concrete contradiction requires another owner.

## Round-1 evidence

Round 1 Evidence PR:

```text
#149 — CLOSED / UNMERGED
reviewed candidate   96c99fa6edfed5b015c5f93dbec9a40d806b8c95
final Fable HEAD     4990a8b1c0e3123a135989577d0f13bda2175932
review CI            #1100 SUCCESS
verdict              NOT CONVERGED
MATERIAL             1
```

Round 1 findings:

```text
F1 MATERIAL  H1 mutable-status cleanup incomplete
F2 MINOR     whitespace demotion also demoted git conflict-marker detection
F3 MINOR     recurrence guard for sole-roadmap status law remains narrow
```

Lead adjudication:

```text
F1 ACCEPTED
F2 ACCEPTED
F3 ACKNOWLEDGED; broad durable-doc regex rejected as brittle/low-signal
```

No Product/T1→T7/T8 semantic authority was reopened.

## Corrected delta under review

Relative to the exact Round-1-reviewed candidate `96c99fa6...`, the corrected candidate changes only:

```text
docs/product/alignment.md
docs/architecture/technical-baseline.md
docs/architecture/backend.md
docs/architecture/interfaces.md
docs/architecture/persistence.md
.github/workflows/ci.yml
docs/roadmap.md
docs/work/current/t8h-global-coherence.md
```

No wire-contract or operation-census file changed in the Round-1 correction delta.

### F1 correction intent

`docs/product/alignment.md`

```text
remove stale T1 ACTIVE / Current gate / active staging packet / stale NEXT
preserve immutable A1-A10 + 4+1 + T1→T7 approval facts
route current program stage / implementation permission / next action only to roadmap
```

`docs/architecture/technical-baseline.md`

```text
preserve immutable T8-A closure
frame T8-B as historical handoff, not current next stage
route mutable program state only to roadmap
```

`docs/architecture/backend.md`

```text
preserve immutable T8-B closure
frame T8-C as historical handoff, not current next stage
remove live current implementation assertion
```

`docs/architecture/interfaces.md`

```text
preserve immutable T8-C closure
make T8-D/T8-E sections authority partition rather than current progression
frame T8-D as historical handoff
route mutable program state only to roadmap
```

`docs/architecture/persistence.md`

```text
remove direct stale T8-E ACTIVE / T8-F→T8-H NOT OPEN contradiction
preserve immutable T8-D closure
frame T8-E as historical handoff
route mutable program state only to roadmap
```

Documentation governance remains relevant:

```text
docs/development/documentation.md
```

It explicitly makes `docs/roadmap.md` the only mutable stage/gate/implementation-status/next-action authority and classifies imported pre-reset provenance strings as non-navigational history. Do not demand cosmetic recursive normalization unless a surviving string still acts as live current-program authority.

### F2 correction intent

The candidate continues to use `git diff --check` as hygiene evidence, but now classifies its diagnostics:

```text
leftover conflict marker -> BLOCKING ERROR
other diff-check output  -> WARNING
```

The Lead separately executed a deterministic Git probe showing:

```text
<<<<<<< / ======= / >>>>>>>
  -> git diff --check: leftover conflict marker

trailing spaces
  -> git diff --check: trailing whitespace
```

Do not trust the Lead claim; inspect the workflow logic and falsify it if there is a real bypass.

### F3 adjudication under attack

The Lead did **not** add a blanket grep over all durable Product/Architecture/Decision prose for words like ACTIVE/BLOCKED/NEXT.

Reason:

```text
such strings may be immutable ratification state,
historical handoff/provenance,
or actual mutable program state;
a broad regex cannot reliably distinguish them and would recreate low-signal red CI
```

The chosen control model is:

```text
roadmap = sole mutable status authority by explicit documentation law
root routers are mechanically prevented from becoming parallel roadmaps
material durable-stage contradictions are reviewable defects
historical/provenance text is permitted when clearly frozen/non-authoritative
```

Your job is to attack whether that is sufficient for this repository **without inventing a generic documentation linter by preference**.

## Required bounded attacks

### R2-1 — F1 closure

Try to find a **surviving live current-program assertion** outside `docs/roadmap.md` in the implicated Whole-Product/T8-A→T8-G authority chain that:

- says a stage is currently ACTIVE / NOT OPEN / NEXT when that is no longer true;
- claims current implementation permission/state as its own authority;
- points to an active staging packet/current next action that is no longer current;
- acts as a competing current router rather than frozen ratification/handoff/provenance.

Distinguish carefully between:

```text
immutable status: CLOSED / OPERATOR-RATIFIED / PROMOTED
historical handoff: at ratification the downstream consumer was X
imported source provenance: old wiki/path/branch labels retained as history
current program status: what is active/open/blocked/next now
```

Only the last class violates the sole-roadmap law.

If you find a survivor, identify exact file/section/claim and explain why existing provenance/handoff framing does not neutralize it.

### R2-2 — F2 conflict-marker correctness

Attack the exact workflow behavior:

- whitespace must remain non-blocking;
- a `git diff --check` `leftover conflict marker` finding must cause `required` to fail;
- the filtering must not accidentally swallow a conflict marker because of shell behavior;
- no second custom conflict grammar should be required if Git's diagnostic is sufficient.

A concrete bypass is MATERIAL only if it lets a merge-conflict marker survive the required candidate gate.

### R2-3 — F3 no-broad-regex adjudication

Try to falsify the Lead's decision not to add a broad durable-doc regex.

A valid MATERIAL finding must demonstrate a **specific low-noise mechanical property** that the current controls fail to protect and that can be enforced without conflating immutable/historical text with current mutable status.

Do not require a generic text-policing framework merely because a human review found F1.

If you believe a bounded guard is justified, specify:

```text
exact protected property
exact files/surface it should cover
exact low-false-positive rule
negative example that must fail
legitimate historical/ratification example that must pass
```

Otherwise explicitly uphold the no-broad-regex adjudication.

### R2-4 — regression / census

Verify the Round-1 correction did not change Product/T8 semantics:

```text
application operation census = exactly 78
operation 79 = absent
new Permission = none
new semantic owner = none
new persistence authority = none
new runtime capability = none
H2 wire-contract meaning unchanged
H3 application/maintenance meaning unchanged
T9 remains NOT OPEN
Product implementation remains BLOCKED
```

The Round-1 correction is intended to be status/routing/CI precision only.

## Finding standard

Classify proportionally.

### MATERIAL

A real surviving contradiction, unsafe gate bypass, duplicate current authority, semantic regression, census change, or other property that prevents T8-H convergence.

### MINOR

A bounded clarity/proof issue that does not change protected architecture or make the candidate unsafe to ratify.

### NO FINDING

Preference, wording taste, cosmetic provenance cleanup, hypothetical future scale, demand for extra ceremony, or a CI rule without a low-noise protected property.

For each finding state:

```text
F#
Severity: MATERIAL | MINOR
Claim
Evidence
Owning authority/property
Why real defect vs preference
Smallest correction
Reopen scope, if any
```

## Verdict

If no MATERIAL finding survives, explicitly write:

```text
CONVERGED
MATERIAL findings = 0
Round 3 = NOT JUSTIFIED
```

If one or more MATERIAL findings survive, explicitly write:

```text
NOT CONVERGED
```

and identify only the smallest owning authority/scope that must reopen.

## Write / push contract

Your ONLY writable review artifact is:

```text
docs/work/current/ai-dialog.md
```

When the review is complete:

1. Write the completed Round-2 review into this file.
2. Verify this remains the **only delta** relative to `arch/t8h-global-coherence @ b940d4e105a8b837ecdac7f71233ff10d735cd5e`.
3. Commit the completed review on `review/t8h-fable-r2`.
4. **PUSH the final commit to `origin/review/t8h-fable-r2`.**
5. Do not stop after editing locally. The Lead reads the pushed commit from the Evidence PR.
6. Do not merge the Evidence PR.
7. Do not edit `arch/t8h-global-coherence` or `main`.
8. Do not begin T9 or implement Product code.

After push, report:

```text
final review commit SHA
confirmation pushed to origin/review/t8h-fable-r2
verdict
MATERIAL finding count
whether Round 3 is justified
```
