package application

import (
	"context"
	"fmt"
	"strings"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
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
	newFamily, err := domain.NewDocumentFamily(*f)
	if err != nil {
		return fmt.Errorf("normalize family create payload: %w", err)
	}
	if err := s.families.Create(ctx, newFamily); err != nil {
		return fmt.Errorf("persist family create: %w", err)
	}
	if s.govLogger != nil {
		payload, err := marshalGovernancePayload(map[string]string{"code": string(newFamily.Code), "name": newFamily.Name})
		if err != nil {
			return err
		}
		tenantID, _ := tenant.FromContext(ctx)
		actorUserID, _ := authn.UserIDFromContext(ctx)
		if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
			TenantID:     tenantID,
			EventType:    domain.GovernanceEventTypeFamilyCreated,
			ActorUserID:  actorUserID,
			ResourceType: "document_family",
			ResourceID:   string(newFamily.Code),
			PayloadJSON:  payload,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *FamilyService) Update(ctx context.Context, f *domain.DocumentFamily) (*domain.DocumentFamily, error) {
	if strings.TrimSpace(f.Name) == "" {
		return nil, domain.ErrFamilyNameRequired
	}
	tx, err := s.families.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update family tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	existing, err := s.families.GetByCodeForUpdate(ctx, tx, string(f.Code))
	if err != nil {
		return nil, fmt.Errorf("load family for update: %w", err)
	}
	normalized, err := domain.NewDocumentFamily(domain.DocumentFamily{
		Code:        existing.Code,
		Name:        f.Name,
		Description: f.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("normalize family update payload: %w", err)
	}
	existing.Name = normalized.Name
	existing.Description = normalized.Description
	if err := s.families.UpdateTx(ctx, tx, existing); err != nil {
		return nil, fmt.Errorf("persist family update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update family tx: %w", err)
	}
	committed = true
	if s.govLogger != nil {
		payload, err := marshalGovernancePayload(map[string]string{"code": string(existing.Code), "name": existing.Name})
		if err != nil {
			return nil, err
		}
		tenantID, _ := tenant.FromContext(ctx)
		actorUserID, _ := authn.UserIDFromContext(ctx)
		if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
			TenantID:     tenantID,
			EventType:    domain.GovernanceEventTypeFamilyUpdated,
			ActorUserID:  actorUserID,
			ResourceType: "document_family",
			ResourceID:   string(existing.Code),
			PayloadJSON:  payload,
		}); err != nil {
			return nil, err
		}
	}
	return existing, nil
}

func (s *FamilyService) Deactivate(ctx context.Context, code string) error {
	tx, err := s.families.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin deactivate family tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	f, err := s.families.GetByCodeForUpdate(ctx, tx, code)
	if err != nil {
		return fmt.Errorf("load family for deactivate: %w", err)
	}
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	hasProfiles, err := s.families.HasActiveProfilesTx(ctx, tx, tenantID, code)
	if err != nil {
		return fmt.Errorf("check active profiles before family deactivate: %w", err)
	}
	if hasProfiles {
		return domain.ErrFamilyHasProfiles
	}
	if err := f.Deactivate(); err != nil {
		return err
	}
	if err := s.families.UpdateTx(ctx, tx, f); err != nil {
		return fmt.Errorf("persist family deactivate: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit deactivate family tx: %w", err)
	}
	committed = true
	if s.govLogger != nil {
		payload, err := marshalGovernancePayload(map[string]string{"code": code})
		if err != nil {
			return err
		}
		actorUserID, _ := authn.UserIDFromContext(ctx)
		if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
			TenantID:     tenantID,
			EventType:    domain.GovernanceEventTypeFamilyDeactivated,
			ActorUserID:  actorUserID,
			ResourceType: "document_family",
			ResourceID:   code,
			PayloadJSON:  payload,
		}); err != nil {
			return err
		}
	}
	return nil
}
