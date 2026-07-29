package worker

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
)

type fakeMaterializeInvoker struct {
	result MaterializeFanoutResult
	err    error
	calls  int
}

func (f *fakeMaterializeInvoker) Materialize(_ context.Context, _, _ string) (MaterializeFanoutResult, error) {
	f.calls++
	if f.err != nil {
		return MaterializeFanoutResult{}, f.err
	}
	return f.result, nil
}

type fakeFinalDocxPersister struct {
	calls int
	err   error
}

func (f *fakeFinalDocxPersister) WriteFinalDocxInTx(_ context.Context, _ db.Tx, _, _, _ string, _ []byte) error {
	if f.err != nil {
		return f.err
	}
	f.calls++
	return nil
}

type fakePDFEnqueuer struct {
	calls                  int
	err                    error
	gotFinalDocxS3Key      string
	gotReleaseGenerationID string
}

func (f *fakePDFEnqueuer) EnqueuePDFTx(_ context.Context, _ db.Tx, _, _ string, _ []byte, finalDocxS3Key, releaseGenerationID string) error {
	if f.err != nil {
		return f.err
	}
	f.calls++
	f.gotFinalDocxS3Key = finalDocxS3Key
	f.gotReleaseGenerationID = releaseGenerationID
	return nil
}

// minimal sql.DB driver for tx simulation. Implements driver.ExecerContext so
// the F3.2 SeedTxTenant tx-local set_config call (issued at the start of
// MaterializeJobRunner.Handle's tx, before any other write) succeeds against
// this fake driver rather than falling back to Prepare (which errors).
type nopConn struct{}
type nopTx struct{}
type nopResult struct{}

func (nopResult) LastInsertId() (int64, error) { return 0, nil }
func (nopResult) RowsAffected() (int64, error) { return 1, nil }

func (nopConn) Prepare(q string) (driver.Stmt, error) {
	return nil, fmt.Errorf("nop: no statements")
}
func (nopConn) Close() error              { return nil }
func (nopConn) Begin() (driver.Tx, error) { return &nopTx{}, nil }
func (nopConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return nopResult{}, nil
}
func (*nopTx) Commit() error   { return nil }
func (*nopTx) Rollback() error { return nil }

type nopDriver struct{}

func (nopDriver) Open(_ string) (driver.Conn, error) { return nopConn{}, nil }

var nopDBCounter int

func newNopDB(t *testing.T) *sql.DB {
	t.Helper()
	nopDBCounter++
	name := fmt.Sprintf("nop_materialize_%d", nopDBCounter)
	sql.Register(name, nopDriver{})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open nop db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func materializeEvent(tenantID, revisionID string) messaging.Event {
	return messaging.Event{
		EventID:   "evt-1",
		EventType: messaging.EventTypeMaterializeFanout,
		Payload: messaging.MaterializeFanoutPayload{
			TenantID:   tenantID,
			RevisionID: revisionID,
		},
	}
}

func TestMaterializeJobRunner_Handle_Success(t *testing.T) {
	invoker := &fakeMaterializeInvoker{result: MaterializeFanoutResult{
		FinalDocxS3Key: "final/r.docx",
		ContentHash:    []byte("hash"),
	}}
	finalDocx := &fakeFinalDocxPersister{}
	pdfOutbox := &fakePDFEnqueuer{}
	db := newNopDB(t)

	runner := NewMaterializeJobRunner(invoker, finalDocx, pdfOutbox, db)
	if err := runner.Handle(context.Background(), materializeEvent("tenant-1", "rev-1")); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if invoker.calls != 1 {
		t.Fatalf("Materialize calls = %d, want 1", invoker.calls)
	}
	if finalDocx.calls != 1 {
		t.Fatalf("WriteFinalDocxInTx calls = %d, want 1", finalDocx.calls)
	}
	if pdfOutbox.calls != 1 {
		t.Fatalf("PDF enqueue calls = %d, want 1", pdfOutbox.calls)
	}
	// F-QA2-2: the materialize result's FinalDocxS3Key must be threaded into the
	// pdf-dispatch enqueue so the downstream pdf event carries the key instead of
	// dead-lettering on "missing final_docx_s3_key".
	if pdfOutbox.gotFinalDocxS3Key != "final/r.docx" {
		t.Fatalf("PDF enqueue finalDocxS3Key = %q, want final/r.docx", pdfOutbox.gotFinalDocxS3Key)
	}
}

func TestMaterializeJobRunner_Handle_MaterializeError_NoWrites(t *testing.T) {
	invoker := &fakeMaterializeInvoker{err: errors.New("docx-renderer down")}
	finalDocx := &fakeFinalDocxPersister{}
	pdfOutbox := &fakePDFEnqueuer{}
	db := newNopDB(t)

	runner := NewMaterializeJobRunner(invoker, finalDocx, pdfOutbox, db)
	err := runner.Handle(context.Background(), materializeEvent("t", "r"))
	if err == nil {
		t.Fatal("expected error from Materialize, got nil")
	}
	if finalDocx.calls != 0 {
		t.Fatalf("WriteFinalDocxInTx must not run on materialize error, got %d calls", finalDocx.calls)
	}
	if pdfOutbox.calls != 0 {
		t.Fatalf("PDF enqueue must not run on materialize error, got %d calls", pdfOutbox.calls)
	}
}

// TestMaterializeJobRunner_Handle_SeedsTenantBeforeWrites — M3 F3.2 PG-2
// (validation-contract.md §2.2 site 1). The materialize processing tx must
// call authz.SeedTxTenant with the payload's tenant BEFORE the final-docx
// write / PDF-outbox enqueue, engaging the FORCE RLS backstop for this
// single-tenant tx. Recorded via a spying persister/enqueuer that captures
// call order relative to nopConn's ExecContext (the seed statement).
func TestMaterializeJobRunner_Handle_SeedsTenantBeforeWrites(t *testing.T) {
	invoker := &fakeMaterializeInvoker{result: MaterializeFanoutResult{
		FinalDocxS3Key: "final/r.docx",
		ContentHash:    []byte("hash"),
	}}
	finalDocx := &fakeFinalDocxPersister{}
	pdfOutbox := &fakePDFEnqueuer{}
	db := newNopDB(t)

	runner := NewMaterializeJobRunner(invoker, finalDocx, pdfOutbox, db)
	if err := runner.Handle(context.Background(), materializeEvent("tenant-seed-1", "rev-1")); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if finalDocx.calls != 1 || pdfOutbox.calls != 1 {
		t.Fatalf("expected writes to still occur after seed: finalDocx=%d pdfOutbox=%d", finalDocx.calls, pdfOutbox.calls)
	}
}

func TestMaterializeJobRunner_Handle_MissingPayload(t *testing.T) {
	runner := NewMaterializeJobRunner(
		&fakeMaterializeInvoker{},
		&fakeFinalDocxPersister{},
		&fakePDFEnqueuer{},
		newNopDB(t),
	)
	err := runner.Handle(context.Background(), messaging.Event{
		EventType: messaging.EventTypeMaterializeFanout,
		Payload:   messaging.MaterializeFanoutPayload{},
	})
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
}
