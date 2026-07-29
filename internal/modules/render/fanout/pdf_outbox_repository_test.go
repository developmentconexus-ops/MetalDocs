package fanout

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPDFOutboxRepository_Enqueue_UsesTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO metaldocs.pdf_dispatch_outbox").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("row-1"))
	mock.ExpectCommit()
	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPDFOutboxRepository(db)
	id, err := repo.Enqueue(context.Background(), tx, "t1", "r1", []byte("hash"), "")
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if id != "row-1" {
		t.Fatalf("id = %q, want %q", id, "row-1")
	}
	_ = tx.Commit()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestPDFOutboxRepository_EnqueuePDF_PersistsFinalDocxKey pins F-QA2-2: the
// pdf-specific enqueue INSERTs the renderer-produced final_docx_s3_key as the
// 4th column/arg alongside content_hash, so the dispatched pdf event carries
// the key instead of dead-lettering on "missing final_docx_s3_key".
func TestPDFOutboxRepository_EnqueuePDF_PersistsFinalDocxKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO metaldocs\.pdf_dispatch_outbox \(tenant_id, revision_id, content_hash, final_docx_s3_key, release_generation_id\)`).
		WithArgs("t1", "r1", []byte("hash"), "tenants/t1/r1/frozen.docx", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("row-1"))
	mock.ExpectCommit()
	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPDFOutboxRepository(db)
	id, err := repo.EnqueuePDF(context.Background(), tx, "t1", "r1", []byte("hash"), "tenants/t1/r1/frozen.docx", "")
	if err != nil {
		t.Fatalf("EnqueuePDF: %v", err)
	}
	if id != "row-1" {
		t.Fatalf("id = %q, want %q", id, "row-1")
	}
	_ = tx.Commit()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestPDFOutboxRepository_EnqueuePDF_EmptyKeyFailsClosed pins the fail-closed
// invariant (F-QA2-2 / no-fallback): an empty final_docx_s3_key is a malformed
// pdf dispatch — the write is rejected before any INSERT runs, never silently
// enqueuing a row that would dead-letter downstream.
func TestPDFOutboxRepository_EnqueuePDF_EmptyKeyFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	// No ExpectQuery: the guard must reject before touching the DB.
	mock.ExpectBegin()
	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPDFOutboxRepository(db)
	id, err := repo.EnqueuePDF(context.Background(), tx, "t1", "r1", []byte("hash"), "", "")
	if err == nil {
		t.Fatal("expected error on empty final_docx_s3_key, got nil")
	}
	if id != "" {
		t.Fatalf("id = %q, want empty on fail-closed", id)
	}
}

func TestPDFOutboxRepository_Enqueue_NilTxRejected(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	// A nil tx must fail loud, not silently autocommit the outbox row outside the
	// caller's business transaction (db.Tx contract / transactional-outbox atomicity).
	repo := NewPDFOutboxRepository(db)
	id, err := repo.Enqueue(context.Background(), nil, "t1", "r1", []byte("hash"), "")
	if err == nil {
		t.Fatal("expected error on nil tx, got nil")
	}
	if id != "" {
		t.Fatalf("id = %q, want empty on error", id)
	}
}

// TestPDFOutboxRepository_Enqueue_Idempotent pins the ON CONFLICT DO NOTHING
// dedup-skip contract: a duplicate enqueue for the same (tenant_id,
// revision_id) yields zero RETURNING rows (sql.ErrNoRows), which Enqueue must
// translate to an empty id and a nil error — not an error — since dedup is a
// successful outcome for the caller (dispatchjobs.Enqueuer uses the empty id
// to skip the paired River insert).
func TestPDFOutboxRepository_Enqueue_Idempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	// ON CONFLICT DO NOTHING → 0 rows returned on a duplicate; Enqueue must not error.
	mock.ExpectQuery("INSERT INTO metaldocs.pdf_dispatch_outbox").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()
	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPDFOutboxRepository(db)
	id, err := repo.Enqueue(context.Background(), tx, "t1", "r1", []byte("hash"), "")
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if id != "" {
		t.Fatalf("id = %q, want empty on dedup skip", id)
	}
	_ = tx.Commit()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestMaterializeOutboxRepository_Enqueue_NilTxRejected(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	repo := NewMaterializeOutboxRepository(db)
	id, err := repo.Enqueue(context.Background(), nil, "t1", "r1", []byte("hash"), "")
	if err == nil {
		t.Fatal("expected error on nil tx, got nil")
	}
	if id != "" {
		t.Fatalf("id = %q, want empty on error", id)
	}
}

// M3 F3.2 (validation-contract.md §2.2 site 5) — MarkDispatched/MarkFailed
// now open their own tx and seed authz.SeedTxTenant with the row's tenant
// BEFORE the write, so every mark-write case below expects Begin → seed →
// UPDATE → Commit in that order.

func TestPDFOutboxRepository_MarkDispatched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("t1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE metaldocs.pdf_dispatch_outbox").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	repo := NewPDFOutboxRepository(db)
	if err := repo.MarkDispatched(context.Background(), "t1", "id-1"); err != nil {
		t.Fatalf("MarkDispatched: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// --- APP-07: pin the terminal state-machine contract documented in
// wiki/backend/flows/async-job-pipeline.md §7 "Staging outbox tables". River
// now owns retry/backoff scheduling for delivery (M5 F5.3 T4); MarkFailed
// always dead-letters. ---

// TestPDFOutboxRepository_MarkFailed_FinalizePath_SetsDeadLetteredAndFailedStatus
// pins that finalize=true takes the terminal branch: status='failed' and
// dead_lettered_at=NOW() are set (mirroring the platform outbox consumer's
// dead_lettered_at pattern), attempts is still incremented for observability,
// and — unlike the retry branch — next_retry_at/claimed_at are not touched.
func TestPDFOutboxRepository_MarkFailed_FinalizePath_SetsDeadLetteredAndFailedStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("t-final").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE metaldocs\.pdf_dispatch_outbox\s+SET status='failed', last_error=\$2, attempts=attempts\+1, dead_lettered_at=NOW\(\)\s+WHERE id=\$1::uuid`).
		WithArgs("id-final", "permanent defect").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	repo := NewPDFOutboxRepository(db)
	if err := repo.MarkFailed(context.Background(), "t-final", "id-final", "permanent defect"); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestPDFOutboxRepository_MarkFailed_RowNotFound pins that MarkFailed
// surfaces a "row not found" error when RowsAffected is 0, rather than
// silently succeeding.
func TestPDFOutboxRepository_MarkFailed_RowNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("t-missing").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE metaldocs.pdf_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	repo := NewPDFOutboxRepository(db)
	if err := repo.MarkFailed(context.Background(), "t-missing", "missing", "err"); err == nil {
		t.Fatal("expected row-not-found error, got nil")
	}
}
