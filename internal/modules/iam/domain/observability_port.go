package domain

import (
	"context"
	"errors"
)

// ObservabilityRepository is the read-only port the IAM observability
// service uses to power GET /iam/usage and GET /iam/kpi. Every method is
// tenant-scoped; implementations MUST NEVER return a row outside the
// requested tenant.
//
// Storage and ApiCalls return (-1, nil) when the underlying source is not
// wired yet (see wiki/modules/iam-tech-debt.md). Returning a negative is
// preferable to fabricating a zero because the FE renders "—" instead of a
// misleading "0 of N".
type ObservabilityRepository interface {
	// CountActiveSeats returns COUNT(iam_users WHERE is_active = true).
	CountActiveSeats(ctx context.Context, tenantID string) (int, error)

	// StorageUsedBytes aggregates blob-layer byte size for the tenant.
	// Returns -1 when the aggregation source is not wired.
	StorageUsedBytes(ctx context.Context, tenantID string) (int64, error)

	// CountAuditEventsByActionPrefix counts rows in metaldocs.audit_events
	// for the tenant whose action starts with prefix, within the given
	// trailing window in seconds. An empty prefix counts every row.
	// Returns -1 when neither the action prefix nor a request-log
	// replacement is wired.
	CountAuditEventsByActionPrefix(ctx context.Context, tenantID, actionPrefix string, withinSeconds int) (int64, error)

	// CountActiveUsersInWindow returns COUNT(DISTINCT actor_id) on
	// audit events for the tenant within the trailing window in seconds.
	CountActiveUsersInWindow(ctx context.Context, tenantID string, withinSeconds int) (int64, error)

	// GetTenantPlan returns the plan envelope row for the tenant. Returns
	// (TenantPlan{}, ErrTenantPlanNotFound) when no row exists.
	GetTenantPlan(ctx context.Context, tenantID string) (TenantPlan, error)

	// CountLockedAccounts returns the number of currently-locked accounts
	// inside the tenant (auth_identities.locked_until > now()).
	CountLockedAccounts(ctx context.Context, tenantID string) (int, error)

	// CountFailedLoginsInWindow counts audit_events with action
	// 'auth.login.failed' for the tenant inside the trailing window.
	CountFailedLoginsInWindow(ctx context.Context, tenantID string, withinSeconds int) (int64, error)

	// CountDormantUsers returns iam_users WHERE is_active AND
	// (last_login_at IS NULL OR last_login_at < now() - $thresholdDays * 1 day).
	CountDormantUsers(ctx context.Context, tenantID string, thresholdDays int) (int, error)

	// RoleDistribution returns one bucket per role for the tenant from
	// iam_user_roles. Empty roles are omitted.
	RoleDistribution(ctx context.Context, tenantID string) ([]RoleCount, error)
}

// ErrTenantPlanNotFound is returned by GetTenantPlan when no plan row
// exists for the requested tenant.
var ErrTenantPlanNotFound = errors.New("iam: tenant plan not found")
