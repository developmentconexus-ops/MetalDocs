package application

import (
	"context"
	"database/sql"

	"metaldocs/internal/modules/render/resolvers"
)

// DocumentContextBuilder builds resolvers.ResolveInput for a document revision.
// It loads the document's area code from the DB and wires the reader dependencies
// needed by built-in computed-placeholder resolvers.
type DocumentContextBuilder struct {
	db             *sql.DB
	revReader      resolvers.RevisionReader
	workflowReader resolvers.WorkflowReader
	registryReader resolvers.RegistryReader
	documentReader resolvers.DocumentReader
}

// NewDocumentContextBuilder wires a DocumentContextBuilder.
// registryReader may be nil if no computed placeholders use the doc_code resolver.
func NewDocumentContextBuilder(
	db *sql.DB,
	revReader resolvers.RevisionReader,
	workflowReader resolvers.WorkflowReader,
	registryReader resolvers.RegistryReader,
	documentReader resolvers.DocumentReader,
) *DocumentContextBuilder {
	return &DocumentContextBuilder{
		db:             db,
		revReader:      revReader,
		workflowReader: workflowReader,
		registryReader: registryReader,
		documentReader: documentReader,
	}
}

func (b *DocumentContextBuilder) loadDocumentSnapshot(ctx context.Context, tenantID, revisionID string) (areaCode, controlledDocumentID string, err error) {
	err = b.db.QueryRowContext(ctx,
		`SELECT coalesce(process_area_code_snapshot, ''), coalesce(controlled_document_id::text, '') FROM documents WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, revisionID,
	).Scan(&areaCode, &controlledDocumentID)
	return areaCode, controlledDocumentID, err
}

// Build returns the ResolveInput for a revision being approved.
func (b *DocumentContextBuilder) Build(ctx context.Context, tenantID, revisionID string, _ ApproverContext) (resolvers.ResolveInput, error) {
	areaCode, controlledDocumentID, err := b.loadDocumentSnapshot(ctx, tenantID, revisionID)
	if err != nil {
		return resolvers.ResolveInput{}, err
	}
	return resolvers.ResolveInput{
		TenantID:             tenantID,
		RevisionID:           revisionID,
		ControlledDocumentID: controlledDocumentID,
		AreaCodeSnapshot:     areaCode,
		RevisionReader:       b.revReader,
		WorkflowReader:       b.workflowReader,
		RegistryReader:       b.registryReader,
		DocumentReader:       b.documentReader,
	}, nil
}

// BuildForDraft returns the ResolveInput for a draft revision (no approver context).
func (b *DocumentContextBuilder) BuildForDraft(ctx context.Context, tenantID, revisionID string) (resolvers.ResolveInput, error) {
	return b.Build(ctx, tenantID, revisionID, ApproverContext{})
}
