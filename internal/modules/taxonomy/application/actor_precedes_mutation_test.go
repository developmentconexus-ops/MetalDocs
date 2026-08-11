package application

import (
	"context"
	"errors"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/tenant"
)

// A3.3 T1 — the actor must be resolved BEFORE any transaction, repository call,
// row lock or domain mutation, not next to the GovernanceEvent at the end of the
// method.
//
// "The transaction rolled back" is a strictly weaker property than "the
// transaction never started": a rollback still means the service opened a tx,
// took FOR UPDATE row locks and ran domain mutations on behalf of a caller
// nobody had identified. These tests assert the strong property directly, by
// counting repository calls rather than by inspecting the returned error alone.
//
// The proof lives at the application layer on purpose. The delivery-layer 401 is
// proven elsewhere (approval submit, templates create-version); the finding
// here is BELOW delivery, so an HTTP test could not have caught it — an
// application service is reachable from any caller that holds a context.

// actorlessContexts enumerates the three shapes of "no actor" this boundary has
// to reject identically. The whitespace case matters because the low-level
// identity accessor stores whatever string it was given: only the canonical
// accessor's TrimSpace turns "   " into absence.
//
// Every one of them carries a tenant. Without it the family methods would stop
// at their own tenant.FromContext guard, and the test would pass for a reason
// that has nothing to do with the actor — it would still be green if the actor
// check moved back below BeginTx. The tenant is what makes the counters, not an
// unrelated guard, the thing being asserted.
func actorlessContexts() []struct {
	name string
	ctx  context.Context
} {
	base := tenant.WithTenantID(context.Background(), "tenant-a")
	return []struct {
		name string
		ctx  context.Context
	}{
		{"no auth context at all", base},
		{"empty actor id", iamdomain.WithAuthContext(base, "", []iamdomain.Role{})},
		{"whitespace-only actor id", iamdomain.WithAuthContext(base, "   ", []iamdomain.Role{})},
	}
}

// ---- Area -----------------------------------------------------------------

// countingAreaRepo wraps the existing fake with call counters. It overrides only
// the methods the T1 property is about; everything else is the fake's behavior.
type countingAreaRepo struct {
	fakeAreaRepository
	beginTx  int
	createTx int
	updateTx int
	lockRow  int
}

func newCountingAreaRepo() *countingAreaRepo {
	return &countingAreaRepo{fakeAreaRepository: *newFakeAreaRepository()}
}

func (r *countingAreaRepo) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	r.beginTx++
	return r.fakeAreaRepository.BeginTx(ctx)
}

func (r *countingAreaRepo) CreateTx(ctx context.Context, tx domain.FamilyTx, a *domain.ProcessArea) error {
	r.createTx++
	return r.fakeAreaRepository.CreateTx(ctx, tx, a)
}

func (r *countingAreaRepo) UpdateTx(ctx context.Context, tx domain.FamilyTx, a *domain.ProcessArea) error {
	r.updateTx++
	return r.fakeAreaRepository.UpdateTx(ctx, tx, a)
}

func (r *countingAreaRepo) GetByCodeForUpdate(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	r.lockRow++
	return r.fakeAreaRepository.GetByCodeForUpdate(ctx, tx, tenantID, code)
}

func TestAreaServiceT1_MissingActorReachesNoRepositoryWork(t *testing.T) {
	methods := []struct {
		name string
		call func(*AreaService, context.Context) error
	}{
		{"Create", func(s *AreaService, ctx context.Context) error {
			return s.Create(ctx, &domain.ProcessArea{Code: "AR-01", TenantID: "tenant-a", Name: "Finance"})
		}},
		{"Update", func(s *AreaService, ctx context.Context) error {
			return s.Update(ctx, &domain.ProcessArea{Code: "AR-01", TenantID: "tenant-a", Name: "Finance"})
		}},
	}

	for _, m := range methods {
		for _, c := range actorlessContexts() {
			t.Run(m.name+"/"+c.name, func(t *testing.T) {
				repo := newCountingAreaRepo()
				repo.put(&domain.ProcessArea{Code: "AR-01", TenantID: "tenant-a", Name: "Finance"})
				logger := &fakeGovernanceLogger{}

				err := m.call(NewAreaService(repo, logger), c.ctx)

				if !errors.Is(err, authn.ErrMissingActor) {
					t.Fatalf("expected authn.ErrMissingActor, got %v", err)
				}
				if repo.beginTx != 0 {
					t.Fatalf("BeginTx must not be called without an actor, got %d", repo.beginTx)
				}
				if repo.createTx != 0 || repo.updateTx != 0 {
					t.Fatalf("no repository mutation may run without an actor, got CreateTx=%d UpdateTx=%d", repo.createTx, repo.updateTx)
				}
				if repo.lockRow != 0 {
					t.Fatalf("GetByCodeForUpdate must not take a row lock without an actor, got %d", repo.lockRow)
				}
				if len(logger.events) != 0 {
					t.Fatalf("governance logger must not be called without an actor, got %d events", len(logger.events))
				}
			})
		}
	}
}

