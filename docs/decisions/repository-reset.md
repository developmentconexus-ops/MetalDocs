---
id: repository-reset
kind: authority
owner: architecture
summary: Ratified clean-slate repository reset and durable provenance/reopen contract.
---

# Repository clean-slate reset

## Decision

The MetalDocs live tree is architecture-first and excludes the superseded application implementation.

```text
PRESERVE CURRENT PRODUCT / ARCHITECTURE TRUTH
+
REACHABLE REQUIRED UNMERGED PROVENANCE
+
MINIMAL DOCUMENTATION + AGENT + CI SPINE
-
SUPERSEDED APPLICATION / DB / OPENAPI / FRONTEND / DEPLOY
-
SUPERSEDED SCRIPTS / TESTS / VERIFIERS / ROADMAP / HARNESS / QA / REVIEW ESTATE
```

This reset is operator-ratified. Repository-governance storage/mechanism changes do not reopen the clean-slate Product/technical decision.

## Why

The ratified technical posture gives MetalDocs clean-slate physical freedom, requires no historical business-data migration for Launch, and keeps implementation blocked until later readiness gates close. Keeping superseded implementation in the live tree created false constraints and context bloat.

## Required unmerged provenance

Some byte-level provenance is intentionally preserved outside the live tree because Git history on `main` cannot preserve content that was never merged.

Current explicit durable refs required by this authority/its frontend locators:

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

The first two refs preserve required unmerged pre-reset authority/review lineage. The T11 refs preserve exact operator-LOCKED frontend Evidence for later P11 reconstruction and are routed by their durable Evidence locators.

### Reachability law

These refs must not be moved/deleted while a current authority names their exact provenance as required.

Current `required` CI does **not** network/SHA-check every historical ref on every unrelated PR. That is deliberate: provenance correctness is proved when the current claim or cleanup action depends on it.

Before removing an original branch/ref that might be the last required path to unmerged material:

```text
identify current consumer
→ verify surviving durable authority
→ resolve the named archive/Evidence ref to the expected exact commit/blob
→ only then remove the redundant ordinary branch
```

Merged legacy implementation remains recoverable through `main` Git history; no working-tree archive is required.

## What survives in the live tree

Only current named consumers survive, including:

```text
Product contract / alignment / ownership
current T1→T10 Product and architecture authorities
current bounded T11 decisions already integrated
forward obligations still relevant to future stages
local repository methods / routers / objective CI spine
```

Historical mechanisms return only through the proof-backed reuse gate in `docs/architecture/technical-baseline.md`.

## Relationship to transition

This reset was source-tree cleanup, not production/business-data cutover. Future runtime/data transition remains owned by `docs/roadmap.md` + `docs/architecture/transition.md`.

## Retired controls and future implementation gate

Old implementation/dependency CI was retired because its protected implementation population in the live tree is zero.

Secret scanning is not currently active because executable implementation/dependency surfaces are absent. **Before the first future implementation/code/schema/runtime commit is authorized, an appropriate secret-scanning control must be restored and proved capable of firing.**

## Reopen triggers

Reopen this reset only if concrete Evidence proves a removed mechanism is:

- an externally binding compatibility contract; or
- independently the smallest sustainable realization of a current accepted requirement.

Sunk cost, old test coverage, historical branch existence, or retired CI expectations are not reopen triggers.
