// Package postgres backs the Sessions & Security service with tenant-scoped
// SQL queries. Every query takes tenantID as the FIRST positional parameter
// — that contract is enforced visually by reading top-to-bottom (and by the
// unit tests in the security package, which seed two tenants and assert
// isolation).
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	securitydomain "metaldocs/internal/modules/security/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) MfaCoverage(ctx context.Context, tenantID string) (securitydomain.MfaCoverage, error) {
	const totalQ = `
SELECT COUNT(*) FILTER (WHERE u.deactivated_at IS NULL)                       AS total,
       COUNT(*) FILTER (WHERE u.deactivated_at IS NULL AND u.mfa_enabled)     AS enabled
FROM metaldocs.iam_users u
WHERE u.tenant_id = $1::uuid
`
	var total, enabled int
	if err := r.db.QueryRowContext(ctx, totalQ, tenantID).Scan(&total, &enabled); err != nil {
		return securitydomain.MfaCoverage{}, fmt.Errorf("mfa coverage total: %w", err)
	}

	const byRoleQ = `
SELECT ur.role_code,
       COUNT(*)                                       AS total,
       COUNT(*) FILTER (WHERE u.mfa_enabled)          AS enabled
FROM metaldocs.iam_user_roles ur
JOIN metaldocs.iam_users u
  ON u.user_id   = ur.user_id
 AND u.tenant_id = ur.tenant_id
WHERE ur.tenant_id = $1::uuid
  AND u.deactivated_at IS NULL
GROUP BY ur.role_code
ORDER BY ur.role_code
`
	rows, err := r.db.QueryContext(ctx, byRoleQ, tenantID)
	if err != nil {
		return securitydomain.MfaCoverage{}, fmt.Errorf("mfa coverage by role: %w", err)
	}
	defer rows.Close()
	var slices []securitydomain.MfaRoleSlice
	for rows.Next() {
		var s securitydomain.MfaRoleSlice
		if err := rows.Scan(&s.Role, &s.Total, &s.MfaEnabled); err != nil {
			return securitydomain.MfaCoverage{}, fmt.Errorf("scan mfa role slice: %w", err)
		}
		if s.Total > 0 {
			s.Pct = float32(s.MfaEnabled) * 100 / float32(s.Total)
		}
		slices = append(slices, s)
	}
	if err := rows.Err(); err != nil {
		return securitydomain.MfaCoverage{}, fmt.Errorf("iterate mfa role slices: %w", err)
	}

	cov := securitydomain.MfaCoverage{
		TotalUsers: total,
		MfaEnabled: enabled,
		ByRole:     slices,
	}
	if total > 0 {
		cov.MfaEnabledPct = float32(enabled) * 100 / float32(total)
	}
	return cov, nil
}

