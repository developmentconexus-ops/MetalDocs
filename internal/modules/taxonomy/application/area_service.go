package application

import (
	"context"
	"fmt"
	"time"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
)

type AreaService struct {
	areas     domain.AreaRepository
	govLogger domain.GovernanceLogger
	now       func() time.Time
}

func NewAreaService(areas domain.AreaRepository, govLogger domain.GovernanceLogger) *AreaService {
	return &AreaService{
		areas:     areas,
		govLogger: govLogger,
		now:       time.Now,
	}
}

func (s *AreaService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.ProcessArea, error) {
	areas, err := s.areas.List(ctx, tenantID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: list areas: %w", err)
	}
	return areas, nil
}

func (s *AreaService) Get(ctx context.Context, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	area, err := s.areas.GetByCode(ctx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: get area %q: %w", code, err)
	}
	return area, nil
}

func (s *AreaService) Create(ctx context.Context, a *domain.ProcessArea) error {
	newArea, err := domain.NewProcessArea(*a)
	if err != nil {
		return fmt.Errorf("taxonomy: validate area create: %w", err)
	}
	tx, err := s.areas.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin area create tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := s.areas.CreateTx(ctx, tx, newArea); err != nil {
		return fmt.Errorf("taxonomy: create area %q: %w", newArea.Code, err)
	}
	payload, err := marshalGovernancePayload(map[string]string{
		"name": newArea.Name,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal area create governance payload: %w", err)
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.LogTx(ctx, sqlTxFromFamilyTx(tx), domain.GovernanceEvent{
		TenantID:     newArea.TenantID,
		EventType:    domain.GovernanceEventTypeAreaCreated,
		ActorUserID:  actorUserID,
		ResourceType: "process_area",
		ResourceID:   string(newArea.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log area create governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit area create tx: %w", err)
	}
	committed = true
	return nil
}

func (s *AreaService) Update(ctx context.Context, a *domain.ProcessArea) error {
	normalized, err := domain.NewProcessArea(*a)
	if err != nil {
		return fmt.Errorf("taxonomy: validate area update: %w", err)
	}
	tx, err := s.areas.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin area update tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	existing, err := s.areas.GetByCodeForUpdate(ctx, tx, a.TenantID, a.Code)
	if err != nil {
		return fmt.Errorf("taxonomy: get area %q for update: %w", a.Code, err)
	}
	existing.Name = normalized.Name
	existing.Description = normalized.Description
	existing.ParentCode = normalized.ParentCode
	existing.OwnerUserID = normalized.OwnerUserID
	existing.DefaultApproverRole = normalized.DefaultApproverRole
	existing.ArchivedAt = normalized.ArchivedAt
	if err := s.areas.UpdateTx(ctx, tx, existing); err != nil {
		return fmt.Errorf("taxonomy: update area %q: %w", existing.Code, err)
	}
	payload, err := marshalGovernancePayload(map[string]string{
		"name": existing.Name,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal area update governance payload: %w", err)
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.LogTx(ctx, sqlTxFromFamilyTx(tx), domain.GovernanceEvent{
		TenantID:     existing.TenantID,
		EventType:    domain.GovernanceEventTypeAreaUpdated,
		ActorUserID:  actorUserID,
		ResourceType: "process_area",
		ResourceID:   string(existing.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log area update governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit area update tx: %w", err)
	}
	committed = true
	return nil
}

func (s *AreaService) Archive(ctx context.Context, tenantID string, areaCode domain.AreaCode, actorID string) error {
	tx, err := s.areas.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin archive area tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	area, err := s.areas.GetByCodeForUpdate(ctx, tx, tenantID, areaCode)
	if err != nil {
		return fmt.Errorf("taxonomy: lock area %q: %w", areaCode, err)
	}
	if err := area.Archive(s.now()); err != nil {
		return fmt.Errorf("taxonomy: archive area %q: %w", areaCode, err)
	}
	if err := s.areas.UpdateTx(ctx, tx, area); err != nil {
		return fmt.Errorf("taxonomy: update archived area %q: %w", areaCode, err)
	}
	if err := s.govLogger.LogTx(ctx, sqlTxFromFamilyTx(tx), domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeAreaArchived,
		ActorUserID:  actorID,
		ResourceType: "process_area",
		ResourceID:   string(areaCode),
		PayloadJSON:  []byte(`{}`),
	}); err != nil {
		return fmt.Errorf("taxonomy: log area archive governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit archive area tx: %w", err)
	}
	committed = true
	return nil
}
