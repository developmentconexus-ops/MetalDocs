package infrastructure

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

func TestFamilyRepository_HasActiveProfiles_TenantScoped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewFamilyRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), "tenant-a"), "actor-1", nil)

	mock.ExpectBegin()
	expectDocumentViewAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1 AND family_code = $2 AND archived_at IS NULL
)`)).
		WithArgs("tenant-a", "fam-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	exists, err := repo.HasActiveProfiles(ctx, "tenant-a", "fam-1")
	if err != nil {
		t.Fatalf("HasActiveProfiles: %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func expectDocumentViewAuthz(mock sqlmock.Sqlmock, actorID, tenantID string) {
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.tenant_id', $1, true)")).
		WithArgs(tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.actor_id', $1, true)")).
		WithArgs(actorID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(actorID))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(tenantID))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles
   WHERE user_id   = $1
     AND tenant_id = $2::uuid
     AND role_code = 'system_admin'
)`)).
		WithArgs(actorID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(nil))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func family(code string) *domain.DocumentFamily {
	return &domain.DocumentFamily{
		Code:        code,
		Name:        "Procedimentos",
		Description: "Familia global para fluxo de editor",
		IsActive:    true,
	}
}
