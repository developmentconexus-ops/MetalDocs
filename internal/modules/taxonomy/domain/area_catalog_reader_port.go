package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

// AreaCatalogReader is the narrow read surface cross-module consumers use to read
// metaldocs.document_process_areas without reaching across the module boundary
// into taxonomy's owned base table (ADR-0039 D1/D3(b); H-G class, M2/F2.3).
// Taxonomy owns document_process_areas; callers that need an area's name or its
// existence depend on this port instead of issuing raw SQL against the table.
//
// Each method takes a db.DB executor so one stateless adapter serves both in-tx
// and off-tx callers: pass a *sql.Tx for an in-tx read (B7 — the documents
// create tx, after the revision-lock acquire) or a *sql.DB pool for an off-tx
// read (B8 — the iam pre-invite existence check). *sql.Tx and *sql.DB both
// satisfy db.DB structurally (internal/platform/db/tx.go). These are plain,
// non-recording SELECTs and MUST stay non-recording even in-tx (HS-PRE-1: no
// authz-recording read inside a lock-holding tx).
type AreaCatalogReader interface {
	// AreaName returns document_process_areas.name for (tenantID, code). found is
	// false when no such area row exists in that tenant — matching the
	// sql.ErrNoRows-tolerant B7 call site (missing → NULL area_name_snapshot).
	AreaName(ctx context.Context, exec db.DB, tenantID, code string) (name string, found bool, err error)

	// AreaExists reports whether a document_process_areas row exists for
	// (tenantID, code). The off-tx B8 boundary existence check; the iam caller
	// passes its own pool as exec.
	AreaExists(ctx context.Context, exec db.DB, tenantID, code string) (bool, error)
}

// NoopAreaCatalogReader is the explicit null-object for callers that never
// exercise an area-catalog read (tests, paths where no area code is present). It
// satisfies AreaCatalogReader and always reports empty/absent with nil errors.
type NoopAreaCatalogReader struct{}

// AreaName always reports not-found with a nil error.
func (NoopAreaCatalogReader) AreaName(_ context.Context, _ db.DB, _, _ string) (string, bool, error) {
	return "", false, nil
}

// AreaExists always reports false with a nil error.
func (NoopAreaCatalogReader) AreaExists(_ context.Context, _ db.DB, _, _ string) (bool, error) {
	return false, nil
}
