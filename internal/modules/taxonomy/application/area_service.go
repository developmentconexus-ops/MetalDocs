package application

import (
	"context"
	"encoding/json"
	"time"

	"metaldocs/internal/modules/taxonomy/domain"
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
	if err := s.areas.Create(ctx, a); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"name": a.Name,
	})
	_ = s.govLogger.Log(ctx, domain.GovernanceEvent{
		EventType:    "area.created",
		ResourceType: "process_area",
		ResourceID:   a.Code,
		PayloadJSON:  payload,
	})
	return nil
}

func (s *AreaService) Update(ctx context.Context, a *domain.ProcessArea) error {
	if err := s.areas.Update(ctx, a); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"name": a.Name,
	})
	_ = s.govLogger.Log(ctx, domain.GovernanceEvent{
		EventType:    "area.updated",
		ResourceType: "process_area",
		ResourceID:   a.Code,
		PayloadJSON:  payload,
	})
	return nil
}

func (s *AreaService) SetParent(ctx context.Context, tenantID, areaCode string, parentCode *string, actorID string) error {
	tx, err := s.areas.BeginTx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	area, err := s.areas.GetByCodeForUpdate(ctx, tx, tenantID, areaCode)
	if err != nil {
		return err
	}
	if !area.IsActive() {
		return domain.ErrAreaArchived
	}

	if parentCode != nil {
		if *parentCode == areaCode {
			return domain.ErrAreaParentCycle
		}
		if _, err := s.areas.GetByCodeForUpdate(ctx, tx, tenantID, *parentCode); err != nil {
			return err
		}
		ancestors, err := s.areas.ListAncestorsTx(ctx, tx, tenantID, *parentCode)
		if err != nil {
			return err
		}
		for _, ancestorCode := range ancestors {
			if ancestorCode == areaCode {
				return domain.ErrAreaParentCycle
			}
		}
	}

	area.ParentCode = parentCode
	if err := s.areas.UpdateTx(ctx, tx, area); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    "area.parent_changed",
		ActorUserID:  actorID,
		ResourceType: "process_area",
		ResourceID:   areaCode,
		PayloadJSON:  []byte(`{}`),
	})
}

func (s *AreaService) Archive(ctx context.Context, tenantID, areaCode, actorID string) error {
	tx, err := s.areas.BeginTx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	area, err := s.areas.GetByCodeForUpdate(ctx, tx, tenantID, areaCode)
	if err != nil {
		return err
	}
	if err := area.Archive(s.now()); err != nil {
		return err
	}
	if err := s.areas.UpdateTx(ctx, tx, area); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    "area.archived",
		ActorUserID:  actorID,
		ResourceType: "process_area",
		ResourceID:   areaCode,
		PayloadJSON:  []byte(`{}`),
	})
}
