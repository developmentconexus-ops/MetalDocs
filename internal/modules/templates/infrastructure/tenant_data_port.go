package infrastructure

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the templates module's M7 F7.3 TenantDataPort. It owns
// public.templates_template and public.templates_template_version.
//
// Erase order handles one same-module wrinkle the tables' own FKs don't
// solve automatically: templates_template.published_version_id ->
// templates_template_version.id and templates_template_version.template_id ->
// templates_template.id form a two-table cycle (no ON DELETE CASCADE either
// direction). published_version_id is nulled out first so
// templates_template_version can then be deleted. (A second wrinkle —
// public.templates_approval_config as a non-cascading in-module child — was
// retired with the legacy template-approval path: the table is dropped by
// db/migrations/0302_drop_templates_approval_config.sql, ADR 0082 phase c /
// unit 3.1a S5.)
//
// Cross-module FK note (orchestrator concern): controlled_documents and
// metaldocs.document_profiles both reference templates_template_version(id)
// with plain (non-cascading) FKs — the controlleddocuments and taxonomy
// ports' EraseTenantData MUST run BEFORE this port in the orchestrator's
// fan-out sequence.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the templates TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

// Module returns the owning module name, "templates".
func (p *TenantDataPort) Module() string { return "templates" }

// Tables returns the tenant-scoped tables this port owns: templates_template
// and templates_template_version.
func (p *TenantDataPort) Tables() []string {
	return []string{
		"public.templates_template",
		"public.templates_template_version",
	}
}

// ExportTenantData exports every row of this port's owned tables for
// tenantID, one tenantdata.TableExport per table.
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

// EraseTenantData deletes tenantID's rows from this port's owned tables
// within tx, breaking the templates_template <-> templates_template_version
// cycle first (see the type doc), and returns the per-table deleted-row
// counts.
func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := make(map[string]int64)

	// Break the templates_template <-> templates_template_version cycle.
	if _, err := tx.ExecContext(ctx, `
UPDATE public.templates_template
   SET published_version_id = NULL
 WHERE tenant_id = $1::uuid AND published_version_id IS NOT NULL`, tenantID); err != nil {
		return nil, err
	}

	n, err := tenantdata.EraseTable(ctx, tx, "public.templates_template_version", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.templates_template_version"] = n

	n, err = tenantdata.EraseTable(ctx, tx, "public.templates_template", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.templates_template"] = n

	return counts, nil
}
