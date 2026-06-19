# F6.5 Spec — IAM routes_memberships Typed Responses

## Feature ID
F6.5

## Milestone
M6 — HS-5 contract sweep (grade-A completion)

## Problem
`internal/modules/iam/delivery/http/routes_memberships.go` contained two
`map[string]any` response sites, violating the H-D typed-response constraint.

## Sites
| Line | Op | Map |
|------|----|----|
| ~168 | listAreaMemberships | `map[string]any{"items": dtos}` |
| ~235 | grantAreaMembership | `map[string]any{"user_id":…, "tenant_id":…, "area_code":…, "role":…}` |

## Fix
**Site 1:** introduce local wrapper `listMembershipsResponse{Items []membershipDTO}`
placed immediately after `membershipDTO` for locality. Do NOT use the codegen
`ListMembershipsResponse` — its `Items []AreaMembership` wire shape differs.

**Site 2:** use `iamapi.GrantAreaMembershipResponse` from
`internal/modules/iam/api/api.gen.go`. Wire shape matches exactly:
`area_code`, `role` (as `UserRole`), `tenant_id`, `user_id`.

## Acceptance
- `go build ./...` clean
- `go test -count=1 ./internal/modules/iam/...` all pass
- `grep -n 'map\[string\]any' routes_memberships.go` → 0 matches
