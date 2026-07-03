package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

// CDFieldReader is the narrow read surface cross-module consumers use to fetch
// individual controlled_documents fields without reaching across the module
// boundary into controlleddocuments' owned base table (ADR-0039 D1/D3(b);
// H-G class, M2/F2.1). controlleddocuments owns metaldocs.controlled_documents;
// callers that need one of its fields depend on this port instead of issuing
// raw SQL against the table.
//
// Each method takes a db.DB executor so one stateless adapter serves both
// off-tx and in-tx callers: pass a *sql.DB (pool) for an off-tx read, or the
// caller's *sql.Tx for an in-tx read — *sql.Tx satisfies db.DB structurally
// (internal/platform/db/tx.go), so the read runs inside the caller's existing
// transaction with no adapter. These are plain, non-recording SELECTs and MUST
// stay non-recording even when run in-tx (HS-PRE-1: no authz-recording read
// inside a lock-holding tx).
type CDFieldReader interface {
	// ProfileCode returns controlled_documents.profile_code for (tenantID, cdID).
	// Returns ("", nil) when the controlled document is absent in that tenant —
	// matching the current sql.ErrNoRows-tolerant call site (B1).
	ProfileCode(ctx context.Context, exec db.DB, tenantID, cdID string) (string, error)

	// ProcessAreaCode returns controlled_documents.process_area_code for
	// (tenantID, cdID). found is false when the controlled document is absent
	// (the caller coalesces a missing row to ""), distinguishing it from a row
	// present with an empty area. Used as the LEFT-JOIN fallback term in the
	// document/instance area-resolver COALESCE chains (B5, B6).
	ProcessAreaCode(ctx context.Context, exec db.DB, tenantID, cdID string) (areaCode string, found bool, err error)
}

// NoopCDFieldReader is the explicit null-object for callers that never exercise
// a controlled-document field read (tests, paths where the CD link is absent).
// It satisfies CDFieldReader and always returns empty results with nil errors.
// Mirrors iam/domain.NoopUserDisplayNameReader.
type NoopCDFieldReader struct{}

// ProfileCode always returns ("", nil).
func (NoopCDFieldReader) ProfileCode(_ context.Context, _ db.DB, _, _ string) (string, error) {
	return "", nil
}

// ProcessAreaCode always returns ("", false, nil).
func (NoopCDFieldReader) ProcessAreaCode(_ context.Context, _ db.DB, _, _ string) (string, bool, error) {
	return "", false, nil
}
