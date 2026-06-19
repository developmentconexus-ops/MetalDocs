# F6.5 Plan — IAM routes_memberships Typed Responses

## Steps

1. Read `routes_memberships.go` in full — confirm both `map[string]any` sites.
2. Read `internal/modules/iam/api/api.gen.go` — confirm `GrantAreaMembershipResponse`
   and `UserRole` types.
3. Add `iamapi "metaldocs/internal/modules/iam/api"` import.
4. Add `listMembershipsResponse` struct after `membershipDTO`.
5. Swap Site 1: `writeJSON(w, http.StatusOK, listMembershipsResponse{Items: dtos})`.
6. Swap Site 2: `writeJSON(w, http.StatusCreated, iamapi.GrantAreaMembershipResponse{…})`.
7. `go build ./...` — must be clean.
8. `go test -count=1 ./internal/modules/iam/...` — all pass.
9. Confirm 0 `map[string]any` hits in file.
10. Write artifacts + commit.
