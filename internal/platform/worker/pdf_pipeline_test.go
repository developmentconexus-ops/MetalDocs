package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/render/gotenberg"
	"metaldocs/internal/platform/servicebus"
)

// fakePDFStore is an in-memory object store satisfying the (unexported)
// object-store surface that GotenbergPDFClient depends on.
type fakePDFStore struct {
	objects map[string][]byte
	saved   map[string][]byte
}

func (f *fakePDFStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	b, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (f *fakePDFStore) Save(_ context.Context, key string, content []byte) error {
	if f.saved == nil {
		f.saved = map[string][]byte{}
	}
	f.saved[key] = append([]byte(nil), content...)
	return nil
}

func newGotenbergPDFConverter(t *testing.T, gotenbergURL string, store *fakePDFStore) *servicebus.GotenbergPDFClient {
	t.Helper()
	client, err := gotenberg.NewClient(gotenbergURL)
	if err != nil {
		t.Fatalf("gotenberg.NewClient: %v", err)
	}
	return servicebus.NewGotenbergPDFClient(store, client)
}

// TestPDFPipeline_RealClient exercises the real Gotenberg client + GotenbergPDFClient
// against a fake Gotenberg server, then verifies PDFJobRunner writes the result via
// the persister. Tests the actual HTTP + object-store path — not mocked at the interface level.
func TestPDFPipeline_RealClient(t *testing.T) {
	const (
		tenantID   = "tenant-abc"
		revisionID = "rev-xyz"
		docxKey    = "tenants/tenant-abc/revisions/rev-xyz/frozen.docx"
		outputKey  = "tenants/tenant-abc/revisions/rev-xyz/final.pdf"
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms/libreoffice/convert" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.7 rendered"))
	}))
	defer srv.Close()

	store := &fakePDFStore{objects: map[string][]byte{docxKey: []byte("PK fake docx")}}
	converter := newGotenbergPDFConverter(t, srv.URL, store)
	persister := &fakePDFPersister{}
	runner := NewPDFJobRunner(converter, persister)

	event := makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       tenantID,
		RevisionID:     revisionID,
		FinalDocxS3Key: docxKey,
	})
	if err := runner.Handle(context.Background(), event); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(persister.calls) != 1 {
		t.Fatalf("WritePDF called %d times, want 1", len(persister.calls))
	}
	call := persister.calls[0]
	if call.tenant != tenantID {
		t.Errorf("tenant = %q, want %q", call.tenant, tenantID)
	}
	if call.docID != revisionID {
		t.Errorf("docID = %q, want %q", call.docID, revisionID)
	}
	if call.s3Key != outputKey {
		t.Errorf("s3Key = %q, want %q", call.s3Key, outputKey)
	}
	if len(call.pdfHash) == 0 {
		t.Error("pdfHash empty")
	}
	if _, ok := store.saved[outputKey]; !ok {
		t.Errorf("PDF was not saved to %q", outputKey)
	}
}

func TestPDFPipeline_RealClient_GotenbergError(t *testing.T) {
	const docxKey = "some/key.docx"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	store := &fakePDFStore{objects: map[string][]byte{docxKey: []byte("PK fake docx")}}
	converter := newGotenbergPDFConverter(t, srv.URL, store)
	persister := &fakePDFPersister{}
	runner := NewPDFJobRunner(converter, persister)

	err := runner.Handle(context.Background(), makePDFEvent(messaging.PDFConvertPayload{
		TenantID:       "t1",
		RevisionID:     "r1",
		FinalDocxS3Key: docxKey,
	}))
	if err == nil {
		t.Fatal("expected error from gotenberg, got nil")
	}
	if len(persister.calls) != 0 {
		t.Error("WritePDF must not be called when convert fails")
	}
}

// TestPDFPipeline_WorkerLoop verifies the full Service.RunOnce -> PDFJobRunner ->
// GotenbergPDFClient -> WritePDF chain using a fake Gotenberg server.
func TestPDFPipeline_WorkerLoop(t *testing.T) {
	const (
		tenantID   = "t-loop"
		revisionID = "r-loop"
		docxKey    = "tenants/t-loop/revisions/r-loop/frozen.docx"
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.7 loop"))
	}))
	defer srv.Close()

	store := &fakePDFStore{objects: map[string][]byte{docxKey: []byte("PK fake docx")}}
	consumer := &fakeConsumer{events: []messaging.Event{
		makePDFEvent(messaging.PDFConvertPayload{
			TenantID:       tenantID,
			RevisionID:     revisionID,
			FinalDocxS3Key: docxKey,
		}),
	}}
	converter := newGotenbergPDFConverter(t, srv.URL, store)
	persister := &fakePDFPersister{}
	runner := NewPDFJobRunner(converter, persister)

	cfg := config.WorkerConfig{MaxAttempts: 3, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg).WithPDFRunner(runner)

	if err := svc.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if len(persister.calls) != 1 {
		t.Errorf("WritePDF called %d times, want 1", len(persister.calls))
	}
}
