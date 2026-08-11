package application

import (
	"context"
	"fmt"
	"strings"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/tenant"
)

// FamilyService is the use-case orchestrator for DocumentFamily: List, Get,
// Create, Update, Deactivate. Every mutating method emits a governance event
// via govLogger inside the same transaction as the mutation (T-004/T-005
// closed).
type FamilyService struct {
	families  domain.FamilyRepository
	govLogger domain.GovernanceLogger
}

// NewFamilyService builds a FamilyService. It panics if govLogger is nil —
// taxonomy has no silent-audit-loss fallback (fail-loud by design).
func NewFamilyService(families domain.FamilyRepository, govLogger domain.GovernanceLogger) *FamilyService {
	if govLogger == nil {
		panic("taxonomy: FamilyService requires a non-nil GovernanceLogger")
	}
	return &FamilyService{families: families, govLogger: govLogger}
}

// List returns families for tenantID, including inactive ones when
// includeInactive is true.
func (s *FamilyService) List(ctx context.Context, tenantID string, includeInactive bool) ([]domain.DocumentFamily, error) {
	families, err := s.families.List(ctx, tenantID, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: list families: %w", err)
	}
	return families, nil
}

// Get returns the family identified by (tenantID, code), or a wrapped
// domain.ErrFamilyNotFound if no such row exists.
func (s *FamilyService) Get(ctx context.Context, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	family, err := s.families.GetByCode(ctx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: get family %q: %w", code, err)
	}
	return family, nil
}

// Create validates f, inserts it, and logs a family.created governance
// event, all inside a single transaction (BeginTx...Commit); any failure
// rolls the transaction back so no partial family/event pair is persisted.
func (s *FamilyService) Create(ctx context.Context, f *domain.DocumentFamily) error {
	newFamily, err := domain.NewDocumentFamily(*f)
	if err != nil {
		return fmt.Errorf("taxonomy: validate family create: %w", err)
	}
	// A3.3 (T1): ActorUserID is the governance-event attribution for this
	// mutation, so an absent actor is a precondition failure, not a late one.
	// It resolves here — after the purely local validation above, before
	// BeginTx — so a request with no principal opens no transaction and calls
	// no repository method at all. "Rolled back" is a weaker property than
	// "never started".
	actorUserID, err := authn.RequireUserID(ctx)
	if err != nil {
		return err
	}
	tx, err := s.families.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin family create tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := s.families.CreateTx(ctx, tx, newFamily); err != nil {
		return fmt.Errorf("taxonomy: create family %q: %w", newFamily.Code, err)
	}
	payload, err := marshalGovernancePayload(map[string]string{"code": string(newFamily.Code), "name": newFamily.Name})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal family create governance payload: %w", err)
	}
	tenantID, _ := tenant.FromContext(ctx)
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeFamilyCreated,
		ActorUserID:  actorUserID,
		ResourceType: "document_family",
		ResourceID:   string(newFamily.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log family create governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit family create tx: %w", err)
	}
	committed = true
	return nil
}

// Update locks the family row (GetByCodeForUpdate FOR UPDATE), applies the
// Name/Description changes from f, persists them, and logs a
// family.updated governance event — all inside one transaction. Code is
// never mutated (immutability is DB-enforced regardless). Returns
// domain.ErrFamilyNameRequired if f.Name is blank.
func (s *FamilyService) Update(ctx context.Context, f *domain.DocumentFamily) (*domain.DocumentFamily, error) {
	if strings.TrimSpace(f.Name) == "" {
		return nil, domain.ErrFamilyNameRequired
	}
	// A3.3 (T1): resolved before BeginTx / GetByCodeForUpdate, so an actorless
	// request never takes the FOR UPDATE row lock. See Create.
	actorUserID, err := authn.RequireUserID(ctx)
	if err != nil {
		return nil, err
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

	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: resolve tenant for family update: %w", err)
	}
	existing, err := s.families.GetByCodeForUpdate(ctx, tx, tenantID, f.Code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: lock family %q: %w", f.Code, err)
	}
	normalized, err := domain.NewDocumentFamily(domain.DocumentFamily{
		Code:        existing.Code,
		TenantID:    existing.TenantID,
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
	payload, err := marshalGovernancePayload(map[string]string{"code": string(existing.Code), "name": existing.Name})
	if err != nil {
		return nil, fmt.Errorf("taxonomy: marshal family update governance payload: %w", err)
	}
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeFamilyUpdated,
		ActorUserID:  actorUserID,
		ResourceType: "document_family",
		ResourceID:   string(existing.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return nil, fmt.Errorf("taxonomy: log family update governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("taxonomy: commit update family tx: %w", err)
	}
	committed = true
	return existing, nil
}

// Deactivate locks the family row (FOR UPDATE), checks HasActiveProfilesTx
// (tenant-scoped, same tx) to reject with domain.ErrFamilyHasProfiles if any
// profile still references the family, then flips is_active and logs a
// family.deactivated governance event — all inside one transaction
// (T-007: closes the TOCTOU window between the check and the write).
func (s *FamilyService) Deactivate(ctx context.Context, code domain.FamilyCode) error {
	// A3.3 (T1): this method has no local input validation to run first, so the
	// actor is the very first thing resolved. Before this it was read after the
	// row lock, the HasActiveProfilesTx check and f.Deactivate() — an actorless
	// request could reach a domain mutation before anyone asked who was making
	// it. Now it opens no transaction at all.
	actorUserID, err := authn.RequireUserID(ctx)
	if err != nil {
		return err
	}
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

	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: resolve tenant for family deactivate: %w", err)
	}
	f, err := s.families.GetByCodeForUpdate(ctx, tx, tenantID, code)
	if err != nil {
		return fmt.Errorf("taxonomy: lock family %q: %w", code, err)
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
	payload, err := marshalGovernancePayload(map[string]string{"code": string(code)})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal family deactivate governance payload: %w", err)
	}
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeFamilyDeactivated,
		ActorUserID:  actorUserID,
		ResourceType: "document_family",
		ResourceID:   string(code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log family deactivate governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit deactivate family tx: %w", err)
	}
	committed = true
	return nil
}
