---
extends:
  - ~/.claude/rules/ecc/golang/patterns.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#h11--idempotency-actor_user_id-text-should-be-uuid--fk-and-unbounded-response_body-jsonb-db
  - "wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#m9--idempotencyreplaystatus-int-not-range-validated---writeheader0-possible-on-corrupt-row-t"
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#m11--no-fk-from-idempotency_keystenant_id-to-tenants-db
enforced_by:
  - sqlclosecheck
  - rowserrcheck
  - bodyclose
---
# Persistence

## pgx v5 as the Driver

New persistence code uses `github.com/jackc/pgx/v5`. `github.com/lib/pq` and `database/sql` remain where already present, but new stores should not add new `lib/pq` usage.

## Parameterized Queries Only

Do:

```go
rows, err := db.QueryContext(ctx, "SELECT id FROM documents WHERE tenant_id = $1", tenantID)
```

Do not:

```go
rows, err := db.QueryContext(ctx, "SELECT id FROM documents WHERE tenant_id = '"+tenantID+"'")
```

No `fmt.Sprintf` SQL with user-controlled values.

## RowsAffected Discipline

UPDATE and DELETE paths check `RowsAffected()`. A zero-row mutation returns a domain error instead of silently succeeding.

```go
n, _ := res.RowsAffected()
if n == 0 {
	return ErrNotFound
}
```

## Rows.Close Discipline

Call `defer rows.Close()` immediately after a successful query and check `rows.Err()` after iteration.

```go
rows, err := db.QueryContext(ctx, q, tenantID)
if err != nil {
	return fmt.Errorf("documents: list: %w", err)
}
defer rows.Close()
for rows.Next() {
	// scan
}
if err := rows.Err(); err != nil {
	return fmt.Errorf("documents: list rows: %w", err)
}
```

## Transaction Boundaries

Transactions open in store packages. The same function that calls `Begin` owns commit and rollback. Services orchestrate use cases and call store methods; they do not keep transactions open across HTTP or external calls.

## Connection Pool Hygiene

Do not hold a connection across a network round-trip. Every DB call receives the request or job context. Context cancellation must reach the driver.

## No New sqlmock

`github.com/DATA-DOG/go-sqlmock` is present in `go.mod` as legacy. New DB behavior tests use real Postgres through testcontainers or the CI Postgres service.
