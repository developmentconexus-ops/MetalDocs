package worker

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/servicebus"
)

// fakeAuthzSeam is a no-op TxAuthzSeam double: WithBackgroundBypass returns
// ctx unchanged, and the tx-scoped methods succeed without touching tx. Unit
// tests here exercise the runner's control flow, not the real iam/authz
// bridge (that is covered by the integration tests in
// pdf_job_runner_integration_test.go / materialize_job_runner_integration_test.go).
type fakeAuthzSeam struct {
	seedErr     error
	bypassErr   error
	seedCalls   int
	bypassCalls int
}

func (f *fakeAuthzSeam) WithBackgroundBypass(ctx context.Context) context.Context {
	return ctx
}

func (f *fakeAuthzSeam) SeedTxTenant(_ context.Context, _ *sql.Tx, _ string) error {
	f.seedCalls++
	return f.seedErr
}

func (f *fakeAuthzSeam) BypassSystem(_ context.Context, _ *sql.Tx) error {
	f.bypassCalls++
	return f.bypassErr
}

type fakePDFConverter struct {
	calls  int
	req    servicebus.ConvertPDFRequest
	result servicebus.ConvertPDFResult
	err    error
}

func (f *fakePDFConverter) ConvertPDF(_ context.Context, req servicebus.ConvertPDFRequest) (servicebus.ConvertPDFResult, error) {
	f.calls++
	f.req = req
	if f.err != nil {
		return servicebus.ConvertPDFResult{}, f.err
	}
	return f.result, nil
}

type fakePDFPersister struct {
	calls []pdfPersistCall
	err   error
}

type pdfPersistCall struct {
	tenant      TenantID
	docID       DocumentID
	s3Key       StorageKey
	pdfHash     []byte
	generatedAt time.Time
}

func (f *fakePDFPersister) WritePDF(_ context.Context, req PDFWriteRequest) error {
	f.calls = append(f.calls, pdfPersistCall{
		tenant:      req.TenantID,
		docID:       req.DocumentID,
		s3Key:       req.StorageKey,
		pdfHash:     append([]byte(nil), req.PDFHash...),
		generatedAt: req.GeneratedAt,
	})
	return f.err
}

func makePDFEvent(payload messaging.PDFConvertPayload) messaging.Event {
	return messaging.Event{
		EventID:   "event-1",
		EventType: messaging.EventTypePDFConvert,
		Payload:   payload,
	}
}

func TestPDFJobRunner_Handle_Success(t *testing.T) {
	hash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	converter := &fakePDFConverter{
		result: servicebus.ConvertPDFResult{
			OutputKey:   "tenants/tenant-1/revisions/rev-1/final.pdf",
			ContentHash: hash,
		},
	}
	persister := &fakePDFPersister{}
	runner := NewPDFJobRunner(converter, persister, &fakeAuthzSeam{})

	err := runner.Handle(context.Background(), makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       "tenant-1",
		RevisionID:     "rev-1",
		FinalDocxS3Key: "tenants/tenant-1/revisions/rev-1/final.docx",
	}))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if converter.calls != 1 {
		t.Fatalf("ConvertPDF calls = %d, want 1", converter.calls)
	}
	if converter.req.DocxKey != "tenants/tenant-1/revisions/rev-1/final.docx" {
		t.Fatalf("DocxKey = %q", converter.req.DocxKey)
	}
	if converter.req.OutputKey != "tenants/tenant-1/revisions/rev-1/final.pdf" {
		t.Fatalf("OutputKey = %q", converter.req.OutputKey)
	}

	if len(persister.calls) != 1 {
		t.Fatalf("WritePDF calls = %d, want 1", len(persister.calls))
	}
	call := persister.calls[0]
	if call.tenant != "tenant-1" || call.docID != "rev-1" || call.s3Key != StorageKey(converter.result.OutputKey) {
		t.Fatalf("WritePDF args = tenant %q docID %q s3Key %q", call.tenant, call.docID, call.s3Key)
	}
	wantHash, err := hex.DecodeString(hash)
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	if hex.EncodeToString(call.pdfHash) != hex.EncodeToString(wantHash) {
		t.Fatalf("pdfHash = %x, want %x", call.pdfHash, wantHash)
	}
	if call.generatedAt.IsZero() {
		t.Fatalf("generatedAt is zero")
	}
	if call.generatedAt.Location() != time.UTC {
		t.Fatalf("generatedAt location = %v, want UTC", call.generatedAt.Location())
	}
}

