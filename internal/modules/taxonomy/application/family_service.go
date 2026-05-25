package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

type FamilyService struct {
	families  domain.FamilyRepository
	govLogger domain.GovernanceLogger
}

func NewFamilyService(families domain.FamilyRepository, govLogger domain.GovernanceLogger) *FamilyService {
	return &FamilyService{families: families, govLogger: govLogger}
}

func (s *FamilyService) List(ctx context.Context, includeInactive bool) ([]domain.DocumentFamily, error) {
	return s.families.List(ctx, includeInactive)
}

func (s *FamilyService) Get(ctx context.Context, code string) (*domain.DocumentFamily, error) {
	return s.families.GetByCode(ctx, code)
}

func (s *FamilyService) Create(ctx context.Context, f *domain.DocumentFamily) error {
	f.IsActive = true
	if err := s.families.Create(ctx, f); err != nil {
		return err
	}
	if s.govLogger != nil {
		payload, _ := json.Marshal(map[string]string{"code": f.Code, "name": f.Name})
		_ = s.govLogger.Log(ctx, domain.GovernanceEvent{
			EventType:    "family.created",
			ResourceType: "document_family",
			ResourceID:   f.Code,
			PayloadJSON:  payload,
		})
	}
	return nil
}

func (s *FamilyService) Update(ctx context.Context, f *domain.DocumentFamily) (*domain.DocumentFamily, error) {
	if strings.TrimSpace(f.Name) == "" {
		return nil, errors.New("family name must not be empty")
	}
	tx, err := s.families.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	existing, err := s.families.GetByCodeForUpdate(ctx, tx, f.Code)
	if err != nil {
		return nil, err
	}
	existing.Name = f.Name
	existing.Description = f.Description
	if err := s.families.UpdateTx(ctx, tx, existing); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	if s.govLogger != nil {
		payload, _ := json.Marshal(map[string]string{"code": existing.Code, "name": existing.Name})
		_ = s.govLogger.Log(ctx, domain.GovernanceEvent{
			EventType:    "family.updated",
			ResourceType: "document_family",
			ResourceID:   existing.Code,
			PayloadJSON:  payload,
		})
	}
	return existing, nil
}

func (s *FamilyService) Deactivate(ctx context.Context, code string) error {
	tx, err := s.families.BeginTx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	f, err := s.families.GetByCodeForUpdate(ctx, tx, code)
	if err != nil {
		return err
	}
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	hasProfiles, err := s.families.HasActiveProfilesTx(ctx, tx, tenantID, code)
	if err != nil {
		return err
	}
	if hasProfiles {
		return domain.ErrFamilyHasProfiles
	}
	if err := f.Deactivate(); err != nil {
		return err
	}
	if err := s.families.UpdateTx(ctx, tx, f); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	if s.govLogger != nil {
		payload, _ := json.Marshal(map[string]string{"code": code})
		_ = s.govLogger.Log(ctx, domain.GovernanceEvent{
			EventType:    "family.deactivated",
			ResourceType: "document_family",
			ResourceID:   code,
			PayloadJSON:  payload,
		})
	}
	return nil
}
