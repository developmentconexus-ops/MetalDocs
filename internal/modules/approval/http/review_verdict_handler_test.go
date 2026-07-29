package approvalhttp

// review_verdict_handler_test.go — HTTP-layer tests for ReviewVerdictHandler,
// reusing the shared fakes defined in signoff_handler_test.go (fakeIdempStore,
// fakeReplayHandle). No real DB.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

type fakeReviewVerdictService struct {
	gotReq application.ReviewVerdictRequest
	result application.ReviewVerdictResult
	err    error
	calls  int
}

func (f *fakeReviewVerdictService) RecordVerdict(_ context.Context, _ db.TxRunner, req application.ReviewVerdictRequest) (application.ReviewVerdictResult, error) {
	f.calls++
	f.gotReq = req
	if f.err != nil {
		return application.ReviewVerdictResult{}, f.err
	}
	return f.result, nil
}

func reviewVerdictTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/review-verdict", h.ReviewVerdictHandler)
	return mux
}

// TestReviewVerdictHandler_FreshCallCarriesVerdictID pins the F-QA4-7-class fix
// on the review-verdict route: the approval_review_verdicts row id threaded out
// on application.ReviewVerdictResult must reach the wire, not the previously
// hardcoded "".
func TestReviewVerdictHandler_FreshCallCarriesVerdictID(t *testing.T) {
	fakeSvc := &fakeReviewVerdictService{result: application.ReviewVerdictResult{
		VerdictID:      "verdict-abc-1",
		StageCompleted: true,
	}}
	h := &Handler{reviewVerdictSvc: fakeSvc, idempStore: &fakeIdempStore{}}
	mux := reviewVerdictTestMux(h)

	body := `{"verdict":"ready","comment":"looks good"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/review-verdict", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
	req.Header.Set("Idempotency-Key", "22222222-2222-4222-8222-222222222222")
	req.Header.Set("If-Match", "v3")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var out contracts.ReviewVerdictResponse
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.VerdictID != "verdict-abc-1" {
		t.Fatalf("verdict_id = %q, want %q", out.VerdictID, "verdict-abc-1")
	}
	if out.WasReplay || out.Outcome != "stage_completed" {
		t.Fatalf("response = %+v, want fresh stage_completed", out)
	}
	if fakeSvc.gotReq.TenantID != "tenant-1" || fakeSvc.gotReq.InstanceID != "inst-1" ||
		fakeSvc.gotReq.StageInstanceID != "stg-1" || fakeSvc.gotReq.ActorUserID != "actor-1" {
		t.Fatalf("unexpected request mapped to service: %+v", fakeSvc.gotReq)
	}
}

// TestReviewVerdictHandler_ReplayReturnsOriginalVerdictID drives the real
// record→replay lifecycle through the faithful in-memory store: the first call
// persists the verdict id in the replay envelope, and a retry with the SAME key
// round-trips that id instead of "".
func TestReviewVerdictHandler_ReplayReturnsOriginalVerdictID(t *testing.T) {
	store := &fakeIdempStore{}
	fakeSvc := &fakeReviewVerdictService{result: application.ReviewVerdictResult{
		VerdictID:        "verdict-replay-1",
		StageCompleted:   true,
		InstanceApproved: true,
	}}
	h := &Handler{reviewVerdictSvc: fakeSvc, idempStore: store}
	mux := reviewVerdictTestMux(h)

	body := `{"verdict":"ready","comment":"looks good"}`
	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/approval/instances/inst-1/stages/stg-1/review-verdict", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-1", []iamdomain.Role{}))
		req.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")
		req.Header.Set("If-Match", "v3")
		return req
	}

	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, newReq())
	if rr1.Code != http.StatusOK {
		t.Fatalf("first call status = %d, want 200; body: %s", rr1.Code, rr1.Body.String())
	}
	var out1 contracts.ReviewVerdictResponse
	if err := json.NewDecoder(rr1.Body).Decode(&out1); err != nil {
		t.Fatalf("decode first: %v", err)
	}
	if out1.WasReplay || out1.VerdictID != "verdict-replay-1" || out1.Outcome != "approved" {
		t.Fatalf("first response = %+v, want fresh approved with the persisted verdict id", out1)
	}

	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, newReq())
	if rr2.Code != http.StatusOK {
		t.Fatalf("retry status = %d, want 200; body: %s", rr2.Code, rr2.Body.String())
	}
	var out2 contracts.ReviewVerdictResponse
	if err := json.NewDecoder(rr2.Body).Decode(&out2); err != nil {
		t.Fatalf("decode retry: %v", err)
	}
	if !out2.WasReplay || out2.Outcome != "approved" {
		t.Fatalf("retry response = %+v, want was_replay=true outcome=approved", out2)
	}
	if out2.VerdictID != out1.VerdictID {
		t.Fatalf("replayed verdict_id = %q, want %q (same as first call)", out2.VerdictID, out1.VerdictID)
	}
	if fakeSvc.calls != 1 {
		t.Fatalf("review verdict service calls = %d, want 1 (replay must not re-run the verdict)", fakeSvc.calls)
	}
}
