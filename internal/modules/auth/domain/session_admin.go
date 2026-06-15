package domain

import "database/sql"

// SessionAdminQuery is the parameter object for session administration queries.
//
//   - TenantID is required (the tenant of the calling admin)
//   - UserID, when set, narrows to a single user's sessions
//   - IncludeRevoked toggles the "active only" filter off — by default the
//     query hides revoked + expired sessions, matching the People tab UX
//     where the admin wants to see live sessions to revoke
//   - Limit caps the row count (max 200; OpenAPI cap)
type SessionAdminQuery struct {
	TenantID       string
	UserID         string
	IncludeRevoked bool
	Limit          int
}

// SessionListItem is the row shape returned to the Admin Center
// Sessions & Security tab. It carries only auth-owned auth_sessions columns;
// display names are owned by IAM (metaldocs.iam_users) and resolved by the IAM
// consumer via the UserDisplayNameReader port (M4/F4.4), so auth never reaches
// across the module boundary into iam_users.
//
// Tenant scoping is enforced inside the repository implementation: every query
// is filtered on s.tenant_id = $1, so a session belonging to another tenant is
// not visible even if its user_id collides.
type SessionListItem struct {
	SessionID  string
	UserID     string
	IPAddress  string
	UserAgent  string
	CreatedAt  sql.NullTime
	LastSeenAt sql.NullTime
	ExpiresAt  sql.NullTime
}
