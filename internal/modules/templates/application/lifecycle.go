package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
)

// ArchiveCmd identifies the template to archive.
type ArchiveCmd struct {
	TenantID, ActorUserID, TemplateID string
}

// PublishTemplateVersionCmd identifies the template version to publish —
// either directly from draft (bypassing the review/approve workflow) or
// completing a version the kernel signoff flow already brought to approved
// (ADR 0082).
type PublishTemplateVersionCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
}

// PublishTemplateVersionResult holds the now-published version. Publish no
// longer spawns the next revision (M1·T2); use CreateNextVersion (POST
// /api/v1/templates/{id}/versions) to start a new draft deliberately.
type PublishTemplateVersionResult struct {
	PublishedVersion *domain.TemplateVersion
}

// PublishTemplateVersion publishes a version that is either a draft
// (bypassing the review/approve workflow entirely) or already approved
// (completing the kernel signoff flow, which ends at approved — ADR 0082).
// ADR 0088 removed the "must carry a committed content hash" precondition:
// every version is materialized at creation, so a never-edited draft publishes
// like any other. The actor must pass segregation-of-duties (publisher must
// not be the author or reviewer) — identity-based, not role-based (ADR 0022:
// authz is capabilities, never roles). Any other source status is rejected
// with ErrInvalidStateTransition. On success, the version is published, the
// template's PublishedVersionID is updated (ADR 0065 — revision/number
// projected onto the read model, not stored on the aggregate), the previously
// published version (if any) is transitioned to obsolete, and AuditPublished
// (schema_key detail always server-derived, never client input; plus
// AuditObsoleted when applicable) events are appended — all atomically in one
// transaction. A CAS conflict from a concurrent transition is remapped
// to ErrConcurrentTransition (409 instead of 412). Publish no longer spawns
// the next revision (M1·T2); callers use CreateNextVersion for that.
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
	// Publish accepts two source statuses: draft (direct publish, bypassing
	// review/approve entirely) and approved (completing the kernel signoff
	// flow, which ends at approved — ADR 0082; legacy Approve, which used to
	// own approved->published, is retired in a later slice). Any other status
	// is rejected unchanged.
	if version.Status != domain.VersionStatusDraft && version.Status != domain.VersionStatusApproved {
		return nil, domain.ErrInvalidStateTransition
	}

	// ADR 0088 §4: the emptiness gate that used to sit here ("content_hash == ''
	// => ErrContentHashMismatch") is DELETED, not reworded. Under ADR 0088 §1
	// every version row carries the verified hash of its object from the moment
	// it exists (DB CHECK chk_template_version_content_hash_non_draft is now
	// unconditional, migration 0317), so the branch was unreachable. Its real
	// job had drifted to "force a real edit before publishing", which is a
	// judgment for the reviewer, not a shape fabricated out of a null column.
	// A blank or unchanged version publishing is intended behaviour.

	// T-004: SoD — publisher must not be the author or the reviewer. This is
	// identity-based (actor user ID vs. author/reviewer user IDs), not
	// role-based, so it stands on its own under ADR 0022 (authz is
	// capabilities, never roles) — unlike the role-binding tier-2 check this
	// used to sit beside, which is deleted (ADR 0082 §Transitional
	// coexistence, unit 3.1a slice S2). Capability (Tier 1,
	// CapTemplatePublish) is enforced inside the tx below.
	if err := domain.CheckSegregation(domain.SegregationRoleApprover, cmd.ActorUserID, version.AuthorID, version.ReviewerID); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	version.Status = domain.VersionStatusPublished
	version.PublishedAt = &now
	// Direct publish (draft source) stamps ApprovedAt at publish time —
	// preserved legacy behavior. Kernel completion (approved source) must NOT
	// clobber the true approval time already stamped by
	// ApprovalCompletionWriter.MarkTemplateVersionApproved at signoff
	// completion (reviewer finding, unit 3.1a S2).
	if version.ApprovedAt == nil {
		version.ApprovedAt = &now
	}

	// Capture the template's currently-published version (if any) BEFORE
	// overwriting the pointer below — this is the version ObsoletePreviousPublishedTx
	// is about to transition to 'obsolete', and it's what the AuditObsoleted
	// event must reference.
	obsoletedVersionID := template.PublishedVersionID

	// ADR 0065: aggregate carries only the published-version FK; the read
	// model projects revision/number/status from the version row on read.
	template.PublishedVersionID = &version.ID

	schemaKey := templateSchemaKey(cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditPublished, map[string]any{"schema_key": schemaKey}, now)
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
// crash outcome is a safe orphan).
//
// ADR 0088: the copy is followed by a Confirm against the source's hash, and
// the draft is constructed WITH the verified hash it returns — the same single
// meaning content_hash carries everywhere else (the hash of the object THIS
// version points at). The superseded stance ("leave ContentHash empty so the
// publish gate still forces a real edit") is gone with the gate it served:
// whether an unchanged revision deserves approval is the reviewer's judgment,
// not a shape expressed by a null column.
func (s *Service) spawnNextDraft(ctx context.Context, tenantID, templateID, actorID string, nextNum int, source *domain.TemplateVersion) (*domain.TemplateVersion, error) {
	if len(source.ContentHash) != 64 {
		// Fail closed: under ADR 0088 every version carries a real hash, so a
		// source without one is a corrupt row (or a pre-0317 leftover), never a
		// licence to spawn another content-less draft from it. That is a server
		// -side data defect, not a missing user upload — it maps to 500, not to
		// a 4xx telling the user to upload something they cannot upload.
		return nil, fmt.Errorf(
			"templates: source version %s has no verified content hash (length=%d): %w",
			source.ID, len(source.ContentHash), domain.ErrContentMaterializationFailed,
		)
	}
	dstKey := templateDocxKey(tenantID, templateID, nextNum)
	contentHash, err := s.copyAndConfirm(ctx, tenantID, source.DocxStorageKey, dstKey, source.ContentHash)
	if err != nil {
		return nil, err
	}
	return domain.NewTemplateVersionDraft(
		s.uuid.New(),
		tenantID,
		templateID,
		actorID,
		dstKey,
		contentHash,
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
