---
extends:
  - ~/.claude/rules/ecc/golang/patterns.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/cmd-metaldocs-api.md#c2-permission-resolver-defaults-to-fail-open
  - wiki/reviews/2026-05-21-go-backend-review/cmd-metaldocs-api.md#c3-server-error-shutdown-path-bypasses-deferred-cleanup
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#h10--ratelimitconfigquotas-is-a-public-mutable-map-with-no-validation-zero-value-panics-t
enforced_by:
  - depguard
  - manual-review
---
# Package Layout

## Module Template

```text
internal/modules/<name>/
  domain/         # types, interfaces, sentinel errors
  application/    # service, use-case orchestration
  store/          # DB access
  delivery/http/  # handler, routes, request/response types
```

## Import Direction Law

Imports flow `delivery -> application -> store -> domain`; reverse imports are forbidden.

Domain packages do not import application, store, delivery, or platform packages except deliberately tiny standard-library dependencies. Delivery/application/store can import `internal/platform/*` for shared infrastructure.

## Platform Packages

`internal/platform/` is shared infrastructure: `authn`, `problem`, `idempotency`, `ratelimit`, `security`, `tenant`, `httpresponse`, `config`, `db`, and observability helpers. Module-specific behavior stays in `internal/modules/<name>/`.

## Constructor Invariant Pattern

Finding: H10 -> Fixed: pending/refactor target

Types with required fields expose constructors that validate mandatory values and return `(T, error)`. Public mutable maps and valid-looking zero values are forbidden for config with invariants.

```go
func NewConfig(quotas map[RouteKey]int) (Config, error) {
	for route, quota := range quotas {
		if quota < 1 {
			return Config{}, fmt.Errorf("ratelimit: invalid quota for %s", route)
		}
	}
	return Config{quotas: maps.Clone(quotas)}, nil
}
```

## cmd Entrypoints

`apps/api/cmd/metaldocs-api` is the HTTP server entrypoint. Business logic lives in modules and platform packages. Wiring lives in `apps/api/internal/wiring/`; that package may import across modules because composition is its job.

## No Import Cycles

If two packages need each other, extract shared types to a third package, usually `domain`.

## Generated Code

`internal/api/v2/types_gen.go` and `api.gen.go` are excluded from lint. Never edit generated files by hand.
