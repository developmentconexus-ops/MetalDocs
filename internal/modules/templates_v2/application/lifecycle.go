package application

import (
	"context"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates_v2/domain"
)

type SubmitForReviewCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
}

func (s *Service) SubmitForReview(ctx context.Context, cmd SubmitForReviewCmd) (*domain.TemplateVersion, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, err
	}

	version, err := s.repo.GetVersion(ctx, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	if version.ContentHash == "" {
		return nil, domain.ErrUploadMissing
	}

	config, err := s.repo.GetApprovalConfig(ctx, cmd.TemplateID)
	if err != nil {
		return nil, err
	}

	version.PendingReviewerRole = config.ReviewerRole
	version.PendingApproverRole = config.ApproverRole
	if err := version.CanTransition(domain.VersionStatusInReview, config.HasReviewer()); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	version.Status = domain.VersionStatusInReview
	version.SubmittedAt = &now

	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates submit: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateSubmit), "tenant"); err != nil {
			return nil, fmt.Errorf("templates submit: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, version); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else if err := s.repo.UpdateVersion(ctx, version); err != nil {
		return nil, err
	}

	if err := s.repo.AppendAudit(ctx, &domain.AuditEvent{
		TenantID:   cmd.TenantID,
		TemplateID: cmd.TemplateID,
		VersionID:  &version.ID,
		ActorID:    cmd.ActorUserID,
		Action:     domain.AuditSubmitted,
		Details: map[string]any{
			"reviewer_role": config.ReviewerRole,
			"approver_role": config.ApproverRole,
		},
		OccurredAt: s.clock.Now(),
	}); err != nil {
		return nil, err
	}

	return version, nil
}

type ReviewCmd struct {
	TenantID, ActorUserID string
	ActorRoles            []string
	TemplateID            string
	VersionNumber         int
	Accept                bool
	Reason                string
}

func (s *Service) Review(ctx context.Context, cmd ReviewCmd) (*domain.TemplateVersion, error) {
	if _, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, err
	}

	version, err := s.repo.GetVersion(ctx, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	if version.Status != domain.VersionStatusInReview {
		return nil, domain.ErrInvalidStateTransition
	}
	if version.PendingReviewerRole == nil {
		return nil, domain.ErrInvalidStateTransition
	}
	if !containsRole(cmd.ActorRoles, *version.PendingReviewerRole) {
		return nil, domain.ErrForbiddenRole
	}
	if err := domain.CheckSegregation("reviewer", cmd.ActorUserID, version.AuthorID, nil); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if cmd.Accept {
		if err := version.CanTransition(domain.VersionStatusApproved, true); err != nil {
			return nil, err
		}
		version.Status = domain.VersionStatusApproved
		version.ReviewerID = &cmd.ActorUserID
		version.ReviewedAt = &now

		if err := s.updateVersionWithAuthz(ctx, version, string(iamdomain.CapTemplateEdit)); err != nil {
			return nil, err
		}
		if err := s.repo.AppendAudit(ctx, &domain.AuditEvent{
			TenantID:   cmd.TenantID,
			TemplateID: cmd.TemplateID,
			VersionID:  &version.ID,
			ActorID:    cmd.ActorUserID,
			Action:     domain.AuditReviewed,
			Details:    map[string]any{},
			OccurredAt: s.clock.Now(),
		}); err != nil {
			return nil, err
		}
		return version, nil
	}

	if err := version.CanTransition(domain.VersionStatusDraft, true); err != nil {
		return nil, err
	}
	version.Status = domain.VersionStatusDraft
	version.SubmittedAt = nil

	if err := s.updateVersionWithAuthz(ctx, version, string(iamdomain.CapTemplateEdit)); err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, &domain.AuditEvent{
		TenantID:   cmd.TenantID,
		TemplateID: cmd.TemplateID,
		VersionID:  &version.ID,
		ActorID:    cmd.ActorUserID,
		Action:     domain.AuditRejected,
		Details: map[string]any{
			"reason": cmd.Reason,
			"stage":  "reviewer",
		},
		OccurredAt: s.clock.Now(),
	}); err != nil {
		return nil, err
	}

	return version, nil
}

type ApproveCmd struct {
	TenantID, ActorUserID string
	ActorRoles            []string
	TemplateID            string
	VersionNumber         int
	Accept                bool
	Reason                string
}

