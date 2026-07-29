package dispatchjobs

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"metaldocs/internal/platform/db"
)

// fakeStagingRepo is a minimal fake satisfying the enqueuer's stagingEnqueuer
// dependency (mirrors *fanout.StagingOutboxRepository.Enqueue's shape). Kept
// unexported and local to this test file, same pattern as outboxMarker /
// fakeOutboxMarker in workers_test.go.
type fakeStagingRepo struct {
	id                         string
	err                        error
	calls                      int
	gotTenantID, gotRevisionID string
	gotContentHash             []byte
	gotFinalDocxS3Key          string
}

func (f *fakeStagingRepo) Enqueue(_ context.Context, _ db.Tx, tenantID, revisionID string, contentHash []byte, releaseGenerationID string) (string, error) {
	f.calls++
	f.gotTenantID = tenantID
	f.gotRevisionID = revisionID
	f.gotContentHash = contentHash
	return f.id, f.err
}

func (f *fakeStagingRepo) EnqueuePDF(_ context.Context, _ db.Tx, tenantID, revisionID string, contentHash []byte, finalDocxS3Key, releaseGenerationID string) (string, error) {
	f.calls++
	f.gotTenantID = tenantID
	f.gotRevisionID = revisionID
	f.gotContentHash = contentHash
	f.gotFinalDocxS3Key = finalDocxS3Key
	return f.id, f.err
}

// fakeRiverInserter is a recording fake for the unexported riverInserter
// dependency, letting the enqueuer be unit-tested without a real *sql.Tx or
// database (constructing a real *sql.Tx is not possible outside package
// database/sql). It captures the args/opts passed to InsertTx.
type fakeRiverInserter struct {
	calls   int
	gotArgs river.JobArgs
	gotOpts *river.InsertOpts
	err     error
}

func (f *fakeRiverInserter) InsertTx(_ context.Context, _ *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	f.calls++
	f.gotArgs = args
	f.gotOpts = opts
	if f.err != nil {
		return nil, f.err
	}
	return &rivertype.JobInsertResult{}, nil
}

// fakeDBTx is a non-*sql.Tx db.Tx implementation used to exercise the
// fail-loud type-assertion branch (case c).
type fakeDBTx struct{}

func (fakeDBTx) ExecContext(context.Context, string, ...any) (sql.Result, error) { return nil, nil }
func (fakeDBTx) QueryContext(context.Context, string, ...any) (*sql.Rows, error) { return nil, nil }
func (fakeDBTx) QueryRowContext(context.Context, string, ...any) *sql.Row        { return nil }

