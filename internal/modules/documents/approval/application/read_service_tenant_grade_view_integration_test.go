//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	approvalrepo "metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// F0.3 (superseded by F8, spec.md §6.3) — CapDocumentView is declared
// ScopeTenant (capability_scope.go:51). LoadInstance/LoadActiveInstanceByDocument
// intentionally pass authz.Require the "tenant" sentinel for CapDocumentView so
// tier-2 authz does not narrow the tenant-grade view capability to area-grade.
//
// F0.3 originally stopped there and let a bare CapDocumentView holder load ANY
// in-flight instance. F8 (spec.md §6.3) tightens this: passing the tier-2 gate
// is necessary but not sufficient — the query/repository layer additionally
// enforces requireInstanceVisible (read_service.go), restricting visibility to
// {submitting author, current-or-past stage pool member (eligible_actor_ids),
// CapApprovalOversee holder (tenant-wide), CapDocumentEdit holder (checked
// against the INSTANCE'S OWN resolved area, never a tenant-wide skip)}. A bare
// document.view holder outside that set is denied with the DISTINCT
// infrastructure.ErrInstanceNotVisible (404, "cross-boundary = not-found"),
// not granted and not authz.ErrCapDenied (403) — see errors.go / http/errors.go.
//
// This file's original "Granted" tests (a bare tenant-grade viewer sees any
// instance) are rewritten below to prove the corrected, narrower contract:
// bare viewer DENIED (ErrInstanceNotVisible, not ErrCapDenied); pool member,
// oversee holder, and area-correct edit holder all GRANTED; no-grant-at-all
// actor still denied with the pre-existing ErrCapDenied (tier-2 gate, unchanged);
// system_admin still bypasses (unchanged).

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
		// document.view must be evaluated regardless of where in the tenant the
		// role is held (the tier-2 gate itself is unaffected by F8).
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

// grantRoleInArea grants roleCode to userID scoped to areaCode (the document's
// OWN area, unlike seedTenantGradeViewFixture's otherArea helper above) — used
// by the CapDocumentEdit-area-correctness tests, where the grant must line up
// with the instance's resolved area for authz.Require to pass.
func grantRoleInArea(t *testing.T, db *sql.DB, tenantID, userID, areaCode, roleCode string) {
	t.Helper()
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.user_process_areas
			    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
			userID, tenantID, areaCode, roleCode,
		)
		return err
	})
}

func newReadServiceForIntegration(t *testing.T, db *sql.DB) *ReadService {
	t.Helper()
	repo := approvalrepo.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	return newReadService(repo, testdb.NewCDFieldReader(t))
}

func ctxWithIdentity(tenantID, actorID string) context.Context {
	ctx := tenant.WithTenantID(context.Background(), tenantID)
	return iamdomain.WithAuthContext(ctx, actorID, nil)
}

// --- Denied (F8): bare tenant-grade viewer, not author/pool/oversee/edit ------

func TestLoadInstance_BareTenantGradeViewer_NotVisible_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	// "viewer" role only has document.view — passes the tier-2 gate but has
	// none of the F8 visibility-set memberships (not author, no stages to be a
	// pool member of, no oversee/edit grant).
	tenantID, actorID, _, instanceID := seedTenantGradeViewFixture(t, db, "viewer")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	_, err := svc.LoadInstance(ctxWithIdentity(tenantID, actorID), runner, tenantID, instanceID)
	if !errors.Is(err, infrastructure.ErrInstanceNotVisible) {
		t.Fatalf("LoadInstance for bare tenant-grade viewer = %v, want infrastructure.ErrInstanceNotVisible (F8 spec.md §6.3: cross-boundary = not-found)", err)
	}
}

func TestLoadActiveInstanceByDocument_BareTenantGradeViewer_NotVisible_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, actorID, docID, _ := seedTenantGradeViewFixture(t, db, "viewer")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	_, err := svc.LoadActiveInstanceByDocument(ctxWithIdentity(tenantID, actorID), runner, tenantID, docID)
	if !errors.Is(err, infrastructure.ErrInstanceNotVisible) {
		t.Fatalf("LoadActiveInstanceByDocument for bare tenant-grade viewer = %v, want infrastructure.ErrInstanceNotVisible (F8 spec.md §6.3)", err)
	}
}

// --- Granted (F8): submitting author is always visible -----------------------

