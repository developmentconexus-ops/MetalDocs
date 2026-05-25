package application

import (
	"context"
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
	return s.profiles.List(ctx, tenantID, includeArchived)
}

func (s *ProfileService) Get(ctx context.Context, tenantID, code string) (*domain.DocumentProfile, error) {
	return s.profiles.GetByCode(ctx, tenantID, code)
}

func (s *ProfileService) Create(ctx context.Context, p *domain.DocumentProfile) error {
	newProfile, err := domain.NewDocumentProfile(*p)
	if err != nil {
		return err
	}
	if err := s.profiles.Create(ctx, newProfile); err != nil {
		return err
	}

	payload, err := marshalGovernancePayload(map[string]string{
		"code": newProfile.Code,
		"name": newProfile.Name,
	})
	if err != nil {
		return err
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     newProfile.TenantID,
		EventType:    domain.GovernanceEventTypeProfileCreated,
		ActorUserID:  actorUserID,
		ResourceType: "document_profile",
		ResourceID:   newProfile.Code,
		PayloadJSON:  payload,
	}); err != nil {
		return err
	}
	return nil
}

func (s *ProfileService) Update(ctx context.Context, p *domain.DocumentProfile) error {
	if err := s.profiles.Update(ctx, p); err != nil {
		return err
	}

	payload, err := marshalGovernancePayload(map[string]string{
		"code": p.Code,
		"name": p.Name,
	})
	if err != nil {
		return err
	}
	actorUserID, _ := authn.UserIDFromContext(ctx)
	if err := s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     p.TenantID,
		EventType:    domain.GovernanceEventTypeProfileUpdated,
		ActorUserID:  actorUserID,
		ResourceType: "document_profile",
		ResourceID:   p.Code,
		PayloadJSON:  payload,
	}); err != nil {
		return err
	}
	return nil
}

func (s *ProfileService) SetDefaultTemplate(ctx context.Context, tenantID, profileCode, templateVersionID, actorID string) error {
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	profile, err := s.profiles.GetByCodeForUpdate(ctx, tx, tenantID, profileCode)
	if err != nil {
		return err
	}
	if !profile.IsActive() {
		return domain.ErrProfileArchived
	}

	published, templateProfileCode, err := s.tplCheck.IsPublished(ctx, templateVersionID)
	if err != nil {
		return err
	}
	if !published {
		return domain.ErrTemplateNotPublished
	}
	if templateProfileCode != profileCode {
		return domain.ErrTemplateProfileMismatch
	}

	profile.DefaultTemplateVersionID = &templateVersionID
	if err := s.profiles.UpdateTx(ctx, tx, profile); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	payload, err := marshalGovernancePayload(map[string]string{
		"template_version_id": templateVersionID,
	})
	if err != nil {
		return err
	}
	return s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeProfileDefaultTemplateChange,
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   profileCode,
		PayloadJSON:  payload,
	})
}

func (s *ProfileService) Archive(ctx context.Context, tenantID, profileCode, actorID string) error {
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	profile, err := s.profiles.GetByCodeForUpdate(ctx, tx, tenantID, profileCode)
	if err != nil {
		return err
	}
	if err := profile.Archive(s.now()); err != nil {
		return err
	}
	if err := s.profiles.UpdateTx(ctx, tx, profile); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeProfileArchived,
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   profileCode,
		PayloadJSON:  []byte(`{}`),
	})
}
