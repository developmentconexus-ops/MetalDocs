package domain

import "context"

// Repository is the read-only port the Service uses to power the
// Sessions & Security admin tab. All calls are tenant-scoped via tenantID;
// implementations must NEVER return a row outside the requested tenant.
type Repository interface {
	MfaCoverage(ctx context.Context, tenantID string) (MfaCoverage, error)
	ListLockouts(ctx context.Context, tenantID string) ([]Lockout, error)

	// CountRecentFailedLoginsByUser returns user_id → fail count for failed
	// logins inside the window. Drives the repeated-failed-login signal.
	CountRecentFailedLoginsByUser(ctx context.Context, tenantID string, withinSeconds int, minThreshold int) (map[string]RecentFailureSummary, error)

	// CountRecentLockouts returns the number of accounts in the tenant
	// whose locked_until was set within the last windowSeconds. Drives the
	// lockout-spike signal.
	CountRecentLockouts(ctx context.Context, tenantID string, withinSeconds int) (int, error)

	// ListNewDeviceLogins returns sessions created within the last
	// windowSeconds where the (user_id, user_agent) pair has never been
	// observed in the prior lookbackSeconds. Drives the new-device-login
	// signal.
	ListNewDeviceLogins(ctx context.Context, tenantID string, windowSeconds int, lookbackSeconds int) ([]NewDeviceLogin, error)

	// ListOffHoursAdminActions returns audit events authored by admin-role
	// actors between offHoursStartHour and offHoursEndHour (server-local
	// time, UTC for v1) within the last windowSeconds.
	ListOffHoursAdminActions(ctx context.Context, tenantID string, windowSeconds int, adminRoles []string, offHoursStartHour, offHoursEndHour int) ([]OffHoursAction, error)
}

// RecentFailureSummary carries enough context to render a single
// repeated-failed-login card without a second query.
type RecentFailureSummary struct {
	UserID       string
	DisplayName  string
	FailCount    int
	LastFailedAt string
}

type NewDeviceLogin struct {
	SessionID   string
	UserID      string
	DisplayName string
	UserAgent   string
	CreatedAt   string
}

type OffHoursAction struct {
	EventID      string
	ActorID      string
	ActorRole    string
	Action       string
	ResourceType string
	ResourceID   string
	OccurredAt   string
}
