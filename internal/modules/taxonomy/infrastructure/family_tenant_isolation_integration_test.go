//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"errors"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// TestFamilyRepository_TenantIsolation is the SEC-01 regression guard: before
// metaldocs.document_families carried tenant_id, any actor holding
// taxonomy.manage in ANY tenant could see and mutate every other tenant's
// families (the table was globally shared — no RLS, no tenant predicate in
// any FamilyRepository query). This test proves the fix end-to-end through
// the real repository (setAuthzGUC -> authz.Require -> RLS), not just at the
// SQL layer:
//   - tenant A creates a family via FamilyRepository.Create
//   - tenant A can read it back via GetByCode
//   - tenant B (a different, real actor) gets ErrFamilyNotFound reading the
//     same code — RLS + the tenant-scoped WHERE clause both hide the row
//   - tenant B cannot mutate it either (Update against tenant B's context
//     reports not-found rather than silently succeeding cross-tenant)
//   - tenant A's List does not include a family created under tenant B
//
// system_admin short-circuits authz.Require's capability lookup (ADR 0022),
// so granting each test actor that role lets the test exercise the full
// tier-2 authz + RLS stack without also standing up a full role->capability
// fixture — the property under test is tenant isolation, not capability
// grants (those are covered by the authz package's own test suite).
func TestFamilyRepository_TenantIsolation(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	userA := testdb.NewUser(t, db, testdb.WithTenant(tenantA.ID), testdb.WithRole("system_admin"))
	userB := testdb.NewUser(t, db, testdb.WithTenant(tenantB.ID), testdb.WithRole("system_admin"))

	ctxA := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tenantA.ID), userA.ID, nil)
	ctxB := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), tenantB.ID), userB.ID, nil)

	repo := NewFamilyRepository(db)

	code := domain.FamilyCode("iso-" + tenantA.ID[:8])
	famA := &domain.DocumentFamily{
		Code:        code,
		TenantID:    tenantA.ID,
		Name:        "Tenant A Family",
		Description: "seeded by TestFamilyRepository_TenantIsolation",
		IsActive:    true,
	}
	if err := repo.Create(ctxA, famA); err != nil {
		t.Fatalf("tenant A Create: %v", err)
	}

	// Tenant A reads its own family back.
	got, err := repo.GetByCode(ctxA, tenantA.ID, code)
	if err != nil {
		t.Fatalf("tenant A GetByCode: %v", err)
	}
	if got.Name != "Tenant A Family" {
		t.Fatalf("tenant A GetByCode: name = %q, want %q", got.Name, "Tenant A Family")
	}

	// Tenant B, querying the SAME code, must not see tenant A's family.
	if _, err := repo.GetByCode(ctxB, tenantB.ID, code); !errors.Is(err, domain.ErrFamilyNotFound) {
		t.Fatalf("cross-tenant GetByCode: want ErrFamilyNotFound, got %v", err)
	}

	// Tenant B cannot mutate tenant A's family under its own tenant scope.
	mutateAttempt := &domain.DocumentFamily{
		Code:        code,
		TenantID:    tenantB.ID,
		Name:        "Hijacked",
		Description: "should never apply",
		IsActive:    true,
	}
	if err := repo.Update(ctxB, mutateAttempt); !errors.Is(err, domain.ErrFamilyNotFound) {
		t.Fatalf("cross-tenant Update: want ErrFamilyNotFound, got %v", err)
	}

	// Tenant A's family must remain unmutated (Update above must have been a no-op).
	stillA, err := repo.GetByCode(ctxA, tenantA.ID, code)
	if err != nil {
		t.Fatalf("tenant A re-read after cross-tenant Update attempt: %v", err)
	}
	if stillA.Name != "Tenant A Family" {
		t.Fatalf("cross-tenant Update leaked through: name = %q, want %q", stillA.Name, "Tenant A Family")
	}

	// Tenant B's List must not include tenant A's family.
	listB, err := repo.List(ctxB, tenantB.ID, true)
	if err != nil {
		t.Fatalf("tenant B List: %v", err)
	}
	for _, f := range listB {
		if f.Code == code {
			t.Fatalf("tenant B List leaked tenant A family %q", code)
		}
	}

	// Tenant A's List must include its own family.
	listA, err := repo.List(ctxA, tenantA.ID, true)
	if err != nil {
		t.Fatalf("tenant A List: %v", err)
	}
	found := false
	for _, f := range listA {
		if f.Code == code {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("tenant A List missing its own family %q", code)
	}
}
