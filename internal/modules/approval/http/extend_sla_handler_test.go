package approvalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

type fakeSLAExtensionService struct {
	gotReq application.ExtendRequest
	err    error
	called bool
}

func (f *fakeSLAExtensionService) Extend(_ context.Context, _ db.TxRunner, req application.ExtendRequest) error {
	f.called = true
	f.gotReq = req
	return f.err
}

func extendSLATestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/approval/instances/{instance_id}/extend-sla", h.ExtendSLAHandler)
	return mux
}

func extendSLARequest(body, idempKey string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/extend-sla", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	if idempKey != "" {
		req.Header.Set("Idempotency-Key", idempKey)
	}
	return req
}

const validIdempKey = "11111111-1111-4111-8111-111111111111"

// TestExtendSLAHandler_MissingIdempotencyKeyRejectedBeforeService proves the
// ordering, not just the status: a missing Idempotency-Key must never reach
// the service. Asserting only rr.Code == 400 would pass even if the
// idempotency check ran AFTER the service call (e.g. the service still ran
// and mutated state, then the response got overwritten) — the assertion on
// fakeSvc.called is what pins the fail-closed-BEFORE-the-service guarantee.
func TestExtendSLAHandler_MissingIdempotencyKeyRejectedBeforeService(t *testing.T) {
	fakeSvc := &fakeSLAExtensionService{}
	h := &Handler{slaExtensionSvc: fakeSvc, readSvc: &fakeReadService{}, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}

	req := extendSLARequest(`{"due_at":"2026-09-01T12:00:00Z","reason":"deadline pushed"}`, "")
	rr := httptest.NewRecorder()
	extendSLATestMux(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (missing Idempotency-Key)", rr.Code, http.StatusBadRequest)
	}
	if fakeSvc.called {
		t.Fatalf("service was called for a request with no Idempotency-Key; it must fail closed before the service runs")
	}
}

// TestExtendSLAHandler_MalformedIdempotencyKeyRejectedBeforeService mirrors
// the missing-key case for a syntactically invalid key.
func TestExtendSLAHandler_MalformedIdempotencyKeyRejectedBeforeService(t *testing.T) {
	fakeSvc := &fakeSLAExtensionService{}
	h := &Handler{slaExtensionSvc: fakeSvc, readSvc: &fakeReadService{}, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}

	req := extendSLARequest(`{"due_at":"2026-09-01T12:00:00Z","reason":"deadline pushed"}`, "not-a-uuid")
	rr := httptest.NewRecorder()
	extendSLATestMux(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (malformed Idempotency-Key)", rr.Code, http.StatusBadRequest)
	}
	if fakeSvc.called {
		t.Fatalf("service was called for a request with a malformed Idempotency-Key; it must fail closed before the service runs")
	}
}

