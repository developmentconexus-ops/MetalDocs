package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
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
	StorageKey                        string
}

func (s *Service) PresignTemplateUpload(ctx context.Context, cmd PresignTemplateUploadCmd) (*PresignAutosaveResult, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, err
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	key := version.DocxStorageKey
	url, err := s.presign.PresignPUT(ctx, key, autosaveUploadTTL)
	if err != nil {
		return nil, err
	}
	return &PresignAutosaveResult{
		UploadURL:  url,
		StorageKey: key,
		ExpiresAt:  s.clock.Now().Add(autosaveUploadTTL),
	}, nil
}

func (s *Service) PresignAutosave(ctx context.Context, cmd PresignAutosaveCmd) (*PresignAutosaveResult, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, err
	}

	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}

	url, err := s.presign.PresignPUT(ctx, version.DocxStorageKey, autosaveUploadTTL)
	if err != nil {
		return nil, err
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

type SaveTemplateDraftCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
	ExpectedLockVersion               int
	DocxStorageKey                    string
	SchemaStorageKey                  string
	DocxContentHash                   string
	SchemaContentHash                 string
}

func (s *Service) SaveTemplateDraft(ctx context.Context, cmd SaveTemplateDraftCmd) error {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return err
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return err
	}
	if version.Status != domain.VersionStatusDraft {
		return domain.ErrInvalidStateTransition
	}
	audit := &domain.AuditEvent{
		TenantID:   cmd.TenantID,
		TemplateID: cmd.TemplateID,
		VersionID:  &version.ID,
		ActorID:    cmd.ActorUserID,
		Action:     domain.AuditSaved,
		Details: map[string]any{
			"docx_content_hash":   cmd.DocxContentHash,
			"schema_content_hash": cmd.SchemaContentHash,
			"schema_storage_key":  cmd.SchemaStorageKey,
			"expected_lock":       cmd.ExpectedLockVersion,
		},
		OccurredAt: s.clock.Now(),
	}
	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("templates save draft: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return fmt.Errorf("templates save draft: set authz context: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return fmt.Errorf("templates save draft: authz: %w", err)
		}
		if err := s.repo.UpdateVersionDraftCASTx(ctx, tx, version.ID, cmd.ExpectedLockVersion, cmd.DocxStorageKey, cmd.DocxContentHash); err != nil {
			return err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return err
		}
		return tx.Commit()
	}
	if err := s.repo.UpdateVersionDraftCAS(ctx, version.ID, cmd.ExpectedLockVersion, cmd.DocxStorageKey, cmd.DocxContentHash); err != nil {
		return err
	}
	return s.repo.AppendAudit(ctx, audit)
}

func (s *Service) CommitAutosave(ctx context.Context, cmd CommitAutosaveCmd) (*domain.TemplateVersion, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, err
	}

	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}

	actualHash, err := s.presign.HeadContentHash(ctx, version.DocxStorageKey)
	if err != nil {
		if errors.Is(err, domain.ErrUploadMissing) {
			return nil, domain.ErrUploadMissing
		}
		return nil, err
	}
	if actualHash != cmd.ExpectedContentHash {
		_ = s.presign.Delete(ctx, version.DocxStorageKey)
		return nil, domain.ErrContentHashMismatch
	}

	version.ContentHash = actualHash
	audit := &domain.AuditEvent{
		TenantID:   cmd.TenantID,
		TemplateID: cmd.TemplateID,
		VersionID:  &version.ID,
		ActorID:    cmd.ActorUserID,
		Action:     domain.AuditSaved,
		Details:    map[string]any{"content_hash": actualHash},
		OccurredAt: s.clock.Now(),
	}
	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates commit autosave: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return nil, fmt.Errorf("templates commit autosave: set authz context: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return nil, fmt.Errorf("templates commit autosave: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, version); err != nil {
			return nil, err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return version, nil
	}
	if err := s.repo.UpdateVersion(ctx, version); err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return nil, err
	}

	return version, nil
}
