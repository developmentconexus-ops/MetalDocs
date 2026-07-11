//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	controlleddocumentsinfrastructure "metaldocs/internal/modules/controlleddocuments/infrastructure"
	docrepo "metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// ctxForActor builds a context carrying BOTH identities GetActiveInstance
// needs: the iamdomain auth-context key that authn.UserIDFromContext reads
// (service.go:677), and the platform/tenant tenant+actor keys the TxRunner
// chokepoint auto-seeds into the metaldocs.tenant_id/actor_id GUCs for the
// in-tx authz.Require(document.view) call (service.go:690-701). Neither key
// pair alone is sufficient — iamdomain.WithAuthContext alone leaves the GUCs
// unset (authz.Require fails closed); authzCtx alone leaves UserIDFromContext
// unsatisfied (ErrActorMissing before authz.Require is ever reached).
func ctxForActor(tenantID, actorID string) context.Context {
	return iamdomain.WithAuthContext(authzCtx(tenantID, actorID), actorID, nil)
}

// SEC-03 (T-006) — GetActiveDocument had no tier-2 authz.Require/capability
// check: only the visibility_scope EXISTS check (CanRead) ran, so any
// authenticated tenant member (default visibility_scope='company') could read
// content hashes, approval state, and published-revision IDs. The fix adds
// authz.Require(document.view, "tenant") inside s.runner.Do, mirroring the
// canonical tenant-grade read-policy pattern in
// documents/approval/application/read_service.go LoadInstance and
// documents/application/view_service.go GetViewURL (ADR 0022: capabilities,
// never roles; document.view is ScopeTenant so callers pass the "tenant"
// sentinel to intentionally disable area filtering).
//
// These tests seed a CD with the DB-default visibility_scope='company' (so
// CanRead is trivially true for any tenant member) to isolate the assertion:
// denial must come from the missing document.view capability grant, not from
// visibility scoping.

func newActiveInstanceServiceForIntegration(t *testing.T, db *sql.DB) *ControlledDocumentService {
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

func TestGetActiveInstance_NoViewGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenant.ID))

	svc := newActiveInstanceServiceForIntegration(t, db)
	ctx := ctxForActor(tenant.ID, user.ID)

	_, err := svc.GetActiveInstance(ctx, tenant.ID, cd.ID)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("GetActiveInstance for actor with no document.view grant = %v, want authz.ErrCapDenied", err)
	}
	if denied.AreaCode != "tenant" {
		t.Fatalf("ErrCapDenied.AreaCode = %q, want %q (tenant-grade deny envelope)", denied.AreaCode, "tenant")
	}
}

func TestGetActiveInstance_WithViewGrant_PassesAuthz(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	user := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenant.ID))

	// Grant "viewer" (carries document.view per db/reference-data role_capabilities)
	// in an unrelated area — document.view is tenant-grade, so the grant must
	// pass regardless of the CD's own process area.
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.user_process_areas
			    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
			user.ID, tenant.ID, tax.ProcessAreaCode, "viewer",
		)
		return err
	})

	svc := newActiveInstanceServiceForIntegration(t, db)
	ctx := ctxForActor(tenant.ID, user.ID)

	// No active document instance exists for this fresh CD, so the call is
	// expected to fall through to ErrNoActiveInstance — the point of this test
	// is that it is NOT authz.ErrCapDenied, proving the capability check passed.
	_, err := svc.GetActiveInstance(ctx, tenant.ID, cd.ID)
	var denied authz.ErrCapDenied
	if errors.As(err, &denied) {
		t.Fatalf("GetActiveInstance for actor WITH document.view grant = %v, want no authz.ErrCapDenied", err)
	}
}

func TestGetActiveInstance_SystemAdmin_PassesAuthz(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenant.ID))
	testdb.SeedSystemAdmin(t, db, tenant.ID, user.ID, "SEC-03 System Admin")

	svc := newActiveInstanceServiceForIntegration(t, db)
	ctx := ctxForActor(tenant.ID, user.ID)

	_, err := svc.GetActiveInstance(ctx, tenant.ID, cd.ID)
	var denied authz.ErrCapDenied
	if errors.As(err, &denied) {
		t.Fatalf("GetActiveInstance for system_admin = %v, want no authz.ErrCapDenied", err)
	}
}
