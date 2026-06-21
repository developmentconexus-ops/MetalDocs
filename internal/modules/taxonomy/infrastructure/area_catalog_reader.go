package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/db"
)

// AreaCatalogReaderPG is the taxonomy-owned Postgres adapter for the
// area-catalog read-port (ADR-0039 D3(b)). It is stateless: every method runs
// its plain, non-recording SELECT on the db.DB executor the caller supplies, so
// one instance serves both the in-tx B7 caller (documents create tx) and the
// off-tx B8 caller (iam pre-invite check). It issues no authz GUC / capability
// assertion — the seam preserves the historical non-recording behavior and keeps
// the in-tx read safe under a lock (HS-PRE-1).
type AreaCatalogReaderPG struct{}

// NewAreaCatalogReaderPG builds the stateless adapter. Wired once at the
// composition root and injected into the documents repository and the iam
// process-area catalog.
func NewAreaCatalogReaderPG() *AreaCatalogReaderPG { return &AreaCatalogReaderPG{} }

var _ taxonomydomain.AreaCatalogReader = (*AreaCatalogReaderPG)(nil)

func (AreaCatalogReaderPG) AreaName(ctx context.Context, exec db.DB, tenantID, code string) (string, bool, error) {
	var name sql.NullString
	err := exec.QueryRowContext(ctx,
		`SELECT name FROM metaldocs.document_process_areas WHERE tenant_id=$1::uuid AND code=$2`,
		tenantID, code,
	).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("read process area name (tenant=%s code=%s): %w", tenantID, code, err)
	}
	// A present row with a NULL name is still "found"; the caller decides how to
	// store an empty name. (document_process_areas.name is NOT NULL today, so
	// name.Valid is always true here — this keeps the contract honest regardless.)
	return name.String, true, nil
}

func (AreaCatalogReaderPG) AreaExists(ctx context.Context, exec db.DB, tenantID, code string) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.document_process_areas
  WHERE tenant_id = $1::uuid
    AND code = $2
)`
	var exists bool
	if err := exec.QueryRowContext(ctx, q, tenantID, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("query process area existence (tenant=%s area=%s): %w", tenantID, code, err)
	}
	return exists, nil
}
