---
extends:
  - ~/.claude/rules/ecc/golang/security.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#critical
  - wiki/architecture/trusted-proxy.md
enforced_by:
  - gosec
  - contextcheck
---
# Security Boundaries

## Fail-Closed Authn: UserIDFromContext

Finding: C5 -> Fixed: d2242313

Any function needing actor identity calls `authn.UserIDFromContext(ctx)` and checks the boolean.

```go
func UserIDFromContext(ctx context.Context) (string, bool) {
	raw := strings.TrimSpace(iamdomain.UserIDFromContext(ctx))
	if raw == "" {
		return "", false
	}
	return raw, true
}
```

If `!ok`, the HTTP layer returns 401 or a server-side configuration error through `problem.Write`. Ignoring the boolean or treating `""` as a real actor is forbidden.

## Trusted-Proxy CIDR

Finding: C1 -> Fixed: def24e4a

`X-Forwarded-For`, `X-Forwarded-Proto`, and `X-Real-IP` are trusted only when the request originates from a CIDR in `METALDOCS_TRUSTED_PROXY_CIDRS`.

```go
func LoadTrustedProxyCIDRs() ([]netip.Prefix, error)
func ParseTrustedProxyCIDRs(raw string) ([]netip.Prefix, error)
```

Empty env returns `nil`, meaning no upstream proxy is trusted. Do not substitute `r.RemoteAddr` with a forwarded header before this check.

## RFC 9457 Problem Envelope

Finding: H9 -> Fixed: e1daeeb3

All 4xx and 5xx responses use `problem.Write`.

```go
_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value"))
return
```

Raw `http.Error`, naked `{"error": "..."}` JSON, and status-only error responses are forbidden.

## CORS Reject-Disallowed

Finding: H7 -> Fixed: 6a0d62a7

CORS middleware rejects origins not in the allowlist. Production never uses wildcard `*`. Normalize scheme and host before comparing, and keep origin protection aligned with trusted-proxy scheme resolution.

## Header Trust Rules

Finding: H4 -> Fixed: def24e4a

Middleware order for header-derived identity is:

1. Trusted-proxy CIDR extraction
2. Rate limiting with trusted client IP
3. CORS and origin protection
4. Authn and authz
5. Handler

Rate limiting must not read a spoofable forwarded IP before the trusted-proxy check.

## Authn Callsite Audit Checklist

Before adding a handler:

- auth middleware is in the chain
- `UserIDFromContext` boolean is checked before use
- unauthenticated paths return via `problem.Write`
- authz runs after authn
