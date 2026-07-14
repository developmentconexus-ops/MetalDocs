---
extends:
  - ~/.claude/rules/ecc/golang/coding-style.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#c3--idempotency-store-has-no-locking--concurrent-same-key-requests-both-execute-the-handler-db--fixed-in-12cae0f9
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#h1--idempotency-recordreplay-write-error-silently-swallowed--broken-idempotency-on-transient-db-error-gsfdb--fixed-in-12cae0f9-incidental-via-c3
  - wiki/reviews/2026-05-21-go-backend-review/cmd-metaldocs-api.md#h4-documentsauditadapter-marshals-payload-falls-back-to-literal--on-error
enforced_by:
  - errcheck
  - errorlint
  - nilerr
---
# Errors and Logging

## Error Wrapping Rule

Every returned error crossing a package boundary is wrapped as `"<subsystem>: <operation>: %w"`.

Do:

```go
return fmt.Errorf("idempotency: begin tx: %w", err)
return fmt.Errorf("idempotency: insert in_flight: %w", err)
```

Do not:

```go
return fmt.Errorf("begin tx failed: %v", err)
return err
```

## Never Swallow

`_ = err` is banned except for documented non-actionable write paths:

```go
_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value"))
```

The audit payload `json.Unmarshal` exception in `internal/modules/audit/delivery/http/handler.go` is tolerated because malformed historical payloads degrade to `{}` for a read-only response. Every other discarded error needs a short comment explaining why it cannot fail.

Do:

```go
if err := store.RecordReplay(ctx, tenantID, actorID, key, hash, status, body); err != nil {
	slog.ErrorContext(ctx, "idempotency: record replay failed", "error", err)
}
```

Do not:

```go
_ = store.RecordReplay(ctx, tenantID, actorID, key, hash, status, body)
```

## errors.Is / errors.As Discipline

Sentinel errors are package-level `var Err... = errors.New(...)`. Callers use `errors.Is` or `errors.As`; string matching is forbidden.

Do:

```go
if errors.Is(err, idempotency.ErrConflict) {
	_ = problem.Write(w, problem.New(http.StatusConflict, "IDEMPOTENCY_KEY_REUSED", "Idempotency key reused"))
	return
}
```

Do not:

```go
if strings.Contains(err.Error(), "conflict") {
	w.WriteHeader(http.StatusConflict)
}
```

## slog Conventions

Use `slog.ErrorContext` and `slog.InfoContext` when a context is available. Messages start with the subsystem and do not interpolate data; fields are structured key/value pairs.

Do:

```go
slog.ErrorContext(ctx, "audit_integrity_validator: validation failed", "error", err)
slog.InfoContext(ctx, "audit_integrity_validator: tick complete", "validated", n)
```

Do not:

```go
log.Printf("audit failed: %v", err)
fmt.Println("validated", n)
```

`log.Fatal` is allowed only in `main`/`cmd` startup paths where the process must exit.

## Log Or Return

Do not log an error and then return the same error to a caller that will log it again. Add context and return. Log at package boundaries where the caller will otherwise lose operational context.
