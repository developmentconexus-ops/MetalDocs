package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/audit/domain"
	platcrypto "metaldocs/internal/platform/crypto"
)

func TestWriterRecordTxStoresHashChainColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

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
	t.Cleanup(func() { _ = db.Close() })

	// limit=3 → reader probes for limit+1 (4); only 3 rows exist → no probe row.
	// AnyArg() for the leading filter args + a literal last arg asserts the probe
	// value (limit+1) is the FINAL bound regardless of how many filter args precede
	// it, so a future filter that shifts arg order can't silently pass this guard.
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs(sqlmock.AnyArg(), 4).
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
	t.Cleanup(func() { _ = db.Close() })

	// limit=3 → probe for 4; 4 rows returned → trimmed to 3 + hasMore=true.
	// AnyArg() for the leading filter args + a literal last arg asserts the probe
	// value (limit+1) is the FINAL bound regardless of preceding filter args (see
	// TestWriterListEventsExactPageHasNoMore).
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs(sqlmock.AnyArg(), 4).
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
	t.Cleanup(func() { _ = db.Close() })

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
	t.Cleanup(func() { _ = db.Close() })

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
	t.Cleanup(func() { _ = db.Close() })

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

// ─── DEC-8: erased-tenant read-path gate ───────────────────────────────────
//
// openPayload's envelope-only redaction (IsEnvelope false for a plaintext
// row) let a plaintext audit_events row belonging to an erased tenant be
// served untouched through ListEvents. These tests cover: the plaintext
// case (the actual gap), the envelope case (must still redact, and must not
// even attempt decrypt), the not-erased case (no over-blocking — a normal
// tenant's payloads, plaintext and envelope alike, must pass through
// unchanged), and fail-closed behavior when the erasure lookup itself
// errors.

// stubTenantErasureChecker is a canned TenantErasureChecker double. err, when
// non-nil, is returned instead of erased/nil — used to exercise ListEvents'
// fail-closed path when the erasure lookup itself fails.
type stubTenantErasureChecker struct {
	erased bool
	err    error
}

func (s stubTenantErasureChecker) IsErased(_ context.Context, _ string) (bool, error) {
	return s.erased, s.err
}

// stubPayloadCrypto is a canned PayloadCrypto double. decryptCalls counts
// DecryptForTenant invocations so a test can assert the erasure gate
// short-circuits BEFORE any decrypt attempt (not merely that it produces
// the same redacted result).
type stubPayloadCrypto struct {
	decrypted    []byte
	decryptErr   error
	decryptCalls int
}

func (s *stubPayloadCrypto) EncryptForTenant(_ context.Context, _ string, _ []byte) (string, bool, error) {
	return "", false, nil
}

func (s *stubPayloadCrypto) EncryptForTenantTx(_ context.Context, _ *sql.Tx, _ string, _ []byte) (string, bool, error) {
	return "", false, nil
}

func (s *stubPayloadCrypto) DecryptForTenant(_ context.Context, _ string, _ string) ([]byte, error) {
	s.decryptCalls++
	return s.decrypted, s.decryptErr
}

func auditListRowsWithPayload(payload string) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"id", "occurred_at", "actor_id", "action",
		"resource_type", "resource_id", "payload", "trace_id", "tenant_id",
	})
	rows.AddRow("evt-1", time.Now().UTC(), "actor", "audit.test", "document", "doc", payload, "trace", "tenant-1")
	return rows
}