func TestPDFJobRunner_Handle_MissingPayloadFields(t *testing.T) {
	converter := &fakePDFConverter{}
	persister := &fakePDFPersister{}
	runner := NewPDFJobRunner(converter, persister, &fakeAuthzSeam{})

	err := runner.Handle(context.Background(), makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       "tenant-1",
		RevisionID:     "",
		FinalDocxS3Key: "tenants/tenant-1/revisions/rev-1/final.docx",
	}))
	if err == nil {
		t.Fatalf("Handle error = nil, want error")
	}
	if converter.calls != 0 {
		t.Fatalf("ConvertPDF calls = %d, want 0", converter.calls)
	}
	if len(persister.calls) != 0 {
		t.Fatalf("WritePDF calls = %d, want 0", len(persister.calls))
	}
}

func TestPDFJobRunner_Handle_ConvertError(t *testing.T) {
	convertErr := errors.New("convert failed")
	converter := &fakePDFConverter{err: convertErr}
	persister := &fakePDFPersister{}
	runner := NewPDFJobRunner(converter, persister, &fakeAuthzSeam{})

	err := runner.Handle(context.Background(), makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       "tenant-1",
		RevisionID:     "rev-1",
		FinalDocxS3Key: "tenants/tenant-1/revisions/rev-1/final.docx",
	}))
	if !errors.Is(err, convertErr) {
		t.Fatalf("Handle error = %v, want wrapped convert error", err)
	}
	if len(persister.calls) != 0 {
		t.Fatalf("WritePDF calls = %d, want 0", len(persister.calls))
	}
}

// ---------------------------------------------------------------------------
// M3 F3.2 PG-2 (validation-contract.md §2.2 site 2) — the pdf job runner must
// wrap its write in a tx seeded with authz.SeedTxTenant BEFORE the write, when
// constructed via NewPDFJobRunnerWithDB.
// ---------------------------------------------------------------------------

// fakePDFPersisterInTx is a spying PDFPersisterInTx + PDFPersister double that
// records the tx it was called with (nil for the untransacted path) so tests
// can assert the tx-wrapped path was actually exercised.
type fakePDFPersisterInTx struct {
	inTxCalls []PDFWriteRequest
	err       error
}

func (f *fakePDFPersisterInTx) WritePDFInTx(_ context.Context, tx db.Tx, req PDFWriteRequest) error {
	if tx == nil {
		return fmt.Errorf("WritePDFInTx: tx must not be nil")
	}
	f.inTxCalls = append(f.inTxCalls, req)
	return f.err
}

func (f *fakePDFPersisterInTx) WritePDF(_ context.Context, _ PDFWriteRequest) error {
	return fmt.Errorf("WritePDF (untransacted) must not be called when db != nil")
}

// pdfNopConn is a minimal driver.Conn/Tx pair supporting ExecContext, so the
// F3.2 SeedTxTenant tx-local set_config call (issued before the PDF write)
// succeeds against a fake driver instead of falling back to Prepare (errors).
type pdfNopConn struct{}
type pdfNopTx struct{}
type pdfNopResult struct{}

func (pdfNopResult) LastInsertId() (int64, error) { return 0, nil }
func (pdfNopResult) RowsAffected() (int64, error) { return 1, nil }

