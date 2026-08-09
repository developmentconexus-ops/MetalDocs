package infrastructure

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the controlleddocuments module's M7 F7.3 TenantDataPort.
// It owns public.controlled_documents, public.controlled_document_area_grants,
// public.controlled_document_user_grants, and public.cd_sequence_counters.
//
// Erase order: deleting public.controlled_documents alone cascades to both
// grant tables (ON DELETE CASCADE on controlled_document_id); their rows are
// pre-counted (same idiom as the documents port) since a cascaded delete's
// row count isn't reported by the parent DELETE's RowsAffected.
// cd_sequence_counters is an independent leaf (no FK from controlled_documents)
// and is deleted directly.
//
// Cross-module FK note (orchestrator concern): public.documents references
// controlled_documents(id) with a plain (non-cascading) FK — the documents
// port's EraseTenantData MUST run BEFORE this port in the orchestrator's
// fan-out sequence, or this port's DELETE on controlled_documents would fail
// with a dangling-reference violation.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the controlleddocuments TenantDataPort backed
// by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

// Module returns the owning module name, "controlleddocuments", for
// tenantdata.Port registration.
func (p *TenantDataPort) Module() string { return "controlleddocuments" }

// Tables returns the controlleddocuments module's tenant-scoped table census:
// controlled_documents, its two grant tables, and cd_sequence_counters.
func (p *TenantDataPort) Tables() []string {
	return []string{
		"public.controlled_documents",
		"public.controlled_document_area_grants",
		"public.controlled_document_user_grants",
		"public.cd_sequence_counters",
	}
}

// ExportTenantData exports every table in Tables() for tenantID.
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

// EraseTenantData erases controlled_documents (cascading to its two grant
// tables) and cd_sequence_counters for tenantID, returning erased row counts
// per table. See the type doc comment for erase-order and cross-module FK
// dependencies.
func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := make(map[string]int64)

	for _, table := range []string{
		"public.controlled_document_area_grants",
		"public.controlled_document_user_grants",
	} {
		var n int64
		if err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM `+table+` WHERE tenant_id = $1::uuid`, tenantID,
		).Scan(&n); err != nil {
			return nil, err
		}
		counts[table] = n
	}

	n, err := tenantdata.EraseTable(ctx, tx, "public.controlled_documents", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.controlled_documents"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "public.cd_sequence_counters", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.cd_sequence_counters"] = n

	return counts, nil
}
