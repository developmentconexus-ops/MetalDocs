package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// requireDocEditDraft opens a short tx, sets authz GUCs, resolves the
// document's area_code, and calls authz.Require for doc.edit_draft (area-grade).
// The tx is rolled back after the check — no writes happen here.
// A missing/empty area resolves to "" → authz.Require denies non-system actors
// (fail-closed, ADR 0022 Phase 8; matches the Phase 7 document.create decision).
func requireDocEditDraft(ctx context.Context, db *sql.DB, tenantID, actorID, docID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fillin authz: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ctx = authz.WithCapCache(ctx)

	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return err
	}
	areaCode, err := loadDocumentAreaCode(ctx, tx, tenantID, docID)
	if err != nil {
		return fmt.Errorf("fillin authz: load area: %w", err)
	}
	return authz.Require(ctx, tx, string(iamdomain.CapDocumentEditDraft), areaCode)
}

func loadDocumentAreaCode(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (string, error) {
	var areaCode string
	err := tx.QueryRowContext(ctx, `
		SELECT process_area_code_snapshot
		  FROM documents
		 WHERE id = $1 AND tenant_id = $2`,
		documentID, tenantID,
	).Scan(&areaCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Fail-closed: empty area → authz.Require denies non-system actors
			// (area-grade caps only — doc.edit_draft, doc.reconstruct). ADR 0022 Phase 8.
			return "", nil
		}
		return "", err
	}
	return areaCode, nil
}
