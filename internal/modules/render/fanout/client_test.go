package fanout

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestClient_Fanout_Success(t *testing.T) {
	var gotPath, gotContentType, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("content-type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content_hash":      "deadbeef",
			"final_docx_s3_key": "out/a.docx",
			"unreplaced_vars":   []string{"missing_var"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "", srv.Client())
	resp, err := c.Fanout(context.Background(), FanoutRequest{
		TenantID:          "t1",
		RevisionID:        "r1",
		BodyDocxS3Key:     "templates/x.docx",
		PlaceholderValues: map[string]string{"doc_code": "ABC"},
		Composition:       json.RawMessage(`{}`),
		ResolvedValues:    map[string]any{"doc_code": "ABC"},
	})
	if err != nil {
		t.Fatalf("fanout err: %v", err)
	}
	if gotPath != "/render/fanout" {
		t.Errorf("path = %q, want /render/fanout", gotPath)
	}
	if gotContentType != "application/json" {
		t.Errorf("content-type = %q", gotContentType)
	}
	if !strings.Contains(gotBody, `"tenant_id":"t1"`) {
		t.Errorf("body missing tenant_id: %s", gotBody)
	}
	if resp.ContentHash != "deadbeef" || resp.FinalDocxS3Key != "out/a.docx" {
		t.Errorf("resp = %+v", resp)
	}
	if len(resp.UnreplacedVars) != 1 || resp.UnreplacedVars[0] != "missing_var" {
		t.Errorf("unreplaced_vars = %v", resp.UnreplacedVars)
	}
}

func TestClient_Fanout_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "", srv.Client())
	_, err := c.Fanout(context.Background(), FanoutRequest{})
	if err == nil {
		t.Fatal("expected error on non-200")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("err should mention status: %v", err)
	}
}

func TestFanout_ClassifiesTemplateDefectAsNonRetryable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"render_failed","kind":"undefined_variable","message":"no var","variable":"foo"}`))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, "", srv.Client()).Fanout(context.Background(), FanoutRequest{})
	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("want *RenderError, got %T (%v)", err, err)
	}
	if re.Kind != "undefined_variable" || re.Variable != "foo" || re.Status != http.StatusBadRequest {
		t.Fatalf("unexpected classification: %+v", re)
	}
	if re.Retryable() {
		t.Fatalf("template defect must not be retryable")
	}
}

// TestClient_Fanout_InjectsTraceparent_WhenSpanActive proves the worker→
// renderer HTTP hop carries the W3C traceparent header when the caller's
// context holds an active span (REQ-OBS-3, T-010 closure). Uses a real SDK
// TracerProvider + in-memory recorder (tracetest) so the propagator has a
// real, non-zero SpanContext to inject — otel.GetTextMapPropagator() is the
// same composite propagator SetupOTel installs process-wide.
func TestClient_Fanout_InjectsTraceparent_WhenSpanActive(t *testing.T) {
	prevPropagator := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))
	defer otel.SetTextMapPropagator(prevPropagator)

	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	var gotTraceparent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("traceparent")
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content_hash":      "deadbeef",
			"final_docx_s3_key": "out/a.docx",
			"unreplaced_vars":   []string{},
		})
	}))
	defer srv.Close()

	ctx, span := tp.Tracer("test").Start(context.Background(), "fanout-test-span")
	defer span.End()

	c := NewClient(srv.URL, "", srv.Client())
	if _, err := c.Fanout(ctx, FanoutRequest{}); err != nil {
		t.Fatalf("fanout err: %v", err)
	}

	if gotTraceparent == "" {
		t.Fatal("expected a traceparent header on the outbound request, got none")
	}
	sc := span.SpanContext()
	if !strings.Contains(gotTraceparent, sc.TraceID().String()) {
		t.Errorf("traceparent %q does not carry the active trace id %q", gotTraceparent, sc.TraceID().String())
	}
}

// TestClient_Fanout_NoTraceparent_WhenPropagatorUnset proves the inject call
// is a harmless no-op when OTel is not installed (unconfigured default): the
// global propagator stays the OTel no-op, so Inject adds no header and the
// request proceeds unchanged — mirrors SetupOTel's inert-by-default contract.
func TestClient_Fanout_NoTraceparent_WhenPropagatorUnset(t *testing.T) {
	prevPropagator := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator()) // no registered propagators
	defer otel.SetTextMapPropagator(prevPropagator)

	var gotTraceparent string
	sawHeader := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent, sawHeader = r.Header.Get("traceparent"), r.Header.Get("traceparent") != ""
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content_hash":      "deadbeef",
			"final_docx_s3_key": "out/a.docx",
			"unreplaced_vars":   []string{},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "", srv.Client())
	if _, err := c.Fanout(context.Background(), FanoutRequest{}); err != nil {
		t.Fatalf("fanout err: %v", err)
	}
	if sawHeader {
		t.Errorf("expected no traceparent header with an empty propagator, got %q", gotTraceparent)
	}
}

func TestFanout_ClassifiesUnknownAsRetryable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"render_failed","kind":"unknown","message":"boom"}`))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, "", srv.Client()).Fanout(context.Background(), FanoutRequest{})
	var re *RenderError
	if !errors.As(err, &re) {
		t.Fatalf("want *RenderError, got %T", err)
	}
	if !re.Retryable() {
		t.Fatalf("unknown/5xx must be retryable")
	}
}
