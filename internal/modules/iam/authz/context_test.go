package authz_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/authz"
)

func TestMustActorID_ReturnsErrWhenGUCMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))

	_, err = authz.MustActorID(context.Background(), tx)
	if !errors.Is(err, authz.ErrActorContextMissing) {
		t.Fatalf("expected ErrActorContextMissing, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMustActorID_ReturnsValueWhenSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("user-123"))

	got, err := authz.MustActorID(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "user-123" {
		t.Fatalf("got %q, want user-123", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMustTenantID_ReturnsErrWhenGUCMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))

	_, err = authz.MustTenantID(context.Background(), tx)
	if !errors.Is(err, authz.ErrTenantContextMissing) {
		t.Fatalf("expected ErrTenantContextMissing, got %v", err)
	}
}

// TestSeedTxIdentity_UsesTransactionLocalGUCs locks the PgBouncer pooling-leak
// guard from ADR 0022 amendment (R6 / authz-industry-evidence.md §3.1): the
// actor/tenant identity GUCs MUST be seeded transaction-local — set_config(..., true)
// — never session-level, or a pooled connection leaks one client's identity to
// the next. The regex asserts BOTH set_config calls carry the `, true)`
// transaction-local flag. A regression to `, false)` (or a missing flag) fails here.
func TestSeedTxIdentity_UsesTransactionLocalGUCs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	// Single combined statement (context.go): SELECT set_config(tenant,$1,true), set_config(actor,$2,true).
	// Pattern requires both GUCs AND the transaction-local `true` 3rd arg for each.
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id', \$1, true\)[\s\S]*set_config\('metaldocs\.actor_id', \$2, true\)`).
		WithArgs("tenant-9", "user-9").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := authz.SeedTxIdentity(context.Background(), tx, "tenant-9", "user-9"); err != nil {
		t.Fatalf("SeedTxIdentity: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
