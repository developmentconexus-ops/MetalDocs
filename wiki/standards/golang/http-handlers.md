---
extends:
  - ~/.claude/rules/ecc/golang/patterns.md
  - ~/.claude/rules/ecc/golang/security.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/cmd-metaldocs-api.md#critical
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#h7
enforced_by:
  - contextcheck
  - gocyclo
---
# HTTP Handlers

## Handler Anatomy

Every handler follows this sequence:

1. Check method and return 405.
2. Parse, trim, and validate query/body fields.
3. Call the service with `r.Context()`.
4. Map domain state to response DTOs.
5. Encode JSON.

```go
if r.Method != http.MethodGet {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}
limit := 50
if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 1 {
		_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value"))
		return
	}
	limit = parsed
}
items, err := h.service.ListEvents(r.Context(), query)
```

## Context Discipline

Handlers pass `r.Context()` to every service and DB path. `context.Background()` is forbidden in request handling.

## Request Validation at Boundary

Trim whitespace on all string inputs before service calls. Validate ranges, enum values, and required fields at the HTTP boundary.

## No Panic in Handlers

Recovery middleware is a last resort. Handlers return explicit errors via `problem.Write`.

## Middleware Ordering

Canonical order:

1. Recovery
2. Trusted-proxy CIDR extraction
3. Rate limiting using trusted IP
4. Idempotency key extraction
5. CORS
6. Authn
7. Authz
8. Handler

Rate limiting must use the trusted client IP. Authz must never run before authn.

## problem.Write for Error Responses

All non-2xx handler errors use `problem.Write`. Return immediately after every `problem.Write`.

## HTTP Method Routing

Use the current `http.ServeMux` pattern and method checks inside handler bodies. Do not introduce a second method-routing style in one module without an architecture update.
