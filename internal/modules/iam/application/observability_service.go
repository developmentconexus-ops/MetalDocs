// Package application — ObservabilityService backs the Admin Center
// Overview "Usage" + "KPI" cards. Every method is tenant-scoped via the
// tenantID argument; the service refuses an empty tenant.
//
// MFA coverage is delegated to the security module (PR-7) rather than
// recomputed here so a single source of truth survives. The dependency is
// expressed via a 1-method interface declared in this package — accept
// interfaces, return structs.
package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// MfaCoveragePctReader is the narrow port the observability service uses
// to fold MFA coverage % into the KPI snapshot. The security module
// service exposes a richer shape (per-role slices); the wiring layer
// adapts it down to this single-method port so the iam package does not
// depend on the security domain.
type MfaCoveragePctReader interface {
	MfaCoveragePct(ctx context.Context, tenantID string) (float32, error)
}

// ObservabilityService composes UsageSnapshot + KpiSnapshot for the
// caller's tenant. Constructed once at startup; safe for concurrent use.
type ObservabilityService struct {
	repo iamdomain.ObservabilityRepository
	mfa  MfaCoveragePctReader
	now  func() time.Time
}

// ErrTenantRequired is returned by GetUsage and GetKpi when tenantID is
// empty after trimming.
var ErrTenantRequired = errors.New("iam observability: tenant id required")

// DormantThresholdDays is the cutoff for the "dormant users 30d" KPI. The
// constant is exported so tests and FE copy can stay in sync if v2
// renegotiates the window.
const DormantThresholdDays = 30

// FailedLoginsWindowSec is the trailing window used by the
// failedLogins24h KPI. Exported for the same reason as DormantThresholdDays.
const FailedLoginsWindowSec = 24 * 60 * 60

// AuditEventsRateWindowSec is the trailing window used to derive the
// auditEventsPerMinute KPI. 60 minutes / 60 = events per minute.
const AuditEventsRateWindowSec = 60 * 60

// APICallActionPrefix is the audit-action prefix counted toward the
// apiCalls usage panes. When no audit events use this prefix (the v1
// surface does not emit per-request audit rows), the repository returns
// -1 and the FE renders "—".
const APICallActionPrefix = "http.request."

// NewObservabilityService wires the observability composer. mfa may be nil,
// in which case KpiSnapshot.MfaCoveragePct is left at its zero value.
func NewObservabilityService(repo iamdomain.ObservabilityRepository, mfa MfaCoveragePctReader) *ObservabilityService {
	return &ObservabilityService{
		repo: repo,
		mfa:  mfa,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

// WithClock is a test seam mirroring the convention in security.Service.
func (s *ObservabilityService) WithClock(now func() time.Time) *ObservabilityService {
	s.now = now
	return s
}

// GetUsage composes the Admin Center Overview "Usage" card: seat, storage,
// and windowed API-call/active-user counts for tenantID. Returns
// ErrTenantRequired for an empty tenant. Tenant plan lookup tolerates
// iamdomain.ErrTenantPlanNotFound (falls through with zero allocations
// rather than failing the whole snapshot).
func (s *ObservabilityService) GetUsage(ctx context.Context, tenantID string) (iamdomain.UsageSnapshot, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return iamdomain.UsageSnapshot{}, ErrTenantRequired
	}
	if s.repo == nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("iam observability: repository not configured")
	}

	seatsUsed, err := s.repo.CountActiveSeats(ctx, tenantID)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("count active seats: %w", err)
	}
	storageUsed, err := s.repo.StorageUsedBytes(ctx, tenantID)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("storage used bytes: %w", err)
	}
	apiCalls24h, err := s.repo.CountAuditEventsByActionPrefix(ctx, tenantID, APICallActionPrefix, 24*60*60)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("count api calls 24h: %w", err)
	}
	apiCalls7d, err := s.repo.CountAuditEventsByActionPrefix(ctx, tenantID, APICallActionPrefix, 7*24*60*60)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("count api calls 7d: %w", err)
	}
	apiCalls30d, err := s.repo.CountAuditEventsByActionPrefix(ctx, tenantID, APICallActionPrefix, 30*24*60*60)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("count api calls 30d: %w", err)
	}
	activeUsers24h, err := s.repo.CountActiveUsersInWindow(ctx, tenantID, 24*60*60)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("count active users 24h: %w", err)
	}
	activeUsers7d, err := s.repo.CountActiveUsersInWindow(ctx, tenantID, 7*24*60*60)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("count active users 7d: %w", err)
	}
	activeUsers30d, err := s.repo.CountActiveUsersInWindow(ctx, tenantID, 30*24*60*60)
	if err != nil {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("count active users 30d: %w", err)
	}

	plan, err := s.repo.GetTenantPlan(ctx, tenantID)
	if err != nil && !errors.Is(err, iamdomain.ErrTenantPlanNotFound) {
		return iamdomain.UsageSnapshot{}, fmt.Errorf("get tenant plan: %w", err)
	}

	snapshot := iamdomain.UsageSnapshot{
		Seats: iamdomain.SeatUsage{
			Used:      seatsUsed,
			Allocated: plan.SeatsAllocated,
		},
		Storage: iamdomain.StorageUsage{
			UsedBytes:      storageUsed,
			AllocatedBytes: plan.StorageAllocatedBytes,
		},
		APICalls: iamdomain.CountWindows{
			Last24h: apiCalls24h,
			Last7d:  apiCalls7d,
			Last30d: apiCalls30d,
		},
		ActiveUsers: iamdomain.CountWindows{
			Last24h: activeUsers24h,
			Last7d:  activeUsers7d,
			Last30d: activeUsers30d,
		},
		PlanTier: plan.Tier,
	}
	return snapshot, nil
}

