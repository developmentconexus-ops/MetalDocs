// Package postgres backs ObservabilityRepository with tenant-scoped SQL
// queries. Every query takes tenantID as the first positional parameter —
// the same contract used by the security repository.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

type ObservabilityRepository struct {
	db *sql.DB
}

func NewObservabilityRepository(db *sql.DB) *ObservabilityRepository {
	return &ObservabilityRepository{db: db}
}

func (r *ObservabilityRepository) CountActiveSeats(ctx context.Context, tenantID string) (int, error) {
	const q = `
SELECT COUNT(*)
FROM metaldocs.iam_users
WHERE tenant_id = $1::uuid
  AND is_active = true
`
	var n int
	if err := r.db.QueryRowContext(ctx, q, tenantID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count active seats: %w", err)
	}
	return n, nil
}

// StorageUsedBytes returns -1: the v1 surface does not yet aggregate
// blob-layer bytes through a tenant-scoped path. Tracked in
// wiki/modules/iam-tech-debt.md ("storage usage source"). The handler
// renders this verbatim; the FE substitutes "—" for negatives.
func (r *ObservabilityRepository) StorageUsedBytes(ctx context.Context, tenantID string) (int64, error) {
	_ = ctx
	_ = tenantID
	return -1, nil
}

// CountAuditEventsByActionPrefix counts audit_events rows for the tenant
// whose action starts with the prefix, inside the trailing window. An
// empty prefix counts every row (used by the audit-events-per-minute KPI).
//
// The v1 surface does not emit per-request audit rows under the
// "http.request." prefix; in that case this returns 0 (not -1) because a
// "no rows match the prefix" answer is still a correct count. The
// observability service maps the apiCalls field directly from this value,
// and the FE will display 0 instead of "—". Once the request-log source
// lands (tech-debt), the call site can switch to a dedicated reader
// without breaking the wire shape.
func (r *ObservabilityRepository) CountAuditEventsByActionPrefix(ctx context.Context, tenantID, actionPrefix string, withinSeconds int) (int64, error) {
	if withinSeconds <= 0 {
		return 0, fmt.Errorf("audit event count: window must be > 0, got %d", withinSeconds)
	}
	if actionPrefix == "" {
		const q = `
SELECT COUNT(*)
FROM metaldocs.audit_events
WHERE tenant_id = $1
  AND occurred_at > now() - make_interval(secs => $2)
`
		var n int64
		if err := r.db.QueryRowContext(ctx, q, tenantID, withinSeconds).Scan(&n); err != nil {
			return 0, fmt.Errorf("audit event count: %w", err)
		}
		return n, nil
	}
	const q = `
SELECT COUNT(*)
FROM metaldocs.audit_events
WHERE tenant_id = $1
  AND action LIKE $2 || '%'
  AND occurred_at > now() - make_interval(secs => $3)
`
	var n int64
	if err := r.db.QueryRowContext(ctx, q, tenantID, actionPrefix, withinSeconds).Scan(&n); err != nil {
		return 0, fmt.Errorf("audit event count by prefix: %w", err)
	}
	return n, nil
}

func (r *ObservabilityRepository) CountActiveUsersInWindow(ctx context.Context, tenantID string, withinSeconds int) (int64, error) {
	if withinSeconds <= 0 {
		return 0, fmt.Errorf("active users: window must be > 0, got %d", withinSeconds)
	}
	const q = `
SELECT COUNT(DISTINCT actor_id)
FROM metaldocs.audit_events
WHERE tenant_id = $1
  AND occurred_at > now() - make_interval(secs => $2)
  AND actor_id <> ''
  AND actor_id <> 'system'
`
	var n int64
	if err := r.db.QueryRowContext(ctx, q, tenantID, withinSeconds).Scan(&n); err != nil {
		return 0, fmt.Errorf("active users in window: %w", err)
	}
	return n, nil
}