func TestLoadInstance_Author_Visible_Granted(t *testing.T) {
	db, _ := testdb.Open(t)
	ten := testdb.NewTenant(t, db)
	author := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(ten.ID), testdb.WithOwner(author.ID), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(doc), testdb.WithOwner(author.ID), testdb.WithStatus("in_progress"),
	)
	// Author still needs to clear the tier-2 CapDocumentView/CapApprovalOversee
	// gate first; grant plain "viewer" so the ONLY reason LoadInstance succeeds
	// is the F8 author-visibility branch, not an incidental oversee/edit grant.
	grantRoleInArea(t, db, ten.ID, author.ID, testdb.NewTaxonomy(t, db, testdb.WithTenant(ten.ID)).ProcessAreaCode, "viewer")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	got, err := svc.LoadInstance(ctxWithIdentity(ten.ID, author.ID), runner, ten.ID, inst.ID)
	if err != nil {
		t.Fatalf("LoadInstance for submitting author = %v, want nil", err)
	}
	if got == nil || got.ID != inst.ID {
		t.Fatalf("LoadInstance returned instance = %#v, want id %q", got, inst.ID)
	}
}

// --- Granted (F8): CapApprovalOversee holder (tenant-wide) --------------------

func TestLoadInstance_OverseeHolder_Visible_Granted(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, _, _, instanceID := seedTenantGradeViewFixture(t, db, "")
	overseer := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	// qms_admin -> approval.oversee (db/reference-data/0001_product_reference_data.sql).
	// Grant in a distinct area than the document's: oversee is tenant-wide.
	otherArea := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID)).ProcessAreaCode
	grantRoleInArea(t, db, tenantID, overseer.ID, otherArea, "qms_admin")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	got, err := svc.LoadInstance(ctxWithIdentity(tenantID, overseer.ID), runner, tenantID, instanceID)
	if err != nil {
		t.Fatalf("LoadInstance for CapApprovalOversee holder = %v, want nil", err)
	}
	if got == nil || got.ID != instanceID {
		t.Fatalf("LoadInstance returned instance = %#v, want id %q", got, instanceID)
	}
}

// --- Granted (F8): CapDocumentEdit holder, area-CORRECT (own instance area) --

func TestLoadInstance_EditHolder_OwnArea_Visible_Granted(t *testing.T) {
	db, _ := testdb.Open(t)
	ten := testdb.NewTenant(t, db)
	area := testdb.NewTaxonomy(t, db, testdb.WithTenant(ten.ID))
	// NewDocument does not forward WithTaxonomy to its auto-built
	// ControlledDoc, so the CD (which carries process_area_code, the
	// loadInstanceAreaCode fallback source when the doc/stage snapshots are
	// NULL) must be built explicitly here and wired via WithControlledDoc.
	cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID), testdb.WithTaxonomy(area))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(ten.ID), testdb.WithStatus("approved"), testdb.WithControlledDoc(cd))
	inst := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("in_progress"))

	editor := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
	// "editor" role -> document.edit (reference data). Grant in the SAME area
	// the instance resolves to (loadInstanceAreaCode) — proves the area-scoped
	// check actually threads the instance's own area, not a "tenant" skip.
	grantRoleInArea(t, db, ten.ID, editor.ID, area.ProcessAreaCode, "editor")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	got, err := svc.LoadInstance(ctxWithIdentity(ten.ID, editor.ID), runner, ten.ID, inst.ID)
	if err != nil {
		t.Fatalf("LoadInstance for same-area CapDocumentEdit holder = %v, want nil", err)
	}
	if got == nil || got.ID != inst.ID {
		t.Fatalf("LoadInstance returned instance = %#v, want id %q", got, inst.ID)
	}
}

// --- Denied (F8): CapDocumentEdit holder, WRONG area --------------------------
//
// This is the regression guard for the exact bug this fix closes: an
// area-scoped CapDocumentEdit grant in a DIFFERENT area than the instance must
// NOT satisfy visibility. Before the fix, requireInstanceVisible passed the
// "tenant" sentinel for CapDocumentEdit, which incorrectly granted edit-anywhere
// semantics regardless of area.
func TestLoadInstance_EditHolder_WrongArea_NotVisible_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	ten := testdb.NewTenant(t, db)
	docArea := testdb.NewTaxonomy(t, db, testdb.WithTenant(ten.ID))
	otherArea := testdb.NewTaxonomy(t, db, testdb.WithTenant(ten.ID))
	cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID), testdb.WithTaxonomy(docArea))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(ten.ID), testdb.WithStatus("approved"), testdb.WithControlledDoc(cd))
	inst := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("in_progress"))

	editor := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
	grantRoleInArea(t, db, ten.ID, editor.ID, otherArea.ProcessAreaCode, "editor")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	_, err := svc.LoadInstance(ctxWithIdentity(ten.ID, editor.ID), runner, ten.ID, inst.ID)
	if !errors.Is(err, infrastructure.ErrInstanceNotVisible) {
		t.Fatalf("LoadInstance for wrong-area CapDocumentEdit holder = %v, want infrastructure.ErrInstanceNotVisible (area-scoped, must not grant edit-anywhere)", err)
	}
}

// --- Denied: no document.view grant anywhere (tier-2 gate, unchanged by F8) --

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
