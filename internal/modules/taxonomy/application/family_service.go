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
	families, err := s.families.List(ctx, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: list families: %w", err)
	}
	return families, nil
}

func (s *FamilyService) Get(ctx context.Context, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	family, err := s.families.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: get family %q: %w", code, err)
	}
	return family, nil
}

func (s *FamilyService) Create(ctx context.Context, f *domain.DocumentFamily) error {
	newFamily, err := domain.NewDocumentFamily(*f)
	if err != nil {
		return fmt.Errorf("taxonomy: validate family create: %w", err)
	}
	if err := s.families.Create(ctx, newFamily); err != nil {
		return fmt.Errorf("taxonomy: create family %q: %w", newFamily.Code, err)
	}
	if s.govLogger != nil {
		payload, err := marshalGovernancePayload(map[string]string{"code": string(newFamily.Code), "name": newFamily.Name})
		if err != nil {
			return fmt.Errorf("taxonomy: marshal family create governance payload: %w", err)
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
			return fmt.Errorf("taxonomy: log family create governance event: %w", err)
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
		return nil, fmt.Errorf("taxonomy: begin update family tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	existing, err := s.families.GetByCodeForUpdate(ctx, tx, f.Code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: lock family %q: %w", f.Code, err)
	}
	normalized, err := domain.NewDocumentFamily(domain.DocumentFamily{
		Code:        existing.Code,
		Name:        f.Name,
		Description: f.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("taxonomy: validate family update: %w", err)
	}
	existing.Name = normalized.Name
	existing.Description = normalized.Description
	if err := s.families.UpdateTx(ctx, tx, existing); err != nil {
		return nil, fmt.Errorf("taxonomy: update family %q: %w", existing.Code, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("taxonomy: commit update family tx: %w", err)
	}
	committed = true
	if s.govLogger != nil {
		payload, err := marshalGovernancePayload(map[string]string{"code": string(existing.Code), "name": existing.Name})
		if err != nil {
			return nil, fmt.Errorf("taxonomy: marshal family update governance payload: %w", err)
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
			return nil, fmt.Errorf("taxonomy: log family update governance event: %w", err)
		}
	}
	return existing, nil
}

func (s *FamilyService) Deactivate(ctx context.Context, code domain.FamilyCode) error {
	tx, err := s.families.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin deactivate family tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	f, err := s.families.GetByCodeForUpdate(ctx, tx, code)
	if err != nil {
		return fmt.Errorf("taxonomy: lock family %q: %w", code, err)
	}
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: resolve tenant for family deactivate: %w", err)
	}
	hasProfiles, err := s.families.HasActiveProfilesTx(ctx, tx, tenantID, code)
	if err != nil {
		return fmt.Errorf("taxonomy: check active profiles for family %q: %w", code, err)
	}
	if hasProfiles {
		return domain.ErrFamilyHasProfiles
	}
	if err := f.Deactivate(); err != nil {
		return fmt.Errorf("taxonomy: deactivate family %q: %w", code, err)
	}
	if err := s.families.UpdateTx(ctx, tx, f); err != nil {
		return fmt.Errorf("taxonomy: update deactivated family %q: %w", code, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit deactivate family tx: %w", err)
	}
	committed = true
	if s.govLogger != nil {
		payload, err := marshalGovernancePayload(map[string]string{"code": string(code)})
		if err != nil {
			return fmt.Errorf("taxonomy: marshal family deactivate governance payload: %w", err)
		}
		actorUserID, _ := authn.UserIDFromContext(ctx)
		if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
			TenantID:     tenantID,
			EventType:    domain.GovernanceEventTypeFamilyDeactivated,
			ActorUserID:  actorUserID,
			ResourceType: "document_family",
			ResourceID:   string(code),
			PayloadJSON:  payload,
		}); err != nil {
			return fmt.Errorf("taxonomy: log family deactivate governance event: %w", err)
		}
	}
	return nil
}
