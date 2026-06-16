# Feature F4.2 — IAM role-membership port

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.2-iam-role-port`
> **Status:** Approved 2026-06-16 — code change may begin.

## Consumer contract

**Consumer:** `security.ListOffHoursAdminActions` in
`internal/modules/security/infrastructure/postgres/repository.go`.

**What it needs:** a set of (userID → roleCode) pairs for a tenant — filtered to a given set of
admin role codes — so it can scope `audit_events` queries to actors with admin roles without
touching `iam_user_roles` directly.

**Required shape after this feature:**
- A new `iamdomain.AdminRoleMemberReader` interface (in `internal/modules/iam/domain/`) with
  method `AdminRoleMembers(ctx, tenantID, roleCodes) (map[string]string, error)`.
- A Postgres impl `iam.AdminRoleMemberRepository` (in `internal/modules/iam/infrastructure/postgres/`)
  that queries `iam_user_roles` once, returning `map[userID → MIN(role_code)]` for the matching roles.
- A `NoopAdminRoleMemberReader` that returns an empty map (nil error).
- `security.Repository` struct gains an `adminRoles iamdomain.AdminRoleMemberReader` field.
- `NewRepository` accepts it as a 4th arg (nil → Noop default).
- `ListOffHoursAdminActions` calls `r.adminRoles.AdminRoleMembers(ctx, tenantID, adminRoles)` to
  get the user/role map, then queries `audit_events` with `actor_id = ANY($userIDs)` (no JOIN);
  `OffHoursAction.ActorRole` is looked up from the map.

## Anchor (re-verified 2026-06-16)

- `internal/modules/security/infrastructure/postgres/repository.go:328` — `func (r *Repository) ListOffHoursAdminActions(...)`
- `internal/modules/security/infrastructure/postgres/repository.go:345` — `JOIN metaldocs.iam_user_roles`

## Non-goals

- No new IAM HTTP endpoint.
- No redesign of IAM's port family or module API.
- No change to `securitydomain.Repository` interface (the application interface is unchanged).
- No new schema/migration.
- No F4.3 (MfaCoverage) scope in this feature.

## Validation Gate

1. `grep -n 'iam_user_roles' internal/modules/security/infrastructure/postgres/repository.go` returns
   0 matches in `ListOffHoursAdminActions` (the comment on the struct field and any surviving
   MfaCoverage byRoleQ reference are named in `evidence.md`; MfaCoverage is F4.3's scope).
2. `AdminRoleMemberReader` interface lives in `iamdomain`; impl in `iam/postgres`; Noop in `iamdomain`.
3. H-PRE-1: `ListOffHoursAdminActions` is not called inside a lock-holding tx — confirmed by
   tracing callers in `evidence.md`.
4. `go build ./...` clean.
5. `go test -count=1 ./internal/modules/security/... ./internal/modules/iam/...` green.
