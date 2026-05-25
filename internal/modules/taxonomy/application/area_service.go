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
	return s.areas.List(ctx, tenantID, includeArchived)
}

func (s *AreaService) Get(ctx context.Context, tenantID, code string) (*domain.ProcessArea, error) {
	return s.areas.GetByCode(ctx, tenantID, code)
}

func (s *AreaService) Create(ctx context.Context, a *domain.ProcessArea) error {
	newArea, err := domain.NewProcessArea(*a)
	if err != nil {
		return fmt.Errorf("normalize area create payload: %w", err)
	}
	if err := s.areas.Create(ctx, newArea); err != nil {
		return fmt.Errorf("persist area create: %w", err)
	}

	payload, err := marshalGovernancePayload(map[string]string{
		"name": newArea.Name,
	})
	if err != nil {
		return fmt.Errorf("marshal area create governance payload: %w", err)
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     newArea.TenantID,
		EventType:    domain.GovernanceEventTypeAreaCreated,
		ActorUserID:  actorUserID,
		ResourceType: "process_area",
		ResourceID:   string(newArea.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("write area create governance event: %w", err)
	}
	return nil
}

func (s *AreaService) Update(ctx context.Context, a *domain.ProcessArea) error {
	if err := s.areas.Update(ctx, a); err != nil {
		return fmt.Errorf("persist area update: %w", err)
	}

	payload, err := marshalGovernancePayload(map[string]string{
		"name": a.Name,
	})
	if err != nil {
		return fmt.Errorf("marshal area update governance payload: %w", err)
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     a.TenantID,
		EventType:    domain.GovernanceEventTypeAreaUpdated,
		ActorUserID:  actorUserID,
		ResourceType: "process_area",
		ResourceID:   string(a.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("write area update governance event: %w", err)
	}
	return nil
}

func (s *AreaService) SetParent(ctx context.Context, tenantID, areaCode string, parentCode *string, actorID string) error {
	tx, err := s.areas.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin set area parent tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	area, err := s.areas.GetByCodeForUpdate(ctx, tx, tenantID, areaCode)
	if err != nil {
		return fmt.Errorf("load area for parent update: %w", err)
	}
	if !area.IsActive() {
		return domain.ErrAreaArchived
	}

	if parentCode != nil {
		if *parentCode == areaCode {
			return domain.ErrAreaParentCycle
		}
		if _, err := s.areas.GetByCodeForUpdate(ctx, tx, tenantID, *parentCode); err != nil {
			return fmt.Errorf("load parent area for parent update: %w", err)
		}
		ancestors, err := s.areas.ListAncestorsTx(ctx, tx, tenantID, *parentCode)
		if err != nil {
			return fmt.Errorf("list area ancestors for parent update: %w", err)
		}
		for _, ancestorCode := range ancestors {
			if ancestorCode == areaCode {
				return domain.ErrAreaParentCycle
			}
		}
	}

	area.ParentCode = parentCode
	if err := s.areas.UpdateTx(ctx, tx, area); err != nil {
		return fmt.Errorf("persist area parent update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit set area parent tx: %w", err)
	}
	committed = true
	return s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeAreaParentChanged,
		ActorUserID:  actorID,
		ResourceType: "process_area",
		ResourceID:   areaCode,
		PayloadJSON:  []byte(`{}`),
	})
}

func (s *AreaService) Archive(ctx context.Context, tenantID, areaCode, actorID string) error {
	tx, err := s.areas.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin archive area tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	area, err := s.areas.GetByCodeForUpdate(ctx, tx, tenantID, areaCode)
	if err != nil {
		return fmt.Errorf("load area for archive: %w", err)
	}
	if err := area.Archive(s.now()); err != nil {
		return err
	}
	if err := s.areas.UpdateTx(ctx, tx, area); err != nil {
		return fmt.Errorf("persist area archive: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit archive area tx: %w", err)
	}
	committed = true
	return s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeAreaArchived,
		ActorUserID:  actorID,
		ResourceType: "process_area",
		ResourceID:   areaCode,
		PayloadJSON:  []byte(`{}`),
	})
}