func (r *ObservabilityRepository) GetTenantPlan(ctx context.Context, tenantID string) (iamdomain.TenantPlan, error) {
	const q = `
SELECT tenant_id::text, plan_tier, seats_allocated, storage_allocated_bytes, api_calls_allocated, created_at, updated_at
FROM metaldocs.tenant_plans
WHERE tenant_id = $1::uuid
`
	var p iamdomain.TenantPlan
	var tier string
	err := r.db.QueryRowContext(ctx, q, tenantID).Scan(
		&p.TenantID, &tier, &p.SeatsAllocated, &p.StorageAllocatedBytes, &p.APICallsAllocated, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return iamdomain.TenantPlan{}, iamdomain.ErrTenantPlanNotFound
	}
	if err != nil {
		return iamdomain.TenantPlan{}, fmt.Errorf("get tenant plan: %w", err)
	}
	p.Tier = iamdomain.PlanTier(tier)
	return p, nil
}

// CountLockedAccounts joins auth_identities (where locked_until lives,
// because the credential surface is tenant-agnostic) onto iam_users via
// user_id to apply the tenant filter.
func (r *ObservabilityRepository) CountLockedAccounts(ctx context.Context, tenantID string) (int, error) {
	const q = `
SELECT COUNT(*)
FROM metaldocs.auth_identities i
JOIN metaldocs.iam_users u ON u.user_id = i.user_id
WHERE u.tenant_id = $1::uuid
  AND i.locked_until IS NOT NULL
  AND i.locked_until > now()
`
	var n int
	if err := r.db.QueryRowContext(ctx, q, tenantID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count locked accounts: %w", err)
	}
	return n, nil
}

func (r *ObservabilityRepository) CountFailedLoginsInWindow(ctx context.Context, tenantID string, withinSeconds int) (int64, error) {
	if withinSeconds <= 0 {
		return 0, fmt.Errorf("failed logins: window must be > 0, got %d", withinSeconds)
	}
	const q = `
SELECT COUNT(*)
FROM metaldocs.audit_events
WHERE tenant_id = $1
  AND action = 'auth.login.failed'
  AND occurred_at > now() - make_interval(secs => $2)
`
	var n int64
	if err := r.db.QueryRowContext(ctx, q, tenantID, withinSeconds).Scan(&n); err != nil {
		return 0, fmt.Errorf("count failed logins: %w", err)
	}
	return n, nil
}

// CountDormantUsers reads last-login from auth_identities (the canonical
// credential surface). iam_users does not carry last_login_at directly —
// PR-4 added context columns but not the timestamp itself.
func (r *ObservabilityRepository) CountDormantUsers(ctx context.Context, tenantID string, thresholdDays int) (int, error) {
	if thresholdDays <= 0 {
		return 0, fmt.Errorf("dormant users: threshold must be > 0, got %d", thresholdDays)
	}
	const q = `
SELECT COUNT(*)
FROM metaldocs.iam_users u
LEFT JOIN metaldocs.auth_identities i ON i.user_id = u.user_id
WHERE u.tenant_id = $1::uuid
  AND u.is_active = true
  AND (i.last_login_at IS NULL OR i.last_login_at < now() - make_interval(days => $2))
`
	var n int
	if err := r.db.QueryRowContext(ctx, q, tenantID, thresholdDays).Scan(&n); err != nil {
		return 0, fmt.Errorf("count dormant users: %w", err)
	}
	return n, nil
}

func (r *ObservabilityRepository) RoleDistribution(ctx context.Context, tenantID string) ([]iamdomain.RoleCount, error) {
	const q = `
SELECT role_code, COUNT(*)
FROM metaldocs.iam_user_roles
WHERE tenant_id = $1::uuid
GROUP BY role_code
ORDER BY role_code
`
	rows, err := r.db.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("role distribution query: %w", err)
	}
	defer rows.Close()
	out := make([]iamdomain.RoleCount, 0, 4)
	for rows.Next() {
		var rc iamdomain.RoleCount
		var role string
		if err := rows.Scan(&role, &rc.Count); err != nil {
			return nil, fmt.Errorf("scan role distribution: %w", err)
		}
		rc.Role = iamdomain.Role(role)
		out = append(out, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role distribution: %w", err)
	}
	return out, nil
}
