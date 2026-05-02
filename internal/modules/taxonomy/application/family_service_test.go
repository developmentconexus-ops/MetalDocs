package application

import (
	"context"
	"testing"

	"metaldocs/internal/modules/taxonomy/domain"
)

type fakeFamilyRepo struct {
	families       map[string]*domain.DocumentFamily
	activeProfiles map[string]bool
}

func newFakeFamilyRepo() *fakeFamilyRepo {
	return &fakeFamilyRepo{
		families:       make(map[string]*domain.DocumentFamily),
		activeProfiles: make(map[string]bool),
	}
}

func (r *fakeFamilyRepo) GetByCode(_ context.Context, code string) (*domain.DocumentFamily, error) {
	f, ok := r.families[code]
	if !ok {
		return nil, domain.ErrFamilyNotFound
	}
	return f, nil
}

func (r *fakeFamilyRepo) List(_ context.Context, includeInactive bool) ([]domain.DocumentFamily, error) {
	out := make([]domain.DocumentFamily, 0)
	for _, f := range r.families {
		if includeInactive || f.IsActive {
			out = append(out, *f)
		}
	}
	return out, nil
}

func (r *fakeFamilyRepo) Create(_ context.Context, f *domain.DocumentFamily) error {
	r.families[f.Code] = f
	return nil
}

func (r *fakeFamilyRepo) Update(_ context.Context, f *domain.DocumentFamily) error {
	if _, ok := r.families[f.Code]; !ok {
		return domain.ErrFamilyNotFound
	}
	r.families[f.Code] = f
	return nil
}

func (r *fakeFamilyRepo) HasActiveProfiles(_ context.Context, familyCode string) (bool, error) {
	return r.activeProfiles[familyCode], nil
}

func TestFamilyService_Create(t *testing.T) {
	repo := newFakeFamilyRepo()
	svc := NewFamilyService(repo)

	f := &domain.DocumentFamily{Code: "policy", Name: "Policy"}
	if err := svc.Create(context.Background(), f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := svc.Get(context.Background(), "policy")
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

func TestFamilyService_Deactivate_BlockedByProfiles(t *testing.T) {
	repo := newFakeFamilyRepo()
	repo.families["policy"] = &domain.DocumentFamily{Code: "policy", IsActive: true}
	repo.activeProfiles["policy"] = true
	svc := NewFamilyService(repo)

	if err := svc.Deactivate(context.Background(), "policy"); err != domain.ErrFamilyHasProfiles {
		t.Fatalf("want ErrFamilyHasProfiles, got %v", err)
	}
}

func TestFamilyService_Deactivate_OK(t *testing.T) {
	repo := newFakeFamilyRepo()
	repo.families["orphan"] = &domain.DocumentFamily{Code: "orphan", IsActive: true}
	svc := NewFamilyService(repo)

	if err := svc.Deactivate(context.Background(), "orphan"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := repo.GetByCode(context.Background(), "orphan")
	if got.IsActive {
		t.Fatal("expected IsActive=false after Deactivate")
	}
}

func TestFamilyService_Update_PreservesIsActive(t *testing.T) {
	repo := newFakeFamilyRepo()
	repo.families["policy"] = &domain.DocumentFamily{Code: "policy", Name: "Old", IsActive: false}
	svc := NewFamilyService(repo)

	if err := svc.Update(context.Background(), &domain.DocumentFamily{Code: "policy", Name: "New"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := repo.GetByCode(context.Background(), "policy")
	if got.Name != "New" {
		t.Fatalf("name = %q, want %q", got.Name, "New")
	}
	if got.IsActive {
		t.Fatal("Update must not re-activate an inactive family")
	}
}

func TestFamilyService_Update_NotFound(t *testing.T) {
	repo := newFakeFamilyRepo()
	svc := NewFamilyService(repo)
	err := svc.Update(context.Background(), &domain.DocumentFamily{Code: "missing", Name: "X"})
	if err != domain.ErrFamilyNotFound {
		t.Fatalf("want ErrFamilyNotFound, got %v", err)
	}
}
