package postgres

import (
	"context"
	"database/sql"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// ProcessAreaCatalog is the Postgres-backed iamapp.AreaCatalogReader. It checks
// that an areaCode exists in metaldocs.document_process_areas (the process-area
// single source of truth) for the caller's tenant before PeopleService burns an
// auth identity on an invite. This mirrors the (tenant_id, area_code) foreign
// key on user_process_areas that already backstops the grant path, surfacing an
// unknown area as a clean validation error at the boundary instead of letting it
// fail later as an opaque FK violation.
//
// M2/F2.3: document_process_areas is taxonomy-owned, so the existence read runs
// through the taxonomy-published AreaCatalogReader port (ADR-0039 D3(b)) instead
// of raw cross-module SQL. The check stays off-tx and non-recording — this
// adapter passes its own pool as the executor.
type ProcessAreaCatalog struct {
	db          *sql.DB
	areaCatalog taxonomydomain.AreaCatalogReader
}

// NewProcessAreaCatalog builds the catalog over the iam pool, delegating the
// existence read to the taxonomy area-catalog port. A nil port is defaulted to
// the Noop reader (reports absent).
func NewProcessAreaCatalog(db *sql.DB, areaCatalog taxonomydomain.AreaCatalogReader) *ProcessAreaCatalog {
	if areaCatalog == nil {
		areaCatalog = taxonomydomain.NoopAreaCatalogReader{}
	}
	return &ProcessAreaCatalog{db: db, areaCatalog: areaCatalog}
}

// AreaCodeExists reports whether areaCode is a known process area for
// tenantID, delegating to the taxonomy-published AreaCatalogReader port over
// the iam pool (off-tx, non-recording).
func (c *ProcessAreaCatalog) AreaCodeExists(ctx context.Context, tenantID, areaCode string) (bool, error) {
	return c.areaCatalog.AreaExists(ctx, c.db, tenantID, areaCode)
}
