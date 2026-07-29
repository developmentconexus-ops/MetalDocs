package dispatchjobs

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"metaldocs/internal/platform/messaging"
)

// fakePublisher records the last published event and can be made to fail.
type fakePublisher struct {
	published *messaging.Event
	err       error
}

func (f *fakePublisher) Publish(_ context.Context, event messaging.Event) error {
	if f.err != nil {
		return f.err
	}
	e := event
	f.published = &e
	return nil
}

// fakeOutboxMarker records MarkDispatched/MarkFailed calls and can be made
// to fail either call.
type fakeOutboxMarker struct {
	dispatchedTenantID, dispatchedID string
	dispatchCalls                    int

	failedTenantID, failedID, failedErrStr string
	failCalls                              int

	markDispatchedErr error
	markFailedErr     error
}

func (f *fakeOutboxMarker) MarkDispatched(_ context.Context, tenantID, id string) error {
	f.dispatchCalls++
	f.dispatchedTenantID = tenantID
	f.dispatchedID = id
	return f.markDispatchedErr
}

func (f *fakeOutboxMarker) MarkFailed(_ context.Context, tenantID, id, errStr string) error {
	f.failCalls++
	f.failedTenantID = tenantID
	f.failedID = id
	f.failedErrStr = errStr
	return f.markFailedErr
}

func testFields() dispatchFields {
	return dispatchFields{
		TenantID:   "tenant-1",
		RevisionID: "revision-1",
		OutboxID:   "outbox-1",
	}
}

// testPDFArgs carries the MATERIALIZED frozen-docx hash (migration 0312:
// pdf_dispatch_outbox.frozen_docx_hash), not a generic "content hash".
func testPDFArgs() PDFDispatchArgs {
	return PDFDispatchArgs{
		dispatchFields: testFields(),
		FrozenDocxHash: []byte{0xDE, 0xAD, 0xBE, 0xEF},
		FinalDocxS3Key: "tenants/tenant-1/revision-1/frozen.docx",
	}
}

// testMaterializeArgs carries the resolved-placeholder VALUES hash
// (migration 0312: materialize_dispatch_outbox.values_hash).
func testMaterializeArgs() MaterializeDispatchArgs {
	return MaterializeDispatchArgs{
		dispatchFields: testFields(),
		ValuesHash:     []byte{0x01, 0x02, 0x03, 0x04},
	}
}

func TestRun_PDFEvent_PublishSuccess_MarksDispatched(t *testing.T) {
	pub := &fakePublisher{}
	repo := &fakeOutboxMarker{}
	args := testPDFArgs()
	fields := args.dispatchFields

	err := run(context.Background(), pub, repo, fields, buildPDFEvent(args), 1, 25)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if pub.published == nil {
		t.Fatal("expected an event to be published")
	}
	if pub.published.EventType != messaging.EventTypePDFConvert {
		t.Errorf("EventType = %q, want %q", pub.published.EventType, messaging.EventTypePDFConvert)
	}
	wantKey := messaging.IdempotencyKey("docgen_v2_pdf:tenant-1:revision-1")
	if pub.published.IdempotencyKey != wantKey {
		t.Errorf("IdempotencyKey = %q, want %q", pub.published.IdempotencyKey, wantKey)
	}
	payload, ok := pub.published.Payload.(messaging.PDFConvertPayload)
	if !ok {
		t.Fatalf("Payload type = %T, want messaging.PDFConvertPayload", pub.published.Payload)
	}
	// The event contract keeps the name ContentHash; the producer-side value is
	// the frozen-docx hash (migration 0312).
	wantHash := hex.EncodeToString(args.FrozenDocxHash)
	if payload.ContentHash != wantHash {
		t.Errorf("ContentHash = %q, want %q", payload.ContentHash, wantHash)
	}
	// F-QA2-2: buildPDFEvent must thread the frozen-docx key into the pdf
	// payload so the pdf job runner never dead-letters on a missing key.
	if payload.FinalDocxS3Key != args.FinalDocxS3Key {
		t.Errorf("FinalDocxS3Key = %q, want %q", payload.FinalDocxS3Key, args.FinalDocxS3Key)
	}

	if repo.dispatchCalls != 1 {
		t.Fatalf("MarkDispatched calls = %d, want 1", repo.dispatchCalls)
	}
	if repo.dispatchedTenantID != fields.TenantID || repo.dispatchedID != fields.OutboxID {
		t.Errorf("MarkDispatched(%q, %q), want (%q, %q)", repo.dispatchedTenantID, repo.dispatchedID, fields.TenantID, fields.OutboxID)
	}
	if repo.failCalls != 0 {
		t.Errorf("MarkFailed calls = %d, want 0", repo.failCalls)
	}
}

