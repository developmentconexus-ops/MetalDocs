---
id: development-documentation
kind: authority
owner: engineering
summary: Defines MetalDocs documentation placement, naming, lifecycle, agent routing, review, and pull-request governance.
---

# Documentation and agent-context governance

> Candidate authority. It becomes current only after independent review, operator ratification, temporary-work cleanup, required checks, and merge of PR #132.

## Purpose

Keep the live repository small, navigable, and unambiguous for humans and agents without losing accepted product, architecture, operational, verification, or provenance truth.

## Current decision

```text
ONE docs/ ROOT
+ SEMANTIC STABLE FILENAMES
+ TASK-ORIENTED INDEX
+ EXPLICIT NAVIGATION MANIFEST
+ SHORT AGENT BOOTSTRAP
+ ONE TEMPORARY PROPOSAL
+ ONE TEMPORARY AI DIALOGUE
+ ONE PR PER RATIFIABLE GATE
+ GIT AS THE ARCHIVE
+ COMPLETE DISPOSITION BEFORE DELETION
+ GATE-SUBJECT PRESERVATION
- wiki/
- docs/superpowers/
- stage/date/version filenames
- amendment chains
- permanent per-round review artifacts
- live-tree archives
- duplicated status and authority
```

Method outcome: **RESTRUCTURE NOW**. Repairing isolated indexes would preserve the defect class: current truth, active work, and historical archive would still compete in the same live tree.

## Documentation root

First-party maintained documentation lives under `docs/`.

Platform-conventional files may remain at repository root:

```text
README.md
AGENTS.md
CLAUDE.md
LICENSE
SECURITY.md
CONTRIBUTING.md
mkdocs.yml
.github/*
```

Third-party-managed documentation may remain under `vendor/` or `third_party/`; it is not MetalDocs authority.

The final live tree does not retain `wiki/`, `docs/superpowers/`, `docs/operator/`, repository-local archive/tombstone trees, or completed candidate/review/adjudication artifacts. Deletion happens only after every current consumer and verification subject has been explicitly disposed.

Git history and closed pull requests preserve historical process and deleted content. History-pinned security allowlists may continue to reference deleted historical paths when the security control still consumes those fingerprints.

## Naming and metadata

Durable filenames use lowercase kebab-case and describe stable subjects:

```text
product/contract.md
architecture/persistence.md
operations/runbooks/restore.md
reference/problem-codes.md
```

They do not carry dates, R10/T-stage codes, versions, `final/new/old`, `candidate/corrected`, `review/adjudication`, `amendment/tombstone`, or `legacy/historical`. Exceptions are limited to identifiers whose number/date is part of the durable subject, including numbered ADRs, dated incidents, and release notes.

Every maintained Markdown page under `docs/` has unique frontmatter:

```yaml
---
id: architecture-persistence
kind: authority
owner: architecture
summary: Defines the target relational model, constraints, and concurrency rules.
---
```

Closed `kind` values are:

```text
authority
work
```

`authority` means a maintained durable page whose subject is current for its declared purpose. `work` means temporary non-authoritative material inside a Draft PR. Ratification is represented by promotion into the durable tree and decision provenance, not by a second status field.

Generated durable pages use the same frontmatter, emitted by their generator. A generated page is never hand-edited to satisfy metadata rules.

A replacement repairs live consumers and deletes the previous page. A compatibility stub requires a demonstrated external consumer or tool that cannot be repaired in the same gate.

## Authority and navigation

One meaning has one current authority.

- `docs/index.md` maps reader intent to the owning page; it does not repeat decisions.
- `docs/status.md` is the sole current stage and implementation-gate authority.
- `docs/decisions/index.md` records compact decision identity, status, authority link, provenance, and reopen trigger without duplicating decision prose.
- `mkdocs.yml` is the explicit durable navigation manifest. Publishing is not required by this profile; build-oriented MkDocs keys are compatibility for future publication, not current product obligations.
- `docs/work/` is never published.

Large governed collections may be reachable through one indexed collection page instead of listing every member directly in `mkdocs.yml`. This applies to ADRs, generated references, database table dictionaries, and similar bounded collections. Every durable page must remain reachable from the navigation graph.

Machine-consumed documentation has explicit homes, including:

```text
docs/decisions/adr-NNNN-<slug>.md
docs/reference/database/<subject>.md
docs/reference/problem-codes.md
docs/reference/requirement-traceability.md
```

Module-debt or equivalent registers survive only if their current gate survives; otherwise the gate and its subject are retired together.

## Agent context

Root `AGENTS.md` is a bounded bootstrap. It contains repository purpose, navigation entrypoints, build/test/verify commands, stable Git/security rules, and pointers to durable repository engineering law and current-system orientation.

Durable homes are:

```text
docs/development/engineering-rules.md
  repository-specific prerequisite stops, mismatch rules, safety and Git constraints

docs/reference/current-system.md
  current implementation orientation needed to work safely before target transition
```

Routine orientation is:

```text
AGENTS.md
→ docs/index.md
→ status/work pointer only when relevant
→ 1–3 task-specific authorities
→ current code evidence only for concrete claims
```

`AGENTS.md` does not copy architecture, stage summaries, decision ledgers, review history, or task prompts. Nested `AGENTS.md` files require repeated evidence of genuinely path-specific needs. `CLAUDE.md` remains a minimal pointer to root `AGENTS.md` and owns no independent product or architecture meaning.

Material engineering decisions apply `developmentconexus-ops/conexus-methodology/METHOD.md`. The canonical Fable workflow is referenced from the organization methodology rather than copied into a local skill. Each Conexus repository owns its own product and architecture truth and implements enforcement on its own verification spine.

## Active work and AI dialogue

A governed architecture or governance PR may maintain:

```text
docs/work/current/index.md
docs/work/current/proposal.md
docs/work/current/plan.md
docs/work/current/ai-dialog.md
```

Rules:

- one active proposal per coherent gate;
- proposal and plan are edited in place;
- no permanent corrected-candidate, review-request, delta-review, adjudication, or tombstone files;
- `ai-dialog.md` appears only for final independent review;
- Fable review, Lead adjudication, bounded Round 2, and operator decision stay in that one file;
- Round 2 occurs only when a material contradiction survives;
- `docs/work/**` may exist only while the pull request is Draft;
- all temporary work files are deleted before the PR is marked ready for review.

CI must run the work-file guard on `ready_for_review` as well as ordinary PR synchronization, so changing PR state cannot bypass the rule.

## Pull-request lifecycle

One coherent ratifiable gate uses one branch, one Draft PR, one active proposal, one final independent review, one operator decision, and one squash merge.

Flow:

```text
branch from current main
→ Draft PR
→ proposal converges
→ final Fable challenge
→ Lead adjudication
→ operator ratification
→ promote durable docs and provenance
→ delete temporary work files
→ required checks green
→ mark ready
→ squash merge
→ next gate starts from updated main
```

A stage may be split only into independently coherent, ratifiable, merge-safe gates—not by conversation or review round. A pushed shared branch is not rebased/force-pushed merely to refresh its base; merge the updated base or use a new clean branch according to repository Git policy.

The PR that promotes a durable authority also records its decision provenance. Until `docs/decisions/index.md` exists, the durable page itself records the promoting PR/review/operator decision; G1 moves that provenance into the compact registry without deleting it from Git history.

## Deletion and retention

Deletion is driven by a complete census, not a hand-written root allowlist.

Before removing documentation, enumerate the tracked documentation estate and give every path exactly one disposition:

```text
KEEP → <target path>
MERGE → <target authority>
GENERATED → <generator + target path>
DELETE → <reason and proof no current consumer remains>
```

An undispositioned path is a stop condition.

A durable document survives only when all are true:

1. A named current consumer exists.
2. It has unique current meaning.
3. Its authority role is explicit.
4. It is reachable from navigation or an indexed governed collection.
5. Its owner is named.
6. Its lifecycle or deletion trigger is clear.

Items 1, 2, and 6 require review-time engineering judgment; items 3–5 are mechanically checkable.

Current-runtime references and runbooks may survive while the running implementation still consumes them. They must name that consumer and deletion trigger; they do not become target architecture authority.

### Gate-subject invariant

No verification gate's declared documentation subject may be deleted unless that gate is repointed or retired in the same pull request and its relevant negative proof is re-run.

This includes PR and non-PR/nightly governance gates. A documentation rebaseline must inspect both the documentation tree and the verification registry/workflows that consume it.

