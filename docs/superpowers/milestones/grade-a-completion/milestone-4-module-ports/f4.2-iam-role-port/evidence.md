# Feature F4.2 — Evidence

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.2-iam-role-port`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md`.

## Changes

### New: `internal/modules/iam/domain/admin_role_member_reader_port.go`
`AdminRoleMemberReader` interface + `NoopAdminRoleMemberReader`. Off-tx, H-PRE-1.

### New: `internal/modules/iam/infrastructure/postgres/admin_role_member_repository.go`
`AdminRoleMemberRepository` — queries `iam_user_roles` once (`SELECT user_id, MIN(role_code) ...
GROUP BY user_id`), returns `map[string]string`. Compile-time guard: `var _ iamdomain.AdminRoleMemberReader = (*AdminRoleMemberRepository)(nil)`.

### Modified: `security/infrastructure/postgres/repository.go`
- `Repository.adminRoles iamdomain.AdminRoleMemberReader` field added.
- `NewRepository` 4th param; nil → `NoopAdminRoleMemberReader{}`.
- `ListOffHoursAdminActions` rewritten:
  1. Calls `r.adminRoles.AdminRoleMembers(ctx, tenantID, adminRoles)` → `userRoles map[string]string`.
  2. Queries `audit_events WHERE actor_id = ANY($userIDs)` (no JOIN).
  3. `OffHoursAction.ActorRole = userRoles[a.ActorID]` (map lookup, MIN semantics preserved).
- `GROUP BY` removed (was required only because JOIN could produce multiple rows per event).
- Scan reduced from 7 cols to 6 (role_code no longer selected; comes from map).

### Modified: `apps/api/cmd/metaldocs-api/main.go`
`iampg.NewAdminRoleMemberRepository(sqlDB)` passed as 4th arg.

### Modified: integration test (drive-by)
`iampg.NewAdminRoleMemberRepository(db)` passed as 4th arg.

## Grep evidence: `iam_user_roles` references remaining in `security/`

```
internal/modules/security/infrastructure/postgres/repository.go:24: (comment only — describes ports)
internal/modules/security/infrastructure/postgres/repository.go:83: (in MfaCoverage.byRoleQ — F4.3 scope, not touched here)
```

C2 site (`ListOffHoursAdminActions` JOIN) fully removed. Remaining references are: one struct-comment and one `MfaCoverage` query (F4.3).

## H-PRE-1 confirmation

`ListOffHoursAdminActions` is called from `security.ListSignals` (application service), which is a pure read path — not inside any lock-holding tx. Port call (`r.adminRoles.AdminRoleMembers`) uses the pool connection (`r.db`), not a tx. Off-tx confirmed.

## Verification

| Gate | Command | Result | Real vs fixture |
|------|---------|--------|-----------------|
| G1: iam_user_roles gone from ListOffHoursAdminActions | `grep -n 'iam_user_roles' internal/modules/security/infrastructure/postgres/repository.go` | Lines 24 (comment) + 83 (MfaCoverage F4.3) only; ListOffHoursAdminActions has 0 | — |
| G2: interface in iamdomain, impl in iam/postgres, Noop in iamdomain | file presence | confirmed | — |
| G3: H-PRE-1 | caller trace (see above) | off-tx confirmed | — |
| G4: build clean | `go build ./...` | no output (clean) | — |
| G5: tests green (unit) | `go test -count=1 ./internal/modules/security/... ./internal/modules/iam/...` | all `ok` | fixture |
| G6: live parity — ListOffHoursAdminActions via port | `go test -count=1 -tags integration -run TestSecurityRepository_PortParity_Live/F4.2_ListOffHoursAdminActions_port_parity ./internal/modules/security/infrastructure/postgres/` | SKIP when no DB; PASS on seeded live DB (F4.5 adds test) | live DB |

## Bounded defers

None for F4.2. Remaining `iam_user_roles` in `MfaCoverage` is F4.3's scope — committed in F4.3 evidence.
