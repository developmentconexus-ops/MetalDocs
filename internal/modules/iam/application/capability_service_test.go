package application

import (
	"context"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestCapabilityService_CanDo_GroupPathEnforcesTenantScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND gr.role = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_user_roles ur
    JOIN metaldocs.role_capabilities rc ON rc.role = ur.role_code
   WHERE ur.user_id = $1
     AND ur.tenant_id = $2::uuid
     AND rc.capability = $3
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
    JOIN metaldocs.role_capabilities rc ON rc.role = gr.role
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND rc.capability = $3
)`)).
		WithArgs("user-1", "tenant-1", "membership.manage").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	svc := NewCapabilityService(db)
	if err := svc.CanDo(context.Background(), "user-1", "tenant-1", "membership.manage"); err != nil {
		t.Fatalf("CanDo: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCapabilityService_CanDo_InvalidCapabilityDenied(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewCapabilityService(db)
	err = svc.CanDo(context.Background(), "user-1", "tenant-1", "invalid.cap")
	if !errors.Is(err, ErrCapabilityDenied) {
		t.Fatalf("CanDo invalid capability error = %v, want ErrCapabilityDenied", err)
	}
}
