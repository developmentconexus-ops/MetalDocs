package domain

import (
	"errors"
	"strings"
	"time"
)

// DocumentFamily is the top-level, tenant-agnostic catalog grouping (e.g.
// "Procedimento") that profiles bind to. document_families carries no
// tenant_id — a family row is visible and mutable from every tenant (T-002).
type DocumentFamily struct {
	Code        FamilyCode `json:"code"`
	TenantID    string     `json:"tenant_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	// IsActive remains an exported field because taxonomy JSON responses bind it directly.
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Sentinel errors returned by DocumentFamily construction and mutation.
// Callers match these with errors.Is; they are stable across the family
// service and repository layers.
var (
	ErrFamilyNotFound        = errors.New("family not found")
	ErrFamilyAlreadyInactive = errors.New("family is already inactive")
	ErrFamilyHasProfiles     = errors.New("family has active profiles and cannot be deactivated")
	ErrFamilyCodeRequired    = errors.New("family code must not be empty")
	ErrFamilyTenantRequired  = errors.New("family tenant must not be empty")
	ErrFamilyNameRequired    = errors.New("family name must not be empty")
)

// FamilyCode is the immutable primary-key identifier of a DocumentFamily.
// Immutability is DB-enforced post-create by trg_reject_families_code_update.
type FamilyCode string

// NewDocumentFamily validates and normalizes input into a new, active
// DocumentFamily. It trims Code/TenantID/Name/Description and rejects empty
// Code, TenantID, or Name with the corresponding sentinel error. It does not
// perform any I/O.
func NewDocumentFamily(input DocumentFamily) (*DocumentFamily, error) {
	family := DocumentFamily{
		Code:        FamilyCode(strings.TrimSpace(string(input.Code))),
		TenantID:    strings.TrimSpace(input.TenantID),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		IsActive:    true,
		CreatedAt:   input.CreatedAt,
	}
	if family.Code == "" {
		return nil, ErrFamilyCodeRequired
	}
	if family.TenantID == "" {
		return nil, ErrFamilyTenantRequired
	}
	if family.Name == "" {
		return nil, ErrFamilyNameRequired
	}
	return &family, nil
}

// Deactivate transitions the family to inactive in-memory. It returns
// ErrFamilyAlreadyInactive if the family is already inactive; it does not
// check for dependent active profiles — callers must run that check
// (FamilyRepository.HasActiveProfilesTx) before calling Deactivate.
func (f *DocumentFamily) Deactivate() error {
	if !f.IsActive {
		return ErrFamilyAlreadyInactive
	}
	f.IsActive = false
	return nil
}