// ---- Family ---------------------------------------------------------------

type countingFamilyRepo struct {
	fakeFamilyRepo
	beginTx           int
	createTx          int
	updateTx          int
	lockRow           int
	hasActiveProfiles int
}

func newCountingFamilyRepo() *countingFamilyRepo {
	return &countingFamilyRepo{fakeFamilyRepo: *newFakeFamilyRepo()}
}

func (r *countingFamilyRepo) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	r.beginTx++
	return r.fakeFamilyRepo.BeginTx(ctx)
}

func (r *countingFamilyRepo) CreateTx(ctx context.Context, tx domain.FamilyTx, f *domain.DocumentFamily) error {
	r.createTx++
	return r.fakeFamilyRepo.CreateTx(ctx, tx, f)
}

func (r *countingFamilyRepo) UpdateTx(ctx context.Context, tx domain.FamilyTx, f *domain.DocumentFamily) error {
	r.updateTx++
	return r.fakeFamilyRepo.UpdateTx(ctx, tx, f)
}

func (r *countingFamilyRepo) GetByCodeForUpdate(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	r.lockRow++
	return r.fakeFamilyRepo.GetByCodeForUpdate(ctx, tx, tenantID, code)
}

func (r *countingFamilyRepo) HasActiveProfilesTx(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.FamilyCode) (bool, error) {
	r.hasActiveProfiles++
	return r.fakeFamilyRepo.HasActiveProfilesTx(ctx, tx, tenantID, code)
}

func TestFamilyServiceT1_MissingActorReachesNoRepositoryOrDomainMutation(t *testing.T) {
	methods := []struct {
		name string
		call func(*FamilyService, context.Context) error
	}{
		{"Create", func(s *FamilyService, ctx context.Context) error {
			return s.Create(ctx, &domain.DocumentFamily{Code: "FAM-01", TenantID: "tenant-a", Name: "Quality"})
		}},
		{"Update", func(s *FamilyService, ctx context.Context) error {
			_, err := s.Update(ctx, &domain.DocumentFamily{Code: "FAM-01", TenantID: "tenant-a", Name: "Quality"})
			return err
		}},
		// Deactivate is the reason this test exists in the form it does: before
		// T1 it resolved the actor AFTER GetByCodeForUpdate, HasActiveProfilesTx
		// and f.Deactivate() — an unidentified caller could flip is_active on a
		// domain object and only then be refused.
		{"Deactivate", func(s *FamilyService, ctx context.Context) error {
			return s.Deactivate(ctx, "FAM-01")
		}},
	}

	for _, m := range methods {
		for _, c := range actorlessContexts() {
			t.Run(m.name+"/"+c.name, func(t *testing.T) {
				repo := newCountingFamilyRepo()
				seeded := &domain.DocumentFamily{Code: "FAM-01", TenantID: "tenant-a", Name: "Quality", IsActive: true}
				repo.families[familyKey("tenant-a", "FAM-01")] = seeded
				logger := &fakeGovernanceLogger{}

				err := m.call(NewFamilyService(repo, logger), c.ctx)

				if !errors.Is(err, authn.ErrMissingActor) {
					t.Fatalf("expected authn.ErrMissingActor, got %v", err)
				}
				if repo.beginTx != 0 {
					t.Fatalf("BeginTx must not be called without an actor, got %d", repo.beginTx)
				}
				if repo.createTx != 0 || repo.updateTx != 0 {
					t.Fatalf("no repository mutation may run without an actor, got CreateTx=%d UpdateTx=%d", repo.createTx, repo.updateTx)
				}
				if repo.lockRow != 0 || repo.hasActiveProfiles != 0 {
					t.Fatalf("no in-tx read may run without an actor, got GetByCodeForUpdate=%d HasActiveProfilesTx=%d", repo.lockRow, repo.hasActiveProfiles)
				}
				// The domain mutation itself: Deactivate() flips IsActive on the
				// loaded aggregate. Nothing may have touched it.
				if !seeded.IsActive {
					t.Fatal("domain mutation ran without an actor: family was deactivated")
				}
				if len(logger.events) != 0 {
					t.Fatalf("governance logger must not be called without an actor, got %d events", len(logger.events))
				}
			})
		}
	}
}

