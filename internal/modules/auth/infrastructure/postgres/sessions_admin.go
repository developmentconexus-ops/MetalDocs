package postgres

import (
	"context"
	"fmt"
	"strings"

	authdomain "metaldocs/internal/modules/auth/domain"
)

// ListActiveSessions returns sessions in the caller's tenant, ordered by
// last_seen_at DESC. Cursor pagination is intentionally deferred until the
// row count justifies the keyset complexity (see openapi.yaml `cursor`
// param — the handler returns CursorPage{HasMore:false} for now).
func (r *Repository) ListActiveSessions(ctx context.Context, q authdomain.SessionAdminQuery) ([]authdomain.SessionListItem, error) {
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

	out := make([]authdomain.SessionListItem, 0, limit)
	for rows.Next() {
		var item authdomain.SessionListItem
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
