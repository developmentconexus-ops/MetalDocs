// Package application holds the taxonomy module's use-case services
// (AreaService, FamilyService, ProfileService): thin orchestrators that
// validate input via domain constructors, resolve the acting user, and
// delegate persistence to the domain repository ports.
package application

import (
	"context"
	"fmt"
	"time"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
)

// AreaService is the use-case orchestrator for ProcessArea: List, Get,
// Create, Update, Archive. Every mutating method emits a governance event
// via govLogger inside the same transaction as the mutation (T-005 closed).
type AreaService struct {
	areas     domain.AreaRepository
	govLogger domain.GovernanceLogger
	now       func() time.Time
}

// NewAreaService builds an AreaService with now defaulted to time.Now.
func NewAreaService(areas domain.AreaRepository, govLogger domain.GovernanceLogger) *AreaService {
	return &AreaService{
		areas:     areas,
		govLogger: govLogger,
		now:       time.Now,
	}
}

// List returns process areas for tenantID, including archived ones when
// includeArchived is true.
func (s *AreaService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.ProcessArea, error) {
	areas, err := s.areas.List(ctx, tenantID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: list areas: %w", err)
	}
	return areas, nil
}

// Get returns the area identified by (tenantID, code), or a wrapped
// domain.ErrAreaNotFound if no such row exists.
func (s *AreaService) Get(ctx context.Context, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	area, err := s.areas.GetByCode(ctx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: get area %q: %w", code, err)
	}
	return area, nil
}

// Create validates a, inserts it, and logs an area.created governance
// event, all inside a single transaction (BeginTx...Commit); any failure
// rolls the transaction back so no partial area/event pair is persisted.
func (s *AreaService) Create(ctx context.Context, a *domain.ProcessArea) error {
	newArea, err := domain.NewProcessArea(*a)
	if err != nil {
		return fmt.Errorf("taxonomy: validate area create: %w", err)
	}
	// A3.3 (T1): ActorUserID is the governance-event attribution for this
	// mutation, so an absent actor is a precondition failure, not a late one.
	// It resolves here — after the purely local validation above, before
	// BeginTx — so a request with no principal opens no transaction, takes no
	// row lock and calls no repository method at all. Resolving it next to the
	// GovernanceEvent below was still fail-closed (the tx rolled back), but
	// "rolled back" is a weaker property than "never started".
	actorUserID, err := authn.RequireUserID(ctx)
	if err != nil {
		return err
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
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
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

// Update locks the area row (GetByCodeForUpdate FOR UPDATE), applies the
// normalized fields from a (Name, Description, ParentCode, OwnerUserID,
// DefaultApproverRole, ArchivedAt), persists them, and logs an
// area.updated governance event — all inside one transaction. Code is
// never mutated (immutability is DB-enforced regardless). Note: this
// method does not run the ListAncestors cycle check on ParentCode changes
// (T-016 — the guarded SetParent entrypoint was deleted as dead code).
func (s *AreaService) Update(ctx context.Context, a *domain.ProcessArea) error {
	normalized, err := domain.NewProcessArea(*a)
	if err != nil {
		return fmt.Errorf("taxonomy: validate area update: %w", err)
	}
	// A3.3 (T1): resolved before BeginTx / GetByCodeForUpdate, so an actorless
	// request never takes the FOR UPDATE row lock. See Create.
	actorUserID, err := authn.RequireUserID(ctx)
	if err != nil {
		return err
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
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
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

// Archive locks the area row (FOR UPDATE), soft-archives it via
// ProcessArea.Archive, persists the change, and logs an area.archived
// governance event — all inside one transaction. Returns
// domain.ErrAreaArchived if the area is already archived.
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
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
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
