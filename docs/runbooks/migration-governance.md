# Runbook: Migration Governance

## Current model

- `migrations/` remains the official upgrade path for existing databases.
- `migrations_baseline/` is the canonical fresh-install path for new local environments.
- New schema changes after the baseline cutoff remain forward-only.

## Rules

1. Do not delete historical migrations during the first baseline rollout.
2. Generate a new baseline only from a trusted fully replayed reference DB.
3. Validate baseline-built DBs against runtime gates before changing default bootstrap behavior.
4. Archive policy applies only after validation and explicit review.
