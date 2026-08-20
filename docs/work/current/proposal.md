---
id: work-repository-information-architecture
kind: work
owner: architecture
summary: Temporary proposal for the docs-only repository information architecture and legacy-deletion governance.
---

# Repository documentation and agent-context proposal

> Temporary, non-authoritative, and deleted before PR #132 becomes merge-ready.

## Decision

```text
ONE docs/ ROOT
+ semantic stable filenames
+ one intent index
+ one navigation manifest
+ short AGENTS bootstrap
+ one temporary proposal
+ one temporary AI dialogue
+ one coherent ratifiable gate per PR
+ Git as archive
+ complete disposition before deletion
+ verification-subject preservation
- wiki/
- docs/superpowers/
- live archives/tombstones
- stage/date/version filenames
- permanent review-round artifacts
- duplicated current authority
```

The target durable authority is `docs/development/documentation.md`.

## Why restructure

The present repository mixes current truth, active work, and process history across `wiki/`, `docs/`, indexes, handoffs, candidates, reviews, and old routing files. The result is authority ambiguity, agent context bloat, stale routing, oversized PRs, and avoidable Git conflicts.

Repairing a few indexes would preserve that defect class. The live tree must instead contain only current maintained truth and bounded active work; Git and closed PRs retain history.

## Target shape

```text
README.md
AGENTS.md
CLAUDE.md
mkdocs.yml

docs/
  index.md
  status.md
  product/
  architecture/
  decisions/
  development/
  operations/
  reference/
  work/current/        # temporary only while PR is Draft
```

Durable files are named by stable subject, for example:

```text
docs/product/contract.md
docs/architecture/persistence.md
docs/development/engineering-rules.md
docs/reference/current-system.md
docs/reference/problem-codes.md
```

R10/T-stage identifiers remain in PR and decision provenance, not durable paths.

## Minimal metadata

Durable and temporary Markdown use:

```yaml
---
id: architecture-persistence
kind: authority
owner: architecture
summary: Defines target relational persistence and concurrency rules.
---
```

Closed kinds:

```text
authority
work
```

A page is current because it is the live durable authority for its subject. Temporary candidacy is represented by `kind: work` and Draft-PR placement, not by a separate status field.

Generated durable pages receive frontmatter from their generator.

## Navigation

`docs/index.md` maps reader intent to authorities. `docs/status.md` owns stage/status. `mkdocs.yml` is the durable navigation manifest; publishing is out of scope.

Large governed collections such as ADRs, generated references, and database-table pages may be reachable through one indexed collection page rather than one top-level navigation row per member.

## Agent routing

`AGENTS.md` becomes a short bootstrap pointing to:

```text
docs/index.md
docs/status.md when stage matters
docs/work/current/ when active governed work exists
docs/development/engineering-rules.md for repository-specific safety law
docs/reference/current-system.md for current implementation orientation
```

`CLAUDE.md` remains a minimal pointer to `AGENTS.md`.

## Active work and Fable

A governed design PR may use only:

```text
docs/work/current/index.md
docs/work/current/proposal.md
docs/work/current/plan.md
docs/work/current/ai-dialog.md
```

The proposal/plan are edited in place. `ai-dialog.md` holds the final Fable review, Lead adjudication, optional bounded Round 2, and operator decision. No permanent review-request, corrected-candidate, delta-review, adjudication, or tombstone documents are created.

`docs/work/**` may exist only while the PR is Draft and must be deleted before the PR is marked ready.

## Deletion model

Deletion is **not** a root allowlist and is **not** a zero-text-occurrence grep.

G1 must start from a complete tracked-document census. Every path receives exactly one disposition:

```text
KEEP → target path
MERGE → target authority
GENERATED → generator + target path
DELETE → reason + proof no current consumer remains
```

An undispositioned path blocks deletion.

A live executable consumer—script, generator, workflow/config, verifier, maintained link, or navigation path—must be repaired or retired in the same PR. Historical/provenance citations in comments, applied migrations, generated history, or commit-pinned security allowlists need not be rewritten merely because the old live path disappears.

## Verification-subject preservation

Machine-consumed documentation is first-class retained truth until its consumer is moved or retired. G1 must explicitly handle at least:

```text
problem-code registry/docs
requirement-trace sources/report
ADRs
current database table dictionary/ownership docs
module-debt registers if their gate survives
current runbooks and engineering rules
```

No verification gate's declared documentation subject may be deleted without repointing or retiring that gate in the same PR and re-running the relevant negative proof.

## Product/R10 parity

G1 reads accepted Product/R10 authority from immutable PR #131 provenance and must prove:

1. every accepted source authority is present in a closed source-to-target map;
2. normative source census is non-empty and uses filesystem-capable search over temporary extracted files;
3. the census includes the repository's actual normative vocabulary (`MUST`, `SHOULD`, `MAY`, `SHALL`, `REQUIRE`, `FORBID`, `PROHIBIT`, `NEVER`, `ALWAYS`, `ONLY`, `SELECT`, `REJECT`, `BLOCKED`, `CLOSED` where normative);
4. semantic contradictions stop the Writer instead of being silently reconciled;
5. accepted T8-E checkpoint decisions already reached in PR #131 are carried into the fresh T8-E proposal after G1.

The text census supports parity; it does not replace semantic review of the closed mapping.

## PR sequence

```text
S0  trustworthy green verification baseline
 ↓
G0  repository documentation profile — PR #132
 ↓
G1  complete documentation census + authority consolidation + consumer repair + legacy deletion + docs verifier
 ↓
close PR #131 as superseded provenance
 ↓
fresh T8-E executable API-contract PR
```

G0/G1 do not authorize product implementation or Product/R10 semantic changes.

## Corrections accepted from independent review

The Fable review at PR #132 HEAD `3b8a25488e1aed5edc6c2b83d64e802b8d66c1c0` identified three blockers and bounded majors. Lead accepts the defect class and selected corrections:

- complete disposition replaces hand-written allowlists;
- gate-subject preservation is a hard invariant;
- executable consumers are repaired, provenance citations are not churned;
- history-pinned secret-scan allowlists may retain deleted paths;
- generated durable docs get generator-owned frontmatter;
- repository safety law and current-system orientation receive durable homes before AGENTS/CLAUDE shrink;
- local proof mirrors actual CI, including `--require-infra`, lint, and affected non-PR governance checks;
- `ready_for_review` must trigger the temporary-work guard;
- G0/G1 provenance is recorded before temporary review files are deleted;
- ADR/reference collections may be indexed collections instead of one nav row per member;
- no unrelated `.claude` permission cleanup is part of this gate.

No Product/R10 or one-docs-root reopen is required.

## Open gate

No legacy deletion starts in PR #132. After Lead adjudication is recorded, the operator decides whether to ratify this corrected profile. Only after ratification and the S0 prerequisite is green may G0 be cleaned for merge and G1 begin from updated `main`.
