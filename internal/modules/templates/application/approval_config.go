package application

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
)

type UpsertApprovalConfigCmd struct {
	TenantID, ActorUserID, TemplateID string
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

	// Elevation rule (ADR 0051 / REQ-AUTHZ-5):
	//   Published template  → require template.manage (in-tx); no creator shortcut.
	//   Unpublished template, creator → domain-ownership shortcut; template.edit suffices.
	//   Unpublished template, non-creator → require template.manage (in-tx).
	// The in-tx CapTemplateEdit check (:77) remains the base gate for all paths.
	hasEverPublished := template.PublishedVersionID != nil
	requireManage := hasEverPublished || template.CreatedBy != cmd.ActorUserID

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

	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return fmt.Errorf("templates approval config: setAuthzGUC: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return fmt.Errorf("templates approval config: authz: %w", err)
		}
		if requireManage {
			if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateManage), "tenant"); err != nil {
				return fmt.Errorf("templates approval config: authz manage: %w", err)
			}
		}
		if err := s.repo.UpsertApprovalConfigTx(ctx, tx, &config); err != nil {
			return wrapAppErr("templates approval config: upsert", err)
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates approval config: append audit", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &config, nil
}
