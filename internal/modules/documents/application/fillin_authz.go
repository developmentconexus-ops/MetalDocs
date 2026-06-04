package application

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// requireDocEditDraft opens a short tx, sets authz GUCs, resolves the
// document's area_code, and calls authz.Require for document.edit (area-grade).
// The tx is rolled back after the check — no writes happen here.
// A missing/empty area resolves to "" → authz.Require denies non-system actors
// (fail-closed, ADR 0022 Phase 8; matches the Phase 7 document.create decision).
// ADR 0022 Phase 10 (F2): the redundant doc.edit_draft cap was merged into the
// canonical CapDocumentEdit — identical grant set, same area-grade enforcement.
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
	// document.edit is area-grade: pass the resolved area as-is. A missing document
	// or null area yields "" which fail-closes (authz.Require denies non-system
	// actors). ADR 0022 Phase 11 (F7): the per-file loadDocumentAreaCode was merged
	// into the shared LoadDocumentAreaCode.
	areaCode, _, err := LoadDocumentAreaCode(ctx, tx, tenantID, docID)
	if err != nil {
		return fmt.Errorf("fillin authz: load area: %w", err)
	}
	return authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode)
}
