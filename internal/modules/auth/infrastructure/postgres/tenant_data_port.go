package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/platform/tenantdata"
)

// TenantDataPort is the auth module's M7 F7.3 TenantDataPort. It erases two
// tables with two different strategies:
//
//   - metaldocs.auth_sessions: a direct tenant_id-predicate DELETE
//     (operational/ephemeral session rows, §3.2 "delete rows", no
//     long-term PII value). This table also carries the port's Tables()
//     entry and its census/coverage registration.
//   - metaldocs.auth_identities: the credential table (identifier + password
//     hash — personal data), erased via a JOIN through metaldocs.iam_users
//     (auth_identities has no tenant_id column of its own; it is keyed by
//     user_id and scoped indirectly through iam_users). This DELETE MUST run
//     BEFORE the iam TenantDataPort's own erase step, because it depends on
//     iam_users rows for tenantID still existing to resolve the join — see
//     tenant_lifecycle_service.go's eraseOrder, which places "auth" before
//     "iam". auth_identities is deliberately NOT added to Tables(): it has no
//     tenant_id/actor_tenant_id column, so the coverage census
//     (tests/integration/tenantdata/coverage_test.go) never expects it
//     there — adding it would break that test's exact-equality assertion.
//
// auth_identities is also deliberately excluded from ExportTenantData: it
// holds credential material (password hashes), which must never appear in a
// tenant data export. The identity-adjacent data a tenant export legitimately
// needs (user ids, display names) is already covered by the iam module's
// export of metaldocs.iam_users.
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
	// auth_identities has no tenant_id column, so it cannot use
	// tenantdata.EraseTable's direct-predicate helper: it is scoped via a
	// join through iam_users. This MUST run before auth_sessions' own
	// erase (order within this method is not itself load-bearing) but,
	// more importantly, before the iam TenantDataPort runs at all — see
	// eraseOrder in tenant_lifecycle_service.go, which places "auth"
	// before "iam" specifically so this join still resolves.
	const eraseIdentitiesQuery = `
DELETE FROM metaldocs.auth_identities
 WHERE user_id IN (SELECT user_id FROM metaldocs.iam_users WHERE tenant_id = $1::uuid)
`
	res, err := tx.ExecContext(ctx, eraseIdentitiesQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenantdata: erase metaldocs.auth_identities: %w", err)
	}
	identitiesErased, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("tenantdata: erase metaldocs.auth_identities: %w", err)
	}

	sessionsErased, err := tenantdata.EraseTable(ctx, tx, "metaldocs.auth_sessions", "tenant_id", tenantID)
	if err != nil {
		return nil, err
	}
	return map[string]int64{
		"metaldocs.auth_sessions":   sessionsErased,
		"metaldocs.auth_identities": identitiesErased,
	}, nil
}
