package repository

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the templates module's M7 F7.3 TenantDataPort. It owns
// public.templates_template and public.templates_template_version.
//
// Erase order handles two same-module wrinkles neither table's own FK
// solves automatically:
//   - templates_template.published_version_id -> templates_template_version.id
//     and templates_template_version.template_id -> templates_template.id form
//     a two-table cycle (no ON DELETE CASCADE either direction). published_version_id
//     is nulled out first so templates_template_version can then be deleted.
//   - public.templates_approval_config (template_id -> templates_template.id,
//     no cascade, no tenant_id column of its own — out of this port's
//     Tables() census but an in-module child table, so invariant 6 still
//     permits touching it here) is deleted before templates_template or the
//     final DELETE would fail on a dangling-reference violation.
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

func (p *TenantDataPort) Module() string { return "templates" }

func (p *TenantDataPort) Tables() []string {
	return []string{
		"public.templates_template",
		"public.templates_template_version",
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

	// Break the templates_template <-> templates_template_version cycle.
	if _, err := tx.ExecContext(ctx, `
UPDATE public.templates_template
   SET published_version_id = NULL
 WHERE tenant_id = $1::uuid AND published_version_id IS NOT NULL`, tenantID); err != nil {
		return nil, err
	}

	// templates_approval_config has no tenant_id of its own (not in
	// Tables()) but is an in-module child of templates_template that would
	// otherwise block the final DELETE.
	if _, err := tx.ExecContext(ctx, `
DELETE FROM public.templates_approval_config
 WHERE template_id IN (SELECT id FROM public.templates_template WHERE tenant_id = $1::uuid)`,
		tenantID,
	); err != nil {
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
