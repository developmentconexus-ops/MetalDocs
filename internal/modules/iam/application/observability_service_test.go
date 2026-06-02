package application_test

import (
	"context"
	"errors"
	"testing"

	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

type fakeObsRepo struct {
	seats          map[string]int
	storage        map[string]int64
	apiByPrefix    map[string]int64 // tenant|prefix|secs -> count
	active         map[string]int64
	plans          map[string]iamdomain.TenantPlan
	locked         map[string]int
	failed         map[string]int64
	dormant        map[string]int
	roleDist       map[string][]iamdomain.RoleCount
	storageErr     error
	dormantThresh  int
	dormantTenant  string
	lastSeenPrefix string
}

func newFakeObsRepo() *fakeObsRepo {
	return &fakeObsRepo{
		seats:       map[string]int{},
		storage:     map[string]int64{},
		apiByPrefix: map[string]int64{},
		active:      map[string]int64{},
		plans:       map[string]iamdomain.TenantPlan{},
		locked:      map[string]int{},
		failed:      map[string]int64{},
		dormant:     map[string]int{},
		roleDist:    map[string][]iamdomain.RoleCount{},
	}
}

func (f *fakeObsRepo) CountActiveSeats(_ context.Context, tenantID string) (int, error) {
	return f.seats[tenantID], nil
}

func (f *fakeObsRepo) StorageUsedBytes(_ context.Context, tenantID string) (int64, error) {
	if f.storageErr != nil {
		return 0, f.storageErr
	}
	if v, ok := f.storage[tenantID]; ok {
		return v, nil
	}
	return -1, nil
}

func (f *fakeObsRepo) CountAuditEventsByActionPrefix(_ context.Context, tenantID, prefix string, within int) (int64, error) {
	key := tenantID + "|" + prefix
	f.lastSeenPrefix = prefix
	_ = within
	return f.apiByPrefix[key], nil
}

func (f *fakeObsRepo) CountActiveUsersInWindow(_ context.Context, tenantID string, within int) (int64, error) {
	_ = within
	return f.active[tenantID], nil
}

func (f *fakeObsRepo) GetTenantPlan(_ context.Context, tenantID string) (iamdomain.TenantPlan, error) {
	if p, ok := f.plans[tenantID]; ok {
		return p, nil
	}
	return iamdomain.TenantPlan{}, iamdomain.ErrTenantPlanNotFound
}

func (f *fakeObsRepo) CountLockedAccounts(_ context.Context, tenantID string) (int, error) {
	return f.locked[tenantID], nil
}

func (f *fakeObsRepo) CountFailedLoginsInWindow(_ context.Context, tenantID string, within int) (int64, error) {
	_ = within
	return f.failed[tenantID], nil
}

func (f *fakeObsRepo) CountDormantUsers(_ context.Context, tenantID string, threshold int) (int, error) {
	f.dormantThresh = threshold
	f.dormantTenant = tenantID
	return f.dormant[tenantID], nil
}

func (f *fakeObsRepo) RoleDistribution(_ context.Context, tenantID string) ([]iamdomain.RoleCount, error) {
	return f.roleDist[tenantID], nil
}

type fakeMfa struct{ pct float32 }

func (f fakeMfa) MfaCoveragePct(_ context.Context, _ string) (float32, error) {
	return f.pct, nil
}

const tA = "11111111-1111-1111-1111-111111111111"
const tB = "22222222-2222-2222-2222-222222222222"

func TestGetUsage_TenantIsolation(t *testing.T) {
	repo := newFakeObsRepo()
	repo.seats[tA] = 7
	repo.seats[tB] = 99
	repo.plans[tA] = iamdomain.TenantPlan{TenantID: tA, Tier: iamdomain.PlanTierPro, SeatsAllocated: 100, StorageAllocatedBytes: 5 << 30}
	svc := iamapp.NewObservabilityService(repo, fakeMfa{})

	got, err := svc.GetUsage(context.Background(), tA)
	if err != nil {
		t.Fatalf("GetUsage(tA): %v", err)
	}
	if got.Seats.Used != 7 || got.Seats.Allocated != 100 {
		t.Fatalf("tenant A seats = %+v, want used=7 allocated=100", got.Seats)
	}
	if got.PlanTier != iamdomain.PlanTierPro {
		t.Fatalf("tenant A planTier = %q, want pro", got.PlanTier)
	}

	// Tenant B never reaches A's plan row.
	gotB, err := svc.GetUsage(context.Background(), tB)
	if err != nil {
		t.Fatalf("GetUsage(tB): %v", err)
	}
	if gotB.Seats.Used != 99 {
		t.Fatalf("tenant B seats.Used = %d, want 99", gotB.Seats.Used)
	}
	if gotB.Seats.Allocated != 0 {
		t.Fatalf("tenant B seats.Allocated = %d, want 0 (no plan row)", gotB.Seats.Allocated)
	}
	if gotB.PlanTier != "" {
		t.Fatalf("tenant B planTier = %q, want empty", gotB.PlanTier)
	}
}

func TestGetUsage_EmptyTenantRejected(t *testing.T) {
	svc := iamapp.NewObservabilityService(newFakeObsRepo(), fakeMfa{})
	if _, err := svc.GetUsage(context.Background(), "  "); !errors.Is(err, iamapp.ErrTenantRequired) {
		t.Fatalf("err = %v, want ErrTenantRequired", err)
	}
}

func TestGetKpi_DormantThresholdIs30Days(t *testing.T) {
	repo := newFakeObsRepo()
	repo.dormant[tA] = 4
	svc := iamapp.NewObservabilityService(repo, fakeMfa{pct: 42.5})

	got, err := svc.GetKpi(context.Background(), tA)
	if err != nil {
		t.Fatalf("GetKpi: %v", err)
	}
	if got.DormantUsers30d != 4 {
		t.Fatalf("dormantUsers30d = %d, want 4", got.DormantUsers30d)
	}
	if repo.dormantThresh != iamapp.DormantThresholdDays {
		t.Fatalf("dormantThresh = %d, want %d", repo.dormantThresh, iamapp.DormantThresholdDays)
	}
	if iamapp.DormantThresholdDays != 30 {
		t.Fatalf("DormantThresholdDays = %d, want 30", iamapp.DormantThresholdDays)
	}
	if got.MfaCoveragePct != 42.5 {
		t.Fatalf("mfaPct = %v, want 42.5", got.MfaCoveragePct)
	}
}

func TestGetKpi_AuditRatePerMinute(t *testing.T) {
	repo := newFakeObsRepo()
	// 120 events in the 60-minute window → 2.0 / minute.
	repo.apiByPrefix[tA+"|"] = 120
	svc := iamapp.NewObservabilityService(repo, fakeMfa{})

	got, err := svc.GetKpi(context.Background(), tA)
	if err != nil {
		t.Fatalf("GetKpi: %v", err)
	}
	if got.AuditEventsPerMinute != 2.0 {
		t.Fatalf("rate = %v, want 2.0", got.AuditEventsPerMinute)
	}
}

func TestGetKpi_RoleDistributionPropagates(t *testing.T) {
	repo := newFakeObsRepo()
	repo.roleDist[tA] = []iamdomain.RoleCount{
		{Role: iamdomain.Role("editor"), Count: 3},
		{Role: iamdomain.Role("viewer"), Count: 5},
	}
	svc := iamapp.NewObservabilityService(repo, fakeMfa{})

	got, err := svc.GetKpi(context.Background(), tA)
	if err != nil {
		t.Fatalf("GetKpi: %v", err)
	}
	if len(got.RoleDistribution) != 2 || got.RoleDistribution[1].Count != 5 {
		t.Fatalf("roleDistribution = %+v", got.RoleDistribution)
	}
}

func TestGetUsage_UsesPlanTierPrefix(t *testing.T) {
	repo := newFakeObsRepo()
	repo.apiByPrefix[tA+"|"+iamapp.ApiCallActionPrefix] = 7
	svc := iamapp.NewObservabilityService(repo, fakeMfa{})

	got, err := svc.GetUsage(context.Background(), tA)
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if got.APICalls.Last24h != 7 {
		t.Fatalf("apiCalls.last24h = %d, want 7", got.APICalls.Last24h)
	}
	if repo.lastSeenPrefix != iamapp.ApiCallActionPrefix {
		t.Fatalf("lastSeenPrefix = %q, want %q", repo.lastSeenPrefix, iamapp.ApiCallActionPrefix)
	}
}
