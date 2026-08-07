package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"metaldocs/internal/modules/controlleddocuments/application"
	cddomain "metaldocs/internal/modules/controlleddocuments/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	platformdb "metaldocs/internal/platform/db"
)

// stubTxRunner aborts immediately without calling fn so downstream nil-deps are never reached.
type stubTxRunner struct{}

func (stubTxRunner) Do(_ context.Context, _ func(tx *sql.Tx) error) error {
	return errors.New("stub runner")
}

func (stubTxRunner) DoReadOnly(_ context.Context, _ func(tx *sql.Tx) error) error {
	return errors.New("stub runner")
}

// stubCDRepo satisfies cddomain.ControlledDocumentRepository.
type stubCDRepo struct{}

func (stubCDRepo) GetByID(_ context.Context, _, _ string) (*cddomain.ControlledDocument, error) {
	return nil, nil
}
func (stubCDRepo) GetByCode(_ context.Context, _, _, _ string) (*cddomain.ControlledDocument, error) {
	return nil, nil
}
func (stubCDRepo) CodeExists(_ context.Context, _, _, _ string) (bool, error) { return false, nil }
func (stubCDRepo) List(_ context.Context, _ string, _ cddomain.CDFilter) ([]cddomain.ControlledDocument, bool, error) {
	return nil, false, nil
}
func (stubCDRepo) CanRead(_ context.Context, _, _, _ string) (bool, error) { return false, nil }
func (stubCDRepo) GetActiveInstance(_ context.Context, _, _ string) (*cddomain.ActiveDocumentInstance, error) {
	return nil, nil
}
func (stubCDRepo) Create(_ context.Context, _ *cddomain.ControlledDocument) error { return nil }
func (stubCDRepo) CreateTx(_ context.Context, _ platformdb.Tx, _ *cddomain.ControlledDocument) error {
	return nil
}
func (stubCDRepo) UpdateStatus(_ context.Context, _, _ string, _ cddomain.CDStatus, _ time.Time) error {
	return nil
}
func (stubCDRepo) UpdateStatusTx(_ context.Context, _ platformdb.Tx, _, _ string, _ cddomain.CDStatus, _ time.Time) error {
	return nil
}

// stubSeq satisfies cddomain.SequenceAllocator.
type stubSeq struct{}

func (stubSeq) NextAndIncrement(_ context.Context, _ platformdb.Tx, _, _, _ string) (int, error) {
	return 0, nil
}
func (stubSeq) Peek(_ context.Context, _, _, _ string) (int, error)   { return 0, nil }
func (stubSeq) EnsureCounter(_ context.Context, _, _, _ string) error { return nil }

// stubTplCheck satisfies application.TemplateVersionChecker.
type stubTplCheck struct{}

func (stubTplCheck) GetTemplateVersionState(_ context.Context, _, _ string) (*string, string, error) {
	return nil, "", nil
}

// stubProfileReader satisfies application.ProfileReader.
// Returns a non-nil active (ArchivedAt==nil) profile so Create passes the IsActive gate.
type stubProfileReader struct{}

func (stubProfileReader) GetByCode(_ context.Context, _, _ string) (*taxonomydomain.DocumentProfile, error) {
	return &taxonomydomain.DocumentProfile{}, nil
}

// stubAreaReader satisfies application.AreaReader.
// Returns a non-nil active (ArchivedAt==nil) area so Create passes the IsActive gate.
type stubAreaReader struct{}

func (stubAreaReader) GetByCode(_ context.Context, _, _ string) (*taxonomydomain.ProcessArea, error) {
	return &taxonomydomain.ProcessArea{}, nil
}

// stubGovLogger satisfies taxonomydomain.GovernanceLogger.
type stubGovLogger struct{}

func (stubGovLogger) Log(_ context.Context, _ taxonomydomain.GovernanceEvent) error { return nil }
func (stubGovLogger) LogTx(_ context.Context, _ platformdb.Tx, _ taxonomydomain.GovernanceEvent) error {
	return nil
}

func newTestCDService() *application.ControlledDocumentService {
	return application.NewControlledDocumentService(
		stubTxRunner{},
		stubCDRepo{},
		stubSeq{},
		stubTplCheck{},
		stubProfileReader{},
		stubAreaReader{},
		stubGovLogger{},
		nil, // docInit — optional, guarded by nil check not panic
	)
}

func TestCreate_EmitsCdCreateSpan(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	svc := newTestCDService()
	cmd := application.CreateControlledDocumentCmd{
		TenantID:    "tenant-1",
		ProfileCode: "SOP",
	}
	_, _ = svc.Create(context.Background(), cmd)

	spans := sr.Ended()
	var found bool
	for _, s := range spans {
		if s.Name() == "cd.create" {
			found = true
			for _, a := range s.Attributes() {
				if a.Key == attribute.Key("document.profile_code") {
					if a.Value.AsString() != cmd.ProfileCode {
						t.Errorf("document.profile_code: got %q, want %q", a.Value.AsString(), cmd.ProfileCode)
					}
				}
			}
			break
		}
	}
	if !found {
		names := make([]string, len(spans))
		for i, s := range spans {
			names[i] = s.Name()
		}
		t.Fatalf("want span named cd.create, got %v", names)
	}
}

func TestCreate_SpanStatusError_OnFailure(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	svc := newTestCDService()
	cmd := application.CreateControlledDocumentCmd{
		TenantID:    "tenant-1",
		ProfileCode: "SOP",
	}
	_, err := svc.Create(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected an error from stub runner, got nil")
	}

	spans := sr.Ended()
	for _, s := range spans {
		if s.Name() == "cd.create" {
			if s.Status().Code != codes.Error {
				t.Errorf("span status: got %v, want Error", s.Status().Code)
			}
			return
		}
	}
	t.Fatal("span cd.create not found")
}
