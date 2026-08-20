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
GIT AS ROLLBACK / PROVENANCE
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

No deleted file is destroyed historically: Git and closed pull requests retain it.

## What survives

Only current truth with a named consumer survives:

```text
Product Contract
Whole-product alignment
Ownership topology
T1 → T8-D ratified authorities
active T8-E proposal/checkpoint
repository documentation governance
minimal engineering / agent routing
minimal required CI context
```

A historical mechanism can return only through a current proof-backed reuse decision. Copying legacy because it already works is not sufficient.

## Relationship to T10

This decision advances **source-tree cleanup**, not a production or business-data cutover.

The deleted implementation and its data are DEV/test/throwaway and are not a Launch compatibility contract. Git provides source rollback.

T10 continues to own any future runtime/data cutover that exists once a new implementation has actually been built. T10 does not require the superseded source tree to remain live until then.

## Reopen triggers

Reopen only if concrete evidence proves that a removed file or mechanism is an externally binding compatibility contract or is independently the smallest sustainable implementation of a current ratified requirement.

Sunk cost, historical test coverage, old roadmap status, or old CI expectations are not reopen triggers.