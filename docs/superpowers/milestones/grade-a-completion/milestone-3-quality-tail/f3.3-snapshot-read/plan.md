# Feature F3.3 — Plan

> **Spec:** `spec.md` (approved 2026-06-16).

## Tasks

### T1 — Extend `SnapshotReader` interface + update `Pin`

**File:** `internal/modules/documents/application/freeze_service.go`

- Add `ReadFreezeAt(ctx context.Context, tenantID, revisionID string, q ...repository.DBTX) (*time.Time, error)` to `SnapshotReader` interface.
- In `Pin` (≈ line 191): replace
  ```go
  snap, valuesFrozenAt, err := s.snapshots.ReadSnapshotWithFreezeAt(ctx, tenantID, revisionID, tx)
  if err != nil { return fmt.Errorf("pin: read snapshot: %w", err) }
  _ = snap
  ```
  with
  ```go
  valuesFrozenAt, err := s.snapshots.ReadFreezeAt(ctx, tenantID, revisionID, tx)
  if err != nil { return fmt.Errorf("pin: read freeze_at: %w", err) }
  ```

### T2 — Implement `ReadFreezeAt` in `SnapshotRepository`

**File:** `internal/modules/documents/repository/snapshot_repository.go`

Add method with a narrow SELECT:
```sql
SELECT values_frozen_at
  FROM documents
 WHERE tenant_id = $1::uuid AND id = $2::uuid
```
Returns `(*time.Time, error)`. Handles the variadic `DBTX` the same way as `ReadSnapshotWithFreezeAt`.

Note: `ErrNoRows` from `database/sql` should return the same error as `ReadSnapshotWithFreezeAt`
(propagate as-is) — callers treat a missing document as an error.

### T3 — Update `fakeSnapshotReader`

**File:** `internal/modules/documents/application/freeze_service_test.go`

Add method to `fakeSnapshotReader`:
```go
func (f fakeSnapshotReader) ReadFreezeAt(_ context.Context, _, _ string, _ ...repository.DBTX) (*time.Time, error) {
    return f.valuesFrozenAt, f.err
}
```

This fake is used by all three test files (`freeze_service_test.go`, `freeze_pin_test.go`,
`freeze_idempotency_test.go`) in the same package — one update covers all.

## Ordering

T1 first (makes the interface compile), T2 (provides the implementation), T3 (makes tests compile). Single commit after all three.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/documents/application/freeze_service.go` | T1: extend interface + rewrite `Pin` read call |
| `internal/modules/documents/repository/snapshot_repository.go` | T2: `ReadFreezeAt` narrow SELECT |
| `internal/modules/documents/application/freeze_service_test.go` | T3: add `ReadFreezeAt` to fake |
