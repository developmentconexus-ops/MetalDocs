package http

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// newIdempotentMockDB returns a sqlmock *sql.DB that always wins the
// idempotency.Store.BeginReplay claim (metaldocs.idempotency_keys) for the
// three taxonomy create routes (CON-06(d)), so the wrapped handler runs
// normally and its own response is what the test observes. Per claimed
// request the store issues exactly: BEGIN, the claim INSERT...RETURNING
// (QueryRow), then either an UPDATE+COMMIT (CompleteReplay, 2xx) or a
// DELETE+COMMIT (FailReplay, non-2xx) — never a Rollback in either outcome
// (see internal/platform/idempotency/postgres_store.go CompleteReplay /
// FailReplay). Both outcomes share the same Exec-then-Commit shape, so one
// mocked expectation pair per iteration covers both, mirroring the
// controlleddocuments delivery/http newPermissiveMockDB pattern.
func newIdempotentMockDB(t *testing.T) *sql.DB {
	t.Helper()
	anyMatcher := sqlmock.QueryMatcherFunc(func(_, _ string) error { return nil })
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(anyMatcher))
	if err != nil {
		t.Fatalf("newIdempotentMockDB: sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })
	for i := 0; i < 10; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("test-tenant"))
		mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
	}
	return mockDB
}

const testIdempotencyKey = "11111111-1111-1111-1111-111111111111"

// TestRegisterRoutes_AllRoutesRegistered pins every operation declared for the
// taxonomy tag in api/openapi/v1/openapi.yaml to a live mux registration —
// mirrors internal/modules/approval/http/router_test.go. A spec
// route with no ServerInterface method fails to compile (routes_generated.go
// asserts `var _ taxonomyapi.ServerInterface = (*Handler)(nil)`); this test
// closes the remaining gap where a route compiles but is never reachable
// (e.g. dropped from RegisterRoutes, wrong BaseURL). A 404 here means the
// path is unreachable at runtime even though the spec declares it.
func TestRegisterRoutes_AllRoutesRegistered(t *testing.T) {
	mux := http.NewServeMux()
	h := NewHandler(fakeProfileService{}, fakeAreaService{}, fakeFamilyService{}, nil)
	h.RegisterRoutes(mux)

	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/taxonomy/profiles"},
		{method: http.MethodPost, path: "/api/v1/taxonomy/profiles"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/profiles/p-1"},
		{method: http.MethodPatch, path: "/api/v1/taxonomy/profiles/p-1"},
		{method: http.MethodDelete, path: "/api/v1/taxonomy/profiles/p-1"},
		{method: http.MethodPut, path: "/api/v1/taxonomy/profiles/p-1/default-template"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/areas"},
		{method: http.MethodPost, path: "/api/v1/taxonomy/areas"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/areas/a-1"},
		{method: http.MethodPut, path: "/api/v1/taxonomy/areas/a-1"},
		{method: http.MethodDelete, path: "/api/v1/taxonomy/areas/a-1"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/families"},
		{method: http.MethodPost, path: "/api/v1/taxonomy/families"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/families/f-1"},
		{method: http.MethodPatch, path: "/api/v1/taxonomy/families/f-1"},
		{method: http.MethodDelete, path: "/api/v1/taxonomy/families/f-1"},
	}

	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Errorf("route %s %s not registered (got 404)", rt.method, rt.path)
		}
	}
}
