package wiring

import (
	"context"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/render/resolvers"
)

// controlledDocumentsReaderAdapter bridges the controlled-document repository
// to resolvers.RegistryReader.
type controlledDocumentsReaderAdapter struct {
	repo interface {
		GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)
	}
}

// NewControlledDocumentsReader returns a resolvers.RegistryReader backed by the given repository.
func NewControlledDocumentsReader(repo interface {
	GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)
}) resolvers.RegistryReader {
	return controlledDocumentsReaderAdapter{repo: repo}
}

func (a controlledDocumentsReaderAdapter) GetControlledDocument(ctx context.Context, tenantID resolvers.TenantID, controlledDocumentID resolvers.ControlledDocumentID) (resolvers.ControlledDocumentInfo, error) {
	cd, err := a.repo.GetByID(ctx, string(tenantID), string(controlledDocumentID))
	if err != nil {
		return resolvers.ControlledDocumentInfo{}, err
	}
	return resolvers.ControlledDocumentInfo{DocCode: cd.Code}, nil
}

// searchRevisionReaderAdapter bridges a revision reader to resolvers.RevisionReader.
type searchRevisionReaderAdapter struct {
	reader interface {
		GetRevisionNumber(ctx context.Context, tenantID, revisionID string) (int64, error)
		GetEffectiveFrom(ctx context.Context, tenantID, revisionID string) (time.Time, error)
		GetAuthor(ctx context.Context, tenantID, revisionID string) (resolvers.AuthorInfo, error)
	}
}

// NewSearchRevisionReader returns a resolvers.RevisionReader backed by the given reader.
func NewSearchRevisionReader(reader interface {
	GetRevisionNumber(ctx context.Context, tenantID, revisionID string) (int64, error)
	GetEffectiveFrom(ctx context.Context, tenantID, revisionID string) (time.Time, error)
	GetAuthor(ctx context.Context, tenantID, revisionID string) (resolvers.AuthorInfo, error)
}) resolvers.RevisionReader {
	return searchRevisionReaderAdapter{reader: reader}
}

func (a searchRevisionReaderAdapter) GetRevisionNumber(ctx context.Context, tenantID resolvers.TenantID, revisionID resolvers.RevisionID) (int64, error) {
	return a.reader.GetRevisionNumber(ctx, string(tenantID), string(revisionID))
}

func (a searchRevisionReaderAdapter) GetEffectiveFrom(ctx context.Context, tenantID resolvers.TenantID, revisionID resolvers.RevisionID) (time.Time, error) {
	return a.reader.GetEffectiveFrom(ctx, string(tenantID), string(revisionID))
}

func (a searchRevisionReaderAdapter) GetAuthor(ctx context.Context, tenantID resolvers.TenantID, revisionID resolvers.RevisionID) (resolvers.AuthorInfo, error) {
	return a.reader.GetAuthor(ctx, string(tenantID), string(revisionID))
}

// searchWorkflowReaderAdapter bridges a workflow reader to resolvers.WorkflowReader.
type searchWorkflowReaderAdapter struct {
	reader interface {
		GetApprovers(ctx context.Context, tenantID, revisionID, approvalInstanceID string) ([]resolvers.ApproverInfo, error)
		GetFinalApprovalDate(ctx context.Context, tenantID, revisionID string) (time.Time, error)
	}
}

// NewSearchWorkflowReader returns a resolvers.WorkflowReader backed by the given reader.
func NewSearchWorkflowReader(reader interface {
	GetApprovers(ctx context.Context, tenantID, revisionID, approvalInstanceID string) ([]resolvers.ApproverInfo, error)
	GetFinalApprovalDate(ctx context.Context, tenantID, revisionID string) (time.Time, error)
}) resolvers.WorkflowReader {
	return searchWorkflowReaderAdapter{reader: reader}
}

func (a searchWorkflowReaderAdapter) GetApprovers(ctx context.Context, tenantID resolvers.TenantID, revisionID resolvers.RevisionID, approvalInstanceID resolvers.ApprovalInstanceID) ([]resolvers.ApproverInfo, error) {
	return a.reader.GetApprovers(ctx, string(tenantID), string(revisionID), string(approvalInstanceID))
}

func (a searchWorkflowReaderAdapter) GetFinalApprovalDate(ctx context.Context, tenantID resolvers.TenantID, revisionID resolvers.RevisionID) (time.Time, error) {
	return a.reader.GetFinalApprovalDate(ctx, string(tenantID), string(revisionID))
}

// searchDocumentReaderAdapter bridges a document title reader to resolvers.DocumentReader.
type searchDocumentReaderAdapter struct {
	reader interface {
		GetDocumentTitle(ctx context.Context, tenantID, revisionID string) (string, error)
	}
}

// NewSearchDocumentReader returns a resolvers.DocumentReader backed by the given reader.
func NewSearchDocumentReader(reader interface {
	GetDocumentTitle(ctx context.Context, tenantID, revisionID string) (string, error)
}) resolvers.DocumentReader {
	return searchDocumentReaderAdapter{reader: reader}
}

func (a searchDocumentReaderAdapter) GetDocumentTitle(ctx context.Context, tenantID resolvers.TenantID, revisionID resolvers.RevisionID) (string, error) {
	return a.reader.GetDocumentTitle(ctx, string(tenantID), string(revisionID))
}
