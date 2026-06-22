package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// LoadDocumentControlledDocumentID reads the controlled_document_id for a document
// inside the caller's tx. Returns "" when the document has no CD link (NULL) or
// does not exist (not found is not an error — caller decides if "" is fatal).
// Used by F3.3 reader-event emit sites to build the LifecycleEventArgs payload.
func LoadDocumentControlledDocumentID(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (string, error) {
	var cdID sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT controlled_document_id FROM public.documents WHERE id = $1 AND tenant_id = $2`,
		documentID, tenantID,
	).Scan(&cdID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("load document controlled_document_id: %w", err)
	}
	if !cdID.Valid {
		return "", nil
	}
	return cdID.String, nil
}
