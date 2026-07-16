package fanout

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/platform/messaging"
)

// PDFPipelineStateReader reports whether the async PDF pipeline (materialize
// fanout + PDF convert) has permanently died for a given document/revision,
// by checking metaldocs.outbox_events for a dead-lettered row on either hop.
// It matches the documents/application.PDFOutboxStateReader port exactly:
// ReadState(ctx, tenantID, revisionID) (string, error), returning "failed"
// when either hop is dead-lettered, "" when not (caller treats "" as
// pending), and the error unwrapped on any query failure (fail closed — no
// substitute value).
type PDFPipelineStateReader struct {
	db *sql.DB
}

// NewPDFPipelineStateReader returns a reader bound to db.
func NewPDFPipelineStateReader(db *sql.DB) *PDFPipelineStateReader {
	return &PDFPipelineStateReader{db: db}
}

// ReadState returns "failed" when metaldocs.outbox_events has a
// dead-lettered row for either the docx_materialize or docgen_v2_pdf event
// type, scoped to this aggregate (revisionID) and tenant (payload->>'tenant_id',
// outbox_events carries no tenant_id column). Returns "" when no such row
// exists.
func (r *PDFPipelineStateReader) ReadState(ctx context.Context, tenantID, revisionID string) (string, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1 FROM metaldocs.outbox_events
   WHERE aggregate_id = $1
     AND event_type IN ($2, $3)
     AND payload->>'tenant_id' = $4
     AND dead_lettered_at IS NOT NULL
)`,
		revisionID,
		string(messaging.EventTypeMaterializeFanout),
		string(messaging.EventTypePDFConvert),
		tenantID,
	).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("pdf pipeline state: read: %w", err)
	}
	if exists {
		return "failed", nil
	}
	return "", nil
}
