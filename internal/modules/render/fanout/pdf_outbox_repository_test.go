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
	mock.ExpectExec("INSERT INTO metaldocs.pdf_dispatch_outbox").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPDFOutboxRepository(db)
	if err := repo.Enqueue(context.Background(), tx, "t1", "r1", []byte("hash")); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	_ = tx.Commit()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPDFOutboxRepository_Enqueue_Idempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO metaldocs.pdf_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 0))
	repo := NewPDFOutboxRepository(db)
	if err := repo.Enqueue(context.Background(), nil, "t1", "r1", []byte("hash")); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPDFOutboxRepository_MarkDispatched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectExec("UPDATE metaldocs.pdf_dispatch_outbox").WillReturnResult(sqlmock.NewResult(1, 1))
	repo := NewPDFOutboxRepository(db)
	if err := repo.MarkDispatched(context.Background(), "id-1"); err != nil {
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
	mock.ExpectExec("UPDATE metaldocs.pdf_dispatch_outbox").WillReturnResult(sqlmock.NewResult(1, 1))
	repo := NewPDFOutboxRepository(db)
	if err := repo.MarkFailed(context.Background(), "id-1", "bus error", time.Now().Add(30*time.Second), false); err != nil {
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
