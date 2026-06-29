//go:build integration
// +build integration

package application_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	tokenshttp "metaldocs/internal/modules/tokens/delivery/http"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// nativeReservedSet is the hardcoded set of the 8 native/computed token keys the
// render resolver registry registers. Hardcoded here on purpose: tokens is a leaf
// and must NOT depend on the render module, even in tests (SP-2 §5.1, §11 — the
// ReservedNames port + composition-root adapter exist for exactly this reason).
// The production adapter derives the same keys from the resolver registry; this
// test proves the guard behavior + HTTP mapping independently of that coupling.
func nativeReservedSet() staticReserved {
	return staticReserved{
		"doc_code":           {},
		"doc_title":          {},
		"revision_number":    {},
		"effective_date":     {},
		"controlled_by_area": {},
		"author":             {},
		"approvers":          {},
		"approval_date":      {},
	}
}

// buildReservedHandler constructs the real production service wired to the test DB
// with the native reserved set, wrapped in the real HTTP handler. This drives the
// actual POST /api/v1/tokens path (strict decode -> Create guard -> problem+json).
func buildReservedHandler(t *testing.T, db *sql.DB) *tokenshttp.Handler {
	t.Helper()
	svc := buildSvc(db).WithReservedNames(nativeReservedSet())
	return tokenshttp.NewHandler(svc)
}

// postToken issues POST /api/v1/tokens with the given JSON body, carrying the
// tenant and actor identity the handler reads from context (mirroring the
// middleware chain). Returns the recorder for status + body assertions.
func postToken(t *testing.T, h *tokenshttp.Handler, tenantID, actorID, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", bytes.NewReader([]byte(body)))
	r.Header.Set("Content-Type", "application/json")
	ctx := tenant.WithTenantID(r.Context(), tenantID)
	ctx = iamdomain.WithAuthContext(ctx, actorID, nil)
	rr := httptest.NewRecorder()
	h.CreateToken(rr, r.WithContext(ctx))
	return rr
}

func problemCode(t *testing.T, b []byte) string {
	t.Helper()
	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("decode problem body: %v (body=%s)", err, b)
	}
	return out.Code
}

// TestHTTP_CreateToken_ReservedName_Author_422 verifies POST /api/v1/tokens with
// a native name ("author") returns HTTP 422 and problem code "reserved_name".
func TestHTTP_CreateToken_ReservedName_Author_422(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	h := buildReservedHandler(t, db)
	rr := postToken(t, h, tn.ID, user.ID, `{"name":"author","value":"v","label":"l"}`)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 (body=%s)", rr.Code, rr.Body.String())
	}
	if c := problemCode(t, rr.Body.Bytes()); c != string(tokenshttp.CodeTokenReservedName) {
		t.Fatalf("code = %q, want %q", c, tokenshttp.CodeTokenReservedName)
	}
}

// TestHTTP_CreateToken_ReservedName_ApprovalDate_422 verifies approval_date
// (registry-only; absent from templates' static placeholder catalog) is also
// blocked at the HTTP edge with 422 reserved_name (guards SP-2 N1).
func TestHTTP_CreateToken_ReservedName_ApprovalDate_422(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	h := buildReservedHandler(t, db)
	rr := postToken(t, h, tn.ID, user.ID, `{"name":"approval_date","value":"v","label":"l"}`)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 (body=%s)", rr.Code, rr.Body.String())
	}
	if c := problemCode(t, rr.Body.Bytes()); c != string(tokenshttp.CodeTokenReservedName) {
		t.Fatalf("code = %q, want %q", c, tokenshttp.CodeTokenReservedName)
	}
}

// TestHTTP_CreateToken_NonReservedName_201 verifies a non-native name
// ("company_name") passes the guard and is created with HTTP 201.
func TestHTTP_CreateToken_NonReservedName_201(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tn := testdb.NewTenant(t, db)
	user := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	grantSP1Caps(t, db, tn.ID, user.ID)

	h := buildReservedHandler(t, db)
	rr := postToken(t, h, tn.ID, user.ID, `{"name":"company_name","value":"Acme Corp","label":"Company name"}`)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body=%s)", rr.Code, rr.Body.String())
	}
	var out struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode created body: %v (body=%s)", err, rr.Body.String())
	}
	if out.Name != "company_name" {
		t.Fatalf("created name = %q, want %q", out.Name, "company_name")
	}
}
