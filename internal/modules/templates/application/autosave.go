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
		return nil, wrapAppErr("templates presign upload: get template", err)
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, wrapAppErr("templates presign upload: get version", err)
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	key := version.DocxStorageKey
	url, err := s.presign.PresignPUT(ctx, key, autosaveUploadTTL)
	if err != nil {
		return nil, fmt.Errorf("templates presign upload: presign put: %w", err)
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

	url, err := s.presign.PresignPUT(ctx, version.DocxStorageKey, autosaveUploadTTL)
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
		return wrapAppErr("templates save draft: get template", err)
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return wrapAppErr("templates save draft: get version", err)
	}
	if version.Status != domain.VersionStatusDraft {
		return domain.ErrInvalidStateTransition
	}
	audit, err := newAuditEvent(
		cmd.TenantID,
		cmd.TemplateID,
		cmd.ActorUserID,
		&version.ID,
		domain.AuditSaved,
		map[string]any{
			"docx_content_hash":   cmd.DocxContentHash,
			"schema_content_hash": cmd.SchemaContentHash,
			"schema_storage_key":  cmd.SchemaStorageKey,
			"expected_lock":       cmd.ExpectedLockVersion,
		},
		s.clock.Now(),
	)
	if err != nil {
		return err
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
		if err := s.repo.UpdateVersionDraftCASTx(ctx, tx, cmd.TenantID, version.ID, cmd.ExpectedLockVersion, cmd.DocxStorageKey, cmd.DocxContentHash); err != nil {
			return wrapAppErr("templates save draft: update draft", err)
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates save draft: append audit", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("templates save draft: commit tx: %w", err)
		}
		return nil
	}
	if err := s.repo.UpdateVersionDraftCAS(ctx, cmd.TenantID, version.ID, cmd.ExpectedLockVersion, cmd.DocxStorageKey, cmd.DocxContentHash); err != nil {
		return wrapAppErr("templates save draft: update draft", err)
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return wrapAppErr("templates save draft: append audit", err)
	}
	return nil
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

	actualHash, err := s.presign.HeadContentHash(ctx, version.DocxStorageKey)
	if err != nil {
		if errors.Is(err, domain.ErrUploadMissing) {
			return nil, domain.ErrUploadMissing
		}
		return nil, fmt.Errorf("templates commit autosave: head content hash: %w", err)
	}
	if actualHash != cmd.ExpectedContentHash {
		if err := s.presign.Delete(ctx, version.DocxStorageKey); err != nil {
			return nil, errors.Join(domain.ErrContentHashMismatch, fmt.Errorf("delete mismatched upload: %w", err))
		}
		return nil, domain.ErrContentHashMismatch
	}

	version.ContentHash = actualHash
	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditSaved, map[string]any{"content_hash": actualHash}, s.clock.Now())
	if err != nil {
		return nil, err
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
		if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
			return nil, wrapAppErr("templates commit autosave: update version", err)
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return nil, wrapAppErr("templates commit autosave: append audit", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("templates commit autosave: commit tx: %w", err)
		}
		return version, nil
	}
	if err := s.repo.UpdateVersion(ctx, cmd.TenantID, version); err != nil {
		return nil, wrapAppErr("templates commit autosave: update version", err)
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return nil, wrapAppErr("templates commit autosave: append audit", err)
	}

	return version, nil
}
