package application

import (
	"context"
	"fmt"
	"log"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
)

type SubmitForReviewCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
}

func (s *Service) SubmitForReview(ctx context.Context, cmd SubmitForReviewCmd) (*domain.TemplateVersion, error) {
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	if template.SystemOwned {
		return nil, domain.ErrSystemTemplateImmutable
	}

	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}
	if version.ContentHash == "" {
		return nil, domain.ErrUploadMissing
	}

	config, err := s.repo.GetApprovalConfig(ctx, cmd.TenantID, cmd.TemplateID)
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
		// Validation reads above still happen before the write transaction; this tx
		// only serializes the state update itself in the current implementation.
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates submit: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return nil, fmt.Errorf("templates submit: setAuthzGUC: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateSubmit), "tenant"); err != nil {
			return nil, fmt.Errorf("templates submit: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else if err := s.repo.UpdateVersion(ctx, cmd.TenantID, version); err != nil {
		return nil, err
	}

	audit, err := newAuditEvent(
		cmd.TenantID,
		cmd.TemplateID,
		cmd.ActorUserID,
		&version.ID,
		domain.AuditSubmitted,
		map[string]any{
			"reviewer_role": config.ReviewerRole,
			"approver_role": config.ApproverRole,
		},
		s.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return nil, wrapAppErr("templates submit: append audit", err)
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
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	if template.SystemOwned {
		return nil, domain.ErrSystemTemplateImmutable
	}

	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
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
	if err := domain.CheckSegregation(domain.SegregationRoleReviewer, cmd.ActorUserID, version.AuthorID, nil); err != nil {
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

		if err := s.updateVersionWithAuthz(ctx, cmd.TenantID, cmd.ActorUserID, version, string(iamdomain.CapTemplateReview)); err != nil {
			return nil, err
		}
		audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditReviewed, nil, s.clock.Now())
		if err != nil {
			return nil, err
		}
		if err := s.repo.AppendAudit(ctx, audit); err != nil {
			return nil, wrapAppErr("templates review: append audit", err)
		}
		return version, nil
	}

	if err := version.CanTransition(domain.VersionStatusDraft, true); err != nil {
		return nil, err
	}
	version.Status = domain.VersionStatusDraft
	version.SubmittedAt = nil

	if err := s.updateVersionWithAuthz(ctx, cmd.TenantID, cmd.ActorUserID, version, string(iamdomain.CapTemplateReview)); err != nil {
		return nil, err
	}
	audit, err := newAuditEvent(
		cmd.TenantID,
		cmd.TemplateID,
		cmd.ActorUserID,
		&version.ID,
		domain.AuditRejected,
		map[string]any{
			"reason": cmd.Reason,
			"stage":  "reviewer",
		},
		s.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return nil, wrapAppErr("templates review: append audit", err)
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
	if template.SystemOwned {
		return nil, domain.ErrSystemTemplateImmutable
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
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

	if role := version.RoleBindingFor(domain.VersionStatusPublished); role != "" && !containsRole(cmd.ActorRoles, role) {
		return nil, domain.ErrForbiddenRole
	}
	if err := domain.CheckSegregation(domain.SegregationRoleApprover, cmd.ActorUserID, version.AuthorID, version.ReviewerID); err != nil {
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
		approvedNum := version.VersionNumber
		template.PublishedVersionNumber = &approvedNum
		// ADR 0013: surface the 0-based regulated revision code alongside the
		// 1-based lifecycle counter. FE renders REV{nn} directly from this field.
		approvedRev := version.RevisionNumber
		template.CurrentRevisionNumber = &approvedRev

		if s.db != nil {
			tx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				return nil, fmt.Errorf("templates approve: begin tx: %w", err)
			}
			defer func() { _ = tx.Rollback() }()
			if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
				return nil, fmt.Errorf("templates approve: setAuthzGUC: %w", err)
			}
			if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateApprove), "tenant"); err != nil {
				return nil, fmt.Errorf("templates approve: authz: %w", err)
			}
			if err := s.repo.ObsoletePreviousPublishedTx(ctx, tx, cmd.TemplateID, version.ID); err != nil {
				return nil, err
			}
			if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
				return nil, err
			}
			if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
				return nil, err
			}
			audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditPublished, nil, s.clock.Now())
			if err != nil {
				return nil, err
			}
			if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
				return nil, wrapAppErr("templates approve: append audit", err)
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
			if err := s.repo.UpdateVersion(ctx, cmd.TenantID, version); err != nil {
				return nil, err
			}
			audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditPublished, nil, s.clock.Now())
			if err != nil {
				return nil, err
			}
			if err := s.repo.AppendAudit(ctx, audit); err != nil {
				return nil, wrapAppErr("templates approve: append audit", err)
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

	if err := s.updateVersionWithAuthz(ctx, cmd.TenantID, cmd.ActorUserID, version, string(iamdomain.CapTemplateApprove)); err != nil {
		return nil, err
	}
	audit, err := newAuditEvent(
		cmd.TenantID,
		cmd.TemplateID,
		cmd.ActorUserID,
		&version.ID,
		domain.AuditRejected,
		map[string]any{
			"reason": cmd.Reason,
			"stage":  "approver",
		},
		s.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return nil, wrapAppErr("templates approve: append audit", err)
	}

	return version, nil
}

