# Migration Baseline and Legacy Recovery Design

## Status

Proposed

## Problem

MetalDocs currently uses a long chronological SQL migration chain for local bootstrap and recovery. The repository has accumulated multiple schema eras, replacement migrations, destructive cleanup steps, and historical churn. Replaying the full chain for local recovery is slow, noisy, and increasingly hard to reason about. When a runtime bug appears, engineers spend time distinguishing real runtime truth from historical migration artifacts.

The immediate pain is local development and recovery time. The deeper issue is migration governance: there is no clean baseline for fresh environments, and the legacy chain is doing too many jobs at once.

## Goals

- Reduce local/bootstrap recovery time for fresh environments.
- Preserve a safe upgrade path for existing databases.
- Create a canonical baseline that reflects real current runtime truth.
- Separate fresh-install bootstrap from historical upgrade history.
- Define a professional migration governance model for future cleanup.

## Non-Goals

- Rewriting production/shared environment migration history immediately.
- Deleting old migrations before baseline validation is complete.
- Changing runtime module ownership, API contracts, or domain behavior as part of this effort.
- Introducing a speculative database redesign unrelated to current schema truth.

## Constraints

- Existing environments must remain upgradeable through the legacy chain for now.
- Fresh baseline generation must come from a fully recovered, trusted database state.
- The current additive-first migration policy remains in force.
- Destructive cleanup must not become part of the default operational path for fresh environments.

## Decision Summary

MetalDocs will adopt a dual-path migration strategy:

1. The current `migrations/` chain remains the official upgrade path for existing/shared/production-like environments.
2. A new baseline path will be created for fresh local/new environments.
3. Future schema evolution continues as forward migrations after the baseline cutoff.
4. Legacy migration cleanup and archival will happen only after the baseline path is validated against runtime truth.

## Considered Approaches

### Approach A - Dual-path baseline

Keep the current legacy migration chain for upgrades and add a separate fresh-install baseline plus post-baseline tail migrations.

Pros:
- Lowest operational risk.
- Fastest path to better local developer experience.
- Preserves upgrade safety while baseline matures.

Cons:
- Two migration entrypoints exist temporarily.
- Repo remains mixed until archival phase completes.

### Approach B - Hard reset migration repository

Replace the current chain immediately with a new canonical baseline and retire the old path.

Pros:
- Cleanest final repository state.
- Simplest mental model once complete.

Cons:
- Highest risk.
- Hard to guarantee compatibility for existing/shared environments.
- Too aggressive for the current repo state.

### Approach C - Snapshot-only local bootstrap

Avoid migration-governance redesign and only speed up local environments with a DB/schema snapshot restore flow.

Pros:
- Fastest short-term dev speed improvement.
- Minimal immediate repo changes.

Cons:
- Does not solve migration governance.
- Leaves historical churn and runtime/schema reasoning problems untouched.
- Creates more tooling truth instead of cleaner schema truth.

## Recommended Approach

Approach A: dual-path baseline.

This gives the largest practical improvement with the lowest risk. It separates the “developer speed” problem from the “upgrade history safety” problem and creates a controlled path toward future archival and cleanup.

## Governance Model

MetalDocs will explicitly operate with three schema truths:

### 1. Legacy truth

The chronological `migrations/` directory remains authoritative for upgrading existing databases.

Use cases:
- existing local DBs that need full replay for debugging
- shared environments
- any environment whose state evolved through historical migrations

### 2. Baseline truth

A new canonical baseline schema represents the best-known current database shape that the application actually expects at runtime.

Use cases:
- fresh local development environments
- fresh ephemeral environments
- future CI/bootstrap paths after validation

### 3. Tail truth

All schema changes after the baseline cutoff remain normal forward migrations applied after the baseline.

Use cases:
- ongoing development after baseline adoption
- preserving append-only forward schema evolution from the cutoff onward

## Required Recovery Rule

Before baseline generation begins, MetalDocs must perform one full trusted recovery pass.

This is mandatory because the local database may be left in a partially migrated or internally inconsistent state if a prior replay was interrupted.

### Recovery prerequisites

1. Reinitialize the local database into a known-empty state.
2. Re-execute the full legacy migration chain end-to-end without interruption.
3. Validate runtime truth after replay:
   - API startup succeeds
   - auth/session flow succeeds
   - key target routes succeed
   - runtime queries match expected schema shape
