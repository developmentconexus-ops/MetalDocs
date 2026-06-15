//go:build integration

package infrastructure

import (
	"context"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

func TestFamilyRepository_HasActiveProfiles_TenantScoped(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	tnt := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))

	repo := NewFamilyRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tnt.ID), "actor-1", nil)

	exists, err := repo.HasActiveProfiles(ctx, tnt.ID, domain.FamilyCode(tax.FamilyCode))
	if err != nil {
		t.Fatalf("HasActiveProfiles: %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true")
	}
}
