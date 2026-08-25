---
id: repository-operating-method
kind: methodology
owner: development
summary: Local repository operating method for authority recovery, selective context, documentation, Evidence, Git and acceptance-increment continuity.
---

# DevelopmentConexus Repository Operating Method

**Version:** 1.0.0  
**Status:** CANDIDATE — MetalDocs governance rebaseline  
**Scope:** repository operation and continuity; not Product or architecture meaning

## Objective

Keep the repository itself usable as the durable working memory of the Product:

```text
fresh actor
→ recover current state without chat archaeology
→ find the smallest relevant authority pack
→ reason with the applicable method
→ expand only when Evidence can change the conclusion
→ leave one coherent, reviewable acceptance increment
→ preserve required provenance without turning history into live context
```

Optimize for **decision signal per token** and continuity across humans/agents. Context efficiency is a means to correctness, not a correctness ceiling.

This method governs repository operation. Engineering reasoning remains owned by `engineering-method.md`; frontend Product Experience planning remains owned by `frontend-product-experience-planning-method.md`; Product/architecture semantics remain owned by their repository authorities.

## 1. Authority hierarchy

Current accepted repository authority outranks chat, handoff, memory, old PR descriptions, historical snapshots and reviewer preference.

Repository roles:

```text
AGENTS.md
  bootstrap + method/router selection + local hard stops

docs/roadmap.md
  sole mutable current stage/block/gate/next-action authority

docs/index.md
  task/intention router to the smallest useful authority pack

docs/decisions/index.md
  compact registry of current material decisions/dispositions

docs/product/**
  Product meaning

docs/architecture/**
  structural/system meaning

docs/decisions/**
  bounded accepted decisions/refinements/reopen triggers

docs/development/**
  methods + repository-local engineering specialization

docs/work/**
  temporary branch-only work; never durable authority

Git / PR / Evidence / research / tests
  evidence and provenance; not Product/status authority merely by existing
```

One material meaning should have one current owner. Routers, indexes, summaries and Evidence locators point to meaning; they do not restate it into parallel authority.

## 2. Fresh-session recovery

Before relying on prior conversation state, establish as applicable:

```text
repository identity
current branch + HEAD
remote main HEAD
relevant PR base/head/state
required aggregate check
unowned worktree state when local tooling exists
```

Then:

```text
AGENTS.md
→ docs/roadmap.md
→ applicable local method(s)
→ docs/index.md
→ 1–2 task-owning authorities by default
→ targeted sections/operations
```

`docs/index.md` may be consulted before the methods when routing itself is the immediate task; no semantic conclusion depends on that ordering.

Chat/handoffs are routing convenience only. A material active state that cannot be recovered from lawful repository/PR/Evidence state is a continuity defect.

## 3. Selective-context law

### 3.1 Start small

Default repository task context should normally be:

```text
bootstrap/status
+ applicable method(s)
+ 1–2 semantic owners
```

Do not recursively read `docs/`, Git history, closed PRs, Evidence refs, research or removed implementation by default.

### 3.2 Large-owner retrieval

A large owner file does not need to be loaded in full merely because it is authoritative.

Prefer:

```text
identify owner
→ search exact concept / operationId / invariant / heading
→ read the relevant section plus required global laws
→ expand inside that owner only when whole-owner coherence can change the claim
```

Examples of reasons to widen within an owner:

- a global law governs the local operation;
- the same concept is defined in multiple sections;
- a local clause depends on a shared invariant;
- an adversarial counterexample can only be resolved globally.

### 3.3 Expand on materiality, not ceremony

There is no hard file/token count that may override correctness.

Expand when another repository owner, history item, Evidence source, runtime fact or external source can **materially change, challenge or falsify** the conclusion.

When expansion becomes broad, state the reason. If several files are repeatedly required for one concept, investigate whether routing or ownership is fragmented rather than normalizing whole-repo reading.

### 3.4 Negative guidance is useful

`docs/index.md` should tell the actor not only where to start, but what is normally irrelevant:

```text
START WITH
ADD WHEN
DO NOT READ BY DEFAULT
```

Negative guidance is routing, not a prohibition. Material Evidence always permits expansion.

## 4. Documentation model

Create only surfaces with a current consumer.

MetalDocs baseline semantic roots are:

```text
docs/
├── index.md
├── roadmap.md
├── product/
├── architecture/
├── decisions/
├── development/
└── work/          temporary branch-only when needed
```

Do not add archive/history/session/round trees to the live documentation merely to preserve chronology. Git is the normal archive.

Durable filenames use semantic lowercase kebab-case. Historical provenance strings inside imported ratified authorities may remain when changing them would add risk without improving routing.

Minimal frontmatter is useful when it makes ownership/routing machine- or human-readable:

```yaml
id: stable-id
kind: authority | methodology | evidence-locator | work
owner: owner-name
summary: one sentence
```

## 5. Roadmap law — snapshot, not journal

`docs/roadmap.md` owns mutable current program truth and should remain compact.

It should answer:

```text
where are we?
what integrated checkpoints matter now?
what block/stage is open?
what materially blocks progress?
is implementation allowed?
what is the exact next action?
```

It should **not** accumulate:

```text
round-by-round review chronology
wireframe R1→Rn narrative
full finding analyses
Evidence blob history
conversation handoffs
closed implementation detail
```

Those belong to the owning decision, temporary work, PR/Git history or a short Evidence locator when exact recovery is still required.

A stage may remain OPEN across multiple merged acceptance increments. Merge does not imply stage closure.

## 6. Index and decision-register law

### `docs/index.md`

Task/intention router only. Prefer one row per real recurring task family with:

```text
Task / intention
Start with
Add when
Do not read by default
```

The index may name a large owner while advising targeted section/operation retrieval first.

### `docs/decisions/index.md`

Current decision/disposition registry only. Keep entries compact:

```text
ID / subject
disposition
one-line current outcome
owning authority
```

Do not reconstruct review chronology in the register. The owning decision file contains the necessary basis/reopen triggers.

## 7. Temporary work

Use `docs/work/current/` only for temporary branch work that has a real current consumer, such as:

```text
functional P8 HTML
candidate analysis
short execution plan
review-local proof notes
```

Temporary work is non-authoritative and may change rapidly.

Before a PR becomes a merge candidate:

```text
absorb current semantics into durable owners/locators
+ preserve exact byte Evidence only when still required
+ remove docs/work/**
```

Do not preserve temporary work permanently because it was expensive to create.

## 8. Evidence and provenance

Evidence supports or falsifies authority; it does not silently become authority.

Preserve an exact durable ref/blob only when a current consumer requires byte-level reconstruction or unmerged provenance.

Prefer:

```text
one final canonical Evidence checkpoint per accepted block/increment
```

rather than a permanent ref for every intermediate iteration.

Earlier candidates normally remain recoverable through branch/PR/Git history. Create additional durable refs only when an independently named consumer requires them.

An Evidence locator should stay short and own only:

```text
what exact artifact is current
ref / commit / blob identity
what protected behavior it proves
how to retrieve/use it
what reopens or retires it
```

It must not become a second roadmap or a full review diary.

Required refs must remain reachable while current authority names them as required. Verification of a specific ref/blob should be performed when the current claim depends on it; unrelated CI runs need not repeatedly prove historical network state.

## 9. Git is the archive

Merged history is normally sufficient archive for superseded material once surviving current meaning is consolidated.

Before deleting the last ref for important **unmerged** work:

1. consolidate still-current meaning into durable authority;
2. determine whether exact byte lineage is still required;
3. if required, preserve one explicit durable ref/tag and record it;
4. only then remove ordinary branches.

Do not keep branch graveyards as a substitute for deciding what remains current.

## 10. Acceptance increments and PR lifecycle

An acceptance increment is the smallest semantically coherent change that can be accepted or rejected while leaving the repository reconstructable.

It is not defined by LOC, file count, commit count or roadmap-stage boundary.

Default lifecycle:

```text
current main
→ one branch
→ Draft PR
→ execute/iterate the coherent increment
→ operator-only decision gates where required
→ targeted proof
→ one full invariant/adversarial sweep before Ready
→ bounded corrections
→ preserve final exact Evidence if required
→ remove temporary work
→ Ready once the candidate is actually integration-ready
→ required aggregate check
→ explicit merge authorization when required
→ squash merge
→ delete ordinary head branch
→ revalidate main
```

Do not use Ready↔Draft transitions as a normal inner-loop review mechanism. Draft is the workspace; Ready means the candidate is believed integration-ready.

### Split law

Split when a discovered change:

- moves or creates Product/architecture authority;
- has a different operator decision;
- has materially separate proof/review needs;
- can be independently accepted/rejected;
- turns the current PR into a long-lived workspace that is difficult to reconstruct.

Frontend findings may legitimately reopen upstream planning. That is not failure: follow Engineering + Frontend methods, reopen the smallest semantic owner, then boundedly rebaseline affected frontend work.

## 11. Review and proof

Reviewer findings are Evidence. Classify them against current authority before changing anything.

A review finding that exposes a defect against current accepted meaning may be corrected locally. A finding that implies new Product meaning must return to the owning decision.

Run review proportional to materiality. Do not create permanent review-transport documentation in the candidate/main.

A final integration candidate should not depend on the reviewer discovering basic contradictions one by one. Before Ready, run the strongest practical invariant sweep over the complete accepted behavior of the increment.

## 12. Repository-local specialization

Stable reusable reasoning belongs in the three local methods. MetalDocs-specific controls belong in `engineering-rules.md`.

A local specialization should name the real MetalDocs consumer/failure class and should not restate an entire method.

If this Repository Operating Method itself becomes a source of repeated ceremony, context inflation or misclassification, reopen it with the Engineering Method rather than layering another router/standard over it.

## 13. Success test

The repository is healthy when a fresh actor can receive only:

> Read `AGENTS.md` and continue from repository authority.

and then reliably:

```text
recover current stage
select the right method
locate the right Product/architecture owner
read only the necessary sections first
expand when material Evidence demands it
understand what is authority vs Evidence/history
continue the current increment without conversation archaeology
```

If normal work requires all methods, whole-repository reading, old PR archaeology, several handoff documents, or guessing which authority is current, repository operation needs repair.