// TestWriterListEventsRedactsPlaintextForErasedTenant is the DEC-8 regression
// guard: a legacy plaintext row (never enveloped — IsEnvelope is false)
// belonging to a tenant the erasure gate reports as erased must be withheld,
// not served in full. This is the actual gap the fix closes; every other
// test in this block is defense/over-block coverage around it.
func TestWriterListEventsRedactsPlaintextForErasedTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	plaintext := `{"username":"real.person","note":"pii"}`
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs(sqlmock.AnyArg(), 21).
		WillReturnRows(auditListRowsWithPayload(plaintext))

	writer := NewWriter(db).WithTenantErasureCheck(stubTenantErasureChecker{erased: true})
	items, _, err := writer.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].PayloadJSON != redactedErasedTenantPayload {
		t.Fatalf("PayloadJSON = %q, want the erased-tenant marker %q (plaintext leaked)", items[0].PayloadJSON, redactedErasedTenantPayload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestWriterListEventsRedactsEnvelopeForErasedTenantWithoutDecrypting proves
// the erasure gate runs BEFORE the envelope decrypt attempt: for an erased
// tenant, an envelope-shaped row must also come back redacted, and
// DecryptForTenant must never be called (it would be pointless — the erased
// gate has already decided the outcome, and calling it needlessly exercises
// crypto against a key that may be destroyed).
func TestWriterListEventsRedactsEnvelopeForErasedTenantWithoutDecrypting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	envelope := platcrypto.EncodeEnvelope([]byte("sealed-bytes"))
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs(sqlmock.AnyArg(), 21).
		WillReturnRows(auditListRowsWithPayload(envelope))

	crypto := &stubPayloadCrypto{decrypted: []byte(`{"should":"never appear"}`)}
	writer := NewWriter(db).
		WithPayloadCrypto(crypto).
		WithTenantErasureCheck(stubTenantErasureChecker{erased: true})

	items, _, err := writer.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if items[0].PayloadJSON != redactedErasedTenantPayload {
		t.Fatalf("PayloadJSON = %q, want the erased-tenant marker %q", items[0].PayloadJSON, redactedErasedTenantPayload)
	}
	if crypto.decryptCalls != 0 {
		t.Fatalf("DecryptForTenant called %d times, want 0 (erasure gate must short-circuit first)", crypto.decryptCalls)
	}
}

// TestWriterListEventsPassesThroughPlaintextForActiveTenant is the
// over-block guard: a NOT-erased tenant's legacy plaintext payload must pass
// through unchanged. A fail-closed fix that also redacts normal tenants is a
// different defect, not a safer one.
func TestWriterListEventsPassesThroughPlaintextForActiveTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	plaintext := `{"name":"QMS"}`
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs(sqlmock.AnyArg(), 21).
		WillReturnRows(auditListRowsWithPayload(plaintext))

	writer := NewWriter(db).WithTenantErasureCheck(stubTenantErasureChecker{erased: false})
	items, _, err := writer.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if items[0].PayloadJSON != plaintext {
		t.Fatalf("PayloadJSON = %q, want unredacted %q (active tenant over-blocked)", items[0].PayloadJSON, plaintext)
	}
}

// TestWriterListEventsPassesThroughEnvelopeForActiveTenant is the
// over-block guard for the envelope path: a NOT-erased tenant's envelope row
// must still decrypt normally through the pre-existing crypto path.
func TestWriterListEventsPassesThroughEnvelopeForActiveTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	envelope := platcrypto.EncodeEnvelope([]byte("sealed-bytes"))
	mock.ExpectQuery("FROM metaldocs.audit_events").
		WithArgs(sqlmock.AnyArg(), 21).
		WillReturnRows(auditListRowsWithPayload(envelope))

	decrypted := `{"name":"QMS"}`
	crypto := &stubPayloadCrypto{decrypted: []byte(decrypted)}
	writer := NewWriter(db).
		WithPayloadCrypto(crypto).
		WithTenantErasureCheck(stubTenantErasureChecker{erased: false})

	items, _, err := writer.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if items[0].PayloadJSON != decrypted {
		t.Fatalf("PayloadJSON = %q, want decrypted %q", items[0].PayloadJSON, decrypted)
	}
	if crypto.decryptCalls != 1 {
		t.Fatalf("DecryptForTenant called %d times, want 1", crypto.decryptCalls)
	}
}

// TestWriterListEventsFailsClosedOnErasureCheckError proves ListEvents does
// not guess "not erased" when the erasure lookup itself fails — it refuses
// the whole read instead of risking a false negative that would serve
// withheld payload content. No rows should be returned.
func TestWriterListEventsFailsClosedOnErasureCheckError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	lookupErr := errors.New("connection reset")
	writer := NewWriter(db).WithTenantErasureCheck(stubTenantErasureChecker{err: lookupErr})

	items, hasMore, err := writer.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-1"})
	if err == nil {
		t.Fatalf("ListEvents: want error on erasure-check failure, got nil (items=%v)", items)
	}
	if !errors.Is(err, lookupErr) {
		t.Fatalf("ListEvents error = %v, want it to wrap %v", err, lookupErr)
	}
	if items != nil || hasMore {
		t.Fatalf("ListEvents = (%v, %v), want (nil, false) on failure", items, hasMore)
	}
	// No SQL expectations were set on mock at all — asserting no query ran
	// confirms the fail-closed check happens BEFORE the list query, not after.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
