package application

import (
	"context"

	"metaldocs/internal/modules/templates/domain"
)

type UpsertApprovalConfigCmd struct {
	TenantID, ActorUserID, TemplateID string
	ActorRoles                        []string
	ReviewerRole                      *string
	ApproverRole                      string
}

func (s *Service) UpsertApprovalConfig(ctx context.Context, cmd UpsertApprovalConfigCmd) (*domain.ApprovalConfig, error) {
	template, err := s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)
	if err != nil {
		return nil, wrapAppErr("templates approval config: get template", err)
	}
	if template.TenantID != cmd.TenantID {
		return nil, domain.ErrNotFound
	}
	if template.SystemOwned {
		return nil, domain.ErrSystemTemplateImmutable
	}
	if template.IsArchived() {
		return nil, domain.ErrArchived
	}

	hasEverPublished := template.PublishedVersionID != nil
	if hasEverPublished {
		if !containsRole(cmd.ActorRoles, "admin") {
			return nil, domain.ErrForbidden
		}
	} else {
		if template.CreatedBy != cmd.ActorUserID && !containsRole(cmd.ActorRoles, "admin") {
			return nil, domain.ErrForbidden
		}
	}

	if cmd.ApproverRole == "" {
		return nil, domain.ErrInvalidApprovalConfig
	}

	config, err := domain.NewApprovalConfig(cmd.TemplateID, cmd.ApproverRole, cmd.ReviewerRole)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpsertApprovalConfig(ctx, &config); err != nil {
		return nil, wrapAppErr("templates approval config: upsert", err)
	}

	audit, err := newAuditEvent(
		cmd.TenantID,
		cmd.TemplateID,
		cmd.ActorUserID,
		nil,
		domain.AuditApprovalConfigUpdated,
		map[string]any{
			"reviewer_role": cmd.ReviewerRole,
			"approver_role": cmd.ApproverRole,
		},
		s.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AppendAudit(ctx, audit); err != nil {
		return nil, wrapAppErr("templates approval config: append audit", err)
	}

	return &config, nil
}
