package application

import (
	"context"
	"database/sql"
	"log/slog"

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

func (b *DocumentContextBuilder) loadDocumentSnapshot(ctx context.Context, tenantID, revisionID string) (areaCode, controlledDocumentID, areaName string, err error) {
	err = b.db.QueryRowContext(ctx,
		`SELECT coalesce(process_area_code_snapshot, ''), coalesce(controlled_document_id::text, ''), coalesce(area_name_snapshot, '') FROM documents WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, revisionID,
	).Scan(&areaCode, &controlledDocumentID, &areaName)
	return areaCode, controlledDocumentID, areaName, err
}

const activeInstanceSQL = `
SELECT id FROM approval_instances
 WHERE tenant_id = $1::uuid AND document_id = $2::uuid
   AND status IN ('approved', 'in_progress')
 ORDER BY submitted_at DESC LIMIT 1`

// Build returns the ResolveInput for a revision being approved.
func (b *DocumentContextBuilder) Build(ctx context.Context, tenantID, revisionID string, _ ApproverContext) (resolvers.ResolveInput, error) {
	areaCode, controlledDocumentID, areaName, err := b.loadDocumentSnapshot(ctx, tenantID, revisionID)
	if err != nil {
		return resolvers.ResolveInput{}, err
	}
	var instanceID sql.NullString
	if err := b.db.QueryRowContext(ctx, activeInstanceSQL, tenantID, revisionID).Scan(&instanceID); err != nil && err != sql.ErrNoRows {
		// Best-effort: proceed with empty instance ID; resolvers return nil approvers gracefully.
		slog.WarnContext(ctx, "approval instance lookup failed", "revision_id", revisionID, "err", err)
	}
	return resolvers.ResolveInput{
		TenantID:             tenantID,
		RevisionID:           revisionID,
		ControlledDocumentID: controlledDocumentID,
		AreaCodeSnapshot:     areaCode,
		AreaNameSnapshot:     areaName,
		ApprovalInstanceID:   instanceID.String,
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
