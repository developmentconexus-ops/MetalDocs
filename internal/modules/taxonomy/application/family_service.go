package application

import (
	"context"
	"errors"
	"strings"

	"metaldocs/internal/modules/taxonomy/domain"
)

type FamilyService struct {
	families domain.FamilyRepository
}

func NewFamilyService(families domain.FamilyRepository) *FamilyService {
	return &FamilyService{families: families}
}

func (s *FamilyService) List(ctx context.Context, includeInactive bool) ([]domain.DocumentFamily, error) {
	return s.families.List(ctx, includeInactive)
}

func (s *FamilyService) Get(ctx context.Context, code string) (*domain.DocumentFamily, error) {
	return s.families.GetByCode(ctx, code)
}

func (s *FamilyService) Create(ctx context.Context, f *domain.DocumentFamily) error {
	f.IsActive = true
	return s.families.Create(ctx, f)
}

func (s *FamilyService) Update(ctx context.Context, f *domain.DocumentFamily) (*domain.DocumentFamily, error) {
	if strings.TrimSpace(f.Name) == "" {
		return nil, errors.New("family name must not be empty")
	}
	existing, err := s.families.GetByCode(ctx, f.Code)
	if err != nil {
		return nil, err
	}
	existing.Name = f.Name
	existing.Description = f.Description
	if err := s.families.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *FamilyService) Deactivate(ctx context.Context, code string) error {
	f, err := s.families.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	hasProfiles, err := s.families.HasActiveProfiles(ctx, code)
	if err != nil {
		return err
	}
	if hasProfiles {
		return domain.ErrFamilyHasProfiles
	}
	if err := f.Deactivate(); err != nil {
		return err
	}
	return s.families.Update(ctx, f)
}
