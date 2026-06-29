//go:build integration
// +build integration

package application_test

import (
	"errors"
	"testing"

	"metaldocs/internal/modules/render/resolvers"
	"metaldocs/internal/modules/tokens/application"
	"metaldocs/internal/modules/tokens/domain"
	"metaldocs/tests/integration/testdb"
)

// buildReservedSet returns a staticReserved containing the 8 native keys from
// the render resolver registry. Used at the service layer for integration tests
// that validate the reserved-name guard with production-equivalent wiring without
// needing a full HTTP stack.
func buildReservedSet() staticReserved {
	reg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(reg)
	known := reg.Known()
	set := make(staticReserved, len(known))
	for k := range known {
		set[k] = struct{}{}
	}
	return set
}

// TestCreate_ReservedName_RegistryKey_Rejected verifies that a name matching a
// native resolver key (e.g. "author") is rejected with ErrReservedName and that
// the tx never opens (guard fires off-tx). Production wiring: real DB, real
// registry keys, no stub.
func TestCreate_ReservedName_RegistryKey_Rejected(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	svc := buildSvc(db).WithReservedNames(buildReservedSet())

	_, err := svc.Create(ctxFor(user.ID), application.CreateCommand{
		TenantID: tn.ID,
		ActorID:  user.ID,
		Name:     "author",
		Value:    "v",
		Label:    "l",
	})
	if !errors.Is(err, domain.ErrReservedName) {
		t.Fatalf("Create with reserved name 'author' = %v, want ErrReservedName", err)
	}
}

// TestCreate_ReservedName_ApprovalDate_Rejected verifies the approval_date key
// (absent from templates' placeholder catalog) is also blocked — guards SP-2 N1.
func TestCreate_ReservedName_ApprovalDate_Rejected(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	svc := buildSvc(db).WithReservedNames(buildReservedSet())

	_, err := svc.Create(ctxFor(user.ID), application.CreateCommand{
		TenantID: tn.ID,
		ActorID:  user.ID,
		Name:     "approval_date",
		Value:    "v",
		Label:    "l",
	})
	if !errors.Is(err, domain.ErrReservedName) {
		t.Fatalf("Create with reserved name 'approval_date' = %v, want ErrReservedName", err)
	}
}

// TestCreate_NonReservedName_Succeeds verifies that a non-native name (e.g.
// "company_name") passes the guard and results in a successful creation.
func TestCreate_NonReservedName_Succeeds(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	svc := buildSvc(db).WithReservedNames(buildReservedSet())

	entry, err := svc.Create(ctxFor(user.ID), application.CreateCommand{
		TenantID: tn.ID,
		ActorID:  user.ID,
		Name:     "company_name",
		Value:    "Acme Corp",
		Label:    "Company name",
	})
	if err != nil {
		t.Fatalf("Create with non-reserved name = %v, want nil", err)
	}
	if entry.ID == "" {
		t.Fatal("Create returned entry with empty ID")
	}
}
