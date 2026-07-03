//go:build integration
// +build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	controlleddocumentsinfrastructure "metaldocs/internal/modules/controlleddocuments/infrastructure"
	docrepo "metaldocs/internal/modules/documents/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// TST-09 — tenant-isolation coverage for controlleddocuments. Multi-tenant
// pooled isolation (tenant_id everywhere, cross-tenant URL -> 404) is a
// non-negotiable invariant (CLAUDE.md); before this file the six tests below
// were unconditional t.Skip("requires live DB") stubs, leaving the module's
// tenant-scoped SQL guarded only by the DB tripwire suite (capability
// assertion, not tenant scoping). These tests exercise the real Postgres
// repository/service — not fakes — through the canonical testdb factory
// (ADR 0034), proving tenant B can never read, list, or create against
// tenant A's controlled-document rows.

func newIsolationRepo(t *testing.T, db *sql.DB) *controlleddocumentsinfrastructure.PostgresControlledDocumentRepository {
	t.Helper()
	return controlleddocumentsinfrastructure.NewPostgresControlledDocumentRepository(
		db, docrepo.NewActiveInstanceReaderPG(db),
	)
}

// TestTenantIsolation_GetByID_CrossTenant_NotFound pins the repository-level
// read-path invariant: GetByID is scoped by `WHERE tenant_id = $1 AND id =
// $2` (repository.go). A controlled document created under tenant A must be
// invisible to a GetByID call scoped to tenant B, even when the caller knows
// tenant A's real document id (no enumeration leak).
func TestTenantIsolation_GetByID_CrossTenant_NotFound(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	cdA := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenantA.ID))

	repo := newIsolationRepo(t, db)
	ctx := context.Background()

	gotA, err := repo.GetByID(ctx, tenantA.ID, cdA.ID)
	if err != nil {
		t.Fatalf("tenant A GetByID own document: %v", err)
	}
	if gotA.Code != cdA.Code {
		t.Fatalf("tenant A GetByID: code = %q, want %q", gotA.Code, cdA.Code)
	}

	_, err = repo.GetByID(ctx, tenantB.ID, cdA.ID)
	if !errors.Is(err, controlleddocumentsdomain.ErrCDNotFound) {
		t.Fatalf("cross-tenant GetByID(tenantB, cdA.ID) = %v, want ErrCDNotFound (tenant B must not see tenant A's document)", err)
	}
}

// TestTenantIsolation_GetByCode_CrossTenant_NotFound pins GetByCode, the
// lookup path used by manual/auto code-uniqueness checks. A document code is
// only unique within a tenant, so the SAME (profile_code, code) pair must
// resolve independently per tenant and never leak tenant A's row to tenant B.
func TestTenantIsolation_GetByCode_CrossTenant_NotFound(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	cdA := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenantA.ID))

	repo := newIsolationRepo(t, db)
	ctx := context.Background()

	gotA, err := repo.GetByCode(ctx, tenantA.ID, cdA.ProfileCode, cdA.Code)
	if err != nil {
		t.Fatalf("tenant A GetByCode own document: %v", err)
	}
	if gotA.ID != cdA.ID {
		t.Fatalf("tenant A GetByCode: id = %q, want %q", gotA.ID, cdA.ID)
	}

	_, err = repo.GetByCode(ctx, tenantB.ID, cdA.ProfileCode, cdA.Code)
	if !errors.Is(err, controlleddocumentsdomain.ErrCDNotFound) {
		t.Fatalf("cross-tenant GetByCode(tenantB, %q, %q) = %v, want ErrCDNotFound", cdA.ProfileCode, cdA.Code, err)
	}

	// CodeExists must agree: the same (profile_code, code) pair reads as
	// "taken" under tenant A and "free" under tenant B.
	existsA, err := repo.CodeExists(ctx, tenantA.ID, cdA.ProfileCode, cdA.Code)
	if err != nil {
		t.Fatalf("tenant A CodeExists: %v", err)
	}
	if !existsA {
		t.Fatalf("tenant A CodeExists(%q) = false, want true", cdA.Code)
	}
	existsB, err := repo.CodeExists(ctx, tenantB.ID, cdA.ProfileCode, cdA.Code)
	if err != nil {
		t.Fatalf("tenant B CodeExists: %v", err)
	}
	if existsB {
		t.Fatalf("cross-tenant CodeExists leaked: tenant B sees tenant A's code %q as taken", cdA.Code)
	}
}

