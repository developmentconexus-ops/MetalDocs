package domain

import (
	"errors"
	"time"
)

type DocumentFamily struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

var (
	ErrFamilyNotFound        = errors.New("family not found")
	ErrFamilyAlreadyInactive = errors.New("family is already inactive")
	ErrFamilyCodeImmutable   = errors.New("family code is immutable")
	ErrFamilyHasProfiles     = errors.New("family has active profiles and cannot be deactivated")
)

func (f *DocumentFamily) Deactivate() error {
	if !f.IsActive {
		return ErrFamilyAlreadyInactive
	}
	f.IsActive = false
	return nil
}
