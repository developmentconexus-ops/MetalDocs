//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	controlleddocumentsinfrastructure "metaldocs/internal/modules/controlleddocuments/infrastructure"
	docrepo "metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// F0.2 — manual-code CD create must enforce tier-2 area-scoped authz.
// Before the symmetric refactor the manual branch had no tx/identity, so
// authz.Require failed-closed on MustActorID for every non-system-admin caller
// (B2: ErrActorContextMissing). After the refactor the manual branch mirrors
// the auto branch: SeedTxIdentity + Require(controlled_documents.create, area)
// + CreateTx inside s.runner.Do.

// seedActorWithGrant provisions a tenant, taxonomy, and actor user. When role
// != "" it also inserts a user_process_areas grant of that role in the seeded
// area (effective_from=now, NULL effective_to). The role is wired through the
// curated role_capabilities mapping — e.g. "author" carries
// controlled_documents.create. Returns (tenantID, actorUserID, profileCode,
// areaCode).
func seedActorWithGrant(t *testing.T, db *sql.DB, role string) (string, string, string, string) {
	t.Helper()
	tenant := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	user := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))

	if role != "" {
		testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(),
				`INSERT INTO metaldocs.user_process_areas
				    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
				 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
				user.ID, tenant.ID, tax.ProcessAreaCode, role,
			)
			return err
		})
	}
	return tenant.ID, user.ID, tax.ProfileCode, tax.ProcessAreaCode
}

func newManualCodeServiceForIntegration(t *testing.T, db *sql.DB) *ControlledDocumentService {
	t.Helper()
	repo := controlleddocumentsinfrastructure.NewPostgresControlledDocumentRepository(db, docrepo.NewActiveInstanceReaderPG(db))
	return NewControlledDocumentService(
		platformdb.NewTxRunner(db),
		repo,
		&fakeSequenceAllocator{next: 1},
		&fakeTemplateVersionChecker{},
		&fakeProfileReader{},
		&fakeAreaReader{},
		&fakeGovernanceLogger{},
		newInvariantReadyDocumentInitializer(),
	)
}

func manualCmd(tenantID, actorUserID, profile, area string) CreateControlledDocumentCmd {
	return CreateControlledDocumentCmd{
		TenantID:         tenantID,
		ProfileCode:      profile,
		ProcessAreaCode:  area,
		Title:            "F0.2 Manual Code Integration",
		OwnerUserID:      actorUserID,
		ActorUserID:      actorUserID,
		ManualCode:       stringPtr("PO-F02-001"),
		ManualCodeReason: stringPtr("F0.2 integration coverage for manual code create"),
		VisibilityScope:  "company",
	}
}

func TestCreate_ManualCode_NonAdminWithCap_Succeeds(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, profile, area := seedActorWithGrant(t, db, "author")

	svc := newManualCodeServiceForIntegration(t, db)
	res, err := svc.Create(context.Background(), manualCmd(tenantID, actorID, profile, area))
	if err != nil {
		t.Fatalf("non-admin author with controlled_documents.create in area: Create = %v, want nil", err)
	}
	if res == nil || res.ControlledDocument == nil {
		t.Fatalf("Create returned nil result for granted actor")
	}
	if res.ControlledDocument.Code != "PO-F02-001" {
		t.Fatalf("Code = %q, want PO-F02-001", res.ControlledDocument.Code)
	}
}

func TestCreate_ManualCode_NonAdminWithoutCap_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, profile, area := seedActorWithGrant(t, db, "")

	svc := newManualCodeServiceForIntegration(t, db)
	_, err := svc.Create(context.Background(), manualCmd(tenantID, actorID, profile, area))
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("non-admin without cap: Create = %v, want authz.ErrCapDenied", err)
	}
}

func TestCreate_ManualCode_SystemAdmin_Succeeds(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, profile, area := seedActorWithGrant(t, db, "")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "F0.2 System Admin")

	svc := newManualCodeServiceForIntegration(t, db)
	res, err := svc.Create(context.Background(), manualCmd(tenantID, actorID, profile, area))
	if err != nil {
		t.Fatalf("system_admin manual-code Create = %v, want nil", err)
	}
	if res == nil || res.ControlledDocument == nil {
		t.Fatalf("Create returned nil result for system_admin")
	}
}
