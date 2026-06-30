package approvalhttp

// idempotency_middleware_test.go — contract tests proving that the 5 approval
// state-transition handlers (publish, schedule-publish, supersede, obsolete,
// cancel-instance, cancel-by-document) enforce the Idempotency-Key header via
// the platform idempotency.Require() middleware.
//
// F-CD1: missing key → 400 problem+json (middleware, not handler)
// F-CD3: malformed UUID → 400 invalid (middleware, not handler)
//
// Replay (F-CD2) requires a live Postgres store; those proofs live in the
// integration suite (internal/platform/idempotency/middleware_test.go and the
// approval HTTP integration tests). The unit cases here cover the middleware
// header-validation path, which fires before any DB call.

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// idempNoopDB returns a *sql.DB backed by a driver that satisfies
// sql.Open without panicking. The idempotency middleware's missing-key and
// malformed-key checks fire before any DB call, so this driver is never
// actually queried in those test cases.
var idempDBCounter int

type idempNoopConn struct{}

func (c *idempNoopConn) Prepare(_ string) (driver.Stmt, error) { return nil, io.EOF }
func (c *idempNoopConn) Close() error                          { return nil }
func (c *idempNoopConn) Begin() (driver.Tx, error)             { return nil, io.EOF }

type idempNoopDriver struct{}

func (d *idempNoopDriver) Open(_ string) (driver.Conn, error) { return &idempNoopConn{}, nil }

func newIdempNoopDB(t *testing.T) *sql.DB {
	t.Helper()
	idempDBCounter++
	name := fmt.Sprintf("idemp_noop_%d", idempDBCounter)
	sql.Register(name, &idempNoopDriver{})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open idemp noop db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// idempTestMux builds a mux that routes through h.idempotent() for the given
// route, exactly as router.go does in production.
func idempTestMux(h *Handler, method, pattern string, inner http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(method+" "+pattern, h.idempotent(method+" "+pattern, inner))
	return mux
}

func withTenantAndActor(r *http.Request) *http.Request {
	r = r.WithContext(tenant.WithTenantID(r.Context(), "tenant-1"))
	r = r.WithContext(iamdomain.WithAuthContext(r.Context(), "actor-1", []iamdomain.Role{}))
	return r
}

// expectProblemJSON asserts the response is an RFC 9457 problem+json with
// the given HTTP status and problem code.
func expectProblemJSON(t *testing.T, rr *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if rr.Code != wantStatus {
		t.Fatalf("status: got %d, want %d (body: %s)", rr.Code, wantStatus, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("Content-Type: got %q, want application/problem+json", ct)
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != wantCode {
		t.Fatalf("code: got %q, want %q", body.Code, wantCode)
	}
}

// TestIdempotencyMiddleware_MissingKey proves that each of the 5 state-transition
// routes returns 400 idempotency.key.required when the Idempotency-Key header is
// absent. This is F-CD1: the spec now declares the header required, and the
// middleware enforces it before the handler runs.
func TestIdempotencyMiddleware_MissingKey(t *testing.T) {
	db := newIdempNoopDB(t)
	h := &Handler{db: db}

	routes := []struct {
		method  string
		pattern string
		path    string
		body    string
	}{
		{"POST", "/api/v1/documents/{id}/publish", "/api/v1/documents/doc-1/publish", ""},
		{"POST", "/api/v1/documents/{id}/schedule-publish", "/api/v1/documents/doc-1/schedule-publish", `{"effective_from":"2026-05-01T12:00:00Z"}`},
		{"POST", "/api/v1/documents/{id}/supersede", "/api/v1/documents/doc-1/supersede", `{"superseded_document_id":"11111111-1111-1111-1111-111111111111"}`},
		{"POST", "/api/v1/documents/{id}/obsolete", "/api/v1/documents/doc-1/obsolete", `{"reason":"end of life"}`},
		{"POST", "/api/v1/approval/instances/{instance_id}/cancel", "/api/v1/approval/instances/inst-1/cancel", `{"reason":"withdrawn"}`},
		{"POST", "/api/v1/documents/{id}/cancel", "/api/v1/documents/doc-1/cancel", `{"reason":"withdrawn"}`},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.pattern, func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				t.Error("inner handler must not be called when Idempotency-Key is missing")
				w.WriteHeader(http.StatusOK)
			})
			mux := idempTestMux(h, rt.method, rt.pattern, inner)

			var reqBody io.Reader
			if rt.body != "" {
				reqBody = strings.NewReader(rt.body)
			}
			req := httptest.NewRequest(rt.method, rt.path, reqBody)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("If-Match", `"v1"`)
			req = withTenantAndActor(req)
			// No Idempotency-Key header.

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			expectProblemJSON(t, rr, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED")
		})
	}
}

// TestIdempotencyMiddleware_MalformedKey proves that each of the 5 state-transition
// routes returns 400 idempotency.key.invalid when the Idempotency-Key header is
// present but is not a UUID. This is F-CD3: uniform UUID-format validation across
// all mutating approval routes, matching the template routes.
func TestIdempotencyMiddleware_MalformedKey(t *testing.T) {
	db := newIdempNoopDB(t)
	h := &Handler{db: db}

	routes := []struct {
		method  string
		pattern string
		path    string
		body    string
	}{
		{"POST", "/api/v1/documents/{id}/publish", "/api/v1/documents/doc-1/publish", ""},
		{"POST", "/api/v1/documents/{id}/schedule-publish", "/api/v1/documents/doc-1/schedule-publish", `{"effective_from":"2026-05-01T12:00:00Z"}`},
		{"POST", "/api/v1/documents/{id}/supersede", "/api/v1/documents/doc-1/supersede", `{"superseded_document_id":"11111111-1111-1111-1111-111111111111"}`},
		{"POST", "/api/v1/documents/{id}/obsolete", "/api/v1/documents/doc-1/obsolete", `{"reason":"end of life"}`},
		{"POST", "/api/v1/approval/instances/{instance_id}/cancel", "/api/v1/approval/instances/inst-1/cancel", `{"reason":"withdrawn"}`},
		{"POST", "/api/v1/documents/{id}/cancel", "/api/v1/documents/doc-1/cancel", `{"reason":"withdrawn"}`},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.pattern, func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				t.Error("inner handler must not be called when Idempotency-Key is malformed")
				w.WriteHeader(http.StatusOK)
			})
			mux := idempTestMux(h, rt.method, rt.pattern, inner)

			var reqBody io.Reader
			if rt.body != "" {
				reqBody = strings.NewReader(rt.body)
			}
			req := httptest.NewRequest(rt.method, rt.path, reqBody)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("If-Match", `"v1"`)
			req.Header.Set("Idempotency-Key", "not-a-uuid")
			req = withTenantAndActor(req)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			expectProblemJSON(t, rr, http.StatusBadRequest, "IDEMPOTENCY_KEY_INVALID")
		})
	}
}
