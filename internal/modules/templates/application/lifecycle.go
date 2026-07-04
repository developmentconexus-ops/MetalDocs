package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
)

// SubmitForReviewCmd identifies the draft template version to submit into
// the review/approval workflow.
type SubmitForReviewCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
}

// SubmitForReview transitions a draft version to under-review, snapshotting
// the template's current approval configuration onto the version's pending
// reviewer/approver role bindings. The version must be a draft with a
// committed content hash (ErrUploadMissing otherwise); the transition itself
// is validated against domain.CanTransition, which accounts for whether a
// reviewer stage is configured. Appends an AuditSubmitted event in the same
// transaction as the status update. A CAS conflict from a concurrent
// transition is remapped to ErrConcurrentTransition (409 instead of 412).
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
	if err := version.CanTransition(domain.VersionStatusUnderReview, config.HasReviewer()); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	version.Status = domain.VersionStatusUnderReview
	version.SubmittedAt = &now

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

	// Validation reads above still happen before the write transaction; this tx
	// only serializes the state update itself in the current implementation.
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateSubmit), "tenant"); err != nil {
			return fmt.Errorf("templates submit: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
			// CAS lost to a concurrent transition; reclassify so the HTTP layer
			// returns 409 (conflict) rather than 412 (precondition failed).
			if errors.Is(err, domain.ErrStaleLockVersion) {
				return domain.ErrConcurrentTransition
			}
			return err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates submit: append audit", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return version, nil
}

// ReviewCmd carries the reviewer's decision (Accept or reject with Reason)
// on a template version that is under review.
type ReviewCmd struct {
	TenantID, ActorUserID string
	ActorRoles            []string
	TemplateID            string
	VersionNumber         int
	Accept                bool
	Reason                string
}

// Review records the reviewer stage decision for a version that is
// under-review. The version must have a pending reviewer role, and the actor
// must hold that role and pass segregation-of-duties (reviewer must not be
// the author). On accept, the version moves to approved and an
// AuditReviewed event is appended; on reject, it reverts to draft and an
// AuditRejected event (stage "reviewer") is appended. The status update and
// audit append happen atomically via updateVersionWithAuthzAndAudit.
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
	if version.Status != domain.VersionStatusUnderReview {
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

		audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditReviewed, nil, s.clock.Now())
		if err != nil {
			return nil, err
		}
		if err := s.updateVersionWithAuthzAndAudit(ctx, cmd.TenantID, cmd.ActorUserID, version, string(iamdomain.CapTemplateReview), audit); err != nil {
			return nil, err
		}
		return version, nil
	}

	if err := version.CanTransition(domain.VersionStatusDraft, true); err != nil {
		return nil, err
	}
	version.Status = domain.VersionStatusDraft
	version.SubmittedAt = nil

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
	if err := s.updateVersionWithAuthzAndAudit(ctx, cmd.TenantID, cmd.ActorUserID, version, string(iamdomain.CapTemplateReview), audit); err != nil {
		return nil, err
	}

	return version, nil
}

// ApproveCmd carries the approver's decision (Accept or reject with Reason)
// on a template version that has cleared review (or is under review, when
// no reviewer stage is configured).
type ApproveCmd struct {
	TenantID, ActorUserID string
	ActorRoles            []string
	TemplateID            string
	VersionNumber         int
	Accept                bool
	Reason                string
}

// ApproveResult holds the now-published version on the accept path, or
// the reverted draft on the reject path. Approve no longer spawns the next
// revision (M1·T2); use CreateNextVersion (POST
// /api/v1/templates/{id}/versions) to start a new draft deliberately.
type ApproveResult struct {
	Version *domain.TemplateVersion
}

