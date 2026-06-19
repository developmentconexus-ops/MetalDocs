package postgres

import (
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// TestUserTenantsQueryParity pins the UserTenantReader query to the exact shape it
// inherited from auth/infrastructure/postgres (H-G site #1, M5/F5.2). The move is
// parity-preserving by construction; this guard fails if a future edit drops
// DISTINCT, changes the table/predicate, or loses the deterministic ORDER BY.
func TestUserTenantsQueryParity(t *testing.T) {
	for _, want := range []string{
		"DISTINCT tenant_id",
		"FROM metaldocs.iam_user_roles",
		"WHERE user_id = $1",
		"ORDER BY tenant_id",
	} {
		if !strings.Contains(userTenantsQuery, want) {
			t.Fatalf("userTenantsQuery missing %q, got: %s", want, userTenantsQuery)
		}
	}
}

// TestNoopUserTenantReaderReturnsEmptyNonNil pins the null-object contract: empty,
// non-nil, no error.
func TestNoopUserTenantReaderReturnsEmptyNonNil(t *testing.T) {
	ids, err := iamdomain.NoopUserTenantReader{}.UserTenantIDs(nil, "anyone")
	if err != nil {
		t.Fatalf("Noop UserTenantIDs err = %v, want nil", err)
	}
	if ids == nil {
		t.Fatalf("Noop UserTenantIDs = nil, want empty non-nil slice")
	}
	if len(ids) != 0 {
		t.Fatalf("Noop UserTenantIDs = %v, want empty", ids)
	}
}
