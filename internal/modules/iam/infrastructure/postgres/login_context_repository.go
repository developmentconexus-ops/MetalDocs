package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// LoginContextRepository implements iamdomain.LoginContextPort. It updates the
// governance columns on metaldocs.iam_users that the People-tab "Last login"
// drawer reads. A missing row (user has no tenant binding yet) is treated as a
// no-op — no row means the UPDATE affects zero rows, which is not an error.
type LoginContextRepository struct {
	db *sql.DB
}

func NewLoginContextRepository(db *sql.DB) *LoginContextRepository {
	return &LoginContextRepository{db: db}
}

// RecordLoginContext updates metaldocs.iam_users with the IP, User-Agent, and
// optional device label observed on the most recent successful login. Length
// truncation matches the values stored alongside the user's active sessions
// (128 for IP/device label, 512 for User-Agent).
func (r *LoginContextRepository) RecordLoginContext(ctx context.Context, userID, tenantID, ip, userAgent, deviceLabel string) error {
	userID = strings.TrimSpace(userID)
	tenantID = strings.TrimSpace(tenantID)
	if userID == "" || tenantID == "" {
		return nil
	}
	const q = `
UPDATE metaldocs.iam_users
SET last_login_ip = NULLIF($3, ''),
    last_login_user_agent = NULLIF($4, ''),
    last_login_device_label = COALESCE(NULLIF($5, ''), last_login_device_label),
    updated_at = NOW()
WHERE user_id = $1
  AND tenant_id = $2::uuid
`
	if _, err := r.db.ExecContext(ctx, q,
		userID,
		tenantID,
		truncateLoginContextField(ip, 128),
		truncateLoginContextField(userAgent, 512),
		truncateLoginContextField(deviceLabel, 128),
	); err != nil {
		return fmt.Errorf("record login context: %w", err)
	}
	return nil
}

func truncateLoginContextField(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}
