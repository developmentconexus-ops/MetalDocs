package domain

import (
	"errors"
	"strings"
	"time"
)

type DocumentProfile struct {
	Code                     ProfileCode `json:"code"`
	TenantID                 string      `json:"tenantId"`
	FamilyCode               FamilyCode  `json:"familyCode"`
	Name                     string      `json:"name"`
	Description              string      `json:"description"`
	Alias                    string      `json:"alias"`
	ReviewIntervalDays       int         `json:"reviewIntervalDays"`
	DefaultTemplateVersionID *string     `json:"defaultTemplateVersionId"`
	OwnerUserID              *string     `json:"ownerUserId"`
	EditableByRole           string      `json:"editableByRole"`
	ArchivedAt               *time.Time  `json:"archivedAt"`
	CreatedAt                time.Time   `json:"createdAt"`
}

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
)

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
		ArchivedAt:               input.ArchivedAt,
		CreatedAt:                input.CreatedAt,
	}
	if strings.TrimSpace(string(profile.Code)) == "" {
		return nil, ErrProfileCodeRequired
	}
	if profile.TenantID == "" {
		return nil, ErrProfileTenantRequired
	}
	if strings.TrimSpace(string(profile.FamilyCode)) == "" {
		return nil, ErrProfileFamilyRequired
	}
	if profile.Name == "" {
		return nil, ErrProfileNameRequired
	}
	if profile.Alias == "" {
		profile.Alias = string(profile.Code)
		if len(profile.Alias) > 24 {
			profile.Alias = profile.Alias[:24]
		}
	}
	return &profile, nil
}

func (p *DocumentProfile) IsActive() bool {
	return p.ArchivedAt == nil
}

func (p *DocumentProfile) Archive(now time.Time) error {
	if !p.IsActive() {
		return ErrProfileArchived
	}
	p.ArchivedAt = &now
	return nil
}
