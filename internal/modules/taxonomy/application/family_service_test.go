package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeFamilyRepo struct {
	families       map[string]*domain.DocumentFamily
	activeProfiles map[string]bool
	lastTenantID   string
}

type fakeFamilyTx struct{}

func (fakeFamilyTx) Commit() error                                                          { return nil }
func (fakeFamilyTx) Rollback() error                                                        { return nil }
func (fakeFamilyTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error)  { return nil, nil }
func (fakeFamilyTx) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error)  { return nil, nil }
func (fakeFamilyTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row         { return nil }

func newFakeFamilyRepo() *fakeFamilyRepo {
	return &fakeFamilyRepo{
		families:       make(map[string]*domain.DocumentFamily),
		activeProfiles: make(map[string]bool),
	}
}

func familyKey(tenantID, code string) string {
	return tenantID + "/" + code
}

func (r *fakeFamilyRepo) GetByCode(_ context.Context, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	f, ok := r.families[familyKey(tenantID, string(code))]
	if !ok {
		return nil, domain.ErrFamilyNotFound
	}
	return f, nil
}

func (r *fakeFamilyRepo) List(_ context.Context, tenantID string, includeInactive bool) ([]domain.DocumentFamily, error) {
	out := make([]domain.DocumentFamily, 0)
	for _, f := range r.families {
		if f.TenantID != tenantID {
			continue
		}
		if includeInactive || f.IsActive {
			out = append(out, *f)
		}
	}
	return out, nil
}

func (r *fakeFamilyRepo) Create(_ context.Context, f *domain.DocumentFamily) error {
	r.families[familyKey(f.TenantID, string(f.Code))] = f
	return nil
}

func (r *fakeFamilyRepo) CreateTx(_ context.Context, _ domain.FamilyTx, f *domain.DocumentFamily) error {
	r.families[familyKey(f.TenantID, string(f.Code))] = f
	return nil
}

func (r *fakeFamilyRepo) Update(_ context.Context, f *domain.DocumentFamily) error {
	key := familyKey(f.TenantID, string(f.Code))
	if _, ok := r.families[key]; !ok {
		return domain.ErrFamilyNotFound
	}
	r.families[key] = f
	return nil
}

func (r *fakeFamilyRepo) HasActiveProfiles(_ context.Context, _ string, familyCode domain.FamilyCode) (bool, error) {
	return r.activeProfiles[string(familyCode)], nil
}

func (r *fakeFamilyRepo) BeginTx(_ context.Context) (domain.FamilyTx, error) {
	return fakeFamilyTx{}, nil
}

func (r *fakeFamilyRepo) GetByCodeForUpdate(_ context.Context, _ domain.FamilyTx, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	return r.GetByCode(context.Background(), tenantID, code)
}

func (r *fakeFamilyRepo) HasActiveProfilesTx(_ context.Context, _ domain.FamilyTx, tenantID string, familyCode domain.FamilyCode) (bool, error) {
	r.lastTenantID = tenantID
	return r.activeProfiles[string(familyCode)], nil
}

func (r *fakeFamilyRepo) UpdateTx(_ context.Context, _ domain.FamilyTx, f *domain.DocumentFamily) error {
	return r.Update(context.Background(), f)
}

func TestFamilyService_Create(t *testing.T) {
	repo := newFakeFamilyRepo()
	svc := NewFamilyService(repo, &fakeGovernanceLogger{})

	f := &domain.DocumentFamily{Code: "policy", TenantID: "tenant-a", Name: "Policy"}
	if err := svc.Create(context.Background(), f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := svc.Get(context.Background(), "tenant-a", "policy")
	if err != nil {
		t.Fatalf("Get after Create: %v", err)
	}
	if got.Name != "Policy" {
		t.Fatalf("name = %q, want %q", got.Name, "Policy")
	}
	if !got.IsActive {
		t.Fatal("expected IsActive=true after Create")
	}
}

func TestFamilyService_Create_DoesNotMutateCallerPointer_AndLogsTenantActor(t *testing.T) {
	repo := newFakeFamilyRepo()
	logger := &fakeGovernanceLogger{}
	svc := NewFamilyService(repo, logger)
	in := &domain.DocumentFamily{
		Code:        " policy ",
		TenantID:    "tenant-a",
		Name:        " Policy ",
		Description: " Desc ",
		IsActive:    false,
	}
	ctx := tenant.WithTenantID(context.Background(), "tenant-a")
	ctx = iamdomain.WithAuthContext(ctx, "actor-1", nil)

	if err := svc.Create(ctx, in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.IsActive {
		t.Fatal("caller pointer must not be mutated")
	}
	got, err := svc.Get(context.Background(), "tenant-a", "policy")
	if err != nil {
		t.Fatalf("Get after Create: %v", err)
	}
	if !got.IsActive {
		t.Fatal("created family must be active")
	}
	if got.Name != "Policy" || got.Description != "Desc" {
		t.Fatalf("created family must be trimmed, got %#v", got)
	}
	if len(logger.events) != 1 {
		t.Fatalf("expected 1 governance event, got %d", len(logger.events))
	}
	if logger.events[0].TenantID != "tenant-a" || logger.events[0].ActorUserID != "actor-1" {
		t.Fatalf("event tenant/actor = %q/%q", logger.events[0].TenantID, logger.events[0].ActorUserID)
	}
}

func TestFamilyService_Deactivate_BlockedByProfiles(t *testing.T) {
	repo := newFakeFamilyRepo()
	repo.families[familyKey("tenant-a", "policy")] = &domain.DocumentFamily{Code: "policy", TenantID: "tenant-a", IsActive: true}
	repo.activeProfiles["policy"] = true
	svc := NewFamilyService(repo, &fakeGovernanceLogger{})

	ctx := tenant.WithTenantID(context.Background(), "tenant-a")
	if err := svc.Deactivate(ctx, "policy"); err != domain.ErrFamilyHasProfiles {
		t.Fatalf("want ErrFamilyHasProfiles, got %v", err)
	}
	if repo.lastTenantID != "tenant-a" {
		t.Fatalf("tenantID = %q, want tenant-a", repo.lastTenantID)
	}
}

func TestFamilyService_Deactivate_OK(t *testing.T) {
	repo := newFakeFamilyRepo()
	repo.families[familyKey("tenant-a", "orphan")] = &domain.DocumentFamily{Code: "orphan", TenantID: "tenant-a", IsActive: true}
	svc := NewFamilyService(repo, &fakeGovernanceLogger{})

	ctx := tenant.WithTenantID(context.Background(), "tenant-a")
	if err := svc.Deactivate(ctx, "orphan"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := repo.GetByCode(context.Background(), "tenant-a", "orphan")
	if got.IsActive {
		t.Fatal("expected IsActive=false after Deactivate")
	}
}

func TestFamilyService_Update_PreservesIsActive(t *testing.T) {
	repo := newFakeFamilyRepo()
	repo.families[familyKey("tenant-a", "policy")] = &domain.DocumentFamily{Code: "policy", TenantID: "tenant-a", Name: "Old", IsActive: false}
	svc := NewFamilyService(repo, &fakeGovernanceLogger{})

	ctx := tenant.WithTenantID(context.Background(), "tenant-a")
	if _, err := svc.Update(ctx, &domain.DocumentFamily{Code: "policy", Name: "New"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := repo.GetByCode(context.Background(), "tenant-a", "policy")
	if got.Name != "New" {
		t.Fatalf("name = %q, want %q", got.Name, "New")
	}
	if got.IsActive {
		t.Fatal("Update must not re-activate an inactive family")
	}
}

func TestFamilyService_Update_NotFound(t *testing.T) {
	repo := newFakeFamilyRepo()
	svc := NewFamilyService(repo, &fakeGovernanceLogger{})
	ctx := tenant.WithTenantID(context.Background(), "tenant-a")
	_, err := svc.Update(ctx, &domain.DocumentFamily{Code: "missing", Name: "X"})
	if !errors.Is(err, domain.ErrFamilyNotFound) {
		t.Fatalf("want ErrFamilyNotFound, got %v", err)
	}
}

func TestFamilyService_Create_ValidationFromDomainConstructor(t *testing.T) {
	repo := newFakeFamilyRepo()
	svc := NewFamilyService(repo, &fakeGovernanceLogger{})

	err := svc.Create(context.Background(), &domain.DocumentFamily{Code: "  ", TenantID: "tenant-a", Name: "Policy"})
	if !errors.Is(err, domain.ErrFamilyCodeRequired) {
		t.Fatalf("want ErrFamilyCodeRequired, got %v", err)
	}
}

func TestFamilyService_Create_ValidationRequiresTenant(t *testing.T) {
	repo := newFakeFamilyRepo()
	svc := NewFamilyService(repo, &fakeGovernanceLogger{})

	err := svc.Create(context.Background(), &domain.DocumentFamily{Code: "policy", Name: "Policy"})
	if !errors.Is(err, domain.ErrFamilyTenantRequired) {
		t.Fatalf("want ErrFamilyTenantRequired, got %v", err)
	}
}
