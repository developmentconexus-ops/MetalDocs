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

The durable target is `docs/development/documentation.md`.

## Corrected model

The live tree contains current maintained truth plus bounded Draft-PR work; Git and closed PRs retain history.

Target shape:

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
  work/current/        # temporary only while Draft
```

Durable paths use stable subjects, not R10/T-stage/date/version labels.

Metadata is intentionally small:

```yaml
---
id: architecture-persistence
kind: authority
owner: architecture
summary: Defines target relational persistence and concurrency rules.
---
```

Closed kinds are only `authority` and `work`; no separate status field exists.

## Agent routing

Before root agent files are reduced, surviving rules move to:

```text
docs/development/engineering-rules.md
docs/reference/current-system.md
```

Then `AGENTS.md` routes to `docs/index.md`, `docs/status.md`, active work when present, and 1–3 task-specific authorities. `CLAUDE.md` remains a minimal pointer.

## Active-work protocol

```text
docs/work/current/index.md
docs/work/current/proposal.md
docs/work/current/plan.md
docs/work/current/ai-dialog.md
```

The same proposal/plan are edited in place. The single `ai-dialog.md` carries final Fable review, Lead adjudication, any bounded Round 2, and operator decision. `docs/work/**` may exist only while the PR is Draft and is deleted before Ready.

## Deletion and retention

G1 does not delete by root allowlist and does not require zero textual old-path references.

Every tracked document gets exactly one disposition:

```text
KEEP → target path
MERGE → target authority
GENERATED → generator + target path
DELETE → reason + proof no current consumer remains
```

Undispositioned documentation blocks the gate.

Path occurrences are classified:

```text
EXECUTABLE CONSUMER
→ repair or retire

PROVENANCE/HISTORY CITATION
→ no forced churn unless also executable
```

History-pinned secret-scan fingerprints may retain deleted paths while the full-history scanner consumes them.

## Gate-subject preservation

No verification gate's documentation subject may be deleted without repointing or retiring that gate in the same PR and re-proving the affected check.

G1 explicitly handles current machine-consumed classes such as:

```text
problem codes
requirement traceability
ADRs
database table dictionary/ownership docs
module debt when its gate survives
current runbooks and engineering rules
```

Generated durable pages receive metadata from their generator.

## Product/R10 parity

G1 uses a closed source-to-target authority map from immutable PR #131 provenance. A filesystem-capable normative census supports the semantic review, must be non-empty, and covers the actual normative vocabulary. It is never treated as a substitute for semantic mapping review.

Any contradiction stops the Writer; cleanup cannot silently choose a new Product/R10 decision.

## PR sequence

```text
S0 trustworthy green verification baseline
→ G0 PR #132
→ G1 complete census + consolidation + consumer repair + legacy deletion + verifier
→ close PR #131 as superseded provenance
→ fresh T8-E PR
```

No G0/G1 product implementation is authorized.

## Independent review result

Fable reviewed the original profile and returned:

```text
APPROVE REPOSITORY DOCUMENTATION PROFILE WITH MATERIAL FIXES
BLOCKER 3 / MAJOR 8 / LOW 6
```

Lead accepted all defect classes with bounded corrections now reflected in the durable candidate and plan. The selected one-docs-root, semantic naming, AI-dialog, and PR-gate architecture remains unchanged. No Product/R10 reopen or second Fable round is required.

## Current gate

No legacy deletion has started. Explicit operator ratification of the corrected G0 profile is next. After ratification, S0 must restore the trustworthy green baseline before G0 can be cleaned and merged; G1 begins only from the updated clean `main`.
