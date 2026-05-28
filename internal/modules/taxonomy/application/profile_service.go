package application

import (
	"context"
	"fmt"
	"time"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
)

type TemplateVersionChecker interface {
	IsPublished(ctx context.Context, versionID string) (bool, string, error)
}

type ProfileService struct {
	profiles  domain.ProfileRepository
	tplCheck  TemplateVersionChecker
	govLogger domain.GovernanceLogger
	now       func() time.Time
}

func NewProfileService(
	profiles domain.ProfileRepository,
	tplCheck TemplateVersionChecker,
	govLogger domain.GovernanceLogger,
) *ProfileService {
	if govLogger == nil {
		panic("taxonomy: ProfileService govLogger must not be nil")
	}
	if tplCheck == nil {
		panic("taxonomy: ProfileService template checker must not be nil")
	}
	return &ProfileService{profiles: profiles, tplCheck: tplCheck, govLogger: govLogger, now: time.Now}
}

func (s *ProfileService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error) {
	profiles, err := s.profiles.List(ctx, tenantID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: list profiles: %w", err)
	}
	return profiles, nil
}

func (s *ProfileService) Get(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	profile, err := s.profiles.GetByCode(ctx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: get profile %q: %w", code, err)
	}
	return profile, nil
}

func (s *ProfileService) Create(ctx context.Context, p *domain.DocumentProfile) error {
	newProfile, err := domain.NewDocumentProfile(*p)
	if err != nil {
		return fmt.Errorf("taxonomy: validate profile create: %w", err)
	}
	if err := s.profiles.Create(ctx, newProfile); err != nil {
		return fmt.Errorf("taxonomy: create profile %q: %w", newProfile.Code, err)
	}

	payload, err := marshalGovernancePayload(map[string]string{
		"code": string(newProfile.Code),
		"name": newProfile.Name,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal profile create governance payload: %w", err)
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     newProfile.TenantID,
		EventType:    domain.GovernanceEventTypeProfileCreated,
		ActorUserID:  actorUserID,
		ResourceType: "document_profile",
		ResourceID:   string(newProfile.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log profile create governance event: %w", err)
	}
	return nil
}

func (s *ProfileService) Update(ctx context.Context, p *domain.DocumentProfile) error {
	if err := s.profiles.Update(ctx, p); err != nil {
		return fmt.Errorf("taxonomy: update profile %q: %w", p.Code, err)
	}

	payload, err := marshalGovernancePayload(map[string]string{
		"code": string(p.Code),
		"name": p.Name,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal profile update governance payload: %w", err)
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     p.TenantID,
		EventType:    domain.GovernanceEventTypeProfileUpdated,
		ActorUserID:  actorUserID,
		ResourceType: "document_profile",
		ResourceID:   string(p.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log profile update governance event: %w", err)
	}
	return nil
}

func (s *ProfileService) SetDefaultTemplate(ctx context.Context, tenantID string, profileCode domain.ProfileCode, templateVersionID, actorID string) error {
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin set default template tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	profile, err := s.profiles.GetByCodeForUpdate(ctx, tx, tenantID, profileCode)
	if err != nil {
		return fmt.Errorf("taxonomy: lock profile %q: %w", profileCode, err)
	}
	if !profile.IsActive() {
		return domain.ErrProfileArchived
	}

	published, templateProfileCode, err := s.tplCheck.IsPublished(ctx, templateVersionID)
	if err != nil {
		return fmt.Errorf("taxonomy: check template version %q: %w", templateVersionID, err)
	}
	if !published {
		return domain.ErrTemplateNotPublished
	}
	if templateProfileCode != string(profileCode) {
		return domain.ErrTemplateProfileMismatch
	}

	profile.DefaultTemplateVersionID = &templateVersionID
	if err := s.profiles.UpdateTx(ctx, tx, profile); err != nil {
		return fmt.Errorf("taxonomy: update profile default template %q: %w", profileCode, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit set default template tx: %w", err)
	}
	committed = true

	payload, err := marshalGovernancePayload(map[string]string{
		"template_version_id": templateVersionID,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal default template governance payload: %w", err)
	}
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeProfileDefaultTemplateChange,
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   string(profileCode),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log default template governance event: %w", err)
	}
	return nil
}

func (s *ProfileService) Archive(ctx context.Context, tenantID string, profileCode domain.ProfileCode, actorID string) error {
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin archive profile tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	profile, err := s.profiles.GetByCodeForUpdate(ctx, tx, tenantID, profileCode)
	if err != nil {
		return fmt.Errorf("taxonomy: lock profile %q: %w", profileCode, err)
	}
	if err := profile.Archive(s.now()); err != nil {
		return fmt.Errorf("taxonomy: archive profile %q: %w", profileCode, err)
	}
	if err := s.profiles.UpdateTx(ctx, tx, profile); err != nil {
		return fmt.Errorf("taxonomy: update archived profile %q: %w", profileCode, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit archive profile tx: %w", err)
	}
	committed = true
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeProfileArchived,
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   string(profileCode),
		PayloadJSON:  []byte(`{}`),
	}); err != nil {
		return fmt.Errorf("taxonomy: log profile archive governance event: %w", err)
	}
	return nil
}
