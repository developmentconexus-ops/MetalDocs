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

Some required byte-level Evidence is intentionally preserved outside the live tree because it must remain exactly recoverable without becoming current Product/status authority.

The GitHub connector available to this execution does not expose annotated-tag creation. Repository Standard v1 §10 permits "a durable annotated tag or another explicit durable ref". The following explicit durable refs were therefore created and remotely proven byte-identical to their exact source heads:

```text
archive/r10-pr131-pre-reset-20260820
→ d8b1c6d31e704e9552a14faa7764c634a29b081d

archive/repository-governance-pr132-20260820
→ b0ebe54cb010e9837a25f7b778f3d9814d283cb8

evidence/t11-pr162-b01-b09-locks-20260824
→ adf58e448bc5bd3a20cae5b7228d729c031f94ac
```

The first two refs preserve unmerged PR #131/#132 authority/review lineages from the repository reset. The T11 Evidence ref preserves exact operator-LOCKED B01-B09 P8 bytes and supporting temporary planning Evidence required by the current `t11-b01-b09-lock-evidence.md` locator for later P11 reconstruction.

Proof: GitHub compare reports each named ref `identical` to its exact source SHA with `ahead_by=0` and `behind_by=0`.

These durable refs MUST NOT be moved or deleted while this page or another current authority names their byte-level provenance as required. The original source branches may be removed only after equivalent reachability has been independently re-proven. Retirement of the T11 Evidence ref additionally follows the explicit retirement law in `t11-b01-b09-lock-evidence.md`.

Git history on `main` preserves the removed legacy implementation; the refs above preserve required **unmerged** authority/review/Evidence lineages.

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