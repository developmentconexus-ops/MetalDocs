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

func TestFamilyRepository_Create_SetsAuthzGUCBeforeRequire(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	tnt := testdb.NewTenant(t, db)
	repo := NewFamilyRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tnt.ID), "actor-1", nil)

	fam := &domain.DocumentFamily{
		Code:        domain.FamilyCode("procedure"),
		Name:        "Procedimentos",
		Description: "Familia global para fluxo de editor",
		IsActive:    true,
	}

	err := repo.Create(ctx, fam)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestProfileRepository_Create_SetsAuthzGUCBeforeRequire(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	tnt := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	repo := NewProfileRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tnt.ID), "actor-1", nil)

	prof := &domain.DocumentProfile{
		Code:               domain.ProfileCode("pop"),
		TenantID:           tnt.ID,
		FamilyCode:         domain.FamilyCode(tax.FamilyCode),
		Name:               "Procedimento Operacional",
		Description:        "descricao",
		Alias:              "pop",
		ReviewIntervalDays: 365,
		EditableByRole:     "admin",
	}

	err := repo.Create(ctx, prof)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestAreaRepository_Create_SetsAuthzGUCBeforeRequire(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	tnt := testdb.NewTenant(t, db)
	repo := NewAreaRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tnt.ID), "actor-1", nil)

	area := &domain.ProcessArea{
		Code:        domain.AreaCode("general"),
		TenantID:    tnt.ID,
		Name:        "Geral",
		Description: "descricao",
	}

	err := repo.Create(ctx, area)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestFamilyRepository_GetByCodeForUpdate_RequiresAuthzBeforeLock(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	tnt := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	repo := NewFamilyRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tnt.ID), "actor-1", nil)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	defer tx.Rollback()

	_, err = repo.GetByCodeForUpdate(ctx, tx, domain.FamilyCode(tax.FamilyCode))
	if err != nil {
		t.Fatalf("GetByCodeForUpdate: %v", err)
	}
}

func TestProfileRepository_GetByCodeForUpdate_RequiresAuthzBeforeLock(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	tnt := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	repo := NewProfileRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tnt.ID), "actor-1", nil)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	defer tx.Rollback()

	_, err = repo.GetByCodeForUpdate(ctx, tx, tnt.ID, domain.ProfileCode(tax.ProfileCode))
	if err != nil {
		t.Fatalf("GetByCodeForUpdate: %v", err)
	}
}

func TestAreaRepository_GetByCodeForUpdate_RequiresAuthzBeforeLock(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	tnt := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	repo := NewAreaRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tnt.ID), "actor-1", nil)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	defer tx.Rollback()

	_, err = repo.GetByCodeForUpdate(ctx, tx, tnt.ID, domain.AreaCode(tax.ProcessAreaCode))
	if err != nil {
		t.Fatalf("GetByCodeForUpdate: %v", err)
	}
}
