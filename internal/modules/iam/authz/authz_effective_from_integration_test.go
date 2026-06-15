//go:build integration
// +build integration

package authz_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

// seedMembership seeds a tenant, a user, a governed process area, and a single
// user_process_areas membership (role qms_admin, which the curated
// role_capabilities maps to document.submit) with the given effective_from.
// Returns the identity + area + capability needed to drive authz.Require.
func seedMembership(t *testing.T, db *sql.DB, effFrom time.Time) (tenantID, userID, area, capability string) {
	t.Helper()
	ctx := context.Background()

	tn := testdb.NewTenant(t, db)
	tenantID = tn.ID
	user := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	userID = user.ID
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	area = tax.ProcessAreaCode
	capability = "document.submit" // qms_admin holds this in curated role_capabilities

	// user_process_areas carries the membership.manage tripwire — assert it tx-locally.
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, 'qms_admin', $4)`,
			userID, tenantID, area, effFrom,
		)
		return err
	})
	return tenantID, userID, area, capability
}

// TestRequire_FutureDatedMembershipDenied is the RED test for F0.1: a membership
// whose effective_from is in the future must NOT grant access. Against code that
// ignores effective_from, Require wrongly returns nil (premature access, B1).
func TestRequire_FutureDatedMembershipDenied(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantID, userID, area, capability := seedMembership(t, db, time.Now().Add(time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.SeedTxIdentity(ctx, tx, tenantID, userID); err != nil {
		t.Fatalf("SeedTxIdentity: %v", err)
	}

	err = authz.Require(authz.WithCapCache(ctx), tx, capability, area)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("Require with future-dated membership = %v, want ErrCapDenied (premature access must be denied)", err)
	}
}

// TestRequire_CurrentMembershipGranted guards against over-correction: a current
// (effective_from in the past) membership must still grant.
func TestRequire_CurrentMembershipGranted(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenantID, userID, area, capability := seedMembership(t, db, time.Now().Add(-time.Hour))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.SeedTxIdentity(ctx, tx, tenantID, userID); err != nil {
		t.Fatalf("SeedTxIdentity: %v", err)
	}

	if err := authz.Require(authz.WithCapCache(ctx), tx, capability, area); err != nil {
		t.Fatalf("Require with current membership = %v, want grant (nil)", err)
	}
}
