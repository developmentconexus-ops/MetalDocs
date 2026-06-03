// Package domain holds the value types returned by the Sessions & Security
// admin tab (PR-7). All shapes are tenant-scoped — the service + repository
// layers refuse to query without a tenant id.
package domain

import "time"

// MfaCoverage summarises MFA enrollment for a single tenant.
//
// Coverage % is computed against active (non-deactivated) users. Roles with
// zero members are omitted from ByRole — empty rows would be visually noisy
// in the Admin Center cards.
type MfaCoverage struct {
	TotalUsers    int
	MfaEnabled    int
	MfaEnabledPct float32
	ByRole        []MfaRoleSlice
}

type MfaRoleSlice struct {
	Role       string
	Total      int
	MfaEnabled int
	Pct        float32
}

// Lockout is one currently-locked account inside the caller's tenant. It is
// populated from a JOIN of auth_identities + iam_users.
type Lockout struct {
	UserID         string
	DisplayName    string
	FailedAttempts int
	LockedUntil    *time.Time
	LastFailedAt   *time.Time
	LastFailedIP   string
}

// Signal is a rule-based anomaly card. Severity vocabulary is shared with the
// OpenAPI SecuritySignalSeverity enum (info | warn | critical). Kinds map 1:1
// to the SecuritySignalKind enum.
type Signal struct {
	SignalID   string
	Kind       string
	Severity   string
	Summary    string
	Evidence   map[string]any
	DetectedAt time.Time
}
