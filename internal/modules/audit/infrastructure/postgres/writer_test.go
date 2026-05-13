package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/audit/domain"
)

func TestWriterRecordTxStoresHashChainColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("SELECT pg_advisory_xact_lock($1)")).
		WithArgs(auditHashChainLockID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO metaldocs.audit_events").
		WithArgs(
			"evt-1",
			sqlmock.AnyArg(),
			"actor-1",
			"document.rename",
			"document",
			"doc-1",
			`{"name":"QMS"}`,
			"trace-1",
			"tenant-1",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	writer := NewWriter(db)
	err = writer.RecordTx(context.Background(), tx, domain.Event{
		ID:           "evt-1",
		OccurredAt:   time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC),
		ActorID:      "actor-1",
		Action:       "document.rename",
		ResourceType: "document",
		ResourceID:   "doc-1",
		PayloadJSON:  `{"name":"QMS"}`,
		TraceID:      "trace-1",
		TenantID:     "tenant-1",
	})
	if err != nil {
		t.Fatalf("RecordTx: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestWriterValidateIntegrityReportsBrokenChain(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"audit_sequence",
		"id",
		"prev_hash",
		"row_hash",
		"expected_prev_hash",
		"expected_row_hash",
	}).
		AddRow(int64(1), "evt-1", "", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").
		AddRow(int64(2), "evt-2", "wrong", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")

	mock.ExpectQuery("FROM metaldocs.audit_events").
		WillReturnRows(rows)

	writer := NewWriter(db)
	issues, err := writer.ValidateIntegrity(context.Background())
	if err != nil {
		t.Fatalf("ValidateIntegrity: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("issue count = %d, want 2: %#v", len(issues), issues)
	}
	if issues[0].Kind != domain.IntegrityIssuePrevHashMismatch {
		t.Fatalf("first issue kind = %q, want prev hash mismatch", issues[0].Kind)
	}
	if issues[1].Kind != domain.IntegrityIssueRowHashMismatch {
		t.Fatalf("second issue kind = %q, want row hash mismatch", issues[1].Kind)
	}
}

func TestWriterValidateIntegrityAllowsRetainedFirstRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	rows := sqlmock.NewRows([]string{
		"audit_sequence",
		"id",
		"prev_hash",
		"row_hash",
		"expected_prev_hash",
		"expected_row_hash",
	}).AddRow(int64(10), "evt-10", hash, hash, hash, hash)

	mock.ExpectQuery(`ROW_NUMBER\(\) OVER`).
		WillReturnRows(rows)

	writer := NewWriter(db)
	issues, err := writer.ValidateIntegrity(context.Background())
	if err != nil {
		t.Fatalf("ValidateIntegrity: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("issue count = %d, want 0: %#v", len(issues), issues)
	}
}