func (s *Service) Approve(ctx context.Context, cmd ApproveCmd) (*domain.TemplateVersion, error) {
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	version, err := s.repo.GetVersion(ctx, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}

	hasReviewer := version.PendingReviewerRole != nil
	if hasReviewer {
		if version.Status != domain.VersionStatusApproved {
			return nil, domain.ErrInvalidStateTransition
		}
	} else if version.Status != domain.VersionStatusInReview {
		return nil, domain.ErrInvalidStateTransition
	}

	if !containsRole(cmd.ActorRoles, version.PendingApproverRole) {
		return nil, domain.ErrForbiddenRole
	}
	if err := domain.CheckSegregation("approver", cmd.ActorUserID, version.AuthorID, version.ReviewerID); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if cmd.Accept {
		if err := version.CanTransition(domain.VersionStatusPublished, hasReviewer); err != nil {
			return nil, err
		}
		version.Status = domain.VersionStatusPublished
		version.ApproverID = &cmd.ActorUserID
		version.ApprovedAt = &now
		version.PublishedAt = &now

		template.PublishedVersionID = &version.ID

		if s.db != nil {
			tx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				return nil, fmt.Errorf("templates approve: begin tx: %w", err)
			}
			defer func() { _ = tx.Rollback() }()
			if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateApprove), "tenant"); err != nil {
				return nil, fmt.Errorf("templates approve: authz: %w", err)
			}
			if err := s.repo.ObsoletePreviousPublishedTx(ctx, tx, cmd.TemplateID, version.ID); err != nil {
				return nil, err
			}
			if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
				return nil, err
			}
			if err := s.repo.UpdateVersionTx(ctx, tx, version); err != nil {
				return nil, err
			}
			if err := s.repo.AppendAuditTx(ctx, tx, &domain.AuditEvent{
				TenantID:   cmd.TenantID,
				TemplateID: cmd.TemplateID,
				VersionID:  &version.ID,
				ActorID:    cmd.ActorUserID,
				Action:     domain.AuditPublished,
				Details:    map[string]any{},
				OccurredAt: s.clock.Now(),
			}); err != nil {
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				return nil, err
			}
		} else {
			if err := s.repo.ObsoletePreviousPublished(ctx, cmd.TemplateID, version.ID); err != nil {
				return nil, err
			}
			if err := s.repo.UpdateTemplate(ctx, template); err != nil {
				return nil, err
			}
			if err := s.repo.UpdateVersion(ctx, version); err != nil {
				return nil, err
			}
			if err := s.repo.AppendAudit(ctx, &domain.AuditEvent{
				TenantID:   cmd.TenantID,
				TemplateID: cmd.TemplateID,
				VersionID:  &version.ID,
				ActorID:    cmd.ActorUserID,
				Action:     domain.AuditPublished,
				Details:    map[string]any{},
				OccurredAt: s.clock.Now(),
			}); err != nil {
				return nil, err
			}
		}
		return version, nil
	}

	if err := version.CanTransition(domain.VersionStatusDraft, hasReviewer); err != nil {
		return nil, err
	}
	version.Status = domain.VersionStatusDraft
	version.SubmittedAt = nil
	version.ReviewedAt = nil
	version.ApprovedAt = nil

	if err := s.updateVersionWithAuthz(ctx, version, string(iamdomain.CapTemplateApprove)); err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, &domain.AuditEvent{
		TenantID:   cmd.TenantID,
		TemplateID: cmd.TemplateID,
		VersionID:  &version.ID,
		ActorID:    cmd.ActorUserID,
		Action:     domain.AuditRejected,
		Details: map[string]any{
			"reason": cmd.Reason,
			"stage":  "approver",
		},
		OccurredAt: s.clock.Now(),
	}); err != nil {
		return nil, err
	}

	return version, nil
}

type ArchiveCmd struct {
	TenantID, ActorUserID, TemplateID string
}

type PublishTemplateVersionCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
	DocxKey                           string
	SchemaKey                         string
}

type PublishTemplateVersionResult struct {
	PublishedVersion *domain.TemplateVersion
	NextDraft        *domain.TemplateVersion
}