func TestEnqueuer_EnqueuePDFTx_NewRow_InsertsRiverJob(t *testing.T) {
	repo := &fakeStagingRepo{id: "outbox-1"}
	client := &fakeRiverInserter{}
	e := newEnqueuerWithInserter(client, repo, nil, 25)

	err := e.EnqueuePDFTx(context.Background(), &sql.Tx{}, "tenant-1", "revision-1", []byte{0xAB}, "tenants/t1/r1/frozen.docx", "")
	if err != nil {
		t.Fatalf("EnqueuePDFTx: %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("repo.Enqueue calls = %d, want 1", repo.calls)
	}
	if client.calls != 1 {
		t.Fatalf("InsertTx calls = %d, want 1", client.calls)
	}
	args, ok := client.gotArgs.(PDFDispatchArgs)
	if !ok {
		t.Fatalf("InsertTx args type = %T, want PDFDispatchArgs", client.gotArgs)
	}
	if args.Kind() != "pdf_dispatch" {
		t.Errorf("Kind() = %q, want pdf_dispatch", args.Kind())
	}
	if args.OutboxID != "outbox-1" || args.TenantID != "tenant-1" || args.RevisionID != "revision-1" {
		t.Errorf("unexpected dispatchFields: %+v", args.dispatchFields)
	}
	// F-QA2-2: the renderer-produced final_docx_s3_key must be threaded through
	// both the outbox row (repo.gotFinalDocxS3Key) and the River job args
	// (args.FinalDocxS3Key) so the pdf runner never dead-letters on a missing key.
	if repo.gotFinalDocxS3Key != "tenants/t1/r1/frozen.docx" {
		t.Errorf("repo.gotFinalDocxS3Key = %q, want tenants/t1/r1/frozen.docx", repo.gotFinalDocxS3Key)
	}
	if args.FinalDocxS3Key != "tenants/t1/r1/frozen.docx" {
		t.Errorf("args.FinalDocxS3Key = %q, want tenants/t1/r1/frozen.docx", args.FinalDocxS3Key)
	}
	if client.gotOpts == nil || client.gotOpts.Queue != "temporal" {
		t.Errorf("InsertOpts.Queue = %+v, want temporal", client.gotOpts)
	}
	if client.gotOpts.MaxAttempts != 25 {
		t.Errorf("InsertOpts.MaxAttempts = %d, want 25", client.gotOpts.MaxAttempts)
	}
}

func TestEnqueuer_EnqueueMaterializeTx_NewRow_InsertsRiverJob(t *testing.T) {
	repo := &fakeStagingRepo{id: "outbox-2"}
	client := &fakeRiverInserter{}
	e := newEnqueuerWithInserter(client, nil, repo, 10)

	err := e.EnqueueMaterializeTx(context.Background(), &sql.Tx{}, "tenant-2", "revision-2", []byte{0xCD}, "")
	if err != nil {
		t.Fatalf("EnqueueMaterializeTx: %v", err)
	}
	args, ok := client.gotArgs.(MaterializeDispatchArgs)
	if !ok {
		t.Fatalf("InsertTx args type = %T, want MaterializeDispatchArgs", client.gotArgs)
	}
	if args.Kind() != "materialize_dispatch" {
		t.Errorf("Kind() = %q, want materialize_dispatch", args.Kind())
	}
	if args.OutboxID != "outbox-2" {
		t.Errorf("OutboxID = %q, want outbox-2", args.OutboxID)
	}
}

func TestEnqueuer_DedupSkip_NoRiverInsert(t *testing.T) {
	repo := &fakeStagingRepo{id: ""} // empty id => dedup skip
	client := &fakeRiverInserter{}
	e := newEnqueuerWithInserter(client, repo, nil, 25)

	err := e.EnqueuePDFTx(context.Background(), &sql.Tx{}, "tenant-1", "revision-1", []byte{0xAB}, "tenants/t1/r1/frozen.docx", "")
	if err != nil {
		t.Fatalf("EnqueuePDFTx: %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("repo.Enqueue calls = %d, want 1", repo.calls)
	}
	if client.calls != 0 {
		t.Fatalf("InsertTx calls = %d, want 0 on dedup skip", client.calls)
	}
}

func TestEnqueuer_RepoError_PropagatesAndSkipsRiver(t *testing.T) {
	repoErr := errors.New("boom")
	repo := &fakeStagingRepo{err: repoErr}
	client := &fakeRiverInserter{}
	e := newEnqueuerWithInserter(client, repo, nil, 25)

	err := e.EnqueuePDFTx(context.Background(), &sql.Tx{}, "tenant-1", "revision-1", []byte{0xAB}, "tenants/t1/r1/frozen.docx", "")
	if !errors.Is(err, repoErr) {
		t.Fatalf("err = %v, want %v", err, repoErr)
	}
	if client.calls != 0 {
		t.Fatalf("InsertTx calls = %d, want 0 on repo error", client.calls)
	}
}

// TestEnqueuer_NonSQLTx_FailsLoud exercises the type-assertion boundary via
// the real production Enqueuer (constructed with NewEnqueuer, wrapping a real
// riverInserter-shaped client type), passing a non-*sql.Tx db.Tx. This proves
// the fail-loud contract without needing a real database.
func TestEnqueuer_NonSQLTx_FailsLoud(t *testing.T) {
	repo := &fakeStagingRepo{id: "outbox-1"}
	client := &fakeRiverInserter{}
	e := newEnqueuerWithInserter(client, repo, repo, 25)

	err := e.EnqueuePDFTx(context.Background(), fakeDBTx{}, "tenant-1", "revision-1", []byte{0xAB}, "tenants/t1/r1/frozen.docx", "")
	if err == nil {
		t.Fatal("expected error for non-*sql.Tx, got nil")
	}
	if client.calls != 0 {
		t.Fatalf("InsertTx calls = %d, want 0 when tx assertion fails", client.calls)
	}

	err = e.EnqueueMaterializeTx(context.Background(), fakeDBTx{}, "tenant-1", "revision-1", []byte{0xAB}, "")
	if err == nil {
		t.Fatal("expected error for non-*sql.Tx, got nil")
	}
}