func (r *Repository) ListLockouts(ctx context.Context, tenantID string) ([]securitydomain.Lockout, error) {
	// JOIN iam_users.tenant_id binds the lockout to the caller's tenant —
	// auth_identities is a global-PK table (no tenant_id column), so the
	// JOIN is how we get tenant scoping.
	const q = `
SELECT i.user_id,
       COALESCE(NULLIF(u.display_name, ''), i.user_id) AS display_name,
       i.failed_login_attempts,
       i.locked_until,
       i.last_failed_login_at,
       COALESCE(i.last_failed_login_ip, '')
FROM metaldocs.auth_identities i
JOIN metaldocs.iam_users u
  ON u.user_id = i.user_id
WHERE u.tenant_id = $1::uuid
  AND i.locked_until IS NOT NULL
  AND i.locked_until > NOW()
ORDER BY i.locked_until ASC
LIMIT 100
`
	rows, err := r.db.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list lockouts: %w", err)
	}
	defer rows.Close()
	var out []securitydomain.Lockout
	for rows.Next() {
		var l securitydomain.Lockout
		var lockedUntil, lastFailedAt sql.NullTime
		if err := rows.Scan(&l.UserID, &l.DisplayName, &l.FailedAttempts, &lockedUntil, &lastFailedAt, &l.LastFailedIP); err != nil {
			return nil, fmt.Errorf("scan lockout: %w", err)
		}
		if lockedUntil.Valid {
			t := lockedUntil.Time.UTC()
			l.LockedUntil = &t
		}
		if lastFailedAt.Valid {
			t := lastFailedAt.Time.UTC()
			l.LastFailedAt = &t
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) CountRecentFailedLoginsByUser(ctx context.Context, tenantID string, withinSeconds int, minThreshold int) (map[string]securitydomain.RecentFailureSummary, error) {
	// Proxy for "N failures within window": auth_identities only stores the
	// LAST failure timestamp + a monotonic counter that RecordSuccessfulLogin
	// zeros. So "≥ threshold attempts AND last_failed_at within window AND
	// no successful login since" is the closest expression of "N recent
	// failures" we can make without a per-attempt log table (which is a
	// separate scope — see wiki/modules/security-tech-debt.md).
	const q = `
SELECT i.user_id,
       COALESCE(NULLIF(u.display_name, ''), i.user_id) AS display_name,
       i.failed_login_attempts,
       i.last_failed_login_at
FROM metaldocs.auth_identities i
JOIN metaldocs.iam_users u ON u.user_id = i.user_id
WHERE u.tenant_id = $1::uuid
  AND i.failed_login_attempts >= $2
  AND i.last_failed_login_at IS NOT NULL
  AND i.last_failed_login_at >= NOW() - ($3 * INTERVAL '1 second')
LIMIT 100
`
	rows, err := r.db.QueryContext(ctx, q, tenantID, minThreshold, withinSeconds)
	if err != nil {
		return nil, fmt.Errorf("recent failed logins: %w", err)
	}
	defer rows.Close()
	out := map[string]securitydomain.RecentFailureSummary{}
	for rows.Next() {
		var s securitydomain.RecentFailureSummary
		var last sql.NullTime
		if err := rows.Scan(&s.UserID, &s.DisplayName, &s.FailCount, &last); err != nil {
			return nil, fmt.Errorf("scan recent fail: %w", err)
		}
		if last.Valid {
			s.LastFailedAt = last.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		out[s.UserID] = s
	}
	return out, rows.Err()
}

func (r *Repository) CountRecentLockouts(ctx context.Context, tenantID string, withinSeconds int) (int, error) {
	const q = `
SELECT COUNT(*)
FROM metaldocs.auth_identities i
JOIN metaldocs.iam_users u ON u.user_id = i.user_id
WHERE u.tenant_id = $1::uuid
  AND i.locked_until IS NOT NULL
  AND i.locked_until >= NOW() - ($2 * INTERVAL '1 second')
`
	var count int
	if err := r.db.QueryRowContext(ctx, q, tenantID, withinSeconds).Scan(&count); err != nil {
		return 0, fmt.Errorf("recent lockouts: %w", err)
	}
	return count, nil
}

func (r *Repository) ListNewDeviceLogins(ctx context.Context, tenantID string, windowSeconds int, lookbackSeconds int) ([]securitydomain.NewDeviceLogin, error) {
	// A "new device" = a session whose (user_id, user_agent) pair has not
	// appeared in any earlier session in the lookback window. SQL stays
	// honest by anti-joining against the same auth_sessions table.
	const q = `
SELECT s.session_id,
       s.user_id,
       COALESCE(NULLIF(u.display_name, ''), s.user_id) AS display_name,
       COALESCE(s.user_agent, '') AS user_agent,
       s.created_at
FROM metaldocs.auth_sessions s
JOIN metaldocs.iam_users u
  ON u.user_id   = s.user_id
 AND u.tenant_id = s.tenant_id
WHERE s.tenant_id = $1::uuid
  AND s.user_agent IS NOT NULL
  AND s.user_agent <> ''
  AND s.created_at >= NOW() - ($2 * INTERVAL '1 second')
  AND NOT EXISTS (
    SELECT 1
    FROM metaldocs.auth_sessions prior
    WHERE prior.tenant_id  = s.tenant_id
      AND prior.user_id    = s.user_id
      AND prior.user_agent = s.user_agent
      AND prior.session_id <> s.session_id
      AND prior.created_at >= NOW() - ($3 * INTERVAL '1 second')
      AND prior.created_at <  s.created_at
  )
ORDER BY s.created_at DESC
LIMIT 50
`
	rows, err := r.db.QueryContext(ctx, q, tenantID, windowSeconds, lookbackSeconds)
	if err != nil {
		return nil, fmt.Errorf("new device logins: %w", err)
	}
	defer rows.Close()
	var out []securitydomain.NewDeviceLogin
	for rows.Next() {
		var d securitydomain.NewDeviceLogin
		var createdAt sql.NullTime
		if err := rows.Scan(&d.SessionID, &d.UserID, &d.DisplayName, &d.UserAgent, &createdAt); err != nil {
			return nil, fmt.Errorf("scan new device: %w", err)
		}
		if createdAt.Valid {
			d.CreatedAt = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) ListOffHoursAdminActions(ctx context.Context, tenantID string, windowSeconds int, adminRoles []string, offHoursStartHour, offHoursEndHour int) ([]securitydomain.OffHoursAction, error) {
	if len(adminRoles) == 0 {
		return nil, nil
	}
	// Off-hours predicate: hour in [start..24) OR [0..end). UTC for v1; once
	// per-tenant timezone is wired (see iam_tenants.timezone proposal in
	// security-tech-debt.md) replace EXTRACT(hour FROM occurred_at) with
	// EXTRACT(hour FROM occurred_at AT TIME ZONE tz).
	const q = `
SELECT e.id,
       e.actor_id,
       MIN(ur.role_code)    AS role_code,
       e.action,
       e.resource_type,
       e.resource_id,
       e.occurred_at
FROM metaldocs.audit_events e
JOIN metaldocs.iam_user_roles ur
  ON ur.user_id   = e.actor_id
 AND ur.tenant_id = e.tenant_id::uuid
WHERE e.tenant_id = $1
  AND e.occurred_at >= NOW() - ($2 * INTERVAL '1 second')
  AND ur.role_code = ANY($3)
  AND (
    EXTRACT(HOUR FROM e.occurred_at AT TIME ZONE 'UTC') >= $4
    OR EXTRACT(HOUR FROM e.occurred_at AT TIME ZONE 'UTC') <  $5
  )
GROUP BY e.id, e.actor_id, e.action, e.resource_type, e.resource_id, e.occurred_at
ORDER BY e.occurred_at DESC
LIMIT 50
`
	rows, err := r.db.QueryContext(ctx, q, strings.TrimSpace(tenantID), windowSeconds, pq.Array(adminRoles), offHoursStartHour, offHoursEndHour)
	if err != nil {
		return nil, fmt.Errorf("off hours admin actions: %w", err)
	}
	defer rows.Close()
	var out []securitydomain.OffHoursAction
	for rows.Next() {
		var a securitydomain.OffHoursAction
		var occurred sql.NullTime
		if err := rows.Scan(&a.EventID, &a.ActorID, &a.ActorRole, &a.Action, &a.ResourceType, &a.ResourceID, &occurred); err != nil {
			return nil, fmt.Errorf("scan off-hours action: %w", err)
		}
		if occurred.Valid {
			a.OccurredAt = occurred.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
