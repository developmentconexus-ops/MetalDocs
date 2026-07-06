package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

// TestHTTPObservability_Wrap_SetsTraceIDResponseHeader closes the
// log<->trace<->response-header correlation triple (see http.go Wrap):
// today X-Trace-Id is only ever READ from the inbound request/context, never
// written to the outbound response, so a client has no way to correlate its
// request to server logs/traces. Wrap must set X-Trace-Id on the response
// before invoking the downstream handler.
func TestHTTPObservability_Wrap_SetsTraceIDResponseHeader(t *testing.T) {
	o := observability.NewHTTPObservability(nil, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/whoami", nil)
	rec := httptest.NewRecorder()
	o.Wrap(next).ServeHTTP(rec, req)

	got := rec.Header().Get("X-Trace-Id")
	if got == "" {
		t.Fatalf("expected non-empty X-Trace-Id response header, got empty")
	}
}

// TestHTTPObservability_Wrap_EchoesInboundTraceIDOnResponse confirms that when
// the caller supplies a valid inbound X-Trace-Id, the resolved id (echoed per
// the Wrap resolution order) is propagated back out on the response header
// unchanged, so a client-supplied trace id round-trips end to end.
func TestHTTPObservability_Wrap_EchoesInboundTraceIDOnResponse(t *testing.T) {
	o := observability.NewHTTPObservability(nil, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	const inbound = "client-supplied-trace-id-123"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/whoami", nil)
	req.Header.Set("X-Trace-Id", inbound)
	rec := httptest.NewRecorder()
	o.Wrap(next).ServeHTTP(rec, req)

	got := rec.Header().Get("X-Trace-Id")
	if got != inbound {
		t.Fatalf("X-Trace-Id response header: got %q want %q", got, inbound)
	}
}
