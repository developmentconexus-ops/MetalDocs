---
id: development-documentation
kind: authority
status: current
owner: engineering
summary: Defines MetalDocs documentation placement, naming, lifecycle, agent routing, review, and pull-request governance.
---

# Documentation and agent-context governance

> This page becomes authority only after PR #132 is independently reviewed, operator-ratified, and merged.

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

Method outcome:

```text
RESTRUCTURE NOW
```

Updating isolated indexes would preserve the defect class: current truth, active work, and historical archive would still compete in the same tree.

## Documentation root

First-party maintained documentation lives under:

```text
docs/
```

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

Third-party-managed documentation may remain under `vendor/` or `third_party/` and is not MetalDocs authority.

The live tree does not retain:

```text
wiki/
docs/superpowers/
docs/operator/
archive/legacy/tombstone documentation trees
completed candidates/reviews/adjudications
```

Git history and closed pull requests preserve provenance.

## Naming and metadata

Durable filenames describe stable subjects.

Use:

```text
lowercase
kebab-case
short semantic nouns or noun phrases
```

Examples:

```text
product/contract.md
architecture/persistence.md
operations/runbooks/restore.md
```

Durable filenames do not contain:

```text
dates
R10/T-stage codes
v1/v2/revision numbers
final/new/old
candidate/corrected
review/adjudication
amendment/tombstone
legacy/historical
```

Explicit exceptions are limited to numbered ADRs, dated incidents, and release notes where the number/date is part of the document identity.

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

A replacement repairs links and deletes the previous page. It does not keep a compatibility stub unless a real external consumer or unchangeable tool requires the old path.

## Authority and navigation

One meaning has one current authority.

`docs/index.md` maps reader intent to the owning page. It is not a second summary of those pages.

`mkdocs.yml` defines publishable navigation. It includes every durable maintained page exactly once and excludes `docs/work/`.

Current status lives only in:

```text
docs/status.md
```

Indexes, README, AGENTS, PR bodies, and active-work files point to that status instead of copying it.

The compact decision registry lives in:

```text
docs/decisions/index.md
```

It records decision ID, status, authority link, provenance, and reopen trigger. It does not duplicate the full decision prose.

## Agent context

Root `AGENTS.md` is a bounded bootstrap containing only:

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

It does not copy architecture, stage summaries, decision ledgers, review history, or task prompts.

Routine orientation is:

```text
AGENTS.md
→ docs/index.md
→ status/work pointer only when relevant
→ 1–3 task-specific authorities
→ current code evidence only for concrete claims
```

Nested `AGENTS.md` files require repeated evidence of genuinely path-specific needs. `CLAUDE.md` remains a minimal pointer to root `AGENTS.md` and owns no independent meaning.

Material engineering decisions apply:

```text
developmentconexus-ops/conexus-methodology/METHOD.md
```

Repository/product authority remains local. External practices are evidence, not product requirements.

## Active work and AI dialogue

A governed architecture/governance PR may maintain:

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
- `ai-dialog.md` is created only when the proposal is ready for final independent review;
- Fable review, Lead adjudication, bounded Round 2, and operator decision remain in that one file;
- Round 2 occurs only when a material contradiction survives;
- all temporary work files are deleted before the PR becomes merge-ready.

Fable chat handoff stays short:

```text
repository / branch / PR / expected HEAD
read AGENTS.md
review docs/work/current/proposal.md
write only docs/work/current/ai-dialog.md
```

The canonical Fable workflow is referenced from `developmentconexus-ops/conexus-methodology/README.md`; it is not copied into a repository-local skill.

## Pull-request lifecycle

One coherent, ratifiable gate uses:

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

A stage may be split only into independently coherent, ratifiable, merge-safe gates—not by conversation or review round.

PR body and branch names may carry stage/provenance identifiers. Durable filenames do not.

## Deletion and retention

A document survives only when all are true:

1. A named current consumer exists.
2. The document has unique current meaning.
3. Its authority class is explicit.
4. It is reachable from navigation.
5. Its owner is named.
6. Its lifecycle or deletion trigger is clear.

Otherwise:

```text
DELETE FROM LIVE TREE
```

Current-runtime references or runbooks may survive while the running implementation still consumes them. They must state that current consumer and deletion trigger; they do not become target architecture authority.

No local archive is created for deleted material.

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

A documentation consolidation is complete only when:

1. every accepted authority is preserved;
2. every retained operational page has a named current consumer;
3. all live links/tool consumers are repaired;
4. temporary work files are absent;
5. the hygiene guard is proven to fire;
6. required CI is green;
7. superseded PRs can be closed without losing current truth.

## Reopen triggers

Reopen this profile only on material evidence that:

- a real external consumer requires historical documentation URLs;
- a required tool cannot operate with one documentation root;
- safe parallel gates cannot use the proposed work lifecycle;
- deleting a current-state page makes the running system unsafe to operate;
- another Conexus product cannot use the common structure without semantic distortion;
- measured agent routing remains materially ambiguous or expensive.

Preference for the former tree or hypothetical future compatibility is not a reopen trigger.

## Related documents

- `docs/work/current/proposal.md` — temporary design source during PR #132.
- `docs/work/current/plan.md` — temporary execution plan during PR #132.
- `developmentconexus-ops/conexus-methodology/METHOD.md` — organizational engineering method.
- `developmentconexus-ops/conexus-methodology/README.md` — canonical Fable review workflow.
