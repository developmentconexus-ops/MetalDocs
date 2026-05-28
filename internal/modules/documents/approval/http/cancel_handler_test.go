package approvalhttp

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeCancelService struct {
	gotReq application.CancelInput
	result application.CancelResult
	err    error
	called bool
}

func (f *fakeCancelService) CancelInstance(_ context.Context, _ *sql.DB, req application.CancelInput) (application.CancelResult, error) {
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
			svcErr:       repository.ErrInstanceCompleted,
			wantStatus:   http.StatusConflict,
			wantSvcCalls: true,
		},
		{
			name:         "authz denied",
			body:         `{"reason":"request withdrawn"}`,
			svcErr:       authz.ErrCapDenied{Capability: "workflow.instance.cancel", AreaCode: "tenant-1", ActorID: "actor-1"},
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
			req.Header.Set("Idempotency-Key", "idem-1")
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