// TestTenantIsolation_ListControlledDocuments_CrossTenant pins the List
// query path (repository.go List, `WHERE tenant_id = $1`): tenant B's list
// must never include a row created under tenant A, and tenant A's own list
// must still return it.
func TestTenantIsolation_ListControlledDocuments_CrossTenant(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	userA := testdb.NewUser(t, db, testdb.WithTenant(tenantA.ID))
	cdA := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenantA.ID), testdb.WithOwner(userA.ID))

	repo := newIsolationRepo(t, db)
	ctx := context.Background()

	listB, _, err := repo.List(ctx, tenantB.ID, controlleddocumentsdomain.CDFilter{Limit: 50})
	if err != nil {
		t.Fatalf("tenant B List: %v", err)
	}
	for _, doc := range listB {
		if doc.ID == cdA.ID {
			t.Fatalf("tenant B List leaked tenant A's controlled document %q", cdA.ID)
		}
	}

	listA, _, err := repo.List(ctx, tenantA.ID, controlleddocumentsdomain.CDFilter{Limit: 50})
	if err != nil {
		t.Fatalf("tenant A List: %v", err)
	}
	found := false
	for _, doc := range listA {
		if doc.ID == cdA.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("tenant A List missing its own controlled document %q", cdA.ID)
	}
}

// TestTenantIsolation_VisibilityGrants_CrossTenant pins the CD's own
// restricted-visibility grant tables (controlled_document_area_grants /
// controlled_document_user_grants). A restricted CD's area/user grants are
// tenant-scoped rows keyed by (tenant_id, controlled_document_id); this
// replaces the former "ListMemberships_CrossTenant" stub name, which named a
// concept (iam group/role memberships) this module does not own — the
// controlled-documents-owned analogue is the visibility grant set attached
// to a restricted document, resolved via CanRead/GetByID.
func TestTenantIsolation_VisibilityGrants_CrossTenant(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	outsiderB := testdb.NewUser(t, db, testdb.WithTenant(tenantB.ID))
	cdA := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenantA.ID), testdb.WithStatus("active"))

	repo := newIsolationRepo(t, db)
	ctx := context.Background()

	// A tenant-B actor (who does not exist in tenant A at all) must not be
	// able to read tenant A's document via CanRead, regardless of visibility
	// scope — the tenant_id predicate in CanRead's query must exclude the row
	// entirely for a foreign tenant id.
	canRead, err := repo.CanRead(ctx, tenantB.ID, cdA.ID, outsiderB.ID)
	if err != nil {
		t.Fatalf("tenant B CanRead against tenant A's document: %v", err)
	}
	if canRead {
		t.Fatalf("cross-tenant CanRead leaked: tenant B actor %q can read tenant A's document %q", outsiderB.ID, cdA.ID)
	}

	// GetByID under tenant B's scope must also fail closed (not-found), even
	// though the row exists (under tenant A).
	if _, err := repo.GetByID(ctx, tenantB.ID, cdA.ID); !errors.Is(err, controlleddocumentsdomain.ErrCDNotFound) {
		t.Fatalf("cross-tenant GetByID for restricted-visibility probe = %v, want ErrCDNotFound", err)
	}
}

