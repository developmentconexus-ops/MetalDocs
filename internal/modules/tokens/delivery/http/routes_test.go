package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	openapi_types "github.com/oapi-codegen/runtime/types"

	iamdomain "metaldocs/internal/modules/iam/domain"
	tokenshttp "metaldocs/internal/modules/tokens/delivery/http"
	"metaldocs/internal/modules/tokens/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// fakeService satisfies tokenshttp.TokenService.
type fakeService struct {
	createErr error
	getErr    error
	listErr   error
	entry     *domain.Entry
}

func (f *fakeService) Create(ctx context.Context, cmd tokenshttp.CreateCommand) (*domain.Entry, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.entry, nil
}

func (f *fakeService) Get(ctx context.Context, tenantID, id string) (*domain.Entry, error) {
	return f.entry, f.getErr
}

func (f *fakeService) List(ctx context.Context, tenantID string) ([]domain.Entry, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return nil, nil
}

func (f *fakeService) Update(ctx context.Context, cmd tokenshttp.UpdateCommand) (*domain.Entry, error) {
	return f.entry, f.getErr
}

func (f *fakeService) Delete(ctx context.Context, tenantID, actorID, id string) error {
	return f.getErr
}

func reqWithTenant(t *testing.T, method, target, body string) *http.Request {
	t.Helper()
	var buf *bytes.Reader
	if body != "" {
		buf = bytes.NewReader([]byte(body))
	} else {
		buf = bytes.NewReader(nil)
	}
	r := httptest.NewRequest(method, target, buf)
	r.Header.Set("Content-Type", "application/json")
	ctx := tenant.WithTenantID(r.Context(), "11111111-1111-1111-1111-111111111111")
	// A3.3: every tokens mutation route resolves the actor before it touches the
	// service (ActorID is the audit attribution on the token entry), so a request
	// with no auth context now stops at 401 and can no longer reach the case
	// under test. The authn middleware always installs one in production.
	ctx = iamdomain.WithAuthContext(ctx, "user-a", []iamdomain.Role{})
	return r.WithContext(ctx)
}

func decodeCode(t *testing.T, b []byte) string {
	t.Helper()
	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, b)
	}
	return out.Code
}

// testID is a valid UUID for path parameter tests.
var testID = openapi_types.UUID{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}

func TestCreate_DuplicateReturns409(t *testing.T) {
	h := tokenshttp.NewHandler(&fakeService{createErr: &pgconn.PgError{Code: "23505"}})
	rr := httptest.NewRecorder()
	h.CreateToken(rr, reqWithTenant(t, http.MethodPost, "/api/v1/tokens", `{"name":"x","value":"v","label":"l"}`))
	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (body=%s)", rr.Code, rr.Body.String())
	}
	if c := decodeCode(t, rr.Body.Bytes()); c != tokenshttp.CodeTokenAlreadyExists.String() {
		t.Fatalf("code = %q, want %q", c, tokenshttp.CodeTokenAlreadyExists)
	}
}

func TestCreate_UnknownFieldReturns400(t *testing.T) {
	h := tokenshttp.NewHandler(&fakeService{entry: &domain.Entry{}})
	rr := httptest.NewRecorder()
	// "extra" is not a known field — strict decode must reject it with 400.
	h.CreateToken(rr, reqWithTenant(t, http.MethodPost, "/api/v1/tokens", `{"name":"x","value":"v","label":"l","extra":1}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rr.Code, rr.Body.String())
	}
}

func TestCreate_ValidationReturns422(t *testing.T) {
	h := tokenshttp.NewHandler(&fakeService{createErr: &domain.ValidationError{Field: "name", Message: "bad"}})
	rr := httptest.NewRecorder()
	h.CreateToken(rr, reqWithTenant(t, http.MethodPost, "/api/v1/tokens", `{"name":"bad name","value":"v","label":"l"}`))
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 (body=%s)", rr.Code, rr.Body.String())
	}
}

// TestList_MissingActorReturns401 is the PR #108 review-round-1 remediation for
// FINDING 2: Service.List (A3.3) returns authn.ErrMissingActor when the request
// context carries no authenticated actor, but ListTokens mapped every
// Service.List error straight to 500 internal.unknown instead of routing it
// through writeTokenError, which already has a case for this sentinel and every
// other tokens route uses it correctly (see TestUpdate_ImmutableNameReturns422
// and TestCreate_ValidationReturns422 above, both driven through
// writeTokenError).
func TestList_MissingActorReturns401(t *testing.T) {
	h := tokenshttp.NewHandler(&fakeService{listErr: authn.ErrMissingActor})
	rr := httptest.NewRecorder()
	h.ListTokens(rr, reqWithTenant(t, http.MethodGet, "/api/v1/tokens", ``))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body=%s)", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	if c := decodeCode(t, rr.Body.Bytes()); c != problem.CodeAuthUnauthenticated.String() {
		t.Fatalf("code = %q, want %q", c, problem.CodeAuthUnauthenticated)
	}
}

func TestGet_NotFoundReturns404(t *testing.T) {
	h := tokenshttp.NewHandler(&fakeService{getErr: domain.ErrNotFound})
	rr := httptest.NewRecorder()
	h.GetToken(rr, reqWithTenant(t, http.MethodGet, "/api/v1/tokens/11111111-1111-1111-1111-111111111111", ``), testID)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (body=%s)", rr.Code, rr.Body.String())
	}
}

func TestUpdate_ImmutableNameReturns422(t *testing.T) {
	h := tokenshttp.NewHandler(&fakeService{getErr: domain.ErrImmutableName})
	rr := httptest.NewRecorder()
	h.UpdateToken(rr, reqWithTenant(t, http.MethodPut, "/api/v1/tokens/11111111-1111-1111-1111-111111111111", `{"name":"changed","value":"v","label":"l"}`), testID)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 (body=%s)", rr.Code, rr.Body.String())
	}
	if c := decodeCode(t, rr.Body.Bytes()); c != tokenshttp.CodeTokenImmutableField.String() {
		t.Fatalf("code = %q, want %q", c, tokenshttp.CodeTokenImmutableField)
	}
}
