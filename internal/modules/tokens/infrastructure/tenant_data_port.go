package infrastructure

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the tokens module's M7 F7.3 TenantDataPort. It owns
// metaldocs.token_dictionary_entries (§3.2: business/authoring data — crypto-
// shred PII + delete rows; this task deletes rows). No FK dependencies in
// either direction.
type TenantDataPort struct {
	db *sql.DB
}

// NewTenantDataPort constructs the tokens TenantDataPort backed by db.
func NewTenantDataPort(db *sql.DB) *TenantDataPort {
	return &TenantDataPort{db: db}
}

var _ tenantdata.Port = (*TenantDataPort)(nil)

func (p *TenantDataPort) Module() string { return "tokens" }

func (p *TenantDataPort) Tables() []string {
	return []string{"metaldocs.token_dictionary_entries"}
}

func (p *TenantDataPort) ExportTenantData(ctx context.Context, db *sql.DB, tenantID string) ([]tenantdata.TableExport, error) {
	exp, err := tenantdata.ExportTable(ctx, db, "metaldocs.token_dictionary_entries", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return []tenantdata.TableExport{exp}, nil
}

func (p *TenantDataPort) EraseTenantData(ctx context.Context, tx *sql.Tx, tenantID string) (map[string]int64, error) {
	n, err := tenantdata.EraseTable(ctx, tx, "metaldocs.token_dictionary_entries", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return map[string]int64{"metaldocs.token_dictionary_entries": n}, nil
}
