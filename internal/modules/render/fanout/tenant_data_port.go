package fanout

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the render module's M7 F7.3 TenantDataPort. It owns the
// two staging dispatch outbox tables (metaldocs.pdf_dispatch_outbox,
// metaldocs.materialize_dispatch_outbox — the same pair
// stagingOutboxAllowlist in staging_outbox.go recognizes). Both are
// operational/ephemeral outbox rows with no long-term PII value (§3.2:
// delete rows). No FK dependencies in either direction.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the render TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "render" }

func (p *TenantDataPort) Tables() []string {
	return []string{
		"metaldocs.pdf_dispatch_outbox",
		"metaldocs.materialize_dispatch_outbox",
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
	for _, table := range p.Tables() {
		n, err := tenantdata.EraseTable(ctx, tx, table, "tenant_id", tenantID)
		if err != nil {
			return nil, err
		}
		counts[table] = n
	}
	return counts, nil
}
