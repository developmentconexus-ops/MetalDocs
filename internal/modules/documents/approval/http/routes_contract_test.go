package approvalhttp

// routes_contract_test.go — CON-03 regression guard for the generated-router
// migration. Adapted from the templates sibling
// (internal/modules/templates/delivery/http/routes_contract_test.go
// TestGeneratedTemplatesRoutes_IdempotencyStoreEngaged): proves the
// Middlewares closure in router.go actually routes idempotent approval
// mutations through internal/platform/idempotency.Store — not just that the
// generated wrapper enforces Idempotency-Key header presence/format (the
// wrapper validates the header BEFORE Middlewares run, so a missing or
// malformed key can never reach the store and cannot serve as a
// discriminator; see idempotency_middleware_test.go for that coverage).
//
// The discriminator here is the scripted hash-conflict: 422
// IDEMPOTENCY_KEY_REUSED is only producible by idempotency.Store.BeginReplay.
// This is also the regression guard for the r.Pattern key bug (r.Pattern
// already includes the method prefix — prepending r.Method again would make
// every lookup miss and silently disable the store on all 7 idempotent
// approval routes while presence-only tests kept passing).

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// newConflictMockDB returns a *sql.DB scripted for exactly one idempotency
// BeginReplay that ends in ErrConflict: BeginTx → INSERT…ON CONFLICT DO
// NOTHING returns no row (loser, sql.ErrNoRows) → SELECT…FOR UPDATE returns a
// stored payload_hash that cannot match the request's → BeginReplay commits
// and returns ErrConflict → middleware writes 422 IDEMPOTENCY_KEY_REUSED.
// The INSERT expectation pins arg 3 to the exact route template the
// Middlewares closure must construct from r.Pattern, so a wrong lookup key
// (e.g. re-prepending r.Method to r.Pattern) fails the test twice over: the
// store is never engaged (no 422) and the args never match.
func newConflictMockDB(t *testing.T, routeTemplate string) *sql.DB {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("newConflictMockDB: sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO metaldocs\.idempotency_keys`).
		WithArgs("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "user-a", routeTemplate, "33333333-3333-4333-8333-333333333333", sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT status, payload_hash`).
		WithArgs("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "user-a", routeTemplate, "33333333-3333-4333-8333-333333333333").
		WillReturnRows(sqlmock.NewRows([]string{"status", "payload_hash", "response_status", "response_body", "expired"}).
			AddRow("completed", "hash-that-cannot-match-any-request", nil, nil, false))
	mock.ExpectCommit()
	return mockDB
}

// TestGeneratedApprovalRoutes_IdempotencyStoreEngaged is the CON-03 store
// engagement guard (task requirement 4). It exercises one idempotent approval
// route — POST /api/v1/documents/{id}/publish — end to end through
// RegisterRoutes so the request traverses the real generated
// ServerInterfaceWrapper, then router.go's Middlewares closure, then
// h.idempotentHandler, then idempotency.Require/BeginReplay. The handler body
// (PublishHandler) is never reached: BeginReplay's scripted conflict returns
// before next.ServeHTTP runs, so a bare &Handler{db: db} (no application
// services wired) is sufficient — exactly as idempotency_middleware_test.go's
// TestIdempotencyMiddleware_MissingKey/MalformedKey already do for the
// header-validation path.
func TestGeneratedApprovalRoutes_IdempotencyStoreEngaged(t *testing.T) {
	const routeTemplate = "POST /api/v1/documents/{id}/publish"

	db := newConflictMockDB(t, routeTemplate)
	h := &Handler{db: db}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/11111111-1111-1111-1111-111111111111/publish", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "user-a", []iamdomain.Role{}))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 from the scripted idempotency conflict, got %d body=%s", rr.Code, rr.Body.String())
	}
	if body := rr.Body.String(); !bytes.Contains([]byte(body), []byte("IDEMPOTENCY_KEY_REUSED")) {
		t.Fatalf("expected IDEMPOTENCY_KEY_REUSED from the idempotency store path, got body=%s", body)
	}
}

// TestGeneratedApprovalRoutes_NonIdempotentRouteSkipsStore proves the inverse:
// a spec route that is NOT in router.go's idempotentRoutes set must not be
// routed through the store even when it carries a well-formed Idempotency-Key
// (signoff and route-admin routes validate/replay idempotency themselves
// in-handler against a different store — see router.go's idempotentRoutes doc
// comment). With the same conflict script and a valid key, a wrongly-wrapped
// route would return exactly 422 IDEMPOTENCY_KEY_REUSED; any other outcome
// means the generic store was correctly skipped.
func TestGeneratedApprovalRoutes_NonIdempotentRouteSkipsStore(t *testing.T) {
	const routeTemplate = "POST /api/v1/approval/routes"

	db := newConflictMockDB(t, routeTemplate)
	h := &Handler{db: db}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/routes", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "user-a", []iamdomain.Role{}))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if body := rr.Body.String(); bytes.Contains([]byte(body), []byte("IDEMPOTENCY_KEY_REUSED")) {
		t.Fatalf("non-idempotent route unexpectedly went through the generic idempotency store: status=%d body=%s", rr.Code, body)
	}
}
