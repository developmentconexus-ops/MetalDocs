package fanout

import (
	"context"
	"testing"
	"time"

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
	id, err := repo.Enqueue(context.Background(), tx, "t1", "r1", []byte("hash"))
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

func TestPDFOutboxRepository_Enqueue_NilTxRejected(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	// A nil tx must fail loud, not silently autocommit the outbox row outside the
	// caller's business transaction (db.Tx contract / transactional-outbox atomicity).
	repo := NewPDFOutboxRepository(db)
	id, err := repo.Enqueue(context.Background(), nil, "t1", "r1", []byte("hash"))
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
	id, err := repo.Enqueue(context.Background(), tx, "t1", "r1", []byte("hash"))
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
	id, err := repo.Enqueue(context.Background(), nil, "t1", "r1", []byte("hash"))
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

func TestPDFOutboxRepository_MarkFailed_AppliesBackoff(t *testing.T) {
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
	if err := repo.MarkFailed(context.Background(), "t1", "id-1", "bus error", time.Now().Add(30*time.Second), false); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPDFOutboxRepository_ResetStaleClaims(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectExec("UPDATE metaldocs.pdf_dispatch_outbox").WillReturnResult(sqlmock.NewResult(2, 2))
	repo := NewPDFOutboxRepository(db)
	n, err := repo.ResetStaleClaims(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("ResetStaleClaims: %v", err)
	}
	if n != 2 {
		t.Fatalf("want 2, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// --- APP-07: pin the retry/terminal state-machine contract documented in
// wiki/backend/flows/async-job-pipeline.md §7 "Staging outbox tables". ---

// TestPDFOutboxRepository_MarkFailed_RetryPath_ResetsClaimAndBumpsAttempts pins
// that finalize=false takes the retry branch: status goes back to 'pending',
// claimed_at is cleared (so the row is reclaimable), attempts is incremented,
// and next_retry_at is set from the caller-supplied backoff deadline — not the
// finalize/dead-letter branch.
func TestPDFOutboxRepository_MarkFailed_RetryPath_ResetsClaimAndBumpsAttempts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	nextRetry := time.Now().Add(30 * time.Second)
	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("t-retry").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE metaldocs\.pdf_dispatch_outbox\s+SET status='pending', last_error=\$2, attempts=attempts\+1, next_retry_at=\$3, claimed_at=NULL\s+WHERE id=\$1::uuid`).
		WithArgs("id-retry", "transient error", nextRetry).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	repo := NewPDFOutboxRepository(db)
	if err := repo.MarkFailed(context.Background(), "t-retry", "id-retry", "transient error", nextRetry, false); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

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
	if err := repo.MarkFailed(context.Background(), "t-final", "id-final", "permanent defect", time.Time{}, true); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestPDFOutboxRepository_MarkFailed_RowNotFound pins that both branches
// surface a "row not found" error when RowsAffected is 0, rather than
// silently succeeding — same shape for retry and finalize.
func TestPDFOutboxRepository_MarkFailed_RowNotFound(t *testing.T) {
	t.Run("retry branch", func(t *testing.T) {
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
		if err := repo.MarkFailed(context.Background(), "t-missing", "missing", "err", time.Now(), false); err == nil {
			t.Fatal("expected row-not-found error, got nil")
		}
	})
	t.Run("finalize branch", func(t *testing.T) {
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
		if err := repo.MarkFailed(context.Background(), "t-missing", "missing", "err", time.Time{}, true); err == nil {
			t.Fatal("expected row-not-found error, got nil")
		}
	})
}

// TestPDFOutboxRepository_ResetStaleClaims_OnlyTouchesProcessingOlderThanCutoff
// pins the WHERE clause: only status='processing' rows whose claimed_at
// predates the cutoff are reclaimed. pending/dispatched/failed rows and
// recently-claimed processing rows must never be touched by this predicate.
func TestPDFOutboxRepository_ResetStaleClaims_OnlyTouchesProcessingOlderThanCutoff(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	olderThan := 5 * time.Minute
	mock.ExpectExec(`UPDATE metaldocs\.pdf_dispatch_outbox\s+SET status='pending', claimed_at=NULL\s+WHERE status='processing' AND claimed_at < NOW\(\) - \(\$1 \* interval '1 millisecond'\)`).
		WithArgs(olderThan.Milliseconds()).
		WillReturnResult(sqlmock.NewResult(0, 3))
	repo := NewPDFOutboxRepository(db)
	n, err := repo.ResetStaleClaims(context.Background(), olderThan)
	if err != nil {
		t.Fatalf("ResetStaleClaims: %v", err)
	}
	if n != 3 {
		t.Fatalf("want 3, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestPDFOutboxRepository_ClaimPending_RespectsAttemptsAndRetryGate pins the
// ClaimPending predicate: status='pending', next_retry_at <= NOW(), and
// attempts < maxAttempts (the $2 bound parameter), ordered by next_retry_at
// ASC with FOR UPDATE SKIP LOCKED, returning the claimed rows with their
// tenant/revision/content_hash/attempts columns intact.
func TestPDFOutboxRepository_ClaimPending_RespectsAttemptsAndRetryGate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "revision_id", "content_hash", "attempts"}).
		AddRow("id-1", "t1", "r1", []byte("hash1"), 2)
	mock.ExpectQuery(`WITH claimed AS \(\s*SELECT id FROM metaldocs\.pdf_dispatch_outbox\s+WHERE status = 'pending' AND next_retry_at <= NOW\(\) AND attempts < \$2\s+ORDER BY next_retry_at ASC\s+LIMIT \$1\s+FOR UPDATE SKIP LOCKED\s*\)\s*UPDATE metaldocs\.pdf_dispatch_outbox o\s+SET status='processing', claimed_at=NOW\(\)\s+FROM claimed c\s+WHERE o\.id = c\.id\s+RETURNING o\.id, o\.tenant_id, o\.revision_id, o\.content_hash, o\.attempts`).
		WithArgs(10, 5).
		WillReturnRows(rows)
	repo := NewPDFOutboxRepository(db)
	got, err := repo.ClaimPending(context.Background(), 10, 5)
	if err != nil {
		t.Fatalf("ClaimPending: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 row, got %d", len(got))
	}
	if got[0].ID != "id-1" || got[0].TenantID != "t1" || got[0].RevisionID != "r1" || got[0].Attempts != 2 {
		t.Fatalf("unexpected row: %+v", got[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestPDFOutboxRepository_ClaimPending_NoEligibleRowsReturnsEmpty pins that a
// ClaimPending call with no matching rows (all rows are either not 'pending',
// gated by next_retry_at in the future, or at/above maxAttempts) returns an
// empty, non-nil-error slice rather than an error.
func TestPDFOutboxRepository_ClaimPending_NoEligibleRowsReturnsEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "revision_id", "content_hash", "attempts"})
	mock.ExpectQuery("WITH claimed AS").WithArgs(10, 5).WillReturnRows(rows)
	repo := NewPDFOutboxRepository(db)
	got, err := repo.ClaimPending(context.Background(), 10, 5)
	if err != nil {
		t.Fatalf("ClaimPending: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 rows, got %d", len(got))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
