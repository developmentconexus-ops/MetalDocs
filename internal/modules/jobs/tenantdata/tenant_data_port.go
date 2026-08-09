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

// Port is the jobs module's M7 F7.3 TenantDataPort. It owns
// metaldocs.idempotency_keys (§3.2: operational/ephemeral, TTL'd — delete
// rows). No FK dependencies in either direction.
type Port struct {
	db *sql.DB
}

// NewTenantDataPort constructs the jobs TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *Port {
	return &Port{db: db}
}

var _ tenantdata.Port = (*Port)(nil)

// Module returns the owning module name ("jobs"), identifying this port in
// the tenant-data-export/erase registry.
func (p *Port) Module() string { return "jobs" }

// Tables returns the fully qualified tables this port owns.
func (p *Port) Tables() []string {
	return []string{"metaldocs.idempotency_keys"}
}

// ExportTenantData exports every metaldocs.idempotency_keys row for
// tenantID.
func (p *Port) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	exp, err := tenantdata.ExportTable(ctx, db, "metaldocs.idempotency_keys", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return []tenantdata.TableExport{exp}, nil
}

// EraseTenantData deletes every metaldocs.idempotency_keys row for tenantID
// within tx, returning the per-table row count erased.
func (p *Port) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	n, err := tenantdata.EraseTable(ctx, tx, "metaldocs.idempotency_keys", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return map[string]int64{"metaldocs.idempotency_keys": n}, nil
}