// ---- Profile --------------------------------------------------------------

type countingProfileRepo struct {
	fakeProfileRepository
	beginTx  int
	createTx int
	updateTx int
	lockRow  int
}

func newCountingProfileRepo() *countingProfileRepo {
	return &countingProfileRepo{fakeProfileRepository: *newFakeProfileRepository()}
}

func (r *countingProfileRepo) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	r.beginTx++
	return r.fakeProfileRepository.BeginTx(ctx)
}

func (r *countingProfileRepo) CreateTx(ctx context.Context, tx domain.FamilyTx, p *domain.DocumentProfile) error {
	r.createTx++
	return r.fakeProfileRepository.CreateTx(ctx, tx, p)
}

func (r *countingProfileRepo) UpdateTx(ctx context.Context, tx domain.FamilyTx, p *domain.DocumentProfile) error {
	r.updateTx++
	return r.fakeProfileRepository.UpdateTx(ctx, tx, p)
}

func (r *countingProfileRepo) GetByCodeForUpdate(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	r.lockRow++
	return r.fakeProfileRepository.GetByCodeForUpdate(ctx, tx, tenantID, code)
}

func TestProfileServiceT1_MissingActorReachesNoRepositoryWork(t *testing.T) {
	methods := []struct {
		name string
		call func(*ProfileService, context.Context) error
	}{
		{"Create", func(s *ProfileService, ctx context.Context) error {
			return s.Create(ctx, &domain.DocumentProfile{
				Code: "PRF-01", TenantID: "tenant-a", Name: "SOP", FamilyCode: "FAM-01", EditableByRole: "editor",
			})
		}},
		{"Update", func(s *ProfileService, ctx context.Context) error {
			return s.Update(ctx, &domain.DocumentProfile{
				Code: "PRF-01", TenantID: "tenant-a", Name: "SOP", FamilyCode: "FAM-01", EditableByRole: "editor",
			})
		}},
	}

	for _, m := range methods {
		for _, c := range actorlessContexts() {
			t.Run(m.name+"/"+c.name, func(t *testing.T) {
				repo := newCountingProfileRepo()
				repo.put(&domain.DocumentProfile{Code: "PRF-01", TenantID: "tenant-a", EditableByRole: "editor"})
				logger := &fakeGovernanceLogger{}

				err := m.call(NewProfileService(repo, &fakeTemplateVersionChecker{}, logger), c.ctx)

				if !errors.Is(err, authn.ErrMissingActor) {
					t.Fatalf("expected authn.ErrMissingActor, got %v", err)
				}
				if repo.beginTx != 0 {
					t.Fatalf("BeginTx must not be called without an actor, got %d", repo.beginTx)
				}
				if repo.createTx != 0 || repo.updateTx != 0 {
					t.Fatalf("no repository mutation may run without an actor, got CreateTx=%d UpdateTx=%d", repo.createTx, repo.updateTx)
				}
				if repo.lockRow != 0 {
					t.Fatalf("GetByCodeForUpdate must not take a row lock without an actor, got %d", repo.lockRow)
				}
				if len(logger.events) != 0 {
					t.Fatalf("governance logger must not be called without an actor, got %d events", len(logger.events))
				}
			})
		}
	}
}
