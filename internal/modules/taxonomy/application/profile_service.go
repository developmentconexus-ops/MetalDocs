package application

import (
	"context"
	"encoding/json"
	"time"

	"metaldocs/internal/modules/taxonomy/domain"
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
	return &ProfileService{profiles: profiles, tplCheck: tplCheck, govLogger: govLogger, now: time.Now}
}

func (s *ProfileService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error) {
	return s.profiles.List(ctx, tenantID, includeArchived)
}

func (s *ProfileService) Get(ctx context.Context, tenantID, code string) (*domain.DocumentProfile, error) {
	return s.profiles.GetByCode(ctx, tenantID, code)
}

func (s *ProfileService) Create(ctx context.Context, p *domain.DocumentProfile) error {
	if err := s.profiles.Create(ctx, p); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"code": p.Code,
		"name": p.Name,
	})
	_ = s.govLogger.Log(ctx, domain.GovernanceEvent{
		EventType:    "profile.created",
		ResourceType: "document_profile",
		ResourceID:   p.Code,
		PayloadJSON:  payload,
	})
	return nil
}

func (s *ProfileService) Update(ctx context.Context, p *domain.DocumentProfile) error {
	if err := s.profiles.Update(ctx, p); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"code": p.Code,
		"name": p.Name,
	})
	_ = s.govLogger.Log(ctx, domain.GovernanceEvent{
		EventType:    "profile.updated",
		ResourceType: "document_profile",
		ResourceID:   p.Code,
		PayloadJSON:  payload,
	})
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

	payload, _ := json.Marshal(map[string]string{
		"template_version_id": templateVersionID,
	})
	return s.govLogger.Log(ctx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    "profile.default_template_change",
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
		EventType:    "profile.archived",
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   profileCode,
		PayloadJSON:  []byte(`{}`),
	})
}
