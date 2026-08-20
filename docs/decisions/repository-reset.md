---
id: repository-reset
kind: authority
owner: architecture
summary: Authorizes removing the superseded MetalDocs implementation from the live repository before new implementation begins.
---

# Repository clean-slate reset

## Decision

The MetalDocs repository is reset to an architecture-first baseline before new implementation.

```text
PRESERVE CURRENT PRODUCT / ARCHITECTURE TRUTH
+
REACHABLE GIT PROVENANCE
+
MINIMAL DOCUMENTATION + AGENT + CI SPINE
-
SUPERSEDED APPLICATION CODE
-
SUPERSEDED DATABASE / MIGRATIONS
-
SUPERSEDED OPENAPI / GENERATED CLIENTS
-
SUPERSEDED FRONTEND
-
SUPERSEDED DEPLOY / CONTAINERS / ENVIRONMENT FILES
-
SUPERSEDED SCRIPTS / TESTS / VERIFIERS / QUALITY BASELINES
-
SUPERSEDED ROADMAP / HARNESS / QA / REVIEW ARTIFACTS
```

## Why

The ratified technical posture already establishes:

- clean-slate physical target freedom;
- current implementation is evidence, not target authority;
- current DEV/test data has no historical-business compatibility consumer;
- implementation remains blocked while T8-E and later realization gates are unfinished.

Maintaining the old implementation during architecture rederivation creates false constraints, context bloat, stale CI obligations, and pressure to preserve shapes that have already lost architectural authority.

## Scope

The live tree removes the old implementation including its application code, API specification, generated code, database schema/migrations, frontend, workers/jobs, deployment configuration, container definitions, package manifests, scripts, tests, verifier registry, runbooks tied only to that implementation, local skills, roadmaps, QA reports, and historical documentation.

Deletion from the live tree is not permission to lose provenance. The Product/R10 corpus in PR #131 and G0 review history in PR #132 were never merged to `main`; therefore closed PR state alone is insufficient as the rollback guarantee.

## Protected provenance refs

The following source branches remain reachable and MUST NOT be deleted until equivalent immutable archival tags/refs are created and this section is updated:

```text
docs/a8-authz-approval-redesign-ledger
@ d8b1c6d31e704e9552a14faa7764c634a29b081d

  preserves PR #131 ratified authority/review corpus and source blobs

docs/repository-information-architecture
@ b0ebe54cb010e9837a25f7b778f3d9814d283cb8
  preserves PR #132 documentation-governance review provenance
```

Git history on `main` preserves the removed legacy implementation. The refs above specifically preserve **unmerged** authority/review objects until immutable archival refs replace them.

## What survives

Only current truth with a named consumer survives:

```text
Product Contract
Whole-product alignment
Ownership topology
T1 → T8-D ratified authorities
paused accepted T8-E checkpoint
remaining stage program
52 forward decision obligations
repository documentation governance
minimal engineering / agent routing
minimal required CI context
```

A historical mechanism can return only through a current proof-backed reuse decision. Copying legacy because it already works is not sufficient.

## Relationship to T10

This decision advances **source-tree cleanup**, not a production or business-data cutover.

The deleted implementation and its data are DEV/test/throwaway and are not a Launch compatibility contract. Reachable Git history provides source rollback.

T10 continues to own any future runtime/data cutover that exists once a new implementation has actually been built. T10 does not require the superseded source tree to remain live until then.

## Retired controls and reopen triggers

The old dependency/security/implementation CI was retired because the live repository contains no executable application/dependency graph while implementation is blocked.

Secret scanning is intentionally not an active clean-tree job now. **Before the first implementation/code/schema/runtime commit is allowed, T11/T12 or an earlier explicit security gate MUST restore an appropriate secret-scanning control for the new repository shape.**

Reopen this reset only if concrete evidence proves that a removed file or mechanism is an externally binding compatibility contract or is independently the smallest sustainable implementation of a current ratified requirement.

Sunk cost, historical test coverage, old roadmap status, or old CI expectations are not reopen triggers.