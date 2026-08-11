package application_test

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/audit/application"
	"metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/audit/infrastructure/memory"
)

type captureReader struct {
	query domain.ListEventsQuery
}

func (r *captureReader) ListEvents(_ context.Context, query domain.ListEventsQuery) ([]domain.Event, bool, error) {
	r.query = query
	return nil, false, nil
}

func TestNewServicePanicsWithoutReader(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil reader")
		}
	}()

	_ = application.NewService(nil)
}

func TestListEventsFailsWhenServiceReaderMissing(t *testing.T) {
	t.Parallel()

	service := &application.Service{}

	_, _, err := service.ListEvents(context.Background(), domain.ListEventsQuery{TenantID: "tenant-a"})

	if !errors.Is(err, application.ErrReaderRequired) {
		t.Fatalf("expected ErrReaderRequired, got %v", err)
	}
}

func TestListEventsRequiresTenantID(t *testing.T) {
	t.Parallel()

	service := application.NewService(&captureReader{})

	_, _, err := service.ListEvents(t.Context(), domain.ListEventsQuery{})

	if !errors.Is(err, application.ErrTenantRequired) {
		t.Fatalf("expected ErrTenantRequired, got %v", err)
	}
}

func TestListEventsPreservesTenantIDDuringNormalization(t *testing.T) {
	t.Parallel()

	reader := &captureReader{}
	service := application.NewService(reader)

	_, _, err := service.ListEvents(t.Context(), domain.ListEventsQuery{
		ResourceType: " document ",
		ResourceID:   " doc-1 ",
		TenantID:     " tenant-a ",
	})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}

	if reader.query.TenantID != "tenant-a" {
		t.Fatalf("TenantID = %q, want tenant-a", reader.query.TenantID)
	}
	if reader.query.ResourceType != "document" {
		t.Fatalf("ResourceType = %q, want document", reader.query.ResourceType)
	}
	if reader.query.ResourceID != "doc-1" {
		t.Fatalf("ResourceID = %q, want doc-1", reader.query.ResourceID)
	}
	if reader.query.Limit != 50 {
		t.Fatalf("Limit = %d, want 50", reader.query.Limit)
	}
}

func TestGetExportStatus_RequiresActorID(t *testing.T) {
	t.Parallel()

	w := memory.NewWriter()
	exports := memory.NewExportJobRepository()
	svc := application.NewService(&captureReader{}).WithExports(w, exports, w, func(domain.ExportJob) string { return "" })
	_, err := svc.GetExportStatus(context.Background(), "tenant-a", "   ", "export-1")
	if !errors.Is(err, application.ErrActorRequired) {
		t.Fatalf("expected ErrActorRequired for empty actorID, got %v", err)
	}
}

// exportTestTenantID is a syntactically valid UUID — buildExportJob parses
// TenantID with uuid.Parse, unlike ListEvents which treats it as an opaque
// string, so the ExportEvents tests below need a real UUID, not "tenant-a".
const exportTestTenantID = "11111111-1111-1111-1111-111111111111"

func TestExportEvents_RefusesPersistWhenTenantErasedBeforeSave(t *testing.T) {
	t.Parallel()

	w := memory.NewWriter()
	if err := w.Record(context.Background(), mustEvent(t, exportTestTenantID, "actor-1")); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	exports := memory.NewExportJobRepository()
	w.SetErased(exportTestTenantID, true)

	svc := application.NewService(w).
		WithExports(w, exports, w, func(domain.ExportJob) string { return "" })

	_, err := svc.ExportEvents(context.Background(), "actor-1", domain.ExportFormatCSV, domain.ListEventsQuery{TenantID: exportTestTenantID})
	if !errors.Is(err, application.ErrExportTenantErased) {
		t.Fatalf("ExportEvents error = %v, want ErrExportTenantErased", err)
	}
	if got := w.ErasedCheckCalls(); got != 1 {
		t.Fatalf("erasure checks = %d, want exactly 1 (re-checked once, immediately before persist)", got)
	}
	if got := exports.Len(); got != 0 {
		t.Fatalf("export rows persisted = %d, want 0 — erasure committed before Save, nothing should have been written", got)
	}
}