func TestRun_MaterializeEvent_BuildsExpectedEvent(t *testing.T) {
	pub := &fakePublisher{}
	repo := &fakeOutboxMarker{}
	args := testMaterializeArgs()

	if err := run(context.Background(), pub, repo, args.dispatchFields, buildMaterializeEvent(args), 1, 25); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if pub.published.EventType != messaging.EventTypeMaterializeFanout {
		t.Errorf("EventType = %q, want %q", pub.published.EventType, messaging.EventTypeMaterializeFanout)
	}
	wantKey := messaging.IdempotencyKey("materialize_fanout:tenant-1:revision-1")
	if pub.published.IdempotencyKey != wantKey {
		t.Errorf("IdempotencyKey = %q, want %q", pub.published.IdempotencyKey, wantKey)
	}
}

func TestRun_PublishFails_NotTerminal_RetriesWithoutDeadLetter(t *testing.T) {
	publishErr := errors.New("publish boom")
	pub := &fakePublisher{err: publishErr}
	repo := &fakeOutboxMarker{}
	args := testPDFArgs()
	fields := args.dispatchFields

	err := run(context.Background(), pub, repo, fields, buildPDFEvent(args), 1, 25)
	if !errors.Is(err, publishErr) {
		t.Fatalf("run error = %v, want %v", err, publishErr)
	}
	if repo.failCalls != 0 {
		t.Errorf("MarkFailed calls = %d, want 0 (not terminal attempt)", repo.failCalls)
	}
	if repo.dispatchCalls != 0 {
		t.Errorf("MarkDispatched calls = %d, want 0", repo.dispatchCalls)
	}
}

func TestRun_PublishFails_TerminalAttempt_DeadLetters(t *testing.T) {
	publishErr := errors.New("publish boom terminal")
	pub := &fakePublisher{err: publishErr}
	repo := &fakeOutboxMarker{}
	args := testPDFArgs()
	fields := args.dispatchFields

	err := run(context.Background(), pub, repo, fields, buildPDFEvent(args), 25, 25)
	if !errors.Is(err, publishErr) {
		t.Fatalf("run error = %v, want %v", err, publishErr)
	}
	if repo.failCalls != 1 {
		t.Fatalf("MarkFailed calls = %d, want 1", repo.failCalls)
	}
	if repo.failedTenantID != fields.TenantID || repo.failedID != fields.OutboxID {
		t.Errorf("MarkFailed(%q, %q), want (%q, %q)", repo.failedTenantID, repo.failedID, fields.TenantID, fields.OutboxID)
	}
	if repo.failedErrStr != publishErr.Error() {
		t.Errorf("MarkFailed errStr = %q, want %q", repo.failedErrStr, publishErr.Error())
	}
}

func TestRun_MarkDispatchedFails_TerminalAttempt_DeadLetters(t *testing.T) {
	markErr := errors.New("mark dispatched boom")
	pub := &fakePublisher{}
	repo := &fakeOutboxMarker{markDispatchedErr: markErr}
	args := testPDFArgs()
	fields := args.dispatchFields

	err := run(context.Background(), pub, repo, fields, buildPDFEvent(args), 25, 25)
	if !errors.Is(err, markErr) {
		t.Fatalf("run error = %v, want %v", err, markErr)
	}
	if repo.failCalls != 1 {
		t.Fatalf("MarkFailed calls = %d, want 1", repo.failCalls)
	}
}

func TestPDFDispatchWorker_Work_NilJob(t *testing.T) {
	w := &PDFDispatchWorker{}
	if err := w.Work(context.Background(), nil); !errors.Is(err, ErrPDFDispatchJobNil) {
		t.Fatalf("Work(nil) error = %v, want %v", err, ErrPDFDispatchJobNil)
	}
}

func TestMaterializeDispatchWorker_Work_NilJob(t *testing.T) {
	w := &MaterializeDispatchWorker{}
	if err := w.Work(context.Background(), nil); !errors.Is(err, ErrMaterializeDispatchJobNil) {
		t.Fatalf("Work(nil) error = %v, want %v", err, ErrMaterializeDispatchJobNil)
	}
}
