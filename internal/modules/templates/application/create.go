package application

import (
	"context"
	"errors"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
)

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

type CreateTemplateResult struct {
	Template *domain.Template
	Version  *domain.TemplateVersion
}

func (s *Service) CreateTemplate(ctx context.Context, cmd CreateTemplateCmd) (*CreateTemplateResult, error) {
	if _, err := s.repo.GetTemplateByKey(ctx, cmd.TenantID, cmd.Key); err == nil {
		return nil, domain.ErrKeyConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("templates create: check template key uniqueness: %w", err)
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
		template.ID,
		cmd.ActorUserID,
		fmt.Sprintf("templates/%s/versions/1.docx", template.ID),
		1,
		mustMetadataSchema(),
		[]domain.Placeholder{},
		s.clock.Now(),
	)
	version.PendingApproverRole = cmd.ApproverRole
	version.PendingReviewerRole = cmd.ReviewerRole

	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates create: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return nil, fmt.Errorf("templates create: set authz context: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateCreate), "tenant"); err != nil {
			return nil, fmt.Errorf("templates create: authz: %w", err)
		}
		if err := s.repo.CreateTemplateTx(ctx, tx, template); err != nil {
			return nil, err
		}
		if err := s.repo.CreateVersionTx(ctx, tx, version); err != nil {
			return nil, err
		}
		config, err := domain.NewApprovalConfig(template.ID, cmd.ApproverRole, cmd.ReviewerRole)
		if err != nil {
			return nil, err
		}
		if err := s.repo.UpsertApprovalConfigTx(ctx, tx, &config); err != nil {
			return nil, err
		}
		audit, err := newAuditEvent(cmd.TenantID, template.ID, cmd.ActorUserID, &version.ID, domain.AuditCreated, map[string]any{}, s.clock.Now())
		if err != nil {
			return nil, err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.CreateTemplate(ctx, template); err != nil {
			return nil, err
		}
		if err := s.repo.CreateVersion(ctx, version); err != nil {
			return nil, err
		}
		config, err := domain.NewApprovalConfig(template.ID, cmd.ApproverRole, cmd.ReviewerRole)
		if err != nil {
			return nil, err
		}
		if err := s.repo.UpsertApprovalConfig(ctx, &config); err != nil {
			return nil, err
		}
		audit, err := newAuditEvent(cmd.TenantID, template.ID, cmd.ActorUserID, &version.ID, domain.AuditCreated, map[string]any{}, s.clock.Now())
		if err != nil {
			return nil, err
		}
		if err := s.repo.AppendAudit(ctx, audit); err != nil {
			return nil, err
		}
	}

	return &CreateTemplateResult{
		Template: template,
		Version:  version,
	}, nil
}

type CreateVersionCmd struct {
	TenantID    string
	ActorUserID string
	TemplateID  string
}

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
		if source.TemplateID != template.ID {
			return nil, domain.ErrNotFound
		}
	} else {
		source, err = s.repo.GetVersion(ctx, cmd.TenantID, template.ID, template.LatestVersion)
		if err != nil {
			return nil, err
		}
	}

	newNum := template.LatestVersion + 1
	version := domain.NewTemplateVersionDraft(
		s.uuid.New(),
		cmd.TemplateID,
		cmd.ActorUserID,
		fmt.Sprintf("templates/%s/versions/%d.docx", cmd.TemplateID, newNum),
		newNum,
		cloneMetadataSchema(source.MetadataSchema),
		clonePlaceholders(source.PlaceholderSchema),
		s.clock.Now(),
	)

	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates create next version: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := s.repo.CreateVersionTx(ctx, tx, version); err != nil {
			return nil, err
		}
		template.LatestVersion = newNum
		if err := s.repo.UpdateTemplateTx(ctx, tx, template); err != nil {
			return nil, err
		}
		audit, err := newAuditEvent(cmd.TenantID, template.ID, cmd.ActorUserID, &version.ID, domain.AuditCreated, map[string]any{}, s.clock.Now())
		if err != nil {
			return nil, err
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.CreateVersion(ctx, version); err != nil {
			return nil, err
		}
		template.LatestVersion = newNum
		if err := s.repo.UpdateTemplate(ctx, template); err != nil {
			return nil, err
		}
		audit, err := newAuditEvent(cmd.TenantID, template.ID, cmd.ActorUserID, &version.ID, domain.AuditCreated, map[string]any{}, s.clock.Now())
		if err != nil {
			return nil, err
		}
		if err := s.repo.AppendAudit(ctx, audit); err != nil {
			return nil, err
		}
	}

	return version, nil
}

func mustMetadataSchema() domain.MetadataSchema {
	schema, err := domain.NewMetadataSchema("", 0, nil, nil)
	if err != nil {
		panic("templates create: default metadata schema must be valid")
	}
	return schema
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
