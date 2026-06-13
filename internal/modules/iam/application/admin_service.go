package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/requesttrace"
)

type RoleCacheInvalidator interface {
	InvalidateUserTenant(userID, tenantID string)
}

// roleAdminTxRepository is the tx-aware subset of the role admin repository.
// Defined in the application layer (NOT the domain port) so the domain
// interface never names database/sql's *sql.Tx (REQ-FE-2 / nosqltxindomain).
// Satisfied structurally by *iampostgres.RoleAdminRepository and the in-memory
// test fixture (H-3b).
type roleAdminTxRepository interface {
	UpsertUserAndAssignRoleTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, role domain.Role, assignedBy string) error
	ReplaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, role domain.Role, assignedBy string) error
}

type AdminService struct {
	repo        domain.RoleAdminRepository
	invalidator RoleCacheInvalidator
	runner      db.TxRunner
	audit       auditdomain.Writer
}

func NewAdminService(repo domain.RoleAdminRepository, invalidator RoleCacheInvalidator, runner db.TxRunner, audit auditdomain.Writer) *AdminService {
	return &AdminService{repo: repo, invalidator: invalidator, runner: runner, audit: audit}
}

func (s *AdminService) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy, actorID string) error {
	userID = strings.TrimSpace(userID)
	displayName = strings.TrimSpace(displayName)
	tenantID = strings.TrimSpace(tenantID)
	assignedBy = strings.TrimSpace(assignedBy)

	if userID == "" {
		return domain.ErrUserNotFound
	}
	if tenantID == "" {
		return domain.ErrUserNotFound
	}
	if !domain.IsValidRole(role) {
		return domain.ErrInvalidRole
	}
	if displayName == "" {
		displayName = userID
	}
	if assignedBy == "" {
		assignedBy = "system"
	}

	// When runner and audit are both wired, emit the audit row in the same tx
	// as the mutation (H-3b REQ-ASYNC-1). Otherwise fall back to the non-tx
	// repo path (bootstrap / tests without a real DB).
	txRepo, txOK := s.repo.(roleAdminTxRepository)
	if s.runner != nil && s.audit != nil && txOK {
		ev, err := s.buildAuditEvent(ctx, tenantID, userID, actorID, "iam.user.role.upserted", map[string]any{
			"role":        string(role),
			"assigned_by": assignedBy,
		})
		if err != nil {
			return err
		}
		if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
			if err := txRepo.UpsertUserAndAssignRoleTx(ctx, tx, userID, displayName, tenantID, role, assignedBy); err != nil {
				return err
			}
			return s.audit.RecordTx(ctx, tx, ev)
		}); err != nil {
			return err
		}
	} else {
		if err := s.repo.UpsertUserAndAssignRole(ctx, userID, displayName, tenantID, role, assignedBy); err != nil {
			return err
		}
	}

	// Cache invalidation MUST happen post-commit so a rolled-back tx never
	// evicts a cache entry prematurely (H-3b invariant).
	if s.invalidator != nil {
		s.invalidator.InvalidateUserTenant(userID, tenantID)
	}
	return nil
}

func (s *AdminService) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy, actorID string) error {
	userID = strings.TrimSpace(userID)
	displayName = strings.TrimSpace(displayName)
	tenantID = strings.TrimSpace(tenantID)
	assignedBy = strings.TrimSpace(assignedBy)

	if userID == "" {
		return domain.ErrUserNotFound
	}
	if tenantID == "" {
		return domain.ErrUserNotFound
	}
	if displayName == "" {
		displayName = userID
	}
	if assignedBy == "" {
		assignedBy = "system"
	}
	if !domain.IsValidRole(role) {
		return domain.ErrInvalidRole
	}

	// When runner and audit are both wired, emit the audit row in the same tx
	// as the mutation (H-3b REQ-ASYNC-1). Otherwise fall back to the non-tx
	// repo path (bootstrap / tests without a real DB).
	txRepo, txOK := s.repo.(roleAdminTxRepository)
	if s.runner != nil && s.audit != nil && txOK {
		ev, err := s.buildAuditEvent(ctx, tenantID, userID, actorID, "iam.user.roles.replaced", map[string]any{
			"roles": []string{string(role)},
		})
		if err != nil {
			return err
		}
		if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
			if err := txRepo.ReplaceUserRolesTx(ctx, tx, userID, displayName, tenantID, role, assignedBy); err != nil {
				return err
			}
			return s.audit.RecordTx(ctx, tx, ev)
		}); err != nil {
			return err
		}
	} else {
		if err := s.repo.ReplaceUserRoles(ctx, userID, displayName, tenantID, role, assignedBy); err != nil {
			return err
		}
	}

	// Cache invalidation MUST happen post-commit so a rolled-back tx never
	// evicts a cache entry prematurely (H-3b invariant).
	if s.invalidator != nil {
		s.invalidator.InvalidateUserTenant(userID, tenantID)
	}
	return nil
}

// buildAuditEvent stamps the audit Event. actorID is the authenticated principal
// who performed the action (the audit ActorID) and is deliberately distinct from
// the domain-level assignedBy column, which may be client-supplied (H-3b M2: the
// audit actor must never be forgeable via the request body).
func (s *AdminService) buildAuditEvent(ctx context.Context, tenantID, userID, actorID, action string, payload map[string]any) (auditdomain.Event, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return auditdomain.Event{}, err
	}
	return auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       action,
		ResourceType: "user",
		ResourceID:   userID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      requesttrace.Resolve(ctx),
		TenantID:     tenantID,
	}, nil
}
