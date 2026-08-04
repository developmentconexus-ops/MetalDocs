package approvalhttp

// idempotency_middleware_test.go — contract tests proving that EVERY approval
// route the router wraps in the platform idempotency.Require() middleware
// enforces the Idempotency-Key header.
//
// F-CD1: missing key → 400 problem+json (middleware, not handler)
// F-CD3: malformed UUID → 400 invalid (middleware, not handler)
//
// The route table is DERIVED from router.go's `idempotentRoutes` (see
// idempotentRouteCases below), never restated here. A hand-copied enumeration
// goes stale silently: this file used to list POST /documents/{id}/publish,
// /schedule-publish and /supersede, and kept passing long after ADR 0085 Stage
// B deleted those endpoints, because h.idempotentHandler wraps whatever pattern
// a test hands it. Deriving the table means a route added to or removed from
// the middleware set changes these tests automatically.
//
// Replay (F-CD2) requires a live Postgres store; those proofs live in the
// integration suite (internal/platform/idempotency/middleware_test.go and the
// approval HTTP integration tests). The unit cases here cover the middleware
// header-validation path, which fires before any DB call — and before the
// request body is read, which is why no case sends one.

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
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

// idempTestMux builds a mux that routes through h.idempotentHandler() for the
// given route, exactly as router.go's Middlewares closure does in production.
func idempTestMux(h *Handler, method, pattern string, inner http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(method+" "+pattern, h.idempotentHandler(method+" "+pattern, inner))
	return mux
}

// idempotentRouteCase is one derived row: the production route template plus a
// concrete path that matches it.
type idempotentRouteCase struct {
	method  string
	pattern string
	path    string
}

// idempotentRouteCases turns router.go's production `idempotentRoutes` set into
// the test table. Path params are substituted with literals; an unrecognised
// param fails loudly rather than silently producing a request that matches no
// pattern (which would 404 and turn a real gap into a green test).
func idempotentRouteCases(t *testing.T) []idempotentRouteCase {
	t.Helper()

	if len(idempotentRoutes) == 0 {
		t.Fatal("idempotentRoutes is empty — the approval router wraps no route in the idempotency middleware")
	}

	cases := make([]idempotentRouteCase, 0, len(idempotentRoutes))
	for route := range idempotentRoutes {
		method, pattern, ok := strings.Cut(route, " ")
		if !ok {
			t.Fatalf("idempotentRoutes key %q is not a method-qualified pattern", route)
		}
		path := strings.ReplaceAll(pattern, "{id}", "doc-1")
		path = strings.ReplaceAll(path, "{instance_id}", "inst-1")
		if strings.ContainsAny(path, "{}") {
			t.Fatalf("unsubstituted path parameter in %q — add the substitution here", pattern)
		}
		cases = append(cases, idempotentRouteCase{method: method, pattern: pattern, path: path})
	}
	sort.Slice(cases, func(i, j int) bool {
		if cases[i].pattern != cases[j].pattern {
			return cases[i].pattern < cases[j].pattern
		}
		return cases[i].method < cases[j].method
	})
	return cases
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

// TestIdempotencyMiddleware_MissingKey proves that every route in the
// production idempotency set returns 400 idempotency.key.required when the
// Idempotency-Key header is absent. This is F-CD1: the spec declares the header
// required, and the middleware enforces it before the handler runs.
func TestIdempotencyMiddleware_MissingKey(t *testing.T) {
	db := newIdempNoopDB(t)
	h := &Handler{db: db}

	for _, rt := range idempotentRouteCases(t) {
		t.Run(rt.method+" "+rt.pattern, func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				t.Error("inner handler must not be called when Idempotency-Key is missing")
				w.WriteHeader(http.StatusOK)
			})
			mux := idempTestMux(h, rt.method, rt.pattern, inner)

			req := httptest.NewRequest(rt.method, rt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("If-Match", `"v1"`)
			req = withTenantAndActor(req)
			// No Idempotency-Key header.

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			expectProblemJSON(t, rr, http.StatusBadRequest, "request.idempotency_key_required")
		})
	}
}

// TestIdempotencyMiddleware_MalformedKey proves that every route in the
// production idempotency set returns 400 idempotency.key.invalid when the
// Idempotency-Key header is present but is not a UUID. This is F-CD3: uniform
// UUID-format validation across all mutating approval routes, matching the
// template routes.
func TestIdempotencyMiddleware_MalformedKey(t *testing.T) {
	db := newIdempNoopDB(t)
	h := &Handler{db: db}

	for _, rt := range idempotentRouteCases(t) {
		t.Run(rt.method+" "+rt.pattern, func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				t.Error("inner handler must not be called when Idempotency-Key is malformed")
				w.WriteHeader(http.StatusOK)
			})
			mux := idempTestMux(h, rt.method, rt.pattern, inner)

			req := httptest.NewRequest(rt.method, rt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("If-Match", `"v1"`)
			req.Header.Set("Idempotency-Key", "not-a-uuid")
			req = withTenantAndActor(req)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			expectProblemJSON(t, rr, http.StatusBadRequest, "request.idempotency_key_invalid")
		})
	}
}