4. Only after this passes may the baseline be generated.

The baseline must never be derived from a half-recovered or manually patched intermediate database state.

## Execution Model

## Phase 1 - Trusted recovery and schema capture

Objective: establish one trusted reference database.

Steps:
- reset local DB to known-empty state
- run full legacy replay once
- run runtime validation gates
- inspect live schema objects: tables, columns, constraints, indexes, triggers, functions
- record the resulting schema as the source state for baseline generation

Deliverable:
- one validated “reference DB” that represents current system truth

## Phase 2 - Baseline path for fresh environments

Objective: remove full-chain replay from normal fresh local bootstrap.

Steps:
- generate a canonical baseline SQL from the trusted reference DB
- define a baseline cutoff version
- apply baseline + tail migrations for fresh environments
- add a new local bootstrap path that prefers baseline installs
- preserve a separate explicit path for legacy full replay

Deliverables:
- baseline SQL artifact
- baseline-aware bootstrap script/runbook
- fresh-install validation evidence

## Phase 3 - Governance hardening and archival

Objective: reduce repo noise and formalize migration policy.

Steps:
- classify legacy migrations into:
  - required for upgrade history
  - historical-only
  - destructive/dev-reset artifacts
- move historical-only files into an archive strategy once safe
- document when a new baseline may be created in the future
- document how fresh install, legacy replay, and upgrade paths differ

Deliverables:
- migration governance runbook
- archive policy
- cutoff policy for future baselines

## Repository Shape (Target)

The final exact names may vary, but the operational model should look like this:

- `migrations/`
  - legacy chain for upgrade history
- `migrations_baseline/`
  - canonical fresh-install baseline
- `migrations_tail/` or continued numbered migrations after cutoff
  - forward migrations after baseline
- `docs/runbooks/`
  - recovery, baseline generation, and bootstrap rules

If preserving a single numbered directory is operationally easier, the baseline path may instead be managed through scripting rather than directory splitting. The important requirement is the behavioral separation, not the exact folder names.

## Bootstrap Behavior

Local/dev bootstrap should support two explicit modes:

### Mode 1 - Baseline bootstrap (default)

For fresh environments:
- create empty DB
- apply canonical baseline
- apply post-baseline tail migrations
- run runtime validation gates

### Mode 2 - Legacy replay

For debugging or recovery:
- create empty DB
- replay the full legacy chain
- run runtime validation gates

The operator must choose intentionally when using the legacy replay path. It should not remain the hidden default for every normal local bootstrap.

## Validation Rules

Baseline adoption is only valid if all of the following hold:

1. A fresh DB created from baseline + tail matches runtime expectations.
2. Local startup scripts succeed.
3. Contract/runtime gates pass.
4. Critical module flows continue working.
5. The schema generated by baseline + tail is materially equivalent to the trusted reference DB for runtime-used objects.

Validation should compare:
- tables and columns
- PK/FK/unique/check constraints
- indexes
- triggers
- critical functions used by runtime behavior

## Risks

### Risk: baseline captures drift instead of truth

If baseline generation happens from a dirty or manually patched DB, the new path will codify accidental state.

Mitigation:
- mandatory recovery gate before schema capture

### Risk: hidden upgrade assumptions remain in legacy chain

Some shared environments may still rely on historical migration effects not obvious from current schema alone.

Mitigation:
- preserve legacy upgrade path during initial rollout
- do not archive aggressively before verification

### Risk: tooling truth diverges from runtime truth

If baseline scripts and runtime expectations evolve independently, the fresh-install path will drift.

Mitigation:
- require runtime-contract validation after baseline bootstrap
- treat baseline updates like schema changes with verification evidence

## Rollout Recommendation

Recommended rollout:

1. Recover one trusted local reference DB using full legacy replay.
2. Generate baseline from that reference DB.
3. Validate fresh bootstrap using baseline + tail.
4. Make baseline bootstrap the default for local fresh installs.
5. Keep legacy replay available and documented.
6. Audit/archive historical migration noise only after the baseline path proves stable.

## Success Criteria

This project is successful when:

- fresh local bootstrap no longer requires replaying the full historical chain by default
- engineers can still intentionally run the full legacy chain when needed
- a trusted baseline exists and is documented
- runtime gates pass on a fresh baseline-built database
- migration governance is documented clearly enough that future cleanup is intentional rather than ad hoc

