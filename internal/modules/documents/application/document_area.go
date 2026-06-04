package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// LoadDocumentAreaCode resolves a document's process area inside tx, preferring
// the document's own snapshot and falling back to its controlled-document's area.
//
//   - found reports whether the document row exists.
//   - areaCode is "" when the row is missing OR exists with no area set.
//
// This is the single document-keyed area resolver for the documents aggregate
// (ADR 0022 Phase 11, F7). It deliberately bakes in NEITHER the fail-closed ""
// nor the wide-open "tenant" default — the prior three near-identical helpers
// disagreed on empty-area semantics (two returned "" to deny, one COALESCEd to
// "tenant" to wide-open), a copy-paste fail-open footgun. Each call site now
// applies its own documented coalesce against the cap's grade:
//
//   - area-grade caps (document.edit / submit / signoff / publish / supersede):
//     pass areaCode as-is. "" fail-closes — authz.Require denies non-system
//     actors. Do NOT COALESCE to "tenant" (that silently disables area scoping;
//     the "tenant" sentinel is reserved and cannot be a real area — see the
//     document_process_areas guard).
//   - tenant-grade caps (document.view): COALESCE "" -> "tenant" at the call site
//     to keep the area filter intentionally OFF.
//
// The found bool lets a caller distinguish "no such document" from "document with
// no area" when those need different handling.
func LoadDocumentAreaCode(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (areaCode string, found bool, err error) {
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(d.process_area_code_snapshot, cd.process_area_code, '')
		  FROM documents d
		  LEFT JOIN controlled_documents cd
		    ON cd.id = d.controlled_document_id AND cd.tenant_id = d.tenant_id
		 WHERE d.id = $1 AND d.tenant_id = $2`,
		documentID, tenantID,
	).Scan(&areaCode)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load document area code: %w", err)
	}
	return areaCode, true, nil
}