type ArchiveCmd struct {
	TenantID, ActorUserID, TemplateID string
}

type PublishTemplateVersionCmd struct {
	TenantID, ActorUserID, TemplateID string
	ActorRoles                        []string
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
	if template.SystemOwned {
		return nil, domain.ErrSystemTemplateImmutable
	}
	version, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
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
	if err := domain.CheckSegregation(domain.SegregationRoleApprover, cmd.ActorUserID, version.AuthorID, version.ReviewerID); err != nil {
		return nil, err
	}
	// T-004 (residual): Tier 2 authz — actor must hold the version's pending
	// approver role binding, mirroring Service.Approve. Capability (Tier 1) is
	// enforced inside the tx below; both must pass for the publish to commit.
	if role := version.RoleBindingFor(domain.VersionStatusPublished); role != "" && !containsRole(cmd.ActorRoles, role) {
		denied, auditErr := newAuditEvent(
			cmd.TenantID,
			cmd.TemplateID,
			cmd.ActorUserID,
			&version.ID,
			domain.AuditPublishForbiddenRole,
			map[string]any{
				"required_role": role,
				"actor_roles":   cmd.ActorRoles,
			},
			s.clock.Now(),
		)
		if auditErr != nil {
			return nil, wrapAppErr("templates publish: build denied audit", auditErr)
		}
		// Audit infrastructure errors must not swallow the security denial:
		// the caller (and the HTTP layer) must always observe ErrForbiddenRole
		// so the actor sees a stable 403 forbidden_role response. A failed
		// audit append is best-effort observability and logged separately.
		if appendErr := s.repo.AppendAudit(ctx, denied); appendErr != nil {
			log.Printf("templates publish: append denied audit failed: %v", appendErr)
		}
		return nil, domain.ErrForbiddenRole
	}
	if s.db == nil {
		return nil, domain.ErrTransactionRequired
	}

	now := s.clock.Now()
	version.Status = domain.VersionStatusPublished
	version.DocxStorageKey = cmd.DocxKey
	version.PublishedAt = &now
	version.ApprovedAt = &now
	template.PublishedVersionID = &version.ID
	publishedNum := version.VersionNumber
	template.PublishedVersionNumber = &publishedNum
	// ADR 0013: see Approve Accept branch.
	publishedRev := version.RevisionNumber
	template.CurrentRevisionNumber = &publishedRev

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("templates publish: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
		return nil, fmt.Errorf("templates publish: setAuthzGUC: %w", err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTemplatePublish), "tenant"); err != nil {
		return nil, fmt.Errorf("templates publish: authz: %w", err)
	}
	if err := s.repo.ObsoletePreviousPublishedTx(ctx, tx, cmd.TemplateID, version.ID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
		return nil, err
	}
	nextNum := template.LatestVersion + 1
	next := s.buildNextDraftVersion(cmd.TemplateID, cmd.ActorUserID, nextNum, version)
	if err := s.repo.CreateVersionTx(ctx, tx, next); err != nil {
		return nil, err
	}
	template.LatestVersion = nextNum
	if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
		return nil, err
	}
	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditPublished, map[string]any{"schema_key": cmd.SchemaKey}, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
		return nil, wrapAppErr("templates publish: append audit", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &PublishTemplateVersionResult{PublishedVersion: version, NextDraft: next}, nil
}

func (s *Service) buildNextDraftVersion(templateID, actorID string, nextNum int, published *domain.TemplateVersion) *domain.TemplateVersion {
	return domain.NewTemplateVersionDraft(
		s.uuid.New(),
		templateID,
		actorID,
		published.DocxStorageKey,
		nextNum,
		published.MetadataSchema,
		published.PlaceholderSchema,
		s.clock.Now(),
	)
}

func (s *Service) ArchiveTemplate(ctx context.Context, cmd ArchiveCmd) (*domain.Template, error) {
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	if template.SystemOwned {
		return nil, domain.ErrSystemTemplateImmutable
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
		if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return nil, fmt.Errorf("templates archive: setAuthzGUC: %w", err)
		}
		if err := authz.Require(ctx, tx, "template.archive", "tenant"); err != nil {
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

	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, nil, domain.AuditArchived, nil, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return nil, wrapAppErr("templates archive: append audit", err)
	}

	return template, nil
}

func (s *Service) updateVersionWithAuthz(ctx context.Context, tenantID, actorID string, version *domain.TemplateVersion, cap string) error {
	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("templates update version: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := setAuthzGUC(ctx, tx, tenantID, actorID); err != nil {
			return fmt.Errorf("templates update version: setAuthzGUC: %w", err)
		}
		if err := authz.Require(ctx, tx, cap, "tenant"); err != nil {
			return fmt.Errorf("templates update version: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, tenantID, version); err != nil {
			return err
		}
		return tx.Commit()
	}
	return s.repo.UpdateVersion(ctx, tenantID, version)
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
