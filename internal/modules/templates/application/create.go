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

// CreateTemplateCmd carries the fields needed to create a new template plus
// the initial approval role bindings for its first draft version.
type CreateTemplateCmd struct {
	TenantID     string
	ActorUserID  string
	DocTypeCode  string
	Key          string
	Name         string
	Description  string
	ApproverRole string
	ReviewerRole *string
}

// CreateTemplateResult holds the newly created template and its initial
// draft version.
type CreateTemplateResult struct {
	Template *domain.Template
	Version  *domain.TemplateVersion
}

// CreateTemplate creates a new template together with its first draft
// version (version 1) and approval configuration, and records an
// AuditCreated event, all inside one transaction. The template key must be
// unique within the tenant; an existing key is rejected with
// ErrKeyConflict.
func (s *Service) CreateTemplate(ctx context.Context, cmd CreateTemplateCmd) (*CreateTemplateResult, error) {
	if _, err := s.repo.GetTemplateByKey(ctx, cmd.TenantID, cmd.Key); err == nil {
		return nil, domain.ErrKeyConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	template := &domain.Template{
		ID:                 s.uuid.New(),
		TenantID:           cmd.TenantID,
		DocTypeCode:        cmd.DocTypeCode,
		Key:                cmd.Key,
		Name:               cmd.Name,
		Description:        cmd.Description,
		LatestVersion:      1,
		PublishedVersionID: nil,
		CreatedBy:          cmd.ActorUserID,
		CreatedAt:          s.clock.Now(),
	}

	version := domain.NewTemplateVersionDraft(
		s.uuid.New(),
		cmd.TenantID,
		template.ID,
		cmd.ActorUserID,
		templateDocxKey(cmd.TenantID, template.ID, 1),
		1,
		domain.MetadataSchema{},
		[]domain.Placeholder{},
		s.clock.Now(),
	)
	version.PendingApproverRole = cmd.ApproverRole
	version.PendingReviewerRole = cmd.ReviewerRole

	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return fmt.Errorf("templates create: set authz context: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateCreate), "tenant"); err != nil {
			return fmt.Errorf("templates create: authz: %w", err)
		}
		if err := s.repo.CreateTemplateTx(ctx, tx, template); err != nil {
			return err
		}
		if err := s.repo.CreateVersionTx(ctx, tx, version); err != nil {
			return err
		}
		if err := s.repo.UpsertApprovalConfigTx(ctx, tx, &domain.ApprovalConfig{
			TemplateID:   template.ID,
			ApproverRole: cmd.ApproverRole,
			ReviewerRole: cmd.ReviewerRole,
		}); err != nil {
			return err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, &domain.AuditEvent{
			TenantID:   cmd.TenantID,
			TemplateID: template.ID,
			VersionID:  &version.ID,
			ActorID:    cmd.ActorUserID,
			Action:     domain.AuditCreated,
			Details:    map[string]any{},
			OccurredAt: s.clock.Now(),
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &CreateTemplateResult{
		Template: template,
		Version:  version,
	}, nil
}

// CreateVersionCmd identifies the template to spawn a new draft version for.
type CreateVersionCmd struct {
	TenantID    string
	ActorUserID string
	TemplateID  string
}

// CreateNextVersion spawns a new draft version as a byte-copy of the
// template's currently published version (or its latest version, if none has
// ever published), numbered above both the template's latest version and the
// source version. The template must not be system-owned or archived. The new
// version, the updated template.LatestVersion pointer, and an AuditCreated
// event are persisted atomically in one transaction.
func (s *Service) CreateNextVersion(ctx context.Context, cmd CreateVersionCmd) (*domain.TemplateVersion, error) {
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	if template.SystemOwned {
		return nil, domain.ErrSystemTemplateImmutable
	}
	if template.IsArchived() {
		return nil, domain.ErrArchived
	}

	var source *domain.TemplateVersion
	if template.PublishedVersionID != nil {
		source, err = s.repo.GetVersionByID(ctx, cmd.TenantID, *template.PublishedVersionID)
		if err != nil {
			return nil, err
		}
		// T-002: GetVersionByID has no tenant predicate — reject if version belongs to a different template.
		if source.TemplateID != template.ID {
			return nil, domain.ErrNotFound
		}
	} else {
		source, err = s.repo.GetVersion(ctx, cmd.TenantID, template.ID, template.LatestVersion)
		if err != nil {
			return nil, err
		}
	}

	newNum := nextVersionNumber(template.LatestVersion, source.VersionNumber)
	version, err := s.spawnNextDraft(ctx, cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, newNum, source)
	if err != nil {
		return nil, err
	}

	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return fmt.Errorf("templates create next version: setAuthzGUC: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return fmt.Errorf("templates create next version: authz: %w", err)
		}
		if err := s.repo.CreateVersionTx(ctx, tx, version); err != nil {
			return err
		}
		template.LatestVersion = newNum
		if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
			return err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, &domain.AuditEvent{
			TenantID:   cmd.TenantID,
			TemplateID: template.ID,
			VersionID:  &version.ID,
			ActorID:    cmd.ActorUserID,
			Action:     domain.AuditCreated,
			Details:    map[string]any{},
			OccurredAt: s.clock.Now(),
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return version, nil
}

func cloneMetadataSchema(s domain.MetadataSchema) domain.MetadataSchema {
	return domain.MetadataSchema{
		DocCodePattern:      s.DocCodePattern,
		RetentionDays:       s.RetentionDays,
		DistributionDefault: cloneStringSlice(s.DistributionDefault),
		RequiredMetadata:    cloneStringSlice(s.RequiredMetadata),
	}
}

func clonePlaceholders(in []domain.Placeholder) []domain.Placeholder {
	out := make([]domain.Placeholder, len(in))
	for i := range in {
		out[i] = in[i]
		out[i].Options = cloneStringSlice(in[i].Options)
	}
	return out
}

func cloneStringSlice(in []string) []string {
	if in == nil {
		return nil
	}
	return append([]string{}, in...)
}
