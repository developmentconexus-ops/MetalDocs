package approvalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type fakeFastForwardService struct {
	gotReq application.FastForwardRequest
	result application.FastForwardResult
	err    error
	calls  int
}

func (f *fakeFastForwardService) RecordFastForward(_ context.Context, _ db.TxRunner, req application.FastForwardRequest) (application.FastForwardResult, error) {
	f.calls++
	f.gotReq = req
	if f.err != nil {
		return application.FastForwardResult{}, f.err
	}
	return f.result, nil
}

// BeginFastForwardReplay extends the signoff_handler_test.go fakeIdempStore
// (same package) so it satisfies fastForwardIdempStore too, letting tests in
// this file reuse the identical fake slot-ledger machinery (begin/entries/
// slots/fakeReplayHandle) rather than duplicating it.
func (f *fakeIdempStore) BeginFastForwardReplay(_ context.Context, tenantID, actorID, key, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error) {
	return f.begin("fast-forward", tenantID, actorID, key, payloadHash)
}

func fastForwardTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/fast-forward", h.FastForwardHandler)
	return mux
}

func TestFastForwardHandler_Happy(t *testing.T) {
	fakeSvc := &fakeFastForwardService{result: application.FastForwardResult{
		Signoff: application.SignoffResult{SignoffID: "sig-ff-1", InstanceApproved: true},
	}}
	h := &Handler{fastForwardSvc: fakeSvc, fastForwardIdempStore: &fakeIdempStore{entries: map[string]application.SignoffReplay{}}}
	mux := fastForwardTestMux(h)

	body := `{"comment":"fast forward","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "v3")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var out contracts.FastForwardResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Outcome != "approved" {
		t.Fatalf("outcome = %q, want %q", out.Outcome, "approved")
	}
	// F-QA4-7: the signoff leg's persisted id must reach the wire, not "".
	if out.SignoffID != "sig-ff-1" {
		t.Fatalf("signoff_id = %q, want %q", out.SignoffID, "sig-ff-1")
	}
	if fakeSvc.gotReq.TenantID != "tenant-1" || fakeSvc.gotReq.InstanceID != "inst-1" || fakeSvc.gotReq.StageInstanceID != "stg-1" || fakeSvc.gotReq.ActorUserID != "actor-1" {
		t.Fatalf("unexpected request mapped to service: %+v", fakeSvc.gotReq)
	}
	if fakeSvc.gotReq.SignatureMethod != "password_reauth" {
		t.Fatalf("signature_method = %q, want %q", fakeSvc.gotReq.SignatureMethod, "password_reauth")
	}
	if got := fakeSvc.gotReq.SignaturePayload["password_token"]; got != "secret" {
		t.Fatalf("signature payload password_token = %#v", got)
	}
	if fakeSvc.gotReq.ContentHash != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("content hash = %#v", fakeSvc.gotReq.ContentHash)
	}
}

func TestFastForwardHandler_MissingIdempotencyKey(t *testing.T) {
	h := &Handler{fastForwardSvc: &fakeFastForwardService{}}
	mux := fastForwardTestMux(h)

	body := `{"comment":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
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
	if out.Code.String() != "idempotency.key_required" {
		t.Fatalf("error.code = %q, want %q", out.Code, "idempotency.key_required")
	}
}

func TestFastForwardHandler_StageNotCompleted(t *testing.T) {
	h := &Handler{fastForwardSvc: &fakeFastForwardService{err: domain.ErrFastForwardStageNotCompleted}, fastForwardIdempStore: &fakeIdempStore{entries: map[string]application.SignoffReplay{}}}
	mux := fastForwardTestMux(h)

	body := `{"comment":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "v1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
	}
	var out problem.Problem
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Code.String() != "state.fast_forward_stage_not_completed" {
		t.Fatalf("error.code = %q, want %q", out.Code, "state.fast_forward_stage_not_completed")
	}
}

func TestFastForwardHandler_NotEligible(t *testing.T) {
	h := &Handler{fastForwardSvc: &fakeFastForwardService{err: domain.ErrFastForwardNotEligible}, fastForwardIdempStore: &fakeIdempStore{entries: map[string]application.SignoffReplay{}}}
	mux := fastForwardTestMux(h)

	body := `{"comment":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "v1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
	}
	var out problem.Problem
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Code.String() != "state.fast_forward_not_eligible" {
		t.Fatalf("error.code = %q, want %q", out.Code, "state.fast_forward_not_eligible")
	}
}

// TestFastForwardHandler_ReplayReturnsCachedOutcome mirrors
// TestSignoffHandler_ReplayReturnsCachedOutcome (signoff_handler_test.go): a
// second call with the same Idempotency-Key returns the cached outcome
// without invoking the fast-forward service again.
func TestFastForwardHandler_ReplayReturnsCachedOutcome(t *testing.T) {
	fakeSvc := &fakeFastForwardService{}
	h := &Handler{
		fastForwardSvc: fakeSvc,
		fastForwardIdempStore: &fakeIdempStore{entries: map[string]application.SignoffReplay{
			"fast-forward:tenant-1:actor-1:88888888-8888-4888-8888-888888888888": {Outcome: "approved", SignoffID: "sig-replayed-1"},
		}},
	}
	mux := fastForwardTestMux(h)

	body := `{"comment":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "88888888-8888-4888-8888-888888888888")
	req.Header.Set("If-Match", "v3")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var out contracts.FastForwardResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.WasReplay || out.Outcome != "approved" {
		t.Fatalf("response = %+v, want replay approved", out)
	}
	// F-QA4-7: a replay must return the ORIGINAL signoff id, not "".
	if out.SignoffID != "sig-replayed-1" {
		t.Fatalf("replayed signoff_id = %q, want %q", out.SignoffID, "sig-replayed-1")
	}
	if fakeSvc.calls != 0 {
		t.Fatalf("fast-forward service should not run on replay, got %d calls", fakeSvc.calls)
	}
}

// TestFastForwardHandler_NonUUIDIdempotencyKey pins F-QA4-6 on the fast-forward
// route, which — like the signoff routes — is deliberately excluded from the
// idempotency.Require middleware and so must enforce the shared UUID wire rule
// itself.
func TestFastForwardHandler_NonUUIDIdempotencyKey(t *testing.T) {
	fakeSvc := &fakeFastForwardService{}
	h := &Handler{fastForwardSvc: fakeSvc, fastForwardIdempStore: &fakeIdempStore{entries: map[string]application.SignoffReplay{}}}
	mux := fastForwardTestMux(h)

	body := `{"comment":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-free-string")
	req.Header.Set("If-Match", "v1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("content-type = %q, want application/problem+json", ct)
	}
	var out problem.Problem
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Code != problem.CodeIdempotencyKeyInvalid {
		t.Fatalf("error.code = %q, want %q", out.Code, problem.CodeIdempotencyKeyInvalid)
	}
	if fakeSvc.calls != 0 {
		t.Fatalf("fast-forward service ran %d times, want 0", fakeSvc.calls)
	}
}

// TestFastForwardHandler_NilIdempStoreFailsClosed guards the no-fallback
// principle: if the handler was constructed (or misconfigured) without a
// fastForwardIdempStore, it must refuse the request with a 500 rather than
// silently proceeding without replay protection on an idempotency guarantee.
func TestFastForwardHandler_NilIdempStoreFailsClosed(t *testing.T) {
	fakeSvc := &fakeFastForwardService{}
	h := &Handler{fastForwardSvc: fakeSvc, fastForwardIdempStore: nil}
	mux := fastForwardTestMux(h)

	body := `{"comment":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	req.Header.Set("If-Match", "v3")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if fakeSvc.calls != 0 {
		t.Fatalf("fast-forward service should not run when idemp store is nil, got %d calls", fakeSvc.calls)
	}
	var out problem.Problem
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Code.String() != "internal.unknown" {
		t.Fatalf("error.code = %q, want %q", out.Code, "internal.unknown")
	}
}

func TestFastForwardHandler_MissingIfMatch(t *testing.T) {
	h := &Handler{fastForwardSvc: &fakeFastForwardService{}}
	mux := fastForwardTestMux(h)

	body := `{"comment":"","password_token":"secret","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/fast-forward", strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusPreconditionRequired)
	}
}
