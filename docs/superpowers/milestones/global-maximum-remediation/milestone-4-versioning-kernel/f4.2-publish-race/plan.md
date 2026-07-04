# F4.2 plan

Executed via subagent; main reviews + commits.

## Task
New `//go:build integration` test (testdb factory) racing manual publish vs scheduled cutover on one
revision. File: `internal/modules/documents/approval/application/publish_race_integration_test.go` (or
alongside existing scheduler/publish integration tests — match the existing testdb fixture idiom).

- Seed: one document with a `controlled_document_id`, status `approved`, a scheduled-publish job input
  matching a `scheduled` variant — actually seed TWO scenarios or flip status per interleaving so both
  the `approved`-side (manual) and `scheduled`-side (scheduler) predicates are exercised against the SAME
  revision_version.
- Interleaving A: manual publish begins first, scheduled second. Interleaving B: scheduled `FOR UPDATE`
  first, manual second. Use a barrier / shared start signal for determinism.
- Assert per contract §2.2: exactly one winner (1 row affected), loser 0 rows + correct sentinel
  (manual→`ErrStaleRevision`, scheduled→`errScheduledPublishNoOp` mapped to nil no-op), terminal status
  `published` once, `revision_version` +1 exactly, single `document_published` event.
- Gate: `go test -tags integration -run <Name> ./internal/modules/documents/approval/...` green (targeted
  only). If the box cannot run `-tags integration`, author the test regardless + record the run as a
  bounded defer (contract §6) with trigger = CI/capable box.

## Contingency (only if race proves unsafe)
Add a single `PublishRevision` method both `publish_service` (manual) and `scheduler_service` route
through; re-run test to green. Record the outcome (safe-as-is vs choke-point-added) in evidence.md.
