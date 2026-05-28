package infrastructure

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

func TestFamilyRepository_Create_SetsAuthzGUCBeforeRequire(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewFamilyRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), "tenant-a"), "actor-1", nil)

	mock.ExpectBegin()
	expectTaxonomyManageAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO metaldocs.document_families (code, name, description, is_active)
VALUES ($1, $2, $3, $4)`)).
		WithArgs("procedure", "Procedimentos", "Familia global para fluxo de editor", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.Create(ctx, family("procedure"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProfileRepository_Create_SetsAuthzGUCBeforeRequire(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), "tenant-a"), "actor-1", nil)

	mock.ExpectBegin()
	expectTaxonomyManageAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO metaldocs.document_profiles
    (code, tenant_id, family_code, name, description, alias, review_interval_days, default_template_version_id, owner_user_id, editable_by_role, archived_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`)).
		WithArgs("pop", "tenant-a", "procedure", "Procedimento Operacional", "descricao", "pop", 365, nil, nil, "admin", nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.Create(ctx, profile("tenant-a"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAreaRepository_Create_SetsAuthzGUCBeforeRequire(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAreaRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), "tenant-a"), "actor-1", nil)

	mock.ExpectBegin()
	expectTaxonomyManageAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO metaldocs.document_process_areas
    (code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8)`)).
		WithArgs("general", "tenant-a", "Geral", "descricao", nil, nil, nil, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.Create(ctx, area("tenant-a"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFamilyRepository_GetByCodeForUpdate_RequiresAuthzBeforeLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewFamilyRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), "tenant-a"), "actor-1", nil)

	mock.ExpectBegin()
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	expectTaxonomyManageAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT code, name, description, is_active, created_at
FROM metaldocs.document_families
WHERE code = $1
FOR UPDATE`)).
		WithArgs("procedure").
		WillReturnRows(sqlmock.NewRows([]string{"code", "name", "description", "is_active", "created_at"}).
			AddRow("procedure", "Procedimentos", "Familia global para fluxo de editor", true, time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)))
	mock.ExpectRollback()

	if _, err := repo.GetByCodeForUpdate(ctx, tx, "procedure"); err != nil {
		t.Fatalf("GetByCodeForUpdate: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProfileRepository_GetByCodeForUpdate_RequiresAuthzBeforeLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), "tenant-a"), "actor-1", nil)

	mock.ExpectBegin()
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	expectTaxonomyManageAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT code, tenant_id, family_code, name, description, alias, review_interval_days,
       default_template_version_id, owner_user_id, editable_by_role, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1 AND code = $2
FOR UPDATE`)).
		WithArgs("tenant-a", "pop").
		WillReturnRows(sqlmock.NewRows([]string{"code", "tenant_id", "family_code", "name", "description", "alias", "review_interval_days", "default_template_version_id", "owner_user_id", "editable_by_role", "archived_at", "created_at"}).
			AddRow("pop", "tenant-a", "procedure", "Procedimento Operacional", "descricao", "pop", 365, nil, nil, "admin", nil, time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)))
	mock.ExpectRollback()

	if _, err := repo.GetByCodeForUpdate(ctx, tx, "tenant-a", "pop"); err != nil {
		t.Fatalf("GetByCodeForUpdate: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAreaRepository_GetByCodeForUpdate_RequiresAuthzBeforeLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAreaRepository(db)
	ctx := iamdomain.WithAuthContext(tenant.WithTenantID(context.Background(), "tenant-a"), "actor-1", nil)

	mock.ExpectBegin()
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	expectTaxonomyManageAuthz(mock, "actor-1", "tenant-a")
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at, created_at
FROM metaldocs.document_process_areas
WHERE tenant_id = $1 AND code = $2
FOR UPDATE`)).
		WithArgs("tenant-a", "general").
		WillReturnRows(sqlmock.NewRows([]string{"code", "tenant_id", "name", "description", "parent_code", "owner_user_id", "default_approver_role", "archived_at", "created_at"}).
			AddRow("general", "tenant-a", "Geral", "descricao", nil, nil, nil, nil, time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)))
	mock.ExpectRollback()

	if _, err := repo.GetByCodeForUpdate(ctx, tx, "tenant-a", "general"); err != nil {
		t.Fatalf("GetByCodeForUpdate: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func expectTaxonomyManageAuthz(mock sqlmock.Sqlmock, actorID, tenantID string) {
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
	expectSystemAdminBypass(mock, actorID, tenantID, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(nil))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectSystemAdminBypass(mock sqlmock.Sqlmock, actorID, tenantID string, exists bool) {
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id   = $1
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
)`)).
		WithArgs(actorID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
}

func profile(tenantID string) *domain.DocumentProfile {
	return &domain.DocumentProfile{
		Code:               "pop",
		TenantID:           tenantID,
		FamilyCode:         "procedure",
		Name:               "Procedimento Operacional",
		Description:        "descricao",
		Alias:              "pop",
		ReviewIntervalDays: 365,
		EditableByRole:     "admin",
	}
}

func area(tenantID string) *domain.ProcessArea {
	return &domain.ProcessArea{
		Code:        "general",
		TenantID:    tenantID,
		Name:        "Geral",
		Description: "descricao",
	}
}
