package infrastructure

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the taxonomy module's M7 F7.3 TenantDataPort. It owns
// metaldocs.document_families, metaldocs.document_process_areas, and
// metaldocs.document_profiles — the governed taxonomy chain.
//
// Cross-module FK note (orchestrator concern, not this port's): profiles and
// process_areas are referenced by other modules' tables (approval_routes,
// cd_sequence_counters, controlled_documents, controlled_document_area_grants,
// public.user_process_areas). The tenant-erasure orchestrator (a later task)
// MUST erase controlleddocuments/documents/iam's referencing rows before
// calling this port, or these DELETEs will fail on FK violations — flagged
// here as a dependency the fan-out order must respect, not solved by this
// port in isolation (invariant 6: this port only touches its own tables).
//
// Within taxonomy's own three tables the erase order is profiles first
// (references families), then process_areas (self-referencing via
// parent_code, deleted as a single statement since Postgres evaluates FK
// constraints per-row within one DELETE), then families last.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the taxonomy TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "taxonomy" }

func (p *TenantDataPort) Tables() []string {
	return []string{
		"metaldocs.document_families",
		"metaldocs.document_process_areas",
		"metaldocs.document_profiles",
	}
}

func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	var out []tenantdata.TableExport
	for _, table := range p.Tables() {
		exp, err := tenantdata.ExportTable(ctx, db, table, "tenant_id", tenantID)
		if err != nil {
			return nil, err
		}
		out = append(out, exp)
	}
	return out, nil
}

func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := make(map[string]int64)

	n, err := tenantdata.EraseTable(ctx, tx, "metaldocs.document_profiles", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.document_profiles"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.document_process_areas", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.document_process_areas"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "metaldocs.document_families", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["metaldocs.document_families"] = n

	return counts, nil
}
