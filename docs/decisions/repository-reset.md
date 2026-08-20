---
id: repository-reset
kind: authority
owner: architecture
summary: Ratified clean-slate repository reset and its durable provenance/reopen contract.
---

# Repository clean-slate reset

## Decision

The MetalDocs live repository is architecture-first and excludes the superseded implementation.

```text
PRESERVE CURRENT PRODUCT / ARCHITECTURE TRUTH
+
REACHABLE REQUIRED PROVENANCE
+
MINIMAL DOCUMENTATION + AGENT + CI SPINE
-
SUPERSEDED APPLICATION / DB / OPENAPI / FRONTEND / DEPLOY
-
SUPERSEDED SCRIPTS / TESTS / VERIFIERS / ROADMAP / HARNESS / QA / REVIEW ESTATE
```

The reset direction is operator-ratified and is not reopened by Repository Standard alignment.

## Why

The ratified technical posture establishes clean-slate physical freedom, no historical-business migration requirement, and implementation blocked until later architecture/readiness gates close. Keeping the old implementation live created false constraints and context bloat.

## Durable unmerged provenance refs

PR #131 and PR #132 were never merged to `main`; therefore their exact required bytes need explicit reachable refs.

The GitHub connector available to this execution does not expose annotated-tag creation. Repository Standard v1 §10 permits "a durable annotated tag or another explicit durable ref". The following explicit archive refs were therefore created and remotely proven byte-identical to the reviewed heads:

```text
archive/r10-pr131-pre-reset-20260820
→ d8b1c6d31e704e9552a14faa7764c634a29b081d

archive/repository-governance-pr132-20260820
→ b0ebe54cb010e9837a25f7b778f3d9814d283cb8
```

Proof: GitHub compare reports each archive ref `identical` to its exact source SHA with `ahead_by=0` and `behind_by=0`.

These archive refs MUST NOT be deleted while this page or another current authority names their byte-level provenance as required. The original source branches may be removed only after equivalent reachability has been independently re-proven.

Git history on `main` preserves the removed legacy implementation; the archive refs above preserve the **unmerged** authority/review lineages.

## What survives

Only named current consumers survive in the live tree:

```text
Product Contract / whole-product alignment / ownership
T1 → T8-D ratified authorities
accepted T8-E checkpoint
52 forward decision obligations
Repository Standard-aligned documentation / Git / CI envelope
```

Historical mechanisms return only through the proof-backed reuse gate in `docs/architecture/technical-baseline.md`.

## Relationship to transition

This was source-tree cleanup, not production/business-data cutover. Any future runtime/data transition remains owned by the stage progression in `docs/roadmap.md`.

## Retired controls and reopen triggers

Old implementation/dependency CI was retired because its protected implementation population is zero in the architecture-first tree.

Secret scanning is not currently active because executable implementation/dependency surfaces are absent. **Before the first future implementation/code/schema/runtime commit is authorized, an appropriate secret-scanning control must be restored and proved capable of firing.**

Reopen this reset only if concrete evidence proves a removed mechanism is an externally binding compatibility contract or independently the smallest sustainable realization of a current ratified requirement. Sunk cost, old test coverage, roadmap status, or old CI expectations are not reopen triggers.