// Approve records the approver stage decision, publishing the version on
// accept or reverting it to draft on reject. The required source status
// depends on whether a reviewer stage is configured (approved vs.
// under-review); the actor must hold the version's pending approver role
// binding and pass segregation-of-duties against the author and reviewer.
// On accept, the version's content hash must already be committed
// (ErrContentHashMismatch otherwise); the version is published, the
// template's PublishedVersionID is updated (ADR 0065 — revision/number are
// projected onto the read model from the version row, not stored on the
// aggregate), the previously published version (if any) is transitioned to
// obsolete, and AuditPublished (plus AuditObsoleted when applicable) events
// are appended — all atomically in one transaction. On reject, the version
// reverts to draft and an AuditRejected event (stage "approver") is
// appended. A CAS conflict from a concurrent transition is remapped to
// ErrConcurrentTransition (409 instead of 412). Approve no longer spawns the
// next revision (M1·T2); callers use CreateNextVersion for that.
func (s *Service) Approve(ctx context.Context, cmd ApproveCmd) (*ApproveResult, error) {
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
	} else if version.Status != domain.VersionStatusUnderReview {
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
		// T-004: content_hash gate — presigned docx must have been committed.
		if version.ContentHash == "" {
			return nil, domain.ErrContentHashMismatch
		}
		if err := version.CanTransition(domain.VersionStatusPublished, hasReviewer); err != nil {
			return nil, err
		}
		version.Status = domain.VersionStatusPublished
		version.ApproverID = &cmd.ActorUserID
		version.ApprovedAt = &now
		version.PublishedAt = &now

		// Capture the template's currently-published version (if any) BEFORE
		// overwriting the pointer below — this is the version ObsoletePreviousPublishedTx
		// is about to transition to 'obsolete', and it's what the AuditObsoleted
		// event must reference.
		obsoletedVersionID := template.PublishedVersionID

		// ADR 0065: the template aggregate carries only the published-version
		// FK; revision/number projections live on the read model, assembled
		// from the version row on read. Post-command responses read the version
		// directly, so no in-memory projection write is needed here.
		template.PublishedVersionID = &version.ID

		audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditPublished, nil, s.clock.Now())
		if err != nil {
			return nil, err
		}
		var obsoletedAudit *domain.AuditEvent
		if obsoletedVersionID != nil {
			obsoletedAudit, err = newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, obsoletedVersionID, domain.AuditObsoleted, map[string]any{"superseded_by_version_id": version.ID}, s.clock.Now())
			if err != nil {
				return nil, err
			}
		}
		if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
			if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateApprove), "tenant"); err != nil {
				return fmt.Errorf("templates approve: authz: %w", err)
			}
			if err := s.repo.ObsoletePreviousPublishedTx(ctx, tx, cmd.TenantID, cmd.TemplateID, version.ID); err != nil {
				return err
			}
			if obsoletedAudit != nil {
				if err := s.repo.AppendAuditTx(ctx, tx, obsoletedAudit); err != nil {
					return wrapAppErr("templates approve: append obsoleted audit", err)
				}
			}
			if err := s.repo.UpdateTemplateTx(ctx, tx, &template.Template); err != nil {
				return err
			}
			if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
				// CAS lost to a concurrent transition; reclassify so the HTTP layer
				// returns 409 (conflict) rather than 412 (precondition failed).
				if errors.Is(err, domain.ErrStaleLockVersion) {
					return domain.ErrConcurrentTransition
				}
				return err
			}
			if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
				return wrapAppErr("templates approve: append audit", err)
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return &ApproveResult{Version: version}, nil
	}

	if err := version.CanTransition(domain.VersionStatusDraft, hasReviewer); err != nil {
		return nil, err
	}
	version.Status = domain.VersionStatusDraft
	version.SubmittedAt = nil
	version.ReviewedAt = nil
	version.ApprovedAt = nil

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
	if err := s.updateVersionWithAuthzAndAudit(ctx, cmd.TenantID, cmd.ActorUserID, version, string(iamdomain.CapTemplateApprove), audit); err != nil {
		return nil, err
	}

	return &ApproveResult{Version: version}, nil
}

// ArchiveCmd identifies the template to archive.
type ArchiveCmd struct {
	TenantID, ActorUserID, TemplateID string
}

// PublishTemplateVersionCmd identifies the draft template version to publish
// directly (bypassing the review/approve workflow), along with the derived
// schema object key to record on the resulting AuditPublished event.
type PublishTemplateVersionCmd struct {
	TenantID, ActorUserID, TemplateID string
	ActorRoles                        []string
	VersionNumber                     int
	SchemaKey                         string
}

// PublishTemplateVersionResult holds the now-published version. Publish no
// longer spawns the next revision (M1·T2); use CreateNextVersion (POST
// /api/v1/templates/{id}/versions) to start a new draft deliberately.
type PublishTemplateVersionResult struct {
	PublishedVersion *domain.TemplateVersion
}

