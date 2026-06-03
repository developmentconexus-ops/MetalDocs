package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// SessionListItem is the row shape returned to the Admin Center
// Sessions & Security tab. Display name is joined from metaldocs.iam_users
// so the UI can render a recognisable label without a second round-trip.
//
// Tenant scoping is enforced inside ListActiveSessions: every query is
// filtered on s.tenant_id = $1 + the JOIN matches iam_users on the same
// (user_id, tenant_id) tuple, so a session belonging to another tenant is
// not visible even if its user_id collides.
type SessionListItem struct {
	SessionID   string
	UserID      string
	DisplayName string
	IPAddress   string
	UserAgent   string
	CreatedAt   sql.NullTime
	LastSeenAt  sql.NullTime
	ExpiresAt   sql.NullTime
}

// SessionAdminQuery is the parameter object for ListActiveSessions.
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

// ListActiveSessions returns sessions in the caller's tenant, ordered by
// last_seen_at DESC. Cursor pagination is intentionally deferred until the
// row count justifies the keyset complexity (see openapi.yaml `cursor`
// param — the handler returns CursorPage{HasMore:false} for now).
func (r *Repository) ListActiveSessions(ctx context.Context, q SessionAdminQuery) ([]SessionListItem, error) {
	tenantID := strings.TrimSpace(q.TenantID)
	if tenantID == "" {
		return nil, fmt.Errorf("list active sessions: tenant id required")
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	args := []any{tenantID, q.UserID, q.IncludeRevoked, limit}
	const query = `
SELECT s.session_id,
       s.user_id,
       COALESCE(NULLIF(u.display_name, ''), s.user_id) AS display_name,
       COALESCE(s.ip_address, '') AS ip_address,
       COALESCE(s.user_agent, '') AS user_agent,
       s.created_at,
       s.last_seen_at,
       s.expires_at
FROM metaldocs.auth_sessions s
JOIN metaldocs.iam_users u
  ON u.user_id   = s.user_id
 AND u.tenant_id = s.tenant_id
WHERE s.tenant_id = $1::uuid
  AND ($2 = '' OR s.user_id = $2)
  AND ($3 OR (s.revoked_at IS NULL AND s.expires_at > NOW()))
ORDER BY s.last_seen_at DESC, s.session_id
LIMIT $4
`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list active sessions: %w", err)
	}
	defer rows.Close()

	out := make([]SessionListItem, 0, limit)
	for rows.Next() {
		var item SessionListItem
		if err := rows.Scan(
			&item.SessionID,
			&item.UserID,
			&item.DisplayName,
			&item.IPAddress,
			&item.UserAgent,
			&item.CreatedAt,
			&item.LastSeenAt,
			&item.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("scan active session: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active sessions: %w", err)
	}
	return out, nil
}
