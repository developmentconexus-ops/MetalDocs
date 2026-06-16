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

	iamdomain "metaldocs/internal/modules/iam/domain"
	securitydomain "metaldocs/internal/modules/security/domain"
)

type Repository struct {
	db *sql.DB
	// displayNames, members, adminRoles, and mfaUsers are iam-owned ports. Security
	// reports on auth_identities / auth_sessions but does NOT own metaldocs.iam_users
	// or iam_user_roles; it resolves tenant membership, display names, admin-role
	// membership, and MFA coverage through these ports. All read the pool (off-tx, H-PRE-1).
	displayNames iamdomain.UserDisplayNameReader
	members      iamdomain.TenantUserReader
	adminRoles   iamdomain.AdminRoleMemberReader
	mfaUsers     iamdomain.MfaUserReader
}

func NewRepository(db *sql.DB, displayNames iamdomain.UserDisplayNameReader, members iamdomain.TenantUserReader, adminRoles iamdomain.AdminRoleMemberReader, mfaUsers iamdomain.MfaUserReader) *Repository {
	if displayNames == nil {
		displayNames = iamdomain.NoopUserDisplayNameReader{}
	}
	if members == nil {
		members = iamdomain.NoopTenantUserReader{}
	}
	if adminRoles == nil {
		adminRoles = iamdomain.NoopAdminRoleMemberReader{}
	}
	if mfaUsers == nil {
		mfaUsers = iamdomain.NoopMfaUserReader{}
	}
	return &Repository{db: db, displayNames: displayNames, members: members, adminRoles: adminRoles, mfaUsers: mfaUsers}
}

// resolveNames fetches display names for the given user ids via the iam port and
// applies the COALESCE(NULLIF(display_name,''), user_id) fallback consumer-side:
// any id the port omits (absent in tenant or empty display_name) maps to itself.
func (r *Repository) resolveNames(ctx context.Context, tenantID string, ids []string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		out[id] = id // default fallback
	}
	if len(ids) == 0 {
		return out, nil
	}
	names, err := r.displayNames.DisplayNames(ctx, tenantID, ids)
	if err != nil {
		return nil, fmt.Errorf("resolve display names: %w", err)
	}
	for id, name := range names {
		if name != "" {
			out[id] = name
		}
	}
	return out, nil
}

