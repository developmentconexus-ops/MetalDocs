//go:build integration

package documents_integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"metaldocs/internal/testdb"
)

func TestVDocumentFinalized_ReturnsApprovalChangedAt(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()

	tenantID, docID := testdb.InsertDraftDocument(t, db, nil, testdb.DevTenantID)

	approvedAt := time.Now().UTC().Truncate(time.Microsecond)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO metaldocs.document_state_history
			(document_id, from_status, to_status, changed_by, changed_at)
		VALUES ($1, 'under_review', 'approved', 'tester', $2)`,
		docID, approvedAt); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	var got sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT finalized_at FROM metaldocs.v_document_finalized WHERE document_id=$1`,
		docID,
	).Scan(&got); err != nil {
		t.Fatalf("query view: %v", err)
	}
	if !got.Valid || !got.Time.Equal(approvedAt) {
		t.Fatalf("finalized_at = %v, want %v", got, approvedAt)
	}
	_ = tenantID
}
