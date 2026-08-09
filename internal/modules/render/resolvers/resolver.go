package resolvers

import (
	"context"
	"fmt"
	"time"
)

// TenantID identifies the tenant a resolver computation is scoped to.
type TenantID string

// RevisionID identifies the controlled-document revision being resolved.
type RevisionID string

// ControlledDocumentID identifies the parent controlled document of a revision.
type ControlledDocumentID string

// ApprovalInstanceID identifies the approval workflow instance a resolver may
// need to disambiguate approvers/approval dates against.
type ApprovalInstanceID string

// ResolveInput carries the identifiers, snapshot fields, and reader ports a
// ComputedResolver needs to compute its value.
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

// ResolvedValue is the output of a ComputedResolver: the computed value plus
// the provenance metadata (resolver identity/version, inputs hash, timestamp)
// needed to audit and reproduce it.
type ResolvedValue struct {
	// Value is resolver-specific; built-in resolvers currently return string or int64.
	Value       any
	ResolverKey string
	ResolverVer int
	InputsHash  []byte
	ComputedAt  time.Time
}

// ComputedResolver computes a single template placeholder value from a
// ResolveInput. Implementations are registered in a Registry under Key.
type ComputedResolver interface {
	Key() string
	Version() int
	Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error)
}

// ControlledDocumentInfo is the subset of a controlled document's registry
// data resolvers need (currently just its code).
type ControlledDocumentInfo struct {
	DocCode string
}

// AuthorInfo identifies the author of a revision for display purposes.
type AuthorInfo struct {
	UserID      string
	DisplayName string
}

// ApproverInfo identifies one approver of a revision, including when they
// signed off.
type ApproverInfo struct {
	UserID      string
	DisplayName string
	SignedAt    time.Time
}

// RegistryReader reads controlled-document registry data needed by resolvers.
type RegistryReader interface {
	GetControlledDocument(ctx context.Context, tenantID TenantID, controlledDocumentID ControlledDocumentID) (ControlledDocumentInfo, error)
}

// RevisionReader reads revision-scoped data needed by resolvers.
type RevisionReader interface {
	GetRevisionNumber(ctx context.Context, tenantID TenantID, revisionID RevisionID) (int64, error)
	GetEffectiveFrom(ctx context.Context, tenantID TenantID, revisionID RevisionID) (time.Time, error)
	GetAuthor(ctx context.Context, tenantID TenantID, revisionID RevisionID) (AuthorInfo, error)
}

// WorkflowReader reads approval-workflow data needed by resolvers.
type WorkflowReader interface {
	GetApprovers(ctx context.Context, tenantID TenantID, revisionID RevisionID, approvalInstanceID ApprovalInstanceID) ([]ApproverInfo, error)
	// GetFinalApprovalDate is currently revision-scoped only; callers cannot
	// disambiguate multiple approval instances for the same revision here yet.
	GetFinalApprovalDate(ctx context.Context, tenantID TenantID, revisionID RevisionID) (time.Time, error)
}

// DocumentReader reads document-level data needed by resolvers.
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