func TestExportEvents_RefusesPersistWhenErasureCheckErrors(t *testing.T) {
	t.Parallel()

	w := memory.NewWriter()
	if err := w.Record(context.Background(), mustEvent(t, exportTestTenantID, "actor-1")); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	exports := memory.NewExportJobRepository()
	w.SetErasedCheckError(errors.New("boom: erasure lookup unavailable"))

	svc := application.NewService(w).
		WithExports(w, exports, w, func(domain.ExportJob) string { return "" })

	_, err := svc.ExportEvents(context.Background(), "actor-1", domain.ExportFormatCSV, domain.ListEventsQuery{TenantID: exportTestTenantID})
	if !errors.Is(err, application.ErrExportTenantErased) {
		t.Fatalf("ExportEvents error = %v, want ErrExportTenantErased (fail closed on lookup error)", err)
	}
	if got := exports.Len(); got != 0 {
		t.Fatalf("export rows persisted = %d, want 0 — an erasure-check error must fail closed, not persist", got)
	}
}

func TestExportEvents_PersistsWhenTenantActive(t *testing.T) {
	t.Parallel()

	w := memory.NewWriter()
	if err := w.Record(context.Background(), mustEvent(t, exportTestTenantID, "actor-1")); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	exports := memory.NewExportJobRepository()

	svc := application.NewService(w).
		WithExports(w, exports, w, func(domain.ExportJob) string { return "" })

	job, err := svc.ExportEvents(context.Background(), "actor-1", domain.ExportFormatCSV, domain.ListEventsQuery{TenantID: exportTestTenantID})
	if err != nil {
		t.Fatalf("ExportEvents: %v", err)
	}
	if got := exports.Len(); got != 1 {
		t.Fatalf("export rows persisted = %d, want 1 for an active tenant (over-block guard)", got)
	}
	if job.ID == "" {
		t.Fatal("job.ID empty, want a persisted export job")
	}
}

// TestWithExports_NoSeparateErasureSetterToOmit documents, at the type
// level, that WithExports is the only place erasure gets wired: nothing in
// this package exposes a WithErasureCheck (or similar) setter a composition
// root or test harness could forget to call. The guarantee is structural
// (domain.Writer embeds ErasureChecker; WithExports assigns s.erasure from
// the same writer parameter it already requires non-nil), not conventional —
// see refuseIfTenantErased's doc comment and PR #121 review round 1, P1.
func TestWithExports_NoSeparateErasureSetterToOmit(t *testing.T) {
	t.Parallel()

	w := memory.NewWriter()
	exports := memory.NewExportJobRepository()
	svc := application.NewService(&captureReader{}).
		WithExports(w, exports, w, func(domain.ExportJob) string { return "" })
	if err := w.Record(context.Background(), mustEvent(t, exportTestTenantID, "actor-1")); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	w.SetErased(exportTestTenantID, true)

	_, err := svc.ExportEvents(context.Background(), "actor-1", domain.ExportFormatCSV, domain.ListEventsQuery{TenantID: exportTestTenantID})
	if !errors.Is(err, application.ErrExportTenantErased) {
		t.Fatalf("ExportEvents error = %v, want ErrExportTenantErased — a Service built via NewService+WithExports alone, with no erasure setter ever called, must still enforce the gate", err)
	}
}

func mustEvent(t *testing.T, tenantID, actorID string) domain.Event {
	t.Helper()
	event, err := domain.NewEvent(tenantID, "document", "doc-1", actorID, "document.viewed", map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("NewEvent: %v", err)
	}
	return event
}

func TestWithExportsPanicsOnNilDependency(t *testing.T) {
	t.Parallel()

	w := memory.NewWriter()
	exports := memory.NewExportJobRepository()
	urlBuilder := func(domain.ExportJob) string { return "" }

	cases := []struct {
		name    string
		counter domain.Counter
		repo    domain.ExportJobRepository
		writer  domain.Writer
		urlBld  application.SignedURLBuilder
	}{
		{"nil counter", nil, exports, w, urlBuilder},
		{"nil repo", w, nil, w, urlBuilder},
		{"nil writer", w, exports, nil, urlBuilder},
		{"nil urlBuilder", w, exports, w, nil},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for %s", tc.name)
				}
			}()
			application.NewService(&captureReader{}).WithExports(tc.counter, tc.repo, tc.writer, tc.urlBld)
		})
	}
}
