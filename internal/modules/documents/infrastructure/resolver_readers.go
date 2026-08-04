package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
)

// RevisionReader implements resolvers.RevisionReader backed by the documents table.
type RevisionReader struct{ db *sql.DB }

func NewRevisionReader(db *sql.DB) *RevisionReader { return &RevisionReader{db: db} }

func (r *RevisionReader) GetRevisionNumber(ctx context.Context, tenantID, revisionID string) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx,
		`SELECT revision_number FROM documents WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, revisionID).Scan(&n)
	return n, err
}

func (r *RevisionReader) GetEffectiveFrom(ctx context.Context, tenantID, revisionID string) (time.Time, error) {
	var nt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT effective_from FROM documents WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, revisionID).Scan(&nt)
	if err != nil {
		return time.Time{}, err
	}
	if !nt.Valid {
		return time.Time{}, nil
	}
	return nt.Time, nil
}

func (r *RevisionReader) GetAuthor(ctx context.Context, tenantID, revisionID string) (resolvers.AuthorInfo, error) {
	var userID string
	var displayName string
	err := r.db.QueryRowContext(ctx,
		`SELECT created_by::text, COALESCE(created_by_display_name_snapshot, created_by::text) FROM documents WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, revisionID).Scan(&userID, &displayName)
	if err != nil {
		return resolvers.AuthorInfo{}, err
	}
	return resolvers.AuthorInfo{UserID: userID, DisplayName: displayName}, nil
}

