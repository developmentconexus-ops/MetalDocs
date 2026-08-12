//go:build integration

package idempotency_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

func jsonEqual(a, b []byte) bool {
	var va, vb any
	return json.Unmarshal(a, &va) == nil && json.Unmarshal(b, &vb) == nil &&
		func() bool { ra, _ := json.Marshal(va); rb, _ := json.Marshal(vb); return bytes.Equal(ra, rb) }()
}

const testActorMW = "actor-middleware-test"

func withIDs(tenantID, actorID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := tenant.WithTenantID(r.Context(), tenantID)
			ctx = iamdomain.WithAuthContext(ctx, actorID, nil)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func handler201(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(body))
	})
}

func TestMiddleware_MissingHeader_Returns400(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	store := idempotency.New(db, "POST /test")
	h := withIDs(tenant.ID, testActorMW)(idempotency.Require(store)(handler201(`{}`)))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestMiddleware_InvalidUUID_Returns400(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	store := idempotency.New(db, "POST /test")
	h := withIDs(tenant.ID, testActorMW)(idempotency.Require(store)(handler201(`{}`)))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "not-a-uuid")
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

// TestMiddleware_MissingActor_Returns401 pins the client-fault branch: a
// blank actor is something the caller can fix by re-authenticating, so it
// stays 401 (A3.3). This is the branch that must NOT change when the
// tenant-missing branch below is split out to 500.
func TestMiddleware_MissingActor_Returns401(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantRow := testdb.NewTenant(t, db)
	store := idempotency.New(db, "POST /test")
	h := withIDs(tenantRow.ID, "")(idempotency.Require(store)(handler201(`{}`)))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	req.Header.Set("Idempotency-Key", "44444444-4444-4444-8444-444444444444")
	h.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("status: got %d want 401", rec.Code)
	}
}

// TestMiddleware_MissingTenant_Returns500 pins the server-fault branch found
// by PR #122 review (chatgpt-codex-connector, identity.go:40): a missing
// tenant means the auth middleware did not run or the session is corrupt —
// internal/platform/tenant/context.go:15-17 requires this be surfaced as a
// 500 invariant violation, matching every other production consumer of
// tenant.FromContext (e.g. controlleddocuments' injectTenant), not folded
// into the same 401 branch as a missing actor.
func TestMiddleware_MissingTenant_Returns500(t *testing.T) {
	db, _ := testdb.Open(t)
	store := idempotency.New(db, "POST /test")
	h := idempotency.Require(store)(handler201(`{}`))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	req.Header.Set("Idempotency-Key", "55555555-5555-4555-8555-555555555555")
	h.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status: got %d want 500", rec.Code)
	}
}

func TestMiddleware_FirstCall_RecordsAndPasses(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	store := idempotency.New(db, "POST /test")
	h := withIDs(tenant.ID, testActorMW)(idempotency.Require(store)(handler201(`{"id":"1"}`)))
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("first status: got %d want 201", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	req2.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 201 {
		t.Fatalf("replay status: got %d want 201", rec2.Code)
	}
	if got := rec2.Header().Get("Idempotent-Replay"); got != "true" {
		t.Fatalf("replay header: got %q want true", got)
	}
	body, _ := io.ReadAll(rec2.Body)
	if !jsonEqual(body, []byte(`{"id":"1"}`)) {
		t.Fatalf("replay body mismatch: %s", body)
	}
}

// ADR 0089 (commit 75c03821, annex R-8) ratified 409 Conflict for
// conflict.idempotency_key_reused; middleware.go returns http.StatusConflict
// deliberately. These tests previously asserted the pre-ADR-0089 422 and were
// never updated in the rename sweep.
func TestMiddleware_Conflict_Returns409(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	store := idempotency.New(db, "POST /test")
	h := withIDs(tenant.ID, testActorMW)(idempotency.Require(store)(handler201(`{}`)))
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	req.Header.Set("Idempotency-Key", "22222222-2222-4222-8222-222222222222")
	h.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":2}`)))
	req2.Header.Set("Idempotency-Key", "22222222-2222-4222-8222-222222222222")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 409 {
		t.Fatalf("conflict status: got %d want 409", rec2.Code)
	}
}

func TestMiddleware_SameKeyDifferentResourcePath_Returns409(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	store := idempotency.New(db, "POST /test/{id}")
	h := withIDs(tenant.ID, testActorMW)(idempotency.Require(store)(handler201(`{"ok":true}`)))

	req := httptest.NewRequest("POST", "/test/one", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")
	h.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest("POST", "/test/two", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 409 {
		t.Fatalf("status: got %d want 409", rec2.Code)
	}
}