// TestExtendSLAHandler_ProblemCodes pins each of the three registered
// problem codes (openapi.yaml) to its declared HTTP status on the wire.
func TestExtendSLAHandler_ProblemCodes(t *testing.T) {
	// The code assertion is not redundant with the status one: two of these
	// three codes are BOTH 422, so status alone cannot tell them apart, and
	// the blank-reason case would still pass while returning not_forward —
	// which is the exact defect (a code the contract advertises but the wire
	// never returns) that these cases exist to pin down.
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "not forward -> 422 validation.sla_extension_not_forward",
			body:       `{"due_at":"2026-09-01T12:00:00Z","reason":"deadline pushed"}`,
			svcErr:     application.ErrSLAExtensionNotForward,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation.sla_extension_not_forward",
		},
		{
			name:       "no active stage -> 409 state.sla_extension_no_active_stage",
			body:       `{"due_at":"2026-09-01T12:00:00Z","reason":"deadline pushed"}`,
			svcErr:     application.ErrSLAExtensionNoActiveStage,
			wantStatus: http.StatusConflict,
			wantCode:   "state.sla_extension_no_active_stage",
		},
		{
			// Blank reason never reaches the service — the handler
			// short-circuits it directly to the same sentinel/code the
			// service would have returned (finding 1: the contract's
			// advertised 422 must be live behind the route).
			name:       "blank reason -> 422 validation.sla_extension_reason_required",
			body:       `{"due_at":"2026-09-01T12:00:00Z","reason":"   "}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation.sla_extension_reason_required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSvc := &fakeSLAExtensionService{err: tt.svcErr}
			h := &Handler{slaExtensionSvc: fakeSvc, readSvc: &fakeReadService{}, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}

			req := extendSLARequest(tt.body, validIdempKey)
			rr := httptest.NewRecorder()
			extendSLATestMux(h).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			var prob struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &prob); err != nil {
				t.Fatalf("decode problem body: %v (body %s)", err, rr.Body.String())
			}
			if prob.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q — the status alone does not distinguish the two 422s, so a wrong code here means the contract advertises one thing and the wire returns another", prob.Code, tt.wantCode)
			}
		})
	}
}

// TestExtendSLAHandler_ReasonRequired_DoesNotCallService confirms the
// blank-reason short-circuit fires before the service is invoked at all,
// not merely that it maps to the right status.
func TestExtendSLAHandler_ReasonRequired_DoesNotCallService(t *testing.T) {
	fakeSvc := &fakeSLAExtensionService{}
	h := &Handler{slaExtensionSvc: fakeSvc, readSvc: &fakeReadService{}, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}

	req := extendSLARequest(`{"due_at":"2026-09-01T12:00:00Z","reason":""}`, validIdempKey)
	rr := httptest.NewRecorder()
	extendSLATestMux(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
	if fakeSvc.called {
		t.Fatalf("service was called for a blank-reason request; the handler must short-circuit before invoking it")
	}
}

// TestExtendSLAHandler_AcceptsNonUTCOffset locks in finding 2: openapi.yaml
// declares due_at as `format: date-time`, which admits any RFC3339 offset.
// The handler must not reject an offset the published contract accepts.
func TestExtendSLAHandler_AcceptsNonUTCOffset(t *testing.T) {
	fakeSvc := &fakeSLAExtensionService{}
	fakeRead := &fakeReadService{inst: &domain.Instance{
		ID:          "inst-1",
		DocumentID:  "doc-1",
		RouteID:     "route-1",
		TenantID:    "tenant-1",
		Status:      domain.InstanceInProgress,
		SubmittedBy: "actor-1",
		SubmittedAt: time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
	}}
	h := &Handler{slaExtensionSvc: fakeSvc, readSvc: fakeRead, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}

	req := extendSLARequest(`{"due_at":"2026-09-01T12:00:00-03:00","reason":"deadline pushed"}`, validIdempKey)
	rr := httptest.NewRecorder()
	extendSLATestMux(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %s) — a non-UTC offset must be accepted per openapi.yaml's date-time format", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !fakeSvc.called {
		t.Fatalf("service was not called")
	}
	wantDueAt := time.Date(2026, 9, 1, 15, 0, 0, 0, time.UTC) // 12:00 -03:00 == 15:00 UTC
	if !fakeSvc.gotReq.NewDueAt.Equal(wantDueAt) {
		t.Fatalf("NewDueAt = %v, want %v (same instant as -03:00 offset)", fakeSvc.gotReq.NewDueAt, wantDueAt)
	}
}

// TestExtendSLAHandler_HappyPath_DelegatesToGetInstanceHandler asserts the
// success path calls the service with the tenant/actor/instance from the
// request and then delegates to GetInstanceHandler for the response body.
func TestExtendSLAHandler_HappyPath_DelegatesToGetInstanceHandler(t *testing.T) {
	fakeSvc := &fakeSLAExtensionService{}
	fakeRead := &fakeReadService{inst: &domain.Instance{
		ID:              "inst-1",
		DocumentID:      "doc-1",
		RouteID:         "route-1",
		TenantID:        "tenant-1",
		Status:          domain.InstanceInProgress,
		SubmittedBy:     "actor-1",
		SubmittedAt:     time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
		RevisionVersion: 3,
	}}
	h := &Handler{slaExtensionSvc: fakeSvc, readSvc: fakeRead, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}

	req := extendSLARequest(`{"due_at":"2026-09-01T12:00:00Z","reason":"deadline pushed"}`, validIdempKey)
	rr := httptest.NewRecorder()
	extendSLATestMux(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %s)", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !fakeSvc.called {
		t.Fatalf("service was not called")
	}
	if fakeSvc.gotReq.TenantID != "tenant-1" || fakeSvc.gotReq.InstanceID != "inst-1" || fakeSvc.gotReq.ActorID != "actor-1" || fakeSvc.gotReq.Reason != "deadline pushed" {
		t.Fatalf("unexpected service request: %+v", fakeSvc.gotReq)
	}
	wantDueAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	if !fakeSvc.gotReq.NewDueAt.Equal(wantDueAt) {
		t.Fatalf("NewDueAt = %v, want %v", fakeSvc.gotReq.NewDueAt, wantDueAt)
	}

	var out contracts.InstanceResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ID != "inst-1" {
		t.Fatalf("response ID = %q, want %q (handler did not delegate to GetInstanceHandler)", out.ID, "inst-1")
	}
}

// TestExtendSLAHandler_AuthzDenied pins the authz-denied mapping (403),
// matching the module's other tier-2 authz failures.
func TestExtendSLAHandler_AuthzDenied(t *testing.T) {
	fakeSvc := &fakeSLAExtensionService{err: authz.ErrCapDenied{Capability: "approval.sla.extend", AreaCode: "tenant-1", ActorID: "actor-1"}}
	h := &Handler{slaExtensionSvc: fakeSvc, readSvc: &fakeReadService{}, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}

	req := extendSLARequest(`{"due_at":"2026-09-01T12:00:00Z","reason":"deadline pushed"}`, validIdempKey)
	rr := httptest.NewRecorder()
	extendSLATestMux(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}
