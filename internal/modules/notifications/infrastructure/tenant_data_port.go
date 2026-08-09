package notificationsinfra

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the notifications module's M7 F7.3 TenantDataPort. It
// owns metaldocs.notifications — operational/ephemeral, no long-term PII
// value (§3.2: delete rows). No FK dependencies in either direction.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the notifications TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

// Module returns the notifications module's tenant-data port identifier.
func (p *TenantDataPort) Module() string { return "notifications" }

// Tables returns the tables this port owns for tenant data export/erasure.
func (p *TenantDataPort) Tables() []string {
	return []string{"metaldocs.notifications"}
}

// ExportTenantData exports all metaldocs.notifications rows for tenantID.
func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	exp, err := tenantdata.ExportTable(ctx, db, "metaldocs.notifications", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return []tenantdata.TableExport{exp}, nil
}

// EraseTenantData deletes all metaldocs.notifications rows for tenantID.
func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	n, err := tenantdata.EraseTable(ctx, tx, "metaldocs.notifications", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return map[string]int64{"metaldocs.notifications": n}, nil
}