func (r *RevisionReader) GetDocumentTitle(ctx context.Context, tenantID, revisionID string) (string, error) {
	var title string
	err := r.db.QueryRowContext(ctx,
		`SELECT name FROM documents WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, revisionID,
	).Scan(&title)
	return title, err
}

// WorkflowReader implements resolvers.WorkflowReader backed by approval tables.
type WorkflowReader struct{ db *sql.DB }

func NewWorkflowReader(db *sql.DB) *WorkflowReader { return &WorkflowReader{db: db} }

// AutoApprovalDisplayName is what the approvers block renders for an instance
// that reached terminal approval with NO signoffs — the normal case for a
// livre profile, whose active route carries zero stages and therefore
// auto-approves on submit (ADR 0087). Rendering an empty approvers block there
// would read as "nobody approved this", which is wrong and unauditable: the
// document IS approved, by an explicitly configured no-approval route.
const AutoApprovalDisplayName = "Aprovação automática — rota livre"

// ErrNoApprovalDate is returned when an approval date is requested for a
// document that has neither a signoff nor a completed approval instance. It
// fails closed on purpose (no-fallback principle): the previous
// coalesce(..., now()) fabricated a date, and since ADR 0087 the no-signoff
// case is the NORMAL livre path rather than a rare anomaly, so a fabricated
// date would silently become the common one.
var ErrNoApprovalDate = errors.New("documents: no approval date for document (no signoff and no completed approval instance)")

func (r *WorkflowReader) GetApprovers(ctx context.Context, tenantID, revisionID, approvalInstanceID string) ([]resolvers.ApproverInfo, error) {
	if approvalInstanceID == "" {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.actor_user_id, COALESCE(s.actor_display_name_snapshot, s.actor_user_id::text), s.signed_at
		  FROM approval_signoffs s
		 WHERE s.actor_tenant_id = $1::uuid
		   AND s.approval_instance_id = $2::uuid
		   AND s.decision = 'approve'
		 ORDER BY s.signed_at ASC`,
		tenantID, approvalInstanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []resolvers.ApproverInfo
	for rows.Next() {
		var a resolvers.ApproverInfo
		if err := rows.Scan(&a.UserID, &a.DisplayName, &a.SignedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) > 0 {
		return out, nil
	}

	// No signoffs. That alone does NOT mean "auto-approved": a review-only
	// route (e.g. a simples profile whose stages are all review-kind) reaches
	// terminal approval through review verdicts and legitimately produces zero
	// signoff rows. Attributing those to an automatic livre approval would be a
	// false record. So the auto-approval line is gated on the instance's ACTUAL
	// pinned shape — the stage instances materialized at submit — and rendered
	// only when that shape is stageless, which is exactly the ADR 0087 livre
	// route. Every other no-signoff case keeps the pre-existing behavior
	// (empty approvers block).
	var completedAt sql.NullTime
	var submittedBy sql.NullString
	var stageCount int
	err = r.db.QueryRowContext(ctx, `
		SELECT ai.completed_at,
		       ai.submitted_by::text,
		       (SELECT count(*) FROM approval_stage_instances si
		         WHERE si.approval_instance_id = ai.id)
		  FROM approval_instances ai
		 WHERE ai.tenant_id = $1::uuid
		   AND ai.id = $2::uuid
		   AND ai.status = 'approved'`,
		tenantID, approvalInstanceID).Scan(&completedAt, &submittedBy, &stageCount)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, err
	case stageCount > 0:
		// Approved through a staged route without signoffs (review verdicts).
		// Nothing signoff-shaped to render; preserve prior behavior.
		return nil, nil
	case !completedAt.Valid:
		// Approved but no completion timestamp: integrity-critical, so fail
		// closed rather than render a half-truth.
		return nil, ErrNoApprovalDate
	}
	return []resolvers.ApproverInfo{{
		UserID:      submittedBy.String,
		DisplayName: AutoApprovalDisplayName,
		SignedAt:    completedAt.Time,
	}}, nil
}

// GetFinalApprovalDate returns the document's approval date: the last approve
// signoff when the route had stages, else the approval instance's completion
// timestamp (ADR 0087 — a livre profile's zero-stage route auto-approves and
// produces no signoff at all). Never fabricates a date; a document with
// neither yields ErrNoApprovalDate.
func (r *WorkflowReader) GetFinalApprovalDate(ctx context.Context, tenantID, revisionID string) (time.Time, error) {
	var at sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(s.signed_at), MAX(ai.completed_at))
		  FROM approval_instances ai
		  LEFT JOIN approval_signoffs s
		         ON s.approval_instance_id = ai.id
		        AND s.decision = 'approve'
		 WHERE ai.tenant_id = $1::uuid
		   AND ai.document_id = $2::uuid
		   AND ai.status = 'approved'`,
		tenantID, revisionID).Scan(&at)
	if err != nil {
		return time.Time{}, err
	}
	if !at.Valid {
		return time.Time{}, ErrNoApprovalDate
	}
	return at.Time, nil
}

// FanoutInputsReader reads stored fanout inputs for forensic reconstruction.
type FanoutInputsReader struct{ db *sql.DB }

func NewFanoutInputsReader(db *sql.DB) *FanoutInputsReader { return &FanoutInputsReader{db: db} }

// ReadForReconstruction loads the snapshot + stored values needed to re-render a frozen revision.
// Returns the FanoutRequest to replay and the original content_hash for comparison.
func (r *FanoutInputsReader) ReadForReconstruction(ctx context.Context, tenantID, revisionID string) (fanout.FanoutRequest, []byte, error) {
	var bodyDocxKey string
	var compositionRaw []byte
	var contentHash []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT coalesce(body_docx_snapshot_s3_key, ''),
		       coalesce(composition_config_snapshot, '{}')::text,
		       content_hash
		  FROM documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, revisionID).Scan(&bodyDocxKey, &compositionRaw, &contentHash)
	if err != nil {
		return fanout.FanoutRequest{}, nil, err
	}

	fillIn := NewFillInRepository(r.db)

	values, err := fillIn.ListValues(ctx, tenantID, revisionID)
	if err != nil {
		return fanout.FanoutRequest{}, nil, err
	}
	placeholders := make(map[string]string, len(values))
	for _, v := range values {
		if v.ValueText != nil {
			placeholders[v.PlaceholderID] = *v.ValueText
		}
	}

	req := fanout.FanoutRequest{
		TenantID:          tenantID,
		RevisionID:        revisionID,
		BodyDocxS3Key:     bodyDocxKey,
		PlaceholderValues: placeholders,
		Composition:       json.RawMessage(compositionRaw),
	}
	return req, contentHash, nil
}
