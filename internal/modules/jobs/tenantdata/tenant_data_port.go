// Package tenantdata hosts the jobs module's M7 F7.3 TenantDataPort. It
// lives in its own subpackage (rather than any of the existing
// job-implementation packages, e.g. idempotency_janitor) because it is not
// tied to one janitor — it owns metaldocs.idempotency_keys, whose canonical
// Postgres store lives in the neutral internal/platform/idempotency package,
// not under any single internal/modules/jobs/* janitor.
package tenantdata

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the jobs module's M7 F7.3 TenantDataPort. It owns
// metaldocs.idempotency_keys (§3.2: operational/ephemeral, TTL'd — delete
// rows). No FK dependencies in either direction.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the jobs TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "jobs" }

func (p *TenantDataPort) Tables() []string {
	return []string{"metaldocs.idempotency_keys"}
}

func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	exp, err := tenantdata.ExportTable(ctx, db, "metaldocs.idempotency_keys", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return []tenantdata.TableExport{exp}, nil
}

func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	n, err := tenantdata.EraseTable(ctx, tx, "metaldocs.idempotency_keys", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return map[string]int64{"metaldocs.idempotency_keys": n}, nil
}
