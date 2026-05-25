package domain

import (
	"errors"
	"strings"
	"time"
)

type DocumentFamily struct {
	Code        FamilyCode `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
}

var (
	ErrFamilyNotFound        = errors.New("family not found")
	ErrFamilyAlreadyInactive = errors.New("family is already inactive")
	ErrFamilyHasProfiles     = errors.New("family has active profiles and cannot be deactivated")
	ErrFamilyCodeRequired    = errors.New("family code must not be empty")
	ErrFamilyNameRequired    = errors.New("family name must not be empty")
)

func NewDocumentFamily(input DocumentFamily) (*DocumentFamily, error) {
	family := DocumentFamily{
		Code:        FamilyCode(strings.TrimSpace(string(input.Code))),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		IsActive:    true,
		CreatedAt:   input.CreatedAt,
	}
	if strings.TrimSpace(string(family.Code)) == "" {
		return nil, ErrFamilyCodeRequired
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