// GetKpi composes the Admin Center Overview "KPI" card: locked accounts,
// failed logins, dormant users, role distribution, audit-event rate, and
// (when the mfa reader is wired) MFA coverage percentage for tenantID.
// Returns ErrTenantRequired for an empty tenant.
func (s *ObservabilityService) GetKpi(ctx context.Context, tenantID string) (iamdomain.KpiSnapshot, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return iamdomain.KpiSnapshot{}, ErrTenantRequired
	}
	if s.repo == nil {
		return iamdomain.KpiSnapshot{}, fmt.Errorf("iam observability: repository not configured")
	}

	locked, err := s.repo.CountLockedAccounts(ctx, tenantID)
	if err != nil {
		return iamdomain.KpiSnapshot{}, fmt.Errorf("count locked accounts: %w", err)
	}
	failed24h, err := s.repo.CountFailedLoginsInWindow(ctx, tenantID, FailedLoginsWindowSec)
	if err != nil {
		return iamdomain.KpiSnapshot{}, fmt.Errorf("count failed logins 24h: %w", err)
	}
	dormant, err := s.repo.CountDormantUsers(ctx, tenantID, DormantThresholdDays)
	if err != nil {
		return iamdomain.KpiSnapshot{}, fmt.Errorf("count dormant users: %w", err)
	}
	dist, err := s.repo.RoleDistribution(ctx, tenantID)
	if err != nil {
		return iamdomain.KpiSnapshot{}, fmt.Errorf("role distribution: %w", err)
	}
	auditCount, err := s.repo.CountAuditEventsByActionPrefix(ctx, tenantID, "", AuditEventsRateWindowSec)
	if err != nil {
		return iamdomain.KpiSnapshot{}, fmt.Errorf("count audit events for rate: %w", err)
	}
	var ratePerMinute float32
	if auditCount > 0 {
		ratePerMinute = float32(auditCount) / 60.0
	}

	var mfaPct float32
	if s.mfa != nil {
		mfaPct, err = s.mfa.MfaCoveragePct(ctx, tenantID)
		if err != nil {
			return iamdomain.KpiSnapshot{}, fmt.Errorf("mfa coverage pct: %w", err)
		}
	}

	return iamdomain.KpiSnapshot{
		LockedAccounts:       locked,
		MfaCoveragePct:       mfaPct,
		FailedLogins24h:      failed24h,
		DormantUsers30d:      dormant,
		RoleDistribution:     dist,
		AuditEventsPerMinute: ratePerMinute,
	}, nil
}
