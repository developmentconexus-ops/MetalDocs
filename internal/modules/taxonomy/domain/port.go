// Package domain holds the taxonomy module's aggregates (DocumentFamily,
// DocumentProfile, ProcessArea), their sentinel errors, and the repository /
// governance-logger ports the application layer depends on. It has no
// database/sql import (CI guard nosqltxindomain) — persistence concerns stay
// behind FamilyTx and the repository interfaces defined here.
package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

// ProfileRepository is the persistence port for DocumentProfile. Tx-suffixed
// methods operate inside a caller-owned FamilyTx; non-Tx methods own their
// own transaction. Implementations wire authz.Require(CapTaxonomyManage) on
// write paths and CapTaxonomyView on reads.
type ProfileRepository interface {
	GetByCode(ctx context.Context, tenantID string, code ProfileCode) (*DocumentProfile, error)
	GetByCodeForUpdate(ctx context.Context, tx FamilyTx, tenantID string, code ProfileCode) (*DocumentProfile, error)
	List(ctx context.Context, tenantID string, includeArchived bool) ([]DocumentProfile, error)
	Create(ctx context.Context, p *DocumentProfile) error
	CreateTx(ctx context.Context, tx FamilyTx, p *DocumentProfile) error
	Update(ctx context.Context, p *DocumentProfile) error
	UpdateTx(ctx context.Context, tx FamilyTx, p *DocumentProfile) error
	BeginTx(ctx context.Context) (FamilyTx, error)
}

// AreaRepository is the persistence port for ProcessArea. Tx-suffixed
// methods operate inside a caller-owned FamilyTx; non-Tx methods own their
// own transaction. Implementations wire authz.Require(CapTaxonomyManage) on
// write paths and CapTaxonomyView on reads.
type AreaRepository interface {
	GetByCode(ctx context.Context, tenantID string, code AreaCode) (*ProcessArea, error)
	GetByCodeForUpdate(ctx context.Context, tx FamilyTx, tenantID string, code AreaCode) (*ProcessArea, error)
	List(ctx context.Context, tenantID string, includeArchived bool) ([]ProcessArea, error)
	Create(ctx context.Context, a *ProcessArea) error
	CreateTx(ctx context.Context, tx FamilyTx, a *ProcessArea) error
	Update(ctx context.Context, a *ProcessArea) error
	ListAncestors(ctx context.Context, tenantID string, code AreaCode) ([]AreaCode, error)
	ListAncestorsTx(ctx context.Context, tx FamilyTx, tenantID string, code AreaCode) ([]AreaCode, error)
	UpdateTx(ctx context.Context, tx FamilyTx, a *ProcessArea) error
	BeginTx(ctx context.Context) (FamilyTx, error)
}

// GovernanceLogger is the port taxonomy services use to record regulated
// mutations. AuditGovernanceAdapter is the sole implementation, routing
// events to metaldocs.audit_events.
type GovernanceLogger interface {
	// Log writes the governance event outside any caller transaction
	// (post-commit fire-and-forget from the caller's perspective).
	Log(ctx context.Context, e GovernanceEvent) error
	// LogTx writes the governance event inside an open transaction so the
	// audit record is atomically committed with the mutation that caused it
	// (REQ-ASYNC-1, F-07).
	LogTx(ctx context.Context, tx db.Tx, e GovernanceEvent) error
}

// GovernanceEventType identifies the kind of regulated mutation a
// GovernanceEvent records (e.g. profile.created, area.archived).
type GovernanceEventType string

// Governance event type constants emitted by FamilyService, ProfileService,
// and AreaService on every mutating operation.
const (
	GovernanceEventTypeProfileCreated               GovernanceEventType = "profile.created"
	GovernanceEventTypeProfileUpdated               GovernanceEventType = "profile.updated"
	GovernanceEventTypeProfileDefaultTemplateChange GovernanceEventType = "profile.default_template_change"
	GovernanceEventTypeProfileArchived              GovernanceEventType = "profile.archived"
	GovernanceEventTypeAreaCreated                  GovernanceEventType = "area.created"
	GovernanceEventTypeAreaUpdated                  GovernanceEventType = "area.updated"
	GovernanceEventTypeAreaParentChanged            GovernanceEventType = "area.parent_changed"
	GovernanceEventTypeAreaArchived                 GovernanceEventType = "area.archived"
	GovernanceEventTypeFamilyCreated                GovernanceEventType = "family.created"
	GovernanceEventTypeFamilyUpdated                GovernanceEventType = "family.updated"
	GovernanceEventTypeFamilyDeactivated            GovernanceEventType = "family.deactivated"
)

// GovernanceEvent is the audit row payload logged by GovernanceLogger for a
// single regulated mutation (create/update/archive/deactivate) on a family,
// profile, or area.
type GovernanceEvent struct {
	TenantID     string
	EventType    GovernanceEventType
	ActorUserID  string
	ResourceType string
	ResourceID   string
	Reason       string
	PayloadJSON  []byte
}

// FamilyTx is the minimal transaction handle services receive from repositories.
// It embeds db.Tx so the governance logger can write audit rows inside the same
// database transaction without an Unwrap shim.
type FamilyTx interface {
	db.Tx
	Commit() error
	Rollback() error
}

// FamilyRepository is the persistence port for DocumentFamily. Tx-suffixed
// methods operate inside a caller-owned FamilyTx; non-Tx methods own their
// own transaction. Implementations wire authz.Require(CapTaxonomyManage) on
// write paths and CapTaxonomyView on reads.
type FamilyRepository interface {
	GetByCode(ctx context.Context, tenantID string, code FamilyCode) (*DocumentFamily, error)
	List(ctx context.Context, tenantID string, includeInactive bool) ([]DocumentFamily, error)
	Create(ctx context.Context, f *DocumentFamily) error
	CreateTx(ctx context.Context, tx FamilyTx, f *DocumentFamily) error
	Update(ctx context.Context, f *DocumentFamily) error
	HasActiveProfiles(ctx context.Context, tenantID string, familyCode FamilyCode) (bool, error)
	BeginTx(ctx context.Context) (FamilyTx, error)
	GetByCodeForUpdate(ctx context.Context, tx FamilyTx, tenantID string, code FamilyCode) (*DocumentFamily, error)
	HasActiveProfilesTx(ctx context.Context, tx FamilyTx, tenantID string, familyCode FamilyCode) (bool, error)
	// Tx methods intentionally mirror the non-tx methods so services can stage
	// lock/read/write flows without leaking transaction details into callers.
	UpdateTx(ctx context.Context, tx FamilyTx, f *DocumentFamily) error
}
