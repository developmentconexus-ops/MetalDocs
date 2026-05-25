package domain

import (
	"errors"
	"strings"
	"time"
)

type ProcessArea struct {
	Code                AreaCode   `json:"code"`
	TenantID            string     `json:"tenantId"`
	Name                string     `json:"name"`
	Description         string     `json:"description"`
	ParentCode          *string    `json:"parentCode"`
	OwnerUserID         *string    `json:"ownerUserId"`
	DefaultApproverRole *string    `json:"defaultApproverRole"`
	ArchivedAt          *time.Time `json:"archivedAt"`
	CreatedAt           time.Time  `json:"createdAt"`
}

var (
	ErrAreaNotFound       = errors.New("process area not found")
	ErrAreaArchived       = errors.New("process area is archived")
	ErrAreaParentCycle    = errors.New("area parent assignment creates cycle")
	ErrAreaCodeImmutable  = errors.New("area code is immutable")
	ErrAreaCodeRequired   = errors.New("area code must not be empty")
	ErrAreaTenantRequired = errors.New("area tenant must not be empty")
	ErrAreaNameRequired   = errors.New("area name must not be empty")
)

func NewProcessArea(input ProcessArea) (*ProcessArea, error) {
	area := ProcessArea{
		Code:                AreaCode(strings.TrimSpace(string(input.Code))),
		TenantID:            strings.TrimSpace(input.TenantID),
		Name:                strings.TrimSpace(input.Name),
		Description:         strings.TrimSpace(input.Description),
		ParentCode:          trimOptionalString(input.ParentCode),
		OwnerUserID:         trimOptionalString(input.OwnerUserID),
		DefaultApproverRole: trimOptionalString(input.DefaultApproverRole),
		ArchivedAt:          input.ArchivedAt,
		CreatedAt:           input.CreatedAt,
	}
	if strings.TrimSpace(string(area.Code)) == "" {
		return nil, ErrAreaCodeRequired
	}
	if area.TenantID == "" {
		return nil, ErrAreaTenantRequired
	}
	if area.Name == "" {
		return nil, ErrAreaNameRequired
	}
	return &area, nil
}

func (a *ProcessArea) IsActive() bool {
	return a.ArchivedAt == nil
}

func (a *ProcessArea) Archive(now time.Time) error {
	if !a.IsActive() {
		return ErrAreaArchived
	}
	a.ArchivedAt = &now
	return nil
}
