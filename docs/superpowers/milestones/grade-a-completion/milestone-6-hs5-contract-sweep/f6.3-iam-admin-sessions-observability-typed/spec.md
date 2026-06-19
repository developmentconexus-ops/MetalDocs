# F6.3 — IAM Admin / Sessions / Observability Typed Responses

## Consumer Contract (per operation)

| Operation | File | Old shape | New shape |
|---|---|---|---|
| POST /api/v1/iam/users/{user_id}/roles (200) | admin_handler.go | `map[string]any{user_id, role, display_name}` | `iamapi.UpsertUserRoleResponse{UserId, Role, DisplayName *string}` |
| PUT /api/v1/iam/users/{user_id}/roles (200) | admin_handler.go | `map[string]any{user_id, display_name, roles}` | `iamapi.ReplaceUserRolesResponse{UserId, DisplayName, Roles []ReplaceUserRolesResponseRoles}` |
| GET /api/v1/auth/sessions (200) | sessions_handler.go | `map[string]any{items []map[string]any, page map[string]any}` | `iamapi.ListSessionsResponse{Items []SessionItem, Page CursorPage}` |
| GET /api/v1/iam/usage (200) | observability_handler.go | `map[string]any{seats,storage,api_calls,active_users,plan_tier}` | `iamapi.UsageSnapshot` |
| GET /api/v1/iam/kpi (200) | observability_handler.go | `map[string]any{locked_accounts,...,role_distribution []map[string]any}` | `iamapi.IamKpiSnapshot{..., RoleDistribution []IamKpiRoleCount}` |

## Validation Gate

- [ ] `go build ./...` passes after each file edit
- [ ] `go test -count=1 ./internal/modules/iam/...` green
- [ ] `go test -count=1 ./...` green
- [ ] `grep -rn 'map\[string\]any' admin_handler.go sessions_handler.go observability_handler.go` = 0 hits in those three files
- [ ] JSON field names unchanged (backward-compat): `session_id`, `user_id`, `display_name`, `created_at`, `last_seen_at`, `expires_at`, `ip_address`, `user_agent`, `device_label`, `has_more`, `next_cursor`, `seats`, `storage`, `api_calls`, `active_users`, `plan_tier`, `locked_accounts`, `mfa_coverage_pct`, `failed_logins24h`, `dormant_users30d`, `role_distribution`, `audit_events_per_minute`
