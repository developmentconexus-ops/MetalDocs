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

func auditListRows(n int) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"id", "occurred_at", "actor_id", "action",
		"resource_type", "resource_id", "payload", "trace_id", "tenant_id",
	})
	now := time.Now().UTC()
	for i := 0; i < n; i++ {
		rows.AddRow("evt", now, "actor", "audit.test", "document", "doc", `{}`, "trace", "tenant-1")
	}
	return rows
}

// TestWriterListEventsExactPageHasNoMore is the B1 regression guard: an
// exact-multiple last page (rows == limit) must report hasMore=false. The old
// handler-side `len(items) >= limit` heuristic reported a false next page here.
func TestWriterListEventsExactPageHasNoMore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// limit=3 → reader probes for limit+1 (4); only 3 rows exist → no probe row.
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs("tenant-1", 4).
		WillReturnRows(auditListRows(3))

	writer := NewWriter(db)
	items, hasMore, err := writer.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-1", Limit: 3})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if hasMore {
		t.Fatalf("hasMore = true on exact-multiple page, want false")
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestWriterListEventsTrimsProbeRow is the B1 regression guard for the
// next-page case: when the limit+1 probe row exists, hasMore=true and the probe
// is trimmed so the caller sees exactly limit rows.
func TestWriterListEventsTrimsProbeRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// limit=3 → probe for 4; 4 rows returned → trimmed to 3 + hasMore=true.
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs("tenant-1", 4).
		WillReturnRows(auditListRows(4))

	writer := NewWriter(db)
	items, hasMore, err := writer.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-1", Limit: 3})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if !hasMore {
		t.Fatalf("hasMore = false with probe row present, want true")
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3 (probe row trimmed)", len(items))
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
		WithArgs(auditIntegrityValidationWindow).
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
		WithArgs(auditIntegrityValidationWindow).
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

func TestWriterValidateIntegrityStopsCollectingAfterIssueLimit(t *testing.T) {
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
	})
	for i := 0; i < auditIntegrityIssueLimit+5; i++ {
		rows.AddRow(int64(i+1), "evt", "wrong", "bad", "expected", "expected")
	}

	mock.ExpectQuery("FROM recent").
		WithArgs(auditIntegrityValidationWindow).
		WillReturnRows(rows)

	writer := NewWriter(db)
	issues, err := writer.ValidateIntegrity(context.Background())
	if err != nil {
		t.Fatalf("ValidateIntegrity: %v", err)
	}
	if len(issues) != auditIntegrityIssueLimit {
		t.Fatalf("issue count = %d, want %d", len(issues), auditIntegrityIssueLimit)
	}
}