### Consumer classification

Path repair applies to **live executable consumers**, including scripts, workflow/configuration, generators, verification tools, navigation and maintained links that resolve a documentation path at runtime or verification time.

A historical/provenance citation does not require churn merely because its referenced path leaves the live tree. Examples include source comments, applied migrations, generated historical artifacts, commit-pinned security fingerprints, and other text whose purpose is provenance rather than path resolution.

The consolidation must classify ambiguous hits rather than requiring zero textual occurrences of old paths.

## Consolidation proof

Product/R10 consolidation uses both structural mapping and normative-content checks.

Required proof:

1. A closed source manifest from the immutable PR #131 authority HEAD maps every accepted source authority to one or more target authorities.
2. The source census is read with a filesystem-capable command (`grep`, `rg`, or equivalent), not `git grep` against ignored temporary files.
3. The normative vocabulary census covers the repository's actual terms, including at least `MUST`, `SHOULD`, `MAY`, `SHALL`, `REQUIRE`, `FORBID`, `PROHIBIT`, `NEVER`, `ALWAYS`, `ONLY`, `SELECT`, `REJECT`, `BLOCKED`, and `CLOSED` where used normatively.
4. An empty source census is failure, never parity.
5. Any semantic contradiction found during consolidation stops the Writer; it is not silently reconciled.

The grep census is supporting evidence, not a substitute for the closed source-to-target decision map and human/agent semantic parity review.

## Mechanical proof obligations

Repository verification must fail on:

```text
wiki/ or docs/superpowers/ in the final tree
unapproved first-party Markdown roots
archive/legacy/tombstone documentation trees
prohibited durable filenames
missing or invalid frontmatter
duplicate document IDs
broken maintained internal links
orphan durable pages or collections
work pages included in durable navigation
docs/work/** when the PR is not Draft
```

Generated pages must prove metadata through their generator. ADR/reference collections may use a governed collection index rather than one top-level nav entry per member.

Every repository-authored blocking guard has a negative fixture or accepted classified waiver. Fixtures include the minimum valid repository/navigation context needed to prove the intended failure rather than merely exiting non-zero for unrelated setup errors.

Local proof mirrors the actual CI ladder: verification commands use `--require-infra` where CI does, lint is run when Go verifier code changes, and non-PR/nightly governance checks affected by the documentation move are executed explicitly before merge.

A consolidation is complete only when:

1. every accepted authority is preserved;
2. every tracked documentation path has an explicit disposition;
3. every retained operational page has a named current consumer;
4. every live executable consumer is repaired or retired;
5. every affected verification-gate subject is rehomed or its gate is retired with proof;
6. temporary work files are absent;
7. hygiene negative fixtures demonstrate the intended rules fire;
8. required PR, lint, security, and affected governance checks are green;
9. superseded PRs can be closed without losing current truth.

## Transitional state

PR #132 defines this profile but does not yet reorganize the existing tree. Until G1 merges, existing `AGENTS.md`, `wiki/`, current verification subjects, and current runtime documentation remain operative evidence/safety rails. G1 is the named successor gate that atomically rehomes current truth, repairs consumers, activates the new navigation/verifier, and removes legacy roots.

No Writer may apply the final-root deletion rules before the G1 parity and gate-subject proofs are satisfied.

## Reopen triggers

Reopen only on material evidence that:

- a real external consumer requires historical documentation URLs;
- a required tool cannot operate with one documentation root;
- safe parallel gates cannot use the work lifecycle;
- deleting a current-state page makes the running system unsafe to operate;
- another Conexus product cannot use the common structure without semantic distortion;
- measured agent routing remains materially ambiguous or expensive.

Preference for the former tree or hypothetical compatibility is not a reopen trigger.

## Provenance

- Gate: repository documentation profile / G0.
- Pull request: #132.
- Independent review: Fable review preserved at PR #132 commit `3b8a25488e1aed5edc6c2b83d64e802b8d66c1c0`.
- Lead adjudication: preserved in subsequent PR #132 history and temporary `docs/work/current/ai-dialog.md` until merge cleanup.
- Operator ratification: pending.
- Consolidation successor: G1 Product/R10 documentation rebaseline.
- Canonical engineering method: `developmentconexus-ops/conexus-methodology/METHOD.md`.
