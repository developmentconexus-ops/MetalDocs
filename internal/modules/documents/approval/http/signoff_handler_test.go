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
	approvalinfra "metaldocs/internal/modules/documents/approval/infrastructure"
	approvalsignature "metaldocs/internal/modules/documents/approval/infrastructure/signature"
	"metaldocs/internal/modules/iam/authz"
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
	mux.HandleFunc("POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoff", h.SignoffHandler)
	return mux
}

func TestSignoffHandler_HappyApprove(t *testing.T) {
	fakeSvc := &fakeDecisionService{result: application.SignoffResult{InstanceApproved: true}}
	h := &Handler{decisionSvc: fakeSvc}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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

func TestSignoffHandler_ReplayReturnsCachedOutcome(t *testing.T) {
	fakeSvc := &fakeDecisionService{}
	h := &Handler{
		decisionSvc: fakeSvc,
		idempStore: &fakeIdempStore{entries: map[string]string{
			"stage:tenant-1:actor-1:idem-replay": "approved",
		}},
	}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-replay")
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
	if !out.WasReplay || out.Outcome != "approved" {
		t.Fatalf("response = %+v, want replay approved", out)
	}
	if fakeSvc.gotReq.InstanceID != "" {
		t.Fatalf("decision service should not run on replay, got %+v", fakeSvc.gotReq)
	}
}

func TestSignoffHandler_MissingIfMatch(t *testing.T) {
	h := &Handler{decisionSvc: &fakeDecisionService{}}
	mux := signoffTestMux(h)

	body := `{"decision":"approve","reason":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/signoff", strings.NewReader(body))
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
	inst         *domain.Instance
	err          error
	mutationInst *domain.Instance
	mutationErr  error

	loadActiveByDocumentCalled            bool
	loadActiveByDocumentForMutationCalled bool
}

func (f *fakeReadService) LoadInstance(_ context.Context, _ *sql.DB, _, _ string) (*domain.Instance, error) {
	return f.inst, f.err
}

func (f *fakeReadService) LoadActiveInstanceByDocument(_ context.Context, _ *sql.DB, _, _ string) (*domain.Instance, error) {
	f.loadActiveByDocumentCalled = true
	return f.inst, f.err
}

func (f *fakeReadService) LoadActiveInstanceByDocumentForMutation(_ context.Context, _ *sql.DB, _, _ string) (*domain.Instance, error) {
	f.loadActiveByDocumentForMutationCalled = true
	if f.mutationInst != nil || f.mutationErr != nil {
		return f.mutationInst, f.mutationErr
	}
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

func (f *fakeIdempStore) BeginDocumentReplay(_ context.Context, tenantID, actorID, key, _ string) (*approvalinfra.SignoffReplayHandle, *approvalinfra.SignoffReplay, error) {
	k := "document:" + tenantID + ":" + actorID + ":" + key
	v, ok := f.entries[k]
	if !ok {
		return nil, nil, nil
	}
	return nil, &approvalinfra.SignoffReplay{Outcome: v}, nil
}

func (f *fakeIdempStore) BeginStageReplay(_ context.Context, tenantID, actorID, key, _ string) (*approvalinfra.SignoffReplayHandle, *approvalinfra.SignoffReplay, error) {
	k := "stage:" + tenantID + ":" + actorID + ":" + key
	v, ok := f.entries[k]
	if !ok {
		return nil, nil, nil
	}
	return nil, &approvalinfra.SignoffReplay{Outcome: v}, nil
}

func docSignoffTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/documents/{id}/signoff", h.SignoffByDocumentHandler)
	return mux
}

func docCancelTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/documents/{id}/cancel", h.CancelByDocumentHandler)
	return mux
}

// TestSignoffByDocumentHandler_ReplayAfterClose verifies that replaying the same
// Idempotency-Key after the approval instance is already closed returns 200 with
// was_replay:true instead of 404/500.
func TestSignoffByDocumentHandler_ReplayAfterEligibility(t *testing.T) {
	store := &fakeIdempStore{entries: map[string]string{
		"document:tenant-1:actor-1:idem-replay": "approved",
	}}
	h := &Handler{
		decisionSvc: &fakeDecisionService{},
		readSvc: &fakeReadService{inst: &domain.Instance{
			ID:       "inst-1",
			TenantID: "tenant-1",
			Stages: []domain.StageInstance{{
				ID:               "stage-1",
				Status:           domain.StageActive,
				EligibleActorIDs: []string{"actor-1"},
			}},
		}},
		idempStore: store,
	}
	mux := docSignoffTestMux(h)

	body := `{"decision":"approve","reason":"","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-replay")
	req.Header.Set("If-Match", "\"v5\"")

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

func TestSignoffByDocumentHandler_ReplayRequiresCurrentEligibility(t *testing.T) {
	store := &fakeIdempStore{entries: map[string]string{
		"document:tenant-1:actor-1:idem-replay": "approved",
	}}
	h := &Handler{
		decisionSvc: &fakeDecisionService{},
		readSvc: &fakeReadService{inst: &domain.Instance{
			ID:       "inst-1",
			TenantID: "tenant-1",
			Stages: []domain.StageInstance{{
				ID:               "stage-1",
				Status:           domain.StageActive,
				EligibleActorIDs: []string{"actor-2"},
			}},
		}},
		idempStore: store,
	}
	mux := docSignoffTestMux(h)

	body := `{"decision":"approve","reason":"","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-replay")
	req.Header.Set("If-Match", "\"v5\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestCancelByDocumentHandler_UsesIfMatchRevision(t *testing.T) {
	cancelSvc := &fakeCancelService{result: application.CancelResult{DocumentID: "doc-1"}}
	h := &Handler{
		readSvc: &fakeReadService{inst: &domain.Instance{
			ID:         "inst-1",
			DocumentID: "doc-1",
			TenantID:   "tenant-1",
		}},
		cancelSvc: cancelSvc,
	}
	mux := docCancelTestMux(h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/cancel", strings.NewReader(`{"reason":"withdrawn"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "\"v6\"")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if cancelSvc.gotReq.ExpectedRevisionVersion != 6 {
		t.Fatalf("expected revision = %d, want 6", cancelSvc.gotReq.ExpectedRevisionVersion)
	}
	if cancelSvc.gotReq.InstanceID != "inst-1" || cancelSvc.gotReq.ActorUserID != "actor-1" {
		t.Fatalf("unexpected cancel request: %+v", cancelSvc.gotReq)
	}
}

func TestCancelByDocumentHandler_RequiresIfMatch(t *testing.T) {
	h := &Handler{readSvc: &fakeReadService{}, cancelSvc: &fakeCancelService{}}
	mux := docCancelTestMux(h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/cancel", strings.NewReader(`{"reason":"withdrawn"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusPreconditionRequired, rr.Body.String())
	}
}

func TestSignoffByDocumentHandler_UsesIfMatchRevision(t *testing.T) {
	fakeSvc := &fakeDecisionService{result: application.SignoffResult{StageCompleted: true}}
	h := &Handler{
		decisionSvc: fakeSvc,
		readSvc: &fakeReadService{inst: &domain.Instance{
			ID:       "inst-1",
			TenantID: "tenant-1",
			Stages: []domain.StageInstance{{
				ID:               "stage-1",
				Status:           domain.StageActive,
				EligibleActorIDs: []string{"actor-1"},
			}},
		}},
		idempStore: &fakeIdempStore{},
	}
	mux := docSignoffTestMux(h)

	body := `{"decision":"approve","reason":"","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "\"v7\"")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if fakeSvc.gotReq.ExpectedRevisionVersion != 7 {
		t.Fatalf("expected revision = %d, want 7", fakeSvc.gotReq.ExpectedRevisionVersion)
	}
}

func TestSignoffByDocumentHandler_RequiresIfMatch(t *testing.T) {
	h := &Handler{
		decisionSvc: &fakeDecisionService{},
		readSvc:     &fakeReadService{},
	}
	mux := docSignoffTestMux(h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(`{"decision":"approve","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusPreconditionRequired, rr.Body.String())
	}
}

func TestSignoffByDocumentHandler_RejectsIfMatchV0(t *testing.T) {
	h := &Handler{
		decisionSvc: &fakeDecisionService{},
		readSvc:     &fakeReadService{},
	}
	mux := docSignoffTestMux(h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(`{"decision":"approve","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "v0")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestSignoffByDocumentHandler_RequiresValidContentHash(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{
			name: "missing content hash",
			body: `{"decision":"approve","password":"secret"}`,
		},
		{
			name: "invalid content hash",
			body: `{"decision":"approve","password":"secret","content_hash":"abc"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{
				decisionSvc: &fakeDecisionService{},
				readSvc: &fakeReadService{inst: &domain.Instance{
					ID:       "inst-1",
					TenantID: "tenant-1",
					Stages: []domain.StageInstance{{
						ID:               "stage-1",
						Status:           domain.StageActive,
						EligibleActorIDs: []string{"actor-1"},
					}},
				}},
			}
			mux := docSignoffTestMux(h)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", "idem-1")
			req.Header.Set("If-Match", "\"v7\"")
			req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
			req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
			}
		})
	}
}

func TestSignoffByDocumentHandler_UsesMutationLookupInsteadOfDocumentViewRead(t *testing.T) {
	fakeSvc := &fakeDecisionService{result: application.SignoffResult{StageCompleted: true}}
	readSvc := &fakeReadService{
		err: authz.ErrCapDenied{
			Capability: "document.view",
			AreaCode:   "qa",
			ActorID:    "actor-1",
		},
		mutationInst: &domain.Instance{
			ID:       "inst-1",
			TenantID: "tenant-1",
			Stages: []domain.StageInstance{{
				ID:               "stage-1",
				Status:           domain.StageActive,
				EligibleActorIDs: []string{"actor-1"},
			}},
		},
	}
	h := &Handler{
		decisionSvc: fakeSvc,
		readSvc:     readSvc,
		idempStore:  &fakeIdempStore{},
	}
	mux := docSignoffTestMux(h)

	body := `{"decision":"approve","reason":"","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "\"v7\"")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !readSvc.loadActiveByDocumentForMutationCalled {
		t.Fatalf("mutation lookup was not called")
	}
	if readSvc.loadActiveByDocumentCalled {
		t.Fatalf("read lookup should not be called on signoff mutation path")
	}
}

func TestCancelByDocumentHandler_UsesMutationLookupInsteadOfDocumentViewRead(t *testing.T) {
	cancelSvc := &fakeCancelService{result: application.CancelResult{DocumentID: "doc-1"}}
	readSvc := &fakeReadService{
		err: authz.ErrCapDenied{
			Capability: "document.view",
			AreaCode:   "qa",
			ActorID:    "actor-1",
		},
		mutationInst: &domain.Instance{
			ID:         "inst-1",
			DocumentID: "doc-1",
			TenantID:   "tenant-1",
		},
	}
	h := &Handler{
		readSvc:   readSvc,
		cancelSvc: cancelSvc,
	}
	mux := docCancelTestMux(h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/cancel", strings.NewReader(`{"reason":"withdrawn"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("If-Match", "\"v6\"")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !readSvc.loadActiveByDocumentForMutationCalled {
		t.Fatalf("mutation lookup was not called")
	}
	if readSvc.loadActiveByDocumentCalled {
		t.Fatalf("read lookup should not be called on cancel mutation path")
	}
}
