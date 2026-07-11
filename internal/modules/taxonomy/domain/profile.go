package domain

import (
	"errors"
	"strings"
	"time"
)

// DocumentProfile is a per-tenant document type bound to a DocumentFamily.
// It carries the optional default template version consumed by the documents
// wizard (DefaultTemplateVersionID) and is soft-archived via ArchivedAt
// (ADR 0010).
type DocumentProfile struct {
	Code                     ProfileCode `json:"code"`
	TenantID                 string      `json:"tenant_id"`
	FamilyCode               FamilyCode  `json:"family_code"`
	Name                     string      `json:"name"`
	Description              string      `json:"description"`
	Alias                    string      `json:"alias"`
	ReviewIntervalDays       int         `json:"review_interval_days"`
	DefaultTemplateVersionID *string     `json:"default_template_version_id"`
	OwnerUserID              *string     `json:"owner_user_id"`
	EditableByRole           string      `json:"editable_by_role"`
	// GovernanceClass drives the per-profile route-signature policy (G1).
	// Empty is normalized to GovernanceControlado by NewDocumentProfile,
	// mirroring the DB column default.
	GovernanceClass GovernanceClass `json:"governance_class"`
	ArchivedAt      *time.Time      `json:"archived_at"`
	CreatedAt       time.Time       `json:"created_at"`
}

// Sentinel errors returned by DocumentProfile construction and mutation.
// Callers match these with errors.Is; they are stable across the profile
// service and repository layers.
var (
	ErrProfileNotFound         = errors.New("profile not found")
	ErrProfileCodeImmutable    = errors.New("profile code is immutable")
	ErrProfileArchived         = errors.New("profile is archived")
	ErrTemplateNotPublished    = errors.New("template version is not published")
	ErrTemplateProfileMismatch = errors.New("template version belongs to different profile")
	ErrProfileCodeRequired     = errors.New("profile code must not be empty")
	ErrProfileTenantRequired   = errors.New("profile tenant must not be empty")
	ErrProfileFamilyRequired   = errors.New("profile family must not be empty")
	ErrProfileNameRequired     = errors.New("profile name must not be empty")
	ErrInvalidGovernanceClass  = errors.New("profile governance_class must be one of: controlado, simples, livre")
	// ErrClassChangeRouteConflict is returned when a reclassification would
	// leave an active approval route violating the new governance_class (e.g.
	// simples→controlado while a review-only route is active). It is the
	// friendly translation of the DB reclassify-guard trigger (P0001,
	// 'ErrClassChangeRouteConflict:' prefix) and maps to HTTP 409.
	ErrClassChangeRouteConflict = errors.New("profile reclassification conflicts with an active approval route")
	// ErrRouteViolatesProfilePolicy is the friendly translation of the DB
	// route-shape trigger (P0001, 'ErrRouteViolatesProfilePolicy:' prefix): an
	// approval route write violates its profile's governance_class. Surfaced by
	// the approval module; declared here because it originates from the
	// taxonomy-owned governance policy.
	ErrRouteViolatesProfilePolicy = errors.New("approval route violates the profile governance policy")
)

// ProfileCode is the immutable primary-key identifier of a DocumentProfile.
// Immutability is DB-enforced post-create by trg_document_profiles_code_immutable.
type ProfileCode string

// NewDocumentProfile validates and normalizes input into a new
// DocumentProfile. It trims Code/TenantID/FamilyCode/Name/Description/Alias/
// EditableByRole and rejects empty Code, TenantID, FamilyCode, or Name with
// the corresponding sentinel error. An empty Alias defaults to Code
// (truncated to 24 chars). It does not perform any I/O.
func NewDocumentProfile(input DocumentProfile) (*DocumentProfile, error) {
	profile := DocumentProfile{
		Code:                     ProfileCode(strings.TrimSpace(string(input.Code))),
		TenantID:                 strings.TrimSpace(input.TenantID),
		FamilyCode:               FamilyCode(strings.TrimSpace(string(input.FamilyCode))),
		Name:                     strings.TrimSpace(input.Name),
		Description:              strings.TrimSpace(input.Description),
		Alias:                    strings.TrimSpace(input.Alias),
		ReviewIntervalDays:       input.ReviewIntervalDays,
		DefaultTemplateVersionID: trimOptionalString(input.DefaultTemplateVersionID),
		OwnerUserID:              trimOptionalString(input.OwnerUserID),
		EditableByRole:           strings.TrimSpace(input.EditableByRole),
		GovernanceClass:          input.GovernanceClass,
		ArchivedAt:               input.ArchivedAt,
		CreatedAt:                input.CreatedAt,
	}
	if profile.Code == "" {
		return nil, ErrProfileCodeRequired
	}
	if profile.TenantID == "" {
		return nil, ErrProfileTenantRequired
	}
	if profile.FamilyCode == "" {
		return nil, ErrProfileFamilyRequired
	}
	if profile.Name == "" {
		return nil, ErrProfileNameRequired
	}
	if profile.GovernanceClass == "" {
		profile.GovernanceClass = GovernanceControlado
	}
	if err := profile.GovernanceClass.Validate(); err != nil {
		return nil, err
	}
	if profile.Alias == "" {
		profile.Alias = string(profile.Code)
		if len(profile.Alias) > 24 {
			profile.Alias = profile.Alias[:24]
		}
	}
	return &profile, nil
}

// IsActive reports whether the profile has not been archived (ArchivedAt is nil).
func (p *DocumentProfile) IsActive() bool {
	return p.ArchivedAt == nil
}

// Archive soft-archives the profile by stamping ArchivedAt with now
// (ADR 0010). It returns ErrProfileArchived if the profile is already
// archived.
func (p *DocumentProfile) Archive(now time.Time) error {
	if !p.IsActive() {
		return ErrProfileArchived
	}
	p.ArchivedAt = &now
	return nil
}
