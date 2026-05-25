package application

import (
	"context"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
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

	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates approval config: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return nil, fmt.Errorf("templates approval config: setAuthzGUC: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return nil, fmt.Errorf("templates approval config: authz: %w", err)
		}
		if err := s.repo.UpsertApprovalConfigTx(ctx, tx, &config); err != nil {
			return nil, wrapAppErr("templates approval config: upsert", err)
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return nil, wrapAppErr("templates approval config: append audit", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.UpsertApprovalConfig(ctx, &config); err != nil {
			return nil, wrapAppErr("templates approval config: upsert", err)
		}
		if err := s.repo.AppendAudit(ctx, audit); err != nil {
			return nil, wrapAppErr("templates approval config: append audit", err)
		}
	}

	return &config, nil
}