// TestTenantIsolation_SequenceCounters_CrossTenant pins cd_sequence_counters
// isolation (T-005/RLS closed the schema-level gap; this proves the
// application-level NextAndIncrement/Peek/EnsureCounter API is independently
// keyed per tenant too — the SAME profile_code/process_area_code pair must
// allocate its own counter sequence per tenant, never sharing or leaking
// state across tenants).
func TestTenantIsolation_SequenceCounters_CrossTenant(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)

	// Deliberately reuse the SAME profile/area code strings across tenants —
	// the sequence counter's uniqueness key is (tenant_id, profile_code,
	// process_area_code), so if tenant scoping were broken these two tenants
	// would fight over one counter.
	const profileCode = "shared-profile"
	const areaCode = "shared-area"
	testdb.SeedGovernedTaxonomy(t, db, tenantA.ID, profileCode, areaCode)
	testdb.SeedGovernedTaxonomy(t, db, tenantB.ID, profileCode, areaCode)

	alloc := controlleddocumentsinfrastructure.NewPostgresSequenceAllocator(db)
	ctx := context.Background()

	if err := alloc.EnsureCounter(ctx, tenantA.ID, profileCode, areaCode); err != nil {
		t.Fatalf("tenant A EnsureCounter: %v", err)
	}
	if err := alloc.EnsureCounter(ctx, tenantB.ID, profileCode, areaCode); err != nil {
		t.Fatalf("tenant B EnsureCounter: %v", err)
	}

	runner := platformdb.NewTxRunner(db)

	// Advance tenant A's counter twice (1 -> 2 -> 3); tenant B's must remain
	// untouched at 1.
	for i := 0; i < 2; i++ {
		if err := runner.Do(ctx, func(tx *sql.Tx) error {
			_, err := alloc.NextAndIncrement(ctx, tx, tenantA.ID, profileCode, areaCode)
			return err
		}); err != nil {
			t.Fatalf("tenant A NextAndIncrement iteration %d: %v", i, err)
		}
	}

	peekA, err := alloc.Peek(ctx, tenantA.ID, profileCode, areaCode)
	if err != nil {
		t.Fatalf("tenant A Peek: %v", err)
	}
	if peekA != 3 {
		t.Fatalf("tenant A Peek = %d, want 3 (two allocations from base 1)", peekA)
	}

	peekB, err := alloc.Peek(ctx, tenantB.ID, profileCode, areaCode)
	if err != nil {
		t.Fatalf("tenant B Peek: %v", err)
	}
	if peekB != 1 {
		t.Fatalf("cross-tenant sequence leak: tenant B Peek = %d, want 1 (tenant A's allocations must not affect tenant B's counter)", peekB)
	}
}

// TestTenantIsolation_CreateCD_CrossTenantProfile_NotFound pins the
// service-level Create path's profile/area resolution
// (application/service.go Create -> s.profiles.GetByCode /
// s.areas.GetByCode). Both taxonomy repositories scope their SQL by
// tenant_id (taxonomy/infrastructure/repository.go GetByCode), so a create
// request naming a real profile/area code that exists only under a
// DIFFERENT tenant must fail as not-found — never silently create against
// the wrong tenant's taxonomy, and never leak whether the code exists
// elsewhere. This is the service-level analogue of the HTTP-level
// cross-tenant-URL-404 invariant (CLAUDE.md non-negotiable), replacing the
// former stub name "...Returns404" (404 is an HTTP concept; the service
// layer proves the underlying not-found error the handler maps to it).
func TestTenantIsolation_CreateCD_CrossTenantProfile_NotFound(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	taxA := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantA.ID))
	userB := testdb.NewUser(t, db, testdb.WithTenant(tenantB.ID))
	testdb.SeedSystemAdmin(t, db, tenantB.ID, userB.ID, "TST-09 cross-tenant probe")

	repo := newIsolationRepo(t, db)
	svc := NewControlledDocumentService(
		platformdb.NewTxRunner(db),
		repo,
		controlleddocumentsinfrastructure.NewPostgresSequenceAllocator(db),
		&fakeTemplateVersionChecker{},
		controlleddocumentsinfrastructure.NewTaxonomyProfileReader(taxonomyinfra.NewProfileRepository(db)),
		controlleddocumentsinfrastructure.NewTaxonomyAreaReader(taxonomyinfra.NewAreaRepository(db)),
		&fakeGovernanceLogger{},
		newInvariantReadyDocumentInitializer(),
	)

	ctx := iamdomain.WithAuthContext(context.Background(), userB.ID, nil)
	_, err := svc.Create(ctx, CreateControlledDocumentCmd{
		TenantID:         tenantB.ID,
		ProfileCode:      taxA.ProfileCode,     // exists only under tenant A
		ProcessAreaCode:  taxA.ProcessAreaCode, // exists only under tenant A
		Title:            "Cross-Tenant Profile Probe",
		OwnerUserID:      userB.ID,
		ActorUserID:      userB.ID,
		ManualCode:       stringPtr("PO-TST09-001"),
		ManualCodeReason: stringPtr("TST-09 tenant isolation coverage"),
		VisibilityScope:  "company",
	})
	if err == nil {
		t.Fatalf("Create with tenant B naming tenant A's profile/area codes succeeded, want not-found error (cross-tenant taxonomy leak)")
	}
	if !errors.Is(err, taxonomydomain.ErrProfileNotFound) {
		t.Fatalf("Create cross-tenant profile error = %v, want wrapping taxonomy.ErrProfileNotFound", err)
	}
}
