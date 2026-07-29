package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"metaldocs/internal/platform/iamtypes"
)

// ProcessArea is a per-tenant operational area catalog entry, optionally
// nested under another area via ParentCode (self-FK). Cycle prevention on
// ParentCode reassignment is application-layer only (ListAncestors walk),
// not DB-enforced.
type ProcessArea struct {
	Code                AreaCode   `json:"code"`
	TenantID            string     `json:"tenant_id"`
	Name                string     `json:"name"`
	Description         string     `json:"description"`
	ParentCode          *AreaCode  `json:"parent_code"`
	OwnerUserID         *string    `json:"owner_user_id"`
	DefaultApproverRole *string    `json:"default_approver_role"`
	ArchivedAt          *time.Time `json:"archived_at"`
	CreatedAt           time.Time  `json:"created_at"`
}

// Sentinel errors returned by ProcessArea construction and mutation. Callers
// match these with errors.Is; they are stable across the area service and
// repository layers.
var (
	ErrAreaNotFound           = errors.New("process area not found")
	ErrAreaArchived           = errors.New("process area is archived")
	ErrAreaParentCycle        = errors.New("area parent assignment creates cycle")
	ErrAreaParentCodeRequired = errors.New("area parent code must not be empty")
	ErrAreaCodeImmutable      = errors.New("area code is immutable")
	ErrAreaCodeRequired       = errors.New("area code must not be empty")
	ErrAreaTenantRequired     = errors.New("area tenant must not be empty")
	ErrAreaNameRequired       = errors.New("area name must not be empty")
	// ErrInvalidDefaultApproverRole is returned when a non-empty
	// default_approver_role is not AREA-assignable. The vocabulary is
	// single-sourced from internal/platform/iamtypes (areaRoles) — every
	// canonical role EXCEPT system_admin, which is tenant-wide tier-1 and never
	// a user_process_areas membership. Configuring a role no user can hold in an
	// area would silently resolve to an empty approver pool, so it fails closed
	// (ADR 0022 — role strings bind to the registry, never free text).
	ErrInvalidDefaultApproverRole = errors.New("area default_approver_role must be an area-assignable role")
)

// AreaCode is the immutable primary-key identifier of a ProcessArea.
// Immutability is DB-enforced post-create by trg_process_areas_code_immutable.
type AreaCode string

// NewProcessArea validates and normalizes input into a new ProcessArea. It
// trims Code/TenantID/Name/Description/ParentCode/OwnerUserID/
// DefaultApproverRole (blank optional fields collapse to nil) and rejects
// empty Code, TenantID, or Name with the corresponding sentinel error. It
// does not perform any I/O.
func NewProcessArea(input ProcessArea) (*ProcessArea, error) {
	area := ProcessArea{
		Code:                AreaCode(strings.TrimSpace(string(input.Code))),
		TenantID:            strings.TrimSpace(input.TenantID),
		Name:                strings.TrimSpace(input.Name),
		Description:         strings.TrimSpace(input.Description),
		ParentCode:          trimOptionalAreaCode(input.ParentCode),
		OwnerUserID:         trimOptionalString(input.OwnerUserID),
		DefaultApproverRole: trimOptionalString(input.DefaultApproverRole),
		ArchivedAt:          input.ArchivedAt,
		CreatedAt:           input.CreatedAt,
	}
	if area.Code == "" {
		return nil, ErrAreaCodeRequired
	}
	if area.TenantID == "" {
		return nil, ErrAreaTenantRequired
	}
	if area.Name == "" {
		return nil, ErrAreaNameRequired
	}
	if err := ValidateDefaultApproverRole(area.DefaultApproverRole); err != nil {
		return nil, err
	}
	return &area, nil
}

// ValidateDefaultApproverRole reports whether an area's default approver role is
// area-assignable. nil (the field is optional and blank collapses to nil in
// NewProcessArea) is valid — "no default" is a legitimate state. A non-nil value
// must be in iamtypes.areaRoles, which excludes system_admin.
func ValidateDefaultApproverRole(role *string) error {
	if role == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*role)
	if trimmed == "" {
		return nil
	}
	if iamtypes.IsAreaRole(iamtypes.Role(trimmed)) {
		return nil
	}
	return fmt.Errorf("%w: got %q, want one of %v", ErrInvalidDefaultApproverRole, trimmed, iamtypes.AreaRoles())
}

func trimOptionalAreaCode(v *AreaCode) *AreaCode {
	if v == nil {
		return nil
	}
	trimmed := AreaCode(strings.TrimSpace(string(*v)))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// IsActive reports whether the area has not been archived (ArchivedAt is nil).
func (a *ProcessArea) IsActive() bool {
	return a.ArchivedAt == nil
}

// Archive soft-archives the area by stamping ArchivedAt with now (ADR 0010).
// It returns ErrAreaArchived if the area is already archived.
func (a *ProcessArea) Archive(now time.Time) error {
	if !a.IsActive() {
		return ErrAreaArchived
	}
	a.ArchivedAt = &now
	return nil
}