func (r *Repository) MfaCoverage(ctx context.Context, tenantID string) (securitydomain.MfaCoverage, error) {
	// Resolve MFA counts via IAM port (off-tx, H-PRE-1).
	total, enabled, err := r.mfaUsers.TenantMfaCounts(ctx, tenantID)
	if err != nil {
		return securitydomain.MfaCoverage{}, fmt.Errorf("mfa coverage total: %w", err)
	}

	roleCounts, err := r.mfaUsers.TenantMfaCountsByRole(ctx, tenantID)
	if err != nil {
		return securitydomain.MfaCoverage{}, fmt.Errorf("mfa coverage by role: %w", err)
	}

	slices := make([]securitydomain.MfaRoleSlice, 0, len(roleCounts))
	for _, rc := range roleCounts {
		s := securitydomain.MfaRoleSlice{Role: rc.Role, Total: rc.Total, MfaEnabled: rc.MfaEnabled}
		if s.Total > 0 {
			s.Pct = float32(s.MfaEnabled) * 100 / float32(s.Total)
		}
		slices = append(slices, s)
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
	// auth_identities is a global-PK table (no tenant_id column); tenant scope
	// comes from the iam-owned membership port (the user_id set with an iam_users
	// row in this tenant) instead of a cross-module JOIN. Display names from the
	// iam display-name port. (M4/F4.6)
	memberIDs, err := r.members.TenantUserIDs(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list lockouts (members): %w", err)
	}
	if len(memberIDs) == 0 {
		return nil, nil
	}
	const q = `
SELECT i.user_id,
       i.failed_login_attempts,
       i.locked_until,
       i.last_failed_login_at,
       COALESCE(i.last_failed_login_ip, '')
FROM metaldocs.auth_identities i
WHERE i.user_id = ANY($1)
  AND i.locked_until IS NOT NULL
  AND i.locked_until > NOW()
ORDER BY i.locked_until ASC
LIMIT 100
`
	rows, err := r.db.QueryContext(ctx, q, pq.Array(memberIDs))
	if err != nil {
		return nil, fmt.Errorf("list lockouts: %w", err)
	}
	defer rows.Close()
	var out []securitydomain.Lockout
	var ids []string
	for rows.Next() {
		var l securitydomain.Lockout
		var lockedUntil, lastFailedAt sql.NullTime
		if err := rows.Scan(&l.UserID, &l.FailedAttempts, &lockedUntil, &lastFailedAt, &l.LastFailedIP); err != nil {
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
		ids = append(ids, l.UserID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	names, err := r.resolveNames(ctx, tenantID, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].DisplayName = names[out[i].UserID]
	}
	return out, nil
}

func (r *Repository) CountRecentFailedLoginsByUser(ctx context.Context, tenantID string, withinSeconds int, minThreshold int) (map[string]securitydomain.RecentFailureSummary, error) {
	// Proxy for "N failures within window": auth_identities only stores the
	// LAST failure timestamp + a monotonic counter that RecordSuccessfulLogin
	// zeros. So "≥ threshold attempts AND last_failed_at within window AND
	// no successful login since" is the closest expression of "N recent
	// failures" we can make without a per-attempt log table (which is a
	// separate scope — see wiki/modules/security-tech-debt.md).
	// Tenant scope via the iam membership port (auth_identities has no tenant_id);
	// display names via the iam display-name port. (M4/F4.6)
	memberIDs, err := r.members.TenantUserIDs(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("recent failed logins (members): %w", err)
	}
	if len(memberIDs) == 0 {
		return map[string]securitydomain.RecentFailureSummary{}, nil
	}
	const q = `
SELECT i.user_id,
       i.failed_login_attempts,
       i.last_failed_login_at
FROM metaldocs.auth_identities i
WHERE i.user_id = ANY($1)
  AND i.failed_login_attempts >= $2
  AND i.last_failed_login_at IS NOT NULL
  AND i.last_failed_login_at >= NOW() - ($3 * INTERVAL '1 second')
LIMIT 100
`
	rows, err := r.db.QueryContext(ctx, q, pq.Array(memberIDs), minThreshold, withinSeconds)
	if err != nil {
		return nil, fmt.Errorf("recent failed logins: %w", err)
	}
	defer rows.Close()
	out := map[string]securitydomain.RecentFailureSummary{}
	var ids []string
	for rows.Next() {
		var s securitydomain.RecentFailureSummary
		var last sql.NullTime
		if err := rows.Scan(&s.UserID, &s.FailCount, &last); err != nil {
			return nil, fmt.Errorf("scan recent fail: %w", err)
		}
		if last.Valid {
			s.LastFailedAt = last.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		out[s.UserID] = s
		ids = append(ids, s.UserID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	names, err := r.resolveNames(ctx, tenantID, ids)
	if err != nil {
		return nil, err
	}
	for id, s := range out {
		s.DisplayName = names[id]
		out[id] = s
	}
	return out, nil
}

func (r *Repository) CountRecentLockouts(ctx context.Context, tenantID string, withinSeconds int) (int, error) {
	// Tenant scope via the iam membership port (no iam_users JOIN). (M4/F4.6)
	memberIDs, err := r.members.TenantUserIDs(ctx, tenantID)
	if err != nil {
		return 0, fmt.Errorf("recent lockouts (members): %w", err)
	}
	if len(memberIDs) == 0 {
		return 0, nil
	}
	const q = `
SELECT COUNT(*)
FROM metaldocs.auth_identities i
WHERE i.user_id = ANY($1)
  AND i.locked_until IS NOT NULL
  AND i.locked_until >= NOW() - ($2 * INTERVAL '1 second')
`
	var count int
	if err := r.db.QueryRowContext(ctx, q, pq.Array(memberIDs), withinSeconds).Scan(&count); err != nil {
		return 0, fmt.Errorf("recent lockouts: %w", err)
	}
	return count, nil
}

func (r *Repository) ListNewDeviceLogins(ctx context.Context, tenantID string, windowSeconds int, lookbackSeconds int) ([]securitydomain.NewDeviceLogin, error) {
	// A "new device" = a session whose (user_id, user_agent) pair has not
	// appeared in any earlier session in the lookback window. SQL stays
	// honest by anti-joining against the same auth_sessions table.
	// auth_sessions HAS tenant_id, so tenant scope is the owned s.tenant_id column
	// directly (no iam_users JOIN); display names via the iam port. The old INNER
	// JOIN also dropped sessions whose user lacks an iam_users membership row, but
	// a session in tenant T implies a login to T (⇒ membership), so no such orphan
	// rows occur on the real path (documented note, F4.4/F4.6). (M4/F4.6)
	const q = `
SELECT s.session_id,
       s.user_id,
       COALESCE(s.user_agent, '') AS user_agent,
       s.created_at
FROM metaldocs.auth_sessions s
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
	var ids []string
	for rows.Next() {
		var d securitydomain.NewDeviceLogin
		var createdAt sql.NullTime
		if err := rows.Scan(&d.SessionID, &d.UserID, &d.UserAgent, &createdAt); err != nil {
			return nil, fmt.Errorf("scan new device: %w", err)
		}
		if createdAt.Valid {
			d.CreatedAt = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		out = append(out, d)
		ids = append(ids, d.UserID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	names, err := r.resolveNames(ctx, tenantID, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].DisplayName = names[out[i].UserID]
	}
	return out, nil
}

func (r *Repository) ListOffHoursAdminActions(ctx context.Context, tenantID string, windowSeconds int, adminRoles []string, offHoursStartHour, offHoursEndHour int) ([]securitydomain.OffHoursAction, error) {
	if len(adminRoles) == 0 {
		return nil, nil
	}
	// Resolve admin-role membership via IAM port (off-tx, H-PRE-1).
	userRoles, err := r.adminRoles.AdminRoleMembers(ctx, tenantID, adminRoles)
	if err != nil {
		return nil, fmt.Errorf("off-hours-admin: resolve role members: %w", err)
	}
	if len(userRoles) == 0 {
		return nil, nil
	}
	userIDs := make([]string, 0, len(userRoles))
	for uid := range userRoles {
		userIDs = append(userIDs, uid)
	}

	// Off-hours predicate: hour in [start..24) OR [0..end). UTC for v1; once
	// per-tenant timezone is wired (see iam_tenants.timezone proposal in
	// security-tech-debt.md) replace EXTRACT(hour FROM occurred_at) with
	// EXTRACT(hour FROM occurred_at AT TIME ZONE tz).
	const q = `
SELECT e.id,
       e.actor_id,
       e.action,
       e.resource_type,
       e.resource_id,
       e.occurred_at
FROM metaldocs.audit_events e
WHERE e.tenant_id = $1
  AND e.actor_id = ANY($2)
  AND e.occurred_at >= NOW() - ($3 * INTERVAL '1 second')
  AND (
    EXTRACT(HOUR FROM e.occurred_at AT TIME ZONE 'UTC') >= $4
    OR EXTRACT(HOUR FROM e.occurred_at AT TIME ZONE 'UTC') <  $5
  )
ORDER BY e.occurred_at DESC
LIMIT 50
`
	rows, err := r.db.QueryContext(ctx, q, strings.TrimSpace(tenantID), pq.Array(userIDs), windowSeconds, offHoursStartHour, offHoursEndHour)
	if err != nil {
		return nil, fmt.Errorf("off hours admin actions: %w", err)
	}
	defer rows.Close()
	var out []securitydomain.OffHoursAction
	for rows.Next() {
		var a securitydomain.OffHoursAction
		var occurred sql.NullTime
		if err := rows.Scan(&a.EventID, &a.ActorID, &a.Action, &a.ResourceType, &a.ResourceID, &occurred); err != nil {
			return nil, fmt.Errorf("scan off-hours action: %w", err)
		}
		if occurred.Valid {
			a.OccurredAt = occurred.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		a.ActorRole = userRoles[a.ActorID]
		out = append(out, a)
	}
	return out, rows.Err()
}
