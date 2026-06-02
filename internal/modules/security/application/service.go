// Package application is the orchestration layer for the Sessions & Security
// admin tab (PR-7). Every method enforces tenant isolation: callers must pass
// the tenant id resolved from tenant.FromContext on the request — the service
// refuses an empty tenant.
//
// Signal queries live here (rather than in a background job) because the
// windows are short (≤ 24h) and the result set is small. If the workload
// grows enough to justify pre-computation, swap the per-request orchestration
// for a materialised view + cron without changing the HTTP shape.
package application

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	securitydomain "metaldocs/internal/modules/security/domain"
)

// Default tunables. Public so tests can override; not configurable at runtime
// (yet) because the values are tuned to the rule set documented in
// wiki/modules/security-signals.md.
const (
	RepeatedFailedLoginWindowSec = 15 * 60
	RepeatedFailedLoginThreshold = 5

	LockoutSpikeWindowSec = 60 * 60
	LockoutSpikeThreshold = 3

	NewDeviceLoginWindowSec   = 24 * 60 * 60
	NewDeviceLoginLookbackSec = 90 * 24 * 60 * 60

	OffHoursWindowSec    = 24 * 60 * 60
	OffHoursStartHourUTC = 22 // inclusive
	OffHoursEndHourUTC   = 6  // exclusive (i.e. 22:00..06:00 UTC)
)

// AdminRoles is the set of roles whose mutating actions trigger the
// off-hours-admin-action signal. Mirrors the canonical role allowlist
// (qa/iam-admin-center; iamdomain.Role*).
var AdminRoles = []string{"system_admin", "area_admin", "qms_admin"}

type Service struct {
	repo securitydomain.Repository
	now  func() time.Time
}

func NewService(repo securitydomain.Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

// WithClock is a test seam: lets tests pin h.now() so the off-hours rule
// fires deterministically regardless of wall-clock.
func (s *Service) WithClock(now func() time.Time) *Service {
	s.now = now
	return s
}

var ErrTenantRequired = errors.New("security: tenant id required")

func (s *Service) MfaCoverage(ctx context.Context, tenantID string) (securitydomain.MfaCoverage, error) {
	if strings.TrimSpace(tenantID) == "" {
		return securitydomain.MfaCoverage{}, ErrTenantRequired
	}
	return s.repo.MfaCoverage(ctx, tenantID)
}

func (s *Service) ListLockouts(ctx context.Context, tenantID string) ([]securitydomain.Lockout, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantRequired
	}
	return s.repo.ListLockouts(ctx, tenantID)
}

// ListSignals runs every rule once and returns the union of cards. Ordering:
// highest severity first, then most recent detectedAt — matches the UI's
// "what should I look at" reading order.
func (s *Service) ListSignals(ctx context.Context, tenantID string) ([]securitydomain.Signal, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantRequired
	}
	now := s.now()
	var signals []securitydomain.Signal

	repeated, err := s.repo.CountRecentFailedLoginsByUser(ctx, tenantID, RepeatedFailedLoginWindowSec, RepeatedFailedLoginThreshold)
	if err != nil {
		return nil, fmt.Errorf("repeated-failed-login rule: %w", err)
	}
	for userID, summary := range repeated {
		signals = append(signals, securitydomain.Signal{
			SignalID:   stableID("repeated-failed-login", tenantID, userID, summary.LastFailedAt),
			Kind:       "repeated-failed-login",
			Severity:   "warn",
			Summary:    fmt.Sprintf("%s — %d failed logins in the last %d min", summary.DisplayName, summary.FailCount, RepeatedFailedLoginWindowSec/60),
			Evidence:   map[string]any{"userId": summary.UserID, "displayName": summary.DisplayName, "failCount": summary.FailCount, "lastFailedAt": summary.LastFailedAt},
			DetectedAt: now,
		})
	}

	lockouts, err := s.repo.CountRecentLockouts(ctx, tenantID, LockoutSpikeWindowSec)
	if err != nil {
		return nil, fmt.Errorf("lockout-spike rule: %w", err)
	}
	if lockouts >= LockoutSpikeThreshold {
		signals = append(signals, securitydomain.Signal{
			SignalID:   stableID("lockout-spike", tenantID, fmt.Sprintf("%d", lockouts), now.Truncate(time.Hour).Format(time.RFC3339)),
			Kind:       "lockout-spike",
			Severity:   "warn",
			Summary:    fmt.Sprintf("%d accounts locked out in the last %d min", lockouts, LockoutSpikeWindowSec/60),
			Evidence:   map[string]any{"count": lockouts, "windowMinutes": LockoutSpikeWindowSec / 60},
			DetectedAt: now,
		})
	}

	newDevices, err := s.repo.ListNewDeviceLogins(ctx, tenantID, NewDeviceLoginWindowSec, NewDeviceLoginLookbackSec)
	if err != nil {
		return nil, fmt.Errorf("new-device-login rule: %w", err)
	}
	for _, d := range newDevices {
		signals = append(signals, securitydomain.Signal{
			SignalID:   stableID("new-device-login", tenantID, d.SessionID),
			Kind:       "new-device-login",
			Severity:   "info",
			Summary:    fmt.Sprintf("%s logged in from a new device", d.DisplayName),
			Evidence:   map[string]any{"userId": d.UserID, "displayName": d.DisplayName, "userAgent": d.UserAgent, "sessionId": d.SessionID, "createdAt": d.CreatedAt},
			DetectedAt: now,
		})
	}

	offHours, err := s.repo.ListOffHoursAdminActions(ctx, tenantID, OffHoursWindowSec, AdminRoles, OffHoursStartHourUTC, OffHoursEndHourUTC)
	if err != nil {
		return nil, fmt.Errorf("off-hours-admin rule: %w", err)
	}
	for _, e := range offHours {
		signals = append(signals, securitydomain.Signal{
			SignalID:   stableID("off-hours-admin-action", tenantID, e.EventID),
			Kind:       "off-hours-admin-action",
			Severity:   "info",
			Summary:    fmt.Sprintf("%s by %s (%s) outside business hours", e.Action, e.ActorID, e.ActorRole),
			Evidence:   map[string]any{"eventId": e.EventID, "actorId": e.ActorID, "actorRole": e.ActorRole, "action": e.Action, "resourceType": e.ResourceType, "resourceId": e.ResourceID, "occurredAt": e.OccurredAt},
			DetectedAt: now,
		})
	}

	// Sort: severity rank desc, then detectedAt desc.
	sort.SliceStable(signals, func(i, j int) bool {
		ri, rj := severityRank(signals[i].Severity), severityRank(signals[j].Severity)
		if ri != rj {
			return ri > rj
		}
		return signals[i].DetectedAt.After(signals[j].DetectedAt)
	})
	return signals, nil
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 3
	case "warn":
		return 2
	case "info":
		return 1
	}
	return 0
}

// stableID makes the same evidence-tuple produce the same signal ID across
// requests. The UI uses this to de-dupe cards across refreshes and to attach
// per-card dismiss state without persisting it server-side.
func stableID(parts ...string) string {
	h := sha1.New()
	for i, p := range parts {
		if i > 0 {
			_, _ = h.Write([]byte{0})
		}
		_, _ = h.Write([]byte(p))
	}
	return "sig_" + hex.EncodeToString(h.Sum(nil))[:16]
}
