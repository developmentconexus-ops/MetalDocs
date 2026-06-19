# F6.5 Evidence — IAM routes_memberships Typed Responses

## Changes

File: `internal/modules/iam/delivery/http/routes_memberships.go`

1. Added import `iamapi "metaldocs/internal/modules/iam/api"`.
2. Added `listMembershipsResponse` struct (lines 66-68, after `membershipDTO`):
   ```go
   type listMembershipsResponse struct {
       Items []membershipDTO `json:"items"`
   }
   ```
3. Site 1 (listAreaMemberships): replaced
   `map[string]any{"items": dtos}` →
   `listMembershipsResponse{Items: dtos}`
4. Site 2 (grantAreaMembership): replaced
   `map[string]any{"user_id": …, "tenant_id": …, "area_code": …, "role": …}` →
   `iamapi.GrantAreaMembershipResponse{UserId: userID, TenantId: tenantID, AreaCode: areaCode, Role: iamapi.UserRole(role)}`

## Verification

### map[string]any grep
```
$ grep -n 'map\[string\]any' routes_memberships.go
(no output — 0 matches)
```

### go build ./...
```
$ go build ./...
(clean — no output)
```

### go test -count=1 ./internal/modules/iam/...
```
?   metaldocs/internal/modules/iam/api               [no test files]
ok  metaldocs/internal/modules/iam/application        3.151s
ok  metaldocs/internal/modules/iam/authz              2.324s
ok  metaldocs/internal/modules/iam/delivery/http      4.166s
ok  metaldocs/internal/modules/iam/domain             1.172s
ok  metaldocs/internal/modules/iam/infrastructure/memory  1.203s
ok  metaldocs/internal/modules/iam/infrastructure/postgres 3.302s
ok  metaldocs/internal/modules/iam/presence           4.684s
```

## Status
PASS — 0 `map[string]any` in target file; build clean; all IAM tests pass.
