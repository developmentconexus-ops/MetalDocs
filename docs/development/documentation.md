---
id: development-documentation
kind: authority
status: active
owner: engineering
summary: Defines MetalDocs documentation placement, naming, lifecycle, agent routing, review, and pull-request governance.
---

# Documentation and agent-context governance

> Candidate authority. Promote `status` to `current` only after independent review, operator ratification, and merge.

## Purpose

Keep the live repository small, navigable, and unambiguous for humans and agents without losing accepted product, architecture, operational, or verification truth.

## Current decision

```text
ONE docs/ ROOT
+ SEMANTIC STABLE FILENAMES
+ TASK-ORIENTED INDEX
+ EXPLICIT NAVIGATION
+ SHORT AGENT BOOTSTRAP
+ ONE TEMPORARY PROPOSAL
+ ONE TEMPORARY AI DIALOGUE
+ ONE PR PER RATIFIABLE GATE
+ GIT AS THE ONLY ARCHIVE
+ ALLOWLIST-BASED LEGACY DELETION
- wiki/
- docs/superpowers/
- stage/date/version filenames
- amendment chains
- permanent per-round review artifacts
- live-tree archives
- duplicated status and authority
```

Method outcome: **RESTRUCTURE NOW**. Updating isolated indexes would preserve the defect class: current truth, active work, and historical archive would still compete in one tree.

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

The final live tree does not retain:

```text
wiki/
docs/superpowers/
docs/operator/
archive/legacy/tombstone documentation trees
completed candidates/reviews/adjudications
```

Git history and closed pull requests preserve provenance.

## Naming and metadata

Durable filenames use lowercase kebab-case and describe stable subjects:

```text
product/contract.md
architecture/persistence.md
operations/runbooks/restore.md
```

They do not carry dates, R10/T-stage codes, versions, `final/new/old`, `candidate/corrected`, `review/adjudication`, `amendment/tombstone`, or `legacy/historical`. Explicit exceptions are limited to numbered ADRs, dated incidents, and release notes where the number or date is part of the document identity.

Every maintained Markdown page under `docs/` has unique frontmatter:

```yaml
---
id: architecture-persistence
kind: authority
status: current
owner: architecture
summary: Defines the target relational model, constraints, and concurrency rules.
---
```

Closed values:

```text
kind: authority | guide | reference | work
status: current | active
```

A replacement repairs links and deletes the previous page. A compatibility stub requires a real external consumer or unchangeable tool.

## Authority and navigation

One meaning has one current authority.

- `docs/index.md` maps reader intent to the owning page; it does not repeat decisions.
- `mkdocs.yml` lists every publishable durable page exactly once and excludes `docs/work/`.
- `docs/status.md` is the sole current stage and implementation-gate authority.
- `docs/decisions/index.md` records compact decision identity, status, authority link, provenance, and reopen trigger without duplicating decision prose.
- README, AGENTS, indexes, PR bodies, and work files point to authorities instead of copying them.

## Agent context

Root `AGENTS.md` contains only:

```text
repository purpose
docs/index.md entrypoint
docs/status.md pointer when stage matters
docs/work/current/index.md pointer when active work exists
task-specific authority lookup
build/test/verify commands
stable Git/security rules
canonical Method and Fable pointers
```

It does not copy architecture, stage summaries, decision ledgers, review history, or prompts.

Routine orientation:

```text
AGENTS.md
→ docs/index.md
→ status/work pointer only when relevant
→ 1–3 task-specific authorities
→ current code evidence only for concrete claims
```

Nested `AGENTS.md` files require repeated evidence of path-specific needs. `CLAUDE.md` remains a minimal pointer to root `AGENTS.md` and owns no independent meaning.

Material engineering decisions apply `developmentconexus-ops/conexus-methodology/METHOD.md`. Repository and product authority remain local; external practices are evidence, not product requirements.

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
- `ai-dialog.md` appears only for the final independent review;
- Fable review, Lead adjudication, bounded Round 2, and operator decision remain in that one file;
- Round 2 occurs only when a material contradiction survives;
- all temporary work files are deleted before the PR becomes merge-ready.

Fable handoff contains only repository, branch, PR, expected HEAD, and pointers to `AGENTS.md`, `proposal.md`, and `ai-dialog.md`. The canonical workflow lives in `developmentconexus-ops/conexus-methodology/README.md`; it is not copied into a local skill.

## Pull-request lifecycle

One coherent ratifiable gate uses:

```text
one branch
one Draft PR
one active proposal
one final independent review
one operator decision
one squash merge
```

Flow:

```text
branch from current main
→ Draft PR
→ proposal converges
→ final Fable challenge
→ Lead adjudication
→ operator ratification
→ promote durable docs
→ delete temporary work files
→ required checks green
→ squash merge
→ next gate starts from updated main
```

A stage may be split only into independently coherent, ratifiable, merge-safe gates—not by conversation or review round. Stage/provenance identifiers belong in PR metadata and decision provenance, not durable filenames.

## Deletion and retention

A document survives only when all are true:

1. A named current consumer exists.
2. It has unique current meaning.
3. Its authority class is explicit.
4. It is reachable from navigation.
5. Its owner is named.
6. Its lifecycle or deletion trigger is clear.

Otherwise: **delete it from the live tree**.

Current-runtime references or runbooks may survive while the running implementation still consumes them. They must state that consumer and deletion trigger; they do not become target architecture authority. No local archive is created.

## Mechanical proof obligations

Repository verification must fail on:

```text
wiki/ or docs/superpowers/ in the final tree
unapproved first-party Markdown roots
archive/legacy/tombstone directories
prohibited durable filenames
missing or invalid frontmatter
duplicate document IDs
broken internal links
orphan durable pages
pages absent from MkDocs navigation
work pages included in published navigation
temporary work files in a merge-ready PR
```

Every repository-authored blocking guard includes a negative fixture or an accepted classified waiver under the existing verifier rules.

A consolidation is complete only when:

1. every accepted authority is preserved;
2. each retained operational page has a named current consumer;
3. all live links and tool consumers are repaired;
4. temporary work files are absent;
5. the hygiene guard is proven to fire;
6. required CI is green;
7. superseded PRs can be closed without losing current truth.

## Reopen triggers

Reopen only on material evidence that:

- a real external consumer requires historical documentation URLs;
- a required tool cannot operate with one documentation root;
- safe parallel gates cannot use the work lifecycle;
- deleting a current-state page makes the running system unsafe to operate;
- another Conexus product cannot use the common structure without semantic distortion;
- measured agent routing remains materially ambiguous or expensive.

Preference for the former tree or hypothetical compatibility is not a reopen trigger.

## Related documents

- `docs/work/current/proposal.md` — temporary design source during PR #132.
- `docs/work/current/plan.md` — temporary execution plan during PR #132.
- `developmentconexus-ops/conexus-methodology/METHOD.md` — organizational engineering method.
- `developmentconexus-ops/conexus-methodology/README.md` — canonical Fable review workflow.
