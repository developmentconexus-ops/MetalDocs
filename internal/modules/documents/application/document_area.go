package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
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
// cdRead is the controlleddocuments-owned read-port (ADR-0039 D3(b), M2/F2.1):
// the cd.process_area_code term of the resolver is read through it instead of
// joining controlled_documents (owned by another module). It is tx-aware — the
// caller's *sql.Tx is passed straight through as the executor, so the CD read
// runs inside the same transaction as the document read, preserving the prior
// JOIN's snapshot consistency. The COALESCE(snapshot, cd.process_area_code, ”)
// precedence is reproduced in Go: a non-NULL snapshot wins even when "", and
// only a NULL snapshot falls through to the CD area (a missing/NULL CD area ⇒ "").
func LoadDocumentAreaCode(ctx context.Context, tx *sql.Tx, cdRead controlleddocumentsdomain.CDFieldReader, tenantID, documentID string) (areaCode string, found bool, err error) {
	// Nil-safe: a caller that never wires the read-port (unit tests with no CD
	// area) gets the null object, which resolves the CD term to "" — the same
	// fail-closed "" the LEFT JOIN produced for an absent CD. Production wires the
	// real CDFieldReaderPG at the composition root.
	if cdRead == nil {
		cdRead = controlleddocumentsdomain.NoopCDFieldReader{}
	}
	var snapshot sql.NullString
	var cdID sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT process_area_code_snapshot, controlled_document_id
		   FROM documents WHERE id = $1 AND tenant_id = $2`,
		documentID, tenantID,
	).Scan(&snapshot, &cdID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load document area code: %w", err)
	}
	// Non-NULL snapshot wins outright (COALESCE first non-NULL term), including "".
	if snapshot.Valid {
		return snapshot.String, true, nil
	}
	// NULL snapshot: fall through to the controlled document's area via the port.
	// The port returns "" for a missing CD row or a NULL CD area, matching the
	// LEFT JOIN's COALESCE-to-'' tail.
	if cdID.Valid && cdID.String != "" {
		area, _, perr := cdRead.ProcessAreaCode(ctx, tx, tenantID, cdID.String)
		if perr != nil {
			return "", false, fmt.Errorf("load document area code: %w", perr)
		}
		return area, true, nil
	}
	return "", true, nil
}
