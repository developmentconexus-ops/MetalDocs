package postgres

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the auth module's M7 F7.3 TenantDataPort. It owns
// metaldocs.auth_sessions — operational/ephemeral session rows (§3.2:
// "delete rows", no long-term PII value). auth_identities (the credential
// table) carries no tenant_id column of its own (it is keyed by user_id and
// scoped indirectly through iam_users), so it is out of this port's
// TenantDataPort scope and not touched here.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the auth TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "auth" }

func (p *TenantDataPort) Tables() []string {
	return []string{"metaldocs.auth_sessions"}
}

func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	exp, err := tenantdata.ExportTable(ctx, db, "metaldocs.auth_sessions", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return []tenantdata.TableExport{exp}, nil
}

func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	n, err := tenantdata.EraseTable(ctx, tx, "metaldocs.auth_sessions", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return map[string]int64{"metaldocs.auth_sessions": n}, nil
}
