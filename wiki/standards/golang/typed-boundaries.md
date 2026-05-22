---
extends:
  - ~/.claude/rules/ecc/golang/coding-style.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#c5--authnuseridfromcontext-returns--silently-on-missing-key--cross-tenant-list-exposure-sf--fixed-in-d2242313
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#h9
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md#h11
enforced_by:
  - exhaustive
  - revive
---
# Typed Boundaries

## The Rule

`string` is banned for IDs, roles, problem codes, and idempotency keys at hand-written package boundaries. Use distinct named types so invalid values cannot drift between auth, tenant, service, store, and HTTP layers unnoticed.

Generated OpenAPI types are exempt because `internal/api/v2/types_gen.go` and `api.gen.go` are generated and excluded from lint. Hand-written domain, service, and store APIs are not exempt.

## Proven Types

Source: `internal/modules/iam/domain/model.go:3`

```go
type Role string

const (
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)
```

Source: `internal/platform/problem/codes.go:7`

```go
type Code string

const (
	CodeValidationError Code = "VALIDATION_ERROR"
	CodeUnauthenticated Code = "UNAUTHENTICATED"
)
```

## Migration Pattern

Define boundary types in the owning domain package and migrate call sites one layer at a time.

```go
type TenantID string
type UserID string
type IdempotencyKey string
```

Wrap at ingress, then narrow service and store signatures. Do not replace every local helper at once unless the package boundary already changed.

## Anti-Patterns

Do not expose raw string roles:

```go
func CanApprove(role string) bool
```

Do expose domain roles:

```go
func CanApprove(role iamdomain.Role) bool
```

Do not expose raw actor IDs:

```go
func List(ctx context.Context, tenantID string, userID string) ([]Document, error)
```

Do expose named IDs once the owning domain type exists:

```go
func List(ctx context.Context, tenantID tenant.ID, userID iamdomain.UserID) ([]Document, error)
```

No bare `string` parameter named `role`, `tenantID`, `userID`, `errorCode`, or `idempotencyKey` is acceptable across package boundaries.

## When Typed ID Is Not Needed

Short internal helpers can use `string` when they are package-local, do not cross domain/service/store/HTTP boundaries, and cannot be confused with another identifier class.
