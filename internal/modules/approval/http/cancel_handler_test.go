package approvalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

type fakeCancelService struct {
	gotReq application.CancelInput
	result application.CancelResult
	err    error
	called bool
}

func (f *fakeCancelService) CancelInstance(_ context.Context, _ db.TxRunner, req application.CancelInput) (application.CancelResult, error) {
	f.called = true
	f.gotReq = req
	if f.err != nil {
		return application.CancelResult{}, f.err
	}
	return f.result, nil
}

func cancelTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/approval/instances/{instance_id}/cancel", h.CancelHandler)
	return mux
}

func TestCancelHandler(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		svcErr       error
		wantStatus   int
		wantSvcCalls bool
	}{
		{
			name:         "happy path",
			body:         `{"reason":"request withdrawn"}`,
			wantStatus:   http.StatusOK,
			wantSvcCalls: true,
		},
		{
			name:         "reason missing",
			body:         `{"reason":""}`,
			wantStatus:   http.StatusBadRequest,
			wantSvcCalls: false,
		},
		{
			name:         "instance completed",
			body:         `{"reason":"request withdrawn"}`,
			svcErr:       infrastructure.ErrInstanceCompleted,
			wantStatus:   http.StatusConflict,
			wantSvcCalls: true,
		},
		{
			name:         "authz denied",
			body:         `{"reason":"request withdrawn"}`,
			svcErr:       authz.ErrCapDenied{Capability: "document.edit", AreaCode: "tenant-1", ActorID: "actor-1"},
			wantStatus:   http.StatusForbidden,
			wantSvcCalls: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSvc := &fakeCancelService{result: application.CancelResult{DocumentID: "doc-4"}, err: tt.svcErr}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-4/cancel", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
			req.Header.Set("If-Match", "\"v9\"")

			rr := httptest.NewRecorder()
			cancelTestMux(&Handler{cancelSvc: fakeSvc}).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if fakeSvc.called != tt.wantSvcCalls {
				t.Fatalf("service called = %v, want %v", fakeSvc.called, tt.wantSvcCalls)
			}

			if tt.wantSvcCalls {
				if fakeSvc.gotReq.TenantID != "tenant-1" || fakeSvc.gotReq.InstanceID != "inst-4" || fakeSvc.gotReq.ActorUserID != "actor-1" || fakeSvc.gotReq.ExpectedRevisionVersion != 9 {
					t.Fatalf("unexpected service request: %+v", fakeSvc.gotReq)
				}
			}

			if tt.wantStatus == http.StatusOK {
				var out contracts.CancelResponse
				if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if out.DocumentID != "doc-4" {
					t.Fatalf("document_id = %q, want %q", out.DocumentID, "doc-4")
				}
			}
		})
	}
}

// TestCancelHandlerIfMatchPreconditions pins the OCC precondition contract for
// cancel. The wildcard case is the regression guard: the spec used to advertise
// `If-Match: *` as "skip the check", but the parser encoded it as version 0 and
// CancelService feeds ExpectedRevisionVersion verbatim into
// `WHERE revision_version = $3`, so the wildcard silently asserted v0 and the
// live request came back 409 conflict.stale_revision instead of cancelling.
// The contract no longer offers the wildcard: it is rejected up front as a
// malformed precondition (400) and never reaches the service, so no caller can
// mistake a failed OCC guard for a skipped one.
func TestCancelHandlerIfMatchPreconditions(t *testing.T) {
	tests := []struct {
		name         string
		ifMatch      string
		wantStatus   int
		wantSvcCalls bool
		wantVersion  int
	}{
		{name: "wildcard rejected", ifMatch: "*", wantStatus: http.StatusBadRequest},
		{name: "quoted wildcard rejected", ifMatch: `"*"`, wantStatus: http.StatusBadRequest},
		{name: "v0 rejected (governed transition)", ifMatch: `"v0"`, wantStatus: http.StatusBadRequest},
		{name: "missing header rejected", ifMatch: "", wantStatus: http.StatusPreconditionRequired},
		{name: "concrete version accepted", ifMatch: `"v1"`, wantStatus: http.StatusOK, wantSvcCalls: true, wantVersion: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSvc := &fakeCancelService{result: application.CancelResult{DocumentID: "doc-4"}}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-4/cancel", strings.NewReader(`{"reason":"request withdrawn"}`))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
			if tt.ifMatch != "" {
				req.Header.Set("If-Match", tt.ifMatch)
			}

			rr := httptest.NewRecorder()
			cancelTestMux(&Handler{cancelSvc: fakeSvc}).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("If-Match %q: status = %d, want %d (body %s)", tt.ifMatch, rr.Code, tt.wantStatus, rr.Body.String())
			}
			if fakeSvc.called != tt.wantSvcCalls {
				t.Fatalf("If-Match %q: service called = %v, want %v", tt.ifMatch, fakeSvc.called, tt.wantSvcCalls)
			}
			if tt.wantSvcCalls && fakeSvc.gotReq.ExpectedRevisionVersion != tt.wantVersion {
				t.Fatalf("If-Match %q: ExpectedRevisionVersion = %d, want %d", tt.ifMatch, fakeSvc.gotReq.ExpectedRevisionVersion, tt.wantVersion)
			}
		})
	}
}
