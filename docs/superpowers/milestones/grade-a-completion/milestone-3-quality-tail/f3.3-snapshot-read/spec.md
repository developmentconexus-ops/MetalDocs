# Feature F3.3 — Spec

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.3-snapshot-read`
> **Status:** Approved — 2026-06-16
> **Authored before any code.**

## Consumer contract

**Consumer:** `Pin` path — pays only for the I/O it consumes.

`Pin` currently calls `ReadSnapshotWithFreezeAt`, which issues a SQL `SELECT` across 4 columns
(`placeholder_schema_snapshot`, `composition_config_snapshot`, `body_docx_snapshot_s3_key`,
`values_frozen_at`) but uses only `values_frozen_at`. The 3 snapshot columns are fetched and
immediately discarded (`_ = snap`). After this feature:

- `Pin` calls a new `ReadFreezeAt` method that issues a narrow `SELECT values_frozen_at …` query.
  The 3 blob/string snapshot columns are never fetched on the `Pin` path.
- `Freeze` and `Materialize` paths are unchanged — they call `ReadSnapshotWithFreezeAt` and use
  `snap` fully.
- The `SnapshotReader` interface grows one method: `ReadFreezeAt`. All existing tests still
  compile and pass.

## Non-goals

- No change to `Freeze`, `Materialize`, or any other `FreezeService` method.
- No change to `ReadSnapshotWithFreezeAt` behaviour or signature.
- No change to any domain type or public API shape.
- No freeze-pipeline refactor beyond the single call-site swap in `Pin`.

## Validation Gate

| # | Criterion | Command / proof |
|---|-----------|-----------------|
| 1 | `_ = snap` discard gone from `freeze_service.go` | `grep -n '_ = snap' internal/modules/documents/application/freeze_service.go` → 0 matches |
| 2 | `Pin` calls `ReadFreezeAt`, not `ReadSnapshotWithFreezeAt` | `grep -n 'ReadSnapshotWithFreezeAt\|ReadFreezeAt' internal/modules/documents/application/freeze_service.go` — `Pin` body (≈ line 191) shows `ReadFreezeAt`; `Freeze` / `Materialize` show `ReadSnapshotWithFreezeAt` |
| 3 | `SnapshotReader` interface extended with `ReadFreezeAt` | `grep -n 'ReadFreezeAt' internal/modules/documents/application/freeze_service.go` shows interface + Pin usage |
| 4 | `SnapshotRepository` implements `ReadFreezeAt` with narrow SELECT | `grep -n 'ReadFreezeAt\|values_frozen_at' internal/modules/documents/repository/snapshot_repository.go` |
| 5 | `fakeSnapshotReader` updated | `go build ./...` clean |
| 6 | Whole-repo tests green | `go test ./...` PASS |
| 7 | Runtime proof: `Pin` path executes without snapshot blob fetch | test covers idempotency (already-pinned) and normal-pin cases; fake returns only `valuesFrozenAt` with zero `snap` — proves `Pin` no longer depends on `snap` value |

## Pre-spec investigation

| Question | Finding |
|----------|---------|
| What does `ReadSnapshotWithFreezeAt` SELECT? | 4 cols: `placeholder_schema_snapshot`, `composition_config_snapshot`, `body_docx_snapshot_s3_key`, `values_frozen_at` — confirmed at `repository/snapshot_repository.go:66-80`. |
| What does `Pin` use from the result? | Only `valuesFrozenAt` — `snap` is discarded at `freeze_service.go:195`. |
| What do `Freeze` (line 306) and `Materialize` (line 229) use? | Both use `snap` (schema, composition, S3 key) and the timestamp — they must keep `ReadSnapshotWithFreezeAt`. |
| Is adding `ReadFreezeAt` to `SnapshotReader` an HS-2 trigger? | No. `SnapshotReader` is an application-layer interface owned by the documents module (defined in `freeze_service.go`). Adding a method is a bounded surgical extension, not a shared-module architectural redesign. Milestone spec §rabbit-holes excludes "redesigning the freeze pipeline beyond the discard-fix" — this is exactly the discard-fix. |
| Does `BodyDocxBytes` get fetched by `ReadSnapshotWithFreezeAt`? | No — the query scans `body_docx_snapshot_s3_key` (a string), not the bytes blob. The bytes field in `TemplateSnapshot` is populated elsewhere. Schema + composition JSONs ARE fetched and discarded. |
