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

evidence/t11-pr170-b10-locks-20260824
→ b8c607cbd30d61d6bcf6ec1ea734ed1653d2569e
```

The first two refs preserve unmerged PR #131/#132 authority/review lineages from the repository reset. The T11 Evidence refs preserve exact operator-LOCKED frontend Evidence required for later P11 reconstruction: B01-B09 are routed by `t11-b01-b09-lock-evidence.md`; B10 P8/P9/P10 is routed by `t11-b10-lock-evidence.md`.

Proof: each current Evidence/provenance ref is exact-SHA pinned by the repository aggregate gate. The B10 ref was additionally remotely resolved immediately after creation to the exact pre-cleanup PR #170 HEAD above.

These durable refs MUST NOT be moved or deleted while this page or another current authority names their byte-level provenance as required. The original source branches may be removed only after equivalent reachability has been independently re-proven. Retirement of each T11 Evidence ref additionally follows the explicit retirement law in its corresponding Evidence locator.

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
