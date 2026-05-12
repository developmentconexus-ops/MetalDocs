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
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	approvalsignature "metaldocs/internal/modules/documents/approval/infra/signature"
	"metaldocs/internal/modules/documents/approval/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type fakeDecisionService struct {
	gotReq application.SignoffRequest
	result application.SignoffResult
	err    error
}

func (f *fakeDecisionService) RecordSignoff(_ context.Context, _ *sql.DB, req application.SignoffRequest) (application.SignoffResult, error) {
	f.gotReq = req
	if f.err != nil {
		return application.SignoffResult{}, f.err
	}
	return f.result, nil
}

func signoffTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v2/approval/instances/{instance_id}/stages/{stage_id}/signoff", h.SignoffHandler)
	return mux
}

func TestSignoffHandler_HappyApprove(t *testing.T) {
	fakeSvc := &fakeDecisionService{result: application.SignoffResult{InstanceApproved: true}}
	h := &Handler{decisionSvc: fakeSvc}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "v3")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var out contracts.SignoffResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Outcome != "approved" {
		t.Fatalf("outcome = %q, want %q", out.Outcome, "approved")
	}
	if fakeSvc.gotReq.TenantID != "tenant-1" || fakeSvc.gotReq.InstanceID != "inst-1" || fakeSvc.gotReq.StageInstanceID != "stg-1" || fakeSvc.gotReq.ActorUserID != "actor-1" {
		t.Fatalf("unexpected request mapped to service: %+v", fakeSvc.gotReq)
	}
	if fakeSvc.gotReq.Decision != "approve" {
		t.Fatalf("decision = %q, want %q", fakeSvc.gotReq.Decision, "approve")
	}
	if fakeSvc.gotReq.SignatureMethod != "password_reauth" {
		t.Fatalf("signature_method = %q, want %q", fakeSvc.gotReq.SignatureMethod, "password_reauth")
	}
	if got := fakeSvc.gotReq.SignaturePayload["password_token"]; got != "secret" {
		t.Fatalf("signature payload password_token = %#v", got)
	}
	if got := fakeSvc.gotReq.ContentFormData["_content_hash"]; got != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("content form data _content_hash = %#v", got)
	}
}

func TestSignoffHandler_HappyReject(t *testing.T) {
	fakeSvc := &fakeDecisionService{result: application.SignoffResult{InstanceRejected: true}}
	h := &Handler{decisionSvc: fakeSvc}
	mux := signoffTestMux(h)

	body := `{"decision":"reject","reason":"wrong value","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "v2")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var out contracts.SignoffResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Outcome != "rejected" {
		t.Fatalf("outcome = %q, want %q", out.Outcome, "rejected")
	}
}

func TestSignoffHandler_SoDViolation(t *testing.T) {
	h := &Handler{decisionSvc: &fakeDecisionService{err: domain.ErrAuthorCannotSign}}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "v1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestSignoffHandler_SignatureInvalid(t *testing.T) {
	h := &Handler{decisionSvc: &fakeDecisionService{err: approvalsignature.ErrInvalidCredentials}}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"bad","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "v1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSignoffHandler_ContentHashMismatch(t *testing.T) {
	h := &Handler{decisionSvc: &fakeDecisionService{err: ErrContentHashMismatch}}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "v1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusPreconditionFailed)
	}
}

func TestSignoffHandler_MissingIdempotencyKey(t *testing.T) {
	h := &Handler{decisionSvc: &fakeDecisionService{}}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("If-Match", "v1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	var out problem.Problem
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Code != "idempotency.key_required" {
		t.Fatalf("error.code = %q, want %q", out.Code, "idempotency.key_required")
	}
}

func TestSignoffHandler_MissingIfMatch(t *testing.T) {
	h := &Handler{decisionSvc: &fakeDecisionService{}}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusPreconditionRequired)
	}
}

func TestSignoffHandler_MalformedIfMatch(t *testing.T) {
	h := &Handler{decisionSvc: &fakeDecisionService{}}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "invalid")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- fakes for SignoffByDocumentHandler tests ---

type fakeReadService struct {
	inst *domain.Instance
	err  error
}

func (f *fakeReadService) LoadInstance(_ context.Context, _ *sql.DB, _, _, _ string) (*domain.Instance, error) {
	return f.inst, f.err
}

func (f *fakeReadService) LoadActiveInstanceByDocument(_ context.Context, _ *sql.DB, _, _ string) (*domain.Instance, error) {
	return f.inst, f.err
}

func (f *fakeReadService) ListPendingForActor(_ context.Context, _ *sql.DB, _, _, _ string, _, _ int) ([]domain.Instance, error) {
	return nil, nil
}

func (f *fakeReadService) ListInboxItems(_ context.Context, _ *sql.DB, _, _, _ string, _, _ int) ([]application.InboxView, error) {
	return nil, nil
}

func (f *fakeReadService) CountPendingForActor(_ context.Context, _ *sql.DB, _, _, _ string) (int, error) {
	return 0, nil
}

type fakeIdempStore struct {
	entries map[string]string // composite key → outcome
}

func (f *fakeIdempStore) CheckReplay(_ context.Context, tenantID, actorID, key string) (bool, string, error) {
	k := tenantID + ":" + actorID + ":" + key
	v, ok := f.entries[k]
	return ok, v, nil
}

func (f *fakeIdempStore) RecordReplay(_ context.Context, tenantID, actorID, key, outcome string) error {
	k := tenantID + ":" + actorID + ":" + key
	if f.entries == nil {
		f.entries = make(map[string]string)
	}
	f.entries[k] = outcome
	return nil
}

func docSignoffTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v2/documents/{id}/signoff", h.SignoffByDocumentHandler)
	return mux
}

// TestSignoffByDocumentHandler_ReplayAfterClose verifies that replaying the same
// Idempotency-Key after the approval instance is already closed returns 200 with
// was_replay:true instead of 404/500.
func TestSignoffByDocumentHandler_ReplayAfterClose(t *testing.T) {
	store := &fakeIdempStore{entries: map[string]string{
		"tenant-1:actor-1:idem-replay": "approved",
	}}
	h := &Handler{
		decisionSvc: &fakeDecisionService{},
		readSvc:     &fakeReadService{err: repository.ErrNoActiveInstance},
		idempStore:  store,
	}
	mux := docSignoffTestMux(h)

	body := `{"decision":"approve","reason":"","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/documents/doc-1/signoff", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-replay")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var out contracts.SignoffResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.WasReplay {
		t.Fatalf("was_replay = %v, want true", out.WasReplay)
	}
	if out.Outcome != "approved" {
		t.Fatalf("outcome = %q, want %q", out.Outcome, "approved")
	}
}
