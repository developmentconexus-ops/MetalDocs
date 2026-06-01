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
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type fakeDecisionService struct {
	gotReq application.SignoffRequest
	result application.SignoffResult
	err    error
	calls  int
}

func (f *fakeDecisionService) RecordSignoff(_ context.Context, _ *sql.DB, req application.SignoffRequest) (application.SignoffResult, error) {
	f.calls++
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

// fakeIdempStore is a faithful in-memory double of the Postgres idempotency
// store. Slot identity is the composite key (kind, tenant, actor, idempKey) —
// mirroring the real PK (tenant, actor, route_template, key) — and the payload
// hash is enforced as a misuse guard: re-claiming a slot with a different hash
// returns idempotency.ErrConflict, exactly as the production store does at
// postgres_store.go. Slots progress in_flight → completed (Complete) or are
// released (Fail), so the record→replay lifecycle is exercised for real.
type idempSlot struct {
	hash      string
	outcome   string
	completed bool
}

type fakeIdempStore struct {
	// Pre-seeded hash-agnostic outcomes, keyed by composite slot key. Used by
	// tests that assert replay without exercising the record path.
	entries map[string]string

	// Faithful slot ledger populated by the record path.
	slots map[string]*idempSlot

	// Instrumentation from the most recent Begin*Replay call.
	gotHash   string
	completed bool
	failed    bool
}

func (f *fakeIdempStore) begin(kind, tenantID, actorID, key, payloadHash string) (approvalinfra.SignoffReplayCommitter, *approvalinfra.SignoffReplay, error) {
	f.gotHash = payloadHash
	k := kind + ":" + tenantID + ":" + actorID + ":" + key
	if v, ok := f.entries[k]; ok {
		f.completed = true
		return nil, &approvalinfra.SignoffReplay{Outcome: v}, nil
	}
	if f.slots == nil {
		f.slots = map[string]*idempSlot{}
	}
	if s, ok := f.slots[k]; ok {
		if s.hash != payloadHash {
			return nil, nil, idempotency.ErrConflict
		}
		if s.completed {
			f.completed = true
			return nil, &approvalinfra.SignoffReplay{Outcome: s.outcome}, nil
		}
	}
	s := &idempSlot{hash: payloadHash}
	f.slots[k] = s
	return &fakeReplayHandle{store: f, key: k, slot: s}, nil, nil
}

func (f *fakeIdempStore) BeginDocumentReplay(_ context.Context, tenantID, actorID, key, payloadHash string) (approvalinfra.SignoffReplayCommitter, *approvalinfra.SignoffReplay, error) {
	return f.begin("document", tenantID, actorID, key, payloadHash)
}

func (f *fakeIdempStore) BeginStageReplay(_ context.Context, tenantID, actorID, key, payloadHash string) (approvalinfra.SignoffReplayCommitter, *approvalinfra.SignoffReplay, error) {
	return f.begin("stage", tenantID, actorID, key, payloadHash)
}

type fakeReplayHandle struct {
	store *fakeIdempStore
	key   string
	slot  *idempSlot
}

func (h *fakeReplayHandle) Complete(outcome string) error {
	h.slot.completed = true
	h.slot.outcome = outcome
	h.store.completed = true
	return nil
}

func (h *fakeReplayHandle) Fail(_ error) error {
	// Mirror FailReplay rolling back the in_flight row: the slot is released so a
	// fresh attempt (any hash) can re-claim it. No outcome is recorded.
	delete(h.store.slots, h.key)
	h.store.failed = true
	return nil
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

// TestSignoffByDocumentHandler_ReplayWinsOverPostLoadEligibilityChange verifies
// that once a (tenant, actor, key, payloadHash) tuple has a cached outcome, a
// retry returns the replay even if the underlying instance / stage has rotated
// so the original actor is no longer eligible. This is the corrected
// RFC-style idempotency contract — replay before state validation.
func TestSignoffByDocumentHandler_ReplayWinsOverPostLoadEligibilityChange(t *testing.T) {
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

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var out contracts.SignoffResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.WasReplay || out.Outcome != "approved" {
		t.Fatalf("response = %+v, want was_replay=true outcome=approved", out)
	}
}

// TestSignoffByDocumentHandler_FirstCallTerminalStateReturnsError verifies that
// a FIRST call (no cached outcome) against an instance whose active stage is
// nil returns the terminal-state error. The replay-before-validation reorder
// must not swallow the error or commit a successful outcome to the slot.
func TestSignoffByDocumentHandler_FirstCallTerminalStateReturnsError(t *testing.T) {
	store := &fakeIdempStore{entries: map[string]string{}}
	decisionSvc := &fakeDecisionService{}
	h := &Handler{
		decisionSvc: decisionSvc,
		readSvc: &fakeReadService{inst: &domain.Instance{
			ID:       "inst-1",
			TenantID: "tenant-1",
			Stages: []domain.StageInstance{{
				ID:               "stage-1",
				Status:           domain.StageCompleted,
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
	req.Header.Set("Idempotency-Key", "idem-first-terminal")
	req.Header.Set("If-Match", "\"v5\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusConflict, rr.Body.String())
	}
	if store.completed {
		t.Fatalf("idempotency slot completed on terminal-state path; want no successful commit")
	}
	if decisionSvc.gotReq.InstanceID != "" {
		t.Fatalf("decision service should not run on terminal-state path, got %+v", decisionSvc.gotReq)
	}
}

// TestSignoffByDocumentHandler_FingerprintExcludesServerResolvedState verifies
// that the doc-scoped replay fingerprint is derived only from client-stable
// inputs (path document ID + body), never from the server-resolved instance or
// active-stage IDs. This is the root-cause guard for F-002: hashing mutable
// state made an identical retry produce a different fingerprint after the stage
// rotated or the instance went terminal.
func TestSignoffByDocumentHandler_FingerprintExcludesServerResolvedState(t *testing.T) {
	const (
		decisionStr = "approve"
		reason      = "F-002 fingerprint"
		contentHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	store := &fakeIdempStore{}
	h := &Handler{
		decisionSvc: &fakeDecisionService{result: application.SignoffResult{InstanceApproved: true}},
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

	body := `{"decision":"approve","reason":"F-002 fingerprint","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "idem-fp")
	req.Header.Set("If-Match", "\"v5\"")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	wantClientStable := signoffPayloadHash("doc-1", "", "", domain.Decision(decisionStr), reason, contentHash)
	if store.gotHash != wantClientStable {
		t.Fatalf("fingerprint = %q, want client-stable %q", store.gotHash, wantClientStable)
	}
	// The fingerprint MUST NOT match one computed from the resolved instance/stage
	// IDs — that was the bug.
	withServerState := signoffPayloadHash("doc-1", "inst-1", "stage-1", domain.Decision(decisionStr), reason, contentHash)
	if store.gotHash == withServerState {
		t.Fatalf("fingerprint must not depend on resolved instance/stage IDs")
	}
}

// TestSignoffByDocumentHandler_RecordsThenReplaysAcrossTerminalTransition is the
// authoritative F-002 regression: a first call records the outcome while the
// stage is active, then the SAME key replays the cached outcome after the
// instance has gone terminal — driving the real active→terminal recording
// transition end-to-end (not a pre-seeded slot). This is the exact flow that the
// live API previously returned 500 on.
func TestSignoffByDocumentHandler_RecordsThenReplaysAcrossTerminalTransition(t *testing.T) {
	store := &fakeIdempStore{}
	decisionSvc := &fakeDecisionService{result: application.SignoffResult{InstanceApproved: true}}
	read := &fakeReadService{inst: &domain.Instance{
		ID:       "inst-1",
		TenantID: "tenant-1",
		Stages: []domain.StageInstance{{
			ID:               "stage-1",
			Status:           domain.StageActive,
			EligibleActorIDs: []string{"actor-1"},
		}},
	}}
	h := &Handler{decisionSvc: decisionSvc, readSvc: read, idempStore: store}
	mux := docSignoffTestMux(h)

	body := `{"decision":"approve","reason":"","password":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
		req.Header.Set("Idempotency-Key", "idem-terminal")
		req.Header.Set("If-Match", "\"v5\"")
		return req
	}

	// First call: stage active, actor eligible → records the outcome.
	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, newReq())
	if rr1.Code != http.StatusOK {
		t.Fatalf("first call status = %d, want 200; body: %s", rr1.Code, rr1.Body.String())
	}
	var out1 contracts.SignoffResponse
	if err := json.NewDecoder(rr1.Body).Decode(&out1); err != nil {
		t.Fatalf("decode first: %v", err)
	}
	if out1.WasReplay || out1.Outcome != "approved" {
		t.Fatalf("first response = %+v, want fresh approved", out1)
	}

	// Instance now goes terminal out-of-band (no active stage).
	read.inst = &domain.Instance{
		ID:       "inst-1",
		TenantID: "tenant-1",
		Stages: []domain.StageInstance{{
			ID:     "stage-1",
			Status: domain.StageCompleted,
		}},
	}

	// Retry with the SAME key: must replay the cached outcome, not 409/500.
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, newReq())
	if rr2.Code != http.StatusOK {
		t.Fatalf("retry status = %d, want 200 (replay after terminal); body: %s", rr2.Code, rr2.Body.String())
	}
	var out2 contracts.SignoffResponse
	if err := json.NewDecoder(rr2.Body).Decode(&out2); err != nil {
		t.Fatalf("decode retry: %v", err)
	}
	if !out2.WasReplay || out2.Outcome != "approved" {
		t.Fatalf("retry response = %+v, want was_replay=true outcome=approved", out2)
	}
	if decisionSvc.calls != 1 {
		t.Fatalf("decision service calls = %d, want 1 (replay must not re-run the signoff)", decisionSvc.calls)
	}
}

// TestSignoffByDocumentHandler_KeyReuseDifferentBodyConflicts verifies the
// misuse guard surfaces as 409 idempotency.key_conflict (not 500): once a key is
// recorded, reusing it with a different request body (different content hash)
// must be rejected as a conflict so the caller rotates the key.
func TestSignoffByDocumentHandler_KeyReuseDifferentBodyConflicts(t *testing.T) {
	store := &fakeIdempStore{}
	read := &fakeReadService{inst: &domain.Instance{
		ID:       "inst-1",
		TenantID: "tenant-1",
		Stages: []domain.StageInstance{{
			ID:               "stage-1",
			Status:           domain.StageActive,
			EligibleActorIDs: []string{"actor-1"},
		}},
	}}
	h := &Handler{
		decisionSvc: &fakeDecisionService{result: application.SignoffResult{InstanceApproved: true}},
		readSvc:     read,
		idempStore:  store,
	}
	mux := docSignoffTestMux(h)

	mk := func(contentHash string) *http.Request {
		body := `{"decision":"approve","reason":"","password":"secret","content_hash":"` + contentHash + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/signoff", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
		req.Header.Set("Idempotency-Key", "idem-reuse")
		req.Header.Set("If-Match", "\"v5\"")
		return req
	}

	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, mk("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	if rr1.Code != http.StatusOK {
		t.Fatalf("first call status = %d, want 200; body: %s", rr1.Code, rr1.Body.String())
	}

	// Same key, different content hash → conflict.
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, mk("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"))
	if rr2.Code != http.StatusConflict {
		t.Fatalf("reuse status = %d, want %d; body: %s", rr2.Code, http.StatusConflict, rr2.Body.String())
	}
	var prob problem.Problem
	if err := json.NewDecoder(rr2.Body).Decode(&prob); err != nil {
		t.Fatalf("decode conflict: %v", err)
	}
	if prob.Code != "idempotency.key_conflict" {
		t.Fatalf("error.code = %q, want %q", prob.Code, "idempotency.key_conflict")
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
