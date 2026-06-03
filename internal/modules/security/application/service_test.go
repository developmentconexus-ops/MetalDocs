package application

import (
	"context"
	"errors"
	"testing"
	"time"

	securitydomain "metaldocs/internal/modules/security/domain"
)

type fakeRepo struct {
	mfa            map[string]securitydomain.MfaCoverage
	lockouts       map[string][]securitydomain.Lockout
	failedByUser   map[string]map[string]securitydomain.RecentFailureSummary
	lockoutCount   map[string]int
	newDeviceLogin map[string][]securitydomain.NewDeviceLogin
	offHours       map[string][]securitydomain.OffHoursAction
}

func (r *fakeRepo) MfaCoverage(_ context.Context, t string) (securitydomain.MfaCoverage, error) {
	return r.mfa[t], nil
}
func (r *fakeRepo) ListLockouts(_ context.Context, t string) ([]securitydomain.Lockout, error) {
	return r.lockouts[t], nil
}
func (r *fakeRepo) CountRecentFailedLoginsByUser(_ context.Context, t string, _, _ int) (map[string]securitydomain.RecentFailureSummary, error) {
	return r.failedByUser[t], nil
}
func (r *fakeRepo) CountRecentLockouts(_ context.Context, t string, _ int) (int, error) {
	return r.lockoutCount[t], nil
}
func (r *fakeRepo) ListNewDeviceLogins(_ context.Context, t string, _, _ int) ([]securitydomain.NewDeviceLogin, error) {
	return r.newDeviceLogin[t], nil
}
func (r *fakeRepo) ListOffHoursAdminActions(_ context.Context, t string, _ int, _ []string, _, _ int) ([]securitydomain.OffHoursAction, error) {
	return r.offHours[t], nil
}

func TestService_TenantRequired(t *testing.T) {
	svc := NewService(&fakeRepo{})
	if _, err := svc.MfaCoverage(context.Background(), ""); !errors.Is(err, ErrTenantRequired) {
		t.Fatalf("MfaCoverage empty tenant: want ErrTenantRequired, got %v", err)
	}
	if _, err := svc.ListLockouts(context.Background(), ""); !errors.Is(err, ErrTenantRequired) {
		t.Fatalf("ListLockouts empty tenant: want ErrTenantRequired, got %v", err)
	}
	if _, err := svc.ListSignals(context.Background(), ""); !errors.Is(err, ErrTenantRequired) {
		t.Fatalf("ListSignals empty tenant: want ErrTenantRequired, got %v", err)
	}
}

func TestService_ListSignals_RuleFires(t *testing.T) {
	repo := &fakeRepo{
		failedByUser: map[string]map[string]securitydomain.RecentFailureSummary{
			"tenant-a": {
				"user-1": {UserID: "user-1", DisplayName: "Alice", FailCount: 7, LastFailedAt: "2026-06-02T10:00:00Z"},
			},
		},
		lockoutCount: map[string]int{"tenant-a": 5},
		newDeviceLogin: map[string][]securitydomain.NewDeviceLogin{
			"tenant-a": {{SessionID: "sess-1", UserID: "user-1", DisplayName: "Alice", UserAgent: "Mozilla/5.0", CreatedAt: "2026-06-02T09:00:00Z"}},
		},
		offHours: map[string][]securitydomain.OffHoursAction{
			"tenant-a": {{EventID: "evt-1", ActorID: "admin-1", ActorRole: "system_admin", Action: "iam.user.role.upserted", ResourceType: "user", ResourceID: "user-2", OccurredAt: "2026-06-02T23:30:00Z"}},
		},
	}
	svc := NewService(repo)
	signals, err := svc.ListSignals(context.Background(), "tenant-a")
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	kinds := map[string]int{}
	for _, s := range signals {
		kinds[s.Kind]++
	}
	want := []string{"repeated-failed-login", "lockout-spike", "new-device-login", "off-hours-admin-action"}
	for _, k := range want {
		if kinds[k] == 0 {
			t.Errorf("expected signal kind %q to fire, got %v", k, kinds)
		}
	}
	// Severity ordering: warn (lockout-spike, repeated-failed-login) before info.
	if len(signals) >= 2 {
		first := severityRank(signals[0].Severity)
		last := severityRank(signals[len(signals)-1].Severity)
		if first < last {
			t.Errorf("signals not severity-sorted: %v -> %v", signals[0].Severity, signals[len(signals)-1].Severity)
		}
	}
}

func TestService_ListSignals_TenantIsolation(t *testing.T) {
	// Tenant A has firing signals; tenant B shares no data. The service
	// must not return tenant A's signals when asked about tenant B.
	repo := &fakeRepo{
		failedByUser: map[string]map[string]securitydomain.RecentFailureSummary{
			"tenant-a": {"user-1": {UserID: "user-1", DisplayName: "Alice", FailCount: 9, LastFailedAt: "2026-06-02T10:00:00Z"}},
		},
		lockoutCount: map[string]int{"tenant-a": 10},
		newDeviceLogin: map[string][]securitydomain.NewDeviceLogin{
			"tenant-a": {{SessionID: "sess-1", UserID: "user-1", DisplayName: "Alice", UserAgent: "x", CreatedAt: "now"}},
		},
		offHours: map[string][]securitydomain.OffHoursAction{
			"tenant-a": {{EventID: "evt-1", ActorID: "admin-1", ActorRole: "system_admin", Action: "iam.user.role.upserted", ResourceType: "user", ResourceID: "user-2", OccurredAt: "23:30"}},
		},
	}
	svc := NewService(repo)
	got, err := svc.ListSignals(context.Background(), "tenant-b")
	if err != nil {
		t.Fatalf("ListSignals tenant-b: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("tenant isolation broken: tenant-b saw %d signals from tenant-a", len(got))
	}
}

func TestService_LockoutSpike_BelowThreshold(t *testing.T) {
	repo := &fakeRepo{lockoutCount: map[string]int{"tenant-a": LockoutSpikeThreshold - 1}}
	svc := NewService(repo)
	signals, err := svc.ListSignals(context.Background(), "tenant-a")
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	for _, s := range signals {
		if s.Kind == "lockout-spike" {
			t.Fatalf("lockout-spike fired below threshold")
		}
	}
}

func TestService_StableSignalIDDeterministic(t *testing.T) {
	a := stableID("kind", "tenant", "evidence")
	b := stableID("kind", "tenant", "evidence")
	if a != b {
		t.Fatalf("stableID not deterministic: %s vs %s", a, b)
	}
	c := stableID("kind", "tenant", "other")
	if a == c {
		t.Fatalf("stableID collision across different evidence: %s", a)
	}
}

var _ = time.Second // keep time import in case future tests use it
