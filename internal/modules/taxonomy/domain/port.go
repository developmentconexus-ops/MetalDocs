package domain

import "context"

type ProfileRepository interface {
	GetByCode(ctx context.Context, tenantID, code string) (*DocumentProfile, error)
	List(ctx context.Context, tenantID string, includeArchived bool) ([]DocumentProfile, error)
	Create(ctx context.Context, p *DocumentProfile) error
	Update(ctx context.Context, p *DocumentProfile) error
}

type AreaRepository interface {
	GetByCode(ctx context.Context, tenantID, code string) (*ProcessArea, error)
	List(ctx context.Context, tenantID string, includeArchived bool) ([]ProcessArea, error)
	Create(ctx context.Context, a *ProcessArea) error
	Update(ctx context.Context, a *ProcessArea) error
	ListAncestors(ctx context.Context, tenantID, code string) ([]string, error)
}

type GovernanceLogger interface {
	Log(ctx context.Context, e GovernanceEvent) error
}

type GovernanceEvent struct {
	TenantID     string
	EventType    string
	ActorUserID  string
	ResourceType string
	ResourceID   string
	Reason       string
	PayloadJSON  []byte
}

type FamilyTx interface {
	Commit() error
	Rollback() error
}

type FamilyRepository interface {
	GetByCode(ctx context.Context, code string) (*DocumentFamily, error)
	List(ctx context.Context, includeInactive bool) ([]DocumentFamily, error)
	Create(ctx context.Context, f *DocumentFamily) error
	Update(ctx context.Context, f *DocumentFamily) error
	HasActiveProfiles(ctx context.Context, tenantID, familyCode string) (bool, error)
	BeginTx(ctx context.Context) (FamilyTx, error)
	GetByCodeForUpdate(ctx context.Context, tx FamilyTx, code string) (*DocumentFamily, error)
	HasActiveProfilesTx(ctx context.Context, tx FamilyTx, tenantID, familyCode string) (bool, error)
	UpdateTx(ctx context.Context, tx FamilyTx, f *DocumentFamily) error
}
