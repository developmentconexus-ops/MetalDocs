package repository

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the documents module's M7 F7.3 TenantDataPort. It owns
// public.documents, public.document_comments, public.document_exports,
// public.document_placeholder_values, and public.editor_sessions.
//
// Erase order: deleting public.documents alone is sufficient. Every other
// table in this port's set reaches documents.id via an ON DELETE CASCADE FK
// chain: document_comments, document_exports, editor_sessions, and
// document_revisions (not tenant-scoped itself — no tenant_id column, so out
// of the export/erasure census, but tenant-owned via its document_id FK) all
// cascade from documents; document_placeholder_values cascades one hop
// further from document_revisions. Rows Tables() reports as
// "public.document_comments" etc. are counted from the cascade via a
// pre-delete count, since ExecContext's RowsAffected only reports the
// directly-targeted table's row count, not cascaded rows.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the documents TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "documents" }

func (p *TenantDataPort) Tables() []string {
	return []string{
		"public.documents",
		"public.document_comments",
		"public.document_exports",
		"public.document_placeholder_values",
		"public.editor_sessions",
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

// countTenantRows returns the tenant's current row count for table (used to
// report cascade-deleted rows the DELETE statement's own RowsAffected would
// otherwise miss). Must be called BEFORE the cascading DELETE runs.
func countTenantRows(ctx context.Context, tx *sql.Tx, table, tenantColumn, tenantID string) (int64, error) {
	var n int64
	query := `SELECT count(*) FROM ` + table + ` WHERE ` + tenantColumn + ` = $1::uuid`
	if err := tx.QueryRowContext(ctx, query, tenantID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	counts := make(map[string]int64)

	// Pre-count the cascade-only tables (their own DELETE never runs
	// directly — public.documents' cascade removes them).
	for _, table := range []string{
		"public.document_comments",
		"public.document_exports",
		"public.document_placeholder_values",
		"public.editor_sessions",
	} {
		n, err := countTenantRows(ctx, tx, table, "tenant_id", tenantID)
		if err != nil {
			return nil, err
		}
		counts[table] = n
	}

	n, err := tenantdata.EraseTable(ctx, tx, "public.documents", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	counts["public.documents"] = n

	return counts, nil
}
