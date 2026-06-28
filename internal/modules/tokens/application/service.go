package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/tokens/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/requesttrace"
)

const auditResourceType = "token_dictionary_entry"

// authzRequireFunc is the in-tx capability check seam. Production wires
// authz.Require; tests stub it. Signature matches authz.Require.
type authzRequireFunc func(ctx context.Context, tx *sql.Tx, capability, areaCode string) error

// seedTxFunc seeds the tx-local authz GUCs. A seam (matches authz.SeedTxIdentity)
// so unit tests can run the service flow with a nil tx without panicking.
type seedTxFunc func(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error

// AuditRecorderForTest is the exported alias of the auditRecorder port so
// out-of-package tests can satisfy it.
type AuditRecorderForTest = auditRecorder

// Service is the tokens application service. It owns the transaction boundary
// and wires authz + audit into every state-changing operation.
type Service struct {
	runner  txRunner
	repo    domain.Repository
	audit   auditRecorder
	require authzRequireFunc
	seed    seedTxFunc
}

// NewService is the production constructor. It pins the real authz primitives.
func NewService(runner txRunner, repo domain.Repository, audit auditRecorder) *Service {
	if runner == nil || repo == nil || audit == nil {
		panic("tokens.application: runner, repo, audit are required")
	}
	return &Service{runner: runner, repo: repo, audit: audit, require: authz.Require, seed: authz.SeedTxIdentity}
}

// NewServiceForTest injects stub authz primitives (the real ones need a DB).
func NewServiceForTest(runner txRunner, repo domain.Repository, audit auditRecorder, require authzRequireFunc, seed seedTxFunc) *Service {
	return &Service{runner: runner, repo: repo, audit: audit, require: require, seed: seed}
}

// Compile-time proof the service satisfies the published DictionaryReader (SP-2 surface).
var _ domain.DictionaryReader = (*Service)(nil)

// CreateCommand carries the inputs for creating a new token dictionary entry.
type CreateCommand struct {
	TenantID, ActorID, Name, Value, Label string
	Description                           *string
}

// UpdateCommand carries the inputs for updating an existing token dictionary entry.
type UpdateCommand struct {
	TenantID, ActorID, ID, Name, Value, Label string
	Description                               *string
}

// Create creates a new entry, enforcing authz and recording an audit event.
func (s *Service) Create(ctx context.Context, cmd CreateCommand) (*domain.Entry, error) {
	entry, err := domain.NewEntry(domain.NewEntryInput{
		TenantID:    cmd.TenantID,
		ActorID:     cmd.ActorID,
		Name:        cmd.Name,
		Value:       cmd.Value,
		Label:       cmd.Label,
		Description: cmd.Description,
	})
	if err != nil {
		return nil, err
	}
	var out *domain.Entry
	err = s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := s.seed(ctx, tx, cmd.TenantID, cmd.ActorID); err != nil {
			return err
		}
		if err := s.require(ctx, tx, string(iamdomain.CapTokenDictionaryManage), "tenant"); err != nil {
			return err
		}
		created, err := s.repo.Create(ctx, tx, entry)
		if err != nil {
			return err
		}
		out = created
		return s.record(ctx, tx, cmd.TenantID, cmd.ActorID, "tokens.entry.created", created)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Update updates an existing entry, enforcing name immutability, authz, and audit.
func (s *Service) Update(ctx context.Context, cmd UpdateCommand) (*domain.Entry, error) {
	var out *domain.Entry
	err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := s.seed(ctx, tx, cmd.TenantID, cmd.ActorID); err != nil {
			return err
		}
		if err := s.require(ctx, tx, string(iamdomain.CapTokenDictionaryManage), "tenant"); err != nil {
			return err
		}
		existing, err := s.repo.GetByID(ctx, tx, cmd.TenantID, cmd.ID)
		if err != nil {
			return err
		}
		// Name is immutable once set. Reject any attempt to change it even if caller
		// sends the field blank (blank != original name).
		if cmd.Name != existing.Name {
			return domain.ErrImmutableName
		}
		if err := existing.ApplyUpdate(cmd.Value, cmd.Label, cmd.Description, cmd.ActorID); err != nil {
			return err
		}
		updated, err := s.repo.Update(ctx, tx, existing)
		if err != nil {
			return err
		}
		out = updated
		return s.record(ctx, tx, cmd.TenantID, cmd.ActorID, "tokens.entry.updated", updated)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Delete deletes an entry after authz check and records an audit event.
func (s *Service) Delete(ctx context.Context, tenantID, actorID, id string) error {
	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := s.seed(ctx, tx, tenantID, actorID); err != nil {
			return err
		}
		if err := s.require(ctx, tx, string(iamdomain.CapTokenDictionaryManage), "tenant"); err != nil {
			return err
		}
		existing, err := s.repo.GetByID(ctx, tx, tenantID, id)
		if err != nil {
			return err
		}
		if err := s.repo.Delete(ctx, tx, tenantID, id); err != nil {
			return err
		}
		return s.record(ctx, tx, tenantID, actorID, "tokens.entry.deleted", existing)
	})
}

// Get fetches a single entry by ID. Read operations are not audited.
func (s *Service) Get(ctx context.Context, tenantID, id string) (*domain.Entry, error) {
	var out *domain.Entry
	err := s.runner.DoReadOnly(ctx, func(tx *sql.Tx) error {
		actorID := iamdomain.UserIDFromContext(ctx)
		if err := s.seed(ctx, tx, tenantID, actorID); err != nil {
			return err
		}
		if err := s.require(ctx, tx, string(iamdomain.CapTokenView), "tenant"); err != nil {
			return err
		}
		e, err := s.repo.GetByID(ctx, tx, tenantID, id)
		if err != nil {
			return err
		}
		out = e
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetByName satisfies domain.DictionaryReader. Read operations are not audited.
func (s *Service) GetByName(ctx context.Context, tenantID, name string) (*domain.Entry, error) {
	var out *domain.Entry
	err := s.runner.DoReadOnly(ctx, func(tx *sql.Tx) error {
		actorID := iamdomain.UserIDFromContext(ctx)
		if err := s.seed(ctx, tx, tenantID, actorID); err != nil {
			return err
		}
		if err := s.require(ctx, tx, string(iamdomain.CapTokenView), "tenant"); err != nil {
			return err
		}
		e, err := s.repo.GetByName(ctx, tx, tenantID, name)
		if err != nil {
			return err
		}
		out = e
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// List satisfies domain.DictionaryReader. Read operations are not audited.
func (s *Service) List(ctx context.Context, tenantID string) ([]domain.Entry, error) {
	var out []domain.Entry
	err := s.runner.DoReadOnly(ctx, func(tx *sql.Tx) error {
		actorID := iamdomain.UserIDFromContext(ctx)
		if err := s.seed(ctx, tx, tenantID, actorID); err != nil {
			return err
		}
		if err := s.require(ctx, tx, string(iamdomain.CapTokenView), "tenant"); err != nil {
			return err
		}
		items, err := s.repo.List(ctx, tx, tenantID)
		if err != nil {
			return err
		}
		out = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// record builds and persists an audit event inside the current transaction.
func (s *Service) record(ctx context.Context, tx *sql.Tx, tenantID, actorID, action string, e *domain.Entry) error {
	payload, err := json.Marshal(map[string]string{"id": e.ID, "name": e.Name})
	if err != nil {
		return fmt.Errorf("tokens: marshal audit payload: %w", err)
	}
	ev := auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       action,
		ResourceType: auditResourceType,
		ResourceID:   e.ID,
		PayloadJSON:  string(payload),
		TraceID:      requesttrace.Resolve(ctx),
		TenantID:     tenantID,
	}
	// *sql.Tx satisfies db.Tx structurally (ExecContext/QueryContext/QueryRowContext).
	var rec db.Tx = tx
	return s.audit.RecordTx(ctx, rec, ev)
}
