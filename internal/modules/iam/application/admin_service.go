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

// RoleCacheInvalidator is the narrow port for evicting a user's cached role
// set. Implemented by CachedRoleProvider. Callers MUST invoke
// InvalidateUserTenant post-commit after any write that changes a user's
// roles or tenant membership; otherwise the cache serves stale authz
// decisions until TTL expiry.
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

// AdminService owns the legacy single-tenant-role assignment write paths
// (upsert-or-assign and replace). When runner and audit are both wired, the
// mutation and its audit row commit atomically in one tx (H-3b REQ-ASYNC-1);
// otherwise it falls back to a non-tx repo call for bootstrap/test paths.
type AdminService struct {
	repo        domain.RoleAdminRepository
	invalidator RoleCacheInvalidator
	runner      db.TxRunner
	audit       auditdomain.Writer
}

// NewAdminService constructs the service. invalidator, runner, and audit may
// be nil; nil runner/audit forces the non-tx repo fallback path on every
// write, and a nil invalidator skips post-commit cache eviction.
func NewAdminService(repo domain.RoleAdminRepository, invalidator RoleCacheInvalidator, runner db.TxRunner, audit auditdomain.Writer) *AdminService {
	return &AdminService{repo: repo, invalidator: invalidator, runner: runner, audit: audit}
}

// UpsertUserAndAssignRole creates the user (if absent) and assigns role,
// atomically with an audit row when runner/audit are wired. Returns
// domain.ErrUserNotFound for an empty userID/tenantID and domain.ErrInvalidRole
// for a role outside the canonical set. Invalidates the actor's cached roles
// post-commit (H-3b invariant) — never before, so a rolled-back tx cannot
// evict a still-valid cache entry.
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

// ReplaceUserRoles replaces the user's single tenant role with role,
// atomically with an audit row when runner/audit are wired. Returns
// domain.ErrUserNotFound for an empty userID/tenantID and domain.ErrInvalidRole
// for a role outside the canonical set. Invalidates the actor's cached roles
// post-commit (H-3b invariant).
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
