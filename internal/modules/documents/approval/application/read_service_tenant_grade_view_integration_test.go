//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// F0.3 — CapDocumentView is declared ScopeTenant (capability_scope.go:51).
// Two read paths (LoadInstance, LoadActiveInstanceByDocument) historically
// resolved a real area code and passed it to authz.Require, narrowing the
// tenant-grade view to area-grade. The fix aligns both call sites with the
// canonical sibling documents/application/view_service.go:71 by passing the
// "tenant" sentinel. These tests prove the consumer contract: a tenant-role-only
// viewer reads a document carrying any real area code; an actor with no view
// grant is denied with ErrCapDenied{AreaCode:"tenant"}; system_admin bypasses.

func seedTenantGradeViewFixture(t *testing.T, db *sql.DB, role string) (tenantID, actorID, docID, instanceID string) {
	t.Helper()
	ten := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(ten.ID), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(doc),
		testdb.WithStatus("in_progress"),
	)

	if role != "" {
		// Grant the role in a DIFFERENT area than the document's: tenant-grade
		// document.view must grant regardless of where in the tenant the role
		// is held.
		otherArea := testdb.NewTaxonomy(t, db, testdb.WithTenant(ten.ID)).ProcessAreaCode
		testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(),
				`INSERT INTO metaldocs.user_process_areas
				    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
				 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
				user.ID, ten.ID, otherArea, role,
			)
			return err
		})
	}

	return ten.ID, user.ID, inst.DocumentID, inst.ID
}

func newReadServiceForIntegration(t *testing.T, db *sql.DB) *ReadService {
	t.Helper()
	repo := approvalrepo.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	return newReadService(repo, cdinfra.NewCDFieldReaderPG())
}

func ctxWithIdentity(tenantID, actorID string) context.Context {
	ctx := tenant.WithTenantID(context.Background(), tenantID)
	return iamdomain.WithAuthContext(ctx, actorID, nil)
}

// --- Granted: tenant-grade viewer, doc carries an area code -------------------

func TestLoadInstance_TenantGradeViewer_DocWithAreaCode_Granted(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, _, instanceID := seedTenantGradeViewFixture(t, db, "viewer")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	inst, err := svc.LoadInstance(ctxWithIdentity(tenantID, actorID), runner, tenantID, instanceID)
	if err != nil {
		t.Fatalf("LoadInstance for tenant-grade viewer = %v, want nil (tenant-grade document.view must grant across areas)", err)
	}
	if inst == nil || inst.ID != instanceID {
		t.Fatalf("LoadInstance returned instance = %#v, want id %q", inst, instanceID)
	}
}

func TestLoadActiveInstanceByDocument_TenantGradeViewer_DocWithAreaCode_Granted(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, docID, instanceID := seedTenantGradeViewFixture(t, db, "viewer")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	inst, err := svc.LoadActiveInstanceByDocument(ctxWithIdentity(tenantID, actorID), runner, tenantID, docID)
	if err != nil {
		t.Fatalf("LoadActiveInstanceByDocument for tenant-grade viewer = %v, want nil", err)
	}
	if inst == nil || inst.ID != instanceID {
		t.Fatalf("LoadActiveInstanceByDocument returned instance = %#v, want id %q", inst, instanceID)
	}
}

// --- Denied: no document.view grant anywhere ----------------------------------

func TestLoadInstance_NoViewGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, _, instanceID := seedTenantGradeViewFixture(t, db, "")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	_, err := svc.LoadInstance(ctxWithIdentity(tenantID, actorID), runner, tenantID, instanceID)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("LoadInstance for no-grant actor = %v, want authz.ErrCapDenied", err)
	}
	if denied.AreaCode != "tenant" {
		t.Fatalf("ErrCapDenied.AreaCode = %q, want %q (tenant-grade deny envelope)", denied.AreaCode, "tenant")
	}
}

func TestLoadActiveInstanceByDocument_NoViewGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, docID, _ := seedTenantGradeViewFixture(t, db, "")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	_, err := svc.LoadActiveInstanceByDocument(ctxWithIdentity(tenantID, actorID), runner, tenantID, docID)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("LoadActiveInstanceByDocument for no-grant actor = %v, want authz.ErrCapDenied", err)
	}
	if denied.AreaCode != "tenant" {
		t.Fatalf("ErrCapDenied.AreaCode = %q, want %q (tenant-grade deny envelope)", denied.AreaCode, "tenant")
	}
}

// --- system_admin bypass still works ------------------------------------------

func TestLoadInstance_SystemAdmin_Granted(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, _, instanceID := seedTenantGradeViewFixture(t, db, "")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "F0.3 System Admin LoadInstance")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	inst, err := svc.LoadInstance(ctxWithIdentity(tenantID, actorID), runner, tenantID, instanceID)
	if err != nil {
		t.Fatalf("LoadInstance for system_admin = %v, want nil", err)
	}
	if inst == nil || inst.ID != instanceID {
		t.Fatalf("LoadInstance returned instance = %#v, want id %q", inst, instanceID)
	}
}

func TestLoadActiveInstanceByDocument_SystemAdmin_Granted(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, docID, instanceID := seedTenantGradeViewFixture(t, db, "")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "F0.3 System Admin LoadActiveByDocument")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	inst, err := svc.LoadActiveInstanceByDocument(ctxWithIdentity(tenantID, actorID), runner, tenantID, docID)
	if err != nil {
		t.Fatalf("LoadActiveInstanceByDocument for system_admin = %v, want nil", err)
	}
	if inst == nil || inst.ID != instanceID {
		t.Fatalf("LoadActiveInstanceByDocument returned instance = %#v, want id %q", inst, instanceID)
	}
}