func (s *Service) PublishTemplateVersion(ctx context.Context, cmd PublishTemplateVersionCmd) (*PublishTemplateVersionResult, error) {
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	version, err := s.repo.GetVersion(ctx, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}

	// T-004: content_hash gate — presigned docx must have been committed.
	if version.ContentHash == "" {
		return nil, domain.ErrContentHashMismatch
	}

	// T-004: SoD — publisher must not be the author or the reviewer.
	if err := domain.CheckSegregation("approver", cmd.ActorUserID, version.AuthorID, version.ReviewerID); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	version.Status = domain.VersionStatusPublished
	version.DocxStorageKey = cmd.DocxKey
	version.PublishedAt = &now
	version.ApprovedAt = &now
	template.PublishedVersionID = &version.ID

	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates publish: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplatePublish), "tenant"); err != nil {
			return nil, fmt.Errorf("templates publish: authz: %w", err)
		}
		if err := s.repo.ObsoletePreviousPublishedTx(ctx, tx, cmd.TemplateID, version.ID); err != nil {
			return nil, err
		}
		if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
			return nil, err
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, version); err != nil {
			return nil, err
		}
		nextNum := template.LatestVersion + 1
		next := &domain.TemplateVersion{
			ID:                s.uuid.New(),
			TemplateID:        cmd.TemplateID,
			VersionNumber:     nextNum,
			Status:            domain.VersionStatusDraft,
			DocxStorageKey:    fmt.Sprintf("templates/%s/versions/%d.docx", cmd.TemplateID, nextNum),
			ContentHash:       "",
			MetadataSchema:    cloneMetadataSchema(version.MetadataSchema),
			PlaceholderSchema: clonePlaceholders(version.PlaceholderSchema),
			AuthorID:          cmd.ActorUserID,
			CreatedAt:         s.clock.Now(),
		}
		if err := s.repo.CreateVersionTx(ctx, tx, next); err != nil {
			return nil, err
		}
		template.LatestVersion = nextNum
		if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
			return nil, err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, &domain.AuditEvent{
			TenantID:   cmd.TenantID,
			TemplateID: cmd.TemplateID,
			VersionID:  &version.ID,
			ActorID:    cmd.ActorUserID,
			Action:     domain.AuditPublished,
			Details:    map[string]any{"schema_key": cmd.SchemaKey},
			OccurredAt: now,
		}); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &PublishTemplateVersionResult{PublishedVersion: version, NextDraft: next}, nil
	} else {
		if err := s.repo.ObsoletePreviousPublished(ctx, cmd.TemplateID, version.ID); err != nil {
			return nil, err
		}
		if err := s.repo.UpdateTemplate(ctx, template); err != nil {
			return nil, err
		}
		if err := s.repo.UpdateVersion(ctx, version); err != nil {
			return nil, err
		}
		if err := s.repo.AppendAudit(ctx, &domain.AuditEvent{
			TenantID:   cmd.TenantID,
			TemplateID: cmd.TemplateID,
			VersionID:  &version.ID,
			ActorID:    cmd.ActorUserID,
			Action:     domain.AuditPublished,
			Details:    map[string]any{"schema_key": cmd.SchemaKey},
			OccurredAt: now,
		}); err != nil {
			return nil, err
		}
	}

	next, err := s.CreateNextVersion(ctx, CreateVersionCmd{
		TenantID:    cmd.TenantID,
		ActorUserID: cmd.ActorUserID,
		TemplateID:  cmd.TemplateID,
	})
	if err != nil {
		return nil, err
	}
	return &PublishTemplateVersionResult{PublishedVersion: version, NextDraft: next}, nil
}

func (s *Service) ArchiveTemplate(ctx context.Context, cmd ArchiveCmd) (*domain.Template, error) {
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	if template.IsArchived() {
		return template, nil
	}

	now := s.clock.Now()
	template.ArchivedAt = &now

	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates archive: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return nil, fmt.Errorf("templates archive: authz: %w", err)
		}
		if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else if err := s.repo.UpdateTemplate(ctx, template); err != nil {
		return nil, err
	}

	if err := s.repo.AppendAudit(ctx, &domain.AuditEvent{
		TenantID:   cmd.TenantID,
		TemplateID: cmd.TemplateID,
		VersionID:  nil,
		ActorID:    cmd.ActorUserID,
		Action:     domain.AuditArchived,
		Details:    map[string]any{},
		OccurredAt: s.clock.Now(),
	}); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *Service) updateVersionWithAuthz(ctx context.Context, version *domain.TemplateVersion, cap string) error {
	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("templates update version: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := authz.Require(ctx, tx, cap, "tenant"); err != nil {
			return fmt.Errorf("templates update version: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, version); err != nil {
			return err
		}
		return tx.Commit()
	}
	return s.repo.UpdateVersion(ctx, version)
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
