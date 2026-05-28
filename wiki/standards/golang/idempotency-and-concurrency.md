---
extends:
  - ~/.claude/rules/ecc/golang/patterns.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#c3--idempotency-store-has-no-locking--concurrent-same-key-requests-both-execute-the-handler-db--fixed-in-12cae0f9
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#c4--idempotency-schema-defines-in_flight--failed-states-but-go-code-never-writes-them-db--fixed-in-12cae0f9
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#h11--idempotency-actor_user_id-text-should-be-uuid--fk-and-unbounded-response_body-jsonb-db
enforced_by:
  - manual-review
---
# Idempotency and Concurrency

## Two-Phase Write Pattern

Finding: C3 -> Fixed: 12cae0f9
Finding: C4 -> Fixed: 12cae0f9
Finding: C3/C4 evidence -> Fixed: 07312d58

Mutating handlers that use idempotency call `BeginReplay` before any state change and then call exactly one of `CompleteReplay` or `FailReplay`.

```go
handle, replay, err := store.BeginReplay(ctx, tenantID, actorID, key, payloadHash)
```

Outcomes and obligations:

- `(handle, nil, nil)`: this caller owns the key; run the handler and call `CompleteReplay` on success or `FailReplay` on panic/non-2xx.
- `(nil, replay, nil)`: cache hit; write `replay` and skip the handler.
- `(nil, nil, ErrConflict)`: same key with different payload; return 409.

## ON CONFLICT DO NOTHING RETURNING

The canonical claim query uses `INSERT ... ON CONFLICT DO NOTHING RETURNING`. Only the winning transaction sees `RETURNING`; losers block on `SELECT ... FOR UPDATE` and read the completed row or retry after rollback. This avoids check-then-act races.

## Retry-Safe Handler Semantics

A retry-safe handler produces the same visible outcome for repeated calls with the same idempotency key. POST, PUT, PATCH, and DELETE handlers are either idempotency-key gated or naturally idempotent, such as DELETE treating already-absent state as success.

## H11 Go-Layer Guards

Finding: H11 -> Fixed: 12cae0f9

The store rejects empty actors before any DB round-trip and refuses persisted response bodies above 64 KiB.

```go
if actorID == "" {
	return nil, nil, errors.New("idempotency: actorID must not be empty")
}
```

## Replay-Race Fix

C4 documented schema/code drift: `in_flight` and `failed` existed in DDL but the Go store only wrote `completed`. The two-phase API restores the intended state machine and prevents two concurrent same-key requests from both executing the handler.

## defer-FailReplay Pattern

Use a sentinel so panics and early non-success returns free the in-flight slot:

```go
completed := false
defer func() {
	if !completed {
		_ = store.FailReplay(handle, err)
	}
}()
```

## Out of Scope

Distributed locks and saga orchestration are v2 topics. Do not introduce them for local HTTP idempotency.
