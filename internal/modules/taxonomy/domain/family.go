package domain

import (
	"errors"
	"strings"
	"time"
)

type DocumentFamily struct {
	Code        FamilyCode `json:"code"`
	TenantID    string     `json:"tenant_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	// IsActive remains an exported field because taxonomy JSON responses bind it directly.
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	ErrFamilyNotFound        = errors.New("family not found")
	ErrFamilyAlreadyInactive = errors.New("family is already inactive")
	ErrFamilyHasProfiles     = errors.New("family has active profiles and cannot be deactivated")
	ErrFamilyCodeRequired    = errors.New("family code must not be empty")
	ErrFamilyTenantRequired  = errors.New("family tenant must not be empty")
	ErrFamilyNameRequired    = errors.New("family name must not be empty")
)

type FamilyCode string

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

func (f *DocumentFamily) Deactivate() error {
	if !f.IsActive {
		return ErrFamilyAlreadyInactive
	}
	f.IsActive = false
	return nil
}
