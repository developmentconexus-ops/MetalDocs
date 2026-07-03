// Package iamtypes holds the canonical Role enum as a neutral platform type.
//
// Role previously lived in internal/modules/iam/domain, which made
// internal/modules/auth import iam/domain (for Role) while iam imported
// auth/domain (for ManagedUser et al.) — a bidirectional module dependency
// that was non-circular only by accident of which sub-packages were involved
// (ARC-06 / wiki/modules/auth-tech-debt.md T-007, wiki/modules/iam-tech-debt.md
// T-010, wiki/backlog/iam-refactor.md R-010). Role is a pure value type with no
// dependency on IAM's application or infrastructure layers, so it belongs in a
// neutral platform package that both auth and iam (and any other module) can
// import without creating a module-to-module edge.
//
// iam/domain.Role remains a type alias to this package (Role = iamtypes.Role)
// so existing call sites across the codebase keep compiling unchanged; new
// code should prefer importing iamtypes directly when Role is the only thing
// needed from iam/domain.
package iamtypes

import "sort"

// Role is a canonical tenant role code (ADR 0022 §capability catalog; roles are
// the assignment/grouping unit — authorization itself is always by capability,
// never by role).
type Role string

// Canonical tenant roles (8). Source: wiki/modules/iam.md §canonical-roles.
const (
	RoleApprover    Role = "approver"
	RoleAreaAdmin   Role = "area_admin"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleQmsAdmin    Role = "qms_admin"
	RoleSigner      Role = "signer"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)

var validRoles = map[Role]struct{}{
	RoleApprover:    {},
	RoleAreaAdmin:   {},
	RoleAuthor:      {},
	RoleEditor:      {},
	RoleQmsAdmin:    {},
	RoleSigner:      {},
	RoleSystemAdmin: {},
	RoleViewer:      {},
}

// IsValidRole reports whether role is one of the eight canonical roles.
func IsValidRole(role Role) bool {
	_, ok := validRoles[role]
	return ok
}

// areaRoles is the subset of canonical roles assignable as an AREA membership
// (a public.user_process_areas row). It is every canonical role EXCEPT
// system_admin: system_admin is a tenant-wide tier-1 role that bypasses tier-2
// and is never an area membership. This set is the Go single source of truth for
// area-assignable roles and mirrors the user_process_areas role CHECK constraint
// in the DB. Consumers that resolve actors by an area role (e.g. the approval
// module's stage required_role, joined against user_process_areas.role) bind
// against this set so a role no user can ever hold cannot be configured
// (ADR 0022 — role strings are bound to the registry, never free text).
var areaRoles = map[Role]struct{}{
	RoleApprover:  {},
	RoleAreaAdmin: {},
	RoleAuthor:    {},
	RoleEditor:    {},
	RoleQmsAdmin:  {},
	RoleSigner:    {},
	RoleViewer:    {},
}

// IsAreaRole reports whether role can be held as an area membership
// (user_process_areas). system_admin is intentionally excluded — it is tenant-wide.
func IsAreaRole(role Role) bool {
	_, ok := areaRoles[role]
	return ok
}

// AreaRoles returns the canonical area-assignable roles, sorted — for validation
// messages and diagnostics.
func AreaRoles() []Role {
	out := make([]Role, 0, len(areaRoles))
	for r := range areaRoles {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
