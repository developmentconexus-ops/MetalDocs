package resolvers

import (
	"context"
	"fmt"
	"time"
)

type TenantID string
type RevisionID string
type ControlledDocumentID string
type ApprovalInstanceID string

type ResolveInput struct {
	TenantID             TenantID
	RevisionID           RevisionID
	ControlledDocumentID ControlledDocumentID
	ProfileCodeSnapshot  string
	AreaCodeSnapshot     string
	AreaNameSnapshot     string             // D2 (populated in Phase 3)
	ApprovalInstanceID   ApprovalInstanceID // D8
	RegistryReader       RegistryReader
	RevisionReader       RevisionReader
	WorkflowReader       WorkflowReader
	DocumentReader       DocumentReader
}

type ResolvedValue struct {
	// Value is resolver-specific; built-in resolvers currently return string or int64.
	Value       any
	ResolverKey string
	ResolverVer int
	InputsHash  []byte
	ComputedAt  time.Time
}

type ComputedResolver interface {
	Key() string
	Version() int
	Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error)
}

type ControlledDocumentInfo struct {
	DocCode string
}

type AuthorInfo struct {
	UserID      string
	DisplayName string
}

type ApproverInfo struct {
	UserID      string
	DisplayName string
	SignedAt    time.Time
}

type RegistryReader interface {
	GetControlledDocument(ctx context.Context, tenantID TenantID, controlledDocumentID ControlledDocumentID) (ControlledDocumentInfo, error)
}

type RevisionReader interface {
	GetRevisionNumber(ctx context.Context, tenantID TenantID, revisionID RevisionID) (int64, error)
	GetEffectiveFrom(ctx context.Context, tenantID TenantID, revisionID RevisionID) (time.Time, error)
	GetAuthor(ctx context.Context, tenantID TenantID, revisionID RevisionID) (AuthorInfo, error)
}

type WorkflowReader interface {
	GetApprovers(ctx context.Context, tenantID TenantID, revisionID RevisionID, approvalInstanceID ApprovalInstanceID) ([]ApproverInfo, error)
	// GetFinalApprovalDate is currently revision-scoped only; callers cannot
	// disambiguate multiple approval instances for the same revision here yet.
	GetFinalApprovalDate(ctx context.Context, tenantID TenantID, revisionID RevisionID) (time.Time, error)
}

type DocumentReader interface {
	GetDocumentTitle(ctx context.Context, tenantID TenantID, revisionID RevisionID) (string, error)
}

func requireTenantID(resolverKey string, tenantID TenantID) error {
	if tenantID == "" {
		return fmt.Errorf("%s: TenantID is required", resolverKey)
	}
	return nil
}

func requireRevisionID(resolverKey string, revisionID RevisionID) error {
	if revisionID == "" {
		return fmt.Errorf("%s: RevisionID is required", resolverKey)
	}
	return nil
}

func requireControlledDocumentID(resolverKey string, controlledDocumentID ControlledDocumentID) error {
	if controlledDocumentID == "" {
		return fmt.Errorf("%s: ControlledDocumentID is required", resolverKey)
	}
	return nil
}