func (pdfNopConn) Prepare(q string) (driver.Stmt, error) {
	return nil, fmt.Errorf("pdfNopConn: no statements")
}
func (pdfNopConn) Close() error              { return nil }
func (pdfNopConn) Begin() (driver.Tx, error) { return &pdfNopTx{}, nil }
func (pdfNopConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return pdfNopResult{}, nil
}
func (*pdfNopTx) Commit() error   { return nil }
func (*pdfNopTx) Rollback() error { return nil }

type pdfNopDriver struct{}

func (pdfNopDriver) Open(_ string) (driver.Conn, error) { return pdfNopConn{}, nil }

var pdfNopDBCounter int

func newPDFNopDB(t *testing.T) *sql.DB {
	t.Helper()
	pdfNopDBCounter++
	name := fmt.Sprintf("nop_pdf_%d", pdfNopDBCounter)
	sql.Register(name, pdfNopDriver{})
	sqlDB, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open nop db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}

func TestPDFJobRunner_HandleWithDB_WrapsWriteInSeededTx(t *testing.T) {
	converter := &fakePDFConverter{
		result: servicebus.ConvertPDFResult{
			OutputKey:   "tenants/tenant-1/revisions/rev-1/final.pdf",
			ContentHash: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
	}
	persister := &fakePDFPersisterInTx{}
	sqlDB := newPDFNopDB(t)

	runner := NewPDFJobRunnerWithDB(converter, persister, &fakeAuthzSeam{}, sqlDB)
	err := runner.Handle(context.Background(), makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       "tenant-1",
		RevisionID:     "rev-1",
		FinalDocxS3Key: "tenants/tenant-1/revisions/rev-1/final.docx",
	}))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(persister.inTxCalls) != 1 {
		t.Fatalf("WritePDFInTx calls = %d, want 1 (tx-wrapped path must be used)", len(persister.inTxCalls))
	}
	if persister.inTxCalls[0].TenantID != "tenant-1" {
		t.Fatalf("WritePDFInTx tenant = %q, want tenant-1", persister.inTxCalls[0].TenantID)
	}
}

func TestPDFJobRunner_HandleWithDB_NonTxPersisterErrors(t *testing.T) {
	converter := &fakePDFConverter{
		result: servicebus.ConvertPDFResult{
			OutputKey:   "tenants/tenant-1/revisions/rev-1/final.pdf",
			ContentHash: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
	}
	// A plain PDFPersister (no PDFPersisterInTx) must fail loudly rather than
	// silently skip the RLS seed.
	persister := &fakePDFPersister{}
	sqlDB := newPDFNopDB(t)

	runner := NewPDFJobRunnerWithDB(converter, persister, &fakeAuthzSeam{}, sqlDB)
	err := runner.Handle(context.Background(), makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       "tenant-1",
		RevisionID:     "rev-1",
		FinalDocxS3Key: "tenants/tenant-1/revisions/rev-1/final.docx",
	}))
	if err == nil {
		t.Fatal("expected error when persister does not implement PDFPersisterInTx")
	}
}

func TestPDFJobRunner_Handle_PersistError(t *testing.T) {
	persistErr := errors.New("persist failed")
	converter := &fakePDFConverter{
		result: servicebus.ConvertPDFResult{
			OutputKey:   "tenants/tenant-1/revisions/rev-1/final.pdf",
			ContentHash: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
	}
	persister := &fakePDFPersister{err: persistErr}
	runner := NewPDFJobRunner(converter, persister, &fakeAuthzSeam{})

	err := runner.Handle(context.Background(), makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       "tenant-1",
		RevisionID:     "rev-1",
		FinalDocxS3Key: "tenants/tenant-1/revisions/rev-1/final.docx",
	}))
	if !errors.Is(err, persistErr) {
		t.Fatalf("Handle error = %v, want wrapped persist error", err)
	}
	if len(persister.calls) != 1 {
		t.Fatalf("WritePDF calls = %d, want 1", len(persister.calls))
	}
}
