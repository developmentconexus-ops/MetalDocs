## 1. Entry point
| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | (unclear: route handwritten — no OpenAPI declaration) | api/openapi/v1/openapi.yaml (no `listAreaMemberships` match) |
| Generated server stub | n/a — hand-written stdlib mux | internal/modules/iam/delivery/http/routes_memberships.go:30 |
| Handler | `(*MembershipHandler).listMemberships` | internal/modules/iam/delivery/http/routes_memberships.go:35 |

## 2. Call chain
1. `apps/api/cmd/metaldocs-api/main.go:386` composed handler chain (`authMiddleware.Wrap(iamMiddleware.Wrap(...mux))`) — tier-1 IAM middleware runs before route dispatch then -> calls: `internal/modules/iam/delivery/http/middleware.go:49` `(*Middleware).Wrap`
2. `internal/modules/iam/delivery/http/middleware.go:61` `resolver(r.Method, r.URL.Path)` + `m.caps.CanDo(...)` gate — resolves `/api/v2/iam/area-memberships` capability and enforces tier-1 authz then -> calls: `apps/api/cmd/metaldocs-api/permissions.go:196` `newPermissionResolver` mapping and `internal/modules/iam/application/capability_service.go:31` `(*CapabilityService).CanDo`
3. `internal/modules/iam/application/capability_service.go:64` `s.db.QueryRowContext(...).Scan(&allowed)` — DB capability check (`iam_user_roles`/`iam_group_*`/`role_capabilities`) before handler then -> returns to middleware -> `next.ServeHTTP`
4. `internal/modules/iam/delivery/http/routes_memberships.go:35` `(*MembershipHandler).listMemberships` — validates `userId` (query or authenticated actor), resolves tenant, invokes use case then -> calls: `internal/modules/iam/application/area_membership_service.go:45` `(*AreaMembershipService).ListActive`
5. `internal/modules/iam/application/area_membership_service.go:46` `repo.ListActive(ctx, userID, tenantID, now)` — application pass-through for active memberships then -> calls: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:21` `(*UserAreaRepository).ListActive`
6. `internal/modules/iam/infrastructure/postgres/user_area_repository.go:31` `r.db.QueryContext(ctx, q, userID, tenantID, now)` — SELECT active rows from `user_process_areas`; scans rows and returns response items (DB driver boundary).

## 3. State changes
none

## 4. SQL touched
| File:line | Verb | Table(s) | Auth-area arg |
|---|---|---|---|
| internal/modules/iam/application/capability_service.go:64 | SELECT (EXISTS) | `metaldocs.iam_user_roles`, `metaldocs.iam_group_members`, `metaldocs.iam_group_roles`, `metaldocs.role_capabilities` | tenant-level (`tenantID`), capability from resolver (`membership.manage`) |
| internal/modules/iam/infrastructure/postgres/user_area_repository.go:31 | SELECT | `user_process_areas` | tenant-level filter (`tenantID`), user filter (`userID`) |

Tripwire pairing: N/A (read)

## 5. Response shape
- 2xx schema ref: (unclear: route handwritten — no OpenAPI declaration for `listAreaMemberships`); concrete handler payload is `200` JSON object `{ "items": []domain.UserProcessArea }` from `internal/modules/iam/delivery/http/routes_memberships.go:55`.
- Error responses + Problem type URI: route returns `{code,message}` via `writeMembershipAPIError` (`internal/modules/iam/delivery/http/routes_memberships.go:137`) and middleware may return `{error:{code,message,details,trace_id}}` via `writeAPIError` (`internal/modules/iam/delivery/http/middleware.go:129`); RFC 9457 Problem `type` URI is not used here (unclear: no Problem payload on this flow).

## 6. Cross-references
- Idempotency: no (no idempotency middleware/store in this route path; handler registered directly at `internal/modules/iam/delivery/http/routes_memberships.go:30`)
- Pagination: no + cursor field name: (unclear: none in handler response)
- Audit log emission: no for GET list flow (membership service logger used by grant/revoke paths only; `ListActive` is pass-through at `internal/modules/iam/application/area_membership_service.go:45`)