// PublishTemplateVersion publishes a draft version directly, without going
// through the review/approve workflow. The version must be a draft with a
// committed content hash (ErrContentHashMismatch otherwise) and the actor
// must pass segregation-of-duties (publisher must not be the author or
// reviewer) and hold the version's pending approver role binding — a role
// mismatch is audited as AuditPublishForbiddenRole (best-effort, non-fatal)
// before returning ErrForbiddenRole. On success, the version is published,
// the template's PublishedVersionID is updated (ADR 0065 — revision/number
// projected onto the read model, not stored on the aggregate), the previously
// published version (if any) is transitioned to obsolete, and AuditPublished
// (plus AuditObsoleted
// when applicable) events are appended — all atomically in one transaction.
// A CAS conflict from a concurrent transition is remapped to
// ErrConcurrentTransition (409 instead of 412). Publish no longer spawns the
// next revision (M1·T2); callers use CreateNextVersion for that.
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
			slog.Warn("templates publish: append denied audit failed", "err", appendErr)
		}
		return nil, domain.ErrForbiddenRole
	}
	now := s.clock.Now()
	version.Status = domain.VersionStatusPublished
	version.PublishedAt = &now
	version.ApprovedAt = &now

	// Capture the template's currently-published version (if any) BEFORE
	// overwriting the pointer below — this is the version ObsoletePreviousPublishedTx
	// is about to transition to 'obsolete', and it's what the AuditObsoleted
	// event must reference.
	obsoletedVersionID := template.PublishedVersionID

	// ADR 0065: aggregate carries only the published-version FK; the read
	// model projects revision/number/status from the version row on read.
	template.PublishedVersionID = &version.ID

	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditPublished, map[string]any{"schema_key": cmd.SchemaKey}, now)
	if err != nil {
		return nil, err
	}
	var obsoletedAudit *domain.AuditEvent
	if obsoletedVersionID != nil {
		obsoletedAudit, err = newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, obsoletedVersionID, domain.AuditObsoleted, map[string]any{"superseded_by_version_id": version.ID}, now)
		if err != nil {
			return nil, err
		}
	}
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplatePublish), "tenant"); err != nil {
			return fmt.Errorf("templates publish: authz: %w", err)
		}
		if err := s.repo.ObsoletePreviousPublishedTx(ctx, tx, cmd.TenantID, cmd.TemplateID, version.ID); err != nil {
			return err
		}
		if obsoletedAudit != nil {
			if err := s.repo.AppendAuditTx(ctx, tx, obsoletedAudit); err != nil {
				return wrapAppErr("templates publish: append obsoleted audit", err)
			}
		}
		if err := s.repo.UpdateTemplateTx(ctx, tx, &template.Template); err != nil {
			return err
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, cmd.TenantID, version); err != nil {
			// CAS lost to a concurrent transition; reclassify so the HTTP layer
			// returns 409 (conflict) rather than 412 (precondition failed).
			if errors.Is(err, domain.ErrStaleLockVersion) {
				return domain.ErrConcurrentTransition
			}
			return err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates publish: append audit", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &PublishTemplateVersionResult{PublishedVersion: version}, nil
}

// nextVersionNumber allocates the next version slot. The new draft must be
// numbered above both the template's latest version and the source version it
// spawns from. Unifies what Approve/Publish/CreateNextVersion previously did
// inconsistently.
func nextVersionNumber(latestVersion, sourceVersionNumber int) int {
	n := latestVersion + 1
	if sourceVersionNumber >= n {
		n = sourceVersionNumber + 1
	}
	return n
}

// spawnNextDraft builds the next draft version as a byte-copy of its source at
// the draft's OWN canonical key (never the source's key — that shared-key bug
// overwrites the immutable source object). The copy runs PRE-TX (store-then-
// reference: the object exists before the referencing row commits; the only
// crash outcome is a safe orphan). ContentHash is left empty (the draft
// constructor's default) so the publish gate still forces a real edit before the
// new revision can publish.
func (s *Service) spawnNextDraft(ctx context.Context, tenantID, templateID, actorID string, nextNum int, source *domain.TemplateVersion) (*domain.TemplateVersion, error) {
	dstKey := templateDocxKey(tenantID, templateID, nextNum)
	if err := s.presign.Copy(ctx, tenantID, source.DocxStorageKey, dstKey); err != nil {
		return nil, fmt.Errorf("templates: copy docx to %s: %w", dstKey, err)
	}
	return domain.NewTemplateVersionDraft(
		s.uuid.New(),
		tenantID,
		templateID,
		actorID,
		dstKey,
		nextNum,
		cloneMetadataSchema(source.MetadataSchema),
		clonePlaceholders(source.PlaceholderSchema),
		s.clock.Now(),
	), nil
}

// ArchiveTemplate soft-archives a template by stamping ArchivedAt and
// appending an AuditArchived event, atomically in one transaction. The
// template must not be system-owned. Archiving an already-archived template
// is a no-op that returns the template unchanged.
func (s *Service) ArchiveTemplate(ctx context.Context, cmd ArchiveCmd) (*domain.TemplateRead, error) {
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

	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, nil, domain.AuditArchived, nil, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateArchive), "tenant"); err != nil {
			return fmt.Errorf("templates archive: authz: %w", err)
		}
		if err := s.repo.UpdateTemplateTx(ctx, tx, &template.Template); err != nil {
			return err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates archive: append audit", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return template, nil
}

// updateVersionWithAuthzAndAudit runs the authz check and the version update
// inside a single transaction, then appends the audit event atomically (F-07 / REQ-ASYNC-1).
// ErrStaleLockVersion from UpdateVersionTx is remapped to ErrConcurrentTransition
// so that status-transition callers surface a 409 rather than a 412.
func (s *Service) updateVersionWithAuthzAndAudit(ctx context.Context, tenantID, actorID string, version *domain.TemplateVersion, cap string, audit *domain.AuditEvent) error {
	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.Require(ctx, tx, cap, "tenant"); err != nil {
			return fmt.Errorf("templates update version: authz: %w", err)
		}
		if err := s.repo.UpdateVersionTx(ctx, tx, tenantID, version); err != nil {
			// CAS lost to a concurrent transition; reclassify so the HTTP layer
			// returns 409 (conflict) rather than 412 (precondition failed).
			if errors.Is(err, domain.ErrStaleLockVersion) {
				return domain.ErrConcurrentTransition
			}
			return err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates update version: append audit", err)
		}
		return nil
	})
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
