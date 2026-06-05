package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// ProcessAreaCatalog is the Postgres-backed iamapp.AreaCatalogReader. It checks
// that an areaCode exists in metaldocs.document_process_areas (the process-area
// single source of truth) for the caller's tenant before PeopleService burns an
// auth identity on an invite. This mirrors the (tenant_id, area_code) foreign
// key on user_process_areas that already backstops the grant path, surfacing an
// unknown area as a clean validation error at the boundary instead of letting it
// fail later as an opaque FK violation.
type ProcessAreaCatalog struct {
	db *sql.DB
}

func NewProcessAreaCatalog(db *sql.DB) *ProcessAreaCatalog {
	return &ProcessAreaCatalog{db: db}
}

func (c *ProcessAreaCatalog) AreaCodeExists(ctx context.Context, tenantID, areaCode string) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.document_process_areas
  WHERE tenant_id = $1::uuid
    AND code = $2
)`
	var exists bool
	if err := c.db.QueryRowContext(ctx, q, tenantID, areaCode).Scan(&exists); err != nil {
		return false, fmt.Errorf("query process area existence (tenant=%s area=%s): %w", tenantID, areaCode, err)
	}
	return exists, nil
}
