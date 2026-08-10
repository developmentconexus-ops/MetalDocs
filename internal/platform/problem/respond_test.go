package problem

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/platform/requesttrace"
)

// captureLogs swaps the default slog logger for the duration of the test and
// returns the captured records. Not parallel-safe by construction: slog's
// default logger is process-global, so these tests must not call t.Parallel.
func captureLogs(t *testing.T, level slog.Level) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

func requestWithTrace(traceID string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/documents", nil)
	return r.WithContext(requesttrace.WithTraceID(r.Context(), traceID))
}

func TestRespond_4xx_WritesRFC9457AndDoesNotLog(t *testing.T) {
	buf := captureLogs(t, slog.LevelDebug)
	rr := httptest.NewRecorder()

	Respond(rr, requestWithTrace("trace-4xx"), NewFor(CodeRequestInvalid, "Validation failed").WithDetail("bad payload"))

	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("content-type = %q, want application/problem+json", got)
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	var got Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not a problem document: %v (%s)", err, rr.Body.String())
	}
	if got.Code != CodeRequestInvalid || got.Title != "Validation failed" || got.Detail != "bad payload" {
		t.Fatalf("unexpected body: %+v", got)
	}

	// Policy: HTTPObservability already logs one access line per request, so a
	// causeless 4xx must NOT produce a second line here.
	if buf.Len() != 0 {
		t.Fatalf("causeless 4xx logged: %s", buf.String())
	}
}

func TestRespond_4xxWithCause_LogsAtDebugAndKeepsCauseOffTheWire(t *testing.T) {
	buf := captureLogs(t, slog.LevelDebug)
	rr := httptest.NewRecorder()

	RespondCause(rr, requestWithTrace("trace-cause"),
		NewFor(CodeRequestInvalid, "Validation failed"),
		errors.New("secret-internal-detail"))

	if strings.Contains(rr.Body.String(), "secret-internal-detail") {
		t.Fatalf("cause leaked into the response body: %s", rr.Body.String())
	}
	line := buf.String()
	if !strings.Contains(line, `"level":"DEBUG"`) {
		t.Fatalf("4xx-with-cause should log at DEBUG, got: %s", line)
	}
	if !strings.Contains(line, "secret-internal-detail") || !strings.Contains(line, "trace-cause") {
		t.Fatalf("log line missing cause or trace id: %s", line)
	}
}

func TestRespond_5xx_LogsErrorWithTraceCorrelation(t *testing.T) {
	buf := captureLogs(t, slog.LevelDebug)
	rr := httptest.NewRecorder()

	RespondCause(rr, requestWithTrace("trace-5xx"),
		NewFor(CodeInternalUnknown, "Internal server error"),
		errors.New("pool exhausted"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	var rec map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec); err != nil {
		t.Fatalf("no structured log record emitted for a 5xx: %v (%s)", err, buf.String())
	}
	if rec["level"] != "ERROR" {
		t.Fatalf("5xx must log at ERROR, got %v", rec["level"])
	}
	for k, want := range map[string]any{
		"trace_id": "trace-5xx",
		"status":   float64(500),
		"code":     CodeInternalUnknown.String(),
		"method":   http.MethodPost,
		"path":     "/api/v1/documents",
		"err":      "pool exhausted",
	} {
		if rec[k] != want {
			t.Fatalf("log field %q = %v, want %v (record: %v)", k, rec[k], want, rec)
		}
	}
	// The generic client body must not carry the internal cause.
	if strings.Contains(rr.Body.String(), "pool exhausted") {
		t.Fatalf("internal cause leaked to client: %s", rr.Body.String())
	}
}

func TestRespond_NeverLogsRequestBodyOrHeaders(t *testing.T) {
	buf := captureLogs(t, slog.LevelDebug)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"password":"hunter2"}`))
	r.Header.Set("Cookie", "session=super-secret-token")
	r.Header.Set("Authorization", "Bearer super-secret-token")

	RespondCause(httptest.NewRecorder(), r, NewFor(CodeInternalUnknown, "Internal server error"),
		errors.New("boom"))

	if strings.Contains(buf.String(), "hunter2") || strings.Contains(buf.String(), "super-secret-token") {
		t.Fatalf("writer logged request secrets: %s", buf.String())
	}
}

func TestRespond_WritesExactlyOnce(t *testing.T) {
	captureLogs(t, slog.LevelDebug)
	cw := &countingWriter{ResponseWriter: httptest.NewRecorder()}

	Respond(cw, requestWithTrace("trace-once"), NewFor(CodeNotFoundResource, "Not found"))

	if cw.headerCalls != 1 {
		t.Fatalf("WriteHeader called %d times, want 1", cw.headerCalls)
	}
	if cw.writeCalls != 1 {
		t.Fatalf("Write called %d times, want 1", cw.writeCalls)
	}
}

func TestRespond_NilProblem_FailsClosedWith500(t *testing.T) {
	buf := captureLogs(t, slog.LevelDebug)
	rr := httptest.NewRecorder()

	Respond(rr, requestWithTrace("trace-nil"), nil)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (never a bare 200)", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("content-type = %q", got)
	}
	if !strings.Contains(buf.String(), "nil problem") {
		t.Fatalf("nil problem was not reported: %s", buf.String())
	}
}

func TestRespond_NilRequest_StillEmitsProblem(t *testing.T) {
	captureLogs(t, slog.LevelDebug)
	rr := httptest.NewRecorder()

	Respond(rr, nil, NewFor(CodeNotFoundResource, "Not found"))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

type countingWriter struct {
	http.ResponseWriter
	headerCalls int
	writeCalls  int
}

func (w *countingWriter) WriteHeader(code int) {
	w.headerCalls++
	w.ResponseWriter.WriteHeader(code)
}

func (w *countingWriter) Write(b []byte) (int, error) {
	w.writeCalls++
	return w.ResponseWriter.Write(b)
}
