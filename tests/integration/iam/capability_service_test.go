//go:build integration

package iam_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"metaldocs/internal/modules/iam/application"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

const devTenant = tenant.DevTenantID

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", testdb.DSN(t))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(context.Background()); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}
	return db
}

func TestCapabilityService_SystemAdmin_AllowsAnyCapability(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	userID := testdb.DeterministicID(t, "system-admin")

	insertUserRole(t, db, userID, "system_admin")
	svc := application.NewCapabilityService(db)

	for _, cap := range []string{
		"doc.view", "doc.create", "doc.submit", "doc.signoff",
		"user.manage", "taxonomy.manage", "membership.manage",
	} {
		if err := svc.CanDo(ctx, userID, devTenant, cap); err != nil {
			t.Fatalf("CanDo(%q) = %v, want nil", cap, err)
		}
	}
}

func TestCapabilityService_Viewer_AllowedDocView_DeniedDocCreate(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	userID := testdb.DeterministicID(t, "viewer")

	insertUserRole(t, db, userID, "viewer")
	svc := application.NewCapabilityService(db)

	if err := svc.CanDo(ctx, userID, devTenant, "doc.view"); err != nil {
		t.Fatalf("CanDo(doc.view) = %v, want nil", err)
	}
	if err := svc.CanDo(ctx, userID, devTenant, "doc.create"); !errors.Is(err, application.ErrCapabilityDenied) {
		t.Fatalf("CanDo(doc.create) = %v, want ErrCapabilityDenied", err)
	}
}

func TestCapabilityService_UnknownUser_Denied(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	userID := testdb.DeterministicID(t, "unknown")

	svc := application.NewCapabilityService(db)
	if err := svc.CanDo(ctx, userID, devTenant, "doc.view"); !errors.Is(err, application.ErrCapabilityDenied) {
		t.Fatalf("CanDo(doc.view) = %v, want ErrCapabilityDenied", err)
	}
}

func insertUserRole(t *testing.T, db *sql.DB, userID, role string) {
	t.Helper()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
		 VALUES ($1, $2, $3::uuid)
		 ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id`,
		userID, userID, devTenant,
	); err != nil {
		t.Fatalf("insert iam_users: %v", err)
	}

	// uq_iam_user_roles_user_tenant: UNIQUE(tenant_id, user_id) — one role per user
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id, assigned_by)
		 VALUES ($1, $2, $3::uuid, $1)
		 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code, assigned_by = EXCLUDED.assigned_by`,
		userID, role, devTenant,
	); err != nil {
		t.Fatalf("insert iam_user_roles: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID)   //nolint:errcheck
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)        //nolint:errcheck
	})
}
