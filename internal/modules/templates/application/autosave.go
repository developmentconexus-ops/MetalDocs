package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/objectstore"
)

const autosaveUploadTTL = 10 * time.Minute

type PresignAutosaveCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
}

type PresignAutosaveResult struct {
	UploadURL  string
	StorageKey string
	ExpiresAt  time.Time
}

type PresignTemplateUploadCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
}

func (s *Service) PresignTemplateUpload(ctx context.Context, cmd PresignTemplateUploadCmd) (*PresignAutosaveResult, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, wrapAppErr("templates presign upload: get template", err)
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, wrapAppErr("templates presign upload: get version", err)
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	url, err := s.presign.PresignPut(ctx, cmd.TenantID, version.DocxStorageKey, autosaveUploadTTL)
	if err != nil {
		return nil, fmt.Errorf("templates presign upload: presign put: %w", err)
	}
	return &PresignAutosaveResult{
		UploadURL:  url,
		StorageKey: version.DocxStorageKey,
		ExpiresAt:  s.clock.Now().Add(autosaveUploadTTL),
	}, nil
}

func (s *Service) PresignTemplateSchemaUpload(ctx context.Context, cmd PresignTemplateUploadCmd) (*PresignAutosaveResult, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, wrapAppErr("templates presign schema upload: get template", err)
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, wrapAppErr("templates presign schema upload: get version", err)
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	key := templateSchemaKey(cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	url, err := s.presign.PresignPut(ctx, cmd.TenantID, key, autosaveUploadTTL)
	if err != nil {
		return nil, fmt.Errorf("templates presign schema upload: presign put: %w", err)
	}
	return &PresignAutosaveResult{
		UploadURL:  url,
		StorageKey: key,
		ExpiresAt:  s.clock.Now().Add(autosaveUploadTTL),
	}, nil
}

func (s *Service) PresignAutosave(ctx context.Context, cmd PresignAutosaveCmd) (*PresignAutosaveResult, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, wrapAppErr("templates presign autosave: get template", err)
	}

	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, wrapAppErr("templates presign autosave: get version", err)
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}

	url, err := s.presign.PresignPut(ctx, cmd.TenantID, version.DocxStorageKey, autosaveUploadTTL)
	if err != nil {
		return nil, fmt.Errorf("templates presign autosave: presign put: %w", err)
	}

	return &PresignAutosaveResult{
		UploadURL:  url,
		StorageKey: version.DocxStorageKey,
		ExpiresAt:  s.clock.Now().Add(autosaveUploadTTL),
	}, nil
}

type CommitAutosaveCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
	ExpectedContentHash               string
}

func (s *Service) CommitAutosave(ctx context.Context, cmd CommitAutosaveCmd) (*domain.TemplateVersion, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, wrapAppErr("templates commit autosave: get template", err)
	}

	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, wrapAppErr("templates commit autosave: get version", err)
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}

	vp, err := s.presign.Confirm(ctx, cmd.TenantID, version.DocxStorageKey, cmd.ExpectedContentHash)
	if err != nil {
		switch {
		case errors.Is(err, objectstore.ErrObjectMissing):
			return nil, domain.ErrUploadMissing
		case errors.Is(err, objectstore.ErrHashMismatch):
			return nil, domain.ErrContentHashMismatch
		case errors.Is(err, objectstore.ErrObjectTooLarge):
			return nil, domain.ErrUploadTooLarge
		default:
			return nil, fmt.Errorf("templates commit autosave: confirm: %w", err)
		}
	}

	version.ContentHash = vp.ContentHash
	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditSaved, map[string]any{"content_hash": vp.ContentHash}, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return fmt.Errorf("templates commit autosave: set authz context: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return fmt.Errorf("templates commit autosave: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
			return wrapAppErr("templates commit autosave: update version", err)
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates commit autosave: append audit", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return version, nil
}
