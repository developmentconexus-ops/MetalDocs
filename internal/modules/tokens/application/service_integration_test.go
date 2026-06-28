//go:build integration
// +build integration

package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	auditpostgres "metaldocs/internal/modules/audit/infrastructure/postgres"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/tokens/application"
	"metaldocs/internal/modules/tokens/domain"
	tokensinfra "metaldocs/internal/modules/tokens/infrastructure"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// buildSvc constructs the real production service wired to the provided
// isolated test database. All dependencies are concrete — no stubs.
func buildSvc(db *sql.DB) *application.Service {
	runner := platformdb.NewTxRunner(db)
	repo := tokensinfra.NewPostgresRepository()
	aw := auditpostgres.NewWriter(db)
	return application.NewService(runner, repo, aw)
}

// grantSP1Caps seeds a user_process_areas membership for actor in tenant with
// role=qms_admin. The curated bootstrap loads role_capabilities rows that map
// qms_admin → token.view and qms_admin → token_dictionary.manage (seeded via
// db/reference-data in Task 3). The service passes areaCode="tenant" to authz.Require, which
// degenerates to a tenant-scope check (skips area filter) so any valid area_code
// row satisfies the query.
//
// user_process_areas carries the membership.manage tripwire
// (trg_require_cap_asserted), so we assert it tx-locally via SeedWithCaps —
// pool-safe, never leaks session state.
func grantSP1Caps(t *testing.T, db *sql.DB, tenantID, userID string) {
	t.Helper()
	// Seed a governed process-area row so the area_code FK parent exists.
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	areaCode := tax.ProcessAreaCode

	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, 'qms_admin', $4)`,
			userID, tenantID, areaCode, time.Now().Add(-time.Hour),
		)
		return err
	})
}

// ctxFor returns a context carrying the actor user ID for read paths (Get, List,
// GetByName). Those methods call iamdomain.UserIDFromContext(ctx) to obtain the
// actor for SeedTxIdentity.
func ctxFor(userID string) context.Context {
	return iamdomain.WithAuthContext(context.Background(), userID, nil)
}

// TestCreate_GrantedManage_Succeeds verifies the happy path: a qms_admin user
// (token_dictionary.manage) can create a dictionary entry and receives back a
// populated Entry with a server-assigned ID and correct Name.
func TestCreate_GrantedManage_Succeeds(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	svc := buildSvc(db)
	entry, err := svc.Create(ctxFor(user.ID), application.CreateCommand{
		TenantID: tn.ID,
		ActorID:  user.ID,
		Name:     "company_slogan",
		Value:    "Excellence in Quality",
		Label:    "Slogan da empresa",
	})
	if err != nil {
		t.Fatalf("Create = %v, want nil", err)
	}
	if entry.ID == "" {
		t.Fatal("Create returned entry with empty ID")
	}
	if entry.Name != "company_slogan" {
		t.Fatalf("Create entry.Name = %q, want %q", entry.Name, "company_slogan")
	}
}

// TestCreate_Unauthorized_Denied verifies that a bare user (no process-area
// membership) receives authz.ErrCapDenied when attempting to create an entry.
func TestCreate_Unauthorized_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	// bare user: iam_users row exists but no user_process_areas membership.
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))

	svc := buildSvc(db)
	_, err := svc.Create(ctxFor(user.ID), application.CreateCommand{
		TenantID: tn.ID,
		ActorID:  user.ID,
		Name:     "unauthorised_entry",
		Value:    "v",
		Label:    "l",
	})
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("Create without grants = %v (%T), want authz.ErrCapDenied", err, err)
	}
}

// TestGet_CrossTenant_NotFound verifies that tenant B cannot read an entry
// created by tenant A. The service seeds the GUC metaldocs.tenant_id for tenant
// B's request; the RLS policy on token_dictionary_entries filters by tenant_id,
// so tenant A's row is invisible and the service returns ErrNotFound.
func TestGet_CrossTenant_NotFound(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	// Tenant A creates an entry.
	tnA := testdb.NewTenant(t, db)
	userA := testdb.NewUser(t, db, testdb.WithTenant(tnA.ID))
	grantSP1Caps(t, db, tnA.ID, userA.ID)

	svc := buildSvc(db)
	entry, err := svc.Create(ctxFor(userA.ID), application.CreateCommand{
		TenantID: tnA.ID,
		ActorID:  userA.ID,
		Name:     "cross_tenant_entry",
		Value:    "v",
		Label:    "l",
	})
	if err != nil {
		t.Fatalf("Create (tenant A) = %v, want nil", err)
	}

	// Tenant B: own user with own caps.
	tnB := testdb.NewTenant(t, db)
	userB := testdb.NewUser(t, db, testdb.WithTenant(tnB.ID))
	grantSP1Caps(t, db, tnB.ID, userB.ID)

	// Tenant B tries to read tenant A's entry using tenant B's tenant_id scope.
	_, err = svc.Get(ctxFor(userB.ID), tnB.ID, entry.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Get cross-tenant = %v (%T), want domain.ErrNotFound", err, err)
	}
}

// TestCreate_DuplicateName_Conflict verifies that creating two entries with the
// same (tenant_id, name) pair returns a PostgreSQL unique-violation error (23505).
// The uq_token_dictionary_tenant_name index enforces this at the DB level.
func TestCreate_DuplicateName_Conflict(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	svc := buildSvc(db)
	cmd := application.CreateCommand{
		TenantID: tn.ID,
		ActorID:  user.ID,
		Name:     "duplicate_name",
		Value:    "v",
		Label:    "l",
	}

	if _, err := svc.Create(ctxFor(user.ID), cmd); err != nil {
		t.Fatalf("Create (first) = %v, want nil", err)
	}

	_, err := svc.Create(ctxFor(user.ID), cmd)
	if err == nil {
		t.Fatal("Create (duplicate) = nil, want unique-violation error")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("Create (duplicate) = %v (%T), want *pgconn.PgError", err, err)
	}
	const uniqueViolation = "23505"
	if pgErr.Code != uniqueViolation {
		t.Fatalf("PgError.Code = %q, want %q (unique_violation)", pgErr.Code, uniqueViolation)
	}
}

// TestCreate_RecordsAuditEvent verifies that a successful Create writes a row to
// metaldocs.audit_events with action="tokens.entry.created" for the correct
// resource_id and tenant.
func TestCreate_RecordsAuditEvent(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	svc := buildSvc(db)
	entry, err := svc.Create(ctxFor(user.ID), application.CreateCommand{
		TenantID: tn.ID,
		ActorID:  user.ID,
		Name:     "audit_target",
		Value:    "v",
		Label:    "l",
	})
	if err != nil {
		t.Fatalf("Create = %v, want nil", err)
	}

	var count int
	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM metaldocs.audit_events
		  WHERE action = 'tokens.entry.created'
		    AND resource_id = $1
		    AND tenant_id = $2`,
		entry.ID, tn.ID,
	).Scan(&count); err != nil {
		t.Fatalf("query audit_events: %v", err)
	}
	if count != 1 {
		t.Fatalf("audit_events count for tokens.entry.created = %d, want 1", count)
	}
}
