package idempotency_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/testsupport/pgtest"
)

func jsonEqual(a, b []byte) bool {
	var va, vb any
	return json.Unmarshal(a, &va) == nil && json.Unmarshal(b, &vb) == nil &&
		func() bool { ra, _ := json.Marshal(va); rb, _ := json.Marshal(vb); return bytes.Equal(ra, rb) }()
}

const (
	testTenantMW = "00000000-0000-4000-8000-000000000010"
	testActorMW  = "actor-middleware-test"
)

func withIDs(tenant, actor string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "tenant", tenant)
			ctx = context.WithValue(ctx, "actor", actor)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func actorFromCtx(ctx context.Context) (string, string) {
	return ctx.Value("tenant").(string), ctx.Value("actor").(string)
}

func handler201(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(body))
	})
}

func TestMiddleware_MissingHeader_Returns400(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /test")
	h := withIDs(testTenantMW, testActorMW)(idempotency.Require(store, actorFromCtx)(handler201(`{}`)))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestMiddleware_InvalidUUID_Returns400(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /test")
	h := withIDs(testTenantMW, testActorMW)(idempotency.Require(store, actorFromCtx)(handler201(`{}`)))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "not-a-uuid")
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestMiddleware_FirstCall_RecordsAndPasses(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /test")
	h := withIDs(testTenantMW, testActorMW)(idempotency.Require(store, actorFromCtx)(handler201(`{"id":"1"}`)))
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
	body, _ := io.ReadAll(rec2.Body)
	if !jsonEqual(body, []byte(`{"id":"1"}`)) {
		t.Fatalf("replay body mismatch: %s", body)
	}
}

func TestMiddleware_Conflict_Returns422(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /test")
	h := withIDs(testTenantMW, testActorMW)(idempotency.Require(store, actorFromCtx)(handler201(`{}`)))
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":1}`)))
	req.Header.Set("Idempotency-Key", "22222222-2222-4222-8222-222222222222")
	h.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"x":2}`)))
	req2.Header.Set("Idempotency-Key", "22222222-2222-4222-8222-222222222222")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 422 {
		t.Fatalf("conflict status: got %d want 422", rec2.Code)
	}
}
