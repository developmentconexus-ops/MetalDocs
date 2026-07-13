package approvalhttp

import (
	"context"
	"encoding/json"
	"errors"
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

type fakeSubmitService struct {
	result application.SubmitResult
	err    error
	gotReq application.SubmitRequest
	called bool
}

func (f *fakeSubmitService) SubmitRevisionForReview(_ context.Context, _ db.TxRunner, req application.SubmitRequest) (application.SubmitResult, error) {
	f.called = true
	f.gotReq = req
	if f.err != nil {
		return application.SubmitResult{}, f.err
	}
	return f.result, nil
}

// PreviewRoute is a no-op stub only to satisfy the submitService interface —
// this fake drives the submit handler tests, which never call PreviewRoute.
// PreviewRoute's own behavior is covered in
// internal/modules/approval/application/route_preview_test.go and the
// resolution-parity integration test under tests/integration/approval/.
func (f *fakeSubmitService) PreviewRoute(_ context.Context, _ db.TxRunner, _, _ string) (application.RoutePreview, error) {
	return application.RoutePreview{}, nil
}

func submitTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/documents/{id}/submit", h.SubmitHandler)
	return mux
}

func TestSubmitHandler(t *testing.T) {
	tests := []struct {
		name           string
		ifMatch        string
		idempotencyKey string
		body           string
		svcErr         error
		wantStatus     int
		wantInstanceID string
		wantETag       string
	}{
		{
			name:           "happy path",
			ifMatch:        "\"v3\"",
			idempotencyKey: "22222222-2222-2222-2222-222222222222",
			body:           `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			svcErr:         nil,
			wantStatus:     http.StatusCreated,
			wantInstanceID: "inst-123",
			wantETag:       "\"v4\"",
		},
		{
			name:           "missing idempotency key",
			ifMatch:        "\"v3\"",
			idempotencyKey: "",
			body:           `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			wantStatus:     http.StatusBadRequest,
		},
		{
			name:           "missing if-match",
			ifMatch:        "",
			idempotencyKey: "22222222-2222-2222-2222-222222222222",
			body:           `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			wantStatus:     http.StatusPreconditionRequired,
		},
		{
			name:           "malformed if-match",
			ifMatch:        "oops",
			idempotencyKey: "22222222-2222-2222-2222-222222222222",
			body:           `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			wantStatus:     http.StatusBadRequest,
		},
		{
			// ADR 0073: an omitted route_id is valid (in-tx resolution), so the
			// validate-failure case must use a PRESENT-but-malformed route_id.
			name:           "validate fails",
			ifMatch:        "\"v1\"",
			idempotencyKey: "22222222-2222-2222-2222-222222222222",
			body:           `{"route_id":"not-a-uuid","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			wantStatus:     http.StatusBadRequest,
		},
		{
			name:           "service stale revision",
			ifMatch:        "\"v2\"",
			idempotencyKey: "22222222-2222-2222-2222-222222222222",
			body:           `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			svcErr:         infrastructure.ErrStaleRevision,
			wantStatus:     http.StatusConflict,
		},
		{
			name:           "service capability denied",
			ifMatch:        "\"v2\"",
			idempotencyKey: "22222222-2222-2222-2222-222222222222",
			body:           `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			svcErr:         authz.ErrCapDenied{Capability: "doc.submit", AreaCode: "tenant", ActorID: "actor-1"},
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "service generic error",
			ifMatch:        "\"v2\"",
			idempotencyKey: "22222222-2222-2222-2222-222222222222",
			body:           `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			svcErr:         errors.New("boom"),
			wantStatus:     http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeSubmitService{result: application.SubmitResult{InstanceID: "inst-123"}, err: tt.svcErr}
			h := &Handler{submitSvc: svc}
			mux := submitTestMux(h)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/submit", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
			if tt.idempotencyKey != "" {
				req.Header.Set("Idempotency-Key", tt.idempotencyKey)
			}
			req.Header.Set("X-Request-ID", "req-123")
			if tt.ifMatch != "" {
				req.Header.Set("If-Match", tt.ifMatch)
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			// Missing idempotency key must fail closed before the service is invoked.
			if tt.name == "missing idempotency key" && svc.called {
				t.Fatalf("service was called despite missing idempotency key")
			}

			if tt.wantStatus == http.StatusCreated {
				var out contracts.SubmitResponse
				if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if out.InstanceID != tt.wantInstanceID {
					t.Fatalf("instance_id = %q, want %q", out.InstanceID, tt.wantInstanceID)
				}
				if got := rr.Header().Get("ETag"); got != tt.wantETag {
					t.Fatalf("etag = %q, want %q", got, tt.wantETag)
				}
				if svc.gotReq.IdempotencyKey != tt.idempotencyKey {
					t.Fatalf("threaded idempotency key = %q, want %q", svc.gotReq.IdempotencyKey, tt.idempotencyKey)
				}
			}
		})
	}
}
